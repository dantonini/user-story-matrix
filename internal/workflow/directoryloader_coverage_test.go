// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

// TestLoadWorkflowFromDirectory_ErrorCases tests error cases in LoadWorkflowFromDirectory
func TestLoadWorkflowFromDirectory_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(fs *io.MockFileSystem)
		dirPath       string
		expectedError string
	}{
		{
			name: "Directory does not exist",
			setupMock: func(fs *io.MockFileSystem) {
				// Don't add the directory
			},
			dirPath:       "/path/to/workflow",
			expectedError: "workflow directory not found",
		},
		{
			name: "Workflow YAML does not exist",
			setupMock: func(fs *io.MockFileSystem) {
				fs.AddDirectory("/path/to/workflow")
				fs.AddDirectory("/path/to/workflow/prompts")
			},
			dirPath:       "/path/to/workflow",
			expectedError: "workflow configuration file not found",
		},
		{
			name: "Prompts directory does not exist",
			setupMock: func(fs *io.MockFileSystem) {
				fs.AddDirectory("/path/to/workflow")
				fs.AddFile("/path/to/workflow/workflow.yaml", []byte("name: test-workflow\ndescription: Test workflow"))
			},
			dirPath:       "/path/to/workflow",
			expectedError: "prompts directory not found",
		},
		{
			name: "Read workflow.yaml error",
			setupMock: func(fs *io.MockFileSystem) {
				fs.AddDirectory("/path/to/workflow")
				fs.AddDirectory("/path/to/workflow/prompts")
				fs.AddFile("/path/to/workflow/workflow.yaml", []byte("dummy content"))
				fs.SetWriteFileError("/path/to/workflow/workflow.yaml", errors.New("read error"))
			},
			dirPath:       "/path/to/workflow",
			expectedError: "invalid YAML in workflow configuration",
		},
		{
			name: "Invalid YAML in workflow.yaml",
			setupMock: func(fs *io.MockFileSystem) {
				fs.AddDirectory("/path/to/workflow")
				fs.AddDirectory("/path/to/workflow/prompts")
				fs.AddFile("/path/to/workflow/workflow.yaml", []byte("invalid: yaml: content\n  indentation is wrong"))
			},
			dirPath:       "/path/to/workflow",
			expectedError: "invalid YAML in workflow configuration",
		},
		{
			name: "Validation error - missing prompt",
			setupMock: func(fs *io.MockFileSystem) {
				fs.AddDirectory("/path/to/workflow")
				fs.AddDirectory("/path/to/workflow/prompts")
				// Create YAML with steps but missing prompt files
				yaml := `
name: test-workflow
description: Test workflow
steps:
  - id: step1
    description: Step 1
    prompt: prompts/step1.md
  - id: step2
    description: Step 2
    prompt: prompts/step2.md
`
				fs.AddFile("/path/to/workflow/workflow.yaml", []byte(yaml))
				// Only add one prompt file
				fs.AddFile("/path/to/workflow/prompts/step1.md", []byte("Step 1 prompt"))
			},
			dirPath:       "/path/to/workflow",
			expectedError: "workflow validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock filesystem
			fs := io.NewMockFileSystem()
			tt.setupMock(fs)

			// Call function
			_, _, err := LoadWorkflowFromDirectory(fs, tt.dirPath)

			// Check error
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestValidateDirectoryWorkflow_ExtraTests adds additional tests to increase coverage
func TestValidateDirectoryWorkflow_ExtraTests(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(fs *io.MockFileSystem)
		dirPath       string
		expectedError string
	}{
		{
			name: "Directory does not exist",
			setupMock: func(fs *io.MockFileSystem) {
				// Don't add the directory
			},
			dirPath:       "/path/to/workflow",
			expectedError: "workflow directory not found",
		},
		{
			name: "Read workflow.yaml error",
			setupMock: func(fs *io.MockFileSystem) {
				fs.AddDirectory("/path/to/workflow")
				fs.AddFile("/path/to/workflow/workflow.yaml", []byte("dummy content"))
				fs.SetWriteFileError("/path/to/workflow/workflow.yaml", errors.New("read error"))
			},
			dirPath:       "/path/to/workflow",
			expectedError: "invalid YAML in workflow configuration",
		},
		{
			name: "Invalid external workflow - missing name",
			setupMock: func(fs *io.MockFileSystem) {
				fs.AddDirectory("/path/to/workflow")
				yaml := `
description: Workflow without a name
steps:
  - id: step1
    description: Step 1
    prompt: This is an embedded prompt
`
				fs.AddFile("/path/to/workflow/workflow.yaml", []byte(yaml))
			},
			dirPath:       "/path/to/workflow",
			expectedError: "workflow must have a name",
		},
		{
			name: "Missing prompts directory",
			setupMock: func(fs *io.MockFileSystem) {
				fs.AddDirectory("/path/to/workflow")
				yaml := `
name: test-workflow
description: Test workflow
steps:
  - id: step1
    description: Step 1
    prompt: prompts/step1.md
`
				fs.AddFile("/path/to/workflow/workflow.yaml", []byte(yaml))
				// Don't add prompts directory
			},
			dirPath:       "/path/to/workflow",
			expectedError: "prompt file for step",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock filesystem
			fs := io.NewMockFileSystem()
			tt.setupMock(fs)

			// Call function
			errs := ValidateDirectoryWorkflow(fs, tt.dirPath)

			// Check error
			assert.NotEmpty(t, errs, "Expected validation errors")
			assert.Contains(t, errs[0].Error(), tt.expectedError)
		})
	}
} 
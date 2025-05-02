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

// TestExtractStandardWorkflow_ErrorHandling tests error handling in ExtractStandardWorkflow
func TestExtractStandardWorkflow_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(fs *io.MockFileSystem)
		outputDir     string
		expectedError string
	}{
		{
			name: "Create output directory error",
			setupMock: func(fs *io.MockFileSystem) {
				fs.SetMkdirAllError("output/dir", errors.New("mkdir error"))
			},
			outputDir:     "output/dir",
			expectedError: "failed to create output directory",
		},
		{
			name: "Create prompts directory error",
			setupMock: func(fs *io.MockFileSystem) {
				// Let the output dir creation succeed
				fs.SetMkdirAllError("output/dir/prompts", errors.New("mkdir error"))
			},
			outputDir:     "output/dir",
			expectedError: "failed to create prompts directory",
		},
		{
			name: "Write README.md error",
			setupMock: func(fs *io.MockFileSystem) {
				// Create the directories but fail on README.md
				fs.SetWriteFileError("output/dir/README.md", errors.New("write error"))
			},
			outputDir:     "output/dir",
			expectedError: "failed to create README.md",
		},
		{
			name: "Write prompts README.md error",
			setupMock: func(fs *io.MockFileSystem) {
				// Create the directories but fail on prompts/README.md
				fs.SetWriteFileError("output/dir/prompts/README.md", errors.New("write error"))
			},
			outputDir:     "output/dir",
			expectedError: "failed to create prompts README.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock filesystem
			fs := io.NewMockFileSystem()
			tt.setupMock(fs)

			// Call function
			err := ExtractStandardWorkflow(fs, tt.outputDir)

			// Check error
			assert.Error(t, err, "Expected an error")
			assert.Contains(t, err.Error(), tt.expectedError, "Expected error message to contain %q", tt.expectedError)
		})
	}
}

// TestGenerateWorkflowYAML_ErrorHandling tests error handling in generateWorkflowYAML
func TestGenerateWorkflowYAML_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(fs *io.MockFileSystem)
		outputPath    string
		expectedError string
	}{
		{
			name: "Write file error",
			setupMock: func(fs *io.MockFileSystem) {
				fs.SetWriteFileError("output/dir/workflow.yaml", errors.New("write error"))
			},
			outputPath:    "output/dir/workflow.yaml",
			expectedError: "failed to write workflow YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock filesystem
			fs := io.NewMockFileSystem()
			tt.setupMock(fs)

			// Call function with minimal steps
			steps := []WorkflowStep{
				{ID: "step1", Description: "Step 1", Prompt: "Step 1 prompt"},
			}
			promptPaths := map[string]string{"step1": "prompts/step1.md"}

			// Call function
			err := generateWorkflowYAML(fs, steps, tt.outputPath, promptPaths)

			// Check error
			assert.Error(t, err, "Expected an error")
			assert.Contains(t, err.Error(), tt.expectedError, "Expected error message to contain %q", tt.expectedError)
		})
	}
}

// TestExtractPromptToFile_ErrorHandling tests error handling in extractPromptToFile
func TestExtractPromptToFile_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(fs *io.MockFileSystem)
		promptsDir    string
		step          WorkflowStep
		expectedError string
	}{
		{
			name: "Write file error",
			setupMock: func(fs *io.MockFileSystem) {
				fs.SetWriteFileError("output/dir/prompts/step1.md", errors.New("write error"))
			},
			promptsDir: "output/dir/prompts",
			step: WorkflowStep{
				ID:          "step1",
				Description: "Step 1",
				Prompt:      "Test prompt content",
			},
			expectedError: "failed to write prompt file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock filesystem
			fs := io.NewMockFileSystem()
			tt.setupMock(fs)

			// Call function
			_, err := extractPromptToFile(fs, tt.promptsDir, tt.step)

			// Check error
			assert.Error(t, err, "Expected an error")
			assert.Contains(t, err.Error(), tt.expectedError, "Expected error message to contain %q", tt.expectedError)
		})
	}
} 
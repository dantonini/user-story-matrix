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

// TestValidateTemplate_AdditionalCases tests more edge cases for ValidateTemplate
func TestValidateTemplate_AdditionalCases(t *testing.T) {
	tests := []struct {
		name           string
		fs             io.FileSystem
		templatePath   string
		expectedErrors []string
	}{
		{
			name:           "Non-existent file",
			fs:             io.NewMockFileSystem(),
			templatePath:   "non-existent.tmpl",
			expectedErrors: []string{"prompt file not found"},
		},
		{
			name: "File read error",
			fs: func() io.FileSystem {
				fs := io.NewMockFileSystemWithErrors()
				fs.AddFile("error.tmpl", []byte("test"))
				fs.SetReadError("error.tmpl", errors.New("read error"))
				return fs
			}(),
			templatePath:   "error.tmpl",
			expectedErrors: []string{"failed to read prompt file"},
		},
		{
			name: "Invalid template syntax",
			fs: func() io.FileSystem {
				fs := io.NewMockFileSystem()
				fs.AddFile("invalid.tmpl", []byte("Hello {{.name"))
				return fs
			}(),
			templatePath:   "invalid.tmpl",
			expectedErrors: []string{"invalid template syntax"},
		},
		{
			name: "Absolute path with invalid template",
			fs: func() io.FileSystem {
				fs := io.NewMockFileSystem()
				fs.AddFile("/absolute/path/invalid.tmpl", []byte("Hello {{.name"))
				return fs
			}(),
			templatePath:   "/absolute/path/invalid.tmpl",
			expectedErrors: []string{"invalid template syntax"},
		},
		{
			name: "Valid absolute path template",
			fs: func() io.FileSystem {
				fs := io.NewMockFileSystem()
				fs.AddFile("/absolute/path/valid.tmpl", []byte("Hello {{.name}}"))
				return fs
			}(),
			templatePath:   "/absolute/path/valid.tmpl",
			expectedErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTemplateRenderer(tt.fs, "")
			err := renderer.ValidateTemplate(tt.templatePath)
			
			if tt.expectedErrors != nil {
				assert.Error(t, err, "Expected an error")
				for _, expectedErrText := range tt.expectedErrors {
					assert.Contains(t, err.Error(), expectedErrText, "Expected error message to contain %q", expectedErrText)
				}
			} else {
				assert.NoError(t, err, "Expected no error")
			}
		})
	}
}

// TestRenderPrompt_ErrorCases tests error cases for RenderPrompt
func TestRenderPrompt_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		fs            io.FileSystem
		templatePath  string
		variables     map[string]interface{}
		expectedError string
	}{
		{
			name:          "Non-existent file",
			fs:            io.NewMockFileSystem(),
			templatePath:  "non-existent.tmpl",
			variables:     map[string]interface{}{"name": "World"},
			expectedError: "prompt file not found",
		},
		{
			name: "File read error",
			fs: func() io.FileSystem {
				fs := io.NewMockFileSystemWithErrors()
				fs.AddFile("error.tmpl", []byte("test"))
				fs.SetReadError("error.tmpl", errors.New("read error"))
				return fs
			}(),
			templatePath:  "error.tmpl",
			variables:     map[string]interface{}{"name": "World"},
			expectedError: "failed to read prompt file",
		},
		{
			name: "Parse error",
			fs: func() io.FileSystem {
				fs := io.NewMockFileSystem()
				fs.AddFile("invalid.tmpl", []byte("Hello {{.name"))
				return fs
			}(),
			templatePath:  "invalid.tmpl",
			variables:     map[string]interface{}{"name": "World"},
			expectedError: "invalid template syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTemplateRenderer(tt.fs, "")
			_, err := renderer.RenderPrompt(tt.templatePath, tt.variables)

			assert.Error(t, err, "Expected an error")
			assert.Contains(t, err.Error(), tt.expectedError, "Expected error message to contain %q", tt.expectedError)
		})
	}
} 
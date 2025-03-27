// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// testFileSystem is a mock implementation of FileSystem for testing
type testFileSystem struct {
	files    map[string][]byte
	exists   map[string]bool
	mkdirErr error
}

func newTestFileSystem() *testFileSystem {
	return &testFileSystem{
		files:  make(map[string][]byte),
		exists: make(map[string]bool),
	}
}

func (m *testFileSystem) ReadFile(path string) ([]byte, error) {
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func (m *testFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.files[path] = data
	m.exists[path] = true
	return nil
}

func (m *testFileSystem) Exists(path string) bool {
	return m.exists[path]
}

func (m *testFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirErr != nil {
		return m.mkdirErr
	}
	m.exists[path] = true
	return nil
}

// testUserOutput is a mock implementation of UserOutput for testing
type testUserOutput struct {
	messages         []string
	progressMessages []string
	errorMessages    []string
	warningMessages  []string
	stepMessages     []string
	successMessages  []string
}

func newTestUserOutput() *testUserOutput {
	return &testUserOutput{
		messages:         make([]string, 0),
		progressMessages: make([]string, 0),
		errorMessages:    make([]string, 0),
		warningMessages:  make([]string, 0),
		stepMessages:     make([]string, 0),
		successMessages:  make([]string, 0),
	}
}

func (t *testUserOutput) Print(msg string) {
	t.messages = append(t.messages, msg)
}

func (t *testUserOutput) PrintSuccess(msg string) {
	t.successMessages = append(t.successMessages, msg)
}

func (t *testUserOutput) PrintError(msg string) {
	t.errorMessages = append(t.errorMessages, msg)
}

func (t *testUserOutput) PrintWarning(msg string) {
	t.warningMessages = append(t.warningMessages, msg)
}

func (t *testUserOutput) PrintProgress(msg string) {
	t.progressMessages = append(t.progressMessages, msg)
}

func (t *testUserOutput) PrintStep(current int, total int, msg string) {
	t.stepMessages = append(t.stepMessages, fmt.Sprintf("Step %d/%d: %s", current, total, msg))
}

func TestStepExecutor_ExecuteStep(t *testing.T) {
	tests := []struct {
		name           string
		changeRequest  string
		step          WorkflowStep
		wantSuccess    bool
		wantErrorText  string
		expectedOutput []string
	}{
		{
			name: "Successful execution",
			changeRequest: `# Test Change Request
This is a test change request.`,
			step: WorkflowStep{
				ID:          "01-laying-the-foundation",
				Description: "Laying the foundation",
				IsTest:      false,
				OutputFile:  "%s.01-laying-the-foundation.md",
			},
			wantSuccess: true,
			expectedOutput: []string{
				"# Laying the foundation",
				"## Architecture & Design",
				"This step focuses on setting up the architecture and structure for the implementation.",
				"### Key Activities",
				"1. Create necessary packages and interfaces",
				"2. Define core data structures",
				"3. Establish file organization",
				"4. Set up testing infrastructure",
				"## Change Request Context",
				"This step was executed for change request:",
				"# Test Change Request",
				"This is a test change request.",
				"Step ID: 01-laying-the-foundation",
				"Step Description: Laying the foundation",
				"Is Test Step: false",
			},
		},
		{
			name:          "File not found",
			changeRequest: "",
			step: WorkflowStep{
				ID:          "01-laying-the-foundation",
				Description: "Laying the foundation",
				IsTest:      false,
				OutputFile:  "%s.01-laying-the-foundation.md",
			},
			wantSuccess:   false,
			wantErrorText: fmt.Sprintf(ErrFileNotFound, "change-request.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			fs := newTestFileSystem()
			io := newTestUserOutput()

			// Create executor
			executor := NewStepExecutor(fs, io)

			// Set up mock file system
			if tt.changeRequest != "" {
				fs.files["change-request.md"] = []byte(tt.changeRequest)
			}

			// Execute step
			success, err := executor.ExecuteStep("change-request.md", tt.step, "output.md")

			// Check success/failure
			if tt.wantSuccess {
				if !success || err != nil {
					t.Errorf("ExecuteStep() success = %v, error = %v, want success = true, error = nil", success, err)
				}

				// Check output file content
				output, exists := fs.files["output.md"]
				if !exists {
					t.Error("ExecuteStep() did not create output file")
				} else {
					outputStr := string(output)
					for _, expectedLine := range tt.expectedOutput {
						if !strings.Contains(outputStr, expectedLine) {
							t.Errorf("Expected line not found in output: %s", expectedLine)
						}
					}
				}
			} else {
				if success || err == nil {
					t.Error("ExecuteStep() expected error, got nil")
				} else if err.Error() != tt.wantErrorText {
					t.Errorf("ExecuteStep() error = %v, want %v", err, tt.wantErrorText)
				}
			}
		})
	}
}

func TestStepExecutor_ExecuteStep_FileSystemErrors(t *testing.T) {
	// Test cases for file system errors
	tests := []struct {
		name          string
		setupFS       func(*testFileSystem)
		wantSuccess   bool
		wantErrorText string
	}{
		{
			name: "change request file not found",
			setupFS: func(fs *testFileSystem) {
				// Don't add the change request file
			},
			wantSuccess:   false,
			wantErrorText: fmt.Sprintf(ErrFileNotFound, "change-request.md"),
		},
		{
			name: "mkdir error",
			setupFS: func(fs *testFileSystem) {
				fs.files["change-request.md"] = []byte("Test content")
				fs.exists["change-request.md"] = true
				fs.mkdirErr = fmt.Errorf("mkdir error")
			},
			wantSuccess:   false,
			wantErrorText: fmt.Sprintf(ErrOutputFileCreateFailed, "mkdir error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fs := newTestFileSystem()
			io := newTestUserOutput()
			executor := NewStepExecutor(fs, io)

			// Setup file system state
			tt.setupFS(fs)

			// Execute
			step := WorkflowStep{
				ID:          "01-laying-the-foundation",
				Description: "Laying the foundation",
				IsTest:      false,
			}
			success, err := executor.ExecuteStep("change-request.md", step, "output/test.md")

			// Check success/error
			if success != tt.wantSuccess {
				t.Errorf("ExecuteStep() success = %v, want %v", success, tt.wantSuccess)
			}
			if err == nil || err.Error() != tt.wantErrorText {
				t.Errorf("ExecuteStep() error = %v, want %v", err, tt.wantErrorText)
			}
		})
	}
} 
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
	messages []string
}

func newTestUserOutput() *testUserOutput {
	return &testUserOutput{
		messages: make([]string, 0),
	}
}

func (m *testUserOutput) Print(message string) {
	m.messages = append(m.messages, message)
}

func (m *testUserOutput) PrintSuccess(message string) {
	m.messages = append(m.messages, "SUCCESS: "+message)
}

func (m *testUserOutput) PrintError(message string) {
	m.messages = append(m.messages, "ERROR: "+message)
}

func TestStepExecutor_ExecuteStep(t *testing.T) {
	// Test cases
	tests := []struct {
		name             string
		changeRequest    string
		step            WorkflowStep
		outputFile      string
		wantSuccess     bool
		wantError       bool
		wantOutputLines []string
	}{
		{
			name:          "successful execution - foundation step",
			changeRequest: "Test change request content",
			step: WorkflowStep{
				ID:          "01-laying-the-foundation",
				Description: "Laying the foundation",
				IsTest:      false,
			},
			outputFile:  "output.md",
			wantSuccess: true,
			wantOutputLines: []string{
				"# Laying the foundation",
				"## Architecture & Design",
				"This step focuses on setting up the architecture and structure for the implementation.",
				"### Key Activities",
				"1. Create necessary packages and interfaces",
				"2. Define core data structures",
				"3. Establish file organization",
				"4. Set up testing infrastructure",
				"## Change Request Context",
				"This step was executed for change request: Test change request content",
				"Step ID: 01-laying-the-foundation",
				"Step Description: Laying the foundation",
				"Is Test Step: false",
			},
		},
		{
			name:          "successful execution - mvi step",
			changeRequest: "Test change request content",
			step: WorkflowStep{
				ID:          "02-mvi",
				Description: "Minimum Viable Implementation",
				IsTest:      false,
			},
			outputFile:  "output.md",
			wantSuccess: true,
			wantOutputLines: []string{
				"# Minimum Viable Implementation",
				"## Minimum Viable Implementation",
				"This step implements the core functionality with minimal features.",
				"### Implementation Focus",
				"1. Core business logic",
				"2. Essential functionality",
				"3. Basic error handling",
				"4. Minimal user interface",
				"## Change Request Context",
				"This step was executed for change request: Test change request content",
				"Step ID: 02-mvi",
				"Step Description: Minimum Viable Implementation",
				"Is Test Step: false",
			},
		},
		{
			name:          "invalid step ID",
			changeRequest: "Test change request content",
			step: WorkflowStep{
				ID:          "invalid-step",
				Description: "Invalid Step",
				IsTest:      false,
			},
			outputFile:  "output.md",
			wantSuccess: false,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fs := newTestFileSystem()
			io := newTestUserOutput()
			executor := NewStepExecutor(fs, io)

			// Add change request file to mock filesystem
			fs.files["change-request.md"] = []byte(tt.changeRequest)
			fs.exists["change-request.md"] = true

			// Execute
			success, err := executor.ExecuteStep("change-request.md", tt.step, tt.outputFile)

			// Check success/error
			if success != tt.wantSuccess {
				t.Errorf("ExecuteStep() success = %v, want %v", success, tt.wantSuccess)
			}
			if (err != nil) != tt.wantError {
				t.Errorf("ExecuteStep() error = %v, wantError %v", err, tt.wantError)
			}

			// If we expect success, check the output file content
			if tt.wantSuccess {
				output := string(fs.files[tt.outputFile])
				lines := strings.Split(strings.TrimSpace(output), "\n")

				// Check each expected line
				for _, wantLine := range tt.wantOutputLines {
					found := false
					for _, line := range lines {
						if strings.TrimSpace(line) == strings.TrimSpace(wantLine) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected line not found in output: %s", wantLine)
					}
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
			wantErrorText: "failed to read change request file",
		},
		{
			name: "mkdir error",
			setupFS: func(fs *testFileSystem) {
				fs.files["change-request.md"] = []byte("Test content")
				fs.exists["change-request.md"] = true
				fs.mkdirErr = fmt.Errorf("mkdir error")
			},
			wantSuccess:   false,
			wantErrorText: "failed to create output directory",
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
			if err == nil || !strings.Contains(err.Error(), tt.wantErrorText) {
				t.Errorf("ExecuteStep() error = %v, want error containing %q", err, tt.wantErrorText)
			}
		})
	}
} 
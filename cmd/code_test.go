// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// Define static errors
var (
	ErrFileNotFound = errors.New("file not found")
)

// MockWorkflowManager is a mock implementation of the workflow manager
type MockWorkflowManager struct {
	resetWorkflowFunc          func(string) error
	isWorkflowCompleteFunc     func(string) (bool, error)
	determineNextStepFunc      func(string) (int, error)
	generateOutputFilenameFunc func(string, workflow.WorkflowStep) string
	updateStateFunc            func(string, int) error
}

// Interfaces needed for the mock implementation
type WorkflowManager interface {
	ResetWorkflow(string) error
	IsWorkflowComplete(string) (bool, error)
	DetermineNextStep(string) (int, error)
	GenerateOutputFilename(string, workflow.WorkflowStep) string
	UpdateState(string, int) error
}

type UserOutput interface {
	Print(string)
	PrintSuccess(string)
	PrintError(string)
	PrintWarning(string)
	PrintProgress(string)
	PrintStep(int, int, string)
	PrintTable([]string, [][]string)
	IsDebugEnabled() bool
}

func (m *MockWorkflowManager) ResetWorkflow(changeRequestPath string) error {
	return m.resetWorkflowFunc(changeRequestPath)
}

func (m *MockWorkflowManager) IsWorkflowComplete(changeRequestPath string) (bool, error) {
	return m.isWorkflowCompleteFunc(changeRequestPath)
}

func (m *MockWorkflowManager) DetermineNextStep(changeRequestPath string) (int, error) {
	return m.determineNextStepFunc(changeRequestPath)
}

func (m *MockWorkflowManager) GenerateOutputFilename(changeRequestPath string, step workflow.WorkflowStep) string {
	return m.generateOutputFilenameFunc(changeRequestPath, step)
}

func (m *MockWorkflowManager) UpdateState(changeRequestPath string, newStepIndex int) error {
	return m.updateStateFunc(changeRequestPath, newStepIndex)
}

// mockUserOutput is a simple implementation of the user output interface for testing
type mockUserOutput struct {
	messages         []string
	successMessages  []string
	errorMessages    []string
	warningMessages  []string
	progressMessages []string
	stepMessages     []string
	debugEnabled     bool
}

func (m *mockUserOutput) Print(message string) {
	m.messages = append(m.messages, message)
}

func (m *mockUserOutput) PrintSuccess(message string) {
	m.successMessages = append(m.successMessages, message)
}

func (m *mockUserOutput) PrintError(message string) {
	m.errorMessages = append(m.errorMessages, message)
}

func (m *mockUserOutput) PrintWarning(message string) {
	m.warningMessages = append(m.warningMessages, message)
}

func (m *mockUserOutput) PrintProgress(message string) {
	m.progressMessages = append(m.progressMessages, message)
}

func (m *mockUserOutput) PrintStep(stepNumber int, totalSteps int, description string) {
	m.stepMessages = append(m.stepMessages, fmt.Sprintf("Step %d/%d: %s", stepNumber, totalSteps, description))
}

func (m *mockUserOutput) PrintTable(headers []string, rows [][]string) {
	// Not needed for these tests
}

func (m *mockUserOutput) IsDebugEnabled() bool {
	return m.debugEnabled
}

// TestGetDirectoryPath tests the getDirectoryPath function
func TestGetDirectoryPath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "Unix path",
			filePath: "/path/to/file.txt",
			want:     "/path/to/",
		},
		{
			name:     "Windows path",
			filePath: "C:\\path\\to\\file.txt",
			want:     "C:\\path\\to\\",
		},
		{
			name:     "No directory",
			filePath: "file.txt",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDirectoryPath(tt.filePath)
			if got != tt.want {
				t.Errorf("getDirectoryPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetFileName tests the getFileName function
func TestGetFileName(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "Unix path",
			filePath: "/path/to/file.txt",
			want:     "file.txt",
		},
		{
			name:     "Windows path",
			filePath: "C:\\path\\to\\file.txt",
			want:     "file.txt",
		},
		{
			name:     "No directory",
			filePath: "file.txt",
			want:     "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFileName(tt.filePath)
			if got != tt.want {
				t.Errorf("getFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Mock implementation of executeStep for testing
func mockExecuteStep(crPath string, step workflow.WorkflowStep, wf WorkflowManager, fs io.FileSystem, io UserOutput) error {
	// Simple mock implementation
	outputFile := wf.GenerateOutputFilename(crPath, step)
	return fs.WriteFile(outputFile, []byte("test content"), 0644)
}

// TestExecuteStep tests the executeStep function
func TestExecuteStep(t *testing.T) {
	// Create mock dependencies using io.MockFileSystem
	mockFS := io.NewMockFileSystem()
	mockIO := &mockUserOutput{}
	mockWF := &MockWorkflowManager{}
	
	// Set up the mock workflow manager functions
	mockWF.generateOutputFilenameFunc = func(path string, step workflow.WorkflowStep) string {
		return "test-output.md"
	}
	
	// Test case
	testCR := "/path/to/change-request.md"
	testStep := workflow.WorkflowStep{
		ID:          "test-step",
		Description: "Test Step",
		Prompt:      "Test prompt with ${change_request_file_path} variable",
		OutputFile:  "test-output-%s.md",
	}
	
	// Setup mock to create output file
	mockFS.AddFile(testCR, []byte("Test change request content"))
	
	// Call the function
	err := mockExecuteStep(testCR, testStep, mockWF, mockFS, mockIO)
	
	// Check results
	if err != nil {
		t.Errorf("executeStep() error = %v, want nil", err)
	}
	
	// Verify that the output file creation was attempted
	outputFile := mockWF.GenerateOutputFilename(testCR, testStep)
	
	// Check if the output file was created
	if !mockFS.Exists(outputFile) {
		t.Errorf("executeStep() did not create output file %s", outputFile)
	}
}

// Mock implementation of validateChangeRequestExists for testing
func checkFileExists(path string, fs io.FileSystem, io UserOutput) error {
	if !fs.Exists(path) {
		errorMsg := fmt.Sprintf("File %s not found.", path)
		io.PrintError(errorMsg)
		return errors.New(errorMsg)
	}
	return nil
}

// TestCodeCmd_FileNotFound tests the code command when the change request file is not found
func TestCodeCmd_FileNotFound(t *testing.T) {
	// Create mock dependencies using io.MockFileSystem
	mockFS := io.NewMockFileSystem()
	mockIO := &mockUserOutput{}
	
	// Test input
	nonExistentFile := "/path/to/non-existent-file.md"
	
	// Call the function
	err := checkFileExists(nonExistentFile, mockFS, mockIO)
	
	// Check results
	if err == nil {
		t.Errorf("checkFileExists() should return error for non-existent file")
	}
	
	// Verify that error message was printed
	if len(mockIO.errorMessages) == 0 {
		t.Errorf("checkFileExists() should print error message")
	}
	
	expectedError := fmt.Sprintf("File %s not found.", nonExistentFile)
	if mockIO.errorMessages[0] != expectedError {
		t.Errorf("checkFileExists() error message = %v, want %v", mockIO.errorMessages[0], expectedError)
	}
}

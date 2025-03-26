// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// MockFileSystem implements FileSystem interface for testing
type MockFileSystem struct {
	files        map[string][]byte
	existsFunc   func(path string) bool
	readFileFunc func(path string) ([]byte, error)
	writeFileFunc func(path string, data []byte, perm os.FileMode) error
}

// NewMockFileSystem creates a new MockFileSystem instance
func NewMockFileSystem() *MockFileSystem {
	m := &MockFileSystem{
		files: make(map[string][]byte),
	}
	
	m.existsFunc = func(path string) bool {
		_, exists := m.files[path]
		return exists
	}
	
	m.readFileFunc = func(path string) ([]byte, error) {
		data, exists := m.files[path]
		if !exists {
			return nil, errors.New("file not found")
		}
		return data, nil
	}
	
	m.writeFileFunc = func(path string, data []byte, perm os.FileMode) error {
		m.files[path] = data
		return nil
	}
	
	return m
}

// SetExistsFunc sets a custom function for file existence checks
func (m *MockFileSystem) SetExistsFunc(f func(path string) bool) {
	m.existsFunc = f
}

// SetReadFileFunc sets a custom function for file reading
func (m *MockFileSystem) SetReadFileFunc(f func(path string) ([]byte, error)) {
	m.readFileFunc = f
}

// SetWriteFileFunc sets a custom function for file writing
func (m *MockFileSystem) SetWriteFileFunc(f func(path string, data []byte, perm os.FileMode) error) {
	m.writeFileFunc = f
}

// ReadFile implements FileSystem.ReadFile
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	return m.readFileFunc(path)
}

// WriteFile implements FileSystem.WriteFile
func (m *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return m.writeFileFunc(path, data, perm)
}

// Exists implements FileSystem.Exists
func (m *MockFileSystem) Exists(path string) bool {
	return m.existsFunc(path)
}

// MockIO implements UserOutput interface for testing
type MockIO struct {
	messages       []string
	successMessages []string
	errorMessages   []string
}

// NewMockIO creates a new MockIO instance
func NewMockIO() *MockIO {
	return &MockIO{
		messages:        []string{},
		successMessages: []string{},
		errorMessages:   []string{},
	}
}

// Print implements UserOutput.Print
func (m *MockIO) Print(message string) {
	m.messages = append(m.messages, message)
}

// PrintSuccess implements UserOutput.PrintSuccess
func (m *MockIO) PrintSuccess(message string) {
	m.successMessages = append(m.successMessages, message)
}

// PrintError implements UserOutput.PrintError
func (m *MockIO) PrintError(message string) {
	m.errorMessages = append(m.errorMessages, message)
}

func TestGenerateStateFilePath(t *testing.T) {
	tests := []struct {
		name              string
		changeRequestPath string
		want              string
	}{
		{
			name:              "Simple path",
			changeRequestPath: "/path/to/change-request.blueprint.md",
			want:              "/path/to/.change-request.blueprint.md.step",
		},
		{
			name:              "Path with no directory",
			changeRequestPath: "change-request.blueprint.md",
			want:              ".change-request.blueprint.md.step",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateStateFilePath(tt.changeRequestPath)
			if got != tt.want {
				t.Errorf("GenerateStateFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkflowManager_LoadState_NoStateFile(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	
	// Set up mocks
	fs.SetExistsFunc(func(path string) bool {
		return false
	})
	
	// Call the function
	state, err := wm.LoadState(changeRequestPath)
	
	// Check results
	if err != nil {
		t.Errorf("LoadState() error = %v, want nil", err)
	}
	
	// Verify state values
	if state.ChangeRequestPath != changeRequestPath {
		t.Errorf("LoadState() ChangeRequestPath = %v, want %v", state.ChangeRequestPath, changeRequestPath)
	}
	if state.CurrentStepIndex != 0 {
		t.Errorf("LoadState() CurrentStepIndex = %v, want 0", state.CurrentStepIndex)
	}
	if len(state.CompletedSteps) != 0 {
		t.Errorf("LoadState() CompletedSteps = %v, want empty slice", state.CompletedSteps)
	}
}

func TestWorkflowManager_LoadState_WithValidStateFile(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	
	// Create test state
	testState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  2,
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-laying-the-foundation", "01-laying-the-foundation-test"},
	}
	
	// Marshal state to JSON
	stateData, err := json.Marshal(testState)
	if err != nil {
		t.Fatalf("Failed to marshal test state: %v", err)
	}
	
	// Set up mocks
	fs.files[stateFilePath] = stateData
	
	// Call the function
	state, err := wm.LoadState(changeRequestPath)
	
	// Check results
	if err != nil {
		t.Errorf("LoadState() error = %v, want nil", err)
	}
	
	// Verify state values
	if state.ChangeRequestPath != testState.ChangeRequestPath {
		t.Errorf("LoadState() ChangeRequestPath = %v, want %v", state.ChangeRequestPath, testState.ChangeRequestPath)
	}
	if state.CurrentStepIndex != testState.CurrentStepIndex {
		t.Errorf("LoadState() CurrentStepIndex = %v, want %v", state.CurrentStepIndex, testState.CurrentStepIndex)
	}
	if !reflect.DeepEqual(state.CompletedSteps, testState.CompletedSteps) {
		t.Errorf("LoadState() CompletedSteps = %v, want %v", state.CompletedSteps, testState.CompletedSteps)
	}
}

func TestWorkflowManager_LoadState_WithInvalidStateFile(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	
	// Set up mocks with invalid JSON data
	fs.files[stateFilePath] = []byte("invalid json")
	
	// Call the function
	_, err := wm.LoadState(changeRequestPath)
	
	// Check results
	if err == nil {
		t.Errorf("LoadState() error = nil, want an error")
	}
}

func TestWorkflowManager_LoadState_WithInvalidStepIndex(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	
	// Create test state with invalid step index
	testState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  99, // Invalid index
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-laying-the-foundation", "01-laying-the-foundation-test"},
	}
	
	// Marshal state to JSON
	stateData, err := json.Marshal(testState)
	if err != nil {
		t.Fatalf("Failed to marshal test state: %v", err)
	}
	
	// Set up mocks
	fs.files[stateFilePath] = stateData
	
	// Call the function
	state, err := wm.LoadState(changeRequestPath)
	
	// Check results
	if err != nil {
		t.Errorf("LoadState() error = %v, want nil", err)
	}
	
	// Verify state values were reset
	if state.CurrentStepIndex != 0 {
		t.Errorf("LoadState() CurrentStepIndex = %v, want 0", state.CurrentStepIndex)
	}
	if len(state.CompletedSteps) != 0 {
		t.Errorf("LoadState() CompletedSteps = %v, want empty slice", state.CompletedSteps)
	}
	
	// Verify error message was printed
	if len(io.errorMessages) != 1 {
		t.Errorf("LoadState() should print one error message")
	}
}

func TestWorkflowManager_SaveState(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	
	// Create test state
	testState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  2,
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-laying-the-foundation", "01-laying-the-foundation-test"},
	}
	
	// Call the function
	err := wm.SaveState(testState)
	
	// Check results
	if err != nil {
		t.Errorf("SaveState() error = %v, want nil", err)
	}
	
	// Verify file was written
	if _, exists := fs.files[stateFilePath]; !exists {
		t.Errorf("SaveState() did not write to the expected file path")
	}
	
	// Verify file contents
	savedData := fs.files[stateFilePath]
	var savedState WorkflowState
	if err := json.Unmarshal(savedData, &savedState); err != nil {
		t.Errorf("SaveState() wrote invalid JSON: %v", err)
	}
	
	// Compare state values
	if savedState.ChangeRequestPath != testState.ChangeRequestPath {
		t.Errorf("SaveState() ChangeRequestPath = %v, want %v", savedState.ChangeRequestPath, testState.ChangeRequestPath)
	}
	if savedState.CurrentStepIndex != testState.CurrentStepIndex {
		t.Errorf("SaveState() CurrentStepIndex = %v, want %v", savedState.CurrentStepIndex, testState.CurrentStepIndex)
	}
	if !reflect.DeepEqual(savedState.CompletedSteps, testState.CompletedSteps) {
		t.Errorf("SaveState() CompletedSteps = %v, want %v", savedState.CompletedSteps, testState.CompletedSteps)
	}
}

func TestWorkflowManager_DetermineNextStep_NoStateFile(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	
	// Set up mocks
	fs.SetExistsFunc(func(path string) bool {
		return false
	})
	
	// Call the function
	step, err := wm.DetermineNextStep(changeRequestPath)
	
	// Check results
	if err != nil {
		t.Errorf("DetermineNextStep() error = %v, want nil", err)
	}
	if step != 0 {
		t.Errorf("DetermineNextStep() = %v, want 0", step)
	}
}

func TestWorkflowManager_DetermineNextStep_WorkflowComplete(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	
	// Create test state with completed workflow
	testState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  len(StandardWorkflowSteps), // Completed all steps
		LastModified:      time.Now(),
		CompletedSteps:    []string{}, // Not important for this test
	}
	
	// Marshal state to JSON
	stateData, err := json.Marshal(testState)
	if err != nil {
		t.Fatalf("Failed to marshal test state: %v", err)
	}
	
	// Set up mocks
	fs.files[stateFilePath] = stateData
	
	// Call the function
	step, err := wm.DetermineNextStep(changeRequestPath)
	
	// Check results
	if err != nil {
		t.Errorf("DetermineNextStep() error = %v, want nil", err)
	}
	if step != -1 {
		t.Errorf("DetermineNextStep() = %v, want -1", step)
	}
}

func TestWorkflowManager_UpdateState(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	newStepIndex := 3
	
	// Call the function
	err := wm.UpdateState(changeRequestPath, newStepIndex)
	
	// Check results
	if err != nil {
		t.Errorf("UpdateState() error = %v, want nil", err)
	}
	
	// Load the saved state to verify
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	savedData := fs.files[stateFilePath]
	var savedState WorkflowState
	if err := json.Unmarshal(savedData, &savedState); err != nil {
		t.Errorf("UpdateState() wrote invalid JSON: %v", err)
	}
	
	// Verify state values
	if savedState.CurrentStepIndex != newStepIndex {
		t.Errorf("UpdateState() CurrentStepIndex = %v, want %v", savedState.CurrentStepIndex, newStepIndex)
	}
	
	// Verify completed steps
	expectedCompletedSteps := []string{
		StandardWorkflowSteps[0].ID,
		StandardWorkflowSteps[1].ID,
		StandardWorkflowSteps[2].ID,
	}
	if !reflect.DeepEqual(savedState.CompletedSteps, expectedCompletedSteps) {
		t.Errorf("UpdateState() CompletedSteps = %v, want %v", savedState.CompletedSteps, expectedCompletedSteps)
	}
}

func TestWorkflowManager_GenerateOutputFilename(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	step := StandardWorkflowSteps[0]
	
	// Call the function
	filename := wm.GenerateOutputFilename(changeRequestPath, step)
	
	// Define expected result
	expected := filepath.Join("/path/to", "change-request.01-laying-the-foundation.md")
	
	// Check results
	if filename != expected {
		t.Errorf("GenerateOutputFilename() = %v, want %v", filename, expected)
	}
}

func TestWorkflowManager_IsWorkflowComplete(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	
	// Test cases
	tests := []struct {
		name      string
		stepIndex int
		want      bool
	}{
		{
			name:      "Not complete",
			stepIndex: 4,
			want:      false,
		},
		{
			name:      "Complete",
			stepIndex: len(StandardWorkflowSteps),
			want:      true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test state
			testState := WorkflowState{
				ChangeRequestPath: changeRequestPath,
				CurrentStepIndex:  tt.stepIndex,
				LastModified:      time.Now(),
				CompletedSteps:    []string{},
			}
			
			// Marshal state to JSON
			stateData, err := json.Marshal(testState)
			if err != nil {
				t.Fatalf("Failed to marshal test state: %v", err)
			}
			
			// Set up mocks
			fs.files[stateFilePath] = stateData
			
			// Call the function
			got, err := wm.IsWorkflowComplete(changeRequestPath)
			
			// Check results
			if err != nil {
				t.Errorf("IsWorkflowComplete() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("IsWorkflowComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkflowManager_ResetWorkflow(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	
	// Create test state with some progress
	testState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  3,
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-laying-the-foundation", "01-laying-the-foundation-test", "02-mvi"},
	}
	
	// Marshal state to JSON
	stateData, err := json.Marshal(testState)
	if err != nil {
		t.Fatalf("Failed to marshal test state: %v", err)
	}
	
	// Set up mocks
	fs.files[stateFilePath] = stateData
	
	// Call the function
	err = wm.ResetWorkflow(changeRequestPath)
	
	// Check results
	if err != nil {
		t.Errorf("ResetWorkflow() error = %v, want nil", err)
	}
	
	// Load the saved state to verify
	savedData := fs.files[stateFilePath]
	var savedState WorkflowState
	if err := json.Unmarshal(savedData, &savedState); err != nil {
		t.Errorf("ResetWorkflow() wrote invalid JSON: %v", err)
	}
	
	// Verify state values were reset
	if savedState.CurrentStepIndex != 0 {
		t.Errorf("ResetWorkflow() CurrentStepIndex = %v, want 0", savedState.CurrentStepIndex)
	}
	if len(savedState.CompletedSteps) != 0 {
		t.Errorf("ResetWorkflow() CompletedSteps = %v, want empty slice", savedState.CompletedSteps)
	}
} 
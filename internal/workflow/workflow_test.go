// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// MockFileSystem implements FileSystem interface for testing
type MockFileSystem struct {
	files        map[string][]byte
	existsFunc   func(path string) bool
	readFileFunc func(path string) ([]byte, error)
	writeFileFunc func(path string, data []byte, perm os.FileMode) error
	mkdirErr     error
	exists       map[string]bool
}

// NewMockFileSystem creates a new MockFileSystem instance
func NewMockFileSystem() *MockFileSystem {
	m := &MockFileSystem{
		files: make(map[string][]byte),
		exists: make(map[string]bool),
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

// MkdirAll implements FileSystem.MkdirAll
func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirErr != nil {
		return m.mkdirErr
	}
	m.exists[path] = true
	return nil
}

// MockIO implements UserOutput interface for testing
type MockIO struct {
	messages        []string
	successMessages []string
	errorMessages   []string
	warningMessages []string
	progressMessages []string
	stepMessages    []string
	debugEnabled    bool
}

// NewMockIO creates a new MockIO instance
func NewMockIO() *MockIO {
	return &MockIO{
		messages:        []string{},
		successMessages: []string{},
		errorMessages:   []string{},
		warningMessages: []string{},
		progressMessages: []string{},
		stepMessages:    []string{},
		debugEnabled:    false,
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

// PrintWarning implements UserOutput.PrintWarning
func (m *MockIO) PrintWarning(message string) {
	m.warningMessages = append(m.warningMessages, message)
}

// PrintProgress implements UserOutput.PrintProgress
func (m *MockIO) PrintProgress(message string) {
	m.progressMessages = append(m.progressMessages, message)
}

// PrintStep implements UserOutput.PrintStep
func (m *MockIO) PrintStep(stepNumber int, totalSteps int, description string) {
	message := fmt.Sprintf("Step %d/%d: %s", stepNumber, totalSteps, description)
	m.stepMessages = append(m.stepMessages, message)
}

// PrintTable implements UserOutput.PrintTable
func (m *MockIO) PrintTable(headers []string, rows [][]string) {
	// Not needed for these tests
}

// IsDebugEnabled implements UserOutput.IsDebugEnabled
func (m *MockIO) IsDebugEnabled() bool {
	return m.debugEnabled
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
	state, err := wm.LoadState(changeRequestPath)
	
	// Check results - now we expect an error for invalid state file
	if err == nil {
		t.Errorf("LoadState() should return error for invalid state file")
	}
	
	// Verify state values were reset
	if state.CurrentStepIndex != 0 {
		t.Errorf("LoadState() CurrentStepIndex = %v, want 0", state.CurrentStepIndex)
	}
	
	// Verify warning message was printed (if any)
	expectedWarning := fmt.Sprintf(ErrInvalidStateFile, changeRequestPath)
	foundWarning := false
	
	for _, msg := range io.warningMessages {
		if msg == expectedWarning {
			foundWarning = true
			break
		}
	}
	
	if !foundWarning && len(io.warningMessages) > 0 {
		t.Errorf("LoadState() did not print expected warning: %v, got: %v", expectedWarning, io.warningMessages)
	}
	
	// Verify progress message was printed (if any)
	foundProgress := false
	for _, msg := range io.progressMessages {
		if msg == ProgressValidating {
			foundProgress = true
			break
		}
	}
	
	if !foundProgress && len(io.progressMessages) > 0 {
		t.Errorf("LoadState() did not print expected progress: %v, got: %v", ProgressValidating, io.progressMessages)
	}
}

func TestWorkflowManager_LoadState_WithInvalidStepIndex(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Enable debug flag so warnings are printed
	io.debugEnabled = true
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	
	// Create test state with invalid step index
	testState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  99, // Invalid step index
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
	
	// Verify warning message was printed
	if len(io.warningMessages) != 1 {
		t.Errorf("LoadState() should print one warning message")
	} else {
		expectedWarning := fmt.Sprintf(ErrUnrecognizedStep, stateFilePath)
		if io.warningMessages[0] != expectedWarning {
			t.Errorf("LoadState() warning = %v, want %v", io.warningMessages[0], expectedWarning)
		}
	}
}

func TestWorkflowManager_SaveState(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create test state
	state := WorkflowState{
		ChangeRequestPath: "/path/to/change-request.blueprint.md",
		CurrentStepIndex:  2,
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-laying-the-foundation", "01-laying-the-foundation-test"},
	}
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Test successful save
	t.Run("Successful save", func(t *testing.T) {
		// Reset mock
		fs = NewMockFileSystem()
		io = NewMockIO()
		
		// Enable debug mode to print progress messages
		io.debugEnabled = true
		
		wm = NewWorkflowManager(fs, io)
		
		// Call SaveState
		err := wm.SaveState(state)
		
		// Verify results
		if err != nil {
			t.Errorf("SaveState() error = %v, want nil", err)
		}
		
		// Verify that file was written
		stateFilePath := GenerateStateFilePath(state.ChangeRequestPath)
		if _, exists := fs.files[stateFilePath]; !exists {
			t.Errorf("SaveState() didn't write to %s", stateFilePath)
		}
		
		// Verify progress message
		if len(io.progressMessages) == 0 || io.progressMessages[0] != ProgressSavingState {
			t.Errorf("Expected progress message, got %v", io.progressMessages)
		}
	})
	
	// Test write error
	t.Run("Write error", func(t *testing.T) {
		// Reset mock with write error
		fs = NewMockFileSystem()
		io = NewMockIO()
		wm = NewWorkflowManager(fs, io)
		
		// Configure the mock filesystem to return an error when writing
		fs.SetWriteFileFunc(func(path string, data []byte, perm os.FileMode) error {
			return errors.New("write error")
		})
		
		// Call SaveState
		err := wm.SaveState(state)
		
		// Verify results
		if err == nil {
			t.Errorf("SaveState() error = nil, want error")
		}
		
		// Check error message
		expectedError := fmt.Sprintf(ErrStateUpdateFailed, "write error")
		if err.Error() != expectedError {
			t.Errorf("SaveState() error = %v, want %v", err.Error(), expectedError)
		}
	})
}

func TestWorkflowManager_DetermineNextStep_NoStateFile(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Enable debug mode to print step messages
	io.debugEnabled = true
	
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
	
	// Verify step message was printed
	if len(io.stepMessages) != 1 {
		t.Errorf("DetermineNextStep() should print one step message")
	}
	expectedStep := fmt.Sprintf("Step 1/%d: %s", len(StandardWorkflowSteps), StandardWorkflowSteps[0].Description)
	if io.stepMessages[0] != expectedStep {
		t.Errorf("DetermineNextStep() step = %v, want %v", io.stepMessages[0], expectedStep)
	}
}

func TestWorkflowManager_DetermineNextStep_WorkflowComplete(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Enable debug mode to print success messages
	io.debugEnabled = true
	
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
	
	// Verify success message was printed
	if len(io.successMessages) != 1 {
		t.Errorf("DetermineNextStep() should print one success message")
	}
	expectedSuccess := fmt.Sprintf(SuccessWorkflowCompleted, changeRequestPath)
	if io.successMessages[0] != expectedSuccess {
		t.Errorf("DetermineNextStep() success = %v, want %v", io.successMessages[0], expectedSuccess)
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

func TestWorkflowManager_UpdateState_ValidationChecks(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Test negative step index
	t.Run("Negative step index", func(t *testing.T) {
		err := wm.UpdateState("/path/to/change-request.blueprint.md", -1)
		
		if err == nil {
			t.Errorf("UpdateState() with negative index should return error")
		}
		
		expectedErr := fmt.Sprintf(ErrStateUpdateFailed, ErrNegativeStepIndex)
		if err.Error() != expectedErr {
			t.Errorf("UpdateState() error = %v, want %v", err.Error(), expectedErr)
		}
	})
	
	// Test step index exceeding number of steps
	t.Run("Exceeding step index", func(t *testing.T) {
		err := wm.UpdateState("/path/to/change-request.blueprint.md", len(StandardWorkflowSteps) + 1)
		
		if err == nil {
			t.Errorf("UpdateState() with excessive index should return error")
		}
		
		expectedErr := fmt.Sprintf(ErrStateUpdateFailed, ErrExceedingStepIndex)
		if err.Error() != expectedErr {
			t.Errorf("UpdateState() error = %v, want %v", err.Error(), expectedErr)
		}
	})
	
	// Test load state error
	t.Run("Load state error", func(t *testing.T) {
		// Reset mocks
		fs = NewMockFileSystem()
		io = NewMockIO()
		
		// Create workflow manager
		wm = NewWorkflowManager(fs, io)
		
		// Configure the mock filesystem to return an error when reading
		fs.SetReadFileFunc(func(path string) ([]byte, error) {
			return nil, errors.New("read error")
		})
		
		// Set up mock to return that file exists
		fs.SetExistsFunc(func(path string) bool {
			return true
		})
		
		err := wm.UpdateState("/path/to/change-request.blueprint.md", 1)
		
		if err == nil {
			t.Errorf("UpdateState() with load state error should return error")
		}
	})
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

func TestWorkflowManager_IsWorkflowComplete_LoadStateError(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Configure the mock filesystem to return an error when reading
	fs.SetReadFileFunc(func(path string) ([]byte, error) {
		return nil, errors.New("read error")
	})
	
	// Set up mock to return that file exists
	fs.SetExistsFunc(func(path string) bool {
		return true
	})
	
	// Call the function
	complete, err := wm.IsWorkflowComplete("/path/to/change-request.blueprint.md")
	
	// Verify results
	if err == nil {
		t.Errorf("IsWorkflowComplete() should return error when load state fails")
	}
	
	if complete {
		t.Errorf("IsWorkflowComplete() returned true with error, expected false")
	}
}

// TestWorkflowManager_DetermineNextStep_ErrorConditions tests error handling for the DetermineNextStep method
func TestWorkflowManager_DetermineNextStep_ErrorConditions(t *testing.T) {
	t.Run("ReadFile error", func(t *testing.T) {
		// Create mocks
		fs := NewMockFileSystem()
		io := NewMockIO()
		
		// Enable debug mode to print warning messages
		io.debugEnabled = true
		
		// Create workflow manager
		wm := NewWorkflowManager(fs, io)
		
		// Configure the mock filesystem to return an error when reading
		fs.SetReadFileFunc(func(path string) ([]byte, error) {
			return nil, errors.New("read error")
		})
		
		// Set up mock to return that file exists
		fs.SetExistsFunc(func(path string) bool {
			return true
		})
		
		// Call the function
		nextStepIndex, err := wm.DetermineNextStep("/path/to/change-request.blueprint.md")
		
		// Verify results
		if err != nil {
			t.Errorf("DetermineNextStep() returned unexpected error: %v", err)
		}
		
		// Should start from step 0 when there's an error with the state file
		if nextStepIndex != 0 {
			t.Errorf("DetermineNextStep() returned %d, want 0", nextStepIndex)
		}
		
		// Should show a warning
		if len(io.warningMessages) == 0 {
			t.Errorf("DetermineNextStep() should print a warning message")
		}
	})
}

// TestWorkflowManager_ResetWorkflow_Error tests error handling for the ResetWorkflow method
func TestWorkflowManager_ResetWorkflow_Error(t *testing.T) {
	// Create mocks
	fs := NewMockFileSystem()
	io := NewMockIO()
	
	// Create workflow manager
	wm := NewWorkflowManager(fs, io)
	
	// Configure the mock filesystem to return an error when writing
	fs.SetWriteFileFunc(func(path string, data []byte, perm os.FileMode) error {
		return errors.New("write error")
	})
	
	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	
	// Call the function
	err := wm.ResetWorkflow(changeRequestPath)
	
	// Check results
	if err == nil {
		t.Error("ResetWorkflow() should return error when SaveState fails")
	}
	
	// Verify error is from SaveState
	if !strings.Contains(err.Error(), "write error") {
		t.Errorf("ResetWorkflow() error = %v, should contain 'write error'", err)
	}
}

func TestWorkflowManager_ValidateWorkflowSteps(t *testing.T) {
	tests := []struct {
		name         string
		steps        []WorkflowStep
		wantErrorNum int
	}{
		{
			name: "Valid steps",
			steps: []WorkflowStep{
				{
					ID:          "01-test",
					Description: "Test step",
					Prompt:      "Valid prompt with ${change_request_file_path}",
					OutputFile:  "output.md",
				},
			},
			wantErrorNum: 0,
		},
		{
			name: "Missing ID",
			steps: []WorkflowStep{
				{
					Description: "Test step",
					Prompt:      "Valid prompt",
					OutputFile:  "output.md",
				},
			},
			wantErrorNum: 1,
		},
		{
			name: "Missing description",
			steps: []WorkflowStep{
				{
					ID:         "01-test",
					Prompt:     "Valid prompt",
					OutputFile: "output.md",
				},
			},
			wantErrorNum: 1,
		},
		{
			name: "Missing output file",
			steps: []WorkflowStep{
				{
					ID:          "01-test",
					Description: "Test step",
					Prompt:      "Valid prompt",
				},
			},
			wantErrorNum: 1,
		},
		{
			name: "Invalid prompt with malformed variable",
			steps: []WorkflowStep{
				{
					ID:          "01-test",
					Description: "Test step",
					Prompt:      "Invalid prompt with ${var with spaces}",
					OutputFile:  "output.md",
				},
			},
			wantErrorNum: 1,
		},
		{
			name: "Invalid prompt with unclosed variable",
			steps: []WorkflowStep{
				{
					ID:          "01-test",
					Description: "Test step",
					Prompt:      "Invalid prompt with ${unclosed",
					OutputFile:  "output.md",
				},
			},
			wantErrorNum: 1,
		},
		{
			name: "Multiple errors",
			steps: []WorkflowStep{
				{
					ID:          "01-test",
					Description: "",
					Prompt:      "Invalid prompt with ${var with spaces}",
					OutputFile:  "",
				},
			},
			wantErrorNum: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newTestFileSystem()
			io := newTestUserOutput()
			
			wm := NewWorkflowManager(fs, io)
			
			errors := wm.ValidateWorkflowSteps(tt.steps)
			
			if len(errors) != tt.wantErrorNum {
				t.Errorf("ValidateWorkflowSteps() got %d errors, want %d errors", len(errors), tt.wantErrorNum)
				for i, err := range errors {
					t.Logf("Error %d: %v", i, err)
				}
			}
		})
	}
} 
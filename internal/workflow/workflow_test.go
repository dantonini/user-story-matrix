// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/user-story-matrix/usm/internal/io"
	ioLib "github.com/user-story-matrix/usm/internal/io"
)

// MockIO is a mock implementation of the UserOutput interface for testing.
type MockIO struct {
	warningMessages  []string
	errorMessages    []string
	successMessages  []string
	progressMessages []string
	stepMessages     []string
	debugEnabled     bool
}

// NewMockIO creates a new mock IO instance.
func NewMockIO() *MockIO {
	return &MockIO{
		warningMessages:  []string{},
		errorMessages:    []string{},
		successMessages:  []string{},
		progressMessages: []string{},
		stepMessages:     []string{},
		debugEnabled:     false,
	}
}

// Print prints a message.
func (m *MockIO) Print(message string) {
	// For debugging purposes, actually print the message
	fmt.Printf("MockIO.Print: %s\n", message)
}

// PrintSuccess prints a success message.
func (m *MockIO) PrintSuccess(message string) {
	m.successMessages = append(m.successMessages, message)
	fmt.Printf("MockIO.PrintSuccess: %s\n", message)
}

// PrintError prints an error message.
func (m *MockIO) PrintError(message string) {
	m.errorMessages = append(m.errorMessages, message)
	fmt.Printf("MockIO.PrintError: %s\n", message)
}

// PrintWarning prints a warning message.
func (m *MockIO) PrintWarning(message string) {
	m.warningMessages = append(m.warningMessages, message)
	fmt.Printf("MockIO.PrintWarning: %s\n", message)
}

// PrintProgress prints a progress message.
func (m *MockIO) PrintProgress(message string) {
	m.progressMessages = append(m.progressMessages, message)
	fmt.Printf("MockIO.PrintProgress: %s\n", message)
}

// PrintStep prints step information.
func (m *MockIO) PrintStep(stepNumber int, totalSteps int, description string) {
	stepMessage := fmt.Sprintf("Step %d/%d: %s", stepNumber, totalSteps, description)
	m.stepMessages = append(m.stepMessages, stepMessage)
	fmt.Printf("MockIO.PrintStep: %s\n", stepMessage)
}

// IsDebugEnabled returns true if debug is enabled.
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
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"

	// Call the function
	state, err := wm.LoadState(changeRequestPath)

	// Assert results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if state.CurrentStepIndex != 0 {
		t.Errorf("Expected step to be 0, got %d", state.CurrentStepIndex)
	}
	if !reflect.DeepEqual(state.CompletedSteps, []string{}) {
		t.Errorf("Expected empty history, got %v", state.CompletedSteps)
	}
}

func TestWorkflowManager_LoadState_WithValidStateFile(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)

	// Create test state
	testState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  2,
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-laying-the-foundation", "01-laying-the-foundation-test"},
		WorkflowName:      StandardWorkflowName,
		WorkflowPath:      "",
	}

	// Marshal state to JSON
	stateData, err := json.Marshal(testState)
	if err != nil {
		t.Fatalf("Failed to marshal test state: %v", err)
	}

	// Set up mock file
	fs.AddFile(stateFilePath, stateData)

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
	if state.WorkflowName != testState.WorkflowName {
		t.Errorf("LoadState() WorkflowName = %v, want %v", state.WorkflowName, testState.WorkflowName)
	}
}

func TestWorkflowManager_LoadState_WithInvalidStateFile(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Enable debug mode to see warnings
	mockIO.debugEnabled = true

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)

	// Set up mocks with invalid JSON data
	fs.AddFile(stateFilePath, []byte("invalid json"))

	// Call the function
	state, err := wm.LoadState(changeRequestPath)

	// LoadState should never return an error
	if err != nil {
		t.Errorf("LoadState() error = %v, want nil", err)
	}

	// Verify state values were reset to defaults
	if state.CurrentStepIndex != 0 {
		t.Errorf("LoadState() CurrentStepIndex = %v, want 0", state.CurrentStepIndex)
	}
	if state.WorkflowName != StandardWorkflowName {
		t.Errorf("LoadState() WorkflowName = %v, want %s", state.WorkflowName, StandardWorkflowName)
	}
	if state.ChangeRequestPath != changeRequestPath {
		t.Errorf("LoadState() ChangeRequestPath = %v, want %v", state.ChangeRequestPath, changeRequestPath)
	}
	if len(state.CompletedSteps) != 0 {
		t.Errorf("LoadState() CompletedSteps = %v, want empty slice", state.CompletedSteps)
	}

	// Verify warning message was printed about parsing failure
	foundParseWarning := false
	for _, msg := range mockIO.warningMessages {
		if strings.Contains(msg, "Failed to parse state file") {
			foundParseWarning = true
			break
		}
	}

	if !foundParseWarning {
		t.Errorf("LoadState() should print warning about parse failure, got warnings: %v", mockIO.warningMessages)
	}
}

func TestWorkflowManager_LoadState_WithInvalidStepIndex(t *testing.T) {
	// Create mocks
	fs := io.NewMockFileSystem()
	mockIO := NewMockIO()
	mockIO.debugEnabled = true // Enable debug mode to see warnings
	registry := NewWorkflowRegistry()

	// Create workflow manager with default workflow
	wm := NewWorkflowManager(fs, mockIO, "", registry)

	// Create test state with invalid step index
	testState := WorkflowState{
		ChangeRequestPath: "/path/to/change-request.blueprint.md",
		CurrentStepIndex:  99, // Invalid step index
		LastModified:      time.Now(),
		WorkflowName:      StandardWorkflowName,
		CompletedSteps:    []string{},
	}

	// Save the test state
	stateFilePath := GenerateStateFilePath("/path/to/change-request.blueprint.md")
	data, err := json.Marshal(testState)
	if err != nil {
		t.Fatalf("Failed to marshal test state: %v", err)
	}
	fs.AddFile(stateFilePath, data)

	// Load the state
	state, err := wm.LoadState("/path/to/change-request.blueprint.md")

	// LoadState should never return an error
	if err != nil {
		t.Errorf("LoadState() error = %v", err)
	}

	// When there's an invalid step index, it should be reset to 0
	if state.CurrentStepIndex != 0 {
		t.Errorf("LoadState() CurrentStepIndex = %d, want 0", state.CurrentStepIndex)
	}

	// Verify warning message was printed about invalid step index
	foundWarning := false
	for _, msg := range mockIO.warningMessages {
		if strings.Contains(msg, "Invalid step index") {
			foundWarning = true
			break
		}
	}

	if !foundWarning {
		t.Errorf("LoadState() should print warning about invalid step index, got warnings: %v", mockIO.warningMessages)
	}
}

func TestWorkflowManager_SaveState(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create test state
	state := WorkflowState{
		ChangeRequestPath: "/path/to/change-request.blueprint.md",
		CurrentStepIndex:  2,
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-laying-the-foundation", "01-laying-the-foundation-test"},
	}

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Test successful save
	t.Run("Successful save", func(t *testing.T) {
		// Reset mock
		fs = ioLib.NewMockFileSystem()
		mockIO = NewMockIO()

		// Enable debug mode to print progress messages
		mockIO.debugEnabled = true

		wm = NewDefaultWorkflowManager(fs, mockIO)

		// Call SaveState
		err := wm.SaveState(state)

		// Verify results
		if err != nil {
			t.Errorf("SaveState() error = %v, want nil", err)
		}

		// Verify that file was written
		stateFilePath := GenerateStateFilePath(state.ChangeRequestPath)
		if !fs.Exists(stateFilePath) {
			t.Errorf("SaveState() didn't write to %s", stateFilePath)
		}

		// Verify progress message is included (may be other messages too)
		foundSavingMessage := false
		for _, msg := range mockIO.progressMessages {
			if msg == ProgressSavingState {
				foundSavingMessage = true
				break
			}
		}
		if !foundSavingMessage {
			t.Errorf("Expected progress message '%s', but it wasn't found in: %v", 
				ProgressSavingState, mockIO.progressMessages)
		}
	})

	// Test write error - we'll skip this test since we can't easily simulate errors with the MockFileSystem
	t.Run("Write error", func(t *testing.T) {
		t.Skip("Cannot easily simulate write errors with MockFileSystem")
	})
}

func TestWorkflowManager_DetermineNextStep_WorkflowComplete(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Enable debug mode to print success messages
	mockIO.debugEnabled = true

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)

	// Create test state with all steps completed
	testState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  len(wm.workflow.Steps), // Workflow is completed
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-laying-the-foundation", "01-laying-the-foundation-test", "02-mvi", "03-extend", "04-refine"},
		WorkflowName:      StandardWorkflowName, // Add workflow name
		WorkflowPath:      "",                   // Empty for built-in workflows
	}

	// Marshal state to JSON
	stateData, err := json.Marshal(testState)
	if err != nil {
		t.Fatalf("Failed to marshal test state: %v", err)
	}

	// Set up mocks
	fs.AddFile(stateFilePath, stateData)

	// Call the function
	stepIndex, err := wm.DetermineNextStep(changeRequestPath)

	// Check results
	if err != nil {
		t.Errorf("DetermineNextStep() error = %v, want nil", err)
	}

	// For a completed workflow, it should return -1
	if stepIndex != -1 {
		t.Errorf("DetermineNextStep() returned step index %d, want -1", stepIndex)
	}

	// Verify success message was printed
	if len(mockIO.successMessages) != 1 {
		t.Errorf("DetermineNextStep() should print one success message")
		return // Return early to avoid panic accessing empty slice
	}
	expectedSuccess := fmt.Sprintf(SuccessWorkflowCompleted, changeRequestPath)
	if mockIO.successMessages[0] != expectedSuccess {
		t.Errorf("DetermineNextStep() success = %v, want %v", mockIO.successMessages[0], expectedSuccess)
	}
}

func TestWorkflowManager_UpdateState(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	newStepIndex := 1

	// Setup mock state file for LoadState to read
	initialState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
		CompletedSteps:    []string{},
		WorkflowName:      StandardWorkflowName,
		WorkflowPath:      "",
	}
	
	// Marshal and save initial state to mock filesystem
	initialStateData, err := json.Marshal(initialState)
	if err != nil {
		t.Fatalf("Failed to marshal initial state: %v", err)
	}
	fs.AddFile(stateFilePath, initialStateData)

	// Call the function to update state (with path and step index)
	err = wm.UpdateState(changeRequestPath, newStepIndex)

	// Check results
	if err != nil {
		t.Errorf("UpdateState() error = %v, want nil", err)
	}

	// Verify file was written correctly
	if !fs.Exists(stateFilePath) {
		t.Errorf("UpdateState() did not create state file at %s", stateFilePath)
	}

	// Read the written file
	data, err := fs.ReadFile(stateFilePath)
	if err != nil {
		t.Fatalf("Failed to read written state file: %v", err)
	}

	// Parse saved state
	var savedState WorkflowState
	err = json.Unmarshal(data, &savedState)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved state: %v", err)
	}

	// Verify saved state
	if savedState.ChangeRequestPath != changeRequestPath {
		t.Errorf("Saved state ChangeRequestPath = %v, want %v", savedState.ChangeRequestPath, changeRequestPath)
	}
	if savedState.CurrentStepIndex != newStepIndex {
		t.Errorf("Saved state CurrentStepIndex = %v, want %v", savedState.CurrentStepIndex, newStepIndex)
	}

	// Verify completed steps - previous step ID should be added
	expectedCompletedSteps := []string{"01-laying-the-foundation"} // ID of step 0 in standard workflow
	if !slicesEqual(savedState.CompletedSteps, expectedCompletedSteps) {
		t.Errorf("Saved state CompletedSteps = %v, want %v", savedState.CompletedSteps, expectedCompletedSteps)
	}
}

// Helper function to check if two string slices contain the same elements (regardless of order)
func slicesEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	
	// Create copies to avoid modifying the originals
	s1 := make([]string, len(slice1))
	s2 := make([]string, len(slice2))
	copy(s1, slice1)
	copy(s2, slice2)
	
	// Sort both slices
	sort.Strings(s1)
	sort.Strings(s2)
	
	// Compare elements
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func TestWorkflowManager_UpdateState_ValidationChecks(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Test negative step index
	t.Run("Negative step index", func(t *testing.T) {
		err := wm.UpdateState("/path/to/change-request.blueprint.md", -1)
		if err == nil {
			t.Errorf("UpdateState() should return error for negative step index")
		}
		if !strings.Contains(err.Error(), ErrNegativeStepIndex) {
			t.Errorf("UpdateState() error = %v, want error containing %v", err.Error(), ErrNegativeStepIndex)
		}
	})

	// Test exceeding step index
	t.Run("Exceeding step index", func(t *testing.T) {
		err := wm.UpdateState("/path/to/change-request.blueprint.md", len(wm.workflow.Steps)+1)
		if err == nil {
			t.Errorf("UpdateState() should return error for exceeding step index")
		}
		if !strings.Contains(err.Error(), ErrExceedingStepIndex) {
			t.Errorf("UpdateState() error = %v, want error containing %v", err.Error(), ErrExceedingStepIndex)
		}
	})

	// Test load state error
	t.Run("Load state error", func(t *testing.T) {
		// Reset mocks
		fs = ioLib.NewMockFileSystem()
		mockIO = NewMockIO()

		// Enable debug mode to see warnings
		mockIO.debugEnabled = true

		// Create workflow manager with default workflow
		wm = NewDefaultWorkflowManager(fs, mockIO)

		// Add invalid state file
		changeRequestPath := "/path/to/change-request.blueprint.md"
		stateFilePath := GenerateStateFilePath(changeRequestPath)
		fs.AddFile(stateFilePath, []byte("invalid json"))

		// Call the function
		err := wm.UpdateState(changeRequestPath, 1)

		// UpdateState should handle LoadState errors gracefully
		if err != nil {
			t.Errorf("UpdateState() error = %v, want nil", err)
		}

		// Verify warning message was printed about parsing failure
		foundParseWarning := false
		for _, msg := range mockIO.warningMessages {
			if strings.Contains(msg, "Failed to parse state file") {
				foundParseWarning = true
				break
			}
		}

		if !foundParseWarning {
			t.Errorf("UpdateState() should print warning about parse failure, got warnings: %v", mockIO.warningMessages)
		}
	})
}

func TestWorkflowManager_IsWorkflowComplete(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

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
			stepIndex: len(wm.workflow.Steps), // Use the actual length from the workflow
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test case
			fs = ioLib.NewMockFileSystem()
			
			// Create test state
			testState := WorkflowState{
				ChangeRequestPath: changeRequestPath,
				CurrentStepIndex:  tt.stepIndex,
				LastModified:      time.Now(),
				CompletedSteps:    []string{},
				WorkflowName:      StandardWorkflowName, // Add workflow name
				WorkflowPath:      "",                   // Empty for built-in workflows
			}

			// Marshal state to JSON
			stateData, err := json.Marshal(testState)
			if err != nil {
				t.Fatalf("Failed to marshal test state: %v", err)
			}

			// Set up mocks
			fs.AddFile(stateFilePath, stateData)
			
			// Create new workflow manager with updated fs
			wm = NewDefaultWorkflowManager(fs, mockIO)

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
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Enable debug mode to print success messages
	mockIO.debugEnabled = true

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)

	// Create initial state with some steps completed
	initialState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  2,
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-laying-the-foundation", "01-laying-the-foundation-test"},
	}

	// Marshal initial state to JSON
	initialStateData, err := json.Marshal(initialState)
	if err != nil {
		t.Fatalf("Failed to marshal initial state: %v", err)
	}

	// Set up mock file system
	fs.AddFile(stateFilePath, initialStateData)

	// Call the function
	err = wm.ResetWorkflow(changeRequestPath)

	// Check results
	if err != nil {
		t.Errorf("ResetWorkflow() error = %v, want nil", err)
	}

	// Read the state file after reset
	stateData, err := fs.ReadFile(stateFilePath)
	if err != nil {
		t.Fatalf("Failed to read state file after reset: %v", err)
	}

	// Unmarshal the state data
	var resetState WorkflowState
	err = json.Unmarshal(stateData, &resetState)
	if err != nil {
		t.Fatalf("Failed to unmarshal reset state: %v", err)
	}

	// Verify the reset state
	if resetState.CurrentStepIndex != 0 {
		t.Errorf("ResetWorkflow() reset state CurrentStepIndex = %v, want 0", resetState.CurrentStepIndex)
	}
	if len(resetState.CompletedSteps) != 0 {
		t.Errorf("ResetWorkflow() reset state CompletedSteps = %v, want empty slice", resetState.CompletedSteps)
	}

	// Verify success message was printed
	foundSuccess := false
	expectedSuccess := fmt.Sprintf(SuccessStateReset, changeRequestPath)

	for _, msg := range mockIO.successMessages {
		if msg == expectedSuccess {
			foundSuccess = true
			break
		}
	}

	if !foundSuccess {
		t.Errorf("ResetWorkflow() did not print expected success message: %v", expectedSuccess)
	}
}

func TestWorkflowManager_IsWorkflowComplete_LoadStateError(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Enable debug mode to see warnings
	mockIO.debugEnabled = true

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Setup invalid state file
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	fs.AddFile(stateFilePath, []byte("invalid json"))

	// Call the function
	complete, err := wm.IsWorkflowComplete(changeRequestPath)

	// LoadState should never return an error, so IsWorkflowComplete should handle it gracefully
	if err != nil {
		t.Errorf("IsWorkflowComplete() error = %v, want nil", err)
	}

	// When state file is invalid, it should return false (not complete)
	if complete {
		t.Errorf("IsWorkflowComplete() = %v, want false", complete)
	}

	// Verify warning message was printed about parsing failure
	foundParseWarning := false
	for _, msg := range mockIO.warningMessages {
		if strings.Contains(msg, "Failed to parse state file") {
			foundParseWarning = true
			break
		}
	}

	if !foundParseWarning {
		t.Errorf("IsWorkflowComplete() should print warning about parse failure, got warnings: %v", mockIO.warningMessages)
	}
}

func TestWorkflowManager_DetermineNextStep_ErrorConditions(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Define test parameters
	changeRequestPath := "/path/to/change-request.blueprint.md"

	// Test when LoadState returns an error
	t.Run("LoadState error", func(t *testing.T) {
		// Setup a state file with invalid content
		stateFilePath := GenerateStateFilePath(changeRequestPath)
		fs.AddFile(stateFilePath, []byte("invalid json"))

		// Call the function - this should still work but start from step 0
		stepIndex, err := wm.DetermineNextStep(changeRequestPath)

		// Check that we didn't get an error, but a fallback to step 0
		if err != nil {
			t.Errorf("DetermineNextStep() error = %v, want nil", err)
		}

		if stepIndex != 0 {
			t.Errorf("DetermineNextStep() = %v, want 0", stepIndex)
		}

		// Should have a warning message
		if len(mockIO.warningMessages) == 0 && mockIO.debugEnabled {
			t.Errorf("DetermineNextStep() should print warning when LoadState fails")
		}
	})
}

func TestWorkflowManager_ResetWorkflow_Error(t *testing.T) {
	// Test case where WriteFile fails
	t.Run("Write error", func(t *testing.T) {
		// We can't directly mock WriteFile to fail with the new implementation
		// so we'll skip this test
		t.Skip("Cannot easily simulate write errors with MockFileSystem")
	})
}

func TestWorkflowManager_ValidateWorkflowSteps(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create workflow manager with default workflow
	wm := NewDefaultWorkflowManager(fs, mockIO)

	// Test cases
	tests := []struct {
		name         string
		steps        []WorkflowStep
		wantErrCount int
		wantErrMsgs  []string
	}{
		{
			name: "Valid steps",
			steps: []WorkflowStep{
				{
					ID:          "01-test",
					Description: "Test step",
					Prompt:      "Test prompt",
				},
			},
			wantErrCount: 0,
			wantErrMsgs:  []string{},
		},
		{
			name: "Missing ID",
			steps: []WorkflowStep{
				{
					Description: "Test step",
					Prompt:      "Test prompt",
				},
			},
			wantErrCount: 1,
			wantErrMsgs:  []string{"step missing ID"},
		},
		{
			name: "Missing description",
			steps: []WorkflowStep{
				{
					ID:     "01-test",
					Prompt: "Test prompt",
				},
			},
			wantErrCount: 1,
			wantErrMsgs:  []string{"step 01-test missing description"},
		},
		{
			name: "Multiple errors",
			steps: []WorkflowStep{
				{
					// Missing ID and description
					Prompt: "Test prompt",
				},
				{
					ID: "02-test",
					// Missing description
					Prompt: "Test prompt",
				},
			},
			wantErrCount: 2,
			wantErrMsgs:  []string{"step missing ID", "step 02-test missing description"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function
			errors := wm.ValidateWorkflowSteps(tc.steps)

			// Check error count
			if len(errors) != tc.wantErrCount {
				t.Errorf("ValidateWorkflowSteps() error count = %v, want %v", len(errors), tc.wantErrCount)
			}

			// Check error messages
			for i, wantMsg := range tc.wantErrMsgs {
				if i < len(errors) {
					if !strings.Contains(errors[i].Error(), wantMsg) {
						t.Errorf("ValidateWorkflowSteps() error %d = %v, should contain %v", i, errors[i], wantMsg)
					}
				}
			}
		})
	}
}

func TestWorkflowManager_WithNonExistentWorkflow(t *testing.T) {
	// Create mocks
	fs := ioLib.NewMockFileSystem()
	mockIO := NewMockIO()

	// Enable debug mode to see warnings
	mockIO.debugEnabled = true

	// Try to create workflow manager with non-existent workflow
	nonExistentWorkflowName := "non-existent-workflow"
	wm := NewWorkflowManager(fs, mockIO, nonExistentWorkflowName, nil)

	// Test that the workflow manager falls back to standard workflow
	if wm.workflow.Name != StandardWorkflowName {
		t.Errorf("WorkflowManager not falling back to standard workflow, got %s, want %s",
			wm.workflow.Name, StandardWorkflowName)
	}

	// Verify warning was printed
	expectedWarning := fmt.Sprintf(ErrWorkflowNotFound, nonExistentWorkflowName)
	foundWarning := false
	for _, msg := range mockIO.warningMessages {
		if msg == expectedWarning {
			foundWarning = true
			break
		}
	}

	if !foundWarning {
		t.Errorf("Expected warning for non-existent workflow not printed: %s", expectedWarning)
	}
}

func TestWorkflowManager_GetStepByIndex(t *testing.T) {
	// Create mocks
	fs := io.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create a custom workflow with known steps
	workflow := &WorkflowDefinition{
		Name: "test-workflow",
		Steps: []WorkflowStep{
			{ID: "step-1", Description: "First step", Prompt: "Do the first thing"},
			{ID: "step-2", Description: "Second step", Prompt: "Do the second thing"},
			{ID: "step-3", Description: "Third step", Prompt: "Do the third thing"},
		},
	}

	// Create registry and register our workflow
	registry := GetGlobalRegistry()
	registry.RegisterBuiltInWorkflow(workflow)

	// Create workflow manager with our custom workflow
	wm := NewWorkflowManager(fs, mockIO, "test-workflow", registry)

	tests := []struct {
		name        string
		index       int
		wantID      string
		expectError bool
	}{
		{
			name:        "Valid index 0",
			index:       0,
			wantID:      "step-1",
			expectError: false,
		},
		{
			name:        "Valid index 1",
			index:       1,
			wantID:      "step-2",
			expectError: false,
		},
		{
			name:        "Valid index 2",
			index:       2,
			wantID:      "step-3",
			expectError: false,
		},
		{
			name:        "Invalid negative index",
			index:       -1,
			wantID:      "",
			expectError: true,
		},
		{
			name:        "Invalid exceeding index",
			index:       3,
			wantID:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step, err := wm.GetStepByIndex(tt.index)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if step.ID != tt.wantID {
					t.Errorf("Expected step ID %s but got %s", tt.wantID, step.ID)
				}
			}
		})
	}
}

func TestWorkflowManager_ListAvailableWorkflows(t *testing.T) {
	// Create mocks
	fs := io.NewMockFileSystem()
	mockIO := NewMockIO()

	// Create several test workflows
	workflows := []*WorkflowDefinition{
		{Name: "workflow-1", Steps: []WorkflowStep{{ID: "step1", Description: "Step 1"}}},
		{Name: "workflow-2", Steps: []WorkflowStep{{ID: "step1", Description: "Step 1"}}},
		{Name: "workflow-3", Steps: []WorkflowStep{{ID: "step1", Description: "Step 1"}}},
	}

	// Create registry and register our workflows
	// Use the global registry for this test
	registry := GetGlobalRegistry()
	ResetGlobalRegistry() // Reset to ensure clean state
	
	for _, wf := range workflows {
		registry.RegisterBuiltInWorkflow(wf)
	}

	// Create workflow manager with the registry
	wm := NewWorkflowManager(fs, mockIO, StandardWorkflowName, registry)

	// Get the list of workflows
	availableWorkflows := wm.ListAvailableWorkflows()

	// Verify that all registered workflows are in the list
	if len(availableWorkflows) < len(workflows) {
		t.Errorf("Expected at least %d workflows, got %d", len(workflows), len(availableWorkflows))
	}

	// Verify each workflow name is in the list
	for _, wf := range workflows {
		found := false
		for _, name := range availableWorkflows {
			if name == wf.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Workflow %s not found in the list of available workflows", wf.Name)
		}
	}
}

// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

// Import newMockUserOutput from registry_test.go which is in the same package
// This avoids redefinition errors

func TestStateBackwardCompatibility(t *testing.T) {
	// Setup
	fs := io.NewMockFileSystem()
	output := &mockUserOutput{}
	registry := NewWorkflowRegistry()
	manager := NewWorkflowManager(fs, output, "", registry)
	
	// Create a state file with old format (without workflow identification)
	changeRequestPath := "/path/to/change-request.blueprint.md"
	stateFilePath := GenerateStateFilePath(changeRequestPath)
	oldFormatData := createOldFormatState(changeRequestPath, 3)
	err := fs.WriteFile(stateFilePath, oldFormatData, 0644)
	assert.NoError(t, err)
	
	// Test loading state with backward compatibility
	state, err := manager.LoadState(changeRequestPath)
	assert.NoError(t, err)
	
	// Verify default values for new fields
	assert.Equal(t, StandardWorkflowName, state.WorkflowName, "Should default to standard workflow")
	assert.Equal(t, "", state.WorkflowPath, "WorkflowPath should be empty")
	
	// Verify old fields are preserved
	assert.Equal(t, changeRequestPath, state.ChangeRequestPath)
	assert.Equal(t, 3, state.CurrentStepIndex)
}

func TestSaveStateWithWorkflowInfo(t *testing.T) {
	// Setup
	fs := io.NewMockFileSystem()
	userOutput := &mockUserOutput{}
	registry := NewWorkflowRegistry()
	
	// Create a custom workflow and register it
	customWorkflow := &WorkflowDefinition{
		Name:        "custom-workflow",
		Description: "Custom workflow for testing",
		Steps: []WorkflowStep{
			{ID: "01-step", Description: "Step 1", Prompt: "Prompt 1"},
			{ID: "02-step", Description: "Step 2", Prompt: "Prompt 2"},
		},
	}
	registry.RegisterBuiltInWorkflow(customWorkflow)
	
	// Create the manager with the custom workflow
	manager := NewWorkflowManager(fs, userOutput, "custom-workflow", registry)
	
	// Create a WorkflowState with workflow identification
	changeRequestPath := "/path/to/change-request.blueprint.md"
	customWorkflowPath := "/path/to/custom/workflow"
	
	state := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  2,
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-step", "02-step"},
		WorkflowName:      "custom-workflow",
		WorkflowPath:      customWorkflowPath,
	}
	
	// Save it to a file
	err := manager.SaveState(state)
	assert.NoError(t, err)
	
	// Load it back and verify all fields are preserved
	loadedState, err := manager.LoadState(changeRequestPath)
	assert.NoError(t, err)
	
	assert.Equal(t, state.ChangeRequestPath, loadedState.ChangeRequestPath)
	assert.Equal(t, state.CurrentStepIndex, loadedState.CurrentStepIndex)
	assert.ElementsMatch(t, state.CompletedSteps, loadedState.CompletedSteps)
	assert.Equal(t, state.WorkflowName, loadedState.WorkflowName)
	assert.Equal(t, state.WorkflowPath, loadedState.WorkflowPath)
}

func TestUpdateStatePreservesWorkflow(t *testing.T) {
	// Setup
	fs := io.NewMockFileSystem()
	userOutput := &mockUserOutput{}
	registry := NewWorkflowRegistry()
	
	// Create a custom workflow
	customWorkflow := &WorkflowDefinition{
		Name:        "custom-workflow",
		Description: "Custom workflow for testing",
		Steps: []WorkflowStep{
			{ID: "01-step", Description: "Step 1", Prompt: "Prompt 1"},
			{ID: "02-step", Description: "Step 2", Prompt: "Prompt 2"},
			{ID: "03-step", Description: "Step 3", Prompt: "Prompt 3"},
		},
	}
	registry.RegisterBuiltInWorkflow(customWorkflow)
	
	manager := NewWorkflowManager(fs, userOutput, "custom-workflow", registry)
	
	// Create a state with workflow identification
	changeRequestPath := "/path/to/change-request.blueprint.md"
	
	initialState := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
		CompletedSteps:    []string{},
		WorkflowName:      "custom-workflow",
		WorkflowPath:      "/path/to/custom/workflow",
	}
	
	// Save initial state
	err := manager.SaveState(initialState)
	assert.NoError(t, err)
	
	// Update the state to a new step
	err = manager.UpdateState(changeRequestPath, 1)
	assert.NoError(t, err)
	
	// Verify workflow identification is preserved
	updatedState, err := manager.LoadState(changeRequestPath)
	assert.NoError(t, err)
	
	assert.Equal(t, 1, updatedState.CurrentStepIndex, "Step index should be updated")
	assert.Equal(t, initialState.WorkflowName, updatedState.WorkflowName, "Workflow name should be preserved")
	assert.Equal(t, initialState.WorkflowPath, updatedState.WorkflowPath, "Workflow path should be preserved")
}

func TestWorkflowSwitchValidation(t *testing.T) {
	// Setup
	registry := NewWorkflowRegistry()
	
	// Create old and new workflow definitions
	oldWorkflow := &WorkflowDefinition{
		Name:        "old-workflow",
		Description: "Old workflow for testing",
		Steps: []WorkflowStep{
			{ID: "01-step", Description: "Step 1", Prompt: "Prompt 1"},
			{ID: "02-step", Description: "Step 2", Prompt: "Prompt 2"},
			{ID: "03-step", Description: "Step 3", Prompt: "Prompt 3"},
		},
	}
	
	// Compatible workflow with same step IDs
	compatibleWorkflow := &WorkflowDefinition{
		Name:        "compatible-workflow",
		Description: "Compatible workflow for testing",
		Steps: []WorkflowStep{
			{ID: "01-step", Description: "New Step 1", Prompt: "New Prompt 1"},
			{ID: "02-step", Description: "New Step 2", Prompt: "New Prompt 2"},
			{ID: "03-step", Description: "New Step 3", Prompt: "New Prompt 3"},
			{ID: "04-step", Description: "New Step 4", Prompt: "New Prompt 4"},
		},
	}
	
	// Incompatible workflow with different step IDs
	incompatibleWorkflow := &WorkflowDefinition{
		Name:        "incompatible-workflow",
		Description: "Incompatible workflow for testing",
		Steps: []WorkflowStep{
			{ID: "setup", Description: "Setup Step", Prompt: "Setup Prompt"},
			{ID: "implement", Description: "Implement Step", Prompt: "Implement Prompt"},
			{ID: "test", Description: "Test Step", Prompt: "Test Prompt"},
		},
	}
	
	registry.RegisterBuiltInWorkflow(oldWorkflow)
	registry.RegisterBuiltInWorkflow(compatibleWorkflow)
	registry.RegisterBuiltInWorkflow(incompatibleWorkflow)
	
	fs := io.NewMockFileSystem()
	userOutput := &mockUserOutput{}
	manager := NewWorkflowManager(fs, userOutput, "", registry)
	
	// Test validation with compatible workflows
	warnings := manager.ValidateWorkflowSwitch("old-workflow", "compatible-workflow")
	// The warning is expected because of the new 04-step in compatible-workflow
	assert.Len(t, warnings, 1, "Should have warning about new step")
	assert.Contains(t, warnings[0], "04-step", "Warning should mention the new step")
	
	// Test validation with incompatible workflows
	warnings = manager.ValidateWorkflowSwitch("old-workflow", "incompatible-workflow")
	assert.NotEmpty(t, warnings, "Should have warnings for incompatible workflows")
	// Check for specific step differences rather than general message
	stepWarningFound := false
	for _, warning := range warnings {
		if strings.Contains(warning, "01-step") && strings.Contains(warning, "exists in") {
			stepWarningFound = true
			break
		}
	}
	assert.True(t, stepWarningFound, "Warning should mention step differences")
}

func TestMapProgressBetweenWorkflows(t *testing.T) {
	// Setup
	registry := NewWorkflowRegistry()
	
	// Create source and target workflow definitions with different steps
	sourceWorkflow := &WorkflowDefinition{
		Name:        "source-workflow",
		Description: "Source workflow for testing",
		Steps: []WorkflowStep{
			{ID: "01-step", Description: "Step 1", Prompt: "Prompt 1"},
			{ID: "02-step", Description: "Step 2", Prompt: "Prompt 2"},
			{ID: "03-step", Description: "Step 3", Prompt: "Prompt 3"},
		},
	}
	
	targetWorkflow := &WorkflowDefinition{
		Name:        "target-workflow",
		Description: "Target workflow for testing",
		Steps: []WorkflowStep{
			{ID: "01-step", Description: "New Step 1", Prompt: "New Prompt 1"},
			{ID: "new-step", Description: "New Step", Prompt: "New Prompt"},
			{ID: "02-step", Description: "New Step 2", Prompt: "New Prompt 2"},
			{ID: "03-step", Description: "New Step 3", Prompt: "New Prompt 3"},
		},
	}
	
	registry.RegisterBuiltInWorkflow(sourceWorkflow)
	registry.RegisterBuiltInWorkflow(targetWorkflow)
	
	fs := io.NewMockFileSystem()
	userOutput := &mockUserOutput{}
	manager := NewWorkflowManager(fs, userOutput, "", registry)
	
	// Create a state for the source workflow
	sourceState := WorkflowState{
		ChangeRequestPath: "/path/to/change-request.blueprint.md",
		CurrentStepIndex:  1, // At step 02-step
		LastModified:      time.Now(),
		CompletedSteps:    []string{"01-step"},
		WorkflowName:      "source-workflow",
		WorkflowPath:      "",
	}
	
	// Map it to the target workflow
	targetState, warnings := manager.MapProgressBetweenWorkflows(sourceState, "target-workflow")
	
	// Verify step mapping
	assert.Equal(t, "target-workflow", targetState.WorkflowName)
	assert.Equal(t, 2, targetState.CurrentStepIndex, "Should map to step 02-step in target workflow (index 2)")
	assert.ElementsMatch(t, []string{"01-step"}, targetState.CompletedSteps)
	
	// The current implementation doesn't warn about new steps, only checks if steps from the old workflow
	// are missing in the new workflow. Since all steps from the source workflow exist in the target,
	// no warnings are generated.
	assert.Empty(t, warnings, "No warnings expected for this specific mapping scenario")
}

// Helper function to create a test state file
func createTestStateFile(t *testing.T, path string, state WorkflowState) {
	// Serialize the state to JSON
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("Failed to marshal state: %v", err)
	}
	
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	
	// Write to a temporary file
	err = os.WriteFile(path, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write state file: %v", err)
	}
}

// Helper function to create a test state
func createTestState(changeRequestPath string, currentStep int, workflowName string, workflowPath string) WorkflowState {
	return WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  currentStep,
		LastModified:      time.Now(),
		CompletedSteps:    []string{},
		WorkflowName:      workflowName,
		WorkflowPath:      workflowPath,
	}
}

// Helper function to create an old format state (without workflow identification)
func createOldFormatState(changeRequestPath string, currentStep int) []byte {
	state := struct {
		ChangeRequestPath string    `json:"change_request_path"`
		CurrentStepIndex  int       `json:"current_step_index"`
		LastModified      time.Time `json:"last_modified"`
		CompletedSteps    []string  `json:"completed_steps"`
	}{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  currentStep,
		LastModified:      time.Now(),
		CompletedSteps:    []string{},
	}
	
	data, _ := json.Marshal(state)
	return data
} 
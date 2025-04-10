// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
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
	
	// Workflow with different order of steps
	reorderedWorkflow := &WorkflowDefinition{
		Name:        "reordered-workflow",
		Description: "Workflow with reordered steps",
		Steps: []WorkflowStep{
			{ID: "03-step", Description: "Step 3", Prompt: "Prompt 3"},
			{ID: "01-step", Description: "Step 1", Prompt: "Prompt 1"},
			{ID: "02-step", Description: "Step 2", Prompt: "Prompt 2"},
		},
	}
	
	registry.RegisterBuiltInWorkflow(oldWorkflow)
	registry.RegisterBuiltInWorkflow(compatibleWorkflow)
	registry.RegisterBuiltInWorkflow(incompatibleWorkflow)
	registry.RegisterBuiltInWorkflow(reorderedWorkflow)
	
	fs := io.NewMockFileSystem()
	userOutput := &mockUserOutput{}
	manager := NewWorkflowManager(fs, userOutput, "", registry)
	
	// Test validation with compatible workflows
	t.Run("Compatible workflow with new step", func(t *testing.T) {
		warnings := manager.ValidateWorkflowSwitch("old-workflow", "compatible-workflow")
		// The warning is expected because of the new 04-step in compatible-workflow
		assert.Len(t, warnings, 1, "Should have warning about new step")
		assert.Contains(t, warnings[0], "04-step", "Warning should mention the new step")
	})
	
	// Test validation with incompatible workflows
	t.Run("Incompatible workflow with different steps", func(t *testing.T) {
		warnings := manager.ValidateWorkflowSwitch("old-workflow", "incompatible-workflow")
		assert.NotEmpty(t, warnings, "Should have warnings for incompatible workflows")
		
		// Check for specific step differences
		stepWarningsCount := 0
		for _, warning := range warnings {
			if strings.Contains(warning, "exists in") {
				stepWarningsCount++
			}
		}
		assert.True(t, stepWarningsCount >= 3, "Should have at least 3 warnings about step differences")
	})
	
	// Test validation with non-existent old workflow
	t.Run("Non-existent old workflow", func(t *testing.T) {
		warnings := manager.ValidateWorkflowSwitch("non-existent-workflow", "compatible-workflow")
		assert.Len(t, warnings, 1, "Should have one warning about non-existent workflow")
		assert.Contains(t, warnings[0], "Source workflow", "Warning should mention source workflow")
		assert.Contains(t, warnings[0], "not found", "Warning should mention not found")
	})
	
	// Test validation with non-existent new workflow
	t.Run("Non-existent new workflow", func(t *testing.T) {
		warnings := manager.ValidateWorkflowSwitch("old-workflow", "non-existent-workflow")
		assert.Len(t, warnings, 1, "Should have one warning about non-existent workflow")
		assert.Contains(t, warnings[0], "Target workflow", "Warning should mention target workflow")
		assert.Contains(t, warnings[0], "not found", "Warning should mention not found")
	})
	
	// Test validation with reordered steps
	t.Run("Reordered steps", func(t *testing.T) {
		warnings := manager.ValidateWorkflowSwitch("old-workflow", "reordered-workflow")
		
		// Since the steps are the same but in different order, we expect an order warning
		hasOrderWarning := false
		for _, warning := range warnings {
			if strings.Contains(warning, "different order") {
				hasOrderWarning = true
				break
			}
		}
		assert.True(t, hasOrderWarning, "Should have warning about different order")
	})
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
	
	incompatibleWorkflow := &WorkflowDefinition{
		Name:        "incompatible-workflow",
		Description: "Incompatible workflow for testing",
		Steps: []WorkflowStep{
			{ID: "setup", Description: "Setup Step", Prompt: "Setup Prompt"},
			{ID: "implement", Description: "Implement Step", Prompt: "Implement Prompt"},
			{ID: "test", Description: "Test Step", Prompt: "Test Prompt"},
		},
	}
	
	registry.RegisterBuiltInWorkflow(sourceWorkflow)
	registry.RegisterBuiltInWorkflow(targetWorkflow)
	registry.RegisterBuiltInWorkflow(incompatibleWorkflow)
	
	fs := io.NewMockFileSystem()
	userOutput := &mockUserOutput{}
	manager := NewWorkflowManager(fs, userOutput, "", registry)
	
	// Test case 1: Standard mapping with step in common
	t.Run("Standard mapping with step in common", func(t *testing.T) {
		sourceState := WorkflowState{
			ChangeRequestPath: "/path/to/change-request.blueprint.md",
			CurrentStepIndex:  1, // At step 02-step
			LastModified:      time.Now(),
			CompletedSteps:    []string{"01-step"},
			WorkflowName:      "source-workflow",
			WorkflowPath:      "",
		}
		
		targetState, warnings := manager.MapProgressBetweenWorkflows(sourceState, "target-workflow")
		
		// Verify step mapping
		assert.Equal(t, "target-workflow", targetState.WorkflowName)
		assert.Equal(t, 2, targetState.CurrentStepIndex, "Should map to step 02-step in target workflow (index 2)")
		assert.ElementsMatch(t, []string{"01-step"}, targetState.CompletedSteps)
		
		// The current implementation doesn't warn about new steps, only checks if steps from the old workflow
		// are missing in the new workflow. Since all steps from the source workflow exist in the target,
		// no warnings are generated.
		assert.Empty(t, warnings, "No warnings expected for this specific mapping scenario")
	})
	
	// Test case 2: Mapping to incompatible workflow
	t.Run("Mapping to incompatible workflow", func(t *testing.T) {
		sourceState := WorkflowState{
			ChangeRequestPath: "/path/to/change-request.blueprint.md",
			CurrentStepIndex:  1, // At step 02-step
			LastModified:      time.Now(),
			CompletedSteps:    []string{"01-step"},
			WorkflowName:      "source-workflow",
			WorkflowPath:      "",
		}
		
		targetState, warnings := manager.MapProgressBetweenWorkflows(sourceState, "incompatible-workflow")
		
		// Verify state
		assert.Equal(t, "incompatible-workflow", targetState.WorkflowName)
		assert.Equal(t, 0, targetState.CurrentStepIndex, "Should reset to first step when no steps match")
		assert.Empty(t, targetState.CompletedSteps, "No steps should be marked completed")
		
		// Should have warnings about missing steps
		assert.NotEmpty(t, warnings, "Should have warnings for incompatible workflows")
		hasCompletedStepWarning := false
		for _, warning := range warnings {
			if strings.Contains(warning, "Completed step") && strings.Contains(warning, "not found") {
				hasCompletedStepWarning = true
				break
			}
		}
		assert.True(t, hasCompletedStepWarning, "Should warn about completed steps not found in target workflow")
	})
	
	// Test case 3: Mapping with non-existent target workflow
	t.Run("Mapping to non-existent workflow", func(t *testing.T) {
		sourceState := WorkflowState{
			ChangeRequestPath: "/path/to/change-request.blueprint.md",
			CurrentStepIndex:  1,
			LastModified:      time.Now(),
			CompletedSteps:    []string{"01-step"},
			WorkflowName:      "source-workflow",
			WorkflowPath:      "",
		}
		
		targetState, warnings := manager.MapProgressBetweenWorkflows(sourceState, "non-existent-workflow")
		
		// Should return the original state
		assert.Equal(t, sourceState.WorkflowName, targetState.WorkflowName)
		assert.Equal(t, sourceState.CurrentStepIndex, targetState.CurrentStepIndex)
		
		// Should have warning about non-existent workflow
		assert.NotEmpty(t, warnings, "Should have warnings for non-existent workflow")
		hasTargetNotFoundWarning := false
		for _, warning := range warnings {
			if strings.Contains(warning, "not found, keeping current workflow") {
				hasTargetNotFoundWarning = true
				break
			}
		}
		assert.True(t, hasTargetNotFoundWarning, "Should warn about target workflow not found")
	})
	
	// Test case 4: Mapping with non-existent source workflow
	t.Run("Mapping from non-existent workflow", func(t *testing.T) {
		sourceState := WorkflowState{
			ChangeRequestPath: "/path/to/change-request.blueprint.md",
			CurrentStepIndex:  1,
			LastModified:      time.Now(),
			CompletedSteps:    []string{"01-step"},
			WorkflowName:      "non-existent-workflow", // This workflow doesn't exist
			WorkflowPath:      "",
		}
		
		targetState, warnings := manager.MapProgressBetweenWorkflows(sourceState, "target-workflow")
		
		// Should map to target workflow with reset state
		assert.Equal(t, "target-workflow", targetState.WorkflowName)
		
		// Should have warning about non-existent source workflow
		assert.NotEmpty(t, warnings, "Should have warnings for non-existent source workflow")
		hasSourceNotFoundWarning := false
		for _, warning := range warnings {
			if strings.Contains(warning, "Source workflow") && strings.Contains(warning, "not found") {
				hasSourceNotFoundWarning = true
				break
			}
		}
		assert.True(t, hasSourceNotFoundWarning, "Should warn about source workflow not found")
	})
	
	// Test case 5: Mapping with invalid current step index
	t.Run("Mapping with invalid current step index", func(t *testing.T) {
		sourceState := WorkflowState{
			ChangeRequestPath: "/path/to/change-request.blueprint.md",
			CurrentStepIndex:  99, // Out of bounds
			LastModified:      time.Now(),
			CompletedSteps:    []string{"01-step"},
			WorkflowName:      "source-workflow",
			WorkflowPath:      "",
		}
		
		targetState, warnings := manager.MapProgressBetweenWorkflows(sourceState, "target-workflow")
		
		// Should map to target workflow but reset the step index
		assert.Equal(t, "target-workflow", targetState.WorkflowName)
		assert.Equal(t, 0, targetState.CurrentStepIndex, "Should reset to first step when current step index is invalid")
		
		// Should have warning about invalid step index
		assert.NotEmpty(t, warnings, "Should have warnings for invalid step index")
		hasInvalidStepWarning := false
		for _, warning := range warnings {
			if strings.Contains(warning, "Invalid current step index") {
				hasInvalidStepWarning = true
				break
			}
		}
		assert.True(t, hasInvalidStepWarning, "Should warn about invalid current step index")
	})
	
	// Test case 6: Mapping from completed workflow to new workflow
	t.Run("Mapping from completed workflow", func(t *testing.T) {
		sourceState := WorkflowState{
			ChangeRequestPath: "/path/to/change-request.blueprint.md",
			CurrentStepIndex:  3, // Beyond the last step (completed)
			LastModified:      time.Now(),
			CompletedSteps:    []string{"01-step", "02-step", "03-step"},
			WorkflowName:      "source-workflow",
			WorkflowPath:      "",
		}
		
		targetState, warnings := manager.MapProgressBetweenWorkflows(sourceState, "target-workflow")
		
		// Should map to target workflow with the last matching step
		assert.Equal(t, "target-workflow", targetState.WorkflowName)
		// Current implementation should map to a valid step
		assert.GreaterOrEqual(t, targetState.CurrentStepIndex, 0)
		assert.Less(t, targetState.CurrentStepIndex, len(targetWorkflow.Steps))
		
		// Check completed steps are mapped
		for _, step := range targetState.CompletedSteps {
			assert.Contains(t, []string{"01-step", "02-step", "03-step"}, step)
		}
		
		// Should have warning about current step
		assert.NotEmpty(t, warnings, "Should have warnings for mapping from completed workflow")
	})
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

func TestGetStepAtIndex(t *testing.T) {
	// Setup
	registry := NewWorkflowRegistry()
	
	// Create a workflow definition with steps
	workflowDef := &WorkflowDefinition{
		Name:        "test-workflow",
		Description: "Test workflow for GetStepAtIndex",
		Steps: []WorkflowStep{
			{ID: "01-step", Description: "Step 1", Prompt: "Prompt 1"},
			{ID: "02-step", Description: "Step 2", Prompt: "Prompt 2"},
			{ID: "03-step", Description: "Step 3", Prompt: "Prompt 3"},
		},
	}
	
	fs := io.NewMockFileSystem()
	userOutput := &mockUserOutput{}
	manager := NewWorkflowManager(fs, userOutput, "", registry)
	
	// Test valid index
	t.Run("Valid index", func(t *testing.T) {
		step, err := manager.GetStepAtIndex(workflowDef, 1)
		assert.NoError(t, err)
		assert.Equal(t, "02-step", step.ID)
		assert.Equal(t, "Step 2", step.Description)
	})
	
	// Test negative index
	t.Run("Negative index", func(t *testing.T) {
		_, err := manager.GetStepAtIndex(workflowDef, -1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrNegativeStepIndex)
	})
	
	// Test exceeding index
	t.Run("Exceeding index", func(t *testing.T) {
		_, err := manager.GetStepAtIndex(workflowDef, 3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrExceedingStepIndex)
	})
} 
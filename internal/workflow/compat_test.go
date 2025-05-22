// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWorkflowStepFromLegacy(t *testing.T) {
	// Disable legacy warnings to avoid noise in tests
	DisableLegacyWarnings()
	defer EnableLegacyWarnings()
	
	// Test 1: Valid index
	step, err := GetWorkflowStepFromLegacy(0)
	assert.NoError(t, err)
	assert.NotEmpty(t, step.ID)
	assert.NotEmpty(t, step.Description)
	
	// Test 2: Another valid index
	step, err = GetWorkflowStepFromLegacy(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, step.ID)
	assert.NotEmpty(t, step.Description)
	
	// Test 3: Invalid index - negative
	step, err = GetWorkflowStepFromLegacy(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid step index")
	assert.Empty(t, step.ID)
	
	// Test 4: Invalid index - too large
	step, err = GetWorkflowStepFromLegacy(len(UsmCodeStandardWorkflowSteps) + 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid step index")
	assert.Empty(t, step.ID)
}

func TestInitCompatibilityLayer(t *testing.T) {
	// Ensure the registry exists with a standard workflow
	registry := GetGlobalRegistry()
	assert.NotNil(t, registry)
	
	// Call the function to test
	initCompatibilityLayer()
	
	// Verify standard workflow was registered
	standardWorkflow := registry.GetUsmCodeStandardWorkflow()
	assert.NotNil(t, standardWorkflow)
	assert.Equal(t, "standard", standardWorkflow.Name)
	
	// Verify steps were registered
	assert.NotEmpty(t, standardWorkflow.Steps)
}

func TestLegacyWarningSystem(t *testing.T) {
	// Test 1: Warnings initially enabled (default state)
	EnableLegacyWarnings()
	// Reset shown flag
	legacyAccessWarning.shown = false
	
	// Verify initial state
	assert.False(t, legacyAccessWarning.disableWarnings)
	assert.False(t, legacyAccessWarning.shown)
	
	// Call function that uses warning system
	logLegacyAccessWarning()
	
	// Verify warning was logged
	assert.True(t, legacyAccessWarning.shown)
	
	// Test 2: Disable warnings
	DisableLegacyWarnings()
	
	// Verify warnings disabled
	assert.True(t, legacyAccessWarning.disableWarnings)
	
	// Reset shown flag for testing
	legacyAccessWarning.shown = false
	
	// Call warning function again
	logLegacyAccessWarning()
	
	// Verify warning was suppressed (shown flag still false)
	assert.False(t, legacyAccessWarning.shown)
	
	// Test 3: Re-enable warnings
	EnableLegacyWarnings()
	
	// Verify state reset correctly
	assert.False(t, legacyAccessWarning.disableWarnings)
	assert.False(t, legacyAccessWarning.shown)
}

func TestStandardWorkflowStepsConsistency(t *testing.T) {
	// Check that StandardWorkflowSteps and the standard workflow in the registry
	// have the same content
	
	// Get standard workflow from registry
	registry := GetGlobalRegistry()
	standardWorkflow := registry.GetUsmCodeStandardWorkflow()
	
	// Ensure we have a standard workflow
	assert.NotNil(t, standardWorkflow)
	
	// Check if lengths match
	assert.Equal(t, len(UsmCodeStandardWorkflowSteps), len(standardWorkflow.Steps),
		"StandardWorkflowSteps has different number of steps than standard workflow in registry")
	
	// Check each step for consistency
	for i, step := range UsmCodeStandardWorkflowSteps {
		registryStep := standardWorkflow.Steps[i]
		
		// Compare key fields
		assert.Equal(t, step.ID, registryStep.ID, "Step ID mismatch at index %d", i)
		assert.Equal(t, step.Description, registryStep.Description, "Step description mismatch at index %d", i)
		
		// More detailed comparison could include other fields if needed
	}
}

func TestLogLegacyAccessWarning(t *testing.T) {
	// Test when warnings are already shown
	// First, enable warnings and ensure shown flag is false
	EnableLegacyWarnings()
	legacyAccessWarning.shown = false
	
	// First call should set shown to true
	logLegacyAccessWarning()
	assert.True(t, legacyAccessWarning.shown)
	
	// Reset shown for testing purposes
	legacyAccessWarning.shown = true
	
	// Second call should not change anything since shown is already true
	logLegacyAccessWarning()
	assert.True(t, legacyAccessWarning.shown)
	
	// Reset for other tests
	EnableLegacyWarnings()
}

func TestLegacyWarning(t *testing.T) {
	// Reset warning state
	EnableLegacyWarnings()
	
	// Create a buffer to capture warning output
	var buf bytes.Buffer
	SetLegacyWarningWriter(&buf)
	
	// Access via legacy function
	_, err := GetWorkflowStepFromLegacy(0)
	require.NoError(t, err)
	
	// Verify warning was logged
	output := buf.String()
	assert.Contains(t, output, "WARNING: Direct access to StandardWorkflowSteps is deprecated")
	assert.Contains(t, output, "Use WorkflowRegistry.GetWorkflow(\"standard\") instead")
	assert.Contains(t, output, "TestLegacyWarning") // Function name in warning
	
	// Reset buffer
	buf.Reset()
	
	// Access again should not log another warning
	_, err = GetWorkflowStepFromLegacy(0)
	require.NoError(t, err)
	
	// Verify no additional warning
	assert.Empty(t, buf.String())
	
	// Reset warnings and check that warning appears again
	EnableLegacyWarnings()
	
	_, err = GetWorkflowStepFromLegacy(0)
	require.NoError(t, err)
	
	// Verify warning was logged again
	output = buf.String()
	assert.Contains(t, output, "WARNING: Direct access to StandardWorkflowSteps is deprecated")
}

func TestWorkflowCallback(t *testing.T) {
	// Create a clean registry for testing
	oldRegistry := GetGlobalRegistry()
	defer func() {
		// Restore the global registry after test
		globalRegistry = oldRegistry
	}()
	
	registry := ResetGlobalRegistry()
	
	// Prepare test workflow
	testWorkflow := &WorkflowDefinition{
		Name:        "test-workflow",
		Description: "Test workflow",
		Steps: []WorkflowStep{
			{
				ID:          "test-step",
				Description: "Test step",
				Prompt:      "Test prompt",
			},
		},
	}
	
	// Track callback execution
	callbackExecuted := false
	callbackWorkflow := (*WorkflowDefinition)(nil)
	
	// Register callback
	registry.AddWorkflowChangeCallback("test-workflow", func(wf *WorkflowDefinition) {
		callbackExecuted = true
		callbackWorkflow = wf
	})
	
	// Register workflow
	registry.RegisterBuiltInWorkflow(testWorkflow)
	
	// Verify callback was executed
	assert.True(t, callbackExecuted)
	assert.Equal(t, testWorkflow, callbackWorkflow)
}

func TestStandardWorkflowSynchronization(t *testing.T) {
	// Create a clean registry for testing
	oldRegistry := GetGlobalRegistry()
	defer func() {
		// Restore the global registry after test
		globalRegistry = oldRegistry
	}()
	
	registry := ResetGlobalRegistry()
	
	// Store original workflow steps
	originalSteps := make([]WorkflowStep, len(UsmCodeStandardWorkflowSteps))
	copy(originalSteps, UsmCodeStandardWorkflowSteps)
	
	// Modify registry's standard workflow
	modifiedWorkflow := &WorkflowDefinition{
		Name:        UsmCodeStandardWorkflowName,
		Description: "Modified standard workflow",
		Steps: []WorkflowStep{
			{
				ID:          "modified-step",
				Description: "Modified step",
				Prompt:      "Modified prompt",
			},
		},
	}
	
	// Register modified workflow
	registry.RegisterBuiltInWorkflow(modifiedWorkflow)
	
	// Verify StandardWorkflowSteps was updated
	assert.Equal(t, 1, len(UsmCodeStandardWorkflowSteps))
	assert.Equal(t, "modified-step", UsmCodeStandardWorkflowSteps[0].ID)
	
	// Restore original steps
	UsmCodeStandardWorkflowSteps = originalSteps
} 
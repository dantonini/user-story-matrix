// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	step, err = GetWorkflowStepFromLegacy(len(StandardWorkflowSteps) + 10)
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
	standardWorkflow := registry.GetStandardWorkflow()
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
	standardWorkflow := registry.GetStandardWorkflow()
	
	// Ensure we have a standard workflow
	assert.NotNil(t, standardWorkflow)
	
	// Check if lengths match
	assert.Equal(t, len(StandardWorkflowSteps), len(standardWorkflow.Steps),
		"StandardWorkflowSteps has different number of steps than standard workflow in registry")
	
	// Check each step for consistency
	for i, step := range StandardWorkflowSteps {
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
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"testing"
)

func TestGetWorkflowStepFromLegacy(t *testing.T) {
	// Disable legacy warnings to avoid noise in tests
	DisableLegacyWarnings()
	defer EnableLegacyWarnings()
	
	// TODO: This is a placeholder test that will be implemented in MVI phase
	t.Skip("GetWorkflowStepFromLegacy tests will be implemented in MVI phase")
}

func TestInitCompatibilityLayer(t *testing.T) {
	// TODO: Implement tests for initCompatibilityLayer in MVI phase
	t.Skip("initCompatibilityLayer tests will be implemented in MVI phase")
}

func TestLegacyWarningSystem(t *testing.T) {
	// Test enabling/disabling warnings
	DisableLegacyWarnings()
	// TODO: Verify warnings are disabled
	
	EnableLegacyWarnings()
	// TODO: Verify warnings are enabled
	
	// TODO: This is a placeholder test that will be implemented in MVI phase
	t.Skip("Legacy warning system tests will be implemented in MVI phase")
}

func TestStandardWorkflowStepsConsistency(t *testing.T) {
	// Check that StandardWorkflowSteps and the standard workflow in the registry
	// have the same content
	
	// Get standard workflow from registry
	registry := GetGlobalRegistry()
	standardWorkflow := registry.GetStandardWorkflow()
	
	// Check if lengths match
	if len(StandardWorkflowSteps) != len(standardWorkflow.Steps) {
		t.Errorf("StandardWorkflowSteps has %d steps, but standard workflow in registry has %d steps",
			len(StandardWorkflowSteps), len(standardWorkflow.Steps))
	}
	
	// TODO: This test will be expanded in MVI phase
	t.Skip("StandardWorkflowStepsConsistency tests will be expanded in MVI phase")
} 
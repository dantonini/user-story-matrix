// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"runtime"
	"sync"
)

// workflowCompatInit initializes the compatibility layer between
// StandardWorkflowSteps and the workflow registry.
// This function is called during package initialization.
func init() {
	// Ensure registry is created first
	_ = GetGlobalRegistry()
	
	// Initialize the compatibility layer
	initCompatibilityLayer()
}

// legacyAccessWarning holds state to prevent repeated warnings
var legacyAccessWarning struct {
	shown    bool
	mutex    sync.Mutex
	disableWarnings bool
}

// initCompatibilityLayer sets up the compatibility between
// the StandardWorkflowSteps global variable and the registry's standard workflow.
// This allows existing code to continue using StandardWorkflowSteps
// while ensuring consistency with the registry.
func initCompatibilityLayer() {
	// TODO: Implement full two-way synchronization in MVI phase
	
	// For now, just ensure StandardWorkflowSteps is loaded into registry
	registry := GetGlobalRegistry()
	registry.RegisterBuiltInWorkflow(createStandardWorkflow())
}

// GetWorkflowStepFromLegacy provides a compatibility function to access
// workflow steps using the registry while supporting direct indexing into
// StandardWorkflowSteps for legacy code.
//
// Parameters:
//   - index: The 0-based index of the step to retrieve
//
// Returns:
//   - The WorkflowStep at the specified index
//   - Error if the index is invalid
func GetWorkflowStepFromLegacy(index int) (WorkflowStep, error) {
	// Log warning about legacy access (only once per process)
	logLegacyAccessWarning()
	
	// Validate index
	if index < 0 || index >= len(StandardWorkflowSteps) {
		return WorkflowStep{}, fmt.Errorf("invalid step index: %d", index)
	}
	
	// Get step from standard workflow in registry
	workflow := GetGlobalRegistry().GetStandardWorkflow()
	if workflow == nil || index >= len(workflow.Steps) {
		// Fallback to legacy array if registry access fails
		return StandardWorkflowSteps[index], nil
	}
	
	return workflow.Steps[index], nil
}

// logLegacyAccessWarning logs a warning about using the deprecated
// StandardWorkflowSteps. The warning is only shown once per process.
func logLegacyAccessWarning() {
	if legacyAccessWarning.disableWarnings {
		return
	}
	
	legacyAccessWarning.mutex.Lock()
	defer legacyAccessWarning.mutex.Unlock()
	
	if legacyAccessWarning.shown {
		return
	}
	
	// Use runtime.Caller to identify the calling code
	// Note: Variables are intentionally unused in this stub implementation
	// They will be used in the full implementation during MVI phase
	_, _, _, _ = runtime.Caller(2) // Skip this func and GetWorkflowStepFromLegacy
	
	// Log warning about deprecated usage (will be implemented in full in MVI phase)
	// For now, just set the flag to prevent repeated warnings
	legacyAccessWarning.shown = true
}

// DisableLegacyWarnings disables warnings for legacy StandardWorkflowSteps usage
// This is useful for tests or initialization code
func DisableLegacyWarnings() {
	legacyAccessWarning.mutex.Lock()
	defer legacyAccessWarning.mutex.Unlock()
	
	legacyAccessWarning.disableWarnings = true
}

// EnableLegacyWarnings enables warnings for legacy StandardWorkflowSteps usage
func EnableLegacyWarnings() {
	legacyAccessWarning.mutex.Lock()
	defer legacyAccessWarning.mutex.Unlock()
	
	legacyAccessWarning.disableWarnings = false
	legacyAccessWarning.shown = false
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// workflowCompatInit initializes the compatibility layer between
// StandardWorkflowSteps and the workflow registry.
// This function is called during package initialization.
func init() {
	// Ensure registry is created first
	_ = GetGlobalRegistry()
	
	// Initialize the callbacks map
	workflowCallbacks.mutex.Lock()
	workflowCallbacks.callbacks = make(map[string][]workflowChangeCallback)
	workflowCallbacks.mutex.Unlock()
	
	// Initialize the compatibility layer
	initCompatibilityLayer()
}

// legacyAccessWarning holds state to prevent repeated warnings
var legacyAccessWarning struct {
	shown          bool
	mutex          sync.Mutex
	disableWarnings bool
	logWriter       io.Writer
}

// workflowChangeCallback is a function type that handles workflow changes
type workflowChangeCallback func(workflow *WorkflowDefinition)

// workflowCallbacks stores callbacks for workflow changes
var workflowCallbacks struct {
	callbacks map[string][]workflowChangeCallback
	mutex     sync.RWMutex
}

// AddWorkflowChangeCallback registers a callback to be executed when a workflow changes
// This extends the WorkflowRegistry to support synchronization with StandardWorkflowSteps
func (r *WorkflowRegistry) AddWorkflowChangeCallback(workflowName string, callback workflowChangeCallback) {
	workflowCallbacks.mutex.Lock()
	defer workflowCallbacks.mutex.Unlock()
	
	// Initialize the callbacks map if it's nil
	if workflowCallbacks.callbacks == nil {
		workflowCallbacks.callbacks = make(map[string][]workflowChangeCallback)
	}
	
	// Initialize the slice if it doesn't exist
	if _, exists := workflowCallbacks.callbacks[workflowName]; !exists {
		workflowCallbacks.callbacks[workflowName] = make([]workflowChangeCallback, 0)
	}
	
	// Add the callback
	workflowCallbacks.callbacks[workflowName] = append(
		workflowCallbacks.callbacks[workflowName], 
		callback,
	)
}

// notifyWorkflowCallbacks executes all registered callbacks for a workflow
// This should be called whenever a workflow is updated in the registry
func notifyWorkflowCallbacks(workflowName string, workflow *WorkflowDefinition) {
	workflowCallbacks.mutex.RLock()
	defer workflowCallbacks.mutex.RUnlock()
	
	// Check if callbacks map is initialized
	if workflowCallbacks.callbacks == nil {
		return
	}
	
	// Execute all callbacks for this workflow
	if callbacks, exists := workflowCallbacks.callbacks[workflowName]; exists {
		for _, callback := range callbacks {
			callback(workflow)
		}
	}
}

// initCompatibilityLayer sets up the compatibility between
// the StandardWorkflowSteps global variable and the registry's standard workflow.
// This allows existing code to continue using StandardWorkflowSteps
// while ensuring consistency with the registry.
func initCompatibilityLayer() {
	// Set up warning log writer (defaults to stderr)
	legacyAccessWarning.logWriter = os.Stderr
	
	// Set up two-way synchronization between StandardWorkflowSteps and registry
	registry := GetGlobalRegistry()
	
	// Register the standard workflow using the current StandardWorkflowSteps
	workflow := &WorkflowDefinition{
		Name:        StandardWorkflowName,
		Description: "Standard workflow for implementing user stories",
		Steps:       StandardWorkflowSteps,
	}
	
	registry.RegisterBuiltInWorkflow(workflow)
	
	// Set up a hook for when the registry's standard workflow changes
	registry.AddWorkflowChangeCallback(StandardWorkflowName, func(workflow *WorkflowDefinition) {
		// Update StandardWorkflowSteps with the registry's standard workflow
		if workflow != nil && len(workflow.Steps) > 0 {
			// This is a synchronization point to avoid repeatedly logging warnings
			legacyAccessWarning.mutex.Lock()
			StandardWorkflowSteps = workflow.Steps
			legacyAccessWarning.mutex.Unlock()
		}
	})
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
	pc, file, line, ok := runtime.Caller(2) // Skip this func and GetWorkflowStepFromLegacy
	
	// Prepare warning message
	warning := "WARNING: Direct access to StandardWorkflowSteps is deprecated. "
	warning += "Use WorkflowRegistry.GetWorkflow(\"standard\") instead.\n"
	
	// Add caller information if available
	if ok {
		// Get function name from program counter
		fn := runtime.FuncForPC(pc)
		var funcName string
		if fn != nil {
			funcName = fn.Name()
		} else {
			funcName = "unknown_function"
		}
		
		// Format just the filename without the full path
		fileName := filepath.Base(file)
		
		warning += fmt.Sprintf("Called from %s (%s:%d)\n", funcName, fileName, line)
	}
	
	// Log the warning
	fmt.Fprint(legacyAccessWarning.logWriter, warning)
	
	// Set flag to prevent repeated warnings
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

// SetLegacyWarningWriter sets a custom writer for legacy warnings
// This is primarily used for testing to capture warning output
func SetLegacyWarningWriter(writer io.Writer) {
	legacyAccessWarning.mutex.Lock()
	defer legacyAccessWarning.mutex.Unlock()
	
	if writer == nil {
		legacyAccessWarning.logWriter = os.Stderr
	} else {
		legacyAccessWarning.logWriter = writer
	}
} 
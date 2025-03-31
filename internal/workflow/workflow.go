// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WorkflowStep represents a single step in the implementation workflow
type WorkflowStep struct {
	ID          string // Unique identifier (e.g., "01-laying-the-foundation")
	Description string // Human-readable description
	Prompt      string // AI agent instructions with variable interpolation
	OutputFile  string // Template for output filename
}

// WorkflowState tracks the current state of a workflow for a specific change request
type WorkflowState struct {
	ChangeRequestPath string    // Path to the change request file
	CurrentStepIndex  int       // Index of the current step (0-based)
	LastModified      time.Time // When the state was last updated
	CompletedSteps    []string  // List of completed step IDs
}

// WorkflowManager handles workflow-related operations
type WorkflowManager struct {
	fs FileSystem
	io UserOutput
}

// FileSystem defines the file system operations needed by the workflow manager
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Exists(path string) bool
}

// UserOutput defines the interface for displaying output to the user
type UserOutput interface {
	Print(message string)
	PrintSuccess(message string)
	PrintError(message string)
	PrintWarning(message string)
	PrintProgress(message string)
	PrintStep(stepNumber int, totalSteps int, description string)
}

// Error message templates
const (
	ErrFileNotFound            = "❌ Error: File %s not found."
	ErrInvalidStateFile        = "⚠️ Warning: Invalid state file detected for %s. Starting from the beginning."
	ErrStateUpdateFailed       = "❌ Error: Failed to update workflow state: %s"
	ErrStepExecutionFailed     = "❌ Error: Failed to execute step: %s"
	ErrUnrecognizedStep        = "⚠️ Warning: Unrecognized step in %s. Consider resetting the workflow with --reset."
	ErrStateFileCorrupted      = "⚠️ Warning: State file for %s appears to be corrupted. Starting from step 1."
	ErrOutputFileCreateFailed  = "❌ Error: Failed to create output file: %s"
	ErrNegativeStepIndex       = "invalid step index: negative value"
	ErrExceedingStepIndex      = "invalid step index: exceeds number of steps"
	ErrFailedToLoadState       = "failed to load state: %w"
)

// Success message templates
const (
	SuccessStepCompleted     = "✅ Completed step %d of %d: %s"
	SuccessWorkflowCompleted = "🎉 All steps completed successfully for change request: %s"
	SuccessStateReset        = "🔄 Workflow for %s has been reset to the beginning."
)

// Progress message templates
const (
	ProgressExecutingStep = "⏳ Executing step %s: %s"
	ProgressSavingState   = "💾 Saving workflow state..."
	ProgressValidating    = "🔍 Validating workflow state..."
)

// StandardWorkflowSteps defines the predefined sequence of steps in the implementation workflow
var StandardWorkflowSteps = []WorkflowStep{
	{
		ID:          "01-laying-the-foundation",
		Description: "Laying the foundation - Setting up the architecture and structure",
		Prompt:      "You are about to begin a new iteration of software development. Your task is to lay the foundation—that is, to prepare the codebase to safely and effectively accommodate the upcoming changes in the blueprint file at ${change_request_file_path}.",
		OutputFile:  "%s.01-laying-the-foundation.md",
	},
	{
		ID:          "01-laying-the-foundation-test",
		Description: "Laying the foundation testing - Verifying the foundational changes",
		Prompt:      "Review the foundational changes implemented based on the blueprint at ${change_request_file_path}. Verify that the structure is appropriate and tests are in place.",
		OutputFile:  "%s.01-laying-the-foundation-test.md",
	},
	{
		ID:          "02-mvi",
		Description: "Minimum Viable Implementation - Building the core functionality",
		Prompt:      "Implement the core functionality described in the blueprint at ${change_request_file_path}. Focus on meeting the basic acceptance criteria.",
		OutputFile:  "%s.02-mvi.md",
	},
	{
		ID:          "02-mvi-test",
		Description: "Minimum Viable Implementation testing - Verifying the core functionality",
		Prompt:      "Test the minimum viable implementation based on the blueprint at ${change_request_file_path}. Ensure all basic functionality works as expected.",
		OutputFile:  "%s.02-mvi-test.md",
	},
	{
		ID:          "03-extend-functionalities",
		Description: "Extending functionalities - Adding additional features and improvements",
		Prompt:      "Extend the functionality with additional features described in the blueprint at ${change_request_file_path}. Improve the implementation beyond the basic requirements.",
		OutputFile:  "%s.03-extend-functionalities.md",
	},
	{
		ID:          "03-extend-functionalities-test",
		Description: "Extending functionalities testing - Verifying the additional features",
		Prompt:      "Test the extended functionality implemented based on the blueprint at ${change_request_file_path}. Verify all features work correctly.",
		OutputFile:  "%s.03-extend-functionalities-test.md",
	},
	{
		ID:          "04-final-iteration",
		Description: "Final iteration - Polishing and final adjustments",
		Prompt:      "Polish the implementation described in the blueprint at ${change_request_file_path}. Make final adjustments to ensure quality and maintainability.",
		OutputFile:  "%s.04-final-iteration.md",
	},
	{
		ID:          "04-final-iteration-test",
		Description: "Final iteration testing - Final verification and validation",
		Prompt:      "Perform final verification and validation of the implementation based on the blueprint at ${change_request_file_path}. Ensure all requirements are met.",
		OutputFile:  "%s.04-final-iteration-test.md",
	},
}

// NewWorkflowManager creates a new workflow manager instance
func NewWorkflowManager(fs FileSystem, io UserOutput) *WorkflowManager {
	return &WorkflowManager{
		fs: fs,
		io: io,
	}
}

// GenerateStateFilePath generates the path for the state file based on the change request path
func GenerateStateFilePath(changeRequestPath string) string {
	dir := filepath.Dir(changeRequestPath)
	base := filepath.Base(changeRequestPath)
	return filepath.Join(dir, "."+base+".step")
}

// LoadState loads the workflow state from the state file
func (wm *WorkflowManager) LoadState(changeRequestPath string) (WorkflowState, error) {
	state := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
		CompletedSteps:    []string{},
	}

	stateFilePath := GenerateStateFilePath(changeRequestPath)
	if !wm.fs.Exists(stateFilePath) {
		return state, nil
	}

	wm.io.PrintProgress(ProgressValidating)

	data, err := wm.fs.ReadFile(stateFilePath)
	if err != nil {
		wm.io.PrintWarning(fmt.Sprintf(ErrStateFileCorrupted, changeRequestPath))
		return state, err
	}

	if err := json.Unmarshal(data, &state); err != nil {
		wm.io.PrintWarning(fmt.Sprintf(ErrInvalidStateFile, changeRequestPath))
		return state, err
	}

	// Validate the state
	if state.CurrentStepIndex < 0 || state.CurrentStepIndex > len(StandardWorkflowSteps) {
		wm.io.PrintWarning(fmt.Sprintf(ErrUnrecognizedStep, stateFilePath))
		state.CurrentStepIndex = 0
		state.CompletedSteps = []string{}
	}

	return state, nil
}

// SaveState saves the workflow state to the state file
func (wm *WorkflowManager) SaveState(state WorkflowState) error {
	wm.io.PrintProgress(ProgressSavingState)
	
	state.LastModified = time.Now()
	
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf(ErrStateUpdateFailed, err)
	}
	
	stateFilePath := GenerateStateFilePath(state.ChangeRequestPath)
	if err := wm.fs.WriteFile(stateFilePath, data, 0644); err != nil {
		return fmt.Errorf(ErrStateUpdateFailed, err)
	}
	
	return nil
}

// DetermineNextStep determines the next step to execute based on the state
func (wm *WorkflowManager) DetermineNextStep(changeRequestPath string) (int, error) {
	state, err := wm.LoadState(changeRequestPath)
	if err != nil {
		wm.io.PrintWarning(fmt.Sprintf(ErrInvalidStateFile, changeRequestPath))
		return 0, nil // Still start from beginning despite the error
	}
	
	// If we've completed all steps, return a special indicator
	if state.CurrentStepIndex >= len(StandardWorkflowSteps) {
		wm.io.PrintSuccess(fmt.Sprintf(SuccessWorkflowCompleted, changeRequestPath))
		return -1, nil
	}
	
	// Print current step information
	wm.io.PrintStep(state.CurrentStepIndex+1, len(StandardWorkflowSteps), StandardWorkflowSteps[state.CurrentStepIndex].Description)
	
	return state.CurrentStepIndex, nil
}

// UpdateState updates the workflow state after completing a step
func (wm *WorkflowManager) UpdateState(changeRequestPath string, newStepIndex int) error {
	state, err := wm.LoadState(changeRequestPath)
	if err != nil {
		return fmt.Errorf(ErrStateUpdateFailed, err)
	}
	
	// Validate new step index
	if newStepIndex < 0 {
		return fmt.Errorf(ErrStateUpdateFailed, ErrNegativeStepIndex)
	}
	
	if newStepIndex > len(StandardWorkflowSteps) {
		return fmt.Errorf(ErrStateUpdateFailed, ErrExceedingStepIndex)
	}
	
	// Update the state
	state.CurrentStepIndex = newStepIndex
	
	// Update completed steps
	state.CompletedSteps = make([]string, 0, newStepIndex)
	for i := 0; i < newStepIndex; i++ {
		if i < len(StandardWorkflowSteps) {
			state.CompletedSteps = append(state.CompletedSteps, StandardWorkflowSteps[i].ID)
		}
	}
	
	// Print success message for the completed step
	if newStepIndex > 0 && newStepIndex <= len(StandardWorkflowSteps) {
		completedStep := StandardWorkflowSteps[newStepIndex-1]
		wm.io.PrintSuccess(fmt.Sprintf(SuccessStepCompleted, newStepIndex, len(StandardWorkflowSteps), completedStep.Description))
	}
	
	// Save the updated state
	return wm.SaveState(state)
}

// GenerateOutputFilename generates the output filename for a step
func (wm *WorkflowManager) GenerateOutputFilename(changeRequestPath string, step WorkflowStep) string {
	dir := filepath.Dir(changeRequestPath)
	base := filepath.Base(changeRequestPath)
	
	// Remove the .blueprint.md extension if present
	base = strings.TrimSuffix(base, ".blueprint.md")
	
	// Format the output filename using the step's template
	filename := fmt.Sprintf(step.OutputFile, base)
	
	return filepath.Join(dir, filename)
}

// IsWorkflowComplete checks if the workflow is complete
func (wm *WorkflowManager) IsWorkflowComplete(changeRequestPath string) (bool, error) {
	state, err := wm.LoadState(changeRequestPath)
	if err != nil {
		return false, fmt.Errorf(ErrFailedToLoadState, err)
	}
	
	return state.CurrentStepIndex >= len(StandardWorkflowSteps), nil
}

// ResetWorkflow resets the workflow to the beginning
func (wm *WorkflowManager) ResetWorkflow(changeRequestPath string) error {
	state := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
		CompletedSteps:    []string{},
	}
	
	if err := wm.SaveState(state); err != nil {
		return err
	}
	
	wm.io.PrintSuccess(fmt.Sprintf(SuccessStateReset, changeRequestPath))
	return nil
} 
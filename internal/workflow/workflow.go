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
	IsTest      bool   // Whether this is a test step
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
}

// StandardWorkflowSteps defines the predefined sequence of steps in the implementation workflow
var StandardWorkflowSteps = []WorkflowStep{
	{
		ID:          "01-laying-the-foundation",
		Description: "Laying the foundation - Setting up the architecture and structure",
		IsTest:      false,
		OutputFile:  "%s.01-laying-the-foundation.md",
	},
	{
		ID:          "01-laying-the-foundation-test",
		Description: "Laying the foundation testing - Verifying the foundational changes",
		IsTest:      true,
		OutputFile:  "%s.01-laying-the-foundation-test.md",
	},
	{
		ID:          "02-mvi",
		Description: "Minimum Viable Implementation - Building the core functionality",
		IsTest:      false,
		OutputFile:  "%s.02-mvi.md",
	},
	{
		ID:          "02-mvi-test",
		Description: "Minimum Viable Implementation testing - Verifying the core functionality",
		IsTest:      true,
		OutputFile:  "%s.02-mvi-test.md",
	},
	{
		ID:          "03-extend-functionalities",
		Description: "Extending functionalities - Adding additional features and improvements",
		IsTest:      false,
		OutputFile:  "%s.03-extend-functionalities.md",
	},
	{
		ID:          "03-extend-functionalities-test",
		Description: "Extending functionalities testing - Verifying the additional features",
		IsTest:      true,
		OutputFile:  "%s.03-extend-functionalities-test.md",
	},
	{
		ID:          "04-final-iteration",
		Description: "Final iteration - Polishing and final adjustments",
		IsTest:      false,
		OutputFile:  "%s.04-final-iteration.md",
	},
	{
		ID:          "04-final-iteration-test",
		Description: "Final iteration testing - Final verification and validation",
		IsTest:      true,
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

	data, err := wm.fs.ReadFile(stateFilePath)
	if err != nil {
		return state, fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("invalid state file format: %w", err)
	}

	// Validate the state
	if state.CurrentStepIndex < 0 || state.CurrentStepIndex > len(StandardWorkflowSteps) {
		wm.io.PrintError(fmt.Sprintf("Invalid step index %d in state file. Starting from the beginning.", state.CurrentStepIndex))
		state.CurrentStepIndex = 0
		state.CompletedSteps = []string{}
	}

	return state, nil
}

// SaveState saves the workflow state to the state file
func (wm *WorkflowManager) SaveState(state WorkflowState) error {
	state.LastModified = time.Now()
	
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	
	stateFilePath := GenerateStateFilePath(state.ChangeRequestPath)
	if err := wm.fs.WriteFile(stateFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	
	return nil
}

// DetermineNextStep determines the next step to execute based on the state
func (wm *WorkflowManager) DetermineNextStep(changeRequestPath string) (int, error) {
	state, err := wm.LoadState(changeRequestPath)
	if err != nil {
		// If there's an error loading the state, start from the beginning
		wm.io.PrintError(fmt.Sprintf("Error loading workflow state: %s. Starting from the beginning.", err))
		return 0, nil
	}
	
	// If we've completed all steps, return a special indicator
	if state.CurrentStepIndex >= len(StandardWorkflowSteps) {
		return -1, nil
	}
	
	return state.CurrentStepIndex, nil
}

// UpdateState updates the workflow state after completing a step
func (wm *WorkflowManager) UpdateState(changeRequestPath string, newStepIndex int) error {
	state, err := wm.LoadState(changeRequestPath)
	if err != nil {
		return fmt.Errorf("failed to load state for update: %w", err)
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
		return false, fmt.Errorf("failed to load state: %w", err)
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
	
	return wm.SaveState(state)
} 
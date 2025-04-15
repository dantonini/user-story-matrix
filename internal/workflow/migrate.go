// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/user-story-matrix/usm/internal/io"
)

// MigrateStateFile migrates a workflow state file from one workflow to another.
// It loads the existing state, updates the workflow information, and saves it back.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - stateFilePath: Path to the state file
//   - targetWorkflowName: Name of the target workflow
//   - createBackup: Whether to create a backup of the original state file
//
// Returns:
//   - A slice of warning messages (empty if no issues)
//   - Error if the migration fails
func MigrateStateFile(fs io.FileSystem, stateFilePath string, targetWorkflowName string, createBackup bool) ([]string, error) {
	warnings := []string{}

	// Validate arguments
	if fs == nil {
		return nil, fmt.Errorf("filesystem cannot be nil")
	}
	
	if stateFilePath == "" {
		return nil, fmt.Errorf("state file path cannot be empty")
	}
	
	if targetWorkflowName == "" {
		return nil, fmt.Errorf("target workflow name cannot be empty")
	}

	// Check if the file exists
	if !fs.Exists(stateFilePath) {
		return nil, fmt.Errorf("state file not found: %s", stateFilePath)
	}

	// Create backup if requested
	var backupPath string
	var err error
	if createBackup {
		backupPath, err = CreateStateBackup(fs, stateFilePath)
		if err != nil {
			return warnings, fmt.Errorf("failed to create backup before migration: %w", err)
		}
		warnings = append(warnings, fmt.Sprintf("Created backup at %s", backupPath))
	}

	// Read the state file
	content, err := fs.ReadFile(stateFilePath)
	if err != nil {
		return warnings, fmt.Errorf("failed to read state file: %w", err)
	}

	// Check for empty state file
	if len(content) == 0 {
		return warnings, fmt.Errorf("state file is empty: %s", stateFilePath)
	}

	// Unmarshal the state
	var state WorkflowState
	err = json.Unmarshal(content, &state)
	if err != nil {
		return warnings, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Check if the workflow is already set to the target
	if state.WorkflowName == targetWorkflowName {
		warnings = append(warnings, "State file is already using the target workflow")
		return warnings, nil
	}

	// Store the original workflow name for warning message
	originalWorkflow := state.WorkflowName
	if originalWorkflow == "" {
		originalWorkflow = "unknown"
		warnings = append(warnings, "Original workflow name was not specified in the state file")
	}

	// Get the registry to validate workflow names
	registry := GetGlobalRegistry()
	if registry == nil {
		return warnings, fmt.Errorf("workflow registry not initialized")
	}

	// Check if target workflow exists
	targetWorkflow, err := registry.GetWorkflow(targetWorkflowName)
	if err != nil {
		return warnings, fmt.Errorf("target workflow '%s' not found: %w", targetWorkflowName, err)
	}

	// Validate target workflow
	if targetWorkflow == nil || len(targetWorkflow.Steps) == 0 {
		return warnings, fmt.Errorf("invalid target workflow '%s': workflow is empty or nil", targetWorkflowName)
	}

	// Update the state with the new workflow
	state.WorkflowName = targetWorkflowName

	// Attempt to map step progress between workflows
	// This is a best effort to maintain the user's progress
	if originalWorkflow != "unknown" && originalWorkflow != "" {
		sourceWorkflow, err := registry.GetWorkflow(originalWorkflow)
		if err == nil && sourceWorkflow != nil {
			// Create a dummy workflow manager to use its mapping function
			wm := &WorkflowManager{
				fs:       fs,
				registry: registry,
				workflow: sourceWorkflow,
			}

			// Map progress between workflows
			newState, mappingWarnings := wm.MapProgressBetweenWorkflows(state, targetWorkflowName)
			state = newState
			warnings = append(warnings, mappingWarnings...)
		} else {
			warnings = append(warnings, fmt.Sprintf("Could not find original workflow '%s', progress mapping skipped", originalWorkflow))
		}
	}

	// Safety check for step index
	if state.CurrentStepIndex >= len(targetWorkflow.Steps) {
		state.CurrentStepIndex = 0
		warnings = append(warnings, fmt.Sprintf("Reset step index to 0 as the target workflow '%s' has fewer steps", targetWorkflowName))
	}

	// Marshal the updated state
	updatedContent, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return warnings, fmt.Errorf("failed to serialize updated state: %w", err)
	}

	// Write back to the file
	err = fs.WriteFile(stateFilePath, updatedContent, 0644)
	if err != nil {
		return warnings, fmt.Errorf("failed to write updated state: %w", err)
	}

	warnings = append(warnings, fmt.Sprintf("Successfully migrated state file from '%s' to '%s'", originalWorkflow, targetWorkflowName))
	return warnings, nil
}

// AutoMigrateStateFile automatically migrates a state file that doesn't have workflow information.
// This is used to handle legacy state files created before the custom workflow system.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - io: UserOutput interface for user interaction
//   - stateFilePath: Path to the state file
//
// Returns:
//   - Error if the migration fails
func AutoMigrateStateFile(fs io.FileSystem, io UserOutput, stateFilePath string) error {
	// Validate arguments
	if fs == nil {
		return fmt.Errorf("filesystem cannot be nil")
	}
	
	if io == nil {
		return fmt.Errorf("user output interface cannot be nil")
	}
	
	if stateFilePath == "" {
		return fmt.Errorf("state file path cannot be empty")
	}

	// Check if the file exists
	if !fs.Exists(stateFilePath) {
		return fmt.Errorf("state file not found: %s", stateFilePath)
	}

	// Check if this needs migration
	needsMigration, err := needsWorkflowMigration(fs, stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !needsMigration {
		// No migration needed
		return nil
	}

	// Log debug message if debug is enabled
	if io.IsDebugEnabled() {
		io.PrintProgress(fmt.Sprintf("Migrating legacy state file to standard workflow: %s", stateFilePath))
	}

	// Create a backup before migration
	backupPath, err := CreateStateBackup(fs, stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to create backup before migration: %w", err)
	}

	if io.IsDebugEnabled() {
		io.PrintProgress(fmt.Sprintf("Created backup of state file at %s", backupPath))
	}

	// Read the state file
	content, err := fs.ReadFile(stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Check for empty file
	if len(content) == 0 {
		return fmt.Errorf("state file is empty: %s", stateFilePath)
	}

	// Unmarshal the state
	var state WorkflowState
	err = json.Unmarshal(content, &state)
	if err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	// Verify that the standard workflow exists
	registry := GetGlobalRegistry()
	if registry == nil {
		return fmt.Errorf("workflow registry not initialized")
	}
	
	standardWorkflow, err := registry.GetWorkflow(StandardWorkflowName)
	if err != nil || standardWorkflow == nil {
		return fmt.Errorf("standard workflow not found: %w", err)
	}

	// Update the state with the standard workflow
	state.WorkflowName = StandardWorkflowName

	// Safety check for step index
	if state.CurrentStepIndex >= len(standardWorkflow.Steps) {
		state.CurrentStepIndex = 0
		if io.IsDebugEnabled() {
			io.PrintProgress("Reset step index to 0 as it was out of bounds for standard workflow")
		}
	}

	// Marshal the updated state
	updatedContent, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize updated state: %w", err)
	}

	// Write back to the file
	err = fs.WriteFile(stateFilePath, updatedContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated state: %w", err)
	}

	if io.IsDebugEnabled() {
		io.PrintProgress(fmt.Sprintf("Successfully migrated state file to standard workflow: %s", stateFilePath))
	}

	return nil
}

// needsWorkflowMigration checks if a state file needs workflow migration.
// A state file needs migration if it doesn't have workflow information.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - stateFilePath: Path to the state file
//
// Returns:
//   - True if migration is needed, false otherwise
//   - Error if the check fails
func needsWorkflowMigration(fs io.FileSystem, stateFilePath string) (bool, error) {
	// Validate arguments
	if fs == nil {
		return false, fmt.Errorf("filesystem cannot be nil")
	}
	
	if stateFilePath == "" {
		return false, fmt.Errorf("state file path cannot be empty")
	}

	// Read the state file
	content, err := fs.ReadFile(stateFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to read state file: %w", err)
	}

	// Check for empty file
	if len(content) == 0 {
		return false, fmt.Errorf("state file is empty: %s", stateFilePath)
	}

	// Unmarshal the state
	var state WorkflowState
	err = json.Unmarshal(content, &state)
	if err != nil {
		return false, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Check if workflow name is empty
	return state.WorkflowName == "", nil
}

// CreateStateBackup creates a backup of a state file.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - stateFilePath: Path to the state file
//
// Returns:
//   - Path to the backup file
//   - Error if the backup fails
func CreateStateBackup(fs io.FileSystem, stateFilePath string) (string, error) {
	// Validate arguments
	if fs == nil {
		return "", fmt.Errorf("filesystem cannot be nil")
	}
	
	if stateFilePath == "" {
		return "", fmt.Errorf("state file path cannot be empty")
	}
	
	// Check if the source file exists
	if !fs.Exists(stateFilePath) {
		return "", fmt.Errorf("source file not found: %s", stateFilePath)
	}

	// Generate backup path with timestamp
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(stateFilePath)
	basePath := stateFilePath[:len(stateFilePath)-len(ext)]
	backupPath := fmt.Sprintf("%s.backup-%s%s", basePath, timestamp, ext)

	// Read original file
	content, err := fs.ReadFile(stateFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read state file: %w", err)
	}

	// Make sure we're not creating an empty backup
	if len(content) == 0 {
		return "", fmt.Errorf("cannot backup empty file: %s", stateFilePath)
	}

	// Write to backup path
	err = fs.WriteFile(backupPath, content, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
} 
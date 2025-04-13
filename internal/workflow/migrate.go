// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
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
	// TODO: Implement state file migration in MVI phase
	
	// Load existing state
	// This is a stub implementation that will be replaced in the MVI phase
	return []string{"Stub implementation - migration will be implemented in MVI phase"}, nil
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
	
	// TODO: Implement the actual migration in MVI phase
	// This is a stub implementation
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
	// TODO: Implement the actual check in MVI phase
	// This is a stub implementation that always assumes migration is needed
	return true, nil
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
	
	// Write to backup path
	err = fs.WriteFile(backupPath, content, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}
	
	return backupPath, nil
} 
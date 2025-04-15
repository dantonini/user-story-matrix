// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user-story-matrix/usm/internal/io"
)

// TestCreateStateBackup tests the CreateStateBackup function
func TestCreateStateBackup(t *testing.T) {
	// Set up mock filesystem
	fs := io.NewMockFileSystem()

	// Create a test state file
	stateContent := []byte(`{"ChangeRequestPath": "test.md", "CurrentStepIndex": 0}`)
	stateFilePath := "/test/.test.md.step"
	err := fs.WriteFile(stateFilePath, stateContent, 0644)
	require.NoError(t, err)

	// Create backup
	backupPath, err := CreateStateBackup(fs, stateFilePath)
	require.NoError(t, err)

	// Verify backup was created
	assert.True(t, fs.Exists(backupPath))

	// Verify backup content
	backupContent, err := fs.ReadFile(backupPath)
	require.NoError(t, err)
	assert.Equal(t, stateContent, backupContent)

	// Verify backup path format
	assert.Contains(t, backupPath, "/test/.test.md.backup-")
	assert.Contains(t, backupPath, ".step")
}

// TestNeedsWorkflowMigration tests the needsWorkflowMigration function
func TestNeedsWorkflowMigration(t *testing.T) {
	// Set up mock filesystem
	fs := io.NewMockFileSystem()

	// Create a state file with no workflow name
	stateNoWorkflow := WorkflowState{
		ChangeRequestPath: "test.md",
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
	}
	stateNoWorkflowContent, err := json.Marshal(stateNoWorkflow)
	require.NoError(t, err)
	stateNoWorkflowPath := "/test/.test-no-workflow.md.step"
	err = fs.WriteFile(stateNoWorkflowPath, stateNoWorkflowContent, 0644)
	require.NoError(t, err)

	// Create a state file with workflow name
	stateWithWorkflow := WorkflowState{
		ChangeRequestPath: "test.md",
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
		WorkflowName:      "standard",
	}
	stateWithWorkflowContent, err := json.Marshal(stateWithWorkflow)
	require.NoError(t, err)
	stateWithWorkflowPath := "/test/.test-with-workflow.md.step"
	err = fs.WriteFile(stateWithWorkflowPath, stateWithWorkflowContent, 0644)
	require.NoError(t, err)

	// Test state file with no workflow name
	needsMigration, err := needsWorkflowMigration(fs, stateNoWorkflowPath)
	require.NoError(t, err)
	assert.True(t, needsMigration)

	// Test state file with workflow name
	needsMigration, err = needsWorkflowMigration(fs, stateWithWorkflowPath)
	require.NoError(t, err)
	assert.False(t, needsMigration)

	// Test nonexistent file
	_, err = needsWorkflowMigration(fs, "/nonexistent.step")
	assert.Error(t, err)
}

// TestAutoMigrateStateFile tests the AutoMigrateStateFile function
func TestAutoMigrateStateFile(t *testing.T) {
	// Set up mock filesystem
	fs := io.NewMockFileSystem()

	// Set up mock user output
	mockOutput := &MockUserOutput{debugEnabled: true}

	// Create a state file with no workflow name
	stateNoWorkflow := WorkflowState{
		ChangeRequestPath: "test.md",
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
	}
	stateNoWorkflowContent, err := json.Marshal(stateNoWorkflow)
	require.NoError(t, err)
	stateNoWorkflowPath := "/test/.test-no-workflow.md.step"
	err = fs.WriteFile(stateNoWorkflowPath, stateNoWorkflowContent, 0644)
	require.NoError(t, err)

	// Migrate state file
	err = AutoMigrateStateFile(fs, mockOutput, stateNoWorkflowPath)
	require.NoError(t, err)

	// Verify migration
	updatedContent, err := fs.ReadFile(stateNoWorkflowPath)
	require.NoError(t, err)

	var updatedState WorkflowState
	err = json.Unmarshal(updatedContent, &updatedState)
	require.NoError(t, err)

	// Verify workflow name is set
	assert.Equal(t, StandardWorkflowName, updatedState.WorkflowName)

	// Verify debug messages
	assert.Contains(t, mockOutput.GetMessages(), "Migrating legacy state file to standard workflow")
	assert.Contains(t, mockOutput.GetMessages(), "Created backup of state file at")
	assert.Contains(t, mockOutput.GetMessages(), "Successfully migrated state file to standard workflow")
}

// TestMigrateStateFile tests the MigrateStateFile function
func TestMigrateStateFile(t *testing.T) {
	// Set up mock filesystem
	fs := io.NewMockFileSystem()

	// Create standard workflow in registry
	standardWorkflow := &WorkflowDefinition{
		Name:        StandardWorkflowName,
		Description: "Standard workflow",
		Steps: []WorkflowStep{
			{
				ID:          "step1",
				Description: "Step 1",
				Prompt:      "Prompt 1",
			},
			{
				ID:          "step2",
				Description: "Step 2",
				Prompt:      "Prompt 2",
			},
		},
	}

	// Create custom workflow in registry
	customWorkflow := &WorkflowDefinition{
		Name:        "custom",
		Description: "Custom workflow",
		Steps: []WorkflowStep{
			{
				ID:          "custom-step1",
				Description: "Custom Step 1",
				Prompt:      "Custom Prompt 1",
			},
		},
	}

	// Reset registry and register workflows
	registry := ResetGlobalRegistry()
	registry.RegisterBuiltInWorkflow(standardWorkflow)
	registry.RegisterBuiltInWorkflow(customWorkflow)

	// Create a state file with standard workflow
	standardState := WorkflowState{
		ChangeRequestPath: "test.md",
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
		WorkflowName:      StandardWorkflowName,
		CompletedSteps:    []string{"step1"},
	}
	standardStateContent, err := json.Marshal(standardState)
	require.NoError(t, err)
	standardStatePath := "/test/.test-standard.md.step"
	err = fs.WriteFile(standardStatePath, standardStateContent, 0644)
	require.NoError(t, err)

	// Create a state file with no workflow name
	noWorkflowState := WorkflowState{
		ChangeRequestPath: "test.md",
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
	}
	noWorkflowStateContent, err := json.Marshal(noWorkflowState)
	require.NoError(t, err)
	noWorkflowStatePath := "/test/.test-no-workflow.md.step"
	err = fs.WriteFile(noWorkflowStatePath, noWorkflowStateContent, 0644)
	require.NoError(t, err)

	// Test migrating from standard to custom
	warnings, err := MigrateStateFile(fs, standardStatePath, "custom", true)
	require.NoError(t, err)

	// Verify migration
	updatedContent, err := fs.ReadFile(standardStatePath)
	require.NoError(t, err)

	var updatedState WorkflowState
	err = json.Unmarshal(updatedContent, &updatedState)
	require.NoError(t, err)

	// Verify workflow name is updated
	assert.Equal(t, "custom", updatedState.WorkflowName)

	// Verify warnings
	assert.Contains(t, warnings[0], "Created backup at")
	assert.Contains(t, warnings[len(warnings)-1], "Successfully migrated state file from 'standard' to 'custom'")

	// Test migrating from no workflow to custom
	warnings, err = MigrateStateFile(fs, noWorkflowStatePath, "custom", false)
	require.NoError(t, err)

	// Verify migration
	updatedContent, err = fs.ReadFile(noWorkflowStatePath)
	require.NoError(t, err)

	err = json.Unmarshal(updatedContent, &updatedState)
	require.NoError(t, err)

	// Verify workflow name is updated
	assert.Equal(t, "custom", updatedState.WorkflowName)

	// Verify warnings
	assert.Contains(t, warnings, "Original workflow name was not specified in the state file")
	assert.Contains(t, warnings, "Successfully migrated state file from 'unknown' to 'custom'")

	// Test migrating to non-existent workflow
	_, err = MigrateStateFile(fs, standardStatePath, "nonexistent", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target workflow 'nonexistent' not found")
}

// MockUserOutput is a mock implementation of UserOutput for testing
type MockUserOutput struct {
	messages      []string
	debugEnabled bool
}

// Print implements UserOutput.Print
func (m *MockUserOutput) Print(message string) {
	m.messages = append(m.messages, message)
}

// PrintSuccess implements UserOutput.PrintSuccess
func (m *MockUserOutput) PrintSuccess(message string) {
	m.messages = append(m.messages, "SUCCESS: "+message)
}

// PrintError implements UserOutput.PrintError
func (m *MockUserOutput) PrintError(message string) {
	m.messages = append(m.messages, "ERROR: "+message)
}

// PrintWarning implements UserOutput.PrintWarning
func (m *MockUserOutput) PrintWarning(message string) {
	m.messages = append(m.messages, "WARNING: "+message)
}

// PrintProgress implements UserOutput.PrintProgress
func (m *MockUserOutput) PrintProgress(message string) {
	m.messages = append(m.messages, "PROGRESS: "+message)
}

// PrintStep implements UserOutput.PrintStep
func (m *MockUserOutput) PrintStep(stepNumber int, totalSteps int, description string) {
	m.messages = append(m.messages, 
		"STEP: "+description+" ("+string(rune('0'+stepNumber))+"/"+string(rune('0'+totalSteps))+")",
	)
}

// IsDebugEnabled implements UserOutput.IsDebugEnabled
func (m *MockUserOutput) IsDebugEnabled() bool {
	return m.debugEnabled
}

// GetMessages returns all messages logged to this mock
func (m *MockUserOutput) GetMessages() string {
	result := ""
	for _, msg := range m.messages {
		result += msg + "\n"
	}
	return result
} 
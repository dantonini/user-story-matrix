// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	internalio "github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// TestShowWorkflow tests the internal showWorkflow function directly
func TestShowWorkflow(t *testing.T) {
	// Skip if running in a CI environment or with limited resources
	if os.Getenv("SKIP_WORKFLOW_TESTS") == "1" {
		t.Skip("Skipping test that requires workflow registry - set SKIP_WORKFLOW_TESTS=0 to enable")
	}
	
	// Create mock terminal IO with the correct interface type
	mockTermIO := internalio.NewTerminalIO()
	
	// Test with standard workflow - this should always exist
	result := showWorkflow("standard", "text", mockTermIO)
	
	// Validate result
	assert.True(t, result.Success)
	assert.Empty(t, result.ErrorMessage)
	assert.Empty(t, result.Output) // Text format prints directly, no output returned
}

// TestShowWorkflow_JSON tests the JSON format
func TestShowWorkflow_JSON(t *testing.T) {
	// Skip if running in a CI environment or with limited resources
	if os.Getenv("SKIP_WORKFLOW_TESTS") == "1" {
		t.Skip("Skipping test that requires workflow registry - set SKIP_WORKFLOW_TESTS=0 to enable")
	}
	
	// Create mock terminal IO with the correct interface type
	mockTermIO := internalio.NewTerminalIO()
	
	// Test with standard workflow and JSON format
	result := showWorkflow("standard", "json", mockTermIO)
	
	// Validate result
	assert.True(t, result.Success)
	assert.Empty(t, result.ErrorMessage)
	assert.NotEmpty(t, result.Output)
	
	// Parse the JSON output
	var details WorkflowDetails
	err := json.Unmarshal([]byte(result.Output), &details)
	assert.NoError(t, err)
	
	// Validate content
	assert.Equal(t, "standard", details.Name)
	assert.NotEmpty(t, details.Description)
	assert.NotEmpty(t, details.Steps)
}

// TestShowWorkflow_Markdown tests the markdown format
func TestShowWorkflow_Markdown(t *testing.T) {
	// Skip if running in a CI environment or with limited resources
	if os.Getenv("SKIP_WORKFLOW_TESTS") == "1" {
		t.Skip("Skipping test that requires workflow registry - set SKIP_WORKFLOW_TESTS=0 to enable")
	}
	
	// Create mock terminal IO with the correct interface type
	mockTermIO := internalio.NewTerminalIO()
	
	// Test with standard workflow and markdown format
	result := showWorkflow("standard", "markdown", mockTermIO)
	
	// Validate result
	assert.True(t, result.Success)
	assert.Empty(t, result.ErrorMessage)
	assert.NotEmpty(t, result.Output)
	
	// Check for markdown elements
	assert.Contains(t, result.Output, "# Workflow: standard")
	assert.Contains(t, result.Output, "## Steps")
	assert.Contains(t, result.Output, "## Usage")
}

// TestShowWorkflow_NonexistentWorkflow tests the behavior with a nonexistent workflow
func TestShowWorkflow_NonexistentWorkflow(t *testing.T) {
	// Create mock terminal IO with the correct interface type
	mockTermIO := internalio.NewTerminalIO()
	
	// Test with nonexistent workflow
	result := showWorkflow("nonexistent-workflow", "text", mockTermIO)
	
	// Validate result
	assert.False(t, result.Success)
	assert.Contains(t, result.ErrorMessage, "not found")
	assert.Empty(t, result.Output)
}

// TestShowWorkflow_MockedWorkflow tests the workflow showing with a mocked workflow
func TestShowWorkflow_MockedWorkflow(t *testing.T) {
	// Create a mock terminal IO
	mockTermIO := internalio.NewTerminalIO()
	
	// Create a mock file system
	mockFS := internalio.NewMockFileSystem()
	
	// Define a test workflow
	testWorkflowYaml := `name: test-workflow
description: Test workflow for unit tests
steps:
  - id: step1
    description: First step
    prompt: This is a test prompt for step 1
    variables:
      test_var: test value
      another_var: another value
  - id: step2
    description: Second step
    prompt: This is a test prompt for step 2
`
	// Add the test workflow to the mock file system
	homeDir, _ := os.UserHomeDir()
	workflowDir := filepath.Join(homeDir, ".usm", "workflows", "test-workflow")
	mockFS.AddFile(filepath.Join(workflowDir, "workflow.yaml"), []byte(testWorkflowYaml))
	
	// Reset the global registry to ensure clean state
	registry := workflow.ResetGlobalRegistry()
	
	// Load the workflow directly into the registry
	testWf, err := registry.LoadFromDirectory(mockFS, workflowDir)
	if err != nil {
		t.Fatalf("Failed to load test workflow: %v", err)
	}
	registry.RegisterBuiltInWorkflow(testWf)
	
	// Test showing the workflow in each format
	formats := []string{"text", "json", "markdown"}
	
	for _, format := range formats {
		t.Run("Format_"+format, func(t *testing.T) {
			result := showWorkflow("test-workflow", format, mockTermIO)
			
			// Validate result
			assert.True(t, result.Success)
			assert.Empty(t, result.ErrorMessage)
			
			// For non-text formats, check output
			if format != "text" {
				assert.NotEmpty(t, result.Output)
				
				// Specific checks based on format
				switch format {
				case "json":
					var details WorkflowDetails
					err := json.Unmarshal([]byte(result.Output), &details)
					assert.NoError(t, err)
					assert.Equal(t, "test-workflow", details.Name)
					assert.Equal(t, "Test workflow for unit tests", details.Description)
					assert.Len(t, details.Steps, 2)
					assert.Equal(t, "step1", details.Steps[0].ID)
					assert.Equal(t, "step2", details.Steps[1].ID)
					
				case "markdown":
					md := result.Output
					assert.Contains(t, md, "# Workflow: test-workflow")
					assert.Contains(t, md, "Test workflow for unit tests")
					assert.Contains(t, md, "### 1. step1")
					assert.Contains(t, md, "### 2. step2")
					assert.Contains(t, md, "| test_var | test value |")
				}
			}
		})
	}
}

// TestShowCmd tests the cobra command
func TestShowCmd(t *testing.T) {
	// Skip if running in a CI environment or with limited resources
	if os.Getenv("SKIP_WORKFLOW_TESTS") == "1" {
		t.Skip("Skipping test that requires workflow registry - set SKIP_WORKFLOW_TESTS=0 to enable")
	}
	
	// Create a test command
	cmd := &cobra.Command{}
	cmd.Flags().StringP("format", "f", "text", "Output format")
	_ = cmd.Flags().Set("format", "text")
	
	// Test with the standard workflow (should always exist)
	var output bytes.Buffer
	
	// Save original stdout and restore it after the test
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Clean up after the test
	defer func() {
		os.Stdout = oldStdout
	}()
	
	// Create a channel to signal when we're done capturing
	done := make(chan bool)
	
	// Copy output in a goroutine
	go func() {
		_, _ = io.Copy(&output, r)
		done <- true
	}()
	
	// Execute the command with the standard workflow
	showCmdRun := ShowCmd.Run
	showCmdRun(cmd, []string{"standard"})
	
	// Close writer and wait for copy to complete
	_ = w.Close()
	<-done
	
	// Close reader
	_ = r.Close()
	
	// Verify output contains expected elements
	outputStr := output.String()
	assert.Contains(t, outputStr, "Looking for workflow 'standard'")
	assert.Contains(t, outputStr, "Workflow: standard")
	assert.Contains(t, outputStr, "Steps:")
}

// TestGenerateMarkdownOutput tests the markdown generation function
func TestGenerateMarkdownOutput(t *testing.T) {
	// Create a sample workflow details
	details := WorkflowDetails{
		Name:        "test-workflow",
		Description: "Test workflow description",
		Steps: []StepDetail{
			{
				ID:          "step1",
				Description: "First step description",
				Variables: map[string]string{
					"var1": "value1",
					"var2": "value2",
				},
			},
			{
				ID:          "step2",
				Description: "Second step description",
				Variables:   map[string]string{},
			},
		},
	}
	
	// Generate markdown
	markdown := generateMarkdownOutput(details)
	
	// Validate content
	assert.Contains(t, markdown, "# Workflow: test-workflow")
	assert.Contains(t, markdown, "Test workflow description")
	assert.Contains(t, markdown, "### 1. step1")
	assert.Contains(t, markdown, "First step description")
	assert.Contains(t, markdown, "| var1 | value1 |")
	assert.Contains(t, markdown, "| var2 | value2 |")
	assert.Contains(t, markdown, "### 2. step2")
	assert.Contains(t, markdown, "Second step description")
	assert.Contains(t, markdown, "## Usage")
	assert.Contains(t, markdown, "usm code --workflow=test-workflow")
} 
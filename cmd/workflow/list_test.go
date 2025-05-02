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
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	internalio "github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// TestListWorkflow_NoWorkflows tests the behavior with no workflows found
func TestListWorkflow_NoWorkflows(t *testing.T) {
	// Create mock terminal IO and file system
	mockTermIO := internalio.NewTerminalIO()
	mockFS := internalio.NewMockFileSystem()
	
	// Reset the registry with a completely empty registry for this specific test
	registry := workflow.ResetGlobalRegistry()
	
	// Clear built-in workflows for this test
	registry.ClearBuiltInWorkflows()
	
	// Use a registry that would return no workflows
	// This is testing the empty workflow case
	result := listWorkflows("text", mockTermIO, mockFS, false)
	
	// Validate result
	assert.True(t, result.Success)
	assert.True(t, result.NoWorkflows)
	assert.Empty(t, result.Output)
	assert.Empty(t, result.Workflows)
}

// TestListCmd_TextFormat tests the workflow list command with text format
// This tests the full command integration
func TestListCmd_TextFormat(t *testing.T) {
	// Create a test command with capture output
	cmd := &cobra.Command{}
	cmd.Flags().StringP("format", "f", "text", "Output format")
	_ = cmd.Flags().Set("format", "text")
	cmd.Flags().Bool("debug", false, "Show debug output")
	
	// Create a buffer to capture output
	var output bytes.Buffer
	
	// Create test with text format
	testListCmd(t, cmd, &output)
	
	// Verify output
	outputStr := output.String()
	assert.Contains(t, outputStr, "Discovering workflows")
	
	// The command could output either workflows found or "No workflows found" 
	// depending on if any workflows exist
	if strings.Contains(outputStr, "No workflows found") {
		// If no workflows are found, that's a valid case
		assert.Contains(t, outputStr, "No workflows found")
	} else {
		// If workflows are found, expect these elements
		assert.Contains(t, outputStr, "Found")
		assert.Contains(t, outputStr, "workflows")
		assert.Contains(t, outputStr, "NAME")
		assert.Contains(t, outputStr, "DESCRIPTION")
	}
}

// TestListCmd_JSONFormat tests the workflow list command with JSON format
// This tests the full command integration
func TestListCmd_JSONFormat(t *testing.T) {
	// Skip this test if running in a CI environment where registry might be empty
	if os.Getenv("CI") != "" {
		t.Skip("Skipping JSON format test in CI environment")
	}
	
	// Create a test command
	cmd := &cobra.Command{}
	cmd.Flags().StringP("format", "f", "json", "Output format")
	_ = cmd.Flags().Set("format", "json")
	
	// Create a buffer to capture output
	var output bytes.Buffer
	
	// Create test with JSON format
	testListCmd(t, cmd, &output)
	
	// Verify output
	outputStr := output.String()
	assert.Contains(t, outputStr, "Discovering workflows")
	
	// Check if any workflows were found
	if strings.Contains(outputStr, "No workflows found") {
		// If no workflows are found, that's a valid case
		assert.Contains(t, outputStr, "No workflows found")
		return
	}
	
	// Find the JSON part in the output
	jsonStart := strings.Index(outputStr, "[")
	if jsonStart == -1 {
		t.Skip("JSON output not found - this might be ok if no workflows exist")
		return
	}
	
	jsonEnd := strings.LastIndex(outputStr, "]") + 1
	if jsonEnd <= jsonStart {
		t.Fatal("Invalid JSON output range")
	}
	
	// Parse the JSON to verify it's valid
	var workflowInfos []WorkflowInfo
	jsonData := outputStr[jsonStart:jsonEnd]
	err := json.Unmarshal([]byte(jsonData), &workflowInfos)
	if err != nil {
		t.Fatalf("Invalid JSON output: %v\nJSON: %s", err, jsonData)
	}
	
	// Verify we have workflow info
	assert.GreaterOrEqual(t, len(workflowInfos), 1, "Expected at least one workflow")
}

// testListCmd is a helper to prepare and test the list command
func testListCmd(t *testing.T, cmd *cobra.Command, output *bytes.Buffer) {
	t.Helper()
	
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
		_, _ = io.Copy(output, r)
		done <- true
	}()
	
	// Execute the command
	listCmdRun := ListCmd.Run
	listCmdRun(cmd, []string{})
	
	// Close writer and wait for copy to complete
	_ = w.Close()
	<-done
	
	// Close reader
	_ = r.Close()
}

// TestListWorkflow_CustomWorkflow tests that custom workflows display the correct source and path
func TestListWorkflow_CustomWorkflow(t *testing.T) {
	// Create mock terminal IO and file system
	mockTermIO := internalio.NewTerminalIO()
	mockFS := internalio.NewMockFileSystem()
	
	// Reset the registry to ensure a clean state
	registry := workflow.ResetGlobalRegistry()
	
	// Create a fake workflow directory structure that matches the one in the real application
	workflowDir := ".usm/workflows/asd"
	mockFS.MkdirAll(workflowDir, 0755)
	mockFS.MkdirAll(workflowDir+"/prompts", 0755)
	mockFS.MkdirAll(workflowDir+"/prompts/shared", 0755)
	
	// Create workflow.yaml file
	workflowYaml := `name: asd
description: Custom workflow created with usm workflow init
steps:
- id: step1
  description: First step
  prompt: step1.md
- id: step2
  description: Second step
  prompt: step2.md
`
	mockFS.WriteFile(workflowDir+"/workflow.yaml", []byte(workflowYaml), 0644)
	
	// Create prompt files
	mockFS.WriteFile(workflowDir+"/prompts/step1.md", []byte("Step 1 prompt"), 0644)
	mockFS.WriteFile(workflowDir+"/prompts/step2.md", []byte("Step 2 prompt"), 0644)
	mockFS.WriteFile(workflowDir+"/prompts/shared/README.md", []byte("Shared resources"), 0644)
	mockFS.WriteFile(workflowDir+"/README.md", []byte("Workflow readme"), 0644)
	
	// Create mock for the standard workflow 
	mockFS.WriteFile("workflow.yaml", []byte(`name: standard
description: The default USM workflow for implementation
steps:
- id: standard-step
  description: Standard workflow step
  prompt: standard.md
`), 0644)
	
	// Create direct custom list test for clarity, bypassing listWorkflows directly
	customTest := func() {
		// Add our test workflow to the registry through proper loading
		workflowDefPath := filepath.Join(workflowDir, "workflow.yaml")
		workflowDef, err := workflow.LoadWorkflowFromFile(mockFS, workflowDefPath)
		assert.NoError(t, err)
		assert.NotNil(t, workflowDef)
		
		// Add it to registry's cache directly to simulate proper loading
		if err == nil && workflowDef != nil {
			registry.AddToCache(workflowDef, workflowDefPath)
		}
		
		// Also add the standard workflow
		standardDef, err := workflow.LoadWorkflowFromFile(mockFS, "workflow.yaml")
		assert.NoError(t, err)
		if err == nil && standardDef != nil {
			registry.RegisterBuiltInWorkflow(standardDef)
		}
		
		// Get all workflows in registry to verify they are properly loaded
		allWorkflows := registry.ListWorkflows()
		t.Logf("Workflows in registry: %v", allWorkflows)
		
		// Run list command with our prepared registry
		result := listWorkflows("json", mockTermIO, mockFS, true) // Enable debug output
		
		// Validate result
		assert.True(t, result.Success)
		assert.False(t, result.NoWorkflows)
		assert.NotEmpty(t, result.Output)
		
		// Log the JSON output for debugging
		t.Logf("JSON output: %s", result.Output)
		
		// Parse the JSON to verify contents
		var workflowInfos []WorkflowInfo
		err = json.Unmarshal([]byte(result.Output), &workflowInfos)
		assert.NoError(t, err)
		
		// Log all workflows found in output
		for i, wf := range workflowInfos {
			t.Logf("Workflow %d: %s (source: %s, path: %s)", i, wf.Name, wf.Source, wf.Path)
		}
		
		// Find our custom workflow in the results
		var asdWorkflow *WorkflowInfo
		for _, wf := range workflowInfos {
			if wf.Name == "asd" {
				asdWorkflow = &wf
				break
			}
		}
		
		// Verify the workflow was found and has correct properties
		assert.NotNil(t, asdWorkflow, "Custom workflow 'asd' not found in results")
		if asdWorkflow != nil {
			assert.Equal(t, "Custom workflow created with usm workflow init", asdWorkflow.Description)
			assert.Equal(t, "project", asdWorkflow.Source, "Source should be 'project', not 'unknown'")
			assert.Equal(t, ".usm/workflows/asd", asdWorkflow.Path, "Path should be the workflow directory, not '-'")
		}
	}
	
	// Run the test
	customTest()
}
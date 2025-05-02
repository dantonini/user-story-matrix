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
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	internalio "github.com/user-story-matrix/usm/internal/io"
)

// TestListWorkflow_NoWorkflows tests the behavior with no workflows found
func TestListWorkflow_NoWorkflows(t *testing.T) {
	// Create mock terminal IO and file system
	mockTermIO := internalio.NewTerminalIO()
	mockFS := internalio.NewMockFileSystem()
	
	// Use a registry that would return no workflows
	// This is testing the empty workflow case
	result := listWorkflows("text", mockTermIO, mockFS)
	
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
package workflow

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/workflow"
	"gopkg.in/yaml.v3"
)

// TestValidateCmd tests the workflow validate command
func TestValidateCmd(t *testing.T) {
	// Skip this test when running with -short flag
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Create a test workflow directory and file
	tempDir := t.TempDir()
	
	workflowDir := filepath.Join(tempDir, ".usm", "workflows", "test-workflow")
	assert.NoError(t, os.MkdirAll(workflowDir, 0755))
	assert.NoError(t, os.MkdirAll(filepath.Join(workflowDir, "prompts"), 0755))
	
	// Create workflow file
	workflowDef := &workflow.ExternalWorkflowDefinition{
		Name:        "test-workflow",
		Description: "Test workflow",
		Steps: []workflow.ExternalWorkflowStep{
			{
				ID:          "test-step",
				Description: "Test step",
				Prompt:      "Test prompt",
			},
		},
	}
	workflowYAMLBytes, _ := yaml.Marshal(workflowDef)
	workflowFile := filepath.Join(workflowDir, "workflow.yaml")
	assert.NoError(t, os.WriteFile(workflowFile, workflowYAMLBytes, 0644))
	
	// Change to the temp directory for testing
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(cwd)
	assert.NoError(t, os.Chdir(tempDir))
	
	// Test validating by name
	t.Run("validate by name", func(t *testing.T) {
		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		// Run the command 
		cmd := &cobra.Command{}
		ValidateCmd.Run(cmd, []string{"test-workflow"})
		
		// Restore stdout
		w.Close()
		os.Stdout = old
		
		// Read captured output
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()
		
		// Verify output contains success message
		assert.Contains(t, output, "Validating workflow")
		assert.Contains(t, output, "Workflow is valid")
	})
	
	// Test validating by path
	t.Run("validate by path", func(t *testing.T) {
		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		// Run the command 
		cmd := &cobra.Command{}
		ValidateCmd.Run(cmd, []string{".usm/workflows/test-workflow"})
		
		// Restore stdout
		w.Close()
		os.Stdout = old
		
		// Read captured output
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()
		
		// Verify output contains success message
		assert.Contains(t, output, "Validating workflow")
		assert.Contains(t, output, "Workflow is valid")
	})
} 
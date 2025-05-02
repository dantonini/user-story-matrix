package workflow

import (
	"bytes"
	stdIO "io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
	"gopkg.in/yaml.v3"
)

// TestValidateCmd tests the Cobra command wrapper
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
		_, _ = stdIO.Copy(&buf, r)
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
		_, _ = stdIO.Copy(&buf, r)
		output := buf.String()
		
		// Verify output contains success message
		assert.Contains(t, output, "Validating workflow")
		assert.Contains(t, output, "Workflow is valid")
	})
}

// TestValidateWorkflow tests the business logic directly
func TestValidateWorkflow(t *testing.T) {
	// Skip this test when running with -short flag
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	
	// Setup test cases
	tests := []struct {
		name           string
		setupFunc      func() (string, string) // returns nameOrPath and workflowDir
		expectSuccess  bool
		expectWarnings bool
		expectErrors   bool
	}{
		{
			name: "validate by name",
			setupFunc: func() (string, string) {
				// Create temp directory with workflow
				tempDir := t.TempDir()
				workflowDir := filepath.Join(tempDir, ".usm", "workflows", "test-workflow")
				os.MkdirAll(workflowDir, 0755)
				os.MkdirAll(filepath.Join(workflowDir, "prompts"), 0755)
				
				// Create workflow file
				workflowDef := &workflow.ExternalWorkflowDefinition{
					Name:        "test-workflow",
					Description: "Test workflow by name",
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
				os.WriteFile(workflowFile, workflowYAMLBytes, 0644)
				
				// Change to the temp directory for testing
				os.Chdir(tempDir)
				
				return "test-workflow", workflowDir
			},
			expectSuccess:  true,
			expectWarnings: false,
			expectErrors:   false,
		},
		{
			name: "validate by path",
			setupFunc: func() (string, string) {
				// Create temp directory with workflow
				tempDir := t.TempDir()
				workflowDir := filepath.Join(tempDir, "custom-path", "my-workflow")
				os.MkdirAll(workflowDir, 0755)
				os.MkdirAll(filepath.Join(workflowDir, "prompts"), 0755)
				
				// Create workflow file
				workflowDef := &workflow.ExternalWorkflowDefinition{
					Name:        "path-workflow",
					Description: "Test workflow by path",
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
				os.WriteFile(workflowFile, workflowYAMLBytes, 0644)
				
				// Change to the temp directory for testing
				os.Chdir(tempDir)
				
				return filepath.Join("custom-path", "my-workflow"), workflowDir
			},
			expectSuccess:  true,
			expectWarnings: false,
			expectErrors:   false,
		},
		{
			name: "workflow not found",
			setupFunc: func() (string, string) {
				// Create empty temp directory
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				
				return "non-existent-workflow", ""
			},
			expectSuccess:  false,
			expectWarnings: false,
			expectErrors:   false,
		},
	}
	
	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current directory
			origDir, _ := os.Getwd()
			defer os.Chdir(origDir)
			
			// Setup test environment
			nameOrPath, _ := tt.setupFunc()
			
			// Create filesystem and output interfaces
			fs := io.NewOSFileSystem()
			output := io.NewTerminalIO()
			
			// Call the business logic function directly
			result := validateWorkflow(nameOrPath, fs, output)
			
			// Check results
			if tt.expectSuccess {
				assert.True(t, result.Success, "Expected validation to succeed")
				assert.NotEmpty(t, result.WorkflowName, "Expected workflow name to be set")
				assert.Empty(t, result.ErrorMessage, "Expected no error message")
			} else {
				assert.False(t, result.Success, "Expected validation to fail")
				assert.NotEmpty(t, result.ErrorMessage, "Expected error message to be set")
			}
			
			if tt.expectWarnings {
				assert.NotEmpty(t, result.Warnings, "Expected warnings")
			}
			
			if tt.expectErrors {
				assert.NotEmpty(t, result.Errors, "Expected errors")
			}
		})
	}
} 
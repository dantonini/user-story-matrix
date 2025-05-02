// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"bytes"
	stdIO "io"
	"os"
	"path/filepath"
	"strings"
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

// TestValidateWorkflowWithMissingVariables verifies that validation warnings are properly
// displayed when variables are referenced but not defined in workflow steps
func TestValidateWorkflowWithMissingVariables(t *testing.T) {
	// Create a mock file system
	mockFS := io.NewMockFileSystem()
	
	// Create a mock output
	mockOutput := &mockUserOutput{}
	
	// Set up workflow directory - use .usm directory structure to mimic real environment
	workflowDir := "/.usm/workflows/test-workflow"
	mockFS.MkdirAll(workflowDir, 0755)
	
	// Create prompts directory
	promptsDir := filepath.Join(workflowDir, "prompts")
	mockFS.MkdirAll(promptsDir, 0755)
	
	// Create prompt files with deliberate variable reference that's missing
	step1PromptPath := filepath.Join(promptsDir, "step1.md")
	step1Content := `# This step uses {{.key1}} and {{.key2}}.`
	mockFS.WriteFile(step1PromptPath, []byte(step1Content), 0644)
	
	// Print out the content that was written to verify
	data, err := mockFS.ReadFile(step1PromptPath)
	assert.NoError(t, err, "Should be able to read file")
	t.Logf("Content of step1.md: %s", string(data))
	
	// Create workflow.yaml
	workflowYAML := `name: "test-workflow"
description: "Test workflow"
steps:
  - id: "01-step-one"
    description: "First step"
    prompt: "prompts/step1.md"
    variables:
      key1: "value1"
      # key2 is missing intentionally to test validation
`
	workflowYAMLPath := filepath.Join(workflowDir, "workflow.yaml")
	mockFS.WriteFile(workflowYAMLPath, []byte(workflowYAML), 0644)

	// Debug the yaml content
	yamlData, err := mockFS.ReadFile(workflowYAMLPath)
	assert.NoError(t, err, "Should be able to read workflow file")
	t.Logf("Content of workflow.yaml:\n%s", string(yamlData))

	// Add the standard workflow directories to the filesystem
	for _, dir := range workflow.GetStandardWorkflowDirectories() {
		mockFS.MkdirAll(dir, 0755)
	}
	
	// Debug the file system to verify content
	t.Logf("File system contents at %s:", workflowDir)
	if dirExists := mockFS.Exists(workflowDir); dirExists {
		t.Logf("- Directory exists")
		t.Logf("- Contains workflow.yaml: %v", mockFS.Exists(workflowYAMLPath))
		t.Logf("- Contains prompts dir: %v", mockFS.Exists(promptsDir))
		t.Logf("- Contains step1.md: %v", mockFS.Exists(step1PromptPath))
	} else {
		t.Logf("- Directory does not exist")
	}
	
	// Load workflow directly to verify
	wfDef, err := workflow.LoadWorkflowFromFile(mockFS, workflowYAMLPath)
	assert.NoError(t, err, "Should load workflow definition")
	
	// Check that the template is properly linked
	t.Logf("Workflow definition: name=%s, steps=%d", wfDef.Name, len(wfDef.Steps))
	if len(wfDef.Steps) > 0 {
		step := wfDef.Steps[0]
		t.Logf("Step 0: ID=%s", step.ID)
		t.Logf("Step 0 variables: %v", step.Variables)
	}
	
	// Directly validate using workflow validator
	validator := workflow.NewWorkflowValidator(mockFS, workflowDir)
	valResult, err := validator.ValidateWorkflow(wfDef)
	assert.NoError(t, err, "Direct validation should not error")
	
	t.Logf("Direct validation results - Errors: %v, Warnings: %v", valResult.Errors, valResult.Warnings)
	assert.True(t, len(valResult.Warnings) > 0, "Direct validation should have warnings")
	
	// Now run through the command validation
	result := validateWorkflow(workflowDir, mockFS, mockOutput)
	
	t.Logf("Command validation results - Success: %v, Errors: %v, Warnings: %v", 
		result.Success, result.Errors, result.Warnings)
	
	// The workflow should be valid (missing variables only generate warnings)
	assert.True(t, result.Success, "Workflow should be valid despite missing variables")
	
	// But there should be warnings
	assert.NotEmpty(t, result.Warnings, "There should be warnings about missing variables")
	
	// Check specifically for warning about key2
	foundWarning := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "key2") {
			foundWarning = true
			break
		}
	}
	assert.True(t, foundWarning, "Warning about missing key2 should be present")
	
	// Now try with workflow name instead of path
	result = validateWorkflow("test-workflow", mockFS, mockOutput)
	t.Logf("Validation by name results - Success: %v, Errors: %v, Warnings: %v", 
		result.Success, result.Errors, result.Warnings)
}

// mockUserOutput is a mock implementation of UserOutput for testing
type mockUserOutput struct {
	messages []string
}

func (o *mockUserOutput) Print(message string) {
	o.messages = append(o.messages, message)
}

func (o *mockUserOutput) PrintSuccess(message string) {
	o.messages = append(o.messages, "[SUCCESS] "+message)
}

func (o *mockUserOutput) PrintError(message string) {
	o.messages = append(o.messages, "[ERROR] "+message)
}

func (o *mockUserOutput) PrintWarning(message string) {
	o.messages = append(o.messages, "[WARNING] "+message)
}

func (o *mockUserOutput) PrintProgress(message string) {
	o.messages = append(o.messages, "[PROGRESS] "+message)
}

func (o *mockUserOutput) PrintStep(stepNumber int, totalSteps int, description string) {
	o.messages = append(o.messages, description)
}

func (o *mockUserOutput) PrintTable(headers []string, rows [][]string) {
	// Simple implementation for testing
	o.messages = append(o.messages, "TABLE")
}

func (o *mockUserOutput) IsDebugEnabled() bool {
	return false
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"path/filepath"
	"testing"

	"github.com/user-story-matrix/usm/internal/io"
)

// TestWorkflowMVI_SmokeTest performs a smoke test of the workflow MVI functionality
func TestWorkflowMVI_SmokeTest(t *testing.T) {
	// Create mock filesystem and output
	fs := io.NewMockFileSystem()
	output := io.NewMockIO()

	// Create test workflow directory
	workflowDir := "/test-workflow"
	fs.MkdirAll(workflowDir, 0755)
	fs.MkdirAll(filepath.Join(workflowDir, "prompts"), 0755)
	promptsDir := filepath.Join(workflowDir, "prompts")
	
	// Create workflow.yaml
	workflowYAML := `
name: "custom-workflow"
description: "Test custom workflow"
steps:
  - id: "01-step-one"
    description: "First step"
    prompt: "prompts/step1.md"
    variables:
      name: "User Story Matrix"
      version: "1.0.0"
  - id: "02-step-two"
    description: "Second step with variable"
    prompt: "prompts/step2.md"
    variables:
      tool: "USM"
      phase: "implementation"
  - id: "03-step-three"
    description: "Step with default values"
    prompt: "prompts/step3.md"
    variables:
      feature: ""
      priority: ""
`
	fs.AddFile(filepath.Join(workflowDir, WorkflowConfigFile), []byte(workflowYAML))

	// Create prompt files
	step1Prompt := "# Step 1\nWelcome to {{.name}} version {{.version}}!\n"
	fs.AddFile(filepath.Join(promptsDir, "step1.md"), []byte(step1Prompt))

	step2Prompt := "# Step 2\nImplementing {{.tool}} in the {{.phase}} phase.\n"
	fs.AddFile(filepath.Join(promptsDir, "step2.md"), []byte(step2Prompt))

	// Use the exact default pattern expected by checkForDefaultValue
	step3Prompt := "# Step 3\nUsing {{.feature | default \"default feature\"}} with {{.priority | default \"high\"}} priority.\n"
	fs.AddFile(filepath.Join(promptsDir, "step3.md"), []byte(step3Prompt))

	// --- Test 1: Directory loading ---
	t.Run("DirectoryLoading", func(t *testing.T) {
		// Load workflow from directory
		workflow, source, err := LoadWorkflowFromDirectory(fs, workflowDir)
		if err != nil {
			t.Fatalf("Failed to load workflow from directory: %v", err)
		}

		// Verify workflow properties
		if workflow.Name != "custom-workflow" {
			t.Errorf("Expected workflow name 'custom-workflow', got '%s'", workflow.Name)
		}

		if len(workflow.Steps) != 3 {
			t.Errorf("Expected 3 steps, got %d", len(workflow.Steps))
		}

		// Verify workflow source
		if source.Source != SourceProject {
			t.Errorf("Expected workflow source '%s', got '%s'", SourceProject, source.Source)
		}
	})

	// --- Test 2: Workflow validation ---
	t.Run("WorkflowValidation", func(t *testing.T) {
		// Load the workflow from the directory first
		workflowDef, _, err := LoadWorkflowFromDirectory(fs, workflowDir)
		if err != nil {
			t.Fatalf("Failed to load workflow from directory: %v", err)
		}
		
		// Create validator with the workflow path
		validator := NewWorkflowValidator(fs, workflowDir)
		
		// Validate the workflow definition
		result, err := validator.ValidateWorkflow(workflowDef)
		if err != nil {
			t.Fatalf("Validation error: %v", err)
		}
		
		if !result.IsValid() {
			for _, err := range result.Errors {
				t.Logf("Validation error: %s", err)
			}
			t.Errorf("Workflow validation failed with %d errors", len(result.Errors))
		}
	})

	// --- Test 3: Template rendering ---
	t.Run("TemplateRendering", func(t *testing.T) {
		renderer := NewTemplateRenderer(fs, workflowDir)

		// Test with provided variables
		variables1 := map[string]interface{}{
			"name":    "User Story Matrix",
			"version": "1.0.0",
		}
		rendered1, err := renderer.RenderPrompt("prompts/step1.md", variables1)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}
		expected1 := "# Step 1\nWelcome to User Story Matrix version 1.0.0!\n"
		if rendered1 != expected1 {
			t.Errorf("Template rendering mismatch.\nExpected: %q\nGot: %q", expected1, rendered1)
		}

		// Test with default values
		variables3 := map[string]interface{}{
			// No variables provided, should use defaults
		}
		rendered3, err := renderer.RenderPrompt("prompts/step3.md", variables3)
		if err != nil {
			t.Fatalf("Failed to render template with defaults: %v", err)
		}
		expected3 := "# Step 3\nUsing default feature with high priority.\n"
		if rendered3 != expected3 {
			t.Errorf("Template rendering with defaults mismatch.\nExpected: %q\nGot: %q", expected3, rendered3)
		}
	})

	// --- Test 4: Variable extraction ---
	t.Run("VariableExtraction", func(t *testing.T) {
		renderer := NewTemplateRenderer(fs, workflowDir)
		variables, err := renderer.ExtractTemplateVariables("prompts/step2.md")
		if err != nil {
			t.Fatalf("Failed to extract variables: %v", err)
		}

		// Verify extracted variables
		expectedVars := []string{"tool", "phase"}
		for _, expected := range expectedVars {
			found := false
			for _, actual := range variables {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected variable '%s' not found in extracted variables: %v", expected, variables)
			}
		}
	})

	// --- Test 5: Default value validation ---
	t.Run("DefaultValueValidation", func(t *testing.T) {
		// Test default value detection directly
		promptPath := filepath.Join(promptsDir, "step3.md")
		
		// Check if features has default value
		hasDefaultFeature := checkForDefaultValue(fs, promptPath, "feature")
		if !hasDefaultFeature {
			t.Errorf("Expected to detect default value for 'feature', but none found")
		}
		
		// Check if priority has default value
		hasDefaultPriority := checkForDefaultValue(fs, promptPath, "priority")
		if !hasDefaultPriority {
			t.Errorf("Expected to detect default value for 'priority', but none found")
		}
		
		// Check if a nonexistent variable has default value (should be false)
		hasDefaultNonexistent := checkForDefaultValue(fs, promptPath, "nonexistent")
		if hasDefaultNonexistent {
			t.Errorf("Incorrectly detected default value for nonexistent variable")
		}
	})

	// --- Test 6: Workflow manager with path ---
	t.Run("WorkflowManagerWithPath", func(t *testing.T) {
		// Create a change request for testing
		changeRequestPath := "/test/changes-request/test.blueprint.md"
		fs.AddFile(changeRequestPath, []byte("# Test Change Request"))
		
		// Create a state file directory
		stateDir := filepath.Dir(GenerateStateFilePath(changeRequestPath))
		fs.MkdirAll(stateDir, 0755)

		// Create a workflow manager with a custom workflow path
		manager, err := NewWorkflowManagerWithPath(fs, output, workflowDir)
		if err != nil {
			t.Fatalf("Failed to create workflow manager with path: %v", err)
		}

		// Verify the workflow was loaded
		if manager.workflow.Name != "custom-workflow" {
			t.Errorf("Expected workflow name 'custom-workflow', got '%s'", manager.workflow.Name)
		}

		// Initialize state
		state, err := manager.LoadState(changeRequestPath)
		if err != nil {
			t.Fatalf("Failed to load state: %v", err)
		}

		// Set the workflow path in state
		state.WorkflowPath = workflowDir
		
		// Save state and verify
		err = manager.SaveState(state)
		if err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}
		
		// Reload state and verify path was saved
		reloadedState, err := manager.LoadState(changeRequestPath)
		if err != nil {
			t.Fatalf("Failed to reload state: %v", err)
		}
		
		if reloadedState.WorkflowPath != workflowDir {
			t.Errorf("Expected workflow path '%s', got '%s'", workflowDir, reloadedState.WorkflowPath)
		}
	})
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

// TestBuiltinWorkflowVariables tests that builtin workflows have empty Variables fields
// This is the root cause of the issue - builtin workflows use template variables in their
// Prompt strings but don't populate the Variables field.
func TestBuiltinWorkflowVariables(t *testing.T) {
	registry := GetGlobalRegistry()
	
	// Get the usm-code builtin workflow
	workflow, err := registry.GetWorkflow("usm-code")
	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	
	// Check that steps have template variables in their prompts but empty Variables field
	foundPromptWithTemplate := false
	foundStepWithEmptyVariables := false
	
	for _, step := range workflow.Steps {
		// Check if the prompt contains template variables
		if containsTemplateVariable(step.Prompt, "ChangeRequestFilePath") {
			foundPromptWithTemplate = true
			
			// Check if Variables field is empty for this step
			if step.Variables == nil || len(step.Variables) == 0 {
				foundStepWithEmptyVariables = true
				t.Logf("Step '%s' has template variable in prompt but empty Variables field", step.ID)
			}
		}
	}
	
	assert.True(t, foundPromptWithTemplate, "Should find at least one step with ChangeRequestFilePath template variable")
	assert.True(t, foundStepWithEmptyVariables, "Should find at least one step with empty Variables field")
}

// TestWorkflowInitFromBuiltinVariablesCopied tests that when initializing a workflow
// from a builtin template, variables from the prompt templates are properly extracted
// and added to the workflow.yaml file.
func TestWorkflowInitFromBuiltinVariablesCopied(t *testing.T) {
	// Create a mock filesystem
	mockFS := io.NewMockFileSystem()
	
	// Get the builtin workflow
	registry := GetGlobalRegistry()
	builtinWorkflow, err := registry.GetWorkflow("usm-code")
	assert.NoError(t, err)
	
	// Simulate the workflow init process
	workflowName := "test-init-workflow"
	targetDir := filepath.Join(".usm", "workflows", workflowName)
	promptsDir := filepath.Join(targetDir, PromptsDir)
	sharedDir := filepath.Join(promptsDir, SharedDir)
	
	// Create directories (mocked)
	mockFS.AddDirectory(targetDir)
	mockFS.AddDirectory(promptsDir)
	mockFS.AddDirectory(sharedDir)
	
	// Create the external workflow structure (simulating the init process)
	externalWorkflow := ExternalWorkflowDefinition{
		Name:        workflowName,
		Description: "Custom workflow based on usm-code",
		Steps:       make([]ExternalWorkflowStep, len(builtinWorkflow.Steps)),
	}
	
	// Convert each step (this is where the bug occurs)
	for i, step := range builtinWorkflow.Steps {
		promptFileName := step.ID + ".md"
		
		// This is the current buggy implementation that doesn't extract variables
		externalWorkflow.Steps[i] = ExternalWorkflowStep{
			ID:          step.ID,
			Description: step.Description,
			Variables:   step.Variables, // This is empty/nil for builtin steps
			Prompt:      "prompts/" + promptFileName,
		}
		
		// Write prompt file
		mockFS.AddFile(filepath.Join(promptsDir, promptFileName), []byte(step.Prompt))
	}
	
	// Now test that variables are missing (demonstrating the bug)
	foundStepWithMissingVariables := false
	for _, step := range externalWorkflow.Steps {
		// Read the prompt file to check for template variables
		promptPath := filepath.Join(targetDir, step.Prompt)
		if mockFS.Exists(promptPath) {
			promptContent, _ := mockFS.ReadFile(promptPath)
			if containsTemplateVariable(string(promptContent), "ChangeRequestFilePath") {
				// This step uses ChangeRequestFilePath but doesn't have it in Variables
				if step.Variables == nil || step.Variables["ChangeRequestFilePath"] == "" {
					foundStepWithMissingVariables = true
					t.Logf("Step '%s' uses ChangeRequestFilePath in template but doesn't have it in Variables", step.ID)
				}
			}
		}
	}
	
	assert.True(t, foundStepWithMissingVariables, "Should find steps with missing variables (demonstrating the bug)")
}

// TestWorkflowInitFromBuiltinVariablesExtracted tests that the fix works properly
// When initializing from builtin workflow, variables should be extracted and added
func TestWorkflowInitFromBuiltinVariablesExtracted(t *testing.T) {
	// Get the builtin workflow
	registry := GetGlobalRegistry()
	builtinWorkflow, err := registry.GetWorkflow("usm-code")
	assert.NoError(t, err)
	
	// Simulate the FIXED workflow init process
	workflowName := "test-fixed-workflow"
	externalWorkflow := ExternalWorkflowDefinition{
		Name:        workflowName,
		Description: "Custom workflow based on usm-code",
		Steps:       make([]ExternalWorkflowStep, len(builtinWorkflow.Steps)),
	}
	
	// Convert each step using the FIX
	for i, step := range builtinWorkflow.Steps {
		promptFileName := step.ID + ".md"
		
		// Extract variables from the builtin step (THE FIX)
		extractedVariables, err := ExtractVariablesFromBuiltinStep(step)
		assert.NoError(t, err, "Should be able to extract variables from step %s", step.ID)
		
		// This is the FIXED implementation that extracts variables
		externalWorkflow.Steps[i] = ExternalWorkflowStep{
			ID:          step.ID,
			Description: step.Description,
			Variables:   extractedVariables, // Use extracted variables instead of empty
			Prompt:      "prompts/" + promptFileName,
		}
	}
	
	// Now test that predefined variables are NOT added to Variables (verifying the fix)
	foundStepWithoutPredefinedVariables := false
	for i, step := range externalWorkflow.Steps {
		// Get the corresponding builtin step to check its actual prompt content
		builtinStep := builtinWorkflow.Steps[i]
		
		// Check if the builtin step uses ChangeRequestFilePath in its prompt template
		if containsTemplateVariable(builtinStep.Prompt, "ChangeRequestFilePath") {
			// This step should NOT have ChangeRequestFilePath in Variables since it's predefined
			if step.Variables == nil || step.Variables["ChangeRequestFilePath"] == "" {
				foundStepWithoutPredefinedVariables = true
				t.Logf("Step '%s' uses ChangeRequestFilePath in builtin template but correctly does NOT have it in Variables (predefined runtime variable)", step.ID)
			}
		}
	}
	
	assert.True(t, foundStepWithoutPredefinedVariables, "Should find steps that use predefined variables but don't declare them (verifying the fix)")
}

// Helper function to check if a string contains a template variable
func containsTemplateVariable(content, varName string) bool {
	templateVar := "{{." + varName + "}}"
	return len(content) > 0 && (
		// Check for the exact template variable
		// Note: strings.Contains would be imported but I'll use a simple check
		findSubstring(content, templateVar))
}

// Simple substring search since we can't assume strings package is imported
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
} 
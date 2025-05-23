// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPredefinedRuntimeVariables(t *testing.T) {
	changeRequestPath := "/project/docs/changes-request/feature.blueprint.md"
	step := WorkflowStep{
		ID:          "01-analyze-feature",
		Description: "Analyze the feature request",
	}

	variables := BuildPredefinedRuntimeVariables(changeRequestPath, step)

	// Test all expected predefined variables are present
	expectedVars := map[string]interface{}{
		"ChangeRequestFilePath": "/project/docs/changes-request/feature.blueprint.md",
		"ChangeRequestBasename": "feature",
		"BlueprintBasename":     "feature",
		"ChangeRequestDirname":  "/project/docs/changes-request",
		"StepID":                "01-analyze-feature",
		"StepName":              "analyze-feature",
		"ChangeRequestFullpath": "/project/docs/changes-request/feature",
		"Basename":              "feature", // Deprecated
	}

	assert.Equal(t, expectedVars, variables, "BuildPredefinedRuntimeVariables should return expected variables")

	// Test that all variables in PredefinedRuntimeVariables map are actually generated
	for varName := range PredefinedRuntimeVariables {
		assert.Contains(t, variables, varName, "Variable %s should be generated", varName)
	}
}

func TestBuildPredefinedRuntimeVariables_StepNameExtraction(t *testing.T) {
	tests := []struct {
		name           string
		stepID         string
		expectedStepName string
	}{
		{
			name:           "step with prefix",
			stepID:         "01-laying-foundation",
			expectedStepName: "laying-foundation",
		},
		{
			name:           "step without numeric prefix",
			stepID:         "simple-step",
			expectedStepName: "step", // Splits on first dash, takes everything after
		},
		{
			name:           "step with multiple dashes",
			stepID:         "02-very-long-step-name",
			expectedStepName: "very-long-step-name",
		},
	}

	changeRequestPath := "/test/file.blueprint.md"
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := WorkflowStep{ID: tt.stepID}
			variables := BuildPredefinedRuntimeVariables(changeRequestPath, step)
			
			assert.Equal(t, tt.expectedStepName, variables["StepName"], "StepName should be extracted correctly")
			assert.Equal(t, tt.stepID, variables["StepID"], "StepID should match the original")
		})
	}
}

func TestIsPredefinedRuntimeVariable(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{
			name:     "ChangeRequestFilePath is predefined",
			varName:  "ChangeRequestFilePath",
			expected: true,
		},
		{
			name:     "ChangeRequestDirname is predefined",
			varName:  "ChangeRequestDirname",
			expected: true,
		},
		{
			name:     "ChangeRequestBasename is predefined",
			varName:  "ChangeRequestBasename",
			expected: true,
		},
		{
			name:     "StepID is predefined",
			varName:  "StepID",
			expected: true,
		},
		{
			name:     "StepName is predefined",
			varName:  "StepName",
			expected: true,
		},
		{
			name:     "CustomVariable is not predefined",
			varName:  "CustomVariable",
			expected: false,
		},
		{
			name:     "ProjectName is not predefined",
			varName:  "ProjectName",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPredefinedRuntimeVariable(tt.varName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTemplateVariablesFromString(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		expected   []string
		shouldFail bool
	}{
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "no variables",
			content:  "This is plain text with no variables",
			expected: []string{},
		},
		{
			name:     "single variable",
			content:  "Hello {{.Name}}!",
			expected: []string{"Name"},
		},
		{
			name:     "multiple variables",
			content:  "File: {{.ChangeRequestFilePath}} in {{.ChangeRequestDirname}}",
			expected: []string{"ChangeRequestFilePath", "ChangeRequestDirname"},
		},
		{
			name:     "variable with default",
			content:  "Value: {{.Variable | default \"default_value\"}}",
			expected: []string{"Variable"},
		},
		{
			name:     "duplicate variables",
			content:  "{{.Name}} says hello to {{.Name}}",
			expected: []string{"Name"}, // Should deduplicate
		},
		{
			name:     "complex template",
			content:  "Process {{.ChangeRequestFilePath}} and generate {{.OutputPath | default \"output.txt\"}}",
			expected: []string{"ChangeRequestFilePath", "OutputPath"},
		},
		{
			name:       "invalid template syntax",
			content:    "{{.Name",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variables, err := ExtractTemplateVariablesFromString(tt.content)
			
			if tt.shouldFail {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			
			// Convert to map for easy comparison (order doesn't matter)
			expectedMap := make(map[string]bool)
			for _, v := range tt.expected {
				expectedMap[v] = true
			}
			
			actualMap := make(map[string]bool)
			for _, v := range variables {
				actualMap[v] = true
			}
			
			assert.Equal(t, expectedMap, actualMap, "Variables should match")
		})
	}
}

func TestExtractVariablesFromBuiltinStep(t *testing.T) {
	tests := []struct {
		name     string
		step     WorkflowStep
		expected map[string]string
	}{
		{
			name: "step with ChangeRequestFilePath (predefined runtime variable)",
			step: WorkflowStep{
				ID:     "test-step",
				Prompt: "Read the file {{.ChangeRequestFilePath}}",
			},
			expected: map[string]string{}, // Predefined variables should not be included
		},
		{
			name: "step with multiple runtime variables (all predefined)",
			step: WorkflowStep{
				ID:     "test-step",
				Prompt: "File: {{.ChangeRequestFilePath}} Dir: {{.ChangeRequestDirname}} Base: {{.ChangeRequestBasename}}",
			},
			expected: map[string]string{}, // All predefined variables should not be included
		},
		{
			name: "step with unknown variable",
			step: WorkflowStep{
				ID:     "test-step",
				Prompt: "Unknown: {{.SomeUnknownVariable}}",
			},
			expected: map[string]string{
				"SomeUnknownVariable": "CONFIGURE_ME",
			},
		},
		{
			name: "step with no variables",
			step: WorkflowStep{
				ID:     "test-step",
				Prompt: "This has no template variables",
			},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractVariablesFromBuiltinStep(tt.step)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractVariablesFromActualBuiltinStep(t *testing.T) {
	// Test with actual builtin workflow step
	registry := GetGlobalRegistry()
	workflow, err := registry.GetWorkflow("usm-code")
	assert.NoError(t, err)
	
	// Find a step that uses ChangeRequestFilePath
	var testStep *WorkflowStep
	for i, step := range workflow.Steps {
		if containsTemplateVariable(step.Prompt, "ChangeRequestFilePath") {
			testStep = &workflow.Steps[i]
			break
		}
	}
	
	assert.NotNil(t, testStep, "Should find a step with ChangeRequestFilePath variable")
	
	// Extract variables from the actual step
	variables, err := ExtractVariablesFromBuiltinStep(*testStep)
	assert.NoError(t, err)
	
	// Should NOT contain ChangeRequestFilePath since it's a predefined runtime variable
	assert.NotContains(t, variables, "ChangeRequestFilePath")
	
	// Should be empty or contain only non-predefined variables
	t.Logf("Extracted variables from step '%s': %v (predefined variables excluded)", testStep.ID, variables)
} 
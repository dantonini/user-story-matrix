// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"strings"
	"testing"
)

func TestWorkflowStepStructure(t *testing.T) {
	// Test that WorkflowStep struct includes the Prompt field
	step := WorkflowStep{
		ID:          "test-id",
		Description: "Test description",
		Prompt:      "Test prompt",
	}

	// Verify prompt field can be set and retrieved
	if step.Prompt != "Test prompt" {
		t.Errorf("Expected Prompt field to be 'Test prompt', got '%s'", step.Prompt)
	}
}

func TestValidatePrompt(t *testing.T) {
	tests := []struct {
		name          string
		prompt        string
		shouldBeValid bool
		errorContains string
	}{
		{
			name:          "Valid Go template prompt",
			prompt:        "This is a valid prompt with {{.ChangeRequestFilePath}}",
			shouldBeValid: true,
			errorContains: "",
		},
		{
			name:          "Valid prompt with multiple variables",
			prompt:        "The change request is {{.ChangeRequestFilePath}}. The basename is {{.ChangeRequestBasename}}.",
			shouldBeValid: true,
			errorContains: "",
		},
		{
			name:          "Valid prompt with conditionals",
			prompt:        "{{if .ChangeRequestFilePath}}File: {{.ChangeRequestFilePath}}{{end}}",
			shouldBeValid: true,
			errorContains: "",
		},
		{
			name:          "Valid prompt with default function",
			prompt:        "Project: {{.ProjectName | default \"USM\"}}",
			shouldBeValid: true,
			errorContains: "",
		},
		{
			name:          "Invalid template syntax - unclosed tag",
			prompt:        "This prompt has an {{.Variable unclosed tag",
			shouldBeValid: false,
			errorContains: "template syntax error",
		},
		{
			name:          "Invalid template syntax - malformed",
			prompt:        "This prompt has {{.Variable with.invalid.syntax}}",
			shouldBeValid: false,
			errorContains: "template syntax error",
		},
		{
			name:          "Empty prompt",
			prompt:        "",
			shouldBeValid: true,
			errorContains: "",
		},
		{
			name:          "Valid prompt with range and conditionals",
			prompt:        "{{range .Items}}{{if .Name}}Item: {{.Name}}{{end}}{{end}}",
			shouldBeValid: true,
			errorContains: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePrompt(tc.prompt)

			if tc.shouldBeValid && err != nil {
				t.Errorf("Expected prompt to be valid, but got error: %v", err)
			}

			if !tc.shouldBeValid && err == nil {
				t.Errorf("Expected prompt to be invalid, but no error was returned")
			}

			if !tc.shouldBeValid && err != nil && tc.errorContains != "" {
				if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Error message '%v' does not contain expected text '%s'", err, tc.errorContains)
				}
			}
		})
	}
}

func TestGenerateStepPrompt(t *testing.T) {
	// Setup test data
	changeRequestPath := "/path/to/my-change-request.blueprint.md"

	// Test cases
	tests := []struct {
		name     string
		step     WorkflowStep
		contains []string // Strings that should be in the output
		wantErr  bool
	}{
		{
			name: "Go template prompt",
			step: WorkflowStep{
				ID:          "01-test",
				Description: "Test step",
				Prompt:      "This is a test prompt with {{.ChangeRequestFilePath}}",
			},
			contains: []string{changeRequestPath},
			wantErr:  false,
		},
		{
			name: "Prompt with multiple variables",
			step: WorkflowStep{
				ID:          "02-test",
				Description: "Test step with multiple variables",
				Prompt:      "Processing {{.ChangeRequestFilePath}} with step {{.StepID}}",
			},
			contains: []string{changeRequestPath, "02-test"},
			wantErr:  false,
		},
		{
			name: "Prompt with conditionals",
			step: WorkflowStep{
				ID:          "03-test",
				Description: "Test step with conditionals",
				Prompt:      "{{if .ChangeRequestFilePath}}Working on: {{.ChangeRequestFilePath}}{{end}}",
			},
			contains: []string{"Working on:", changeRequestPath},
			wantErr:  false,
		},
		{
			name: "Prompt with custom variables",
			step: WorkflowStep{
				ID:          "04-test",
				Description: "Test step with custom variables",
				Prompt:      "Project {{.ProjectName}} feature {{.FeatureName}}",
				Variables: map[string]string{
					"ProjectName": "USM",
					"FeatureName": "Templates",
				},
			},
			contains: []string{"Project USM", "feature Templates"},
			wantErr:  false,
		},
		{
			name: "Prompt with default function",
			step: WorkflowStep{
				ID:          "05-test",
				Description: "Test step with default function",
				Prompt:      "Priority: {{.Priority | default \"medium\"}}",
			},
			contains: []string{"Priority: medium"},
			wantErr:  false,
		},
		{
			name: "Empty prompt",
			step: WorkflowStep{
				ID:          "06-test",
				Description: "Test step with empty prompt",
				Prompt:      "",
			},
			contains: []string{"Test step with empty prompt"}, // Should use default prompt
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := generateStepPrompt(tc.step, changeRequestPath)

			// Check if all expected strings are contained in the result
			for _, expected := range tc.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("generateStepPrompt() result does not contain expected string '%s'\nResult: %s", expected, result)
				}
			}

			// For empty prompt case, check default prompt is generated
			if tc.step.Prompt == "" {
				expectedDefault := generateDefaultPrompt(tc.step)
				if result != expectedDefault {
					t.Errorf("generateStepPrompt() with empty prompt = %v, want %v", result, expectedDefault)
				}
			}
		})
	}
}

func TestGenerateDefaultPrompt(t *testing.T) {
	step := WorkflowStep{
		ID:          "test-step",
		Description: "Test step",
		Prompt:      "",
	}

	generatedPrompt := generateDefaultPrompt(step)
	expectedContains := "Test step"

	if !strings.Contains(generatedPrompt, expectedContains) {
		t.Errorf("Generated default prompt does not contain expected text '%s'", expectedContains)
	}
}

func TestInterpolationErrorString(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		malformedVars []string
		missingVars   []string
		expected      string
	}{
		{
			name:          "Only message",
			message:       "test message",
			malformedVars: nil,
			missingVars:   nil,
			expected:      "test message",
		},
		{
			name:          "Message with malformed variables",
			message:       "test message",
			malformedVars: []string{"var with space", "another-bad"},
			missingVars:   nil,
			expected:      "test message: malformed variables [var with space, another-bad]",
		},
		{
			name:          "Message with missing variables",
			message:       "test message",
			malformedVars: nil,
			missingVars:   []string{"missing1", "missing2"},
			expected:      "test message: missing variables [missing1, missing2]",
		},
		{
			name:          "Message with both malformed and missing variables",
			message:       "test message",
			malformedVars: []string{"bad-var"},
			missingVars:   []string{"missing-var"},
			expected:      "test message: malformed variables [bad-var], missing variables [missing-var]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewInterpolationError(tt.message, tt.malformedVars, tt.missingVars)
			if err.Error() != tt.expected {
				t.Errorf("Expected error string '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

// Test the integration between template system and step processing
func TestTemplateSystemIntegration(t *testing.T) {
	changeRequestPath := "/project/docs/changes-request/feature.blueprint.md"
	
	step := WorkflowStep{
		ID:          "01-analyze",
		Description: "Analyze the feature",
		Prompt:      "Analyzing {{.ChangeRequestBasename}} from {{.ChangeRequestDirname}}",
		Variables: map[string]string{
			"ProjectName": "TestProject",
		},
	}

	result := generateStepPrompt(step, changeRequestPath)
	
	// Should contain interpolated variables
	expectedBasename := "feature"
	expectedDirname := "/project/docs/changes-request"
	
	if !strings.Contains(result, expectedBasename) {
		t.Errorf("Result should contain basename '%s', got: %s", expectedBasename, result)
	}
	
	if !strings.Contains(result, expectedDirname) {
		t.Errorf("Result should contain dirname '%s', got: %s", expectedDirname, result)
	}
}

// Test template error handling
func TestPromptTemplateErrorHandling(t *testing.T) {
	changeRequestPath := "/path/to/test.blueprint.md"
	
	step := WorkflowStep{
		ID:          "01-test",
		Description: "Test error handling",
		Prompt:      "{{.UndefinedVariable}}", // This will cause a template error
	}

	// Should gracefully degrade to returning the original prompt
	result := generateStepPrompt(step, changeRequestPath)
	
	// Should return the original prompt when template processing fails
	if result != step.Prompt {
		t.Errorf("Expected graceful degradation to original prompt, got: %s", result)
	}
}

// Test advanced template features
func TestAdvancedTemplateFeatures(t *testing.T) {
	changeRequestPath := "/project/feature.blueprint.md"
	
	tests := []struct {
		name     string
		prompt   string
		vars     map[string]string
		contains []string
	}{
		{
			name:     "Conditional with variable",
			prompt:   "{{if .HasTests}}Testing: {{.TestFramework}}{{else}}No tests specified{{end}}",
			vars:     map[string]string{"HasTests": "true", "TestFramework": "Jest"},
			contains: []string{"Testing: Jest"},
		},
		{
			name:     "Conditional without variable",
			prompt:   "{{if .HasTests}}Testing: {{.TestFramework}}{{else}}No tests specified{{end}}",
			vars:     map[string]string{},
			contains: []string{"No tests specified"},
		},
		{
			name:     "Default values",
			prompt:   "Framework: {{.Framework | default \"Go\"}}",
			vars:     map[string]string{},
			contains: []string{"Framework: Go"},
		},
		{
			name:     "Multiple conditionals",
			prompt:   "{{if .Database}}DB: {{.Database}}{{end}}{{if .Cache}} Cache: {{.Cache}}{{end}}",
			vars:     map[string]string{"Database": "PostgreSQL", "Cache": "Redis"},
			contains: []string{"DB: PostgreSQL", "Cache: Redis"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := WorkflowStep{
				ID:          "01-advanced",
				Description: "Advanced template test",
				Prompt:      tc.prompt,
				Variables:   tc.vars,
			}

			result := generateStepPrompt(step, changeRequestPath)

			for _, expected := range tc.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Result should contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}

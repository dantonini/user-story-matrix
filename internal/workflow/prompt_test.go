// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"path/filepath"
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

func TestInterpolatePrompt(t *testing.T) {
	// Setup test data
	changeRequestPath := "/path/to/my-change-request.blueprint.md"
	dir := filepath.Dir(changeRequestPath)
	base := strings.TrimSuffix(filepath.Base(changeRequestPath), ".blueprint.md")
	fullpath := filepath.Join(dir, base)
	stepID := "01-test-step"
	stepName := "test-step"
	
	// Create variables for testing
	vars := PromptVariables{
		ChangeRequestFilePath: changeRequestPath,
		ChangeRequestBasename: base,
		BlueprintBasename:     base,
		ChangeRequestDirname:  dir,
		StepID:                stepID,
		StepName:              stepName,
		ChangeRequestFullpath: fullpath,
		Basename:              base, // Deprecated
	}
	
	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{
			name:     "Simple variable substitution",
			prompt:   "The change request is ${change_request_file_path}",
			expected: "The change request is " + changeRequestPath,
		},
		{
			name:     "Multiple variables",
			prompt:   "Processing step ${stepid} (${stepname}) for ${change_request_basename}",
			expected: "Processing step " + stepID + " (" + stepName + ") for " + base,
		},
		{
			name:     "Basename variable (deprecated)",
			prompt:   "Working on ${basename}",
			expected: "Working on " + base,
		},
		{
			name:     "Mix of path and prompt variables",
			prompt:   "Output will be saved to ${dirname}/results/${stepname}.md",
			expected: "Output will be saved to " + dir + "/results/" + stepName + ".md",
		},
		{
			name:     "All variables in one prompt",
			prompt:   "${change_request_file_path} ${change_request_basename} ${blueprint_basename} ${dirname} ${stepid} ${stepname} ${fullpath} ${basename}",
			expected: changeRequestPath + " " + base + " " + base + " " + dir + " " + stepID + " " + stepName + " " + fullpath + " " + base,
		},
		{
			name:     "No variables",
			prompt:   "A prompt with no variables",
			expected: "A prompt with no variables",
		},
		{
			name:     "Unknown variable",
			prompt:   "This ${unknown_var} should remain unchanged",
			expected: "This ${unknown_var} should remain unchanged",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := InterpolatePrompt(tc.prompt, vars)
			if result != tc.expected {
				t.Errorf("InterpolatePrompt() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestInterpolatePromptWithMissingVars(t *testing.T) {
	// Test handling of missing variables
	prompt := "Process ${nonexistent_var} and ${change_request_file_path} and ${another_missing_var}"
	vars := PromptVariables{
		ChangeRequestFilePath: "/path/to/file",
	}
	
	expectedResult := "Process ${nonexistent_var} and /path/to/file and ${another_missing_var}"
	expectedMissingVars := []string{"nonexistent_var", "another_missing_var"}
	
	result, missingVars := InterpolatePromptWithMissingVars(prompt, vars)
	
	if result != expectedResult {
		t.Errorf("Expected result '%s', got '%s'", expectedResult, result)
	}
	
	if len(missingVars) != len(expectedMissingVars) {
		t.Errorf("Expected %d missing variables, got %d", len(expectedMissingVars), len(missingVars))
	}
	
	// Check that all expected missing variables are in the result
	for _, expected := range expectedMissingVars {
		found := false
		for _, actual := range missingVars {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected missing variable '%s' not found in result", expected)
		}
	}
}

func TestInterpolatePromptWithMap(t *testing.T) {
	// Test with extended variables structure using a map
	prompt := "Process ${change_request_file_path} with ${new_variable}"
	
	// Create variable map
	varMap := map[string]string{
		"change_request_file_path": "/path",
		"new_variable":             "test",
	}
	
	expected := "Process /path with test"
	result := interpolatePromptWithMap(prompt, varMap)
	
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test with nested map containing complex variables
	complexVarMap := map[string]string{
		"change_request_file_path": "/path",
		"user_name":                "john",
		"project_id":               "123",
	}
	
	complexPrompt := "User ${user_name} is working on project ${project_id} using ${change_request_file_path}"
	complexExpected := "User john is working on project 123 using /path"
	complexResult := interpolatePromptWithMap(complexPrompt, complexVarMap)
	
	if complexResult != complexExpected {
		t.Errorf("Expected '%s', got '%s'", complexExpected, complexResult)
	}
}

func TestInterpolatePromptWithError(t *testing.T) {
	// Test cases for InterpolatePromptWithError
	tests := []struct {
		name        string
		prompt      string
		vars        PromptVariables
		shouldError bool
	}{
		{
			name:   "Valid interpolation",
			prompt: "The file is ${change_request_file_path}",
			vars: PromptVariables{
				ChangeRequestFilePath: "/path/to/file.md",
			},
			shouldError: false,
		},
		{
			name:   "Missing variable",
			prompt: "The file is ${change_request_basename}",
			vars: PromptVariables{
				ChangeRequestFilePath: "/path/to/file.md",
			},
			shouldError: true,
		},
		{
			name:   "Malformed variable",
			prompt: "The file is ${change request file path}",
			vars: PromptVariables{
				ChangeRequestFilePath: "/path/to/file.md",
			},
			shouldError: true,
		},
		{
			name:   "Unclosed variable",
			prompt: "The file is ${unclosed",
			vars: PromptVariables{
				ChangeRequestFilePath: "/path/to/file.md",
			},
			shouldError: true,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := InterpolatePromptWithError(tc.prompt, tc.vars)
			if (err != nil) != tc.shouldError {
				t.Errorf("InterpolatePromptWithError() error = %v, shouldError %v", err, tc.shouldError)
			}
		})
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
			name:          "Valid prompt",
			prompt:        "This is a valid prompt with ${change_request_file_path}",
			shouldBeValid: true,
			errorContains: "",
		},
		{
			name:          "Valid prompt with multiple variables",
			prompt:        "The change request is ${change_request_file_path}. The basename is ${change_request_basename}.",
			shouldBeValid: true,
			errorContains: "",
		},
		{
			name:          "Invalid prompt with malformed variable",
			prompt:        "This prompt has an ${invalid variable} with spaces",
			shouldBeValid: false,
			errorContains: "prompt contains malformed variables",
		},
		{
			name:          "Invalid prompt with unclosed variable",
			prompt:        "This prompt has an ${unclosed variable",
			shouldBeValid: false,
			errorContains: "unclosed variable",
		},
		{
			name:          "Empty prompt",
			prompt:        "",
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
		expected string
		wantErr  bool
	}{
		{
			name: "Standard prompt",
			step: WorkflowStep{
				ID:          "01-test",
				Description: "Test step",
				Prompt:      "This is a test prompt with ${change_request_file_path}",
			},
			expected: "This is a test prompt with " + changeRequestPath,
			wantErr:  false,
		},
		{
			name: "Prompt with multiple variables",
			step: WorkflowStep{
				ID:          "02-test",
				Description: "Test step with multiple variables",
				Prompt:      "Processing ${change_request_file_path} with step ${stepid}",
			},
			expected: "Processing " + changeRequestPath + " with step 02-test",
			wantErr:  false,
		},
		{
			name: "Invalid prompt",
			step: WorkflowStep{
				ID:          "03-test",
				Description: "Test step with invalid prompt",
				Prompt:      "This has an ${invalid variable}",
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "Empty prompt",
			step: WorkflowStep{
				ID:          "04-test",
				Description: "Test step with empty prompt",
				Prompt:      "",
			},
			expected: "",
			wantErr:  false,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test function
			if tc.wantErr {
				// For error cases, check ValidatePrompt returns an error
				err := ValidatePrompt(tc.step.Prompt)
				if err == nil {
					t.Errorf("Expected ValidatePrompt to return an error for invalid prompt")
				}
			} else {
				result := generateStepPrompt(tc.step, changeRequestPath)
				
				// For non-error cases with explicit expected values
				if tc.expected != "" && result != tc.expected {
					t.Errorf("generateStepPrompt() = %v, want %v", result, tc.expected)
				}
				
				// For empty prompt case, check default prompt is generated
				if tc.step.Prompt == "" {
					expectedDefault := generateDefaultPrompt(tc.step)
					if result != expectedDefault {
						t.Errorf("generateStepPrompt() with empty prompt = %v, want %v", result, expectedDefault)
					}
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
		name         string
		message      string
		malformedVars []string
		missingVars  []string
		expected     string
	}{
		{
			name:         "Only message",
			message:      "test message",
			malformedVars: nil,
			missingVars:  nil,
			expected:     "test message",
		},
		{
			name:         "Message with malformed variables",
			message:      "test message",
			malformedVars: []string{"var with space", "another-bad"},
			missingVars:  nil,
			expected:     "test message: malformed variables [var with space, another-bad]",
		},
		{
			name:         "Message with missing variables",
			message:      "test message",
			malformedVars: nil,
			missingVars:  []string{"missing1", "missing2"},
			expected:     "test message: missing variables [missing1, missing2]",
		},
		{
			name:         "Message with both malformed and missing variables",
			message:      "test message",
			malformedVars: []string{"bad-var"},
			missingVars:  []string{"missing-var"},
			expected:     "test message: malformed variables [bad-var], missing variables [missing-var]",
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

func BenchmarkInterpolation(b *testing.B) {
	// Generate a large prompt with many variable references
	var prompt strings.Builder
	for i := 0; i < 1000; i++ {
		prompt.WriteString(fmt.Sprintf("This is sentence %d with ${change_request_file_path} variable reference.\n", i))
	}
	
	largePath := "/very/long/path/to/a/file/with/a/lot/of/segments/that/might/slow/down/string/operations/in/a/large/text.md"
	vars := PromptVariables{
		ChangeRequestFilePath: largePath,
	}
	
	b.ResetTimer()
	
	// Benchmark InterpolatePrompt
	b.Run("InterpolatePrompt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			InterpolatePrompt(prompt.String(), vars)
		}
	})
	
	// Benchmark InterpolatePromptWithMissingVars
	b.Run("InterpolatePromptWithMissingVars", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			InterpolatePromptWithMissingVars(prompt.String(), vars)
		}
	})
	
	// Benchmark InterpolatePromptWithError
	b.Run("InterpolatePromptWithError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = InterpolatePromptWithError(prompt.String(), vars)
		}
	})
	
	// Create a large map of variables
	varMap := map[string]string{
		"change_request_file_path": largePath,
	}
	for i := 0; i < 50; i++ {
		varMap[fmt.Sprintf("var_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	
	// Benchmark interpolatePromptWithMap
	b.Run("interpolatePromptWithMap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			interpolatePromptWithMap(prompt.String(), varMap)
		}
	})
} 
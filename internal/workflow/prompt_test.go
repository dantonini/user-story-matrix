// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"strings"
	"testing"
)

func TestWorkflowStepStructure(t *testing.T) {
	// Test that WorkflowStep struct includes the Prompt field
	step := WorkflowStep{
		ID:          "test-id",
		Description: "Test description",
		Prompt:      "Test prompt",
		OutputFile:  "test-output.md",
	}

	// Verify prompt field can be set and retrieved
	if step.Prompt != "Test prompt" {
		t.Errorf("Expected Prompt field to be 'Test prompt', got '%s'", step.Prompt)
	}
}

func TestInterpolatePrompt(t *testing.T) {
	// Test basic variable interpolation
	prompt := "Process the file at ${change_request_file_path}"
	vars := PromptVariables{
		ChangeRequestFilePath: "/path/to/file",
	}
	
	expected := "Process the file at /path/to/file"
	result := InterpolatePrompt(prompt, vars)
	
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test with multiple occurrences of the same variable
	prompt = "Path: ${change_request_file_path}, use ${change_request_file_path} for processing"
	expected = "Path: /path/to/file, use /path/to/file for processing"
	result = InterpolatePrompt(prompt, vars)
	
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
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
		"timestamp":                "2025-04-01",
	}
	
	complexPrompt := "User ${user_name} is working on project ${project_id} at ${timestamp} using ${change_request_file_path}"
	complexExpected := "User john is working on project 123 at 2025-04-01 using /path"
	complexResult := interpolatePromptWithMap(complexPrompt, complexVarMap)
	
	if complexResult != complexExpected {
		t.Errorf("Expected '%s', got '%s'", complexExpected, complexResult)
	}
}

func TestInterpolatePromptWithError(t *testing.T) {
	// Test with missing variables
	prompt := "Process ${nonexistent_var} and ${change_request_file_path}"
	vars := PromptVariables{
		ChangeRequestFilePath: "/path/to/file",
	}
	
	result, err := InterpolatePromptWithError(prompt, vars)
	expected := "Process ${nonexistent_var} and /path/to/file"
	
	if result != expected {
		t.Errorf("Expected result '%s', got '%s'", expected, result)
	}
	
	if err == nil {
		t.Error("Expected error for missing variables, got nil")
	} else {
		ierr, ok := err.(*InterpolationError)
		if !ok {
			t.Errorf("Expected *InterpolationError, got %T", err)
		} else {
			if len(ierr.MissingVars) != 1 || ierr.MissingVars[0] != "nonexistent_var" {
				t.Errorf("Expected missing variable 'nonexistent_var', got %v", ierr.MissingVars)
			}
		}
	}
	
	// Test with malformed variables
	malformedPrompt := "Process ${var with spaces} and ${incomplete"
	_, err = InterpolatePromptWithError(malformedPrompt, vars)
	
	if err == nil {
		t.Error("Expected error for malformed variables, got nil")
	} else {
		ierr, ok := err.(*InterpolationError)
		if !ok {
			t.Errorf("Expected *InterpolationError, got %T", err)
		} else {
			if len(ierr.MalformedVars) == 0 {
				t.Error("Expected malformed variables, got none")
			}
		}
	}
	
	// Test with valid prompt
	validPrompt := "Process ${change_request_file_path}"
	result, err = InterpolatePromptWithError(validPrompt, vars)
	expected = "Process /path/to/file"
	
	if result != expected {
		t.Errorf("Expected result '%s', got '%s'", expected, result)
	}
	
	if err != nil {
		t.Errorf("Expected no error for valid prompt, got %v", err)
	}
}

func TestValidatePrompt(t *testing.T) {
	// Test with valid prompt
	validPrompt := "Process ${change_request_file_path} and ${another_var}"
	err := ValidatePrompt(validPrompt)
	
	if err != nil {
		t.Errorf("Expected no error for valid prompt, got %v", err)
	}
	
	// Test with malformed variables
	malformedPrompt := "Process ${var with spaces}"
	err = ValidatePrompt(malformedPrompt)
	
	if err == nil {
		t.Error("Expected error for malformed variables, got nil")
	} else {
		ierr, ok := err.(*InterpolationError)
		if !ok {
			t.Errorf("Expected *InterpolationError, got %T", err)
		} else {
			if len(ierr.MalformedVars) == 0 {
				t.Error("Expected malformed variables, got none")
			}
		}
	}
	
	// Test with unclosed variable
	unclosedPrompt := "Process ${incomplete"
	err = ValidatePrompt(unclosedPrompt)
	
	if err == nil {
		t.Error("Expected error for unclosed variable, got nil")
	}
}

func TestGenerateStepPrompt(t *testing.T) {
	// Test with a step that has a prompt
	stepWithPrompt := WorkflowStep{
		ID:          "test-id",
		Description: "Test description",
		Prompt:      "Process the file at ${change_request_file_path}",
		OutputFile:  "test-output.md",
	}
	
	expected := "Process the file at /path/to/file"
	result := generateStepPrompt(stepWithPrompt, "/path/to/file")
	
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test with a step that has no prompt
	stepWithoutPrompt := WorkflowStep{
		ID:          "test-id",
		Description: "Test description",
		Prompt:      "",
		OutputFile:  "test-output.md",
	}
	
	expected = "Please execute the following step in the workflow: Test description"
	result = generateStepPrompt(stepWithoutPrompt, "/path/to/file")
	
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestGenerateDefaultPrompt(t *testing.T) {
	// Test default prompt generation
	step := WorkflowStep{
		ID:          "test-id",
		Description: "Test description",
		OutputFile:  "test-output.md",
	}
	
	expected := "Please execute the following step in the workflow: Test description"
	result := generateDefaultPrompt(step)
	
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
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
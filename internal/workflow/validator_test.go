// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

// TestNewValidationResult tests the NewValidationResult function
func TestNewValidationResult(t *testing.T) {
	result := NewValidationResult()
	
	assert.True(t, result.IsValid(), "New validation result should be valid by default")
	assert.Empty(t, result.Errors, "New validation result should have no errors")
	assert.Empty(t, result.Warnings, "New validation result should have no warnings")
}

// TestValidationResult_AddError tests the AddError method
func TestValidationResult_AddError(t *testing.T) {
	result := NewValidationResult()
	
	// Add an error
	result.AddError("workflow name is required")
	
	assert.False(t, result.IsValid(), "Validation result should be invalid after adding an error")
	assert.Len(t, result.Errors, 1, "Validation result should have 1 error after adding one")
	
	// Add another error
	result.AddError("workflow description is required")
	
	assert.Len(t, result.Errors, 2, "Validation result should have 2 errors after adding two")
}

// TestValidationResult_AddWarning tests the AddWarning method
func TestValidationResult_AddWarning(t *testing.T) {
	result := NewValidationResult()
	
	// Add a warning
	result.AddWarning("Test warning")
	
	assert.True(t, result.IsValid(), "Validation result should remain valid after adding a warning")
	assert.Len(t, result.Warnings, 1, "Validation result should have 1 warning after adding one")
	
	// Add another warning
	result.AddWarning("Another warning")
	
	assert.Len(t, result.Warnings, 2, "Validation result should have 2 warnings after adding two")
}

// TestNewWorkflowValidator tests the NewWorkflowValidator function
func TestNewWorkflowValidator(t *testing.T) {
	fs := io.NewMockFileSystem()
	workflowPath := "/path/to/workflow"
	validator := NewWorkflowValidator(fs, workflowPath)
	
	assert.NotNil(t, validator, "NewWorkflowValidator should not return nil")
	assert.Equal(t, fs, validator.fs, "NewWorkflowValidator should set the file system correctly")
	assert.Equal(t, workflowPath, validator.workflowPath, "NewWorkflowValidator should set the workflow path correctly")
}

// TestValidateWorkflow tests the ValidateWorkflow function
func TestValidateWorkflow(t *testing.T) {
	// Setup mock filesystem
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	fs.MkdirAll(promptsDir, 0755)

	// Add a test prompt template
	promptPath := filepath.Join(promptsDir, "test-prompt.md")
	fs.WriteFile(promptPath, []byte("Hello {{.name}}!"), 0644)

	// Create a workflow validator
	validator := NewWorkflowValidator(fs, workflowDir)

	// Create test workflow
	tests := []struct {
		name        string
		workflow    *WorkflowDefinition
		wantErrors  int
		wantWarning int
	}{
		{
			name: "Valid workflow",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables:   map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   promptPath,
						},
					},
				},
			},
			wantErrors:  0,
			wantWarning: 0,
		},
		{
			name: "Missing name",
			workflow: &WorkflowDefinition{
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables:   map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   promptPath,
						},
					},
				},
			},
			wantErrors:  1,
			wantWarning: 0,
		},
		{
			name: "Missing description",
			workflow: &WorkflowDefinition{
				Name: "test-workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables:   map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   promptPath,
						},
					},
				},
			},
			wantErrors:  0,
			wantWarning: 1,
		},
		{
			name: "No steps",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps:       []WorkflowStep{},
			},
			wantErrors:  1,
			wantWarning: 0,
		},
		{
			name: "Step missing ID",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						Description: "Test step",
						Variables:   map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   promptPath,
						},
					},
				},
			},
			wantErrors:  1,
			wantWarning: 0,
		},
		{
			name: "Step missing description",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:        "step1",
						Variables: map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   promptPath,
						},
					},
				},
			},
			wantErrors:  0,
			wantWarning: 1,
		},
		{
			name: "Duplicate step IDs",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step 1",
						Variables:   map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   promptPath,
						},
					},
					{
						ID:          "step1", // Duplicate ID
						Description: "Test step 2",
						Variables:   map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   promptPath,
						},
					},
				},
			},
			wantErrors:  1,
			wantWarning: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run validation
			result, err := validator.ValidateWorkflow(tt.workflow)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantErrors, len(result.Errors), "Expected %d errors, got %d: %v", tt.wantErrors, len(result.Errors), result.Errors)
			assert.Equal(t, tt.wantWarning, len(result.Warnings), "Expected %d warnings, got %d: %v", tt.wantWarning, len(result.Warnings), result.Warnings)
		})
	}
}

// TestValidationError tests the ValidationError type
func TestValidationError(t *testing.T) {
	// Test with line number and fix
	err := &ValidationError{
		Message:    "Test error",
		File:       "workflow.yaml",
		LineNumber: 42,
		Fix:        "Fix this error",
	}
	
	errStr := err.Error()
	assert.Contains(t, errStr, "Test error")
	assert.Contains(t, errStr, "workflow.yaml")
	assert.Contains(t, errStr, "line 42")
	assert.Contains(t, errStr, "Fix: Fix this error")
	
	// Test with line number but no fix
	err = &ValidationError{
		Message:    "Test error",
		File:       "workflow.yaml",
		LineNumber: 42,
	}
	
	errStr = err.Error()
	assert.Contains(t, errStr, "Test error")
	assert.Contains(t, errStr, "workflow.yaml")
	assert.Contains(t, errStr, "line 42")
	assert.NotContains(t, errStr, "Fix:")
	
	// Test with no line number but with fix
	err = &ValidationError{
		Message: "Test error",
		File:    "workflow.yaml",
		Fix:     "Fix this error",
	}
	
	errStr = err.Error()
	assert.Contains(t, errStr, "Test error")
	assert.Contains(t, errStr, "workflow.yaml")
	assert.Contains(t, errStr, "Fix: Fix this error")
	assert.NotContains(t, errStr, "line")
	
	// Test with minimum info
	err = &ValidationError{
		Message: "Test error",
		File:    "workflow.yaml",
	}
	
	errStr = err.Error()
	assert.Contains(t, errStr, "Test error")
	assert.Contains(t, errStr, "workflow.yaml")
	assert.NotContains(t, errStr, "line")
	assert.NotContains(t, errStr, "Fix:")
}

// TestCheckForDefaultValue tests the checkForDefaultValue function
func TestCheckForDefaultValue(t *testing.T) {
	// Setup mock filesystem
	fs := io.NewMockFileSystem()

	// Add test prompt templates
	fs.AddFile("/prompt1.md", []byte("Hello {{.name}}!"))
	fs.AddFile("/prompt2.md", []byte("Hello {{.name | default \"Anonymous\"}}!"))
	fs.AddFile("/prompt3.md", []byte("Hello {{ .name | default 123 }}!"))
	
	// Test file read error
	mockFsWithErrors := io.NewMockFileSystemWithErrors()
	mockFsWithErrors.SetReadError("/error.md", assert.AnError)

	tests := []struct {
		name       string
		fs         io.FileSystem
		promptPath string
		varName    string
		want       bool
	}{
		{
			name:       "No default value",
			fs:         fs,
			promptPath: "/prompt1.md",
			varName:    "name",
			want:       false,
		},
		{
			name:       "Has string default value",
			fs:         fs,
			promptPath: "/prompt2.md",
			varName:    "name",
			want:       true,
		},
		{
			name:       "Has numeric default value",
			fs:         fs,
			promptPath: "/prompt3.md",
			varName:    "name",
			want:       true,
		},
		{
			name:       "File read error",
			fs:         mockFsWithErrors,
			promptPath: "/error.md",
			varName:    "name",
			want:       false,
		},
		{
			name:       "Regex compile error (unlikely but for coverage)",
			fs:         fs,
			promptPath: "/prompt1.md",
			varName:    "name[", // Invalid regex character will cause compile error
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkForDefaultValue(tt.fs, tt.promptPath, tt.varName)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestFindLineNumber tests the findLineNumber function
func TestFindLineNumber(t *testing.T) {
	lines := []string{
		"line 1",
		"line 2 with pattern",
		"line 3",
	}
	
	// Test finding existing pattern
	lineNum := findLineNumber(lines, "pattern")
	assert.Equal(t, 2, lineNum, "Should find pattern in line 2")
	
	// Test not finding pattern
	lineNum = findLineNumber(lines, "nonexistent")
	assert.Equal(t, 0, lineNum, "Should return 0 for non-existent pattern")
}

// TestFindLineNumberAfter tests the findLineNumberAfter function
func TestFindLineNumberAfter(t *testing.T) {
	lines := []string{
		"line 1 with pattern",
		"line 2",
		"line 3 with pattern",
	}
	
	// Test finding pattern after line 1
	lineNum := findLineNumberAfter(lines, "pattern", 1)
	assert.Equal(t, 3, lineNum, "Should find pattern in line 3")
	
	// Test finding pattern from beginning
	lineNum = findLineNumberAfter(lines, "pattern", 0)
	assert.Equal(t, 1, lineNum, "Should find pattern in line 1")
	
	// Test not finding pattern after specified line
	lineNum = findLineNumberAfter(lines, "pattern", 3)
	assert.Equal(t, 0, lineNum, "Should return 0 when pattern is not found after specified line")
}

// TestValidatePromptTemplates tests the validatePromptTemplates function
func TestValidatePromptTemplates(t *testing.T) {
	// Setup mock filesystem
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	fs.MkdirAll(promptsDir, 0755)

	// Add test prompt templates
	validPromptPath := filepath.Join(promptsDir, "valid-prompt.md")
	fs.WriteFile(validPromptPath, []byte("Hello {{.name}}!"), 0644)

	invalidPromptPath := filepath.Join(promptsDir, "invalid-prompt.md")
	fs.WriteFile(invalidPromptPath, []byte("Hello {{.name!"), 0644) // Invalid syntax

	// Create a workflow validator
	validator := NewWorkflowValidator(fs, workflowDir)

	// Create test workflows
	tests := []struct {
		name        string
		workflow    *WorkflowDefinition
		wantErrors  int
		wantWarning int
	}{
		{
			name: "Valid template",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables:   map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   validPromptPath,
						},
					},
				},
			},
			wantErrors:  0,
			wantWarning: 0,
		},
		{
			name: "Invalid template syntax",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables:   map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   invalidPromptPath,
						},
					},
				},
			},
			wantErrors:  2, // Expecting two errors: one for invalid template and one for failed extraction
			wantWarning: 0,
		},
		{
			name: "Missing variable",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables:   map[string]string{}, // Missing 'name' variable
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   validPromptPath,
						},
					},
				},
			},
			wantErrors:  0,
			wantWarning: 1, // Warning for missing variable
		},
		{
			name: "Unused variable",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables: map[string]string{
							"name":    "World",
							"unused": "Not used",
						},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   validPromptPath,
						},
					},
				},
			},
			wantErrors:  0,
			wantWarning: 1, // Warning for unused variable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new validation result
			result := NewValidationResult()

			// Run validation
			err := validator.validatePromptTemplates(tt.workflow, result)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantErrors, len(result.Errors), "Expected %d errors, got %d: %v", tt.wantErrors, len(result.Errors), result.Errors)
			assert.Equal(t, tt.wantWarning, len(result.Warnings), "Expected %d warnings, got %d: %v", tt.wantWarning, len(result.Warnings), result.Warnings)
		})
	}
}

// TestValidateVariableReferences tests the ValidateVariableReferences function
func TestValidateVariableReferences(t *testing.T) {
	// Setup mock filesystem
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	fs.MkdirAll(promptsDir, 0755)

	// Add test prompt templates
	simplePromptPath := filepath.Join(promptsDir, "simple-prompt.md")
	fs.WriteFile(simplePromptPath, []byte("Hello {{.name}}!"), 0644)

	defaultValPromptPath := filepath.Join(promptsDir, "default-val-prompt.md")
	fs.WriteFile(defaultValPromptPath, []byte("Hello {{.name | default \"Anonymous\"}}!"), 0644)

	multiVarPromptPath := filepath.Join(promptsDir, "multi-var-prompt.md")
	fs.WriteFile(multiVarPromptPath, []byte("Hello {{.firstName}} {{.lastName}}!"), 0644)

	// Create test workflows
	tests := []struct {
		name          string
		workflow      *WorkflowDefinition
		wantErrCount  int
		wantErrPrefix string
	}{
		{
			name: "All variables provided",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables:   map[string]string{"name": "World"},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   simplePromptPath,
						},
					},
				},
			},
			wantErrCount: 0,
		},
		{
			name: "Missing variable",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables:   map[string]string{},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   simplePromptPath,
						},
					},
				},
			},
			wantErrCount:  1,
			wantErrPrefix: "step 'step1' uses variables", // Error for missing variables
		},
		{
			name: "Multiple missing variables",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables:   map[string]string{},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   multiVarPromptPath,
						},
					},
				},
			},
			wantErrCount:  1,
			wantErrPrefix: "step 'step1' uses variables", // Error for missing variables
		},
		{
			name: "Unused variables",
			workflow: &WorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test workflow",
				Steps: []WorkflowStep{
					{
						ID:          "step1",
						Description: "Test step",
						Variables: map[string]string{
							"name":     "World",
							"unused1":  "Value1",
							"unused2":  "Value2",
						},
						source: promptSource{
							sourceType: promptSourceFile,
							filePath:   simplePromptPath,
						},
					},
				},
			},
			wantErrCount:  1, // Warnings are also added to errors
			wantErrPrefix: "variables [unused", // Warning for unused variables
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run validation
			errors, err := ValidateVariableReferences(fs, workflowDir, tt.workflow)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantErrCount, len(errors), "Expected %d errors, got %d", tt.wantErrCount, len(errors))
			
			if tt.wantErrCount > 0 && len(errors) > 0 {
				assert.Contains(t, errors[0].Error(), tt.wantErrPrefix, "Error message should contain: %s", tt.wantErrPrefix)
			}
		})
	}
}

// TestValidationErrorString tests the Error method of ValidationError
func TestValidationErrorString(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "With line number and fix",
			err: &ValidationError{
				Message:    "Test error",
				File:       "test.yaml",
				LineNumber: 42,
				Fix:        "Fix this error",
			},
			expected: "Test error in test.yaml (line 42) - Fix: Fix this error",
		},
		{
			name: "With line number, no fix",
			err: &ValidationError{
				Message:    "Test error",
				File:       "test.yaml",
				LineNumber: 42,
			},
			expected: "Test error in test.yaml (line 42)",
		},
		{
			name: "Without line number, with fix",
			err: &ValidationError{
				Message: "Test error",
				File:    "test.yaml",
				Fix:     "Fix this error",
			},
			expected: "Test error in test.yaml - Fix: Fix this error",
		},
		{
			name: "Without line number, without fix",
			err: &ValidationError{
				Message: "Test error",
				File:    "test.yaml",
			},
			expected: "Test error in test.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

// TestValidateWorkflowWithDefaultValues tests that the validator correctly handles default values in templates
func TestValidateWorkflowWithDefaultValues(t *testing.T) {
	// Setup mock filesystem
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	fs.MkdirAll(promptsDir, 0755)

	// Add test prompt templates with clearer separation of test cases
	promptWithDefaultsPath := filepath.Join(promptsDir, "with-defaults.md")
	templateContent := `
# Template with variables
- Variable with value: {{.key1}}
- Variable without value or default: {{.key2}}
- Variable with default: {{.optional_var | default "default value"}}
`
	fs.WriteFile(promptWithDefaultsPath, []byte(templateContent), 0644)

	// Create a workflow validator
	validator := NewWorkflowValidator(fs, workflowDir)

	// Create test workflow
	workflow := &WorkflowDefinition{
		Name:        "test-workflow",
		Description: "Test workflow with default values",
		Steps: []WorkflowStep{
			{
				ID:          "step1",
				Description: "Test step with defaults",
				Variables:   map[string]string{"key1": "value1"},
				source: promptSource{
					sourceType: promptSourceFile,
					filePath:   promptWithDefaultsPath,
				},
			},
		},
	}

	// Test the checkForDefaultValue function directly first
	hasDefaultKey1 := checkForDefaultValue(fs, promptWithDefaultsPath, "key1")
	hasDefaultKey2 := checkForDefaultValue(fs, promptWithDefaultsPath, "key2")
	hasDefaultOptional := checkForDefaultValue(fs, promptWithDefaultsPath, "optional_var")
	
	t.Logf("Direct function test - hasDefaultKey1: %v, hasDefaultKey2: %v, hasDefaultOptional: %v", 
		hasDefaultKey1, hasDefaultKey2, hasDefaultOptional)
	
	// Validate workflow
	result, err := validator.ValidateWorkflow(workflow)
	
	// Print debug information
	t.Logf("Validation result - Errors: %v, Warnings: %v", result.Errors, result.Warnings)
	
	// Assertions
	assert.NoError(t, err)
	
	// The validation should succeed with warnings
	assert.True(t, result.IsValid(), "Result should be valid (no errors) despite warnings")
	assert.Greater(t, len(result.Warnings), 0, "There should be at least one warning")
	
	// Check that key2 is reported as missing but optional_var is not (due to default value)
	foundMissingKey2Warning := false
	foundOptionalVarWarning := false
	
	for _, warning := range result.Warnings {
		t.Logf("Warning: %s", warning)
		if strings.Contains(warning, "key2") {
			foundMissingKey2Warning = true
		}
		if strings.Contains(warning, "optional_var") {
			foundOptionalVarWarning = true
		}
	}
	
	assert.True(t, foundMissingKey2Warning, "Warning for missing key2 should be reported")
	assert.False(t, foundOptionalVarWarning, "Warning for optional_var with default should not be reported")
}

// TestValidateMultiStepWorkflow reproduces a real-world issue where workflow validation
// doesn't properly detect missing variables in multi-step workflows
func TestValidateMultiStepWorkflow(t *testing.T) {
	// Setup mock filesystem
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	fs.MkdirAll(promptsDir, 0755)

	// Add test prompt templates resembling the actual issue
	step1PromptPath := filepath.Join(promptsDir, "step1.md")
	step1Content := `
# Step 1: First Step

This is the first step. You can use variable substitution with {{ .key1 }} and {{ .key2 }}.

You can also use default values: {{ .optional_var | default "default value" }}.
`
	fs.WriteFile(step1PromptPath, []byte(step1Content), 0644)

	step2PromptPath := filepath.Join(promptsDir, "step2.md")
	step2Content := `
# Step 2: Second Step

This step uses {{ .key1 }} and {{ .key2 }}.
`
	fs.WriteFile(step2PromptPath, []byte(step2Content), 0644)

	// Create a workflow that resembles the problematic workflow 
	// where key2 is referenced in step1 but only defined in step2
	workflow := &WorkflowDefinition{
		Name:        "multi-step-workflow",
		Description: "Test multi-step workflow with missing variable reference",
		Steps: []WorkflowStep{
			{
				ID:          "01-step-one",
				Description: "First step",
				Variables:   map[string]string{"key1": "value1", "keyA": "value2"},
				source: promptSource{
					sourceType: promptSourceFile,
					filePath:   step1PromptPath,
				},
			},
			{
				ID:          "02-step-two",
				Description: "Second step",
				Variables:   map[string]string{"key1": "value1", "key2": "value2"},
				source: promptSource{
					sourceType: promptSourceFile,
					filePath:   step2PromptPath,
				},
			},
		},
	}

	// Create a workflow validator
	validator := NewWorkflowValidator(fs, workflowDir)
	
	// Validate workflow
	result, err := validator.ValidateWorkflow(workflow)
	
	// Print debug information
	t.Logf("Validation result - Errors: %v, Warnings: %v", result.Errors, result.Warnings)
	
	// Assertions
	assert.NoError(t, err, "Validation should not return an error")
	assert.True(t, result.IsValid(), "Result should be valid (warnings don't make it invalid)")
	
	// We expect to find a warning about key2 missing in step 1
	foundMissingKey2Warning := false
	for _, warning := range result.Warnings {
		t.Logf("Warning: %s", warning)
		if strings.Contains(warning, "key2") && strings.Contains(warning, "01-step-one") {
			foundMissingKey2Warning = true
			break
		}
	}
	
	assert.True(t, foundMissingKey2Warning, "Warning for missing key2 in step 01-step-one should be reported")
}

// TestValidateWorkflowFromYAML reproduces the issue with workflow validation not detecting
// missing variables when loading from a YAML file
func TestValidateWorkflowFromYAML(t *testing.T) {
	// Setup mock filesystem
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	fs.MkdirAll(promptsDir, 0755)

	// Create a prompt file that references a missing variable
	step1PromptPath := filepath.Join(promptsDir, "step1.md")
	step1Content := `This step uses {{.key1}} and {{.key2}}.`
	fs.WriteFile(step1PromptPath, []byte(step1Content), 0644)

	// Create the workflow.yaml file with missing variable
	workflowYAMLPath := filepath.Join(workflowDir, "workflow.yaml")
	workflowYAMLContent := `name: "test-workflow"
description: "Test workflow with missing variable"
steps:
  - id: "step1"
    description: "First step"
    prompt: "prompts/step1.md"
    variables:
      key1: "value1"
      # key2 is missing intentionally
`
	fs.WriteFile(workflowYAMLPath, []byte(workflowYAMLContent), 0644)

	// Load the workflow from the YAML file
	workflowDef, err := LoadWorkflowFromFile(fs, workflowYAMLPath)
	assert.NoError(t, err, "Should load workflow from YAML")

	// Create validator
	validator := NewWorkflowValidator(fs, workflowDir)

	// Validate workflow
	result, err := validator.ValidateWorkflow(workflowDef)
	assert.NoError(t, err, "Validation should not error")

	// The validation should succeed (missing variable is just a warning)
	assert.True(t, result.IsValid(), "Result should be valid")

	// But it should generate a warning about the missing variable
	foundWarning := false
	for _, warning := range result.Warnings {
		t.Logf("Warning: %s", warning)
		if strings.Contains(warning, "key2") {
			foundWarning = true
			break
		}
	}

	// This assertion should pass when the bug is fixed
	assert.True(t, foundWarning, "Should find warning about missing key2 variable")
}

// TestValidateWorkflowRelativePaths tests that the validator reports paths correctly
// without introducing incorrect path prefixes like ../../../
func TestValidateWorkflowRelativePaths(t *testing.T) {
	// Setup mock filesystem
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	fs.MkdirAll(promptsDir, 0755)

	// Create a prompt file
	step1PromptPath := filepath.Join(promptsDir, "step1.md")
	step1Content := `This step uses {{.key1}} and {{.key2}}.`
	fs.WriteFile(step1PromptPath, []byte(step1Content), 0644)

	// Create workflow definition with relative prompt path
	workflow := &WorkflowDefinition{
		Name:        "test-workflow",
		Description: "Test workflow with relative path",
		Steps: []WorkflowStep{
			{
				ID:          "step1",
				Description: "First step",
				Variables:   map[string]string{"key1": "value1"},
				source: promptSource{
					sourceType: promptSourceFile,
					filePath:   "prompts/step1.md", // Relative path as would be stored in workflow.yaml
				},
			},
		},
	}

	// Create validator
	validator := NewWorkflowValidator(fs, workflowDir)

	// Validate workflow
	result, err := validator.ValidateWorkflow(workflow)
	assert.NoError(t, err, "Validation should not error")

	// There should be warnings about missing variables
	foundWarning := false
	for _, warning := range result.Warnings {
		t.Logf("Warning: %s", warning)
		// The warning should contain the proper relative path, not ../../../ etc.
		if strings.Contains(warning, "key2") {
			foundWarning = true
			assert.Contains(t, warning, "prompts/step1.md", "Warning should use the correct relative path")
			assert.NotContains(t, warning, "../", "Warning should not contain ../ path components")
		}
	}

	assert.True(t, foundWarning, "Should find warning about missing key2 variable")
} 
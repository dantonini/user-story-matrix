// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/user-story-matrix/usm/internal/io"
)

// ValidationResult represents the result of a workflow validation operation.
// It contains any errors or warnings found during validation.
type ValidationResult struct {
	// Errors contains validation errors that prevent the workflow from being used
	Errors []string
	
	// Warnings contains non-critical issues that should be addressed but don't prevent usage
	Warnings []string
}

// ValidationError represents a validation error with detailed context information
type ValidationError struct {
	// Message is the error description
	Message string
	
	// File is the file where the error occurred
	File string
	
	// LineNumber is the line number where the error occurred (0 if unknown)
	LineNumber int
	
	// Fix is a suggested fix for the error, if available
	Fix string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.LineNumber > 0 {
		if e.Fix != "" {
			return fmt.Sprintf("%s in %s (line %d) - Fix: %s", e.Message, e.File, e.LineNumber, e.Fix)
		}
		return fmt.Sprintf("%s in %s (line %d)", e.Message, e.File, e.LineNumber)
	}
	
	if e.Fix != "" {
		return fmt.Sprintf("%s in %s - Fix: %s", e.Message, e.File, e.Fix)
	}
	return fmt.Sprintf("%s in %s", e.Message, e.File)
}

// NewValidationResult creates a new empty validation result.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}
}

// AddError adds an error to the validation result.
func (vr *ValidationResult) AddError(err string) {
	vr.Errors = append(vr.Errors, err)
}

// AddWarning adds a warning to the validation result.
func (vr *ValidationResult) AddWarning(warning string) {
	vr.Warnings = append(vr.Warnings, warning)
}

// IsValid returns true if the validation result contains no errors.
func (vr *ValidationResult) IsValid() bool {
	return len(vr.Errors) == 0
}

// WorkflowValidator validates workflow definitions and prompt files.
type WorkflowValidator struct {
	fs           io.FileSystem
	workflowPath string
	renderer     *TemplateRenderer
}

// NewWorkflowValidator creates a new workflow validator.
func NewWorkflowValidator(fs io.FileSystem, workflowPath string) *WorkflowValidator {
	return &WorkflowValidator{
		fs:           fs,
		workflowPath: workflowPath,
		renderer:     NewTemplateRenderer(fs, workflowPath),
	}
}

// ValidateWorkflow performs comprehensive validation of a workflow
func (v *WorkflowValidator) ValidateWorkflow(workflow *WorkflowDefinition) (*ValidationResult, error) {
	result := NewValidationResult()

	// Validate basic workflow structure
	if workflow.Name == "" {
		result.AddError("workflow name is required")
	}

	if workflow.Description == "" {
		result.AddWarning("workflow description is missing")
	}

	if len(workflow.Steps) == 0 {
		result.AddError("workflow must have at least one step")
	}

	// Check for step IDs uniqueness
	stepIDs := make(map[string]bool)
	for i, step := range workflow.Steps {
		if step.ID == "" {
			result.AddError(fmt.Sprintf("step at index %d must have an ID", i))
		} else if stepIDs[step.ID] {
			result.AddError(fmt.Sprintf("duplicate step ID: %s", step.ID))
		} else {
			stepIDs[step.ID] = true
		}

		if step.Description == "" {
			result.AddWarning(fmt.Sprintf("step '%s' description is missing", step.ID))
		}
	}

	// Validate prompt templates
	err := v.validatePromptTemplates(workflow, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// validatePromptTemplates validates the templates referenced in a workflow
func (v *WorkflowValidator) validatePromptTemplates(workflow *WorkflowDefinition, result *ValidationResult) error {
	for _, step := range workflow.Steps {
		// Get prompt source and path
		promptSource := step.source
		
		if promptSource.sourceType == promptSourceFile {
			// Validate template
			promptPath := promptSource.filePath
			promptRelPath, err := filepath.Rel(v.workflowPath, promptPath)
			if err != nil {
				promptRelPath = promptPath
			}
			
			// Validate template syntax
			err = v.renderer.ValidateTemplate(promptPath)
			if err != nil {
				result.AddError(fmt.Sprintf("invalid template in '%s': %s", promptRelPath, err.Error()))
			}
			
			// Extract and validate variables
			variables, err := v.renderer.ExtractTemplateVariables(promptPath)
			if err != nil {
				result.AddError(fmt.Sprintf("failed to extract variables from '%s': %s", promptRelPath, err.Error()))
				continue
			}
			
			// Check if all extracted variables are provided
			for _, varName := range variables {
				if _, exists := step.Variables[varName]; !exists {
					// This is only a warning because the template might have default values
					result.AddWarning(fmt.Sprintf("step '%s' uses variable '%s' in template '%s' but it is not provided in step definition",
						step.ID, varName, promptRelPath))
				}
			}
			
			// Check if there are unused variables defined
			for varName := range step.Variables {
				found := false
				for _, extractedVar := range variables {
					if extractedVar == varName {
						found = true
						break
					}
				}
				
				if !found {
					result.AddWarning(fmt.Sprintf("variable '%s' is defined in step '%s' but not used in template '%s'",
						varName, step.ID, promptRelPath))
				}
			}
		}
	}
	
	return nil
}

// ValidateVariableReferences checks if all variables referenced in templates are provided.
func ValidateVariableReferences(fs io.FileSystem, workflowDir string, workflow *WorkflowDefinition) ([]error, error) {
	errors := make([]error, 0)
	warnings := make([]error, 0)
	
	renderer := NewTemplateRenderer(fs, workflowDir)
	
	for _, step := range workflow.Steps {
		// Get prompt source and path
		promptSource := step.source
		
		if promptSource.sourceType == promptSourceFile {
			// Extract variables from template
			promptPath := promptSource.filePath
			promptRelPath, err := filepath.Rel(workflowDir, promptPath)
			if err != nil {
				promptRelPath = promptPath
			}
			
			variables, err := renderer.ExtractTemplateVariables(promptPath)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to extract variables from '%s': %w", promptRelPath, err))
				continue
			}
			
			// Check if all extracted variables are provided
			missingVars := make([]string, 0)
			for _, varName := range variables {
				if _, exists := step.Variables[varName]; !exists {
					if !checkForDefaultValue(fs, promptPath, varName) {
						missingVars = append(missingVars, varName)
					}
				}
			}
			
			if len(missingVars) > 0 {
				sort.Strings(missingVars)
				errors = append(errors, fmt.Errorf("step '%s' uses variables %v in template '%s' but they are not provided in step definition",
					step.ID, missingVars, promptRelPath))
			}
			
			// Check if there are unused variables defined
			unusedVars := make([]string, 0)
			for varName := range step.Variables {
				found := false
				for _, extractedVar := range variables {
					if extractedVar == varName {
						found = true
						break
					}
				}
				
				if !found {
					unusedVars = append(unusedVars, varName)
				}
			}
			
			if len(unusedVars) > 0 {
				sort.Strings(unusedVars)
				warnings = append(warnings, fmt.Errorf("variables %v are defined in step '%s' but not used in template '%s'",
					unusedVars, step.ID, promptRelPath))
			}
		}
	}
	
	// Include warnings after errors for better readability
	if len(warnings) > 0 {
		errors = append(errors, warnings...)
	}
	
	return errors, nil
}

// checkForDefaultValue checks if a variable has a default value in the template.
func checkForDefaultValue(fs io.FileSystem, promptPath string, varName string) bool {
	data, err := fs.ReadFile(promptPath)
	if err != nil {
		return false
	}
	
	promptContent := string(data)
	
	// Look for default value pattern like {{.varName | default "value"}}
	// This is a simplified approach that doesn't handle all possible template formats
	pattern := fmt.Sprintf(`\{\{\s*\.%s\s*\|\s*default\s+`, regexp.QuoteMeta(varName))
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	
	return re.MatchString(promptContent)
}

// Helper function to find a line number in text
func findLineNumber(lines []string, pattern string) int {
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			return i + 1 // Line numbers are 1-based
		}
	}
	return 0 // Not found
}

// Helper function to find a line number after a certain line
func findLineNumberAfter(lines []string, pattern string, afterLine int) int {
	for i := afterLine; i < len(lines); i++ {
		if strings.Contains(lines[i], pattern) {
			return i + 1 // Line numbers are 1-based
		}
	}
	return 0 // Not found
} 
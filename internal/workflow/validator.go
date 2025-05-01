// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/user-story-matrix/usm/internal/io"
	"gopkg.in/yaml.v3"
)

// ValidationResult represents the result of a workflow validation.
type ValidationResult struct {
	// IsValid indicates whether the workflow is valid.
	IsValid bool
	
	// Errors contains all validation errors.
	Errors []error
	
	// Warnings contains non-critical issues found during validation.
	Warnings []string
}

// NewValidationResult creates a new validation result with default values.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		IsValid:  true,
		Errors:   make([]error, 0),
		Warnings: make([]string, 0),
	}
}

// AddError adds an error to the validation result and marks it as invalid.
func (r *ValidationResult) AddError(err error) {
	r.Errors = append(r.Errors, err)
	r.IsValid = false
}

// AddWarning adds a warning to the validation result.
func (r *ValidationResult) AddWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}

// WorkflowValidator validates workflow definitions and prompt templates.
type WorkflowValidator struct {
	fs io.FileSystem
}

// NewWorkflowValidator creates a new workflow validator.
func NewWorkflowValidator(fs io.FileSystem) *WorkflowValidator {
	return &WorkflowValidator{
		fs: fs,
	}
}

// ValidateWorkflow validates a workflow directory structure.
//
// Parameters:
//   - dirPath: Path to the workflow directory
//
// Returns:
//   - A ValidationResult containing validation status and errors/warnings
func (v *WorkflowValidator) ValidateWorkflow(dirPath string) *ValidationResult {
	result := NewValidationResult()
	
	// Check if directory exists
	if !v.fs.Exists(dirPath) {
		result.AddError(fmt.Errorf("workflow directory not found: %s", dirPath))
		return result
	}
	
	// Check if workflow.yaml exists
	workflowYAMLPath := filepath.Join(dirPath, WorkflowConfigFile)
	if !v.fs.Exists(workflowYAMLPath) {
		result.AddError(fmt.Errorf("workflow configuration file not found: %s", workflowYAMLPath))
		return result
	}
	
	// Check if prompts directory exists
	promptsDirPath := filepath.Join(dirPath, PromptsDir)
	if !v.fs.Exists(promptsDirPath) {
		result.AddError(fmt.Errorf("prompts directory not found: %s", promptsDirPath))
	}
	
	// Validate workflow configuration
	workflowErrors := validateWorkflowConfiguration(v.fs, dirPath)
	for _, err := range workflowErrors {
		result.AddError(err)
	}
	
	// Validate prompt templates
	promptErrors := v.validatePromptTemplates(dirPath)
	for _, err := range promptErrors {
		result.AddError(err)
	}
	
	// Validate variable references
	variableErrors := v.ValidateVariableReferences(dirPath)
	for _, err := range variableErrors {
		result.AddError(err)
	}
	
	return result
}

// validateWorkflowConfiguration validates the workflow.yaml file.
func validateWorkflowConfiguration(fs io.FileSystem, dirPath string) []error {
	errors := make([]error, 0)
	
	// Read the workflow.yaml file
	workflowYAMLPath := filepath.Join(dirPath, WorkflowConfigFile)
	workflowData, err := fs.ReadFile(workflowYAMLPath)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to read workflow configuration: %w", err))
		return errors
	}
	
	// Parse the workflow.yaml file
	var externalWorkflow ExternalWorkflowDefinition
	externalWorkflow, err = parseWorkflowYAML(workflowData)
	if err != nil {
		errors = append(errors, fmt.Errorf("invalid workflow configuration: %w", err))
		return errors
	}
	
	// Validate basic workflow structure
	if err := validateExternalWorkflow(&externalWorkflow); err != nil {
		errors = append(errors, err)
	}
	
	// Validate prompt references
	promptErrors := ValidateWorkflowPromptReferences(fs, dirPath, &externalWorkflow)
	errors = append(errors, promptErrors...)
	
	// Validate step IDs
	stepIDs := make(map[string]bool)
	for _, step := range externalWorkflow.Steps {
		if _, exists := stepIDs[step.ID]; exists {
			errors = append(errors, fmt.Errorf("duplicate step ID '%s'", step.ID))
		}
		stepIDs[step.ID] = true
	}
	
	return errors
}

// validatePromptTemplates validates all prompt templates in the prompts directory.
func (v *WorkflowValidator) validatePromptTemplates(dirPath string) []error {
	errors := make([]error, 0)
	
	// Check if prompts directory exists
	promptsDirPath := filepath.Join(dirPath, PromptsDir)
	if !v.fs.Exists(promptsDirPath) {
		// This error is already added in ValidateWorkflow
		return errors
	}
	
	// Create template renderer
	renderer := NewTemplateRenderer(v.fs, dirPath)
	
	// Read all files in prompts directory
	entries, err := v.fs.ReadDir(promptsDirPath)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to read prompts directory: %w", err))
		return errors
	}
	
	// Validate each prompt template
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		// Only validate markdown files
		fileName := entry.Name()
		if !strings.HasSuffix(strings.ToLower(fileName), ".md") {
			continue
		}
		
		// Validate template syntax
		promptPath := filepath.Join(PromptsDir, fileName)
		if err := renderer.ValidateTemplate(promptPath); err != nil {
			errors = append(errors, err)
		}
	}
	
	return errors
}

// Helper function to parse workflow YAML data
func parseWorkflowYAML(data []byte) (ExternalWorkflowDefinition, error) {
	var externalWorkflow ExternalWorkflowDefinition
	err := yaml.Unmarshal(data, &externalWorkflow)
	if err != nil {
		return ExternalWorkflowDefinition{}, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return externalWorkflow, nil
}

// ValidateVariableReferences checks whether all variables used in prompt templates
// are provided in step definitions.
func (v *WorkflowValidator) ValidateVariableReferences(dirPath string) []error {
	errors := make([]error, 0)
	
	// Read the workflow configuration
	workflowYAMLPath := filepath.Join(dirPath, WorkflowConfigFile)
	workflowData, err := v.fs.ReadFile(workflowYAMLPath)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to read workflow configuration: %w", err))
		return errors
	}
	
	// Parse the workflow configuration
	externalWorkflow, err := parseWorkflowYAML(workflowData)
	if err != nil {
		errors = append(errors, fmt.Errorf("invalid workflow configuration: %w", err))
		return errors
	}
	
	// Create template renderer
	renderer := NewTemplateRenderer(v.fs, dirPath)
	
	// Check each step for variable references
	for _, step := range externalWorkflow.Steps {
		// Skip steps without prompt
		if step.Prompt == "" {
			continue
		}
		
		// Extract variables from prompt
		variables, err := renderer.ExtractTemplateVariables(step.Prompt)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to extract variables from prompt '%s': %w", step.Prompt, err))
			continue
		}
		
		// Check if variables are provided
		for _, varName := range variables {
			if step.Variables == nil || step.Variables[varName] == "" {
				// Variable not provided, check if it has a default value
				hasDefault, err := v.checkForDefaultValue(dirPath, step.Prompt, varName)
				if err != nil {
					errors = append(errors, fmt.Errorf("failed to check for default value for variable '%s' in prompt '%s': %w", 
						varName, step.Prompt, err))
					continue
				}
				
				if !hasDefault {
					errors = append(errors, fmt.Errorf("variable '%s' is used in prompt '%s' but not provided in step '%s' and has no default value", 
						varName, step.Prompt, step.ID))
				}
			}
		}
	}
	
	return errors
}

// checkForDefaultValue checks if a variable has a default value in the template.
func (v *WorkflowValidator) checkForDefaultValue(dirPath string, promptPath string, varName string) (bool, error) {
	// Get full path to prompt file
	fullPath := promptPath
	if !filepath.IsAbs(promptPath) {
		fullPath = filepath.Join(dirPath, promptPath)
	}
	
	// Check if prompt file exists
	if !v.fs.Exists(fullPath) {
		return false, fmt.Errorf("prompt file not found: %s", promptPath)
	}
	
	// Read the prompt file
	promptData, err := v.fs.ReadFile(fullPath)
	if err != nil {
		return false, fmt.Errorf("failed to read prompt file: %w", err)
	}
	
	// Look for default value pattern: {{.varName | default "..."}}
	promptText := string(promptData)
	defaultPattern := fmt.Sprintf("{{.%s\\s*\\|\\s*default", varName)
	
	// Very simple check - a more robust implementation would use proper template parsing
	return strings.Contains(promptText, defaultPattern), nil
} 
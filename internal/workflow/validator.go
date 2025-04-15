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
	// TODO: Implement YAML parsing
	// This is a stub for now - actual implementation will be provided later
	// Will use yaml.Unmarshal from the gopkg.in/yaml.v3 package
	return externalWorkflow, fmt.Errorf("parseWorkflowYAML not implemented yet")
} 
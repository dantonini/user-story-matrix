// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"strings"
)

// InterpolationError represents an error during prompt interpolation
// It provides detailed information about malformed and missing variables
type InterpolationError struct {
	Message       string   // Error message
	MalformedVars []string // Variables with syntax issues
	MissingVars   []string // Variables that weren't available for interpolation
}

// Error implements the error interface for InterpolationError
// It formats the error message to include details about malformed and missing variables
func (e *InterpolationError) Error() string {
	if len(e.MalformedVars) > 0 && len(e.MissingVars) > 0 {
		return fmt.Sprintf("%s: malformed variables [%s], missing variables [%s]",
			e.Message, strings.Join(e.MalformedVars, ", "), strings.Join(e.MissingVars, ", "))
	} else if len(e.MalformedVars) > 0 {
		return fmt.Sprintf("%s: malformed variables [%s]", e.Message, strings.Join(e.MalformedVars, ", "))
	} else if len(e.MissingVars) > 0 {
		return fmt.Sprintf("%s: missing variables [%s]", e.Message, strings.Join(e.MissingVars, ", "))
	}
	return e.Message
}

// NewInterpolationError creates a new InterpolationError with the given details
func NewInterpolationError(message string, malformedVars []string, missingVars []string) *InterpolationError {
	return &InterpolationError{
		Message:       message,
		MalformedVars: malformedVars,
		MissingVars:   missingVars,
	}
}

// ValidatePrompt checks if a prompt has valid Go template syntax and returns any errors
func ValidatePrompt(prompt string) error {
	if prompt == "" {
		return nil // Empty prompts are valid
	}

	// Validate Go template syntax
	err := ValidateTemplate(prompt)
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}

	return nil
}

// generateStepPrompt generates a prompt for a workflow step using the Go template system
func generateStepPrompt(step WorkflowStep, changeRequestPath string) string {
	if step.Prompt == "" {
		// Generate a default prompt based on the step description
		return generateDefaultPrompt(step)
	}

	// Create variables map for the template system using predefined runtime variables
	variables := BuildPredefinedRuntimeVariables(changeRequestPath, step)
	
	// Add custom variables from step definition
	if step.Variables != nil {
		for k, v := range step.Variables {
			variables[k] = v
		}
	}

	// Use the Go template system
	stringVars := make(map[string]string)
	for k, v := range variables {
		if str, ok := v.(string); ok {
			stringVars[k] = str
		} else {
			stringVars[k] = fmt.Sprintf("%v", v)
		}
	}

	result, err := ApplyTemplateVariables(step.Prompt, stringVars)
	if err != nil {
		// If template processing fails, return the original prompt
		// This provides graceful degradation
		return step.Prompt
	}

	return result
}

// generateDefaultPrompt creates a default prompt based on the step description
func generateDefaultPrompt(step WorkflowStep) string {
	return "Please execute the following step in the workflow: " + step.Description
}

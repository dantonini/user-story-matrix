// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"regexp"
	"strings"
)

// PromptVariables contains variables that can be interpolated into a prompt
type PromptVariables struct {
	ChangeRequestFilePath string
	// Additional variables can be added here in the future
}

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

// InterpolatePrompt replaces variables in the format ${variable_name} with their values
// This is a simple implementation that only handles the ChangeRequestFilePath variable
// For more complex interpolation with error handling, use InterpolatePromptWithError
func InterpolatePrompt(prompt string, variables PromptVariables) string {
	result := prompt

	// Replace ${change_request_file_path} with the actual path
	result = strings.ReplaceAll(result, "${change_request_file_path}", variables.ChangeRequestFilePath)

	return result
}

// InterpolatePromptWithError replaces variables and returns an error if any problems are encountered
// It identifies both missing variables (not available in the variables struct) and
// malformed variables (syntax issues like spaces in variable names or unclosed braces)
func InterpolatePromptWithError(prompt string, variables PromptVariables) (string, error) {
	result, missingVars, malformedVars := interpolateWithDetails(prompt, variables)
	
	if len(missingVars) > 0 || len(malformedVars) > 0 {
		return result, NewInterpolationError(
			"prompt interpolation encountered issues",
			malformedVars,
			missingVars,
		)
	}
	
	return result, nil
}

// interpolateWithDetails performs variable interpolation and returns details about issues
// It is the core function used by the other interpolation functions
// Returns:
// - The interpolated string with available variables replaced
// - A list of variables that weren't available for interpolation
// - A list of variables with syntax issues
func interpolateWithDetails(prompt string, variables PromptVariables) (string, []string, []string) {
	result := prompt
	missingVars := []string{}
	malformedVars := []string{}
	
	// Regular expression to find all variables in format ${variable_name}
	// This regex matches valid variable names consisting of letters, numbers, underscores, and hyphens
	reValid := regexp.MustCompile(`\${([a-zA-Z0-9_-]+)}`)
	
	// This regex captures malformed variables like ${var with spaces} or ${missing-closing-brace
	reMalformed := regexp.MustCompile(`\${([^}]*[\s]+[^}]*)}|\${([^}]*)$`)
	
	// First, find malformed variables to avoid treating them as valid ones
	malformedMatches := reMalformed.FindAllStringSubmatch(prompt, -1)
	for _, match := range malformedMatches {
		if len(match) > 1 {
			malformedVar := strings.TrimSpace(match[1])
			if malformedVar != "" {
				malformedVars = append(malformedVars, malformedVar)
			} else if len(match) > 2 && match[2] != "" {
				malformedVars = append(malformedVars, match[2])
			}
		}
	}
	
	// Next, find and replace valid variables
	validMatches := reValid.FindAllStringSubmatch(prompt, -1)
	for _, match := range validMatches {
		if len(match) > 1 {
			varName := match[1]
			if varName == "change_request_file_path" && variables.ChangeRequestFilePath != "" {
				// Replace ${change_request_file_path} with the actual path
				result = strings.ReplaceAll(result, "${"+varName+"}", variables.ChangeRequestFilePath)
			} else {
				// Mark as missing
				missingVars = append(missingVars, varName)
			}
		}
	}
	
	return result, missingVars, malformedVars
}

// InterpolatePromptWithMissingVars replaces variables and returns a list of missing variables
func InterpolatePromptWithMissingVars(prompt string, variables PromptVariables) (string, []string) {
	result, missingVars, _ := interpolateWithDetails(prompt, variables)
	return result, missingVars
}

// interpolatePromptWithMap replaces variables using a map of variable names to values
func interpolatePromptWithMap(prompt string, variables map[string]string) string {
	result := prompt

	// Regular expression to find all variables in format ${variable_name}
	re := regexp.MustCompile(`\${([^}]+)}`)
	matches := re.FindAllStringSubmatch(prompt, -1)

	// Replace each variable with its value if available
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if value, exists := variables[varName]; exists {
				result = strings.ReplaceAll(result, "${"+varName+"}", value)
			}
		}
	}

	return result
}

// ValidatePrompt checks if a prompt has valid variable syntax and returns any errors
func ValidatePrompt(prompt string) error {
	_, _, malformedVars := interpolateWithDetails(prompt, PromptVariables{})
	
	if len(malformedVars) > 0 {
		return NewInterpolationError(
			"prompt contains malformed variables",
			malformedVars,
			nil,
		)
	}
	
	return nil
}

// generateStepPrompt generates a prompt for a workflow step
func generateStepPrompt(step WorkflowStep, changeRequestPath string) string {
	if step.Prompt == "" {
		// Generate a default prompt based on the step description
		return generateDefaultPrompt(step)
	}

	vars := PromptVariables{
		ChangeRequestFilePath: changeRequestPath,
	}
	return InterpolatePrompt(step.Prompt, vars)
}

// generateDefaultPrompt creates a default prompt based on the step description
func generateDefaultPrompt(step WorkflowStep) string {
	return "Please execute the following step in the workflow: " + step.Description
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"regexp"
	"strings"
)

// PromptVariables contains variables that can be interpolated into a prompt
type PromptVariables struct {
	ChangeRequestFilePath string
	// Additional variables can be added here in the future
}

// InterpolatePrompt replaces variables in the format ${variable_name} with their values
func InterpolatePrompt(prompt string, variables PromptVariables) string {
	result := prompt

	// Replace ${change_request_file_path} with the actual path
	result = strings.ReplaceAll(result, "${change_request_file_path}", variables.ChangeRequestFilePath)

	return result
}

// InterpolatePromptWithMissingVars replaces variables and returns a list of missing variables
func InterpolatePromptWithMissingVars(prompt string, variables PromptVariables) (string, []string) {
	result := prompt
	missingVars := []string{}

	// Regular expression to find all variables in format ${variable_name}
	re := regexp.MustCompile(`\${([^}]+)}`)
	matches := re.FindAllStringSubmatch(prompt, -1)

	// Extract variable names
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if varName == "change_request_file_path" && variables.ChangeRequestFilePath != "" {
				// Replace ${change_request_file_path} with the actual path
				result = strings.ReplaceAll(result, "${change_request_file_path}", variables.ChangeRequestFilePath)
			} else {
				// Mark as missing
				missingVars = append(missingVars, varName)
			}
		}
	}

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
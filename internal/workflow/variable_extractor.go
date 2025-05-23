// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"path/filepath"
	"strings"
	"text/template"
)

// PredefinedRuntimeVariables are variables that are automatically available during workflow execution
// and should not be explicitly declared in workflow.yaml files.
var PredefinedRuntimeVariables = map[string]bool{
	"ChangeRequestFilePath": true,
	"ChangeRequestBasename": true,
	"BlueprintBasename":     true,
	"ChangeRequestDirname":  true,
	"StepID":                true,
	"StepName":              true,
	"ChangeRequestFullpath": true,
	"Basename":              true, // Deprecated but still supported
}

// IsPredefinedRuntimeVariable checks if a variable is a predefined runtime variable
// that is automatically available during workflow execution.
func IsPredefinedRuntimeVariable(varName string) bool {
	return PredefinedRuntimeVariables[varName]
}

// BuildPredefinedRuntimeVariables creates a map of predefined runtime variables
// based on the execution context. This is the single source of truth for 
// predefined variable values.
//
// Parameters:
//   - changeRequestPath: Path to the change request file
//   - step: The workflow step being executed
//
// Returns:
//   - A map of predefined variable names to their runtime values
func BuildPredefinedRuntimeVariables(changeRequestPath string, step WorkflowStep) map[string]interface{} {
	// Extract path components for variables
	dir := filepath.Dir(changeRequestPath)
	base := filepath.Base(changeRequestPath)
	base = strings.TrimSuffix(base, ".blueprint.md")
	fullpath := filepath.Join(dir, base)

	// Extract step name from ID
	stepName := step.ID
	if parts := strings.SplitN(step.ID, "-", 2); len(parts) > 1 {
		stepName = parts[1]
	}

	// Create the predefined variables map
	return map[string]interface{}{
		"ChangeRequestFilePath": changeRequestPath,
		"ChangeRequestBasename": base,
		"BlueprintBasename":     base,
		"ChangeRequestDirname":  dir,
		"StepID":                step.ID,
		"StepName":              stepName,
		"ChangeRequestFullpath": fullpath,
		"Basename":              base, // Deprecated but maintained for compatibility
	}
}

// ExtractTemplateVariablesFromString extracts template variables from a string content.
// This is useful for extracting variables from builtin workflow prompts before writing
// them to external files.
//
// Parameters:
//   - content: The template content as a string
//
// Returns:
//   - A slice of variable names, or an error if extraction failed
func ExtractTemplateVariablesFromString(content string) ([]string, error) {
	if content == "" {
		return []string{}, nil
	}
	
	// Create function map with common functions used in templates
	funcMap := template.FuncMap{
		"default": func(defaultVal interface{}, val interface{}) interface{} {
			if val == nil {
				return defaultVal
			}
			if s, ok := val.(string); ok && s == "" {
				return defaultVal
			}
			return val
		},
		"join":  strings.Join,
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"trim":  strings.TrimSpace,
	}
	
	// Parse the template to get the AST
	tmpl, err := template.New("temp").Funcs(funcMap).Parse(content)
	if err != nil {
		return nil, err
	}
	
	// Extract variables from the template AST
	variableMap := make(map[string]bool) // Use map to deduplicate
	for _, node := range tmpl.Tree.Root.Nodes {
		extractVariablesFromNode(node, variableMap)
	}
	
	// Convert map to slice
	variables := make([]string, 0, len(variableMap))
	for varName := range variableMap {
		variables = append(variables, varName)
	}
	
	return variables, nil
}

// ExtractVariablesFromBuiltinStep extracts variables from a builtin workflow step
// and returns a Variables map with appropriate default values for variables that
// need to be explicitly defined (excluding predefined runtime variables).
//
// Parameters:
//   - step: The WorkflowStep from a builtin workflow
//
// Returns:
//   - A map of variable names to default values suitable for external workflow format,
//     excluding predefined runtime variables that are automatically available
func ExtractVariablesFromBuiltinStep(step WorkflowStep) (map[string]string, error) {
	variables, err := ExtractTemplateVariablesFromString(step.Prompt)
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]string)
	
	// Add variables with appropriate default values, but skip predefined runtime variables
	for _, varName := range variables {
		// Skip predefined runtime variables - these are automatically available
		if IsPredefinedRuntimeVariable(varName) {
			continue
		}
		
		// For unknown variables that aren't predefined, mark them as needing configuration
		result[varName] = "CONFIGURE_ME"
	}
	
	return result, nil
} 
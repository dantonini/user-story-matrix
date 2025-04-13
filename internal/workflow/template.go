// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// TemplateContext holds the template execution context including variables and functions
type TemplateContext struct {
	Variables map[string]interface{} // Variables available to templates
	Functions template.FuncMap       // Custom template functions
}

// TemplateProcessor provides template processing capabilities for workflow steps
type TemplateProcessor struct {
	// Cached templates to improve performance for repeated executions
	templateCache map[string]*template.Template
}

// NewTemplateProcessor creates a new template processor with default functions
func NewTemplateProcessor() *TemplateProcessor {
	return &TemplateProcessor{
		templateCache: make(map[string]*template.Template),
	}
}

// Regular expression for finding default function patterns like {{.varname | default "value"}}
var defaultFuncRegex = regexp.MustCompile(`{{\.([a-zA-Z0-9_]+) \| default "([^"]*)"}}`)

// ApplyTemplateVariables processes a template string with the provided variables
// It replaces variables according to Go's text/template syntax
//
// Parameters:
//   - promptContent: The template content with variable placeholders
//   - variables: Map of variable names to values
//
// Returns:
//   - The processed template with variables replaced
//   - Error if template processing fails
func ApplyTemplateVariables(promptContent string, variables map[string]string) (string, error) {
	// First pre-process the template to handle default values
	processedTemplate := promptContent
	
	// Find all default function patterns
	matches := defaultFuncRegex.FindAllStringSubmatch(promptContent, -1)
	for _, match := range matches {
		if len(match) == 3 {
			fullMatch := match[0]         // {{.varname | default "value"}}
			varName := match[1]           // varname
			defaultValue := match[2]      // value
			
			// Check if variable exists and has a non-empty value
			if value, exists := variables[varName]; exists && value != "" {
				// Replace with just the variable reference
				replacement := fmt.Sprintf("{{.%s}}", varName)
				processedTemplate = strings.Replace(processedTemplate, fullMatch, replacement, -1)
			} else {
				// Replace with the default value directly
				processedTemplate = strings.Replace(processedTemplate, fullMatch, defaultValue, -1)
			}
		}
	}
	
	// Now process the template with the standard template engine
	// Convert string map to interface map
	varMap := make(map[string]interface{})
	for k, v := range variables {
		varMap[k] = v
	}
	
	// Add custom functions
	funcMap := template.FuncMap{
		"default": defaultFunction,
		// Additional functions can be added here as needed
	}
	
	// Parse and execute the template
	tmpl, err := template.New("prompt").Funcs(funcMap).Parse(processedTemplate)
	if err != nil {
		return "", fmt.Errorf("template parsing error: %w", err)
	}
	
	var output bytes.Buffer
	err = tmpl.Execute(&output, varMap)
	if err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}
	
	// Replace "<no value>" with empty string for missing variables
	result := strings.ReplaceAll(output.String(), "<no value>", "")
	
	return result, nil
}

// defaultFunction provides a default value for a variable if it doesn't exist or is empty
// Usage in templates: {{.variable_name | default "default value"}}
func defaultFunction(arg, defaultVal interface{}) interface{} {
	// If arg is nil, return the default value
	if arg == nil {
		return defaultVal
	}
	
	// If arg is a string and empty, return the default value
	if s, ok := arg.(string); ok && s == "" {
		return defaultVal
	}
	
	// Otherwise, return the original arg
	return arg
}

// processTemplate handles the actual template processing with error checking and context preparation
// This is an internal function used by ApplyTemplateVariables
func processTemplate(name, content string, variables map[string]interface{}, funcs template.FuncMap) (string, error) {
	// Create a new template with functions
	tmpl, err := template.New(name).Funcs(funcs).Parse(content)
	if err != nil {
		return "", fmt.Errorf("template parsing error: %w", err)
	}
	
	// Execute template with variables
	var output bytes.Buffer
	err = tmpl.Execute(&output, variables)
	if err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}
	
	return output.String(), nil
}

// prepareContext converts string map to interface map and adds custom functions
func prepareContext(variables map[string]string) TemplateContext {
	// Convert string map to interface map
	varMap := make(map[string]interface{})
	for k, v := range variables {
		varMap[k] = v
	}
	
	// Create function map with custom functions
	funcMap := template.FuncMap{
		"default": defaultFunction,
		// Additional functions can be added here as needed
	}
	
	return TemplateContext{
		Variables: varMap,
		Functions: funcMap,
	}
}

// validateTemplate checks a template for potential issues before execution
func validateTemplate(templateContent string) error {
	// Basic validation - check if template can be parsed
	_, err := template.New("validation").Parse(templateContent)
	if err != nil {
		return fmt.Errorf("template validation error: %w", err)
	}
	
	// Check for unclosed tags (basic check)
	openCount := strings.Count(templateContent, "{{")
	closeCount := strings.Count(templateContent, "}}")
	
	if openCount != closeCount {
		return fmt.Errorf("template contains unclosed tags: %d open tags, %d close tags", openCount, closeCount)
	}
	
	return nil
} 
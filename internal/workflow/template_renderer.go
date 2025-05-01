// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/user-story-matrix/usm/internal/io"
)

// TemplateRenderer renders prompt templates with variable substitution.
// It supports Go template syntax, default values, and structured data.
type TemplateRenderer struct {
	fs          io.FileSystem
	workflowDir string
	cache       map[string]*template.Template
}

// NewTemplateRenderer creates a new template renderer for the given workflow directory.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - workflowDir: Path to the workflow directory
//
// Returns:
//   - A new TemplateRenderer instance
func NewTemplateRenderer(fs io.FileSystem, workflowDir string) *TemplateRenderer {
	return &TemplateRenderer{
		fs:          fs,
		workflowDir: workflowDir,
		cache:       make(map[string]*template.Template),
	}
}

// RenderPrompt renders a prompt template with the given variables.
//
// Parameters:
//   - promptPath: Path to the prompt template file, relative to the workflow directory
//   - variables: Variables to substitute in the template
//
// Returns:
//   - The rendered prompt text, or an error if rendering failed
func (r *TemplateRenderer) RenderPrompt(promptPath string, variables map[string]interface{}) (string, error) {
	// Get full path to prompt file
	fullPath := promptPath
	if !filepath.IsAbs(promptPath) {
		fullPath = filepath.Join(r.workflowDir, promptPath)
	}
	
	// Check if prompt file exists
	if !r.fs.Exists(fullPath) {
		return "", fmt.Errorf("prompt file not found: %s", promptPath)
	}
	
	// Read the prompt file
	promptData, err := r.fs.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %w", err)
	}
	
	// Check if template is cached
	tmpl, ok := r.cache[promptPath]
	if !ok {
		// Create new template
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
			"join": strings.Join,
			"lower": strings.ToLower,
			"upper": strings.ToUpper,
			"title": strings.Title,
			"trim": strings.TrimSpace,
		}
		
		tmpl = template.New(filepath.Base(promptPath)).Funcs(funcMap)
		tmpl, err = tmpl.Parse(string(promptData))
		if err != nil {
			return "", fmt.Errorf("invalid template syntax in %s: %w", promptPath, err)
		}
		
		// Cache the template
		r.cache[promptPath] = tmpl
	}
	
	// Execute the template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, variables)
	if err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}
	
	return buf.String(), nil
}

// ValidateTemplate validates a prompt template.
//
// Parameters:
//   - promptPath: Path to the prompt template file, relative to the workflow directory
//
// Returns:
//   - An error if validation failed, nil if the template is valid
func (r *TemplateRenderer) ValidateTemplate(promptPath string) error {
	// Get full path to prompt file
	fullPath := promptPath
	if !filepath.IsAbs(promptPath) {
		fullPath = filepath.Join(r.workflowDir, promptPath)
	}
	
	// Check if prompt file exists
	if !r.fs.Exists(fullPath) {
		return fmt.Errorf("prompt file not found: %s", promptPath)
	}
	
	// Read the prompt file
	promptData, err := r.fs.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read prompt file: %w", err)
	}
	
	// Create new template
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
	}
	
	tmpl := template.New(filepath.Base(promptPath)).Funcs(funcMap)
	_, err = tmpl.Parse(string(promptData))
	if err != nil {
		return fmt.Errorf("invalid template syntax in %s: %w", promptPath, err)
	}
	
	return nil
}

// ExtractTemplateVariables extracts the variables used in a prompt template.
// This implementation uses basic pattern matching to find Go template variables.
//
// Parameters:
//   - promptPath: Path to the prompt template file, relative to the workflow directory
//
// Returns:
//   - A slice of variable names, or an error if extraction failed
func (r *TemplateRenderer) ExtractTemplateVariables(promptPath string) ([]string, error) {
	// Get full path to prompt file
	fullPath := promptPath
	if !filepath.IsAbs(promptPath) {
		fullPath = filepath.Join(r.workflowDir, promptPath)
	}
	
	// Check if prompt file exists
	if !r.fs.Exists(fullPath) {
		return nil, fmt.Errorf("prompt file not found: %s", promptPath)
	}
	
	// Read the prompt file
	promptData, err := r.fs.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file: %w", err)
	}
	
	// Extract variables using regular expressions
	promptText := string(promptData)
	variableMap := make(map[string]bool) // Use map to deduplicate
	
	// Process the template to find all variable references
	// This is a simplified approach that may not catch all variable usages in complex templates
	
	// Pattern 1: {{.variable}} and {{.variable | function}}
	start := 0
	for {
		// Find the next opening brace
		openBrace := strings.Index(promptText[start:], "{{")
		if openBrace == -1 {
			break
		}
		openBrace += start
		
		// Find the corresponding closing brace
		closeBrace := strings.Index(promptText[openBrace:], "}}")
		if closeBrace == -1 {
			break
		}
		closeBrace += openBrace
		
		// Extract the template expression
		expression := strings.TrimSpace(promptText[openBrace+2:closeBrace])
		
		// Process expression
		if strings.HasPrefix(expression, ".") {
			// Simple variable reference like {{.varName}}
			varName := extractVariableName(expression)
			if varName != "" {
				variableMap[varName] = true
			}
		} else if strings.HasPrefix(expression, "if .") || 
		        strings.HasPrefix(expression, "with .") || 
		        strings.HasPrefix(expression, "range .") {
			// Control structure like {{if .varName}} or {{range .items}}
			parts := strings.Fields(expression)
			if len(parts) >= 2 && strings.HasPrefix(parts[1], ".") {
				varName := parts[1][1:] // Remove the dot
				// Remove any trailing characters like parentheses or pipes
				varName = trimNonAlphanumeric(varName)
				if varName != "" {
					variableMap[varName] = true
				}
			}
		} else if strings.Contains(expression, ".") {
			// Function call with variable like {{join .items ", "}}
			// This is a simplistic approach and won't work for all cases
			parts := strings.Fields(expression)
			for _, part := range parts {
				if strings.HasPrefix(part, ".") {
					varName := trimNonAlphanumeric(part[1:])
					if varName != "" {
						variableMap[varName] = true
					}
				}
			}
		}
		
		// Move past this expression
		start = closeBrace + 2
	}
	
	// Convert map to slice
	variables := make([]string, 0, len(variableMap))
	for varName := range variableMap {
		variables = append(variables, varName)
	}
	
	return variables, nil
}

// extractVariableName extracts a variable name from a template expression
func extractVariableName(expr string) string {
	// Remove leading dot
	if !strings.HasPrefix(expr, ".") {
		return ""
	}
	expr = expr[1:]
	
	// Find the end of the variable name (space, pipe, etc.)
	end := 0
	for i, c := range expr {
		if !isValidVariableChar(c) {
			end = i
			break
		}
		end = i + 1
	}
	
	if end == 0 {
		return ""
	}
	
	return expr[:end]
}

// isValidVariableChar checks if a character is valid in a variable name
func isValidVariableChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || 
	       (c >= 'A' && c <= 'Z') || 
	       (c >= '0' && c <= '9') || 
	       c == '_'
}

// trimNonAlphanumeric removes non-alphanumeric characters from the end of a string
func trimNonAlphanumeric(s string) string {
	end := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		if isValidVariableChar(rune(s[i])) {
			end = i + 1
			break
		}
	}
	return s[:end]
}

// sliceContains checks if a string slice contains a string
// Local helper function to avoid naming conflicts
func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
} 
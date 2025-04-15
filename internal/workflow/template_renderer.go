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
// NOTE: This is a placeholder implementation that will need to be enhanced
// to properly extract variables from Go templates.
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
	
	// Extract variables (simplified implementation)
	// A more robust implementation would need to parse the template
	// This is a placeholder that looks for {{.variable}} patterns
	promptText := string(promptData)
	variables := make([]string, 0)
	
	// Find all {{.name}} patterns
	start := 0
	for {
		varStart := strings.Index(promptText[start:], "{{.")
		if varStart == -1 {
			break
		}
		varStart += start + 3 // Skip "{{."
		
		varEnd := strings.Index(promptText[varStart:], "}}")
		if varEnd == -1 {
			break
		}
		
		// Extract variable name
		varText := promptText[varStart : varStart+varEnd]
		varName := strings.Split(varText, " ")[0]
		varName = strings.Split(varName, "|")[0]
		varName = strings.TrimSpace(varName)
		
		// Add to results if not already present
		if !sliceContains(variables, varName) {
			variables = append(variables, varName)
		}
		
		// Move past this variable
		start = varStart + varEnd + 2
	}
	
	return variables, nil
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
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"text/template/parse"

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
	
	// First try with the provided path
	if !r.fs.Exists(fullPath) {
		// If file doesn't exist at the direct path, try to find it in standard locations
		pwd, _ := os.Getwd()
		possiblePaths := []string{
			fullPath,
			filepath.Join(pwd, promptPath),                // Try relative to current directory
			filepath.Join(r.workflowDir, "..", promptPath), // Try one level up
			filepath.Join(r.workflowDir, "..", "..", promptPath), // Try two levels up
			filepath.Join(r.workflowDir, "prompts", filepath.Base(promptPath)), // Try in prompts subdirectory
		}
		
		foundPath := ""
		for _, path := range possiblePaths {
			if r.fs.Exists(path) {
				foundPath = path
				break
			}
		}
		
		if foundPath == "" {
			return "", fmt.Errorf("prompt file not found: %s", promptPath)
		}
		
		fullPath = foundPath
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
			// Title is deprecated - leaving as a comment for the extend phase to implement properly
			// "title": strings.Title,
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
	
	// First try with the provided path
	if !r.fs.Exists(fullPath) {
		// If file doesn't exist at the direct path, try to find it in standard locations
		// Read the current directory structure to help debug
		pwd, _ := os.Getwd()
		possiblePaths := []string{
			fullPath,
			filepath.Join(pwd, promptPath),                // Try relative to current directory
			filepath.Join(r.workflowDir, "..", promptPath), // Try one level up
			filepath.Join(r.workflowDir, "..", "..", promptPath), // Try two levels up
			filepath.Join(r.workflowDir, "prompts", filepath.Base(promptPath)), // Try in prompts subdirectory
		}
		
		foundPath := ""
		for _, path := range possiblePaths {
			if r.fs.Exists(path) {
				foundPath = path
				break
			}
		}
		
		if foundPath == "" {
			return fmt.Errorf("prompt file not found: %s", promptPath)
		}
		
		fullPath = foundPath
	}
	
	// Read the prompt file
	promptData, err := r.fs.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read prompt file: %w", err)
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
	
	tmpl := template.New(filepath.Base(promptPath)).Funcs(funcMap)
	_, err = tmpl.Parse(string(promptData))
	if err != nil {
		return fmt.Errorf("invalid template syntax in %s: %w", promptPath, err)
	}
	
	return nil
}

// ExtractTemplateVariables extracts the variables used in a prompt template.
// This implementation uses the Go template parser to accurately extract variables.
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
	
	// First try with the provided path
	if !r.fs.Exists(fullPath) {
		// If file doesn't exist at the direct path, try to find it in standard locations
		// Read the current directory structure to help debug
		pwd, _ := os.Getwd()
		possiblePaths := []string{
			fullPath,
			filepath.Join(pwd, promptPath),                // Try relative to current directory
			filepath.Join(r.workflowDir, "..", promptPath), // Try one level up
			filepath.Join(r.workflowDir, "..", "..", promptPath), // Try two levels up
			filepath.Join(r.workflowDir, "prompts", filepath.Base(promptPath)), // Try in prompts subdirectory
		}
		
		foundPath := ""
		for _, path := range possiblePaths {
			if r.fs.Exists(path) {
				foundPath = path
				break
			}
		}
		
		if foundPath == "" {
			return nil, fmt.Errorf("prompt file not found: %s", promptPath)
		}
		
		fullPath = foundPath
	}
	
	// Read the prompt file
	promptData, err := r.fs.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file: %w", err)
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
	tmpl, err := template.New(filepath.Base(promptPath)).Funcs(funcMap).Parse(string(promptData))
	if err != nil {
		return nil, fmt.Errorf("invalid template syntax in %s: %w", promptPath, err)
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

// extractVariablesFromNode recursively extracts variable names from a template AST node
func extractVariablesFromNode(node parse.Node, variables map[string]bool) {
	switch n := node.(type) {
	case *parse.ActionNode:
		// ActionNode represents a {{...}} action
		extractVariablesFromPipe(n.Pipe, variables)
	case *parse.IfNode:
		// IfNode represents an {{if ...}} action
		extractVariablesFromPipe(n.Pipe, variables)
		for _, ifNode := range n.List.Nodes {
			extractVariablesFromNode(ifNode, variables)
		}
		if n.ElseList != nil {
			for _, elseNode := range n.ElseList.Nodes {
				extractVariablesFromNode(elseNode, variables)
			}
		}
	case *parse.RangeNode:
		// RangeNode represents a {{range ...}} action
		extractVariablesFromPipe(n.Pipe, variables)
		for _, rangeNode := range n.List.Nodes {
			extractVariablesFromNode(rangeNode, variables)
		}
		if n.ElseList != nil {
			for _, elseNode := range n.ElseList.Nodes {
				extractVariablesFromNode(elseNode, variables)
			}
		}
	case *parse.WithNode:
		// WithNode represents a {{with ...}} action
		extractVariablesFromPipe(n.Pipe, variables)
		for _, withNode := range n.List.Nodes {
			extractVariablesFromNode(withNode, variables)
		}
		if n.ElseList != nil {
			for _, elseNode := range n.ElseList.Nodes {
				extractVariablesFromNode(elseNode, variables)
			}
		}
	case *parse.ListNode:
		// ListNode represents a list of nodes
		if n != nil {
			for _, listNode := range n.Nodes {
				extractVariablesFromNode(listNode, variables)
			}
		}
	case *parse.TemplateNode:
		// TemplateNode represents a {{template ...}} action
		if n.Pipe != nil {
			extractVariablesFromPipe(n.Pipe, variables)
		}
	}
}

// extractVariablesFromPipe extracts variable names from a template pipe node
func extractVariablesFromPipe(pipe *parse.PipeNode, variables map[string]bool) {
	if pipe == nil {
		return
	}
	
	for _, cmd := range pipe.Cmds {
		for _, arg := range cmd.Args {
			extractVariablesFromArg(arg, variables)
		}
	}
}

// extractVariablesFromArg extracts variable names from a template argument node
func extractVariablesFromArg(arg parse.Node, variables map[string]bool) {
	switch n := arg.(type) {
	case *parse.FieldNode:
		// FieldNode represents a .Field or .Field.Field etc.
		if len(n.Ident) > 0 {
			variables[n.Ident[0]] = true
		}
	case *parse.VariableNode:
		// VariableNode represents a $variable
		if len(n.Ident) > 0 {
			variables[n.Ident[0]] = true
		}
	case *parse.PipeNode:
		// Nested pipe
		extractVariablesFromPipe(n, variables)
	case *parse.ChainNode:
		// ChainNode represents a function call with arguments
		extractVariablesFromArg(n.Node, variables)
	}
} 
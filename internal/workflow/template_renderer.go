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
	"regexp"
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

// resolvePromptPath attempts to find a prompt file by checking multiple possible locations.
// It returns the full path to the file if found, or an empty string if not found.
func (r *TemplateRenderer) resolvePromptPath(promptPath string) string {
	// Check if the path is already absolute
	if filepath.IsAbs(promptPath) {
		if r.fs.Exists(promptPath) {
			return promptPath
		}
	}
	
	// Build a list of possible paths to check
	possiblePaths := []string{}
	
	// First check if the path is relative to the workflow directory
	fullPath := filepath.Join(r.workflowDir, promptPath)
	possiblePaths = append(possiblePaths, fullPath)
	
	// If the path doesn't already include "prompts/" directory, try adding it
	if !strings.HasPrefix(promptPath, "prompts/") && !strings.HasPrefix(promptPath, "prompts\\") {
		possiblePaths = append(possiblePaths, filepath.Join(r.workflowDir, "prompts", promptPath))
	}
	
	// Check if filename contains path separators
	if filepath.Base(promptPath) == promptPath {
		// If it's just a filename without directories, look in the prompts directory
		possiblePaths = append(possiblePaths, filepath.Join(r.workflowDir, "prompts", promptPath))
	}
	
	// Add some more fallback paths
	pwd, _ := os.Getwd()
	possiblePaths = append(possiblePaths, 
		filepath.Join(pwd, promptPath), // Relative to current directory
		promptPath, // As is (might be in current directory)
		filepath.Join(r.workflowDir, "..", promptPath), // One level up
		filepath.Join(r.workflowDir, "..", "..", promptPath), // Two levels up
	)
	
	// Check each possible path
	for _, path := range possiblePaths {
		if r.fs.Exists(path) {
			return path
		}
	}
	
	// No path found
	return ""
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
	// Find the prompt file path
	resolvedPath := r.resolvePromptPath(promptPath)
	if resolvedPath == "" {
		return "", fmt.Errorf("prompt file not found: %s (tried paths: %s)", promptPath, "")
	}
	
	// Read the prompt file
	promptData, err := r.fs.ReadFile(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %w", err)
	}
	
	// Create a template name that's based on the path but cleaned up for use as template name
	templateName := filepath.Base(promptPath)
	
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
		
		// Create a root template with our functions
		tmpl = template.New(templateName).Funcs(funcMap)
		
		// Parse the main template
		tmpl, err = tmpl.Parse(string(promptData))
		if err != nil {
			return "", fmt.Errorf("invalid template syntax in %s: %w", resolvedPath, err)
		}
		
		// Find and parse any referenced templates (e.g., "{{ template "shared/footer.md" . }}")
		for _, t := range findTemplateReferences(string(promptData)) {
			// Skip if this template is already defined
			if tmpl.Lookup(t) != nil {
				continue
			}
			
			// Try to find the referenced template file
			// For "shared/footer.md", we would look in prompts/shared/footer.md
			refPath := ""
			if strings.HasPrefix(t, "shared/") {
				// For shared references, look in the shared directory
				refPath = filepath.Join(filepath.Dir(resolvedPath), t)
			} else {
				// For other references, try to resolve directly
				refPath = r.resolvePromptPath(t)
			}
			
			if refPath == "" {
				return "", fmt.Errorf("referenced template not found: %s", t)
			}
			
			// Read the template file
			refData, err := r.fs.ReadFile(refPath)
			if err != nil {
				return "", fmt.Errorf("failed to read template %s: %w", t, err)
			}
			
			// Parse the referenced template with the same name that's used in the include
			_, err = tmpl.New(t).Parse(string(refData))
			if err != nil {
				return "", fmt.Errorf("invalid template syntax in %s: %w", refPath, err)
			}
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

// findTemplateReferences finds all template references in the given template content.
// It looks for patterns like {{ template "name" . }}
func findTemplateReferences(content string) []string {
	// Use a simple regex to find template inclusions
	templateRefRegex := `{{[\s]*template[\s]+"([^"]+)"[^}]*}}`
	re := regexp.MustCompile(templateRefRegex)
	matches := re.FindAllStringSubmatch(content, -1)
	
	// Extract the template names
	templates := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 2 {
			templates = append(templates, match[1])
		}
	}
	
	return templates
}

// ValidateTemplate validates a prompt template.
//
// Parameters:
//   - promptPath: Path to the prompt template file, relative to the workflow directory
//
// Returns:
//   - An error if validation failed, nil if the template is valid
func (r *TemplateRenderer) ValidateTemplate(promptPath string) error {
	// Find the prompt file path
	resolvedPath := r.resolvePromptPath(promptPath)
	if resolvedPath == "" {
		return fmt.Errorf("prompt file not found: %s", promptPath)
	}
	
	// Read the prompt file
	promptData, err := r.fs.ReadFile(resolvedPath)
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
	// Find the prompt file path
	resolvedPath := r.resolvePromptPath(promptPath)
	if resolvedPath == "" {
		return nil, fmt.Errorf("prompt file not found: %s", promptPath)
	}
	
	// Read the prompt file
	promptData, err := r.fs.ReadFile(resolvedPath)
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
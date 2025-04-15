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

// ApplyTemplateVariables processes a template string with the provided variables
// It replaces variables according to Go's text/template syntax
//
// Parameters:
//   - templateContent: The template content with variable placeholders
//   - variables: Map of variable names to values
//
// Returns:
//   - The processed template with variables replaced
//   - Error if template processing fails
//   - Slice of warnings for undefined variables (returned as part of the error if any)
func ApplyTemplateVariables(templateContent string, variables map[string]string) (string, error) {
	// Validate the template before processing
	if err := ValidateTemplate(templateContent); err != nil {
		return "", err
	}

	// Process variables to handle nested structures and arrays
	processedVars := make(map[string]interface{})
	
	// Organize variables by their path prefix
	variablesByPrefix := make(map[string]map[string]string)
	for key, value := range variables {
		if strings.Contains(key, ".") {
			prefix := strings.Split(key, ".")[0]
			if _, exists := variablesByPrefix[prefix]; !exists {
				variablesByPrefix[prefix] = make(map[string]string)
			}
			variablesByPrefix[prefix][key] = value
		} else {
			// Top-level variables go in their own map
			if _, exists := variablesByPrefix[""]; !exists {
				variablesByPrefix[""] = make(map[string]string)
			}
			variablesByPrefix[""][key] = value
		}
	}
	
	// Process top-level variables first
	if topLevelVars, exists := variablesByPrefix[""]; exists {
		for key, value := range topLevelVars {
			if IsArrayContext(key) {
				// Handle top-level arrays
				baseKey := strings.TrimSuffix(key, "[]")
				if value != "" {
					items := strings.Split(value, ",")
					for i := range items {
						items[i] = strings.TrimSpace(items[i])
					}
					processedVars[baseKey] = items
				} else {
					processedVars[baseKey] = []string{}
				}
			} else {
				// Regular variable
				processedVars[key] = value
			}
		}
	}
	
	// Then process nested variables, grouped by their top-level prefix
	for prefix, prefixVars := range variablesByPrefix {
		if prefix == "" {
			continue // Already processed
		}
		
		// Create the top-level object if it doesn't exist
		if _, exists := processedVars[prefix]; !exists {
			processedVars[prefix] = make(map[string]interface{})
		} else if _, isMap := processedVars[prefix].(map[string]interface{}); !isMap {
			// Convert non-map value to map while preserving original value
			originalValue := processedVars[prefix]
			newMap := make(map[string]interface{})
			if originalValue != nil {
				newMap[""] = originalValue
			}
			processedVars[prefix] = newMap
		}
		
		// Process each nested variable in this prefix group
		for key, value := range prefixVars {
			if strings.HasPrefix(key, prefix+".") {
				// Remove the prefix from the key to get the path within the object
				relativePath := key[len(prefix)+1:]
				parts := strings.Split(relativePath, ".")
				
				// Start at the top level object for this prefix
				currentValue, ok := processedVars[prefix].(map[string]interface{})
				if !ok {
					// If this fails, it's a serious type mismatch error, log and skip
					continue
				}
				
				// Build the nested structure for all parts except the last one
				for i := 0; i < len(parts)-1; i++ {
					part := parts[i]
					
					// If this level doesn't exist yet, create it
					if _, exists := currentValue[part]; !exists {
						currentValue[part] = make(map[string]interface{})
					} else if _, isMap := currentValue[part].(map[string]interface{}); !isMap {
						// Convert non-map value to map while preserving original value
						originalValue := currentValue[part]
						newMap := make(map[string]interface{})
						if originalValue != nil {
							newMap[""] = originalValue
						}
						currentValue[part] = newMap
					}
					
					// Move to the next level
					nextValue, ok := currentValue[part].(map[string]interface{})
					if !ok {
						// If this fails, it's a type mismatch, skip this entry
						break
					}
					currentValue = nextValue
				}
				
				// Set the final value at the leaf
				lastPart := parts[len(parts)-1]
				
				// Handle array context for the leaf node
				if IsArrayContext(lastPart) {
					baseKey := strings.TrimSuffix(lastPart, "[]")
					if value != "" {
						items := strings.Split(value, ",")
						for i := range items {
							items[i] = strings.TrimSpace(items[i])
						}
						currentValue[baseKey] = items
					} else {
						currentValue[baseKey] = []string{}
					}
				} else {
					currentValue[lastPart] = value
				}
			}
		}
	}

	// Create a copy of processedVars for dot context to make variables accessible in all scopes
	dotContext := make(map[string]interface{})
	for k, v := range processedVars {
		dotContext[k] = v
	}

	// Pre-process the template to handle default values
	// This regex matches patterns like {{.varname | default "defaultValue"}}
	defaultPattern := regexp.MustCompile(`{{\.([a-zA-Z0-9_]+)\s*\|\s*default\s*"([^"]*)"}}`)
	processedTemplate := defaultPattern.ReplaceAllStringFunc(templateContent, func(match string) string {
		// Extract variable name and default value
		submatches := defaultPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match // Should not happen if regex works correctly
		}
		
		varName := submatches[1]
		defaultValue := submatches[2]
		
		// Check if the variable has a value
		if value, exists := processedVars[varName]; exists && value != "" {
			// Use the actual value
			return fmt.Sprintf("{{.%s}}", varName)
		}
		
		// Add the default value to the variables
		processedVars[varName] = defaultValue
		
		// Use a simple variable reference
		return fmt.Sprintf("{{.%s}}", varName)
	})

	// Pre-process the template to handle variable references inside range loops
	// This regex matches range blocks with potential variable references
	rangeRegex := regexp.MustCompile(`{{range\s+\.([a-zA-Z0-9_]+)}}(.*?){{end}}`)
	processedTemplate = rangeRegex.ReplaceAllStringFunc(processedTemplate, func(match string) string {
		submatches := rangeRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match // Shouldn't happen if regex works
		}

		rangeName := submatches[1]
		content := submatches[2]

		// Look for variable references in the content that aren't dot (current item)
		varRefRegex := regexp.MustCompile(`{{\.([a-zA-Z0-9_]+)}}`)
		varRefs := varRefRegex.FindAllStringSubmatch(content, -1)

		// Nothing to do if no variable references or only self-references
		if len(varRefs) == 0 {
			return match
		}

		// Create a modified content with variable references accessible in range scope
		modifiedContent := content
		for _, ref := range varRefs {
			if len(ref) >= 2 {
				varName := ref[1]
				// Skip if it's referencing the same variable as the range
				if varName == rangeName {
					continue
				}

				// Replace with $ prefix to access the dot context directly
				modifiedContent = strings.ReplaceAll(
					modifiedContent,
					fmt.Sprintf("{{.%s}}", varName),
					fmt.Sprintf("{{$%s}}", varName),
				)

				// Ensure the variable exists in dot context
				if _, exists := dotContext[varName]; exists {
					// Set up variable declaration at the beginning of the range
					modifiedContent = fmt.Sprintf("{{$%s := $.%s}}%s", 
						varName, varName, modifiedContent)
				}
			}
		}

		return fmt.Sprintf("{{range .%s}}%s{{end}}", rangeName, modifiedContent)
	})

	// Create function map with custom functions
	funcMap := template.FuncMap{
		"default": DefaultFunction,
	}

	// Create and execute the template
	tmpl, err := template.New("content").Funcs(funcMap).Parse(processedTemplate)
	if err != nil {
		return "", fmt.Errorf("template parsing error: %w", err)
	}

	// Check for undefined variables before execution
	undefinedVars := FindUndefinedVariables(processedTemplate, processedVars)
	
	// Execute template with variables, making the original context available as "."
	var result bytes.Buffer
	if err := tmpl.Execute(&result, processedVars); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	// Report undefined variables as warnings
	if len(undefinedVars) > 0 {
		return result.String(), fmt.Errorf("undefined template variables: %s", strings.Join(undefinedVars, ", "))
	}

	return result.String(), nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// ValidateTemplate checks a template for potential issues before execution
func ValidateTemplate(templateContent string) error {
	// Basic validation - check if template can be parsed
	funcMap := template.FuncMap{
		"default": DefaultFunction,
	}
	_, err := template.New("validation").Funcs(funcMap).Parse(templateContent)
	if err != nil {
		// Check if this is an unclosed tag error
		if strings.Contains(err.Error(), "unexpected") && strings.Contains(err.Error(), "in define clause") {
			return fmt.Errorf("template contains unclosed tags: %v", err)
		}
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

// DefaultFunction provides a default value for a variable if it doesn't exist or is empty
// Usage in templates: {{.variable_name | default "default value"}}
func DefaultFunction(arg, defaultVal interface{}) interface{} {
	// If arg is nil, return the default value
	if arg == nil {
		return defaultVal
	}
	
	// If arg is a string, check if it's empty
	if s, ok := arg.(string); ok {
		if s == "" {
			return defaultVal
		}
		return s
	}
	
	// Handle other types - return arg if it's not zero-value
	// For simplicity, we just return the original value for non-string types
	return arg
}

// FindUndefinedVariables checks the template for variables that aren't defined in the variables map
// Returns a slice of undefined variable names
func FindUndefinedVariables(templateContent string, variables map[string]interface{}) []string {
	// Regular expression to find all variable references in the template
	varRefRegex := regexp.MustCompile(`{{\.([a-zA-Z0-9_]+)}}`)
	matches := varRefRegex.FindAllStringSubmatch(templateContent, -1)
	
	undefinedVars := make([]string, 0)
	
	// Check each variable reference
	for _, match := range matches {
		if len(match) == 2 {
			varName := match[1]
			// Check if the variable is defined in the variables map
			if _, exists := variables[varName]; !exists {
				// Add to the list of undefined variables if it's not already there
				if !contains(undefinedVars, varName) {
					undefinedVars = append(undefinedVars, varName)
				}
			}
		}
	}
	
	return undefinedVars
}

// IsArrayContext checks if a variable is likely to be used in a range context
// This is a simple heuristic based on common naming patterns
func IsArrayContext(varName string) bool {
	// If the variable explicitly ends with [], it's an array
	if strings.HasSuffix(varName, "[]") {
		return true
	}

	// Check for common plural forms or array-like names
	// Explicitly exclude common words ending with 's' that aren't plurals
	// Also, exclude single letter 's'
	if strings.HasSuffix(varName, "s") && len(varName) > 1 {
		// Exceptions - common words ending with 's' that aren't typically plurals
		exceptions := []string{"address", "status", "business", "process", "canvas", "news", "focus", "bonus"}
		for _, exception := range exceptions {
			if varName == exception {
				return false
			}
		}
		return true
	}
	
	return strings.HasSuffix(varName, "List") || 
		   strings.HasSuffix(varName, "Array") || 
		   strings.HasSuffix(varName, "Items") || 
		   strings.HasSuffix(varName, "Collection")
} 
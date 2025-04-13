// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

// TestDefaultFunctionDirectly tests the default function directly in a template
// to isolate the issue from our ApplyTemplateVariables function
func TestDefaultFunctionDirectly(t *testing.T) {
	// Test the defaultFunction directly first
	result := defaultFunction("World", "Friend")
	t.Logf("Direct function call with \"World\": %v", result)
	assert.Equal(t, "World", result)
	
	// Test with our ApplyTemplateVariables function
	templateStr := "Hello, {{.name | default \"Friend\"}}!"
	
	// Test with existing value
	result1, err := ApplyTemplateVariables(templateStr, map[string]string{"name": "World"})
	assert.NoError(t, err)
	t.Logf("ApplyTemplateVariables result with \"World\": %v", result1)
	assert.Equal(t, "Hello, World!", result1)
	
	// Test with empty value
	result2, err := ApplyTemplateVariables(templateStr, map[string]string{"name": ""})
	assert.NoError(t, err)
	t.Logf("ApplyTemplateVariables result with empty string: %v", result2)
	assert.Equal(t, "Hello, Friend!", result2)
	
	// Test with missing value
	result3, err := ApplyTemplateVariables(templateStr, map[string]string{})
	assert.NoError(t, err)
	t.Logf("ApplyTemplateVariables result with missing value: %v", result3)
	assert.Equal(t, "Hello, Friend!", result3)
	
	// NOTE: The following tests demonstrate why we need our custom implementation
	// to handle default values correctly
	
	// Create a template with the default function
	tmpl := template.New("test")
	
	// Add our default function
	funcMap := template.FuncMap{
		"default": defaultFunction,
	}
	tmpl = tmpl.Funcs(funcMap)
	
	// Parse the template
	parsedTemplate, err := tmpl.Parse(templateStr)
	assert.NoError(t, err)
	
	// Prepare data for the template
	data := map[string]interface{}{"name": "World"}
	t.Logf("Raw template data: %v", data)
	
	// Test with existing value (this fails with standard template.Execute)
	var buf strings.Builder
	err = parsedTemplate.Execute(&buf, data)
	assert.NoError(t, err)
	result = buf.String()
	t.Logf("Standard template result with \"World\": %v", result)
	t.Log("Note: This fails with standard Go templates, which is why we need our custom implementation")
}

func TestApplyTemplateVariables(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		variables   map[string]string
		expected    string
		expectError bool
	}{
		{
			name:        "Basic variable substitution",
			template:    "Hello, {{.name}}!",
			variables:   map[string]string{"name": "World"},
			expected:    "Hello, World!",
			expectError: false,
		},
		{
			name:        "Multiple variables",
			template:    "{{.greeting}}, {{.name}}!",
			variables:   map[string]string{"greeting": "Hello", "name": "World"},
			expected:    "Hello, World!",
			expectError: false,
		},
		{
			name:        "Default function with missing variable",
			template:    "Hello, {{.name | default \"Friend\"}}!",
			variables:   map[string]string{},
			expected:    "Hello, Friend!",
			expectError: false,
		},
		{
			name:        "Default function with empty variable",
			template:    "Hello, {{.name | default \"Friend\"}}!",
			variables:   map[string]string{"name": ""},
			expected:    "Hello, Friend!",
			expectError: false,
		},
		{
			name:        "Default function with existing variable",
			template:    "Hello, {{.name | default \"Friend\"}}!",
			variables:   map[string]string{"name": "World"},
			expected:    "Hello, World!",
			expectError: false,
		},
		{
			name:        "Conditional section (true condition)",
			template:    "{{if .showGreeting}}Hello{{else}}Hi{{end}}, {{.name}}!",
			variables:   map[string]string{"showGreeting": "true", "name": "World"},
			expected:    "Hello, World!",
			expectError: false,
		},
		{
			name:        "Conditional section (false condition)",
			template:    "{{if .showGreeting}}Hello{{else}}Hi{{end}}, {{.name}}!",
			variables:   map[string]string{"showGreeting": "", "name": "World"},
			expected:    "Hi, World!",
			expectError: false,
		},
		{
			name:        "Missing variable (no error)",
			template:    "Hello, {{.unknown}}!",
			variables:   map[string]string{},
			expected:    "Hello, !",
			expectError: false,
		},
		{
			name:        "Invalid template syntax",
			template:    "Hello, {{.name!",
			variables:   map[string]string{"name": "World"},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Unclosed tags",
			template:    "Hello, {{.name",
			variables:   map[string]string{"name": "World"},
			expected:    "",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ApplyTemplateVariables(test.template, test.variables)
			
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if test.name == "Default function with existing variable" {
					// Add debug info for the failing test
					t.Logf("Template: %s", test.template)
					t.Logf("Variables: %v", test.variables)
					t.Logf("Expected: %s", test.expected)
					t.Logf("Actual: %s", result)
					
					// Check behavior of defaultFunction directly
					testVar := test.variables["name"]
					defaultResult := defaultFunction(testVar, "Friend")
					t.Logf("defaultFunction direct result: %v", defaultResult)
				}
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestDefaultFunction(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		defaultValue interface{}
		expected     interface{}
	}{
		{
			name:         "Nil value",
			value:        nil,
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "Empty string",
			value:        "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "Non-empty string",
			value:        "value",
			defaultValue: "default",
			expected:     "value",
		},
		{
			name:         "Integer value",
			value:        42,
			defaultValue: 0,
			expected:     42,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := defaultFunction(test.value, test.defaultValue)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestTemplateProcessor(t *testing.T) {
	processor := NewTemplateProcessor()
	assert.NotNil(t, processor)
	assert.NotNil(t, processor.templateCache)
}

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		expectError bool
	}{
		{
			name:        "Valid template",
			template:    "Hello, {{.name}}!",
			expectError: false,
		},
		{
			name:        "Invalid syntax",
			template:    "Hello, {{.name!",
			expectError: true,
		},
		{
			name:        "Unclosed tags",
			template:    "Hello, {{.name",
			expectError: true,
		},
		{
			name:        "Unbalanced tags",
			template:    "Hello, {{.name}}}}",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateTemplate(test.template)
			
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCustomTemplateHandling(t *testing.T) {
	// TODO: Implement tests for custom template functions and error handling in MVI phase
	t.Skip("Custom template tests will be implemented in MVI phase")
} 
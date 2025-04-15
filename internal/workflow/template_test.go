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
	// Test the DefaultFunction directly first
	result := DefaultFunction("World", "Friend")
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
	// Expect error about undefined variables, but the template should still render
	assert.Equal(t, "Hello, Friend!", result3)
	
	// NOTE: The following tests demonstrate why we need our custom implementation
	// to handle default values correctly
	
	// Create a template with the default function
	tmpl := template.New("test")
	
	// Add our default function
	funcMap := template.FuncMap{
		"default": DefaultFunction,
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
		errorMsg    string // Expected partial error message
	}{
		{
			name:        "Simple variable substitution",
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
			name:        "Default function (preprocessing)",
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
			name:        "Missing variable (warns about undefined variables)",
			template:    "Hello, {{.unknown}}!",
			variables:   map[string]string{},
			expected:    "Hello, !",
			expectError: true,
			errorMsg:    "undefined template variables: unknown",
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
		{
			name:        "Array iteration with range",
			template:    "Items: {{range .items}}{{.}}{{end}}",
			variables:   map[string]string{"items": "one,two,three"},
			expected:    "Items: onetwothree",
			expectError: false,
		},
		{
			name:        "Array iteration with range and delimiter",
			template:    "Items: {{range .items}}{{.}}, {{end}}",
			variables:   map[string]string{"items": "one,two,three"},
			expected:    "Items: one, two, three, ",
			expectError: false,
		},
		{
			name:        "Array iteration with range and conditional",
			template:    "Items: {{range .items}}{{if .}}{{.}}{{else}}empty{{end}}, {{end}}",
			variables:   map[string]string{"items": "one,two,three"},
			expected:    "Items: one, two, three, ",
			expectError: false,
		},
		{
			name:        "Multiple undefined variables",
			template:    "Hello, {{.firstName}} {{.lastName}}!",
			variables:   map[string]string{},
			expected:    "Hello,  !",
			expectError: true,
			errorMsg:    "undefined template variables: firstName, lastName",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ApplyTemplateVariables(test.template, test.variables)
			
			if test.expectError {
				assert.Error(t, err)
				if test.errorMsg != "" {
					assert.Contains(t, err.Error(), test.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				if test.name == "Default function with existing variable" {
					// Add debug info for the failing test
					t.Logf("Template: %s", test.template)
					t.Logf("Variables: %v", test.variables)
					t.Logf("Expected: %s", test.expected)
					t.Logf("Actual: %s", result)
					
					// Check behavior of DefaultFunction directly
					testVar := test.variables["name"]
					defaultResult := DefaultFunction(testVar, "Friend")
					t.Logf("DefaultFunction direct result: %v", defaultResult)
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
			result := DefaultFunction(test.value, test.defaultValue)
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
			err := ValidateTemplate(test.template)
			
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindUndefinedVariables(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		variables   map[string]interface{}
		expected    []string
	}{
		{
			name:        "No undefined variables",
			template:    "Hello, {{.name}}!",
			variables:   map[string]interface{}{"name": "World"},
			expected:    []string{},
		},
		{
			name:        "One undefined variable",
			template:    "Hello, {{.name}}!",
			variables:   map[string]interface{}{},
			expected:    []string{"name"},
		},
		{
			name:        "Multiple undefined variables",
			template:    "Hello, {{.firstName}} {{.lastName}}!",
			variables:   map[string]interface{}{},
			expected:    []string{"firstName", "lastName"},
		},
		{
			name:        "Mix of defined and undefined variables",
			template:    "{{.greeting}}, {{.firstName}} {{.lastName}}!",
			variables:   map[string]interface{}{"greeting": "Hello"},
			expected:    []string{"firstName", "lastName"},
		},
		{
			name:        "Repeated undefined variable",
			template:    "{{.name}} {{.name}} {{.name}}!",
			variables:   map[string]interface{}{},
			expected:    []string{"name"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FindUndefinedVariables(test.template, test.variables)
			assert.ElementsMatch(t, test.expected, result)
		})
	}
}

func TestIsArrayContext(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{
			name:     "Plural form",
			varName:  "items",
			expected: true,
		},
		{
			name:     "List suffix",
			varName:  "itemList",
			expected: true,
		},
		{
			name:     "Array suffix",
			varName:  "itemArray",
			expected: true,
		},
		{
			name:     "Collection suffix",
			varName:  "dataCollection",
			expected: true,
		},
		{
			name:     "Items suffix",
			varName:  "menuItems",
			expected: true,
		},
		{
			name:     "Not an array context",
			varName:  "name",
			expected: false,
		},
		{
			name:     "Not an array context with 's' suffix",
			varName:  "address",
			expected: false,
		},
		{
			name:     "Single letter 's'",
			varName:  "s",
			expected: false,
		},
		{
			name:     "Explicit array suffix []",
			varName:  "data[]",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsArrayContext(test.varName)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsArrayContextDirectly(t *testing.T) {
	// Test with explicit array suffix '[]'
	result := IsArrayContext("data[]")
	assert.True(t, result, "Variable name with '[]' suffix should be recognized as array context")
	
	// Test with plural form
	result = IsArrayContext("items")
	assert.True(t, result, "Variable name with plural 's' should be recognized as array context")
	
	// Test with 'List' suffix
	result = IsArrayContext("itemList")
	assert.True(t, result, "Variable name with 'List' suffix should be recognized as array context")
	
	// Test with non-array name
	result = IsArrayContext("item")
	assert.False(t, result, "Variable name without array indicators should not be recognized as array context")
	
	// Test exception
	result = IsArrayContext("address")
	assert.False(t, result, "Variable name in exceptions list should not be recognized as array context")
}

func TestNestedVariables(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		variables   map[string]string
		expected    string
		expectError bool
	}{
		{
			name:     "Simple nested variable",
			template: "{{.user.name}}",
			variables: map[string]string{
				"user.name": "John",
			},
			expected:    "John",
			expectError: false,
		},
		{
			name:     "Multiple nested variables",
			template: "{{.user.firstname}} {{.user.lastname}}",
			variables: map[string]string{
				"user.firstname": "John",
				"user.lastname":  "Doe",
			},
			expected:    "John Doe",
			expectError: false,
		},
		{
			name:     "Deep nested variables",
			template: "{{.user.address.city}}, {{.user.address.country}}",
			variables: map[string]string{
				"user.address.city":    "New York",
				"user.address.country": "USA",
			},
			expected:    "New York, USA",
			expectError: false,
		},
		{
			name:     "Nested variable with parent also used",
			template: "Name: {{.user.name}}",
			variables: map[string]string{
				"user":      "JohnDoe",
				"user.name": "John",
			},
			expected:    "Name: John",
			expectError: false,
		},
		{
			name:     "Nested array variable with [] suffix",
			template: "{{range .user.hobbies}}{{.}}, {{end}}",
			variables: map[string]string{
				"user.hobbies[]": "reading,gaming,hiking",
			},
			expected:    "reading, gaming, hiking, ",
			expectError: false,
		},
		{
			name:     "Nested array variable with empty value",
			template: "Hobbies: {{range .user.hobbies}}{{.}}{{else}}None{{end}}",
			variables: map[string]string{
				"user.hobbies[]": "",
			},
			expected:    "Hobbies: None",
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ApplyTemplateVariables(test.template, test.variables)
			
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		variables   map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Invalid template syntax",
			template:    "Hello, {{.name | invalid_function}}",
			variables:   map[string]string{"name": "World"},
			expectError: true,
			errorMsg:    "template validation error",
		},
		{
			name:        "Default value with invalid format",
			template:    "{{.name | default missing_quotes}}",
			variables:   map[string]string{},
			expectError: true,
			errorMsg:    "template validation error",
		},
		{
			name:        "Unclosed define clause",
			template:    "{{define \"template\"}}unclosed",
			variables:   map[string]string{},
			expectError: true,
			errorMsg:    "template validation error",
		},
		{
			name: "Template execution error",
			template: `{{range .invalid}}
				{{.value}}
			{{end}}`,
			variables: map[string]string{
				"invalid": "not-an-array", // This will cause execution error
			},
			expectError: true,
			errorMsg:    "template execution error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ApplyTemplateVariables(test.template, test.variables)
			assert.Error(t, err)
			if test.errorMsg != "" {
				assert.Contains(t, err.Error(), test.errorMsg)
			}
		})
	}
}

func TestNestedPathsAndArrayContext(t *testing.T) {
	// Test template with nested paths and array contexts
	templateStr := `
User: {{.user.name}}
Address: {{.user.address.city}}, {{.user.address.country}}
Skills: {{range .user.skills}}{{.}}, {{end}}
Hobbies: {{range .hobbies}}{{.}}, {{end}}
`
	variables := map[string]string{
		"user.name":           "John Doe",
		"user.address.city":   "New York",
		"user.address.country": "USA",
		"user.skills[]":       "programming,design,management",
		"hobbies[]":           "reading,hiking",
		"user":                "Override attempt", // Should be overridden by the nested structure
	}

	// This test should cover many of the nested path handling code paths
	result, err := ApplyTemplateVariables(templateStr, variables)
	assert.NoError(t, err)
	
	expected := `
User: John Doe
Address: New York, USA
Skills: programming, design, management, 
Hobbies: reading, hiking, 
`
	assert.Equal(t, expected, result)
	
	// Test empty array with [] suffix
	variables = map[string]string{
		"items[]": "",
	}
	
	result, err = ApplyTemplateVariables("Items: {{range .items}}{{.}}{{else}}None{{end}}", variables)
	assert.NoError(t, err)
	assert.Equal(t, "Items: None", result)
	
	// Test malformed template to trigger parsing error
	_, err = ApplyTemplateVariables("{{if .condition}}No {{end", variables)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template validation error")
	
	// Test invalid template execution
	_, err = ApplyTemplateVariables("{{range .nonarray}}{{.}}{{end}}", map[string]string{
		"nonarray": "not-an-array",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template execution error")
	
	// Test regex for default values with invalid format
	_, err = ApplyTemplateVariables(`{{.name | default "missing}}`, variables)
	assert.Error(t, err)
}

func TestArrayContextChecking(t *testing.T) {
	// Test with explicit array suffix '[]'
	testCases := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"Empty string", "", false},
		{"Regular variable", "name", false},
		{"Array suffix", "items[]", true},
		{"Mixed suffix", "items[test]", false},
		{"Plural form", "items", true},
		{"Exception word", "address", false},
		{"List suffix", "userList", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsArrayContext(tc.varName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTemplateErrorHandling(t *testing.T) {
	// Test various error conditions
	testCases := []struct {
		name          string
		template      string
		variables     map[string]string
		expectedError string
	}{
		{
			name:          "Invalid template syntax",
			template:      "Hello, {{.name!",
			variables:     map[string]string{},
			expectedError: "template validation error",
		},
		{
			name:          "Unclosed tag",
			template:      "Hello, {{.name",
			variables:     map[string]string{},
			expectedError: "template validation error",
		},
		{
			name:          "Invalid default function",
			template:      "Hello, {{.name | default invalid}}",
			variables:     map[string]string{},
			expectedError: "template validation error",
		},
		{
			name:          "Unbalanced tag count",
			template:      "Hello, {{.name}} {{.other}",
			variables:     map[string]string{},
			expectedError: "template validation error",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ApplyTemplateVariables(tc.template, tc.variables)
			assert.Error(t, err)
			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestEdgeCaseHandling(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		variables   map[string]string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Undefined variable with dollar sign",
			template:    "{{$undefined}}",
			variables:   map[string]string{},
			expected:    "",
			expectError: true,
			errorMsg:    "template validation error: template: validation:1: undefined variable",
		},
		{
			name:        "Nested error in range",
			template:    "{{range .items}}{{if .missing}}{{end}}{{end}}",
			variables:   map[string]string{"items": "one,two"},
			expected:    "",
			expectError: true,
			errorMsg:    "execution error",
		},
		{
			name:        "Unclosed condition",
			template:    "{{if .condition}}Unclosed",
			variables:   map[string]string{"condition": "true"},
			expected:    "",
			expectError: true,
			errorMsg:    "template validation error: template: validation:1: unexpected EOF",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ApplyTemplateVariables(test.template, test.variables)

			if test.expectError {
				assert.Error(t, err)
				if test.errorMsg != "" {
					assert.Contains(t, err.Error(), test.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestCustomTemplateHandling(t *testing.T) {
	// Test advanced template features and error handling
	tests := []struct {
		name        string
		template    string
		variables   map[string]string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Nested conditional statements",
			template:    "{{if .condition1}}{{if .condition2}}Both true{{else}}Only condition1{{end}}{{else}}None true{{end}}",
			variables:   map[string]string{"condition1": "true", "condition2": "true"},
			expected:    "Both true",
			expectError: false,
		},
		{
			name:        "Template with whitespace control",
			template:    "{{- if .show -}}Compact{{- else -}}Also compact{{- end -}}",
			variables:   map[string]string{"show": "true"},
			expected:    "Compact",
			expectError: false,
		},
		{
			name:        "Complex range with indexing",
			template:    "{{range $index, $element := .items}}[{{$index}}:{{$element}}]{{end}}",
			variables:   map[string]string{"items": "one,two,three"},
			expected:    "[0:one][1:two][2:three]",
			expectError: false,
		},
		{
			name:        "Range with empty input",
			template:    "Items: {{range .emptyList}}{{.}}{{else}}No items{{end}}",
			variables:   map[string]string{"emptyList": ""},
			expected:    "Items: No items",
			expectError: false,
		},
		{
			name:        "Complex_nested_ranges",
			template:    "{{range .categories}}{{.}}:{{range $index, $element := $.subcategories}}[{{$element}}]{{end}};{{end}}",
			variables:   map[string]string{
				"categories": "cat1,cat2",
				"subcategories": "sub1,sub2",
			},
			expected:    "cat1:[sub1][sub2];cat2:[sub1][sub2];",
			expectError: false,
		},
		{
			name:        "Template_error: missing_end_tag",
			template:    "{{if .condition}}No end tag",
			variables:   map[string]string{"condition": "true"},
			expected:    "",
			expectError: true,
			errorMsg:    "unexpected EOF",
		},
		{
			name:        "Reference to variable in non-variable scope",
			template:    "{{.outer}} and {{range .items}}{{.outer}}{{end}}",
			variables:   map[string]string{"outer": "outside", "items": "one,two"},
			expected:    "outside and outsideoutside",
			expectError: false,
		},
	}

	for _, test := range tests {
		// The "Reference to variable in non-variable scope" test is now supported
		result, err := ApplyTemplateVariables(test.template, test.variables)
		
		if test.expectError {
			assert.Error(t, err)
			if test.errorMsg != "" {
				assert.Contains(t, err.Error(), test.errorMsg)
			}
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	}
} 
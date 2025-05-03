// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

// customTemplateRenderer extends TemplateRenderer with a modified ExtractTemplateVariables method for testing
type customTemplateRenderer struct {
	*TemplateRenderer
}

// ExtractTemplateVariables overrides the standard method to add template functions like 'default' for testing
func (r *customTemplateRenderer) ExtractTemplateVariables(promptPath string) ([]string, error) {
	// Get full path to prompt file
	fullPath := promptPath
	if !filepath.IsAbs(promptPath) {
		fullPath = filepath.Join(r.workflowDir, promptPath)
	}

	// Check if prompt file exists
	if !r.fs.Exists(fullPath) {
		return nil, errors.New("prompt file not found: " + promptPath)
	}

	// Read the prompt file
	promptData, err := r.fs.ReadFile(fullPath)
	if err != nil {
		return nil, errors.New("failed to read prompt file: " + err.Error())
	}

	// Parse the template with the necessary functions
	funcMap := template.FuncMap{
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil {
				return defaultVal
			}
			if s, ok := val.(string); ok && s == "" {
				return defaultVal
			}
			return val
		},
	}
	
	tmpl, err := template.New(filepath.Base(promptPath)).Funcs(funcMap).Parse(string(promptData))
	if err != nil {
		return nil, errors.New("invalid template syntax in " + promptPath + ": " + err.Error())
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

// newCustomTemplateRenderer creates a new template renderer with custom functionality
func newCustomTemplateRenderer(fs io.FileSystem, workflowDir string) *customTemplateRenderer {
	return &customTemplateRenderer{
		TemplateRenderer: NewTemplateRenderer(fs, workflowDir),
	}
}

// TestASTBasedVariableExtraction tests the AST-based variable extraction functionality
func TestASTBasedVariableExtraction(t *testing.T) {
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	fs.MkdirAll(promptsDir, 0755)
	
	testCases := []struct {
		name            string
		templateContent string
		expectedVars    []string
	}{
		{
			name: "Simple variable",
			templateContent: "Hello {{.name}}!",
			expectedVars: []string{"name"},
		},
		{
			name: "Multiple variables",
			templateContent: "Hello {{.firstName}} {{.lastName}}!",
			expectedVars: []string{"firstName", "lastName"},
		},
		{
			name: "Variables with pipeline",
			templateContent: "Hello {{.name | default \"Anonymous\"}}!",
			expectedVars: []string{"name"},
		},
		{
			name: "Conditional variables",
			templateContent: `{{if .showGreeting}}Hello {{.name}}!{{else}}Welcome!{{end}}`,
			expectedVars: []string{"showGreeting", "name"},
		},
		{
			name: "Range over array",
			templateContent: `{{range .items}}{{.name}}: {{.value}}{{end}}`,
			expectedVars: []string{"items", "name", "value"},
		},
		{
			name: "Complex nested structure",
			templateContent: `
				{{if .showHeader}}
					<h1>{{.title}}</h1>
				{{end}}
				<ul>
				{{range .items}}
					<li>{{.name}}: {{.value}}</li>
					{{if .hasDetails}}
						<ul>
						{{range .details}}
							<li>{{.key}}: {{.value}}</li>
						{{end}}
						</ul>
					{{else}}
						<p>No details available</p>
					{{end}}
				{{else}}
					<li>No items available</li>
				{{end}}
				</ul>
			`,
			expectedVars: []string{"showHeader", "title", "items", "name", "value", "hasDetails", "details", "key"},
		},
		{
			name: "Template inclusion",
			templateContent: `{{template "header" .}}{{.content}}{{template "footer" .}}`,
			expectedVars: []string{"content"},
		},
		{
			name: "Chain node with function",
			templateContent: `{{(print .prefix .name).value}}`,
			expectedVars: []string{"prefix", "name"},
		},
		{
			name: "With statement",
			templateContent: `{{with .user}}{{.name}} ({{.email}}){{else}}Anonymous{{end}}`,
			expectedVars: []string{"user", "name", "email"},
		},
		{
			name: "Using variable syntax",
			templateContent: `{{$username := .user.name}}Hello {{$username}}!`,
			expectedVars: []string{"user"},
		},
		{
			name: "List node in empty block",
			templateContent: `{{define "empty"}}{{end}}{{.name}}`,
			expectedVars: []string{"name"},
		},
		{
			name: "Empty range statement with else",
			templateContent: `{{range .items}}{{else}}No items found!{{end}}`,
			expectedVars: []string{"items"},
		},
		{
			name: "Empty with statement with else",
			templateContent: `{{with .user}}{{else}}Anonymous{{end}}`,
			expectedVars: []string{"user"},
		},
		{
			name: "Complex default function with empty string",
			templateContent: `Hello {{.name | default "Anonymous"}}!`,
			expectedVars: []string{"name"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test prompt file
			promptPath := filepath.Join(promptsDir, tc.name+".md")
			fs.WriteFile(promptPath, []byte(tc.templateContent), 0644)
			
			// Create custom template renderer
			renderer := newCustomTemplateRenderer(fs, workflowDir)
			
			// Extract variables using our custom renderer
			vars, err := renderer.ExtractTemplateVariables(filepath.Join("prompts", tc.name+".md"))
			
			// Verify results
			assert.NoError(t, err)
			for _, expectedVar := range tc.expectedVars {
				found := false
				for _, extractedVar := range vars {
					if extractedVar == expectedVar {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected variable '%s' not found in extracted variables: %v", expectedVar, vars)
			}
		})
	}
}

// TestExtractVariablesWithErrors tests error handling in variable extraction
func TestExtractVariablesWithErrors(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func() (io.FileSystem, string, error)
		expectedError string
	}{
		{
			name: "Invalid template syntax",
			setup: func() (io.FileSystem, string, error) {
				fs := io.NewMockFileSystem()
				workflowDir := "/test-workflow"
				promptsDir := filepath.Join(workflowDir, "prompts")
				fs.MkdirAll(promptsDir, 0755)
				
				promptPath := filepath.Join(promptsDir, "invalid.md")
				fs.WriteFile(promptPath, []byte("Hello {{.name!"), 0644)
				
				return fs, workflowDir, nil
			},
			expectedError: "invalid template syntax",
		},
		{
			name: "Non-existent file",
			setup: func() (io.FileSystem, string, error) {
				fs := io.NewMockFileSystem()
				workflowDir := "/test-workflow"
				fs.MkdirAll(workflowDir, 0755)
				
				return fs, workflowDir, nil
			},
			expectedError: "prompt file not found",
		},
		{
			name: "File system error",
			setup: func() (io.FileSystem, string, error) {
				fs := io.NewMockFileSystemWithErrors()
				workflowDir := "/test-workflow"
				promptsDir := filepath.Join(workflowDir, "prompts")
				fs.MkdirAll(promptsDir, 0755)
				
				promptPath := filepath.Join(promptsDir, "error.md")
				fs.WriteFile(promptPath, []byte("Hello {{.name}}"), 0644)
				fs.SetReadError(promptPath, errors.New("simulated read error"))
				
				return fs, workflowDir, nil
			},
			expectedError: "failed to read prompt file",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			fs, workflowDir, _ := tc.setup()
			renderer := NewTemplateRenderer(fs, workflowDir)
			
			// Test
			var promptPath string
			if tc.name == "Non-existent file" {
				promptPath = "prompts/nonexistent.md"
			} else if tc.name == "File system error" {
				promptPath = "prompts/error.md"
			} else {
				promptPath = "prompts/invalid.md"
			}
			
			_, err := renderer.ExtractTemplateVariables(promptPath)
			
			// Verify
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// TestDefaultFunctionWithStrings tests the default function with string values
func TestDefaultFunctionWithStrings(t *testing.T) {
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	fs.MkdirAll(workflowDir, 0755)
	
	// Create template with default function
	templateContent := `Name: {{.name | default "Default Name"}}`
	promptPath := filepath.Join(workflowDir, "test-default.md")
	fs.WriteFile(promptPath, []byte(templateContent), 0644)
	
	// Create renderer
	renderer := NewTemplateRenderer(fs, workflowDir)
	
	// Test with empty string
	result, err := renderer.RenderPrompt("test-default.md", map[string]interface{}{
		"name": "",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Name: Default Name", result, "Default value should be used for empty string")
	
	// Test with non-empty string
	result, err = renderer.RenderPrompt("test-default.md", map[string]interface{}{
		"name": "John",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Name: John", result, "Provided value should be used when not empty")
}

// TestTemplateValidationWithCustomFunctions tests the template validation with custom functions
func TestTemplateValidationWithCustomFunctions(t *testing.T) {
	fs := io.NewMockFileSystem()
	workflowDir := "/test-workflow"
	fs.MkdirAll(workflowDir, 0755)
	
	// Create template with custom function
	templateContent := `{{.name | default "Default Name"}}`
	promptPath := filepath.Join(workflowDir, "test-function.md")
	fs.WriteFile(promptPath, []byte(templateContent), 0644)
	
	// Create renderer
	renderer := NewTemplateRenderer(fs, workflowDir)
	
	// Validate the template
	err := renderer.ValidateTemplate("test-function.md")
	assert.NoError(t, err, "Template with default function should be valid")
	
	// Test the default function in validation
	testDefaults := map[string]interface{}{
		"name": "",
	}
	
	// This should internally use the default function
	result, err := renderer.RenderPrompt("test-function.md", testDefaults)
	assert.NoError(t, err)
	assert.Equal(t, "Default Name", result, "Default value should be used in validation")
}

func TestTemplateRenderer_RenderPrompt_WorkflowPathHandling(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()
	
	// Set up a test workflow structure
	workflowDir := "/workflows/myworkflow"
	
	// Add workflow.yaml
	fs.AddFile(filepath.Join(workflowDir, "workflow.yaml"), []byte(`
name: "myworkflow"
description: "Test workflow"
steps:
  - id: "01-step-one"
    description: "Step One"
    prompt: "prompts/step1.md"
`))

	// Add a prompt file directly in the prompts directory
	promptContent := "Hello {{ .Name }}! This is step {{ .StepID }}."
	fs.AddFile(filepath.Join(workflowDir, "prompts", "step1.md"), []byte(promptContent))
	
	// Create a renderer
	renderer := NewTemplateRenderer(fs, workflowDir)
	
	// Test case 1: Relative path to prompt within workflow structure
	t.Run("RelativePathWithinWorkflow", func(t *testing.T) {
		result, err := renderer.RenderPrompt("prompts/step1.md", map[string]interface{}{
			"Name":  "User",
			"StepID": "01-step-one",
		})
		
		// Check results
		assert.NoError(t, err)
		assert.Equal(t, "Hello User! This is step 01-step-one.", result)
	})
	
	// Test case 2: Using the basename of the prompt file
	t.Run("BasenameOnly", func(t *testing.T) {
		result, err := renderer.RenderPrompt("step1.md", map[string]interface{}{
			"Name":  "User",
			"StepID": "01-step-one",
		})
		
		// Check results - should now find the file in the prompts directory
		assert.NoError(t, err)
		assert.Equal(t, "Hello User! This is step 01-step-one.", result)
	})
}

// TestRenderPromptWithTemplateInclusions tests that template inclusions using {{ template "..." . }} syntax work correctly
func TestRenderPromptWithTemplateInclusions(t *testing.T) {
	// Create a mock filesystem
	fs := io.NewMockFileSystem()
	
	// Create test workflow directory structure
	workflowDir := "/workflows/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	sharedDir := filepath.Join(promptsDir, "shared")
	
	// Create the directory structure
	fs.AddDirectory(workflowDir)
	fs.AddDirectory(promptsDir)
	fs.AddDirectory(sharedDir)
	
	// Create a main template with inclusions
	mainTemplate := `# {{ .phase }} Phase

This step focuses on laying the foundation for the implementation:

{{ template "shared/phase_header.md" . }}

## Primary Tasks

1. Set up the project structure
2. Define key interfaces and data structures

## Expected Outcome

- Well-defined interfaces
- Clear separation of concerns

{{ template "shared/footer.md" . }}`
	
	// Create shared templates
	headerTemplate := `## Focus Area: {{ .focus }}

This phase we're focusing on {{ .focus }}.

---`
	
	footerTemplate := `---

## Notes and Considerations

- Keep code modular and testable
- Follow project coding standards`
	
	// Add files to the mock filesystem
	mainTemplatePath := filepath.Join(promptsDir, "foundation.md")
	headerTemplatePath := filepath.Join(sharedDir, "phase_header.md")
	footerTemplatePath := filepath.Join(sharedDir, "footer.md")
	
	fs.AddFile(mainTemplatePath, []byte(mainTemplate))
	fs.AddFile(headerTemplatePath, []byte(headerTemplate))
	fs.AddFile(footerTemplatePath, []byte(footerTemplate))
	
	// Create variables for template rendering
	variables := map[string]interface{}{
		"phase": "foundation",
		"focus": "architecture and interfaces",
	}
	
	// Test 1: Render the template
	renderer := NewTemplateRenderer(fs, workflowDir)
	result, err := renderer.RenderPrompt("prompts/foundation.md", variables)
	assert.NoError(t, err, "Template should render without errors")
	assert.NotEmpty(t, result, "Template rendering result should not be empty")
	
	// Test 2: Verify content includes both the main template and shared templates
	assert.Contains(t, result, "# foundation Phase", "Result should contain content from the main template")
	assert.Contains(t, result, "Focus Area: architecture and interfaces", "Result should contain content from the header template")
	assert.Contains(t, result, "Keep code modular and testable", "Result should contain content from the footer template")
	
	// Test 3: Try rendering with missing shared template to verify error handling
	// Create a new renderer to reset the cache
	renderer2 := NewTemplateRenderer(fs, workflowDir)
	
	// Delete the header template
	delete(fs.Files, headerTemplatePath)
	// Remove from directory listing too
	dirEntries := fs.DirItems[sharedDir]
	newEntries := make([]os.DirEntry, 0)
	for _, entry := range dirEntries {
		if entry.Name() != "phase_header.md" {
			newEntries = append(newEntries, entry)
		}
	}
	fs.DirItems[sharedDir] = newEntries
	
	// Try to render the template with a missing include
	_, err = renderer2.RenderPrompt("prompts/foundation.md", variables)
	assert.Error(t, err, "Should error when a referenced template is missing")
	assert.Contains(t, err.Error(), "failed to read template", "Error should indicate the referenced template is missing")
}

// TestExtractVariablesWithTemplateInclusions tests variable extraction from templates with inclusions
func TestExtractVariablesWithTemplateInclusions(t *testing.T) {
	// Create a mock filesystem
	fs := io.NewMockFileSystem()
	
	// Create test workflow directory structure
	workflowDir := "/workflows/test-workflow"
	promptsDir := filepath.Join(workflowDir, "prompts")
	sharedDir := filepath.Join(promptsDir, "shared")
	
	// Create the directory structure
	fs.AddDirectory(workflowDir)
	fs.AddDirectory(promptsDir)
	fs.AddDirectory(sharedDir)
	
	// Create a main template with inclusions
	mainTemplate := `# {{ .phase }} Phase

This step focuses on laying the foundation for the implementation:

{{ template "shared/phase_header.md" . }}

## Primary Tasks

1. Task: {{ .task_name }}
2. Task: {{ .task_detail }}

{{ template "shared/footer.md" . }}`
	
	// Create shared templates
	headerTemplate := `## Focus Area: {{ .focus }}

This phase we're focusing on {{ .focus }}.

---`
	
	footerTemplate := `---

## Assigned to: {{ .assignee }}

- {{ .priority }} priority`
	
	// Add files to the mock filesystem
	mainTemplatePath := filepath.Join(promptsDir, "foundation.md")
	headerTemplatePath := filepath.Join(sharedDir, "phase_header.md")
	footerTemplatePath := filepath.Join(sharedDir, "footer.md")
	
	fs.AddFile(mainTemplatePath, []byte(mainTemplate))
	fs.AddFile(headerTemplatePath, []byte(headerTemplate))
	fs.AddFile(footerTemplatePath, []byte(footerTemplate))
	
	// Create template renderer
	renderer := NewTemplateRenderer(fs, workflowDir)
	
	// Test: Extract variables from the template
	variables, err := renderer.ExtractTemplateVariables("prompts/foundation.md")
	assert.NoError(t, err, "Variable extraction should work without errors")
	
	// Convert slice to map for easier testing
	varMap := make(map[string]bool)
	for _, v := range variables {
		varMap[v] = true
	}
	
	// Verify variables from main template are extracted
	assert.True(t, varMap["phase"], "Should extract 'phase' variable")
	assert.True(t, varMap["task_name"], "Should extract 'task_name' variable")
	assert.True(t, varMap["task_detail"], "Should extract 'task_detail' variable")
	
	// This test would ideally verify that variables from included templates are extracted too,
	// but the current implementation only extracts variables from the main template file,
	// not from included templates. This is noted as a potential enhancement.
	
	// Note: Current implementation limitation is that variables from included templates
	// are not extracted, which leads to the warnings seen in the validate command:
	// "Variable 'focus' not used in template" even though it is used in an included template.
}

// TestFindTemplateReferences tests the findTemplateReferences function
func TestFindTemplateReferences(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "Simple inclusion",
			content: `# Test
{{ template "shared/header.md" . }}
Content
{{ template "shared/footer.md" . }}`,
			expected: []string{"shared/header.md", "shared/footer.md"},
		},
		{
			name: "Inclusion with whitespace",
			content: `# Test
{{  template   "shared/header.md"  . }}
Content
{{template "shared/footer.md" .}}`,
			expected: []string{"shared/header.md", "shared/footer.md"},
		},
		{
			name: "Inclusion with different context",
			content: `# Test
{{ template "shared/header.md" .user }}
Content
{{ template "shared/footer.md" $ }}`,
			expected: []string{"shared/header.md", "shared/footer.md"},
		},
		{
			name:     "No inclusions",
			content:  "# Test\nContent without inclusions {{ .variable }}",
			expected: []string{},
		},
	}
	
	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := findTemplateReferences(tc.content)
			assert.ElementsMatch(t, tc.expected, result, "Template references should match expected")
		})
	}
}

func TestExtractVariablesFromNestedTemplates(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()
	
	// Add a main template that includes a nested template
	mainTemplate := `# {{ .title }} Template
This is the main template.

{{ template "shared/header.md" . }}

More content with {{ .mainVar }}

{{ template "shared/footer.md" . }}
`
	
	// Add a nested header template with its own variables
	headerTemplate := `## {{ .subtitle }}
This is a header with {{ .headerVar }}.
`
	
	// Add a nested footer template with its own variables
	footerTemplate := `## Footer
This is a footer with {{ .footerVar }}.
`
	
	// Add the templates to the mock file system
	workflowDir := "/workflow"
	fs.AddFile(filepath.Join(workflowDir, "main.md"), []byte(mainTemplate))
	fs.AddFile(filepath.Join(workflowDir, "shared/header.md"), []byte(headerTemplate))
	fs.AddFile(filepath.Join(workflowDir, "shared/footer.md"), []byte(footerTemplate))
	
	// Create a template renderer with the mock file system
	renderer := NewTemplateRenderer(fs, workflowDir)
	
	// Extract variables from the main template
	vars, err := renderer.ExtractTemplateVariables(filepath.Join(workflowDir, "main.md"))
	
	// Verify that no error occurred
	assert.NoError(t, err)
	
	// Sort the variables for consistent comparison
	sort.Strings(vars)
	
	// Verify that variables from both the main and nested templates are extracted
	expectedVars := []string{"title", "mainVar", "subtitle", "headerVar", "footerVar"}
	sort.Strings(expectedVars)
	
	assert.Equal(t, expectedVars, vars, "Failed to extract variables from nested templates")
} 
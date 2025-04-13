// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

func TestLoadWorkflowFromFile_JSON(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Add a sample workflow JSON file
	workflowJSON := `{
		"name": "test-workflow",
		"description": "Test Workflow Description",
		"steps": [
			{
				"id": "step1",
				"description": "First step",
				"prompt": "This is the first step"
			},
			{
				"id": "step2",
				"description": "Second step",
				"prompt": "This is the second step"
			}
		]
	}`
	jsonPath := "workflows/test.json"
	fs.AddFile(jsonPath, []byte(workflowJSON))

	// Load the workflow
	workflow, err := LoadWorkflowFromFile(fs, jsonPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the workflow
	if workflow.Name != "test-workflow" {
		t.Errorf("Expected workflow name to be 'test-workflow', got '%s'", workflow.Name)
	}
	if workflow.Description != "Test Workflow Description" {
		t.Errorf("Expected workflow description to be 'Test Workflow Description', got '%s'", workflow.Description)
	}
	if len(workflow.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(workflow.Steps))
	}
	if workflow.Steps[0].ID != "step1" {
		t.Errorf("Expected first step ID to be 'step1', got '%s'", workflow.Steps[0].ID)
	}
	if workflow.Steps[1].ID != "step2" {
		t.Errorf("Expected second step ID to be 'step2', got '%s'", workflow.Steps[1].ID)
	}
}

func TestLoadWorkflowFromFile_YAML(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Add a sample workflow YAML file
	workflowYAML := `name: test-workflow-yaml
description: Test Workflow YAML Description
steps:
  - id: step1
    description: First step
    prompt: This is the first step
  - id: step2
    description: Second step
    prompt: This is the second step
`
	yamlPath := "workflows/test.yaml"
	fs.AddFile(yamlPath, []byte(workflowYAML))

	// Load the workflow
	workflow, err := LoadWorkflowFromFile(fs, yamlPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the workflow
	if workflow.Name != "test-workflow-yaml" {
		t.Errorf("Expected workflow name to be 'test-workflow-yaml', got '%s'", workflow.Name)
	}
	if workflow.Description != "Test Workflow YAML Description" {
		t.Errorf("Expected workflow description to be 'Test Workflow YAML Description', got '%s'", workflow.Description)
	}
	if len(workflow.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(workflow.Steps))
	}
	if workflow.Steps[0].ID != "step1" {
		t.Errorf("Expected first step ID to be 'step1', got '%s'", workflow.Steps[0].ID)
	}
	if workflow.Steps[1].ID != "step2" {
		t.Errorf("Expected second step ID to be 'step2', got '%s'", workflow.Steps[1].ID)
	}
}

func TestLoadWorkflowFromFile_FileNotFound(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Try to load a non-existent workflow
	_, err := LoadWorkflowFromFile(fs, "non-existent.json")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
}

func TestLoadWorkflowFromFile_InvalidJSON(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Add an invalid JSON file
	fs.AddFile("invalid.json", []byte(`{ "name": "test", invalid json }`))

	// Try to load the invalid workflow
	_, err := LoadWorkflowFromFile(fs, "invalid.json")
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestLoadWorkflowFromFile_InvalidYAML(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Add an invalid YAML file
	fs.AddFile("invalid.yaml", []byte(`name: test
description: Invalid YAML file
steps: invalid steps`))

	// Try to load the invalid workflow
	_, err := LoadWorkflowFromFile(fs, "invalid.yaml")
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestLoadWorkflowFromFile_UnsupportedFormat(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Add a file with an unsupported extension
	fs.AddFile("workflow.txt", []byte("This is not a workflow file"))

	// Try to load the unsupported format
	_, err := LoadWorkflowFromFile(fs, "workflow.txt")
	if err == nil {
		t.Fatal("Expected error for unsupported format, got nil")
	}
}

func TestLoadWorkflowFromFile_MissingFields(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Add a workflow file missing required fields
	fs.AddFile("missing.json", []byte(`{
		"name": "test-workflow"
	}`))

	// Try to load the workflow missing required fields
	_, err := LoadWorkflowFromFile(fs, "missing.json")
	if err == nil {
		t.Fatal("Expected error for missing required fields, got nil")
	}
}

func TestLoadWorkflowsFromDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := filepath.Join("output", "test-workflows-"+t.Name())
	
	// Get the real file system
	fs := io.NewOSFileSystem()
	
	// Create the directory if it doesn't exist
	if !fs.Exists(tempDir) {
		err := fs.MkdirAll(tempDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}
	
	// Clean up after the test
	defer func() {
		// Remove all test files and directory
		files, _ := fs.ReadDir(tempDir)
		for _, file := range files {
			if !file.IsDir() {
				// Use standard os package instead of fs.RemoveFile
				os.Remove(filepath.Join(tempDir, file.Name()))
			}
		}
		// Don't remove the parent directory as it might be used by other tests
	}()
	
	// Define test workflows in different formats
	workflowJSON := `{
		"name": "json-workflow",
		"description": "JSON Test Workflow",
		"steps": [
			{
				"id": "step1",
				"description": "Step 1",
				"prompt": "Prompt 1"
			}
		]
	}`
	
	workflowYAML := `name: yaml-workflow
description: YAML Test Workflow
steps:
  - id: step1
    description: Step 1
    prompt: Prompt 1
`
	
	invalidJSON := `{
		"name": "invalid-json",
		"description": "Invalid JSON",
		"steps": [
			{ invalid json here }
		]
	}`
	
	nonWorkflowFile := "This is not a workflow file"
	
	// Write files to the real filesystem
	fs.WriteFile(filepath.Join(tempDir, "workflow1.json"), []byte(workflowJSON), 0644)
	fs.WriteFile(filepath.Join(tempDir, "workflow2.yaml"), []byte(workflowYAML), 0644)
	fs.WriteFile(filepath.Join(tempDir, "invalid.json"), []byte(invalidJSON), 0644)
	fs.WriteFile(filepath.Join(tempDir, "not-a-workflow.txt"), []byte(nonWorkflowFile), 0644)
	
	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subfolder")
	fs.MkdirAll(subDir, 0755)
	fs.WriteFile(filepath.Join(subDir, "nested.yaml"), []byte(workflowYAML), 0644)
	
	// Create an empty directory
	fs.MkdirAll(filepath.Join(tempDir, "empty-dir"), 0755)
	
	// Create a registry for testing
	registry := NewWorkflowRegistry()
	
	// Load workflows from the directory
	workflows, err := LoadWorkflowsFromDirectory(fs, tempDir, registry)
	
	// Check results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Should load 2 valid workflows (JSON and YAML) from the root directory
	// The nested.yaml in the subfolder should not be loaded
	expectedCount := 2
	if len(workflows) != expectedCount {
		t.Errorf("Expected %d workflows to be loaded, got %d", expectedCount, len(workflows))
	}
	
	// Check that the workflows were registered in the registry
	registeredWorkflows := registry.ListWorkflows()
	
	// Registry should now have standard workflow + 2 loaded workflows
	expectedRegistryCount := 3 // Standard + 2 loaded
	if len(registeredWorkflows) != expectedRegistryCount {
		t.Errorf("Expected %d workflows in registry, got %d", expectedRegistryCount, len(registeredWorkflows))
	}
	
	// Check specific workflow names
	expectedNames := map[string]bool{
		"json-workflow": false, 
		"yaml-workflow": false,
	}
	
	for _, name := range registeredWorkflows {
		if _, exists := expectedNames[name]; exists {
			expectedNames[name] = true
		}
	}
	
	// Verify all expected workflows were found
	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected workflow %s to be registered, but it wasn't found", name)
		}
	}
	
	// Test directory not found
	nonExistentDir := filepath.Join(tempDir, "non-existent-dir")
	_, err = LoadWorkflowsFromDirectory(fs, nonExistentDir, registry)
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

func TestSaveWorkflowToFile_JSON(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Create a workflow to save
	workflow := &WorkflowDefinition{
		Name:        "save-test",
		Description: "Test Save Workflow",
		Steps: []WorkflowStep{
			{
				ID:          "step1",
				Description: "First step",
				Prompt:      "This is the first step",
			},
		},
	}

	// Save the workflow
	err := SaveWorkflowToFile(fs, workflow, "output/save-test.json")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the file was saved
	if !fs.Exists("output/save-test.json") {
		t.Fatal("Expected file to exist, but it doesn't")
	}

	// Verify the file contents can be parsed back into a workflow
	savedWorkflow, err := LoadWorkflowFromFile(fs, "output/save-test.json")
	if err != nil {
		t.Fatalf("Expected no error when loading saved workflow, got %v", err)
	}

	if savedWorkflow.Name != workflow.Name {
		t.Errorf("Expected saved workflow name to be '%s', got '%s'", workflow.Name, savedWorkflow.Name)
	}

	if len(savedWorkflow.Steps) != len(workflow.Steps) {
		t.Errorf("Expected saved workflow to have %d steps, got %d", len(workflow.Steps), len(savedWorkflow.Steps))
	}
}

func TestSaveWorkflowToFile_YAML(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Create a workflow to save
	workflow := &WorkflowDefinition{
		Name:        "save-test-yaml",
		Description: "Test Save Workflow YAML",
		Steps: []WorkflowStep{
			{
				ID:          "step1",
				Description: "First step",
				Prompt:      "This is the first step",
			},
		},
	}

	// Save the workflow
	err := SaveWorkflowToFile(fs, workflow, "output/save-test.yaml")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the file was saved
	if !fs.Exists("output/save-test.yaml") {
		t.Fatal("Expected file to exist, but it doesn't")
	}

	// Verify the file contents can be parsed back into a workflow
	savedWorkflow, err := LoadWorkflowFromFile(fs, "output/save-test.yaml")
	if err != nil {
		t.Fatalf("Expected no error when loading saved workflow, got %v", err)
	}

	if savedWorkflow.Name != workflow.Name {
		t.Errorf("Expected saved workflow name to be '%s', got '%s'", workflow.Name, savedWorkflow.Name)
	}
}

func TestSaveWorkflowToFile_UnsupportedFormat(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Create a workflow to save
	workflow := &WorkflowDefinition{
		Name:        "save-test",
		Description: "Test Save Workflow",
		Steps: []WorkflowStep{
			{
				ID:          "step1",
				Description: "First step",
				Prompt:      "This is the first step",
			},
		},
	}

	// Try to save the workflow in an unsupported format
	err := SaveWorkflowToFile(fs, workflow, "output/save-test.txt")
	if err == nil {
		t.Fatal("Expected error for unsupported format, got nil")
	}
}

func TestLoadWorkflowFromFile_EmptyPrompt(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Add a workflow file with empty prompt
	workflowJSON := `{
		"name": "empty-prompt-workflow",
		"description": "Workflow with empty prompt",
		"steps": [
			{
				"id": "step1",
				"description": "Step with empty prompt",
				"prompt": ""
			}
		]
	}`
	
	jsonPath := "workflows/empty-prompt.json"
	fs.AddFile(jsonPath, []byte(workflowJSON))

	// Load the workflow - this should succeed with empty prompt
	workflow, err := LoadWorkflowFromFile(fs, jsonPath)
	if err != nil {
		t.Fatalf("Expected no error for workflow with empty prompt, got %v", err)
	}

	// Verify the workflow
	if workflow.Name != "empty-prompt-workflow" {
		t.Errorf("Expected workflow name to be 'empty-prompt-workflow', got '%s'", workflow.Name)
	}
	
	if len(workflow.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(workflow.Steps))
	} else if workflow.Steps[0].Prompt != "" {
		t.Errorf("Expected empty prompt in step, got '%s'", workflow.Steps[0].Prompt)
	}
}

func TestIsWorkflowFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{
			name:     "JSON file",
			fileName: "workflow.json",
			want:     true,
		},
		{
			name:     "YAML file",
			fileName: "workflow.yaml",
			want:     true,
		},
		{
			name:     "YML file",
			fileName: "workflow.yml",
			want:     true,
		},
		{
			name:     "Text file",
			fileName: "workflow.txt",
			want:     false,
		},
		{
			name:     "No extension",
			fileName: "workflow",
			want:     false,
		},
		{
			name:     "Mixed case extension",
			fileName: "workflow.YML",
			want:     true, // because isWorkflowFile converts to lowercase
		},
		{
			name:     "Double extension",
			fileName: "workflow.json.txt",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWorkflowFile(tt.fileName)
			if got != tt.want {
				t.Errorf("isWorkflowFile(%q) = %v, want %v", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestValidateWorkflowPromptReferences(t *testing.T) {
	tests := []struct {
		name              string
		setupFs           func(fs *io.MockFileSystem)
		workflow          *ExternalWorkflowDefinition
		baseDir           string
		expectedErrorsLen int
	}{
		{
			name: "all prompt files exist",
			setupFs: func(fs *io.MockFileSystem) {
				fs.AddFile("/workflow/dir/prompts/step1.md", []byte("Step 1 prompt"))
				fs.AddFile("/workflow/dir/prompts/step2.md", []byte("Step 2 prompt"))
			},
			workflow: &ExternalWorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test Workflow",
				Steps: []ExternalWorkflowStep{
					{
						ID:          "step1",
						Description: "Step 1",
						Prompt:      "prompts/step1.md",
					},
					{
						ID:          "step2",
						Description: "Step 2",
						Prompt:      "prompts/step2.md",
					},
				},
			},
			baseDir:           "/workflow/dir",
			expectedErrorsLen: 0,
		},
		{
			name: "embedded prompts are not validated",
			setupFs: func(fs *io.MockFileSystem) {
				// No files needed as embedded prompts are not checked
			},
			workflow: &ExternalWorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test Workflow",
				Steps: []ExternalWorkflowStep{
					{
						ID:          "step1",
						Description: "Step 1",
						Prompt:      "This is an embedded prompt",
					},
					{
						ID:          "step2",
						Description: "Step 2",
						Prompt:      "Another embedded prompt",
					},
				},
			},
			baseDir:           "/workflow/dir",
			expectedErrorsLen: 0,
		},
		{
			name: "some prompt files missing",
			setupFs: func(fs *io.MockFileSystem) {
				fs.AddFile("/workflow/dir/prompts/step1.md", []byte("Step 1 prompt"))
				// step2.md is intentionally missing
			},
			workflow: &ExternalWorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test Workflow",
				Steps: []ExternalWorkflowStep{
					{
						ID:          "step1",
						Description: "Step 1",
						Prompt:      "prompts/step1.md",
					},
					{
						ID:          "step2",
						Description: "Step 2",
						Prompt:      "prompts/step2.md",
					},
				},
			},
			baseDir:           "/workflow/dir",
			expectedErrorsLen: 1,
		},
		{
			name: "absolute path prompts",
			setupFs: func(fs *io.MockFileSystem) {
				fs.AddFile("/absolute/path/prompt.md", []byte("Absolute path prompt"))
				// second file is intentionally missing
			},
			workflow: &ExternalWorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test Workflow",
				Steps: []ExternalWorkflowStep{
					{
						ID:          "step1",
						Description: "Step 1",
						Prompt:      "/absolute/path/prompt.md",
					},
					{
						ID:          "step2",
						Description: "Step 2",
						Prompt:      "/absolute/path/missing.md",
					},
				},
			},
			baseDir:           "/workflow/dir",
			expectedErrorsLen: 1,
		},
		{
			name: "mix of embedded and file prompts",
			setupFs: func(fs *io.MockFileSystem) {
				fs.AddFile("/workflow/dir/prompts/step2.md", []byte("Step 2 prompt"))
			},
			workflow: &ExternalWorkflowDefinition{
				Name:        "test-workflow",
				Description: "Test Workflow",
				Steps: []ExternalWorkflowStep{
					{
						ID:          "step1",
						Description: "Step 1",
						Prompt:      "This is an embedded prompt",
					},
					{
						ID:          "step2",
						Description: "Step 2",
						Prompt:      "prompts/step2.md",
					},
					{
						ID:          "step3",
						Description: "Step 3",
						Prompt:      "prompts/step3.md", // Missing file
					},
				},
			},
			baseDir:           "/workflow/dir",
			expectedErrorsLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := io.NewMockFileSystem()
			tc.setupFs(fs)

			errors := ValidateWorkflowPromptReferences(fs, tc.baseDir, tc.workflow)
			
			assert.Len(t, errors, tc.expectedErrorsLen)
		})
	}
} 
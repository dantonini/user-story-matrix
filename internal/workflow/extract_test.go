// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

// cleanupTestOutput removes any test output directories
func cleanupTestOutput() { //nolint:unused
	os.RemoveAll("output")
}

// TestExtractStandardWorkflow tests extracting the standard workflow to the filesystem
func TestExtractStandardWorkflow(t *testing.T) {
	// Create a mock filesystem
	fs := io.NewMockFileSystem()
	
	// Call the function
	err := ExtractStandardWorkflow(fs, "/test/output")
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify workflow.yaml was created
	if !fs.Exists("/test/output/workflow.yaml") {
		t.Error("workflow.yaml was not created")
	}
	
	// Verify prompts directory was created
	if !fs.Exists("/test/output/prompts") {
		t.Error("prompts directory was not created")
	}
	
	// Verify at least some prompt files were created
	// We don't need to check all of them, just make sure something was written
	files, _ := fs.ReadDir("/test/output/prompts")
	if len(files) == 0 {
		t.Error("No prompt files were created")
	}
}

// TestGenerateWorkflowYAML tests generating a workflow.yaml file
func TestGenerateWorkflowYAML(t *testing.T) {
	// Create a mock filesystem
	fs := io.NewMockFileSystem()
	
	// Create test steps
	steps := []WorkflowStep{
		{
			ID:          "test-step",
			Description: "Test Step",
			Prompt:      "Test prompt",
		},
	}
	
	// Create prompt paths
	promptPaths := map[string]string{
		"test-step": "prompts/test-step.md",
	}
	
	// Call the function
	err := generateWorkflowYAML(fs, steps, "/test/output/workflow.yaml", promptPaths)
	
	// If implemented, test actual output
	if err == nil {
		// Check that file was created
		if !fs.Exists("/test/output/workflow.yaml") {
			t.Error("workflow.yaml file was not created")
		}
		
		// Check file contents
		data, _ := fs.ReadFile("/test/output/workflow.yaml")
		if len(data) == 0 {
			t.Error("workflow.yaml file is empty")
		}
	} else if err.Error() != "not implemented" {
		// If not fully implemented but an error occurred, it should be the expected one
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestExtractPromptToFile tests extracting a prompt to a file
func TestExtractPromptToFile(t *testing.T) {
	// Create a mock filesystem
	fs := io.NewMockFileSystem()
	
	// Create test step
	step := WorkflowStep{
		ID:          "test-step",
		Description: "Test Step",
		Prompt:      "Test prompt",
	}
	
	// Call the function
	promptPath, err := extractPromptToFile(fs, "/test/output/prompts", step)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify returned path
	expectedPath := "prompts/test-step.md"
	if promptPath != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, promptPath)
	}
	
	// Verify file was created
	filePath := "/test/output/prompts/test-step.md"
	if !fs.Exists(filePath) {
		t.Errorf("File was not created at %s", filePath)
	}
	
	// Verify file contents
	content, _ := fs.ReadFile(filePath)
	if string(content) != step.Prompt {
		t.Errorf("Expected content %q, got %q", step.Prompt, string(content))
	}
}

// TestLoadPromptContent tests loading prompt content from a file
func TestLoadPromptContent(t *testing.T) {
	tests := []struct {
		name          string
		step          *WorkflowStep
		setupFs       func(fs *io.MockFileSystem)
		expected      string
		expectedError bool
	}{
		{
			name: "embedded prompt",
			step: &WorkflowStep{
				ID:          "test-step",
				Description: "Test Step",
				Prompt:      "This is an embedded prompt",
				source: promptSource{
					sourceType: promptSourceEmbedded,
				},
			},
			setupFs: func(fs *io.MockFileSystem) {},
			expected: "This is an embedded prompt",
		},
		{
			name: "file prompt exists",
			step: &WorkflowStep{
				ID:          "test-step",
				Description: "Test Step",
				Prompt:      "Original embedded content",
				source: promptSource{
					sourceType: promptSourceFile,
					filePath:   "test-prompt.md",
				},
			},
			setupFs: func(fs *io.MockFileSystem) {
				fs.AddFile("test-prompt.md", []byte("File-based prompt content"))
			},
			expected: "File-based prompt content",
		},
		{
			name: "file prompt doesn't exist - fallback to embedded",
			step: &WorkflowStep{
				ID:          "test-step",
				Description: "Test Step",
				Prompt:      "Original embedded content",
				source: promptSource{
					sourceType: promptSourceFile,
					filePath:   "non-existent-prompt.md",
				},
			},
			setupFs: func(fs *io.MockFileSystem) {
				// Try standard locations for fallback
				fs.AddFile("prompts/non-existent-prompt.md", []byte("Fallback prompt from standard location"))
			},
			expected: "Fallback prompt from standard location",
		},
		{
			name: "path with workflows directory",
			step: &WorkflowStep{
				ID:          "test-step", 
				Description: "Test Step",
				Prompt:      "Original embedded content",
				source: promptSource{
					sourceType: promptSourceFile,
					filePath:   "workflows/custom/prompts/test-prompt.md",
				},
			},
			setupFs: func(fs *io.MockFileSystem) {
				fs.AddFile("workflows/custom/prompts/test-prompt.md", []byte("Custom workflow prompt"))
			},
			expected: "Custom workflow prompt",
		},
		{
			name: "file read error",
			step: &WorkflowStep{
				ID:          "test-step",
				Description: "Test Step",
				Prompt:      "Original embedded content",
				source: promptSource{
					sourceType: promptSourceFile,
					filePath:   "error-prompt.md",
				},
			},
			setupFs: func(fs *io.MockFileSystem) {
				fs.AddFile("error-prompt.md", []byte("Should not read this"))
				fs.SetReadFileError("error-prompt.md", fmt.Errorf("simulated read error"))
			},
			expected:      "Original embedded content",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := io.NewMockFileSystem()
			tc.setupFs(fs)

			content, err := loadPromptContent(tc.step, fs)
			
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			assert.Equal(t, tc.expected, content)
		})
	}
}

// TestFromWorkflowDefinition tests converting a WorkflowDefinition to a WorkflowFileDefinition
func TestFromWorkflowDefinition(t *testing.T) {
	// Create test workflow
	workflow := &WorkflowDefinition{
		Name:        "test",
		Description: "Test workflow",
		Steps: []WorkflowStep{
			{
				ID:          "test-step",
				Description: "Test Step",
				Prompt:      "Test prompt",
			},
		},
	}
	
	// Create prompt paths
	promptPaths := map[string]string{
		"test-step": "prompts/test-step.md",
	}
	
	// Call the function
	result := FromWorkflowDefinition(workflow, promptPaths)
	
	// Verify result
	if result.Name != workflow.Name {
		t.Errorf("Expected name %q, got %q", workflow.Name, result.Name)
	}
	
	if result.Description != workflow.Description {
		t.Errorf("Expected description %q, got %q", workflow.Description, result.Description)
	}
	
	// TODO: Once fully implemented, test step conversion
	// 1. Verify steps are converted correctly
	// 2. Verify prompt paths are used correctly
}

// TestGetRelativePromptPath tests getting the relative path from a workflow directory to a prompt file
func TestGetRelativePromptPath(t *testing.T) {
	// Test cases
	tests := []struct{
		name string
		promptFile string
		workflowDir string
		expected string
	}{
		{
			name: "Simple relative path",
			promptFile: "/test/output/prompts/test.md",
			workflowDir: "/test/output",
			expected: "prompts/test.md",
		},
		{
			name: "Already relative path",
			promptFile: "prompts/test.md",
			workflowDir: "/test/output",
			expected: "prompts/test.md",
		},
		{
			name: "Nested directories",
			promptFile: "/test/output/prompts/subdir/test.md",
			workflowDir: "/test/output",
			expected: "prompts/subdir/test.md",
		},
		{
			name: "Different drive or root",
			promptFile: "/other/path/prompts/test.md",
			workflowDir: "/test/output",
			expected: "../../other/path/prompts/test.md",
		},
	}
	
	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getRelativePromptPath(tc.promptFile, tc.workflowDir)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// TestResolvePromptPath tests resolving a prompt path
func TestResolvePromptPath(t *testing.T) {
	tests := []struct {
		name        string
		promptPath  string
		workflowDir string
		expected    string
	}{
		{
			name:        "absolute path",
			promptPath:  "/absolute/path/to/prompt.md",
			workflowDir: "/workflow/dir",
			expected:    "/absolute/path/to/prompt.md",
		},
		{
			name:        "relative path",
			promptPath:  "prompts/step1.md",
			workflowDir: "/workflow/dir",
			expected:    "/workflow/dir/prompts/step1.md",
		},
		{
			name:        "special prompts prefix",
			promptPath:  "prompts/special.md",
			workflowDir: "/custom/workflows",
			expected:    "/custom/workflows/prompts/special.md",
		},
		{
			name:        "normalize path separators",
			promptPath:  filepath.FromSlash("path/with/backslashes.md"),
			workflowDir: "/workflow/dir",
			expected:    filepath.FromSlash("/workflow/dir/path/with/backslashes.md"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolvePromptPath(tc.promptPath, tc.workflowDir)
			// Convert both expected and result to platform-specific format before comparison
			assert.Equal(t, filepath.Clean(tc.expected), filepath.Clean(result))
		})
	}
}

// TestIntegrationSmoke_ExtractAndLoad is a regression smoke test to verify
// that extraction and loading work together properly
func TestIntegrationSmoke_ExtractAndLoad(t *testing.T) {
	// Create a mock filesystem for testing
	fs := io.NewMockFileSystem()
	
	// Test output directory
	outputDir := "/test/workflow-extract"
	
	// 1. Extract standard workflow to outputDir
	err := ExtractStandardWorkflow(fs, outputDir)
	if err != nil {
		t.Fatalf("Failed to extract workflow: %v", err)
	}
	
	// 2. Load the workflow from the same directory
	registry := NewWorkflowRegistry()
	loadedWorkflow, err := registry.LoadFromDirectory(fs, outputDir)
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}
	
	// 3. Verify the loaded workflow matches the original
	if loadedWorkflow.Name != "standard" {
		t.Errorf("Expected name 'standard', got '%s'", loadedWorkflow.Name)
	}
	
	// Check that we have the same number of steps
	if len(loadedWorkflow.Steps) != len(UsmCodeStandardWorkflowSteps) {
		t.Errorf("Expected %d steps, got %d", len(UsmCodeStandardWorkflowSteps), len(loadedWorkflow.Steps))
	}
	
	// Check that step IDs match
	for i, step := range loadedWorkflow.Steps {
		if i < len(UsmCodeStandardWorkflowSteps) && step.ID != UsmCodeStandardWorkflowSteps[i].ID {
			t.Errorf("Step %d: expected ID '%s', got '%s'", i, UsmCodeStandardWorkflowSteps[i].ID, step.ID)
		}
	}
	
	// Check that step descriptions match
	for i, step := range loadedWorkflow.Steps {
		if i < len(UsmCodeStandardWorkflowSteps) && step.Description != UsmCodeStandardWorkflowSteps[i].Description {
			t.Errorf("Step %d: expected description '%s', got '%s'", 
				i, UsmCodeStandardWorkflowSteps[i].Description, step.Description)
		}
	}
	
	// Check that step prompts match
	for i, step := range loadedWorkflow.Steps {
		if i < len(UsmCodeStandardWorkflowSteps) && step.Prompt != UsmCodeStandardWorkflowSteps[i].Prompt {
			t.Errorf("Step %d: prompt content doesn't match", i)
		}
	}
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Standard workflow template paths
const (
	// StandardTemplateDir is the directory where the standard workflow template is stored
	StandardTemplateDir = "internal/workflow/templates/standard"
	
	// StandardPromptsDir is the directory where standard workflow prompt files are stored
	StandardPromptsDir = "internal/workflow/templates/standard/prompts"
	
	// StandardWorkflowYAML is the filename for the standard workflow definition
	StandardWorkflowYAML = "workflow.yaml"
)

// promptSourceType identifies where a prompt comes from
type promptSourceType int //nolint:unused

const (
	promptSourceEmbedded promptSourceType = iota //nolint:unused
	promptSourceFile //nolint:unused
)

// promptSource tracks the origin of a prompt
type promptSource struct { //nolint:unused
	sourceType promptSourceType
	filePath   string // Only used when sourceType is promptSourceFile
}

// ExtractStandardWorkflow exports the standard workflow to the specified directory
// This creates a workflow.yaml file and individual prompt files for each step.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - outputDir: Target directory to export the workflow to
//
// Returns:
//   - error if the extraction fails
func ExtractStandardWorkflow(fs FileSystem, outputDir string) error {
	// Create output directory if it doesn't exist
	if !fs.Exists(outputDir) {
		if err := fs.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	
	// Create prompts directory if it doesn't exist
	promptsDir := filepath.Join(outputDir, "prompts")
	if !fs.Exists(promptsDir) {
		if err := fs.MkdirAll(promptsDir, 0755); err != nil {
			return fmt.Errorf("failed to create prompts directory: %w", err)
		}
	}
	
	// Extract each prompt to a file and collect paths
	promptPaths := make(map[string]string)
	for _, step := range StandardWorkflowSteps {
		promptPath, err := extractPromptToFile(fs, promptsDir, step)
		if err != nil {
			return fmt.Errorf("failed to extract prompt for step %s: %w", step.ID, err)
		}
		promptPaths[step.ID] = promptPath
	}
	
	// Generate workflow.yaml file
	workflowYAMLPath := filepath.Join(outputDir, StandardWorkflowYAML)
	if err := generateWorkflowYAML(fs, StandardWorkflowSteps, workflowYAMLPath, promptPaths); err != nil {
		return fmt.Errorf("failed to generate workflow.yaml: %w", err)
	}
	
	return nil
}

// generateWorkflowYAML creates a workflow.yaml file from StandardWorkflowSteps
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - steps: Workflow steps to generate YAML from
//   - outputPath: Path to save the workflow.yaml file
//   - promptPaths: Map of step ID to relative prompt file path
//
// Returns:
//   - error if the YAML generation or file writing fails
func generateWorkflowYAML(fs FileSystem, steps []WorkflowStep, outputPath string, promptPaths map[string]string) error {
	// Create WorkflowFileDefinition from steps
	fileDef := FromWorkflowDefinition(&WorkflowDefinition{
		Name:        "standard",
		Description: "Standard development workflow",
		Steps:       steps,
	}, promptPaths)
	
	// Marshal to YAML
	data, err := yaml.Marshal(fileDef)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow to YAML: %w", err)
	}
	
	// Write to file
	if err := fs.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflow YAML: %w", err)
	}
	
	return nil
}

// extractPromptToFile writes a single prompt to a markdown file
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - promptsDir: Directory to save prompt files in
//   - step: The workflow step containing the prompt to extract
//
// Returns:
//   - string containing the relative path to the prompt file
//   - error if writing the prompt file fails
func extractPromptToFile(fs FileSystem, promptsDir string, step WorkflowStep) (string, error) {
	// Generate filename from step ID
	filename := fmt.Sprintf("%s.md", step.ID)
	filePath := filepath.Join(promptsDir, filename)
	
	// Write prompt content to file
	if err := fs.WriteFile(filePath, []byte(step.Prompt), 0644); err != nil {
		return "", fmt.Errorf("failed to write prompt file: %w", err)
	}
	
	// Return relative path to the file
	return filepath.Join("prompts", filename), nil
}

// loadPromptContent loads prompt content, with priority to file sources
// If the prompt is from a file, it reads the file content
// If the file doesn't exist or there's an error, it falls back to embedded prompt
//
// Parameters:
//   - step: The workflow step
//   - fs: FileSystem interface for file operations
//
// Returns:
//   - The prompt content as a string
//   - error if loading fails
func loadPromptContent(step *WorkflowStep, fs FileSystem) (string, error) {
	// Check if step has a file source
	if step.source.sourceType == promptSourceFile && step.source.filePath != "" {
		// Try to read from file if available
		if fs.Exists(step.source.filePath) {
			data, err := fs.ReadFile(step.source.filePath)
			if err == nil {
				return string(data), nil
			}
			// Log the error but don't return it, fall back to embedded prompt
			fmt.Printf("Failed to load prompt from file %s: %v, falling back to embedded prompt\n", 
				step.source.filePath, err)
		}
	}
	
	// Fall back to embedded prompt
	return step.Prompt, nil
}

// setPromptFromFile sets the prompt source to a file path
//
// Parameters:
//   - step: The workflow step to update
//   - path: File path to the prompt
func setPromptFromFile(step *WorkflowStep, path string) { //nolint:unused
	step.source = promptSource{
		sourceType: promptSourceFile,
		filePath:   path,
	}
}

// getRelativePromptPath returns the relative path to a prompt file from the workflow directory
//
// Parameters:
//   - promptFile: Absolute path to prompt file
//   - workflowDir: Absolute path to workflow directory
//
// Returns:
//   - Relative path from workflow directory to prompt file
func getRelativePromptPath(promptFile, workflowDir string) string {
	// Clean paths to ensure consistent format
	promptFile = filepath.Clean(promptFile)
	workflowDir = filepath.Clean(workflowDir)
	
	// Check if prompt file is already relative
	if !filepath.IsAbs(promptFile) {
		return promptFile
	}
	
	// Calculate relative path
	rel, err := filepath.Rel(workflowDir, promptFile)
	if err != nil {
		// If we can't create a relative path, return the original path
		return promptFile
	}
	
	// Ensure forward slashes for cross-platform compatibility in YAML
	return strings.ReplaceAll(rel, "\\", "/")
}

// WorkflowFileDefinition represents the structure of a workflow.yaml file
type WorkflowFileDefinition struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Steps       []WorkflowFileStep `yaml:"steps"`
}

// WorkflowFileStep represents a step in a workflow.yaml file
type WorkflowFileStep struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Prompt      string `yaml:"prompt"` // Path to prompt file, relative to workflow dir
}

// FromWorkflowDefinition creates a WorkflowFileDefinition from a WorkflowDefinition
//
// Parameters:
//   - def: Source WorkflowDefinition
//   - promptPaths: Map of step ID to prompt file relative path
//
// Returns:
//   - WorkflowFileDefinition suitable for serialization to YAML
func FromWorkflowDefinition(def *WorkflowDefinition, promptPaths map[string]string) WorkflowFileDefinition {
	fileSteps := make([]WorkflowFileStep, len(def.Steps))
	
	for i, step := range def.Steps {
		promptPath := "step-prompt.md" // Default value
		if path, exists := promptPaths[step.ID]; exists {
			promptPath = path
		}
		
		fileSteps[i] = WorkflowFileStep{
			ID:          step.ID,
			Description: step.Description,
			Prompt:      promptPath,
		}
	}
	
	return WorkflowFileDefinition{
		Name:        def.Name,
		Description: def.Description,
		Steps:       fileSteps,
	}
} 
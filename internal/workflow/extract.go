// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
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
	// TODO: Implement extraction of standard workflow
	// 1. Create directory structure if it doesn't exist
	// 2. Extract each prompt to a file
	// 3. Create workflow.yaml referencing the prompt files
	
	return fmt.Errorf("not implemented")
}

// generateWorkflowYAML creates a workflow.yaml file from StandardWorkflowSteps
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - steps: Workflow steps to generate YAML from
//   - outputPath: Path to save the workflow.yaml file
//
// Returns:
//   - error if the YAML generation or file writing fails
func generateWorkflowYAML(fs FileSystem, steps []WorkflowStep, outputPath string) error {
	// TODO: Implement workflow.yaml generation
	// 1. Create WorkflowFileDefinition from steps
	// 2. Update prompt references to relative file paths
	// 3. Marshal to YAML
	// 4. Write to file
	
	return fmt.Errorf("not implemented")
}

// extractPromptToFile writes a single prompt to a markdown file
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - promptsDir: Directory to save prompt files in
//   - step: The workflow step containing the prompt to extract
//
// Returns:
//   - error if writing the prompt file fails
//   - string containing the relative path to the prompt file
func extractPromptToFile(fs FileSystem, promptsDir string, step WorkflowStep) (string, error) {
	// TODO: Implement prompt extraction
	// 1. Generate filename from step ID
	// 2. Write prompt content to file
	// 3. Return relative path to the file
	
	return "", fmt.Errorf("not implemented")
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
	// TODO: Implement prompt loading
	// 1. Check if step has a file source
	// 2. Try to read from file if available
	// 3. Fall back to embedded prompt if file read fails
	// 4. Return prompt content
	
	return step.Prompt, nil // Default to embedded prompt for now
}

// setPromptFromFile sets the prompt source to a file path
//
// Parameters:
//   - step: The workflow step to update
//   - path: File path to the prompt
func setPromptFromFile(step *WorkflowStep, path string) { //nolint:unused
	// TODO: Implement setting prompt source
	// 1. Update step's internal source field
	// 2. Keep original prompt as fallback
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
	// TODO: Implement relative path calculation
	// 1. Convert absolute paths to clean format
	// 2. Calculate relative path from workflow dir to prompt file
	
	return ""
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
	// TODO: Implement conversion
	// 1. Create file definition with same name and description
	// 2. Convert each step using prompt paths
	
	return WorkflowFileDefinition{
		Name:        def.Name,
		Description: def.Description,
	}
} 
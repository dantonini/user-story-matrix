// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/user-story-matrix/usm/internal/io"
	"gopkg.in/yaml.v3"
)

// ExternalWorkflowDefinition represents a workflow definition loaded from an external file.
// It contains the same fields as WorkflowDefinition, but structured for easy serialization.
type ExternalWorkflowDefinition struct {
	// Name uniquely identifies the workflow (e.g., "standard", "custom-tutorial")
	Name string `json:"name" yaml:"name"`

	// Description provides a human-readable explanation of the workflow's purpose
	Description string `json:"description" yaml:"description"`

	// Steps contains the ordered sequence of steps that make up this workflow
	Steps []ExternalWorkflowStep `json:"steps" yaml:"steps"`
}

// ExternalWorkflowStep represents a workflow step loaded from an external file.
// It contains the same fields as WorkflowStep, but structured for easy serialization.
type ExternalWorkflowStep struct {
	// ID uniquely identifies the step (e.g., "01-laying-the-foundation")
	ID string `json:"id" yaml:"id"`

	// Description provides a human-readable explanation of what the step does
	Description string `json:"description" yaml:"description"`

	// Prompt contains the instructions for the AI agent
	Prompt string `json:"prompt" yaml:"prompt"`
	
	// Variables for template substitution in the prompt
	Variables map[string]string `json:"variables" yaml:"variables"`
}

// ToWorkflowDefinition converts an ExternalWorkflowDefinition to a WorkflowDefinition.
// This allows the same workflow format to be used regardless of whether it was
// defined in code or loaded from an external file.
//
// Returns:
//   - A WorkflowDefinition created from this ExternalWorkflowDefinition
func (e *ExternalWorkflowDefinition) ToWorkflowDefinition() *WorkflowDefinition {
	steps := make([]WorkflowStep, len(e.Steps))
	for i, step := range e.Steps {
		workflowStep := WorkflowStep{
			ID:          step.ID,
			Description: step.Description,
			Prompt:      step.Prompt,
			Variables:   step.Variables,
		}

		// Initialize the source field for all prompts that look like file paths
		// This is more inclusive than the previous check to catch all potential file references
		if strings.Contains(step.Prompt, "/") || 
		   strings.Contains(step.Prompt, "\\") || 
		   filepath.Ext(step.Prompt) != "" {
			// This is likely a file reference - create a file source
			workflowStep.source = promptSource{
				sourceType: promptSourceFile,
				filePath:   step.Prompt, // For now, store the relative path; it will be resolved later
			}
		} else if step.Prompt != "" {
			// For non-empty, non-file prompts, set as embedded type
			workflowStep.source = promptSource{
				sourceType: promptSourceEmbedded,
			}
		}
		// For empty prompts, the source defaults to zero values which is fine
		
		steps[i] = workflowStep
	}
	
	return &WorkflowDefinition{
		Name:        e.Name,
		Description: e.Description,
		Steps:       steps,
	}
}

// LoadWorkflowsFromDirectory loads all workflow definition files from the specified directory
// and registers them with the provided registry. It supports both YAML and JSON formats.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - directory: Path to the directory containing workflow definitions
//   - registry: WorkflowRegistry to register the loaded workflows with
//
// Returns:
//   - A slice of loaded workflow names, or an error if loading failed
func LoadWorkflowsFromDirectory(fs io.FileSystem, directory string, registry *WorkflowRegistry) ([]*WorkflowDefinition, error) {
	// Check if directory exists
	if !fs.Exists(directory) {
		return nil, fmt.Errorf("directory not found: %s", directory)
	}
	
	// Get list of files in directory
	files, err := fs.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	
	// Load workflows from files
	workflows := make([]*WorkflowDefinition, 0)
	for _, file := range files {
		// Skip directories and non-workflow files
		if file.IsDir() {
			continue
		}
		
		fileName := file.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			continue
		}
		
		// Load workflow from file
		filePath := filepath.Join(directory, fileName)
		workflow, err := LoadWorkflowFromFile(fs, filePath)
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("Error loading workflow from %s: %v\n", filePath, err)
			continue
		}
		
		// Register workflow with registry
		if registry != nil {
			registry.RegisterBuiltInWorkflow(workflow)
		}
		
		workflows = append(workflows, workflow)
	}
	
	return workflows, nil
}

// LoadWorkflowFromFile loads a workflow definition from a file.
// It supports both YAML and JSON formats.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - filePath: Path to the workflow definition file
//
// Returns:
//   - The loaded WorkflowDefinition, or an error if loading failed
func LoadWorkflowFromFile(fs io.FileSystem, filePath string) (*WorkflowDefinition, error) {
	// Check if file exists
	if !fs.Exists(filePath) {
		return nil, fmt.Errorf("workflow file not found: %s", filePath)
	}
	
	// Read the file
	data, err := fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}
	
	// Create an external workflow definition
	var externalWorkflow ExternalWorkflowDefinition
	
	// Determine the file format based on extension
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &externalWorkflow); err != nil {
			return nil, fmt.Errorf("invalid YAML in workflow file: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &externalWorkflow); err != nil {
			return nil, fmt.Errorf("invalid JSON in workflow file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported workflow file format: %s", ext)
	}
	
	// Validate the workflow
	if err := validateExternalWorkflow(&externalWorkflow); err != nil {
		return nil, err
	}
	
	// Convert to WorkflowDefinition
	return externalWorkflow.ToWorkflowDefinition(), nil
}

// SaveWorkflowToFile saves a workflow definition to a file.
// It supports both YAML and JSON formats, determined by the file extension.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - workflow: WorkflowDefinition to save
//   - filePath: Path to save the workflow to
//
// Returns:
//   - An error if saving failed
func SaveWorkflowToFile(fs io.FileSystem, workflow *WorkflowDefinition, filePath string) error {
	// Convert to external workflow format
	externalWorkflow := ExternalWorkflowDefinition{
		Name:        workflow.Name,
		Description: workflow.Description,
		Steps:       make([]ExternalWorkflowStep, len(workflow.Steps)),
	}
	
	for i, step := range workflow.Steps {
		externalWorkflow.Steps[i] = ExternalWorkflowStep{
			ID:          step.ID,
			Description: step.Description,
			Prompt:      step.Prompt,
			Variables:   step.Variables,
		}
	}
	
	// Determine the file format based on extension
	ext := strings.ToLower(filepath.Ext(filePath))
	var data []byte
	var err error
	
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(externalWorkflow)
		if err != nil {
			return fmt.Errorf("failed to marshal workflow to YAML: %w", err)
		}
	case ".json":
		data, err = json.MarshalIndent(externalWorkflow, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal workflow to JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported workflow file format: %s", ext)
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if dir != "." && !fs.Exists(dir) {
		err = fs.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Write the file
	if err := fs.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}
	
	return nil
}

// validateExternalWorkflow validates an external workflow definition.
//
// Parameters:
//   - workflow: ExternalWorkflowDefinition to validate
//
// Returns:
//   - An error if validation failed
func validateExternalWorkflow(workflow *ExternalWorkflowDefinition) error {
	// Ensure the workflow has a name
	if workflow.Name == "" {
		return fmt.Errorf("workflow must have a name")
	}

	// Sanitize workflow name and description
	workflow.Name = sanitizeWorkflowField(workflow.Name)
	workflow.Description = sanitizeWorkflowField(workflow.Description)
	
	// Ensure the workflow has at least one step
	if len(workflow.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}
	
	// Ensure all steps have IDs and prompts
	for i, step := range workflow.Steps {
		if step.ID == "" {
			return fmt.Errorf("step %d must have an ID", i)
		}
		
		// Sanitize step fields
		workflow.Steps[i].ID = sanitizeWorkflowField(step.ID)
		workflow.Steps[i].Description = sanitizeWorkflowField(step.Description)
		
		// Prompt can be empty - will be preserved as is
	}
	
	return nil
}

// sanitizeWorkflowField ensures workflow fields don't contain newlines
// or other problematic characters that could break display
func sanitizeWorkflowField(s string) string {
	// Replace newlines with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	
	// Replace tabs with spaces
	s = strings.ReplaceAll(s, "\t", " ")
	
	// Replace carriage returns with spaces
	s = strings.ReplaceAll(s, "\r", " ")
	
	// Normalize multiple spaces to a single space
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	
	return strings.TrimSpace(s)
}

// ValidateWorkflowPromptReferences checks that all prompt files referenced in a workflow exist
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - baseDir: Base directory for resolving relative path references
//   - workflow: ExternalWorkflowDefinition to validate
//
// Returns:
//   - A slice of errors found during validation, empty if all valid
func ValidateWorkflowPromptReferences(fs io.FileSystem, baseDir string, workflow *ExternalWorkflowDefinition) []error {
	var errors []error
	
	for _, step := range workflow.Steps {
		// Skip steps with embedded prompts (no file extension or path separator)
		if !strings.Contains(step.Prompt, "/") && filepath.Ext(step.Prompt) == "" {
			continue
		}
		
		// Resolve prompt path
		promptPath := step.Prompt
		if !filepath.IsAbs(promptPath) {
			promptPath = filepath.Join(baseDir, promptPath)
		}
		
		// Check if prompt file exists
		if !fs.Exists(promptPath) {
			errors = append(errors, fmt.Errorf("prompt file for step %s not found: %s", step.ID, promptPath))
		}
	}
	
	return errors
}

// isWorkflowFile checks if a filename has a workflow file extension
func isWorkflowFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	return ext == ".yaml" || ext == ".yml" || ext == ".json"
} 
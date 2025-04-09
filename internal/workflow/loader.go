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
		steps[i] = WorkflowStep(step)
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
		return nil, fmt.Errorf("workflow directory not found: %s", directory)
	}
	
	// List files in the directory
	entries, err := fs.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow directory: %w", err)
	}
	
	// Pre-allocate the workflows slice
	workflows := make([]*WorkflowDefinition, 0, len(entries))
	
	// Process each file
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}
		
		fileName := entry.Name()
		
		// Only process workflow definition files
		if !isWorkflowFile(fileName) {
			continue
		}
		
		// Load the workflow
		filePath := filepath.Join(directory, fileName)
		workflow, err := LoadWorkflowFromFile(fs, filePath)
		if err != nil {
			// Log the error but continue processing other files
			fmt.Printf("Warning: Failed to load workflow from %s: %v\n", filePath, err)
			continue
		}
		
		// Register the workflow
		registry.RegisterBuiltInWorkflow(workflow)
		workflows = append(workflows, workflow)
	}
	
	return workflows, nil
}

// LoadWorkflowFromFile loads a workflow definition from a file.
// It supports both YAML and JSON formats, detected based on the file extension.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - filePath: Path to the workflow definition file
//
// Returns:
//   - A WorkflowDefinition loaded from the file, or an error if loading failed
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
		externalWorkflow.Steps[i] = ExternalWorkflowStep(step)
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
	// Check required fields
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	
	if workflow.Description == "" {
		return fmt.Errorf("workflow description is required")
	}
	
	if len(workflow.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}
	
	// Validate each step
	for i, step := range workflow.Steps {
		if step.ID == "" {
			return fmt.Errorf("step %d is missing an ID", i+1)
		}
		
		if step.Description == "" {
			return fmt.Errorf("step %d (%s) is missing a description", i+1, step.ID)
		}
		
		// Validate the prompt for each step
		// Only validate the prompt if it's not empty
		if step.Prompt != "" {
			if err := ValidatePrompt(step.Prompt); err != nil {
				return fmt.Errorf("step %d (%s) has an invalid prompt: %v", i+1, step.ID, err)
			}
		}
	}
	
	return nil
}

// isWorkflowFile checks if a file is a workflow definition file based on its extension.
//
// Parameters:
//   - fileName: Name of the file to check
//
// Returns:
//   - true if the file is a workflow definition file, false otherwise
func isWorkflowFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	return ext == ".yaml" || ext == ".yml" || ext == ".json"
} 
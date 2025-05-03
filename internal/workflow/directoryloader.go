// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/user-story-matrix/usm/internal/io"
	"gopkg.in/yaml.v3"
)

// TODO: Extend phase improvements for directoryloader.go:
// - Add support for shared template fragments in the shared/ directory
// - Implement recursive directory traversal for nested directories
// - Add support for importing templates from other workflows
// - Implement workflow versioning support
// - Add cache invalidation for modified prompt files

// DirectoryWorkflowInfo represents metadata about a workflow stored in a directory
type DirectoryWorkflowInfo struct {
	// Path is the file system path to the workflow directory
	Path string
	
	// Source indicates where the workflow came from (built-in, user, project)
	Source string
}

// Constants for workflow sources
const (
	SourceBuiltIn = "built-in"
	SourceUser    = "user"
	SourceProject = "project"
)

// Constants for directory structure
const (
	WorkflowConfigFile = "workflow.yaml"
	PromptsDir         = "prompts"
	SharedDir          = "shared"
)

// LoadWorkflowFromDirectory loads a workflow definition from a directory structure.
// The directory must contain a workflow.yaml file and a prompts/ subdirectory.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - dirPath: Path to the workflow directory
//
// Returns:
//   - The loaded WorkflowDefinition and its info, or an error if loading failed
func LoadWorkflowFromDirectory(fs io.FileSystem, dirPath string) (*WorkflowDefinition, *DirectoryWorkflowInfo, error) {
	// Check if directory exists
	if !fs.Exists(dirPath) {
		return nil, nil, fmt.Errorf("workflow directory not found: %s", dirPath)
	}
	
	// Check if workflow.yaml exists
	workflowYAMLPath := filepath.Join(dirPath, WorkflowConfigFile)
	if !fs.Exists(workflowYAMLPath) {
		return nil, nil, fmt.Errorf("workflow configuration file not found: %s", workflowYAMLPath)
	}
	
	// Check if prompts directory exists
	promptsDirPath := filepath.Join(dirPath, PromptsDir)
	if !fs.Exists(promptsDirPath) {
		return nil, nil, fmt.Errorf("prompts directory not found: %s", promptsDirPath)
	}
	
	// Read the workflow.yaml file
	data, err := fs.ReadFile(workflowYAMLPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read workflow configuration: %w", err)
	}
	
	// Parse the workflow.yaml file
	var externalWorkflow ExternalWorkflowDefinition
	if err := yaml.Unmarshal(data, &externalWorkflow); err != nil {
		return nil, nil, fmt.Errorf("invalid YAML in workflow configuration: %w", err)
	}
	
	// Determine the workflow source
	source := determineWorkflowSource(dirPath)
	
	// Validate the workflow
	errors := ValidateWorkflowPromptReferences(fs, dirPath, &externalWorkflow)
	if len(errors) > 0 {
		// Combine all errors into a single error message
		errorMsgs := make([]string, len(errors))
		for i, err := range errors {
			errorMsgs[i] = err.Error()
		}
		return nil, nil, fmt.Errorf("workflow validation failed:\n- %s", strings.Join(errorMsgs, "\n- "))
	}
	
	// Create workflow info
	info := &DirectoryWorkflowInfo{
		Path:   dirPath,
		Source: source,
	}
	
	// Convert to internal format
	workflowDef := externalWorkflow.ToWorkflowDefinition()
	
	// Update the prompt file paths to be absolute paths
	// This is important for correctly resolving prompt files, especially in custom workflows
	for i, step := range workflowDef.Steps {
		if step.source.sourceType == promptSourceFile {
			// Get the original prompt path from the external workflow definition
			promptPath := externalWorkflow.Steps[i].Prompt
			
			// If the path isn't absolute, make it absolute relative to the workflow directory
			var absolutePath string
			if filepath.IsAbs(promptPath) {
				absolutePath = promptPath
			} else {
				// First check if the path already contains the "prompts/" directory
				if strings.HasPrefix(promptPath, PromptsDir+"/") || strings.HasPrefix(promptPath, PromptsDir+"\\") {
					// Path already includes prompts/ prefix, just join with dirPath
					absolutePath = filepath.Join(dirPath, promptPath)
				} else {
					// First try joining with dirPath directly
					absolutePath = filepath.Join(dirPath, promptPath)
					
					// If that file doesn't exist, try looking in the prompts directory
					if !fs.Exists(absolutePath) {
						// Try with the file in the prompts directory
						absolutePromptPath := filepath.Join(dirPath, PromptsDir, filepath.Base(promptPath))
						if fs.Exists(absolutePromptPath) {
							absolutePath = absolutePromptPath
						}
					}
				}
			}
			
			workflowDef.Steps[i].source.filePath = absolutePath
		}
	}
	
	return workflowDef, info, nil
}

// ValidateDirectoryWorkflow validates a workflow directory structure.
// It checks that the directory contains a valid workflow.yaml file and 
// that all referenced prompt files exist.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - dirPath: Path to the workflow directory
//
// Returns:
//   - A slice of errors found during validation, empty if valid
func ValidateDirectoryWorkflow(fs io.FileSystem, dirPath string) []error {
	errors := make([]error, 0)
	
	// Check if directory exists
	if !fs.Exists(dirPath) {
		errors = append(errors, fmt.Errorf("workflow directory not found: %s", dirPath))
		return errors
	}
	
	// Check if workflow.yaml exists
	workflowYAMLPath := filepath.Join(dirPath, WorkflowConfigFile)
	if !fs.Exists(workflowYAMLPath) {
		errors = append(errors, fmt.Errorf("workflow configuration file not found: %s", workflowYAMLPath))
		return errors
	}
	
	// Read the workflow.yaml file
	data, err := fs.ReadFile(workflowYAMLPath)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to read workflow configuration: %w", err))
		return errors
	}
	
	// Parse the workflow.yaml file
	var externalWorkflow ExternalWorkflowDefinition
	if err := yaml.Unmarshal(data, &externalWorkflow); err != nil {
		errors = append(errors, fmt.Errorf("invalid YAML in workflow configuration: %w", err))
		return errors
	}
	
	// Validate basic workflow structure
	if err := validateExternalWorkflow(&externalWorkflow); err != nil {
		errors = append(errors, err)
	}
	
	// Validate prompt references
	promptErrors := ValidateWorkflowPromptReferences(fs, dirPath, &externalWorkflow)
	errors = append(errors, promptErrors...)
	
	// Check if prompts directory exists
	promptsDirPath := filepath.Join(dirPath, PromptsDir)
	if !fs.Exists(promptsDirPath) {
		errors = append(errors, fmt.Errorf("prompts directory not found: %s", promptsDirPath))
	}
	
	return errors
}

// determineWorkflowSource determines the source of a workflow based on its path.
// Source can be built-in, user, or project.
func determineWorkflowSource(dirPath string) string {
	// TODO: Implement actual source determination based on path
	// For now, we'll just return project as a placeholder
	return SourceProject
} 
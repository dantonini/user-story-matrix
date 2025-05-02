// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// ValidationResult represents the result of a workflow validation operation
type ValidationResult struct {
	Success      bool
	WorkflowName string
	Errors       []string
	Warnings     []string
	ErrorMessage string
}

// ValidateCmd represents the workflow validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate [name-or-path]",
	Short: "Validate a workflow definition",
	Long: `Validate a workflow definition by name or path.

This command validates that:
- The workflow.yaml file is properly formatted
- Required fields are present and valid
- Step IDs are unique
- Referenced prompt files exist
- Template syntax in prompt files is valid

You can provide either:
- A workflow name (to validate a registered workflow)
- A path to a workflow directory

Examples:
  # Validate a workflow by name
  usm workflow validate my-workflow

  # Validate a workflow by path
  usm workflow validate path/to/workflow
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nameOrPath := args[0]
		
		output := io.NewTerminalIO()
		fs := io.NewOSFileSystem()
		
		// Call the business logic function
		result := validateWorkflow(nameOrPath, fs, output)
		
		// Handle the result
		if !result.Success {
			output.PrintError(result.ErrorMessage)
			os.Exit(1)
			return
		}
		
		// Display successful validation results
		output.PrintProgress(fmt.Sprintf("Validating workflow '%s'", result.WorkflowName))
		output.PrintSuccess("Workflow is valid")
		
		// Display any warnings
		if len(result.Warnings) > 0 {
			output.PrintWarning(fmt.Sprintf("Found %d warnings:", len(result.Warnings)))
			for _, warning := range result.Warnings {
				output.Print(fmt.Sprintf("- %s", warning))
			}
		}
	},
}

// validateWorkflow implements the business logic for workflow validation
// It extracts the validation logic from the command's Run function to make it testable
//
// Parameters:
//   - nameOrPath: Either a workflow name or path to a workflow directory
//   - fs: The filesystem interface for file operations
//   - output: The output interface for user feedback (used only for logging, not errors)
//
// Returns:
//   - ValidationResult containing success status and validation details
func validateWorkflow(nameOrPath string, fs io.FileSystem, output io.UserOutput) ValidationResult {
	// Get the global registry
	registry := workflow.GetGlobalRegistry()
	
	// Discover available workflows
	// This is required to find workflows that were created by the user
	// but might not have been loaded into the registry yet
	registry.DiscoverWorkflows(fs)
	
	// Determine if input is a path or a name
	var workflowPath string
	var workflowDef *workflow.WorkflowDefinition
	
	if fs.Exists(nameOrPath) && isValidWorkflowDir(fs, nameOrPath) {
		// Input is a path
		workflowPath = nameOrPath
		
		// Try to load the workflow from this path
		workflowYAMLPath := filepath.Join(workflowPath, workflow.WorkflowConfigFile)
		var err error
		workflowDef, err = workflow.LoadWorkflowFromFile(fs, workflowYAMLPath)
		if err != nil {
			return ValidationResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("Failed to load workflow: %s", err.Error()),
			}
		}
	} else {
		// Assume input is a name, try to find it in registry
		var err error
		workflowDef, err = registry.GetWorkflow(nameOrPath)
		if err != nil {
			return ValidationResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("Workflow '%s' not found. Use 'usm workflow list' to see available workflows.", nameOrPath),
			}
		}
		
		// First check for source path in the registry's cache
		source, path := registry.GetWorkflowSourceInfo(nameOrPath)
		if path != "-" && path != "" {
			workflowPath = path
		} else {
			// If no source path available, try to find it by name
			workflowPath = findWorkflowByName(fs, nameOrPath)
			if workflowPath == "" {
				// For built-in workflows, we don't need a path for validation
				if source == workflow.SourceBuiltIn {
					// Built-in workflows don't need prompt validation
					return ValidationResult{
						Success:      true,
						WorkflowName: workflowDef.Name,
						Warnings:     []string{"Built-in workflow validated without prompt reference checks"},
					}
				}
				
				// For other workflow types, we need a path to validate prompt references
				return ValidationResult{
					Success:      false,
					ErrorMessage: fmt.Sprintf("Cannot find workflow directory for '%s'", nameOrPath),
				}
			}
		}
	}
	
	// Create validator with the workflow path
	validator := workflow.NewWorkflowValidator(fs, workflowPath)
	
	// Validate workflow
	result, err := validator.ValidateWorkflow(workflowDef)
	if err != nil {
		return ValidationResult{
			Success:      false,
			WorkflowName: workflowDef.Name,
			ErrorMessage: fmt.Sprintf("Validation failed: %s", err.Error()),
		}
	}
	
	// Create the result based on validation
	if result.IsValid() {
		return ValidationResult{
			Success:      true,
			WorkflowName: workflowDef.Name,
			Warnings:     result.Warnings,
		}
	} else {
		return ValidationResult{
			Success:      false,
			WorkflowName: workflowDef.Name,
			Errors:       result.Errors,
			ErrorMessage: fmt.Sprintf("Workflow validation failed with %d errors", len(result.Errors)),
		}
	}
}

// isValidWorkflowDir checks if a directory contains a valid workflow structure
func isValidWorkflowDir(fs io.FileSystem, dirPath string) bool {
	workflowYAMLPath := filepath.Join(dirPath, workflow.WorkflowConfigFile)
	promptsDirPath := filepath.Join(dirPath, workflow.PromptsDir)
	
	return fs.Exists(workflowYAMLPath) && fs.Exists(promptsDirPath)
}

// findWorkflowByName looks for a workflow by name in standard locations
func findWorkflowByName(fs io.FileSystem, name string) string {
	// Get standard workflow directories
	dirs := workflow.GetStandardWorkflowDirectories()
	
	// Check each directory for the named workflow
	for _, dir := range dirs {
		if !fs.Exists(dir) {
			continue
		}
		
		workflowPath := filepath.Join(dir, name)
		if isValidWorkflowDir(fs, workflowPath) {
			return workflowPath
		}
	}
	
	return ""
}

func init() {
	// No flags required for this command
} 
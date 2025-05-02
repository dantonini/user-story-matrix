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
				output.PrintError(fmt.Sprintf("Failed to load workflow: %s", err.Error()))
				os.Exit(1)
				return
			}
		} else {
			// Assume input is a name, try to find it in registry
			var err error
			workflowDef, err = registry.GetWorkflow(nameOrPath)
			if err != nil {
				output.PrintError(fmt.Sprintf("Workflow '%s' not found. Use 'usm workflow list' to see available workflows.", nameOrPath))
				os.Exit(1)
				return
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
						output.PrintProgress(fmt.Sprintf("Validating workflow '%s'", workflowDef.Name))
						output.PrintSuccess("Workflow is valid")
						return
					}
					
					// For other workflow types, we need a path to validate prompt references
					output.PrintError(fmt.Sprintf("Cannot find workflow directory for '%s'", nameOrPath))
					os.Exit(1)
					return
				}
			}
		}
		
		// Create validator with the workflow path
		validator := workflow.NewWorkflowValidator(fs, workflowPath)
		
		// Validate workflow
		output.PrintProgress(fmt.Sprintf("Validating workflow '%s'", workflowDef.Name))
		result, err := validator.ValidateWorkflow(workflowDef)
		if err != nil {
			output.PrintError(fmt.Sprintf("Validation failed: %s", err.Error()))
			os.Exit(1)
			return
		}
		
		// Display results
		if result.IsValid() {
			output.PrintSuccess("Workflow is valid")
			
			// Display any warnings
			if len(result.Warnings) > 0 {
				output.PrintWarning(fmt.Sprintf("Found %d warnings:", len(result.Warnings)))
				for _, warning := range result.Warnings {
					output.Print(fmt.Sprintf("- %s", warning))
				}
			}
		} else {
			output.PrintError(fmt.Sprintf("Workflow validation failed with %d errors:", len(result.Errors)))
			for _, err := range result.Errors {
				output.Print(fmt.Sprintf("- %s", err))
			}
			os.Exit(1)
		}
	},
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
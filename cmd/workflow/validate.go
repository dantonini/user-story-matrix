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
		// Get name or path from args
		nameOrPath := args[0]
		
		// Create output instance
		output := io.NewTerminalIO()
		
		// Create filesystem
		fs := io.NewOSFileSystem()
		
		// Determine if input is a path or a name
		var workflowPath string
		if fs.Exists(nameOrPath) && isValidWorkflowDir(fs, nameOrPath) {
			// Input is a path
			workflowPath = nameOrPath
		} else {
			// Assume input is a name, try to find it in standard locations
			workflowPath = findWorkflowByName(fs, nameOrPath)
			if workflowPath == "" {
				output.PrintError(fmt.Sprintf("Workflow '%s' not found", nameOrPath))
				os.Exit(1)
				return
			}
		}
		
		// Create validator
		validator := workflow.NewWorkflowValidator(fs)
		
		// Validate workflow
		output.PrintProgress(fmt.Sprintf("Validating workflow at %s", workflowPath))
		result := validator.ValidateWorkflow(workflowPath)
		
		// Display results
		if result.IsValid {
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
				output.Print(fmt.Sprintf("- %s", err.Error()))
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
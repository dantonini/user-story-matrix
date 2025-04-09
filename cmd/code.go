// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/workflow"
)

var resetFlag bool

// codeCmd represents the code command
var codeCmd = &cobra.Command{
	Use:   "code [change-request-file]",
	Short: "Execute the next step in a structured implementation workflow",
	Long: `The 'code' command provides a structured approach to implementing change requests.

It breaks down the implementation process into predefined steps, guides you through each step,
and keeps track of your progress. The workflow consists of 8 numbered steps:

1. Laying the foundation
2. Laying the foundation testing
3. Minimum Viable Implementation (MVI)
4. MVI testing
5. Extending functionalities
6. Extending functionalities testing
7. Final iteration
8. Final iteration testing

The command detects which step you're on, executes it by displaying the prompt,
and updates your progress. Progress is stored in a .step file, allowing you to 
resume where you left off.

Example:
  usm code docs/changes-request/2025-03-26-020055-code-command.blueprint.md

Use the --reset flag to start the workflow from the beginning:
  usm code --reset docs/changes-request/2025-03-26-020055-code-command.blueprint.md`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Create filesystem and IO interfaces
		fs := io.NewOSFileSystem()
		term := io.NewTerminalIOWithDebug(debug)

		// Create workflow manager
		wm := workflow.NewDefaultWorkflowManager(fs, term)

		// Get the change request path
		changeRequestPath := args[0]

		// Check if file exists
		if !fs.Exists(changeRequestPath) {
			term.PrintError(fmt.Sprintf("❌ Error: File %s not found.", changeRequestPath))
			os.Exit(1)
		}

		// Handle reset flag
		if resetFlag {
			if err := wm.ResetWorkflow(changeRequestPath); err != nil {
				term.PrintError(fmt.Sprintf("Failed to reset workflow: %s", err))
				os.Exit(1)
			}
			// Success message is shown by the ResetWorkflow method in debug mode
		}

		// Check if workflow is already complete
		complete, err := wm.IsWorkflowComplete(changeRequestPath)
		if err != nil {
			term.PrintError(fmt.Sprintf("Failed to check workflow completion: %s", err))
			os.Exit(1)
		}

		if complete {
			// Only show completion message in debug mode
			if term.IsDebugEnabled() {
				term.PrintSuccess(fmt.Sprintf("✅ All steps completed successfully for change request: %s", changeRequestPath))
			}
			os.Exit(0)
		}

		// Determine which step to execute
		nextStepIndex, err := wm.DetermineNextStep(changeRequestPath)
		if err != nil {
			term.PrintError(fmt.Sprintf("Failed to determine next step: %s", err))
			os.Exit(1)
		}

		// Special case: workflow is complete
		if nextStepIndex == -1 {
			// Only show completion message in debug mode
			if term.IsDebugEnabled() {
				term.PrintSuccess(fmt.Sprintf("✅ All steps completed successfully for change request: %s", changeRequestPath))
			}
			os.Exit(0)
		}

		// We can rely on the workflow manager to determine correct step since it already validated the index
		// Get the current workflow definition and step from the manager
		currentStep, err := wm.GetStepByIndex(nextStepIndex)
		if err != nil {
			term.PrintError(fmt.Sprintf("Invalid step index: %s", err))
			os.Exit(1)
		}

		// Execute the step - now just prints the prompt to stdout
		success, err := executeStep(changeRequestPath, currentStep, fs, term)
		if err != nil {
			term.PrintError(fmt.Sprintf("Failed to execute step: %s", err))
			os.Exit(1)
		}

		if !success {
			term.PrintError("Step execution failed.")
			os.Exit(1)
		}

		// Update state
		if err := wm.UpdateState(changeRequestPath, nextStepIndex+1); err != nil {
			term.PrintError(fmt.Sprintf("Failed to update workflow state: %s", err))
			os.Exit(1)
		}

		// Only show success messages if debug is enabled
		if term.IsDebugEnabled() {
			term.PrintSuccess(fmt.Sprintf("Completed step %d: %s", nextStepIndex+1, currentStep.Description))

			// Check if we've completed all steps
			isComplete, err := wm.IsWorkflowComplete(changeRequestPath)
			if err != nil {
				term.PrintWarning(fmt.Sprintf("Failed to check workflow completion: %s", err))
			} else if isComplete {
				term.PrintSuccess(fmt.Sprintf("✅ All steps completed successfully for change request: %s", changeRequestPath))
			} else {
				// Get the next step
				nextStep, err := wm.GetStepByIndex(nextStepIndex + 1)
				if err == nil {
					term.Print(fmt.Sprintf("\nNext step: %s", nextStep.Description))
				}
			}
		}
	},
}

// executeStep executes a workflow step and prints the processed prompt
func executeStep(changeRequestPath string, step workflow.WorkflowStep, fs io.FileSystem, term io.UserOutput) (bool, error) {
	executor := workflow.NewStepExecutor(fs, term)
	return executor.ExecuteStep(changeRequestPath, step)
}

// getDirectoryPath extracts the directory part of a file path
func getDirectoryPath(filePath string) string {
	return filePath[:len(filePath)-len(getFileName(filePath))]
}

// getFileName extracts the file name part of a file path
func getFileName(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' || filePath[i] == '\\' {
			return filePath[i+1:]
		}
	}
	return filePath
}

func init() {
	rootCmd.AddCommand(codeCmd)
	codeCmd.Flags().BoolVar(&resetFlag, "reset", false, "Reset the workflow and start from the beginning")
	logger.Debug("Code command added to root command")
} 
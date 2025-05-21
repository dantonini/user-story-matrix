// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/workflow"
)

var elaborateWorkflowNameFlag string
var elaborateWorkflowPathFlag string
var elaborateResetFlag bool

// elaborateCmd represents the elaborate command
var elaborateCmd = &cobra.Command{
	Use:   "elaborate [vague-description-file]",
	Short: "Transform a vague feature description into well-defined user stories",
	Long: `The 'elaborate' command helps refine vague feature ideas into well-defined user stories.

It processes a high-level feature description and breaks it down into multiple INVEST user stories
using a structured workflow. This command works similar to the 'code' command but is focused on
refining user stories rather than implementing code.

If a file containing the vague description is not provided, the command will prompt for input.
The workflow guides you through the process of transforming your initial description into
detailed, implementable user stories.

By default, it uses the 'elaborate' workflow, but you can specify a custom workflow with
--workflow or --workflow-path flags.

Examples:
  # Provide the description interactively
  usm elaborate

  # Use a file containing the vague description
  usm elaborate docs/vague/my-feature-idea.md

  # Use a custom workflow by name
  usm elaborate --workflow=my-custom-workflow docs/vague/my-feature-idea.md

  # Use a custom workflow by path
  usm elaborate --workflow-path=path/to/my-workflow docs/vague/my-feature-idea.md

  # Reset the workflow and start from the beginning
  usm elaborate --reset docs/vague/my-feature-idea.md`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Create filesystem and IO interfaces
		fs := io.NewOSFileSystem()
		term := io.NewTerminalIOWithDebug(debug)

		// Create workflow manager with custom workflow if specified
		var wm *workflow.WorkflowManager
		var err error
		
		if elaborateWorkflowPathFlag != "" {
			// Workflow path takes precedence over workflow name
			wm, err = workflow.NewWorkflowManagerWithPath(fs, term, elaborateWorkflowPathFlag)
			if err != nil {
				term.PrintError(fmt.Sprintf("❌ Error: %s", err.Error()))
				os.Exit(1)
			}
		} else if elaborateWorkflowNameFlag != "" {
			// Use named workflow
			wm, err = workflow.NewWorkflowManagerWithName(fs, term, elaborateWorkflowNameFlag)
			if err != nil {
				term.PrintError(fmt.Sprintf("❌ Error: %s", err.Error()))
				os.Exit(1)
			}
		} else {
			// Try to use the elaborate workflow, fallback to default if not found
			wm, err = workflow.NewWorkflowManagerWithName(fs, term, "elaborate")
			if err != nil {
				// Elaborate workflow not found, use default
				term.PrintWarning("Elaborate workflow not found, using default workflow")
				wm = workflow.NewDefaultWorkflowManager(fs, term)
			}
		}

		// Handle the case where no file is provided - prompt for input
		var vagueDescriptionPath string
		if len(args) == 0 {
			// Prompt the user for a vague description
			vagueDescription := promptForVagueDescription(term)
			
			// Create the vague descriptions directory if it doesn't exist
			vagueDir := "docs/vague"
			if err := fs.MkdirAll(vagueDir, 0755); err != nil {
				term.PrintError(fmt.Sprintf("❌ Error creating directory: %s", err.Error()))
				os.Exit(1)
			}
			
			// Generate filename with timestamp
			timestamp := time.Now().Format("20060102_150405")
			fileName := fmt.Sprintf("elaborate_%s.md", timestamp)
			vagueDescriptionPath = filepath.Join(vagueDir, fileName)
			
			// Write the description to the file
			if err := fs.WriteFile(vagueDescriptionPath, []byte(vagueDescription), 0644); err != nil {
				term.PrintError(fmt.Sprintf("❌ Error writing to file: %s", err.Error()))
				os.Exit(1)
			}
			
			// Let user know where the file is stored
			term.PrintSuccess(fmt.Sprintf("Saved vague description to: %s", vagueDescriptionPath))
			term.PrintProgress("The file will be used as input for the elaboration workflow.")
			term.PrintProgress(fmt.Sprintf("To continue later, run: usm elaborate %s", vagueDescriptionPath))
			term.Print("") // Empty line for readability
			os.Exit(0)
		} else {
			vagueDescriptionPath = args[0]
			
			// Check if file exists
			if !fs.Exists(vagueDescriptionPath) {
				term.PrintError(fmt.Sprintf("❌ Error: File %s not found.", vagueDescriptionPath))
				os.Exit(1)
			}
		}

		// Handle reset flag
		if elaborateResetFlag {
			if err := wm.ResetWorkflow(vagueDescriptionPath); err != nil {
				term.PrintError(fmt.Sprintf("Failed to reset workflow: %s", err))
				os.Exit(1)
			}
			// Success message is shown by the ResetWorkflow method in debug mode
		}

		// Check if workflow is already complete
		complete, err := wm.IsWorkflowComplete(vagueDescriptionPath)
		if err != nil {
			term.PrintError(fmt.Sprintf("Failed to check workflow completion: %s", err))
			os.Exit(1)
		}

		if complete {
			// Only show completion message in debug mode
			if term.IsDebugEnabled() {
				term.PrintSuccess(fmt.Sprintf("✅ All steps completed successfully for elaboration: %s", vagueDescriptionPath))
			}
			os.Exit(0)
		}

		// Determine which step to execute
		nextStepIndex, err := wm.DetermineNextStep(vagueDescriptionPath)
		if err != nil {
			term.PrintError(fmt.Sprintf("Failed to determine next step: %s", err))
			os.Exit(1)
		}

		// Special case: workflow is complete
		if nextStepIndex == -1 {
			// Always show completion message
			term.PrintSuccess(fmt.Sprintf("✅ All steps completed successfully for elaboration: %s", vagueDescriptionPath))
			term.PrintProgress("The final elaborated user stories can be found in the last output file.")
			os.Exit(0)
		}

		// Get the current step for displaying the status
		currentStep, err := wm.GetStepByIndex(nextStepIndex)
		if err != nil {
			term.PrintError(fmt.Sprintf("Invalid step index: %s", err))
			os.Exit(1)
		}

		// Calculate total steps
		totalSteps := 0
		// Try to get steps one by one until we hit an error (end of steps)
		for i := 0; ; i++ {
			_, err := wm.GetStepByIndex(i)
			if err != nil {
				totalSteps = i
				break
			}
		}

		// Show workflow status before executing the step
		term.PrintProgress(fmt.Sprintf("Executing step %d of %d: %s", nextStepIndex+1, totalSteps, currentStep.Description))
		term.Print("") // Empty line for readability

		// Execute the step
		success, err := executeStep(vagueDescriptionPath, currentStep, fs, term)
		if err != nil {
			term.PrintError(fmt.Sprintf("Failed to execute step: %s", err))
			os.Exit(1)
		}

		if !success {
			term.PrintError("Step execution failed.")
			os.Exit(1)
		}

		// Update state
		if err := wm.UpdateState(vagueDescriptionPath, nextStepIndex+1); err != nil {
			term.PrintError(fmt.Sprintf("Failed to update workflow state: %s", err))
			os.Exit(1)
		}

		// Always show success messages and next steps
		term.PrintSuccess(fmt.Sprintf("Completed step %d of %d: %s", nextStepIndex+1, totalSteps, currentStep.Description))

		// Check if we've completed all steps
		isComplete, err := wm.IsWorkflowComplete(vagueDescriptionPath)
		if err != nil {
			term.PrintWarning(fmt.Sprintf("Failed to check workflow completion: %s", err))
		} else if isComplete {
			term.PrintSuccess(fmt.Sprintf("✅ All steps completed successfully for elaboration: %s", vagueDescriptionPath))
			term.PrintProgress("The final elaborated user stories can be found in the last output file.")
		} else {
			// Get the next step
			nextStepIndex++
			if nextStepIndex < totalSteps {
				nextStep, err := wm.GetStepByIndex(nextStepIndex)
				if err == nil {
					term.Print(fmt.Sprintf("\nNext step: %s", nextStep.Description))
					term.PrintProgress(fmt.Sprintf("To continue, run: usm elaborate %s", vagueDescriptionPath))
				}
			}
		}
	},
}

// promptForVagueDescription prompts the user to enter a vague feature description
func promptForVagueDescription(term io.UserOutput) string {
	term.Print("Please enter a vague description of the feature you want to elaborate:")
	term.Print("(Type your description and press Enter + Ctrl+D when done)\n")
	
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		term.PrintError(fmt.Sprintf("Error reading input: %s", err.Error()))
		os.Exit(1)
	}
	
	return strings.Join(lines, "\n")
}

func init() {
	rootCmd.AddCommand(elaborateCmd)
	elaborateCmd.Flags().BoolVar(&elaborateResetFlag, "reset", false, "Reset the workflow and start from the beginning")
	elaborateCmd.Flags().StringVar(&elaborateWorkflowNameFlag, "workflow", "", "Use a custom workflow by name")
	elaborateCmd.Flags().StringVar(&elaborateWorkflowPathFlag, "workflow-path", "", "Use a custom workflow from a specific path")
	logger.Debug("Elaborate command added to root command")
} 
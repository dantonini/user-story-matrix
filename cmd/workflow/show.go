// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// ShowCmd represents the workflow show command
var ShowCmd = &cobra.Command{
	Use:   "show [workflow-name]",
	Short: "Show workflow details",
	Long: `Show detailed information about a workflow.

This command displays comprehensive information about a workflow, including:
- Workflow metadata (name, description)
- A list of all steps with their IDs and descriptions
- Information about variables used in the workflow

The command supports different output formats:
--format=text (default): Human-readable text format
--format=markdown: Markdown output for documentation
--format=json: JSON output for programmatic use

Examples:
  # Show details of the standard workflow
  usm workflow show standard

  # Show details of a custom workflow
  usm workflow show my-custom-workflow

  # Show workflow details in markdown format
  usm workflow show my-workflow --format=markdown
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get workflow name
		workflowName := args[0]
		
		// Get output format
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			format = "text" // Default to text format if error
		}
		
		// Create output instance
		output := io.NewTerminalIO()
		
		// Get global registry
		registry := workflow.GetGlobalRegistry()
		
		// Try to find the workflow
		output.PrintProgress(fmt.Sprintf("Looking for workflow '%s'...", workflowName))
		wf, err := registry.GetWorkflow(workflowName)
		if err != nil {
			output.PrintError(fmt.Sprintf("Workflow '%s' not found. Use 'usm workflow list' to see available workflows.", workflowName))
			os.Exit(1)
		}
		
		// Display workflow details
		switch format {
		case "json":
			// TODO: Implement JSON output
			output.PrintError("JSON output format not implemented yet")
		case "markdown":
			// TODO: Implement Markdown output
			output.PrintError("Markdown output format not implemented yet")
		default:
			// Text output (default)
			output.PrintSuccess(fmt.Sprintf("Workflow: %s", wf.Name))
			output.Print(fmt.Sprintf("Description: %s", wf.Description))
			output.Print(fmt.Sprintf("Steps: %d", len(wf.Steps)))
			output.Print("")
			
			// Print table of steps
			headers := []string{"ID", "DESCRIPTION"}
			rows := make([][]string, len(wf.Steps))
			
			for i, step := range wf.Steps {
				rows[i] = []string{step.ID, step.Description}
			}
			
			output.PrintTable(headers, rows)
			
			// Print variables info if any steps have variables
			hasVariables := false
			for _, step := range wf.Steps {
				if len(step.Variables) > 0 {
					hasVariables = true
					break
				}
			}
			
			if hasVariables {
				output.Print("")
				output.Print("Variables:")
				output.Print("This workflow uses template variables in some steps. Use 'usm workflow validate' for validation.")
			}
		}
	},
}

func init() {
	// Add flags
	ShowCmd.Flags().StringP("format", "f", "text", "Output format (text, markdown, json)")
} 
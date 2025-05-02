// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// WorkflowDetails represents a workflow with its steps for output formatting
type WorkflowDetails struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Steps       []StepDetail `json:"steps"`
}

// StepDetail represents a workflow step with its variables for output formatting
type StepDetail struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Variables   map[string]string `json:"variables,omitempty"`
}

// ShowResult represents the result of the show command execution
type ShowResult struct {
	Success      bool
	ErrorMessage string
	Output       string
}

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
		workflowName := args[0]
		
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			format = "text" // Default to text format if error
		}
		
		termIO := io.NewTerminalIO()
		
		// Execute the actual logic
		result := showWorkflow(workflowName, format, termIO)
		
		// Handle the result
		if !result.Success {
			termIO.PrintError(result.ErrorMessage)
			return
		}
		
		// If there's output to print, print it
		if result.Output != "" {
			termIO.Print(result.Output)
		}
	},
}

// showWorkflow contains the actual logic for showing workflow details
// It never calls os.Exit directly
func showWorkflow(workflowName, format string, output *io.TerminalIO) ShowResult {
	registry := workflow.GetGlobalRegistry()
	
	// Try to find the workflow
	output.PrintProgress(fmt.Sprintf("Looking for workflow '%s'...", workflowName))
	wf, err := registry.GetWorkflow(workflowName)
	if err != nil {
		return ShowResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Workflow '%s' not found. Use 'usm workflow list' to see available workflows.", workflowName),
		}
	}
	
	// Create a workflow details structure for output
	details := WorkflowDetails{
		Name:        wf.Name,
		Description: wf.Description,
		Steps:       make([]StepDetail, len(wf.Steps)),
	}
	
	// Extract variables from each step
	for i, step := range wf.Steps {
		details.Steps[i] = StepDetail{
			ID:          step.ID,
			Description: step.Description,
			Variables:   step.Variables,
		}
	}
	
	// Format the output based on the requested format
	var resultOutput string
	
	// Display workflow details based on format
	switch format {
	case "json":
		// JSON output
		jsonData, err := json.MarshalIndent(details, "", "  ")
		if err != nil {
			return ShowResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("Failed to generate JSON: %s", err),
			}
		}
		resultOutput = string(jsonData)
		
	case "markdown":
		// Markdown output
		resultOutput = generateMarkdownOutput(details)
		
	default:
		// Text output (default) - print directly to output rather than returning
		// Basic info
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
		
		// Print variables info
		output.Print("")
		output.Print("Variables:")
		hasVariables := false
		
		for _, step := range wf.Steps {
			if len(step.Variables) > 0 {
				hasVariables = true
				output.Print(fmt.Sprintf("  Step '%s':", step.ID))
				
				for varName, varValue := range step.Variables {
					valueStr := varValue
					if len(valueStr) > 50 {
						valueStr = valueStr[:47] + "..."
					}
					output.Print(fmt.Sprintf("    %s: %s", varName, valueStr))
				}
			}
		}
		
		if !hasVariables {
			output.Print("  No variables defined in any steps.")
			output.Print("  You can add variables in the workflow.yaml file:")
			output.Print("    steps:")
			output.Print("      - id: step-id")
			output.Print("        variables:")
			output.Print("          key: value")
		}
		
		// Text output is printed directly to terminal so we don't need to return it
		resultOutput = ""
	}
	
	return ShowResult{
		Success: true,
		Output:  resultOutput,
	}
}

// generateMarkdownOutput creates a markdown representation of the workflow
func generateMarkdownOutput(details WorkflowDetails) string {
	var sb strings.Builder
	
	// Workflow header
	sb.WriteString(fmt.Sprintf("# Workflow: %s\n\n", details.Name))
	sb.WriteString(fmt.Sprintf("%s\n\n", details.Description))
	
	// Steps section
	sb.WriteString("## Steps\n\n")
	
	for i, step := range details.Steps {
		sb.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, step.ID))
		sb.WriteString(fmt.Sprintf("%s\n\n", step.Description))
		
		// Variables table if any
		if len(step.Variables) > 0 {
			sb.WriteString("#### Variables\n\n")
			sb.WriteString("| Name | Value |\n")
			sb.WriteString("|------|-------|\n")
			
			for name, value := range step.Variables {
				valueStr := value
				if len(valueStr) > 50 {
					valueStr = valueStr[:47] + "..."
				}
				sb.WriteString(fmt.Sprintf("| %s | %s |\n", name, valueStr))
			}
			sb.WriteString("\n")
		}
	}
	
	// Usage example
	sb.WriteString("## Usage\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("usm code --workflow=%s path/to/change-request.blueprint.md\n", details.Name))
	sb.WriteString("```\n")
	
	return sb.String()
}

func init() {
	// Add flags
	ShowCmd.Flags().StringP("format", "f", "text", "Output format (text, markdown, json)")
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// ListCmd represents the workflow list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available workflows",
	Long: `List all available workflows from all sources.

This command discovers workflows from:
1. .usm/workflows/ in the current directory (project-specific)
2. ~/.usm/workflows/ in the user's home directory (user-specific)
3. Built-in workflows provided by the application

Each workflow is displayed with its:
- Name: Unique identifier for the workflow
- Description: Brief explanation of the workflow's purpose
- Source: Where the workflow is defined (built-in, user, project)
- Path: File system location of the workflow definition

The command supports different output formats:
--format=text (default): Human-readable text table
--format=json: JSON output for programmatic use

Examples:
  # List all workflows in default format
  usm workflow list

  # List all workflows in JSON format
  usm workflow list --format=json
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get output format
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			format = "text" // Default to text format if error
		}
		
		// Create output instance
		output := io.NewTerminalIO()
		
		// Create filesystem
		fs := io.NewOSFileSystem()
		
		// Get global registry
		registry := workflow.GetGlobalRegistry()
		
		// Discover workflows
		output.PrintProgress("Discovering workflows...")
		workflows := registry.DiscoverWorkflows(fs)
		
		// Display results
		if len(workflows) == 0 {
			output.Print("No workflows found")
			return
		}
		
		// Sort workflows by name for consistent output
		names := make([]string, 0, len(workflows))
		for name := range workflows {
			names = append(names, name)
		}
		sort.Strings(names)
		
		// Format and display results
		switch format {
		case "json":
			// TODO: Implement JSON output
			output.PrintError("JSON output format not implemented yet")
		default:
			// Text output (default)
			output.PrintSuccess(fmt.Sprintf("Found %d workflows:", len(workflows)))
			
			// Print table header and rows
			headers := []string{"NAME", "DESCRIPTION", "SOURCE", "PATH"}
			rows := make([][]string, len(names))
			
			for i, name := range names {
				wf := workflows[name]
				// TODO: Add source and path information
				source := "built-in" // Placeholder
				path := "-"          // Placeholder
				rows[i] = []string{wf.Name, wf.Description, source, path}
			}
			
			output.PrintTable(headers, rows)
		}
	},
}

func init() {
	// Add flags
	ListCmd.Flags().StringP("format", "f", "text", "Output format (text, json)")
} 
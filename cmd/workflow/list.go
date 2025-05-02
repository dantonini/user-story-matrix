// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// WorkflowInfo represents a workflow for output formatting
type WorkflowInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Path        string `json:"path"`
}

// ListResult represents the result of the list command execution
type ListResult struct {
	Success      bool
	ErrorMessage string
	Output       string
	Workflows    []WorkflowInfo
	NoWorkflows  bool
}

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

		output := io.NewTerminalIO()
		fs := io.NewOSFileSystem()
		
		// Execute the actual logic
		result := listWorkflows(format, output, fs)
		
		// Handle the result
		if !result.Success {
			output.PrintError(result.ErrorMessage)
			return
		}
		
		// If there's output to print, print it
		if result.Output != "" {
			output.Print(result.Output)
		}
	},
}

// listWorkflows contains the actual logic for listing workflows
// It never calls os.Exit directly
func listWorkflows(format string, output io.UserOutput, fs io.FileSystem) ListResult {
	registry := workflow.GetGlobalRegistry()

	output.PrintProgress("Discovering workflows...")
	workflowDefs := registry.DiscoverWorkflows(fs)

	if len(workflowDefs) == 0 {
		output.Print("No workflows found")
		return ListResult{
			Success:     true,
			NoWorkflows: true,
		}
	}

	workflowInfos := make([]WorkflowInfo, 0, len(workflowDefs))

	for name, wf := range workflowDefs {
		// Default source and path
		source := "unknown"
		path := "-"

		// Check if it's a built-in workflow
		if _, err := registry.GetWorkflow(name); err == nil {
			builtInWorkflows := []string{"standard", "old-workflow"} // Known built-in workflows
			for _, bw := range builtInWorkflows {
				if name == bw {
					source = workflow.SourceBuiltIn
					break
				}
			}
		}

		// For non-built-in workflows, try to determine source from path
		if source == "unknown" {
			// Check registryInfo in the registry's cache if available
			// Since we can't directly access the registry's cache, we use heuristics
			if strings.HasPrefix(name, "project-") {
				source = workflow.SourceProject
			} else if strings.HasPrefix(name, "user-") {
				source = workflow.SourceUser
			}

			// Get more accurate path information where possible
			// This is a simplified approach since we can't directly access the registry's cache
			if source == workflow.SourceProject {
				path = ".usm/workflows/" + name
			} else if source == workflow.SourceUser {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					// If we can't get the home directory, use a placeholder
					path = "~/.usm/workflows/" + name
				} else {
					path = filepath.Join(homeDir, ".usm/workflows/", name)
				}
			}
		}

		workflowInfos = append(workflowInfos, WorkflowInfo{
			Name:        wf.Name,
			Description: wf.Description,
			Source:      source,
			Path:        path,
		})
	}

	// Sort workflows by name for consistent output
	sort.Slice(workflowInfos, func(i, j int) bool {
		return workflowInfos[i].Name < workflowInfos[j].Name
	})

	var resultOutput string
	
	// Format and display results
	switch format {
	case "json":
		// JSON output
		jsonData, err := json.MarshalIndent(workflowInfos, "", "  ")
		if err != nil {
			return ListResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("Failed to generate JSON: %s", err),
			}
		}
		resultOutput = string(jsonData)
		
	default:
		// Text output (default)
		output.PrintSuccess(fmt.Sprintf("Found %d workflows:", len(workflowInfos)))

		// Print table header and rows
		headers := []string{"NAME", "DESCRIPTION", "SOURCE", "PATH"}
		rows := make([][]string, len(workflowInfos))

		for i, info := range workflowInfos {
			rows[i] = []string{info.Name, info.Description, info.Source, info.Path}
		}

		output.PrintTable(headers, rows)
		resultOutput = "" // Text output is printed directly to terminal
	}

	return ListResult{
		Success:   true,
		Output:    resultOutput,
		Workflows: workflowInfos,
	}
}

func init() {
	// Add flags
	ListCmd.Flags().StringP("format", "f", "text", "Output format (text, json)")
} 
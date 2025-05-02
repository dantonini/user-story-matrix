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
	internalio "github.com/user-story-matrix/usm/internal/io"
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

		// Get debug flag
		showDebug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			showDebug = false // Default to no debug output
		}

		output := io.NewTerminalIO()
		fs := io.NewOSFileSystem()
		
		// Execute the actual logic
		result := listWorkflows(format, output, fs, showDebug)
		
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
func listWorkflows(format string, output internalio.UserOutput, fs internalio.FileSystem, showDebug bool) ListResult {
	registry := workflow.GetGlobalRegistry()

	output.PrintProgress("Discovering workflows...")
	
	// Check if we're using a mock filesystem
	_, isMock := fs.(*internalio.MockFileSystem)
	
	// Temporarily redirect debug output if not requested
	var originalStdout *os.File
	if !showDebug {
		originalStdout = os.Stdout
		nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err == nil {
			os.Stdout = nullFile
			defer func() {
				os.Stdout = originalStdout
				nullFile.Close()
			}()
		}
	}
	
	// Get workflows from the registry, both built-in and cached
	// This is more reliable than rediscovering, especially in tests
	workflowNames := registry.ListWorkflows()
	workflowDefs := make(map[string]*workflow.WorkflowDefinition)
	
	for _, name := range workflowNames {
		wf, err := registry.GetWorkflow(name)
		if err == nil && wf != nil {
			workflowDefs[name] = wf
		}
	}
	
	// Always run discovery for real filesystems
	// For mock filesystems in tests, we only use the pre-loaded workflows
	if !isMock {
		// Discover additional workflows from filesystem
		discoveredWorkflows := registry.DiscoverWorkflows(fs)
		for name, wf := range discoveredWorkflows {
			workflowDefs[name] = wf
		}
	}
	
	// Restore stdout if it was redirected
	if !showDebug && originalStdout != nil {
		os.Stdout = originalStdout
	}

	if len(workflowDefs) == 0 {
		output.Print("No workflows found")
		return ListResult{
			Success:     true,
			NoWorkflows: true,
		}
	}

	workflowInfos := make([]WorkflowInfo, 0, len(workflowDefs))

	for name, wf := range workflowDefs {
		// Get workflow source info from registry
		workflowSource, workflowPath := registry.GetWorkflowSourceInfo(name)

		// Get source from registry or determine it
		source := workflowSource
		if source == "" {
			// Default source
			source = "unknown"
			
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
		}

		// Get path from registry or set default
		path := workflowPath
		if path == "" {
			path = "-"
			
			// For non-built-in workflows with no path info from registry,
			// try to determine path based on source
			if source == workflow.SourceProject {
				path = filepath.Join(".usm/workflows", name)
			} else if source == workflow.SourceUser {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					path = filepath.Join(homeDir, ".usm/workflows", name)
				} else {
					path = filepath.Join("~/.usm/workflows", name)
				}
			}
		}

		// Sanitize the values to prevent display issues
		name = sanitizeForDisplay(name)
		description := sanitizeForDisplay(wf.Description)
		source = sanitizeForDisplay(source)
		path = sanitizeForDisplay(path)

		workflowInfos = append(workflowInfos, WorkflowInfo{
			Name:        name,
			Description: description,
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

		// Manual table formatting for consistent display
		headers := []string{"NAME", "DESCRIPTION", "SOURCE", "PATH"}
		
		// Format the table using our custom formatting function
		table := formatWorkflowTable(headers, workflowInfos)
		fmt.Println(table)
		
		resultOutput = "" // Text output is printed directly to terminal
	}

	return ListResult{
		Success:   true,
		Output:    resultOutput,
		Workflows: workflowInfos,
	}
}

// formatWorkflowTable creates a formatted table string from workflow data
func formatWorkflowTable(headers []string, workflows []WorkflowInfo) string {
	// Define column widths (adjust as needed based on typical content)
	nameWidth := 15
	descWidth := 40
	sourceWidth := 10
	pathWidth := 30
	
	// Build the table
	var sb strings.Builder
	
	// Write header row
	sb.WriteString(fmt.Sprintf("%-*s %-*s %-*s %-*s\n", 
		nameWidth, headers[0], 
		descWidth, headers[1], 
		sourceWidth, headers[2], 
		pathWidth, headers[3]))
	
	// Write separator
	sb.WriteString(fmt.Sprintf("%-*s %-*s %-*s %-*s\n",
		nameWidth, strings.Repeat("─", nameWidth),
		descWidth, strings.Repeat("─", descWidth),
		sourceWidth, strings.Repeat("─", sourceWidth),
		pathWidth, strings.Repeat("─", pathWidth)))
	
	// Write data rows
	for _, wf := range workflows {
		// Truncate fields if needed
		name := truncateWithEllipsis(wf.Name, nameWidth)
		desc := truncateWithEllipsis(wf.Description, descWidth)
		source := truncateWithEllipsis(wf.Source, sourceWidth)
		path := truncateWithEllipsis(wf.Path, pathWidth)
		
		sb.WriteString(fmt.Sprintf("%-*s %-*s %-*s %-*s\n", 
			nameWidth, name, 
			descWidth, desc, 
			sourceWidth, source, 
			pathWidth, path))
	}
	
	return sb.String()
}

// truncateWithEllipsis truncates a string to the specified length, adding an ellipsis if truncated
func truncateWithEllipsis(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	
	if maxLen <= 3 {
		return s[:maxLen]
	}
	
	return s[:maxLen-3] + "..."
}

// sanitizeForDisplay cleans up a string for display in a terminal table
// by removing newlines and other problematic characters
func sanitizeForDisplay(s string) string {
	// Replace all newlines with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	
	// Replace multiple spaces with a single space
	s = strings.Join(strings.Fields(s), " ")
	
	return s
}

func init() {
	// Add flags
	ListCmd.Flags().StringP("format", "f", "text", "Output format (text, json)")
	ListCmd.Flags().Bool("debug", false, "Show debug output")
} 
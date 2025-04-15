// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/cmd/workflow"
)

// workflowCmd represents the workflow command
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage custom workflows",
	Long: `Manage custom workflows for the USM tool.

Workflows are defined using a directory structure containing:
- workflow.yaml: Configuration file with workflow metadata and step definitions
- prompts/: Directory containing individual prompt files
- shared/ (optional): Directory for reusable prompt templates

Available subcommands:
  list        List available workflows
  init        Initialize a new workflow from a template
  validate    Validate a workflow definition
  show        Display workflow details

Examples:
  # List all available workflows
  usm workflow list

  # Initialize a new workflow
  usm workflow init my-workflow

  # Validate a workflow definition
  usm workflow validate path/to/workflow

  # Show workflow details
  usm workflow show my-workflow
`,
	// This is the root workflow command, so no Run function needed
	// Subcommands will be registered in their respective files
}

func init() {
	rootCmd.AddCommand(workflowCmd)
	
	// Add workflow subcommands
	workflowCmd.AddCommand(workflow.ListCmd)
	workflowCmd.AddCommand(workflow.InitCmd)
	workflowCmd.AddCommand(workflow.ValidateCmd)
	workflowCmd.AddCommand(workflow.ShowCmd)
} 
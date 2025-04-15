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

// Templates for workflow initialization
const (
	defaultTemplate = "default"
	fullTemplate    = "full"
	blankTemplate   = "blank"
)

// InitCmd represents the workflow init command
var InitCmd = &cobra.Command{
	Use:   "init [workflow-name]",
	Short: "Initialize a new workflow",
	Long: `Initialize a new workflow with the specified name.

This command creates a directory structure for a new workflow:
- workflow.yaml: Configuration file with workflow metadata and step definitions
- prompts/: Directory containing individual prompt files
- shared/ (optional): Directory for reusable prompt templates

Available templates:
- default: A minimal sample workflow with basic functionality
- full: A comprehensive workflow with more extensive examples
- blank: A skeleton structure with minimal content

By default, the workflow is created in the .usm/workflows/ directory in the
current project. Use the --global flag to create it in ~/.usm/workflows/ instead.

Examples:
  # Create a new workflow with the default template
  usm workflow init my-workflow

  # Create a new workflow with the full template
  usm workflow init my-workflow --template=full

  # Create a new workflow in the global directory
  usm workflow init my-workflow --global
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get workflow name from args
		workflowName := args[0]
		
		// Get flags
		global, err := cmd.Flags().GetBool("global")
		if err != nil {
			global = false // Default to local if error
		}
		
		templateName, err := cmd.Flags().GetString("template")
		if err != nil {
			templateName = defaultTemplate // Default to default template if error
		}
		
		// Create output instance
		output := io.NewTerminalIO()
		
		// Show which template we're using (debug message)
		if output.IsDebugEnabled() {
			output.Print(fmt.Sprintf("Using template: %s", templateName))
			// NOTE: Full template selection implementation will come later
		}
		
		// Create filesystem
		fs := io.NewOSFileSystem()
		
		// Determine the target directory
		var targetDir string
		if global {
			// Use global workflows directory (user home)
			homeDir := workflow.GetStandardWorkflowDirectories()[1] // Use the second directory (user home)
			targetDir = filepath.Join(homeDir, workflowName)
		} else {
			// Use project workflows directory
			targetDir = filepath.Join(".usm", "workflows", workflowName)
		}
		
		// Check if workflow already exists
		if fs.Exists(targetDir) {
			output.PrintError(fmt.Sprintf("Workflow '%s' already exists at %s", workflowName, targetDir))
			os.Exit(1)
		}
		
		// Create the workflow directory
		output.PrintProgress(fmt.Sprintf("Creating workflow '%s' at %s", workflowName, targetDir))
		
		// Create base directories
		err = fs.MkdirAll(targetDir, 0755)
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to create workflow directory: %s", err.Error()))
			os.Exit(1)
		}
		
		// Create prompts directory
		promptsDir := filepath.Join(targetDir, workflow.PromptsDir)
		err = fs.MkdirAll(promptsDir, 0755)
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to create prompts directory: %s", err.Error()))
			os.Exit(1)
		}
		
		// Create shared directory (optional)
		sharedDir := filepath.Join(targetDir, workflow.SharedDir)
		err = fs.MkdirAll(sharedDir, 0755)
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to create shared directory: %s", err.Error()))
			os.Exit(1)
		}
		
		// TODO: Implement template selection based on templateName
		// For now, we'll just create a minimal workflow as a placeholder
		
		// Create workflow.yaml
		workflowYAML := fmt.Sprintf(`name: "%s"
description: "Custom workflow created with usm workflow init"
steps:
  - id: "01-step-one"
    description: "First step"
    prompt: "prompts/step1.md"
    variables:
      key1: "value1"
      key2: "value2"

  - id: "02-step-two"
    description: "Second step"
    prompt: "prompts/step2.md"
    variables:
      key1: "value1"
      key2: "value2"
`, workflowName)
		
		// Write workflow.yaml
		err = fs.WriteFile(filepath.Join(targetDir, workflow.WorkflowConfigFile), []byte(workflowYAML), 0644)
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to create workflow.yaml: %s", err.Error()))
			os.Exit(1)
		}
		
		// Create step1.md
		step1Content := `# Step 1: First Step

This is the first step of your custom workflow. You can use variable substitution with {{ .key1 }} and {{ .key2 }}.

You can also use default values: {{ .optional_var | default "default value" }}.

## Instructions

1. First instruction
2. Second instruction
3. Third instruction

Feel free to customize this prompt for your specific needs.
`
		
		// Write step1.md
		err = fs.WriteFile(filepath.Join(promptsDir, "step1.md"), []byte(step1Content), 0644)
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to create step1.md: %s", err.Error()))
			os.Exit(1)
		}
		
		// Create step2.md
		step2Content := `# Step 2: Second Step

This is the second step of your custom workflow. You can use variable substitution with {{ .key1 }} and {{ .key2 }}.

You can also use default values: {{ .optional_var | default "default value" }}.

## Instructions

1. First instruction
2. Second instruction
3. Third instruction

Feel free to customize this prompt for your specific needs.
`
		
		// Write step2.md
		err = fs.WriteFile(filepath.Join(promptsDir, "step2.md"), []byte(step2Content), 0644)
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to create step2.md: %s", err.Error()))
			os.Exit(1)
		}
		
		// Create README.md in the shared directory
		readmeContent := `# Shared Templates

This directory contains shared templates that can be included in multiple prompts.

## Usage

You can include these shared templates in your prompts using the following syntax:

{{ template "shared/example.md" . }}

This will include the content of shared/example.md in your prompt and pass the current context to it.
`
		
		// Write README.md
		err = fs.WriteFile(filepath.Join(sharedDir, "README.md"), []byte(readmeContent), 0644)
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to create shared README.md: %s", err.Error()))
			os.Exit(1)
		}
		
		// Print success message
		output.PrintSuccess(fmt.Sprintf("Workflow '%s' created successfully at %s", workflowName, targetDir))
		
		// Print next steps
		output.Print("Next steps:")
		output.Print("1. Edit workflow.yaml to define your workflow steps")
		output.Print("2. Create prompts in the prompts/ directory")
		output.Print("3. Validate your workflow with 'usm workflow validate'")
	},
}

func init() {
	// Add flags
	InitCmd.Flags().BoolP("global", "g", false, "Create workflow in global directory (~/.usm/workflows)")
	InitCmd.Flags().StringP("template", "t", defaultTemplate, "Template to use (default, full, blank)")
} 
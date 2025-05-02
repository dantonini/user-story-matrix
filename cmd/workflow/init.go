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

// WorkflowTemplate contains all the files and content for a template
type WorkflowTemplate struct {
	Name        string
	Description string
	Files       map[string]string
}

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
		
		output := io.NewTerminalIO()
		
		fs := io.NewOSFileSystem()
		
		var targetDir string
		if global {
			// Use global workflows directory (user home)
			homeDir, err := os.UserHomeDir()
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to find user home directory: %s", err.Error()))
				os.Exit(1)
			}
			targetDir = filepath.Join(homeDir, ".usm", "workflows", workflowName)
		} else {
			// Use project workflows directory
			targetDir = filepath.Join(".usm", "workflows", workflowName)
		}
		
		if fs.Exists(targetDir) {
			output.PrintError(fmt.Sprintf("Workflow '%s' already exists at %s", workflowName, targetDir))
			os.Exit(1)
		}
		
		// Get template
		template, ok := getTemplate(templateName, workflowName)
		if !ok {
			output.PrintError(fmt.Sprintf("Unknown template '%s'. Available templates: default, full, blank", templateName))
			os.Exit(1)
		}
		
		// Create the workflow
		output.PrintProgress(fmt.Sprintf("Creating workflow '%s' at %s using template '%s'", workflowName, targetDir, template.Name))
		
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
		
		// Create shared directory
		sharedDir := filepath.Join(promptsDir, workflow.SharedDir)
		err = fs.MkdirAll(sharedDir, 0755)
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to create shared directory: %s", err.Error()))
			os.Exit(1)
		}
		
		// Create all template files
		for filePath, content := range template.Files {
			fullPath := filepath.Join(targetDir, filePath)
			
			// Ensure parent directory exists
			dir := filepath.Dir(fullPath)
			if !fs.Exists(dir) {
				err = fs.MkdirAll(dir, 0755)
				if err != nil {
					output.PrintError(fmt.Sprintf("Failed to create directory %s: %s", dir, err.Error()))
					os.Exit(1)
				}
			}
			
			// Write file
			err = fs.WriteFile(fullPath, []byte(content), 0644)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to create file %s: %s", filePath, err.Error()))
				os.Exit(1)
			}
		}
		
		// Print success message
		output.PrintSuccess(fmt.Sprintf("Workflow '%s' created successfully at %s", workflowName, targetDir))
		
		// Print next steps based on template
		switch templateName {
		case blankTemplate:
			output.Print("Next steps:")
			output.Print("1. Edit workflow.yaml to define your workflow steps")
			output.Print("2. Create prompts in the prompts/ directory")
			output.Print("3. Validate your workflow with 'usm workflow validate'")
		case defaultTemplate:
			output.Print("Next steps:")
			output.Print("1. Customize the workflow steps in workflow.yaml")
			output.Print("2. Edit the prompt content in prompts/ directory")
			output.Print("3. Validate your workflow with 'usm workflow validate'")
		case fullTemplate:
			output.Print("Next steps:")
			output.Print("1. Review the example workflow in workflow.yaml")
			output.Print("2. Customize the prompts in prompts/ directory")
			output.Print("3. Try reusing templates from the shared/ directory")
			output.Print("4. Validate your workflow with 'usm workflow validate'")
		}
	},
}

// getTemplate returns the template configuration for the given template name
func getTemplate(templateName string, workflowName string) (WorkflowTemplate, bool) {
	switch templateName {
	case blankTemplate:
		return getBlankTemplate(workflowName), true
	case fullTemplate:
		return getFullTemplate(workflowName), true
	case defaultTemplate:
		return getDefaultTemplate(workflowName), true
	default:
		return WorkflowTemplate{}, false
	}
}

// getBlankTemplate returns a minimal template with just the basic structure
func getBlankTemplate(workflowName string) WorkflowTemplate {
	template := WorkflowTemplate{
		Name:        "blank",
		Description: "A minimal workflow structure",
		Files:       make(map[string]string),
	}
	
	// Add workflow.yaml
	template.Files[workflow.WorkflowConfigFile] = fmt.Sprintf(`name: "%s"
description: "A minimal workflow"
steps:
  - id: "01-step-one"
    description: "First step"
    prompt: "prompts/step1.md"
`, workflowName)
	
	// Add step1.md
	template.Files[filepath.Join(workflow.PromptsDir, "step1.md")] = `# Step 1

This is a minimal prompt template.

Edit this file to define your prompt.
`
	
	// Add README.md
	template.Files["README.md"] = fmt.Sprintf(`# %s Workflow

A minimal workflow structure that you can customize.

## Directory Structure

- workflow.yaml: Workflow configuration
- prompts/: Prompt files
`, workflowName)
	
	return template
}

// getDefaultTemplate returns the default template with basic functionality
func getDefaultTemplate(workflowName string) WorkflowTemplate {
	template := WorkflowTemplate{
		Name:        "default",
		Description: "A standard workflow with basic functionality",
		Files:       make(map[string]string),
	}
	
	// Add workflow.yaml
	template.Files[workflow.WorkflowConfigFile] = fmt.Sprintf(`name: "%s"
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
	
	// Add step1.md
	template.Files[filepath.Join(workflow.PromptsDir, "step1.md")] = `# Step 1: First Step

This is the first step of your custom workflow. You can use variable substitution with {{ .key1 }} and {{ .key2 }}.

You can also use default values: {{ .optional_var | default "default value" }}.

## Instructions

1. First instruction
2. Second instruction
3. Third instruction

Feel free to customize this prompt for your specific needs.
`
	
	// Add step2.md
	template.Files[filepath.Join(workflow.PromptsDir, "step2.md")] = `# Step 2: Second Step

This is the second step of your custom workflow. You can use variable substitution with {{ .key1 }} and {{ .key2 }}.

You can also use default values: {{ .optional_var | default "default value" }}.

## Instructions

1. First instruction
2. Second instruction
3. Third instruction

Feel free to customize this prompt for your specific needs.
`
	
	// Add shared README.md
	template.Files[filepath.Join(workflow.PromptsDir, workflow.SharedDir, "README.md")] = `# Shared Templates

This directory contains shared templates that can be included in multiple prompts.

## Usage

You can include these shared templates in your prompts using the following syntax:

{{ template "shared/example.md" . }}

This will include the content of shared/example.md in your prompt and pass the current context to it.
`
	
	// Add README.md
	template.Files["README.md"] = fmt.Sprintf(`# %s Workflow

A custom workflow created with the usm workflow init command.

## Directory Structure

- workflow.yaml: Workflow configuration
- prompts/: Prompt files
  - step1.md: First step prompt
  - step2.md: Second step prompt
  - shared/: Shared template fragments

## Usage

To use this workflow:

` + "```bash" + `
usm code --workflow=%s path/to/change-request.blueprint.md
` + "```" + `

## Customization

You can customize this workflow by:

1. Editing workflow.yaml to change steps or add new ones
2. Modifying prompt files in the prompts/ directory
3. Adding shared templates in the prompts/shared/ directory
`, workflowName, workflowName)
	
	return template
}

// getFullTemplate returns a comprehensive template with more examples
func getFullTemplate(workflowName string) WorkflowTemplate {
	template := WorkflowTemplate{
		Name:        "full",
		Description: "A comprehensive workflow with extended functionality",
		Files:       make(map[string]string),
	}
	
	// Add workflow.yaml
	template.Files[workflow.WorkflowConfigFile] = fmt.Sprintf(`name: "%s"
description: "Comprehensive workflow with extended functionality"
steps:
  - id: "01-foundation"
    description: "Lay the foundation"
    prompt: "prompts/foundation.md"
    variables:
      phase: "foundation"
      focus: "architecture and interfaces"

  - id: "02-foundation-testing"
    description: "Test the foundation"
    prompt: "prompts/testing.md"
    variables:
      phase: "foundation"
      focus: "architecture and interfaces"
      coverage_goal: "80%%"

  - id: "03-implementation"
    description: "Implement core functionality"
    prompt: "prompts/implementation.md"
    variables:
      phase: "implementation"
      focus: "core business logic"

  - id: "04-implementation-testing"
    description: "Test the implementation"
    prompt: "prompts/testing.md"
    variables:
      phase: "implementation"
      focus: "core business logic"
      coverage_goal: "85%%"

  - id: "05-extension"
    description: "Extend functionality"
    prompt: "prompts/extension.md"
    variables:
      phase: "extension"
      focus: "additional features"

  - id: "06-extension-testing"
    description: "Test the extensions"
    prompt: "prompts/testing.md"
    variables:
      phase: "extension"
      focus: "additional features"
      coverage_goal: "90%%"

  - id: "07-refinement"
    description: "Refine the implementation"
    prompt: "prompts/refinement.md"
    variables:
      phase: "refinement"
      focus: "code quality and performance"

  - id: "08-documentation"
    description: "Document the implementation"
    prompt: "prompts/documentation.md"
    variables:
      scope: "complete implementation"
      target_audience: "developers"
`, workflowName)
	
	// Add foundation.md
	template.Files[filepath.Join(workflow.PromptsDir, "foundation.md")] = `# {{ .phase | title }} Phase

This step focuses on laying the foundation for the implementation:

{{ template "shared/phase_header.md" . }}

## Primary Tasks

1. Set up the project structure
2. Define key interfaces and data structures
3. Establish the overall architecture
4. Create scaffolding for the main components

## Expected Outcome

By the end of this phase, we should have a solid architectural foundation with:

- Well-defined interfaces
- Clear separation of concerns
- Core data models defined
- Initial project structure in place

{{ template "shared/footer.md" . }}
`
	
	// Add implementation.md
	template.Files[filepath.Join(workflow.PromptsDir, "implementation.md")] = `# {{ .phase | title }} Phase

This step focuses on implementing the core functionality:

{{ template "shared/phase_header.md" . }}

## Primary Tasks

1. Implement the core business logic
2. Create the primary components
3. Connect components through defined interfaces
4. Implement error handling and validation

## Expected Outcome

By the end of this phase, we should have:

- Functioning core features
- Proper error handling
- Validation of inputs
- Initial tests for critical paths

{{ template "shared/footer.md" . }}
`
	
	// Add extension.md
	template.Files[filepath.Join(workflow.PromptsDir, "extension.md")] = `# {{ .phase | title }} Phase

This step focuses on extending the functionality beyond the core:

{{ template "shared/phase_header.md" . }}

## Primary Tasks

1. Implement additional features
2. Add support for edge cases
3. Enhance error handling
4. Add convenience methods and utilities

## Expected Outcome

By the end of this phase, we should have:

- Complete feature set implemented
- Robust handling of edge cases
- Comprehensive error handling
- Enhanced usability with convenience methods

{{ template "shared/footer.md" . }}
`
	
	// Add refinement.md
	template.Files[filepath.Join(workflow.PromptsDir, "refinement.md")] = `# {{ .phase | title }} Phase

This step focuses on refining the implementation:

{{ template "shared/phase_header.md" . }}

## Primary Tasks

1. Optimize performance
2. Improve code quality
3. Reduce technical debt
4. Address any remaining issues

## Expected Outcome

By the end of this phase, we should have:

- Optimized performance
- Clean, maintainable code
- Reduced technical debt
- Resolved all known issues

{{ template "shared/footer.md" . }}
`
	
	// Add testing.md (reused across multiple steps)
	template.Files[filepath.Join(workflow.PromptsDir, "testing.md")] = `# Testing - {{ .phase | title }} Phase

This step focuses on testing the {{ .phase }} phase:

{{ template "shared/phase_header.md" . }}

## Testing Focus

During this testing phase, we should focus on:

1. Unit testing individual components
2. Integration testing between components
3. Ensuring good test coverage (target: {{ .coverage_goal }})
4. Verifying edge cases and error handling

## Expected Outcome

By the end of this testing phase, we should have:

- Comprehensive test suite for the {{ .phase }} phase
- Test coverage meeting or exceeding {{ .coverage_goal }}
- Documentation of test cases and approaches
- Validation that the implementation meets requirements

{{ template "shared/footer.md" . }}
`
	
	// Add documentation.md
	template.Files[filepath.Join(workflow.PromptsDir, "documentation.md")] = `# Documentation Phase

This step focuses on documenting the implementation:

## Documentation Scope

We're documenting the {{ .scope }} with a focus on {{ .target_audience }}.

## Primary Tasks

1. Document the public API
2. Create usage examples
3. Document internal architecture
4. Add inline code documentation

## Expected Outcome

By the end of this documentation phase, we should have:

- Comprehensive API documentation
- Clear usage examples
- Architecture documentation
- Well-commented code

{{ template "shared/footer.md" . }}
`
	
	// Add shared templates
	template.Files[filepath.Join(workflow.PromptsDir, workflow.SharedDir, "phase_header.md")] = `## Focus Area: {{ .focus }}

This phase we're focusing on {{ .focus }}.

---`
	
	template.Files[filepath.Join(workflow.PromptsDir, workflow.SharedDir, "footer.md")] = `---

## Notes and Considerations

- Keep code modular and testable
- Follow project coding standards
- Document key decisions and trade-offs
- Consider backwards compatibility`
	
	// Add README.md
	template.Files["README.md"] = fmt.Sprintf(`# %s Workflow

A comprehensive workflow with extended functionality.

## Directory Structure

- workflow.yaml: Workflow configuration
- prompts/: Prompt files
  - foundation.md: Foundation phase prompt
  - implementation.md: Implementation phase prompt
  - extension.md: Extension phase prompt
  - refinement.md: Refinement phase prompt
  - testing.md: Testing phase prompt (reused with different variables)
  - documentation.md: Documentation phase prompt
  - shared/: Shared template fragments
    - phase_header.md: Common header for phase prompts
    - footer.md: Common footer for all prompts

## Features

This workflow demonstrates several advanced features:

1. Reuse of prompt files across multiple steps (testing.md)
2. Shared template fragments (in shared/ directory)
3. Variable substitution with different values per step
4. Template inclusion using Go template syntax

## Usage

To use this workflow:

` + "```bash" + `
usm code --workflow=%s path/to/change-request.blueprint.md
` + "```" + `

## Customization

You can customize this workflow by:

1. Editing workflow.yaml to change steps or add new ones
2. Modifying prompt files in the prompts/ directory
3. Adding or modifying shared templates in the prompts/shared/ directory
4. Customizing variables for each step
`, workflowName, workflowName)
	
	return template
}

func init() {
	// Add flags
	InitCmd.Flags().BoolP("global", "g", false, "Create workflow in global directory (~/.usm/workflows)")
	InitCmd.Flags().StringP("template", "t", defaultTemplate, "Template to use (default, full, blank)")
} 
---
name: custom workflow cli
created-at: 2025-04-15T05:52:15+02:00
user-stories:
  - title: Define custom workflow with directory structure
    file: docs/user-stories/custom-workflow/core-01-define-custom-workflow-structure.md
    content-hash: c1b0a50f0b6966decd8a8368c15199d4fe14af628363d5775846d3b9a29fcf38
  - title: Reuse prompt files with variable substitution
    file: docs/user-stories/custom-workflow/core-02-reuse-prompt-files-with-variables.md
    content-hash: 9fc75b8dc5135b2b0c9cec2ca1c14723d5a5a68a229cbb1f7b3532ff7d98a404
  - title: Validate workflow definitions
    file: docs/user-stories/custom-workflow/core-04-validate-workflow-definitions.md
    content-hash: 65f8d70ec33a7f9616a53eb5dd393b8423cc1f348094b9daea8892ac8d57bc5d
  - title: Select workflow for code command
    file: docs/user-stories/custom-workflow/ui-01-select-workflow-for-code-command.md
    content-hash: 5fd7e3bd6ed6dffd63ec796fcb0905cf9fca1e7ebb0aaeaeb142103ba8d782b5
  - title: Specify custom workflow by path
    file: docs/user-stories/custom-workflow/ui-02-specify-custom-workflow-by-path.md
    content-hash: f1ae01f2a8341a4dcb836ccbe2e5778a510ef37230ccd10db783ccfa3414d58e
  - title: List available workflows
    file: docs/user-stories/custom-workflow/ui-03-list-available-workflows.md
    content-hash: 58c632c637aa8a2718a1325afdcf9ed3d83e02ad5441192c8ecac08239331b2f
  - title: Initialize new workflow
    file: docs/user-stories/custom-workflow/ui-04-initialize-new-workflow.md
    content-hash: 2f5cd322093a9ea6b351504df172963f87ba0ee808b2a1b6bed02f6d75d8bf3e
  - title: Workflow documentation command
    file: docs/user-stories/custom-workflow/ui-05-workflow-documentation-command.md
    content-hash: c33a188e106e98c1b81c768278eb8edd66eff03cd8a14d9462485d557df60a29
  - title: Create workflow subcommand
    file: docs/user-stories/custom-workflow/ui-07-create-workflow-subcommand.md
    content-hash: 8c3d36fdb18e259c88e9e46547000ad591dca8aa45e180e907d3712accc380dc

---

# Blueprint

## Overview
This blueprint outlines the implementation of a custom workflow system for the USM tool. The feature will allow users to define, validate, and manage custom workflows using a directory-based structure with YAML configuration and separate prompt files. Users will be able to select workflows for the `code` command, initialize new workflows from templates, and manage workflows through a dedicated subcommand interface. The implementation focuses on maintainability, flexibility, and user experience.

## Fundamentals

### Data Structures

1. **Directory-based Workflow Structure**
   - A workflow is defined by a directory containing a `workflow.yaml` file and a `prompts/` subdirectory
   - The `workflow.yaml` file contains workflow metadata and step definitions
   - The `prompts/` directory contains individual markdown files for each prompt
   - Optional `shared/` subdirectory for reusable prompt templates

2. **Workflow Configuration**
   ```yaml
   name: "workflow-name"
   description: "Description of the workflow"
   steps:
     - id: "01-step-one"
       description: "First step description"
       prompt: "prompts/step1.md"
       variables:
         key1: "value1"
         key2: "value2"
   ```

3. **Workflow Registry Enhancement**
   - Extend the existing `WorkflowRegistry` to support directory-based workflows
   - Implement workflow discovery in standard locations:
     1. `.usm/workflows/` in the current directory (project-specific)
     2. `~/.usm/workflows/` in the user's home directory (user-specific)
     3. Built-in workflows provided by the application

4. **Prompt Template System**
   - Support Go template syntax for variable substitution (`{{.variable_name}}`)
   - Allow default values: `{{.variable_name | default "default value"}}`
   - Support structured data (arrays, maps) in variables

5. **Validation System**
   - Validate workflow definitions and prompt files
   - Check for existence of referenced prompt files
   - Verify template syntax and variable references

### Existing Components to Leverage

1. **Workflow Registry**
   - `internal/workflow/registry.go` already contains a robust registry system that can be enhanced:
     - `LoadFromDirectory` method loads a workflow from a directory
     - `DiscoverWorkflows` finds workflows in standard locations
     - `RegisterBuiltInWorkflow` registers workflows with the registry

2. **Workflow Loading**
   - `internal/workflow/loader.go` has utilities for loading and saving workflows:
     - `LoadWorkflowFromFile` loads a workflow from YAML or JSON
     - `SaveWorkflowToFile` saves a workflow to disk
     - `ExternalWorkflowDefinition` and `ExternalWorkflowStep` structures for serialization

3. **Workflow Extraction**
   - `internal/workflow/extract.go` contains functionality to extract the standard workflow:
     - `ExtractStandardWorkflow` creates a directory structure with the standard workflow
     - Can be adapted for the `workflow init` command implementation

4. **Template Processing**
   - `internal/workflow/template.go` has a template system that can be used:
     - `TemplateProcessor` handles template rendering
     - `ApplyTemplateVariables` processes templates with variables

5. **Command Implementation**
   - `cmd/extract_workflow.go` provides a pattern for the new workflow commands:
     - Similar command structure can be used for `workflow` subcommands
     - Demonstrates proper file system and IO handling

### Algorithms

1. **Workflow Discovery Algorithm**
   ```
   function DiscoverWorkflows(fs FileSystem) -> map[string]*WorkflowDefinition:
     result = empty map of workflow name to definition
     
     for each directory in GetStandardWorkflowDirectories():
       if directory exists:
         for each subdirectory in directory:
           if subdirectory contains workflow.yaml:
             try to load workflow from subdirectory
             if successful:
               add workflow to result map
     
     return result
   ```
   - *Can leverage existing `DiscoverWorkflows` method in `internal/workflow/registry.go`*

2. **Directory-based Workflow Loading**
   ```
   function LoadWorkflowFromDirectory(fs FileSystem, path string) -> (*WorkflowDefinition, error):
     // Check if workflow.yaml exists
     workflowYAMLPath = path + "/workflow.yaml"
     if !fs.Exists(workflowYAMLPath):
       return nil, error("workflow.yaml not found")
     
     // Load workflow configuration
     externalWorkflow = parse YAML from workflowYAMLPath
     
     // Validate workflow
     errors = ValidateWorkflowPromptReferences(fs, path, externalWorkflow)
     if errors not empty:
       return nil, concatenated errors
     
     // Transform external workflow to internal format
     return externalWorkflow.ToWorkflowDefinition(), nil
   ```
   - *Can extend existing `LoadFromDirectory` method in `internal/workflow/registry.go`*

3. **Prompt Template Rendering**
   ```
   function RenderPromptTemplate(template string, variables map[string]string) -> (string, error):
     templateEngine = new Go template engine
     parse template with templateEngine
     if parsing error:
       return "", parsing error
     
     // Execute template with variables
     buffer = new string buffer
     execute template with variables into buffer
     if execution error:
       return "", execution error
     
     return buffer.String(), nil
   ```
   - *Can use existing `ApplyTemplateVariables` in `internal/workflow/template.go`*

4. **Workflow Validation**
   ```
   function ValidateWorkflow(fs FileSystem, path string) -> []error:
     errors = empty error list
     
     // Check if workflow.yaml exists
     workflowYAMLPath = path + "/workflow.yaml"
     if !fs.Exists(workflowYAMLPath):
       add "workflow.yaml not found" to errors
       return errors
     
     // Load and parse workflow
     externalWorkflow = parse YAML from workflowYAMLPath
     
     // Check required fields
     if externalWorkflow.Name is empty:
       add "workflow name is required" to errors
     if externalWorkflow.Description is empty:
       add "workflow description is required" to errors
     if externalWorkflow.Steps is empty:
       add "workflow must have at least one step" to errors
     
     // Validate steps
     stepIDs = empty set
     for each step in externalWorkflow.Steps:
       if step.ID is empty:
         add "step ID is required" to errors
       if step.ID is in stepIDs:
         add "duplicate step ID" to errors
       add step.ID to stepIDs
       
       if step.Description is empty:
         add "step description is required" to errors
       
       if step.Prompt is empty:
         add "step prompt is required" to errors
       else:
         // Check if prompt file exists
         promptPath = path + "/" + step.Prompt
         if !fs.Exists(promptPath):
           add "prompt file not found: " + step.Prompt to errors
         else:
           // Validate prompt template
           promptContent = read file at promptPath
           _, templateError = RenderPromptTemplate(promptContent, step.Variables)
           if templateError != nil:
             add "invalid template in " + step.Prompt + ": " + templateError.Error() to errors
     
     return errors
   ```
   - *Can adapt existing validation from `validateExternalWorkflow` in `internal/workflow/loader.go`*

### Refactoring Strategy

1. **Command Layer**
   - Add a new `workflow` root command with subcommands:
     - `workflow list`: List available workflows
     - `workflow init`: Initialize new workflow
     - `workflow validate`: Validate workflow definition
     - `workflow show`: Display workflow details
   - Modify the `code` command to accept `--workflow` and `--workflow-path` flags
   - *Pattern similar to existing `extract_workflow.go` command*

2. **Workflow Layer**
   - Enhance `WorkflowRegistry` to discover and load directory-based workflows
   - Implement workflow validation and template rendering
   - Add support for workflow templates
   - *Extend existing registry and loading functionality*

3. **UI Layer**
   - Implement list display for available workflows
   - Create initialization wizard for new workflows
   - Add validation output formatting
   - *Use existing IO interfaces for consistent user experience*

### How to verify – Detailed User Story Breakdown

#### User Story: Define custom workflow with directory structure

**Acceptance Criteria and Testing Scenarios:**

1. **Directory structure support**
   - Create a test workflow directory following the convention
   - Run `usm workflow validate path/to/workflow`
   - Verify it reports success for a valid workflow

2. **workflow.yaml format**
   - Create workflow.yaml with all required fields
   - Create workflow.yaml missing required fields
   - Validate both and verify appropriate success/error responses

3. **Prompt file resolution**
   - Create workflow with valid prompt references
   - Create workflow with missing prompt references
   - Validate both and verify appropriate success/error responses

#### User Story: Reuse prompt files with variable substitution

**Acceptance Criteria and Testing Scenarios:**

1. **Variable substitution in prompts**
   - Create a prompt with `{{.variable_name}}` placeholders
   - Create a workflow step referencing this prompt with variables
   - Run the workflow and verify variables are substituted correctly

2. **Default values**
   - Create a prompt with `{{.variable_name | default "default value"}}`
   - Run with and without providing the variable
   - Verify default value is used when variable is not provided

3. **Structured data support**
   - Create a prompt using complex variables (arrays, maps)
   - Create a workflow step with structured variable values
   - Verify correct rendering of the template

#### User Story: Validate workflow definitions

**Acceptance Criteria and Testing Scenarios:**

1. **Workflow validate command**
   - Verify `usm workflow validate [name or path]` functionality
   - Test with valid and invalid workflows
   - Check exit codes match expected values (0 for valid, 1 for invalid)

2. **Comprehensive validation**
   - Test validation of all required fields
   - Test uniqueness of step IDs
   - Test existence of prompt files
   - Test template syntax validation
   - Test variable reference validation

3. **Error message quality**
   - Verify error messages include file paths
   - Verify error messages include line numbers when possible
   - Verify error messages include suggested fixes

#### User Story: Select workflow for code command

**Acceptance Criteria and Testing Scenarios:**

1. **Workflow selection flag**
   - Run `usm code --workflow=my-custom-workflow path/to/blueprint.md`
   - Verify the specified workflow is selected and used

2. **Default workflow fallback**
   - Run `usm code path/to/blueprint.md` without workflow flag
   - Verify standard workflow is used

3. **Invalid workflow handling**
   - Run with non-existent workflow name
   - Verify appropriate error message

4. **Workflow search path order**
   - Create workflows with same name in different locations
   - Verify correct precedence (.usm/workflows/ > ~/.usm/workflows/ > built-in)

5. **State persistence**
   - Run code command with workflow selection
   - Run again without selection
   - Verify same workflow is used

#### User Story: Specify custom workflow by path

**Acceptance Criteria and Testing Scenarios:**

1. **Path-based selection**
   - Run `usm code --workflow-path=/path/to/my-workflow path/to/blueprint.md`
   - Verify workflow at specified path is used

2. **Validation of specified path**
   - Test with valid and invalid paths
   - Verify appropriate error messages

3. **Flag precedence**
   - Run with both `--workflow` and `--workflow-path` flags
   - Verify `--workflow-path` takes precedence

4. **Relative path resolution**
   - Run with relative path
   - Verify path is resolved correctly relative to current directory

5. **State persistence**
   - Run code command with workflow path
   - Run again without path specification
   - Verify same workflow is used

#### User Story: List available workflows

**Acceptance Criteria and Testing Scenarios:**

1. **List command functionality**
   - Run `usm workflow list`
   - Verify all available workflows are displayed
   - Verify format includes name, description, source, and path

2. **Source classification**
   - Create workflows in all three locations (built-in, user, project)
   - Verify each is correctly labeled with its source

3. **Output formatting**
   - Test different output formats (text, json)
   - Verify correct formatting in each case

#### User Story: Initialize new workflow

**Acceptance Criteria and Testing Scenarios:**

1. **Initialization command**
   - Run `usm workflow init my-workflow`
   - Verify directory structure is created correctly with a sample workflow
   - Verify the sample workflow contains:
     - A properly formatted workflow.yaml file with basic metadata
     - A prompts directory with at least one sample prompt file
     - Comments explaining how to customize the workflow

2. **Template support**
   - Templates can be specified with `--template` flag but are optional
   - When no template is specified, a minimal sample workflow is created
   - Available templates:
     - `--template=full`: A comprehensive workflow with more extensive examples
     - `--template=blank`: A skeleton structure with minimal content

3. **Global flag**
   - Run `usm workflow init my-workflow --global`
   - Verify workflow is created in ~/.usm/workflows/

4. **Error handling**
   - Try to initialize a workflow that already exists
   - Try to use non-existent template
   - Verify appropriate error messages

#### User Story: Workflow documentation command

**Acceptance Criteria and Testing Scenarios:**

1. **Show command functionality**
   - Run `usm workflow show my-workflow`
   - Verify correct display of workflow name, description, and steps

2. **Output formats**
   - Test different output formats (text, markdown, json)
   - Verify correct formatting in each case

3. **Error handling**
   - Run with non-existent workflow name
   - Verify appropriate error message

#### User Story: Create workflow subcommand

**Acceptance Criteria and Testing Scenarios:**

1. **Command structure**
   - Verify `usm workflow` command exists
   - Verify all subcommands are implemented
   - Test help output for clarity and completeness

2. **List command integration**
   - Verify it discovers workflows from all sources
   - Verify it displays them in the expected format

3. **Init command integration**
   - Verify it creates workflow with correct structure
   - Verify it supports templates and global flag

4. **Validate command integration**
   - Verify it performs comprehensive validation
   - Verify it provides useful error messages

5. **Show command integration**
   - Verify it displays workflow details correctly
   - Verify it supports different output formats

### What is the Plan – Detailed Action Items

#### Implementation Plan for User Story: Define custom workflow with directory structure

1. **Enhance WorkflowRegistry**
   - Extend existing `DiscoverWorkflows` method in `registry.go` to support directory-based workflows
   - Update the `LoadFromDirectory` method to handle prompt file references
   - Add caching mechanism for file-based workflows (already in place, need to extend)

2. **Create Directory-based Loader**
   - Enhance existing `LoadWorkflowFromFile` function in `loader.go` to support directory-based workflows
   - Add validation of directory structure
   - Reuse existing file parsing logic from `loader.go`

3. **Update Workflow Execution**
   - Modify workflow execution to support file-based prompt loading (already partially supported)
   - Ensure backward compatibility with embedded prompts

#### Implementation Plan for User Story: Reuse prompt files with variable substitution

1. **Implement Template Engine**
   - Enhance existing `TemplateProcessor` in `template.go`
   - Add support for variable substitution using Go templates
   - Implement default value functionality using existing template functions

2. **Enhance Step Execution**
   - Modify existing `StepExecutor` in `executor.go` to handle variable substitution
   - Add validation of variable references using existing validation logic
   - Implement variable escaping for security

#### Implementation Plan for User Story: Validate workflow definitions

1. **Create Validation System**
   - Extend existing validation in `loader.go` and `extract.go`
   - Implement comprehensive validation for workflow structure and prompt templates
   - Create a unified validation interface for different validation types

2. **Implement Validation Command**
   - Create new command similar to `extract_workflow.go` for validation
   - Reuse validation logic from enhanced validation system
   - Format and display validation errors

#### Implementation Plan for User Story: Select workflow for code command

1. **Update Code Command**
   - Modify existing `code.go` to accept `--workflow` flag
   - Update `NewWorkflowManager` call to pass the workflow name
   - Leverage existing workflow selection logic in `workflow.go`

2. **Implement Workflow Search**
   - Use existing `GetStandardWorkflowDirectories` in `registry.go`
   - Reuse `DiscoverWorkflows` for workflow discovery
   - Implement precedence rules using existing registry logic

3. **Update State Management**
   - Extend existing `WorkflowState` in `workflow.go` to include workflow selection
   - Update state persistence in `SaveState` method
   - Ensure backward compatibility with existing state files

#### Implementation Plan for User Story: Specify custom workflow by path

1. **Update Code Command**
   - Add `--workflow-path` flag to the code command
   - Enhance `NewWorkflowManager` to accept a workflow path
   - Use existing file system interfaces for path validation

2. **Handle Precedence**
   - Implement logic to prioritize `--workflow-path` over `--workflow`
   - Leverage existing workflow loading from `registry.go`
   - Update state to store workflow path

#### Implementation Plan for User Story: List available workflows

1. **Create List Command**
   - Implement new command similar to `extract_workflow.go`
   - Use existing `ListWorkflows` method from `registry.go`
   - Create table display for workflows using terminal IO

2. **Implement Format Options**
   - Add format flag similar to other commands
   - Implement formatters for each output type
   - Reuse existing IO interfaces for output formatting

#### Implementation Plan for User Story: Initialize new workflow

1. **Create Init Command**
   - Implement new command based on existing `extract_workflow.go`
   - Leverage `ExtractStandardWorkflow` as a pattern for creating workflows
   - Add support for `--global` flag using existing path resolution logic
   - Create a default sample workflow that demonstrates the key concepts:
     - Variable substitution
     - Prompt file organization 
     - Basic workflow structure

2. **Implement Optional Templates**
   - Make templates optional with sensible defaults
   - Default to a minimal but educational sample workflow
   - Implement alternative templates:
     - `full`: More comprehensive with additional examples
     - `blank`: Minimal skeleton for advanced users

3. **Handle Edge Cases**
   - Use existing file system checks for workflow existence
   - Implement error handling for template selection
   - Add user guidance for next steps including helpful comments in the generated files

#### Implementation Plan for User Story: Workflow documentation command

1. **Create Show Command**
   - Implement new command similar to `extract_workflow.go`
   - Use existing `GetWorkflow` method from `registry.go`
   - Create workflow detail display using terminal IO

2. **Implement Format Options**
   - Add format flag similar to other commands
   - Implement formatters for each output type
   - Reuse existing IO interfaces for output formatting

#### Implementation Plan for User Story: Create workflow subcommand

1. **Create Parent Command**
   - Implement root `workflow` command similar to other root commands
   - Use existing command registration pattern
   - Add help documentation

2. **Implement Subcommands**
   - Create all required subcommands following existing patterns
   - Add command-specific flags and options
   - Implement help documentation for each

3. **Integration Testing**
   - Follow existing testing patterns in `extract_workflow_test.go`
   - Verify command interactions
   - Ensure consistent CLI experience

### Detailed Implementation Steps

1. **Create New Files:**
   - `cmd/workflow.go`: Root workflow command
   - `cmd/workflow/list.go`: List command implementation
   - `cmd/workflow/init.go`: Init command implementation
   - `cmd/workflow/validate.go`: Validate command implementation
   - `cmd/workflow/show.go`: Show command implementation
   - `internal/workflow/directoryloader.go`: Directory-based workflow loading
   - `internal/workflow/template.go`: Template-based workflow initialization
   - `internal/workflow/validator.go`: Comprehensive workflow validation
   - `internal/workflow/template_renderer.go`: Prompt template rendering

2. **Modify Existing Files:**
   - `cmd/code.go`: Add workflow selection flags
   - `internal/workflow/workflow.go`: Add support for directory-based workflows
   - `internal/workflow/registry.go`: Enhance workflow discovery
   - `internal/workflow/state.go`: Update state to include workflow selection

3. **Add Workflow Templates:**
   - Create default sample workflow template with educational comments
   - Create full template with more comprehensive examples
   - Create blank template with minimal structure

4. **Update Integration Tests:**
   - Add tests for new commands
   - Add tests for workflow selection
   - Add tests for template-based initialization

5. **Update Documentation:**
   - Add workflow command documentation
   - Update code command documentation to include workflow flags
   - Create examples for custom workflow creation

### Documentation Strategy

The implementation of custom workflows requires comprehensive documentation to ensure users can effectively utilize the new functionality. The documentation strategy will consist of several components targeting different user needs:

1. **Command Reference Documentation**
   - Create detailed documentation for each new command:
     - `usm workflow` - Overview of workflow management
     - `usm workflow list` - How to list available workflows
     - `usm workflow init` - How to create new workflows
     - `usm workflow validate` - How to validate workflow definitions
     - `usm workflow show` - How to display workflow details
   - Update existing command documentation:
     - `usm code` - Add details for `--workflow` and `--workflow-path` flags
     - `usm extract-workflow` - Clarify relationship with new workflow commands

2. **Conceptual Documentation**
   - Create a "Custom Workflows Guide" explaining:
     - The workflow directory structure
     - YAML configuration format
     - Prompt file organization
     - Variable substitution
     - Template syntax
   - Add a "Workflow Best Practices" guide covering:
     - Organizing complex workflows
     - Reusing prompt fragments
     - Managing workflow versions
     - Testing workflows before deployment

3. **Tutorials and Examples**
   - Create step-by-step tutorials:
     - "Creating Your First Custom Workflow"
     - "Converting an Existing Process to a Workflow"
     - "Using Variables for Flexible Workflows"
   - Provide annotated example workflows:
     - Minimal example with basic functionality
     - Complex example demonstrating advanced features
     - Multi-team workflow with shared components

4. **Documentation Location and Format**
   - CLI help text: Concise documentation accessible via `--help`
   - In-repo documentation: README files and examples in the codebase

This documentation strategy ensures that users at different levels of expertise can effectively use, customize, and extend the custom workflow functionality.

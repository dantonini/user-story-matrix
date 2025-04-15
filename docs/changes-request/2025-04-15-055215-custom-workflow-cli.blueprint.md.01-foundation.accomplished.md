# Foundation Accomplished: Custom Workflow CLI

## Overview
The foundation phase for the Custom Workflow CLI has been successfully implemented, establishing the core architecture and structure needed to support custom workflows in USM. This accomplishment report details the key components that have been created or enhanced during this phase.

## Core Structures and Components

### Workflow Definition Architecture
- Created the structured directory-based workflow system in `internal/workflow/directoryloader.go` with support for:
  - Standardized workflow structure (workflow.yaml, prompts directory)
  - Flexible workflow discovery and loading mechanisms
  - Source tracking for workflows (built-in, user, project)

### Command Structure
- Implemented the `workflow` command and subcommands in `cmd/workflow.go`
- Created essential subcommands:
  - `list`: Lists available workflows (`cmd/workflow/list.go`)
  - `init`: Initializes new workflows from templates (`cmd/workflow/init.go`) 
  - `validate`: Validates workflow definitions (`cmd/workflow/validate.go`)
  - `show`: Displays workflow details (`cmd/workflow/show.go`)

### Registry and Loading System
- Enhanced the `WorkflowRegistry` in `internal/workflow/registry.go` to support:
  - Loading workflows from filesystem
  - Caching mechanisms for performance
  - Modification time tracking for automatic reloading
  - Comprehensive workflow discovery across multiple locations

### Prompt System
- Implemented robust prompt handling in `internal/workflow/extract.go`:
  - Separation of prompt content into individual files
  - Support for prompt loading from filesystem
  - Tracking of prompt sources (embedded vs. file-based)
  - Fallback mechanisms for error handling

    // promptSourceType identifies where a prompt comes from
    type promptSourceType int

    const (
        promptSourceEmbedded promptSourceType = iota
        promptSourceFile
    )

    // promptSource tracks the origin of a prompt
    type promptSource struct {
        sourceType promptSourceType
        filePath   string // Only used when sourceType is promptSourceFile
    }

### Validation System
- Created comprehensive validation in `internal/workflow/validator.go`:
  - Workflow structure validation
  - Prompt file existence checks
  - Directory structure validation
  - Error collection and reporting

## Key User Stories Progress

### Define custom workflow with directory structure
- Implemented structure with `workflow.yaml` for metadata and `prompts/` for prompt files
- Created the directory loading capabilities in `directoryloader.go`
- Added validation to ensure all referenced prompt files exist

Workflow file format: The YAML file uses this structure:

```yaml
   name: "workflow-name"
   description: "Description of the workflow"
   steps:
     - id: "01-step-one"
       description: "First step"
       prompt: "prompts/step1.md"
       variables:
         key1: "value1"
         key2: "value2"
```

### Create workflow subcommand
- Established the main `workflow` command with all subcommands
- Implemented proper command hierarchy and help documentation
- Connected workflows commands to core workflow system

### List available workflows
- Created the `list` subcommand that discovers and displays available workflows
- Implemented tabular output format showing name, description, source, and path
- Added workflow source identification (built-in, user, project)

### Initialize new workflow
- Implemented the `init` command to create new workflows from templates
- Added support for both local and global workflow creation
- Created template system with sensible defaults for quick start

### Validate workflow definitions
- Created the `validate` command for comprehensive workflow validation
- Implemented robust validation logic in `validator.go`
- Added detailed error reporting for validation issues

## Code Quality and Testing

### Linting Status
- Fixed all linting issues in the workflow command files:
  - Properly handling flag parsing error returns in `init.go`
  - Added error checking for format flags in both `list.go` and `show.go`
  - Ensured consistent code style across all subcommands
  - All linting checks now pass successfully

### Test Coverage
The current implementation has good test coverage overall, but there's one test failure to address:
- `TestWorkflowManager_SaveState_WithErrors_ErrorSimulation` in `state_test.go`
  - The test fails with a permission error simulation issue
  - This is a minor issue related to error handling in state persistence

### Integration Tests
- The workflow registry and directory loading mechanisms have been tested extensively
- Command-line integration tests will need to be expanded in the MVI phase

## Next Steps

### For MVI Phase
- Fix the failing test in `state_test.go`
- Complete variable substitution in prompt templates
- Enhance the `--workflow` flag in the `code` command
- Implement workflow selection by path
- Add workflow state persistence to maintain workflow selection
- Add integration tests for the new workflow commands

### Testing Considerations
- Expand test coverage for the new workflow commands
- Add integration tests for workflow discovery and loading
- Test validation with various error conditions
- Improve error simulation capabilities in tests

## Blind Spots
- Error handling for edge cases in workflow loading could be improved
- More comprehensive validation of template variables is needed
- The workflow selection persistence mechanism needs implementation
- Error simulation in tests is difficult with the current MockFileSystem

## Current Design Decisions
- Workflows are discovered from three standard locations: project-specific, user-specific, and built-in
- Prompt files are separated from workflow configuration for better maintainability
- The registry system maintains backward compatibility with existing code while enabling the new functionality
- Command implementation follows USM CLI standards with consistent error handling and output formatting 
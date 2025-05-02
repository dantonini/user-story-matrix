# Workflow Package

This package implements the custom workflow functionality for the USM tool.

## Overview

The workflow package provides the core functionality for:
- Loading and parsing workflow definitions from directories
- Validating workflow structure and prompt templates
- Template rendering with variable substitution
- Workflow state management
- Workflow switching and state migration

## Key Components

- **WorkflowRegistry**: Central registry for all available workflows
- **WorkflowManager**: Manages workflow execution and state
- **TemplateRenderer**: Renders prompt templates with variable substitution
- **WorkflowValidator**: Validates workflow definitions and prompt files
- **DirectoryLoader**: Loads workflows from a directory structure

## Custom Workflow Directory Structure

A custom workflow follows this structure:
```
workflow-name/
├── workflow.yaml      # Core workflow definition
└── prompts/           # Directory for prompt files
    ├── step1.md
    ├── step2.md
    └── shared/        # Optional directory for reusable prompts
```

The `workflow.yaml` defines the workflow metadata and steps:
```yaml
name: "workflow-name"
description: "Description of the workflow"
steps:
  - id: "01-step-one"
    description: "First step description"
    prompt: "prompts/step1.md"
    variables:
      name: "User Story Matrix"
      version: "1.0.0"
```

## TODO for Extend Phase

The following items need to be addressed in the extend phase:

### Validation Improvements
- [ ] Enhance `checkForDefaultValue` in validator.go to properly detect default values with more robust parsing
- [ ] Add more comprehensive validation for template syntax
- [ ] Improve variable extraction reliability by using a more robust parser

### User Experience
- [ ] Complete the workflow list command with better formatting
- [ ] Finalize the workflow init command implementation
- [ ] Implement comprehensive workflow documentation command

### Template System
- [ ] Add support for shared template fragments
- [ ] Implement conditional steps in workflows
- [ ] Add support for nested variable structures

### Error Handling
- [ ] Improve error messages with more specific guidance
- [ ] Add recovery mechanisms for common errors
- [ ] Enhance template error reporting with line numbers

### Testing
- [ ] Add more edge case tests for template rendering
- [ ] Enhance error simulation capabilities in tests
- [ ] Add integration tests for command-line interface

## Integration Points

- The custom workflow functionality integrates with the `code` command via the `--workflow` and `--workflow-path` flags
- The workflow state is persisted in `.step` files alongside change request files
- The registry discovers workflows from project, user, and built-in locations 
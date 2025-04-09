# Standard Workflow Template

This directory contains the standard workflow template used by the USM (User Story Matrix) tool. The standard workflow provides a structured approach to software development broken down into 4 phases.

## Directory Structure

```
standard/
├── README.md             # This file
├── workflow.yaml         # The workflow definition file
└── prompts/              # Directory containing prompt files
    ├── README.md         # Documentation for prompt files
    ├── 01-foundation.md  # Prompt for laying the foundation phase
    ├── 02-mvi.md         # Prompt for Minimum Viable Implementation phase
    ├── 03-extend.md      # Prompt for extending functionalities phase
    └── 04-refine.md      # Prompt for refinement phase
```

## Usage

This template is used by the following commands:

1. `usm extract-workflow` - Extracts this template to a specified directory
2. `usm workflow init` - Initializes a new workflow based on this template

## Customization

You can customize this template by:

1. Editing the workflow.yaml file to modify the steps or their descriptions
2. Editing the prompt files in the prompts/ directory to change the instructions for each step
3. Adding new prompt files and referencing them in the workflow.yaml file

## File Format

The workflow.yaml file follows this structure:

```yaml
name: "standard"
description: "Standard 4-phase development workflow"
steps:
  - id: "01-laying-the-foundation"
    description: "Laying the foundation - Setting up the architecture and structure"
    prompt: "prompts/01-foundation.md"
  # Additional steps...
```

Each prompt file contains the instructions that will be shown to the user when they reach that step in the workflow. 
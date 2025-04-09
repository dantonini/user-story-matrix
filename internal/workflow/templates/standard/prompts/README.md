# Workflow Prompt Files

This directory contains prompt files used by the standard workflow. Each prompt file contains the instructions that will be shown to the user when executing a workflow step.

## Naming Convention

Prompt files follow this naming convention:

```
{step-id}.md
```

For example, the prompt for step "01-laying-the-foundation" would be in a file named "01-foundation.md".

## Format

Prompt files are Markdown files containing the instructions for each step. They can include placeholders for variables that will be replaced at runtime:

```markdown
You are a senior software engineer implementing a change request.

Your current task is to ${step_description}.

The change request is located at: ${change_request_file_path}

...
```

## Available Variables

The following variables can be used in prompts:

- `${change_request_file_path}` - The absolute path to the change request file
- `${change_request_dirname}` - The directory containing the change request file
- `${change_request_basename}` - The filename of the change request file (without path)
- `${workflow_name}` - The name of the current workflow
- `${step_description}` - The description of the current step
- `${step_id}` - The ID of the current step

## Customization

You can customize these prompt files to:

1. Change the instructions for each step
2. Add more context or specific requirements
3. Reference different tools or approaches

When modifying prompts, make sure to maintain the same variables as they are used by the workflow engine.

## Purpose

Each file in this directory contains the instructions that will be presented to the AI agent for a specific workflow step. These prompts guide the agent through the implementation process, providing context, requirements, and expectations for each phase.

## File Naming Convention

Prompt files follow this naming convention:

```
{step-id}.md
```

For example:
- `01-laying-the-foundation.md`
- `02-mvi.md`
- `03-extend-functionalities.md`

## Format

Prompt files are written in Markdown format and may contain template variables that will be resolved at runtime. For example:

```markdown
You are a senior software engineer about to begin a new iteration of software development based on a set of user stories described in a blueprint document.

Your task is to lay the foundation for implementing ${change_request_basename}.

Now, let's start the work:
- Read the user stories using ./usm cat ${change_request_file_path}
- Read the blueprint using cat ${change_request_file_path}
```

## Customization

You can customize these prompts to better suit your development workflow or to add specialized instructions for specific types of tasks.

## Notes

This directory will be populated when you run the `usm extract-workflow` command for the first time. 
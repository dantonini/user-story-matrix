# Define custom workflow with directory structure
As a  
software engineer using USM,  
I want  
to define custom workflows using a directory-based structure with a YAML configuration file and separate prompt files,  
So that  
I can create flexible, maintainable workflow definitions without struggling with lengthy YAML blocks.

## Acceptance Criteria
- USM supports loading workflow definitions from a directory structure
- The directory structure follows a convention:
  ```
  workflow-name/
  ├── workflow.yaml      # Core workflow definition
  └── prompts/           # Directory for prompt files
      ├── step1.md
      ├── step2.md
      └── shared/        # Optional directory for reusable prompts
  ```
- The workflow.yaml file defines the workflow metadata and steps:
  ```yaml
  name: "workflow-name"
  description: "Description of the workflow"
  steps:
    - id: "01-step-one"
      description: "First step description"
      prompt: "prompts/step1.md"
    - id: "02-step-two"
      description: "Second step description"
      prompt: "prompts/step2.md"
  ```
- Prompt files are stored as separate markdown files for better readability and editing
- The system validates that all referenced prompt files exist
- When a workflow is loaded, all prompt files are resolved and validated

## Priority: MUST HAVE
This is the core functionality required for custom workflow definitions. 
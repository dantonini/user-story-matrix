---
file_path: docs/user-stories/custom-workflow/ui-07-create-workflow-subcommand.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: 8c3d36fdb18e259c88e9e46547000ad591dca8aa45e180e907d3712accc380dc
---

# Create workflow subcommand
As a  
CLI user,  
I want  
to create a new `workflow` subcommand with management operations,  
So that  
users can list, initialize, and manage their custom workflows.

## Acceptance Criteria
- A new `workflow` command is added to the root command:
  ```go
  var workflowCmd = &cobra.Command{
      Use:   "workflow",
      Short: "Manage development workflows",
      Long:  `Create, list, and manage development workflows`,
  }
  ```
- The following subcommands are implemented:
  - `workflow list`: List available workflows
  - `workflow init`: Initialize a new workflow
  - `workflow validate`: Validate a workflow definition
  - `workflow show`: Display workflow details
- The `workflow list` command:
  - Discovers workflows in standard locations
  - Displays them in a table format
  - Indicates the source of each workflow (built-in, user, project)
- The `workflow init` command:
  - Creates a new workflow from a template
  - Generates the directory structure and files
  - Provides clear output about what was created
- The `workflow validate` command:
  - Performs comprehensive validation of a workflow
  - Displays detailed error messages for issues found
- The `workflow show` command:
  - Displays details about a specific workflow
  - Shows all steps and their descriptions
- All commands include help documentation
- The workflow command is properly integrated into the CLI

## Priority: SHOULD HAVE
These commands provide a better user experience for working with custom workflows. 

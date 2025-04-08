# Add workflow loading from filesystem
As a  
USM developer,  
I want  
to implement the ability to load workflow definitions from the filesystem,  
So that  
the system can support user-defined workflows in addition to built-in ones.

## Acceptance Criteria
- The `WorkflowRegistry` is extended to load workflow definitions from disk:
  ```go
  func (r *WorkflowRegistry) LoadFromDirectory(path string) (*WorkflowDefinition, error)
  ```
- The registry can discover workflows in standard locations:
  ```go
  func (r *WorkflowRegistry) DiscoverWorkflows() map[string]*WorkflowDefinition
  ```
- The workflow loading process:
  1. Validates the workflow.yaml file format
  2. Resolves all prompt file references
  3. Loads prompt content into memory
  4. Returns a complete WorkflowDefinition
- The system handles errors gracefully:
  - Missing files
  - Invalid YAML syntax
  - Missing required fields
  - References to non-existent prompt files
- The registry caches loaded workflows for performance
- The registry has a method to reload workflows when they change on disk
- When loading a workflow, the system logs:
  - The workflow being loaded
  - The source location
  - Any validation warnings

## Priority: MUST HAVE
This functionality is required for loading user-defined workflows from the filesystem. 
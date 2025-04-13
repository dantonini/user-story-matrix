---
file_path: docs/user-stories/custom-workflow/ui-06-update-code-command-for-workflow-selection.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: 5249e0ecae0ca87a89459fc36acfbc7cfb943b96fffcb3471199d48216dd2657
---

# Update code command for workflow selection
As a  
CLI user,  
I want  
to update the code command to support workflow selection,  
So that  
users can choose which workflow to use when executing the command.

## Acceptance Criteria
- The `code` command is updated to accept workflow selection flags:
  ```go
  codeCmd.Flags().StringVar(&workflowName, "workflow", "standard", "Name of the workflow to use")
  codeCmd.Flags().StringVar(&workflowPath, "workflow-path", "", "Path to a custom workflow directory")
  ```
- The command handling logic is updated:
  ```go
  Run: func(cmd *cobra.Command, args []string) {
      // Create filesystem and IO interfaces
      fs := io.NewOSFileSystem()
      term := io.NewTerminalIOWithDebug(debug)
      
      // Create workflow registry
      registry := workflow.NewWorkflowRegistry(fs)
      
      // Determine which workflow to use
      var selectedWorkflow *workflow.WorkflowDefinition
      var err error
      
      if workflowPath != "" {
          // Load from specific path
          selectedWorkflow, err = registry.LoadFromDirectory(workflowPath)
          if err != nil {
              term.PrintError(fmt.Sprintf("Failed to load workflow from path: %s", err))
              os.Exit(1)
          }
      } else {
          // Load by name
          selectedWorkflow, err = registry.GetWorkflow(workflowName)
          if err != nil {
              term.PrintError(fmt.Sprintf("Failed to load workflow '%s': %s", workflowName, err))
              os.Exit(1)
          }
      }
      
      // Create workflow manager with the selected workflow
      wm := workflow.NewWorkflowManager(fs, term, selectedWorkflow)
      
      // ... rest of the command logic ...
  }
  ```
- The command provides clear error messages when:
  - The specified workflow doesn't exist
  - The workflow path is invalid
  - The workflow definition is invalid
- The selected workflow is stored in the state file for continuity
- Command help is updated to document the new flags
- Tab completion is implemented for the workflow name flag

## Priority: MUST HAVE
This update is required to enable users to select custom workflows. 

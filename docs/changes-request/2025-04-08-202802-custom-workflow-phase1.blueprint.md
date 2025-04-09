---
name: custom-workflow-phase1
created-at: 2025-04-08T20:28:02+02:00
user-stories:
  - title: Refactor StandardWorkflowSteps structure
    file: docs/user-stories/custom-workflow/dev-01-refactor-standardworkflowsteps-structure.md
    content-hash: 

---

# Blueprint

## Overview

This change request focuses on refactoring the StandardWorkflowSteps structure to prepare the codebase for supporting custom workflow definitions. The current implementation relies on a global variable directly, which limits flexibility and makes it difficult to introduce custom workflows. This refactoring will transform the direct access pattern into a more modular, registry-based approach that maintains backward compatibility while enabling future extensibility.

## Fundamentals

### Data Structures

1. **WorkflowDefinition**: A structured representation of a complete workflow:
   ```go
   // WorkflowDefinition represents a complete workflow
   type WorkflowDefinition struct {
       Name        string
       Description string
       Steps       []WorkflowStep
   }
   ```

2. **WorkflowRegistry**: A registry to manage and provide access to available workflows:
   ```go
   // WorkflowRegistry manages available workflows
   type WorkflowRegistry struct {
       builtInWorkflows map[string]*WorkflowDefinition
   }
   ```

### Refactoring Strategy

The refactoring will follow these key principles:
1. **Interface-based approach**: Create a consistent interface for accessing workflows
2. **Backward compatibility**: Maintain existing behavior of StandardWorkflowSteps
3. **Future extensibility**: Design for easy addition of custom workflows
4. **Gradual migration**: Allow for an incremental transition away from direct StandardWorkflowSteps usage

## How to verify – Detailed User Story Breakdown

### User Story: Refactor StandardWorkflowSteps structure

#### Acceptance Criteria 1: Refactor StandardWorkflowSteps into WorkflowDefinition

**Testing Scenarios:**
- Create a new WorkflowDefinition struct with the required fields
- Verify that a WorkflowDefinition can be instantiated with a name, description, and steps
- Verify that the WorkflowDefinition can be accessed through a getter method

#### Acceptance Criteria 2: Create a WorkflowRegistry

**Testing Scenarios:**
- Create a new WorkflowRegistry struct that manages workflows
- Verify that a workflow can be registered with the registry
- Verify that a workflow can be retrieved from the registry by name
- Verify that attempting to retrieve a non-existent workflow returns an appropriate error

#### Acceptance Criteria 3: Convert StandardWorkflowSteps to a standard workflow definition

**Testing Scenarios:**
- Verify that the existing StandardWorkflowSteps is converted to a WorkflowDefinition
- Verify that the standard workflow is registered with the registry
- Verify that the standard workflow can be retrieved from the registry using a predefined constant name

#### Acceptance Criteria 4: Update WorkflowManager to work with WorkflowDefinition

**Testing Scenarios:**
- Verify that WorkflowManager can be instantiated with a specific WorkflowDefinition
- Verify that WorkflowManager falls back to the standard workflow when no definition is provided
- Verify that all WorkflowManager methods now work with the specified workflow definition

#### Acceptance Criteria 5: Maintain backward compatibility

**Testing Scenarios:**
- Verify that all existing tests continue to pass without modification
- Verify that StandardWorkflowSteps can still be accessed directly for backward compatibility
- Verify that the behavior of the system remains unchanged when using the default workflow

## What is the Plan – Detailed Action Items

### 1. Create the new data structures

1. **Define WorkflowDefinition and WorkflowRegistry types**:
   Create a new file `internal/workflow/registry.go` containing:
   - WorkflowDefinition struct
   - WorkflowRegistry struct with necessary methods
   - Standard workflow name constant

   ```go
   // Pseudo-code structure
   const StandardWorkflowName = "standard"

   // WorkflowDefinition represents a complete workflow
   type WorkflowDefinition struct {
       Name        string
       Description string
       Steps       []WorkflowStep
   }

   // WorkflowRegistry manages available workflows
   type WorkflowRegistry struct {
       builtInWorkflows map[string]*WorkflowDefinition
   }
   
   // NewWorkflowRegistry creates a new registry with the standard workflow pre-registered
   func NewWorkflowRegistry() *WorkflowRegistry {
       registry := &WorkflowRegistry{
           builtInWorkflows: make(map[string]*WorkflowDefinition),
       }
       
       // Register the standard workflow
       registry.RegisterBuiltInWorkflow(createStandardWorkflow())
       
       return registry
   }
   
   // RegisterBuiltInWorkflow adds a workflow to the registry
   func (r *WorkflowRegistry) RegisterBuiltInWorkflow(workflow *WorkflowDefinition) {
       r.builtInWorkflows[workflow.Name] = workflow
   }
   
   // GetWorkflow retrieves a workflow by name
   func (r *WorkflowRegistry) GetWorkflow(name string) (*WorkflowDefinition, error) {
       workflow, exists := r.builtInWorkflows[name]
       if !exists {
           return nil, fmt.Errorf("workflow '%s' not found", name)
       }
       return workflow, nil
   }
   
   // GetStandardWorkflow returns the standard workflow
   func (r *WorkflowRegistry) GetStandardWorkflow() *WorkflowDefinition {
       workflow, _ := r.GetWorkflow(StandardWorkflowName)
       return workflow
   }
   ```

2. **Create a function to convert the existing StandardWorkflowSteps to a WorkflowDefinition**:
   Add to `internal/workflow/registry.go`:
   ```go
   // createStandardWorkflow converts the existing StandardWorkflowSteps to a WorkflowDefinition
   func createStandardWorkflow() *WorkflowDefinition {
       return &WorkflowDefinition{
           Name:        StandardWorkflowName,
           Description: "Standard USM implementation workflow",
           Steps:       StandardWorkflowSteps,
       }
   }
   ```

### 2. Update WorkflowManager to use WorkflowDefinition

1. **Modify WorkflowManager struct to include a workflow definition**:
   Update `internal/workflow/workflow.go`:
   ```go
   // WorkflowManager handles workflow-related operations
   type WorkflowManager struct {
       fs       FileSystem
       io       UserOutput
       registry *WorkflowRegistry
       workflow *WorkflowDefinition // Current workflow being used
   }
   ```

2. **Update WorkflowManager constructor**:
   Modify `NewWorkflowManager` in `internal/workflow/workflow.go`:
   ```go
   // NewWorkflowManager creates a new workflow manager with optional workflow definition
   func NewWorkflowManager(fs FileSystem, io UserOutput, workflowName string) *WorkflowManager {
       registry := NewWorkflowRegistry()
       
       var workflow *WorkflowDefinition
       if workflowName != "" {
           // Try to get the specified workflow
           wf, err := registry.GetWorkflow(workflowName)
           if err == nil {
               workflow = wf
           } else {
               // Log warning and fall back to standard workflow
               io.PrintWarning(fmt.Sprintf("Workflow '%s' not found, using standard workflow", workflowName))
               workflow = registry.GetStandardWorkflow()
           }
       } else {
           // Use standard workflow by default
           workflow = registry.GetStandardWorkflow()
       }
       
       return &WorkflowManager{
           fs:       fs,
           io:       io,
           registry: registry,
           workflow: workflow,
       }
   }
   ```

3. **Update all methods that reference StandardWorkflowSteps**:
   Find all methods in `internal/workflow/workflow.go` that directly reference StandardWorkflowSteps and update them to use `wm.workflow.Steps` instead:

   ```go
   // Example updates (all similar references need to be changed):
   
   // Before:
   if state.CurrentStepIndex < 0 || state.CurrentStepIndex > len(StandardWorkflowSteps) {
   
   // After:
   if state.CurrentStepIndex < 0 || state.CurrentStepIndex > len(wm.workflow.Steps) {
   
   // Before:
   wm.io.PrintStep(state.CurrentStepIndex+1, len(StandardWorkflowSteps), StandardWorkflowSteps[state.CurrentStepIndex].Description)
   
   // After:
   wm.io.PrintStep(state.CurrentStepIndex+1, len(wm.workflow.Steps), wm.workflow.Steps[state.CurrentStepIndex].Description)
   ```

### 3. Maintain backward compatibility

1. **Keep StandardWorkflowSteps as a global variable**:
   Maintain the existing `StandardWorkflowSteps` variable in `internal/workflow/workflow.go`, but update its usage context to indicate it's being maintained for backward compatibility.

2. **Update the WorkflowManager constructor to maintain backward compatibility**:
   Add a simpler version of the constructor that doesn't require specifying a workflow name:
   ```go
   // NewDefaultWorkflowManager creates a new workflow manager with the standard workflow
   func NewDefaultWorkflowManager(fs FileSystem, io UserOutput) *WorkflowManager {
       return NewWorkflowManager(fs, io, "")
   }
   ```

3. **Update commands that use WorkflowManager**:
   In `cmd/code.go`, update the WorkflowManager instantiation:
   ```go
   // Before:
   wm := workflow.NewWorkflowManager(fs, term)
   
   // After:
   wm := workflow.NewDefaultWorkflowManager(fs, term)
   ```

### 4. Update tests to validate the refactoring

1. **Create tests for the new registry and workflow definition**:
   Create `internal/workflow/registry_test.go`:
   ```go
   // Test cases should include:
   // - Creating a new registry
   // - Registering a workflow
   // - Getting a workflow by name
   // - Getting the standard workflow
   // - Handling non-existent workflows
   ```

2. **Update existing tests to work with the new structure**:
   In `internal/workflow/workflow_test.go`, update the WorkflowManager instantiation and any direct references to StandardWorkflowSteps.

### 5. Documentation

1. **Add documentation to the new types and methods**:
   - Add detailed comments to all new types and methods
   - Document the refactoring approach and backward compatibility considerations

2. **Update any existing documentation that references the workflow structure**:
   - Ensure any documentation or user guides are updated to reflect the new design

## Implementation Order

1. Create the new data structures in `registry.go` with tests
2. Update the WorkflowManager struct and constructor
3. Update WorkflowManager methods to use the workflow field
4. Add backward compatibility support
5. Update all commands and client code
6. Run all tests to verify backward compatibility
7. Update documentation

This approach allows for incremental development and testing, ensuring that each part of the refactoring can be validated before proceeding to the next step.

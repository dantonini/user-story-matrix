---
name: custom workflow phase 3
created-at: 2025-04-13T15:26:26+02:00
user-stories:
  - title: Implement template variables support
    file: docs/user-stories/custom-workflow/dev-05-implement-template-variables-support.md
    content-hash: 6611f6d754505eae72e149a3592de0cd9946e7f5d1dfce4004b9cfb53404d622
  - title: Deprecate StandardWorkflowSteps
    file: docs/user-stories/custom-workflow/dev-06-deprecate-standardworkflowsteps.md
    content-hash: 2de9142b0af5ea751897be5aabc480f7e55f13109e811eb69a810a63a727135d
  - title: Migrate legacy workflows
    file: docs/user-stories/custom-workflow/dev-07-migrate-legacy-workflows.md
    content-hash: 6736a44a6aedbd78c8e3057913f4b09918c554c53ed412187cb5bfe9bc8660e7

---

# Blueprint

## Overview

This change request aims to enhance the USM workflow system by improving its flexibility, reusability, and backward compatibility. The implementation will:

1. Add template variable support to workflow prompt templates, allowing for dynamic content with conditional sections and default values
2. Deprecate the existing `StandardWorkflowSteps` global variable while ensuring backward compatibility
3. Implement a seamless migration path from the legacy workflow system to the new custom workflow system

These changes will allow for more customizable workflows with reusable template components while maintaining compatibility with existing projects.

## Fundamentals

### Data Structures

#### 1. Extended WorkflowStep
The `WorkflowStep` struct will be extended to include variables for template substitution:

```go
type WorkflowStep struct {
    ID          string              // Unique identifier
    Description string              // Human-readable description
    Prompt      string              // Path to prompt file
    Variables   map[string]string   // Variables for template substitution
    source      promptSource        // Internal field (unchanged)
}
```

#### 2. Template Context
A new structure to hold the template execution context:

```go
type TemplateContext struct {
    Variables map[string]interface{}  // Variables available to templates
    Functions template.FuncMap        // Custom template functions
}
```

#### 3. External Workflow Structure
The serialized workflow definition also needs to be updated:

```go
type ExternalWorkflowStep struct {
    ID          string                 `json:"id" yaml:"id"`
    Description string                 `json:"description" yaml:"description"`
    Prompt      string                 `json:"prompt" yaml:"prompt"`
    Variables   map[string]string      `json:"variables" yaml:"variables"`
}
```

### Algorithms

#### 1. Template Processing

```go
// Pseudo-code for template processing
func ApplyTemplateVariables(promptContent string, variables map[string]string) (string, error) {
    // Create template context with variables and custom functions
    context := TemplateContext{
        Variables: variables,
        Functions: template.FuncMap{
            "default": defaultFunction,
            // Other custom functions
        },
    }
    
    // Parse template
    tmpl, err := template.New("prompt").Funcs(context.Functions).Parse(promptContent)
    if err != nil {
        return "", fmt.Errorf("template parsing error: %w", err)
    }
    
    // Execute template with variables
    var output bytes.Buffer
    err = tmpl.Execute(&output, context.Variables)
    if err != nil {
        return "", fmt.Errorf("template execution error: %w", err)
    }
    
    return output.String(), nil
}
```

#### 2. Legacy Workflow Mapping

```go
// Pseudo-code for legacy workflow mapping
func GetWorkflowStepFromLegacy(index int) (WorkflowStep, error) {
    // Validate index
    if index < 0 || index >= len(StandardWorkflowSteps) {
        return WorkflowStep{}, errors.New("invalid step index")
    }
    
    // Get step from standard workflow in registry
    workflow := GetGlobalRegistry().GetStandardWorkflow()
    if workflow == nil || index >= len(workflow.Steps) {
        // Fallback to legacy array if registry access fails
        return StandardWorkflowSteps[index], nil
    }
    
    return workflow.Steps[index], nil
}
```

#### 3. State File Migration

```go
// Pseudo-code for state file migration
func MigrateStateFile(stateFile string) error {
    // Load existing state
    state, err := LoadWorkflowState(stateFile)
    if err != nil {
        return err
    }
    
    // Check if workflow name is already set
    if state.WorkflowName != "" {
        // Already migrated
        return nil
    }
    
    // Add standard workflow information
    state.WorkflowName = StandardWorkflowName
    
    // Save updated state
    return SaveWorkflowState(state)
}
```

### Refactoring Strategy

1. The refactoring will focus on adding new functionality without breaking existing code
2. The changes will be introduced gradually with proper deprecation notices
3. For backward compatibility, the `StandardWorkflowSteps` array will continue to work but will be synchronized with the registry
4. All steps will include comprehensive unit tests to ensure compatibility

## How to verify – Detailed User Story Breakdown

### User Story 1: Implement template variables support

#### Acceptance Criteria 1: WorkflowStep struct update
- Verify that the `WorkflowStep` struct includes a `Variables` field of type `map[string]string`
- Verify that the `Variables` field is properly documented in code comments
- Verify that all related structures (e.g., `ExternalWorkflowStep`) also include the variables field
- Verify that existing code using the struct continues to work without changes

**Testing Scenario:**
1. Create a `WorkflowStep` instance with variables
2. Verify the variables can be set and retrieved
3. Verify serialization/deserialization works correctly
4. Verify backward compatibility with existing code

#### Acceptance Criteria 2: Template processing system
- Verify that the `ApplyTemplateVariables` function exists and has the correct signature
- Verify it processes templates correctly with the provided variables
- Verify it returns appropriate errors when the template is invalid

**Testing Scenario:**
1. Call `ApplyTemplateVariables` with a simple template and variables
2. Verify the output matches the expected result
3. Try with invalid templates and verify error handling

#### Acceptance Criteria 3: Go's text/template package usage
- Verify that the implementation uses the `text/template` package
- Verify that template syntax follows Go's conventions

**Testing Scenario:**
1. Examine code to confirm use of `text/template`
2. Test templates using standard Go template syntax

#### Acceptance Criteria 4: Template variable syntax
- Verify that templates use the `{{.variable_name}}` syntax
- Verify that this syntax works correctly in actual templates

**Testing Scenario:**
1. Create templates with various variable references
2. Process templates and verify variable substitution works

#### Acceptance Criteria 5: Advanced template features
- Verify support for default values using `{{.variable_name | default "default value"}}`
- Verify support for conditional sections with if/else
- Verify support for iteration over arrays

**Testing Scenario:**
1. Test templates with default values for missing variables
2. Test templates with conditional sections that evaluate to both true and false
3. Test templates with array iteration using sample data

#### Acceptance Criteria 6: Security (escaping)
- Verify that variables are properly escaped to prevent injection vulnerabilities
- Verify that HTML content in variables doesn't render as HTML

**Testing Scenario:**
1. Test templates with variables containing HTML or template syntax
2. Verify output doesn't interpret these as code/markup

#### Acceptance Criteria 7: Undefined variable warnings
- Verify that the system provides warnings for undefined variables
- Verify these warnings are helpful for debugging

**Testing Scenario:**
1. Process a template with references to undefined variables
2. Verify warnings are generated and are helpful

#### Acceptance Criteria 8: Error handling
- Verify that template errors are caught and handled gracefully
- Verify that meaningful error messages are returned

**Testing Scenario:**
1. Process templates with syntax errors
2. Verify errors are caught and explain the issue

#### Acceptance Criteria 9: Documentation
- Verify that documentation exists for template syntax
- Verify that available functions are documented

**Testing Scenario:**
1. Check for documentation in code comments
2. Check for user-facing documentation

### User Story 2: Deprecate StandardWorkflowSteps

#### Acceptance Criteria 1: Deprecation notice
- Verify that `StandardWorkflowSteps` is marked as deprecated with the correct comment
- Verify the comment includes guidance on the alternative approach

**Testing Scenario:**
1. Check code for appropriate deprecation comment
2. Verify comment references `WorkflowRegistry.GetWorkflow("standard")`

#### Acceptance Criteria 2: Static analysis linter rules
- Verify that linter rules are added to flag direct usage
- Verify the linter rules provide helpful messages

**Testing Scenario:**
1. Write code that directly uses `StandardWorkflowSteps`
2. Run linter and verify warning is shown

#### Acceptance Criteria 3: Compatibility layer
- Verify that a compatibility layer exists between legacy and new APIs
- Verify that it correctly maps between the two systems

**Testing Scenario:**
1. Test code using both legacy and new approaches
2. Verify both approaches produce the same results

#### Acceptance Criteria 4: Backward compatibility
- Verify that existing code using `StandardWorkflowSteps` continues to work
- Verify that no runtime errors occur from the deprecation

**Testing Scenario:**
1. Run existing code that uses `StandardWorkflowSteps`
2. Verify functionality works as expected

#### Acceptance Criteria 5: Documentation updates
- Verify that documentation is updated to reflect the deprecation
- Verify migration guides exist for updating code

**Testing Scenario:**
1. Check updated documentation for deprecation notices
2. Verify migration guides include concrete examples

#### Acceptance Criteria 6: Unit tests
- Verify that unit tests exist for both old and new approaches
- Verify that these tests confirm functionality works in both cases

**Testing Scenario:**
1. Review unit tests for legacy and new workflow usage
2. Verify test coverage for compatibility layer

#### Acceptance Criteria 7: Removal plan
- Verify a plan exists for complete removal in a future version
- Verify this plan is documented

**Testing Scenario:**
1. Check documentation for removal timeline
2. Verify plan includes migration strategies

### User Story 3: Migrate legacy workflows

#### Acceptance Criteria 1: Seamless migration
- Verify that the system converts `StandardWorkflowSteps` to a built-in workflow template
- Verify compatibility with existing state files
- Verify no manual migration is required for users

**Testing Scenario:**
1. Create a project using the legacy workflow
2. Upgrade to the new version with migration support
3. Verify workflows continue to function without manual intervention

#### Acceptance Criteria 2: State file detection
- Verify the system assumes "standard" workflow for existing state files
- Verify appropriate warnings are logged in debug mode
- Verify state files are updated with workflow information

**Testing Scenario:**
1. Create a legacy state file without workflow information
2. Open it with the new system
3. Verify it's correctly identified as standard workflow
4. Check the state file is updated with workflow information

#### Acceptance Criteria 3: Backward compatibility
- Verify the system handles direct references to `StandardWorkflowSteps`
- Verify appropriate deprecation warnings are logged
- Verify functionality continues to work correctly

**Testing Scenario:**
1. Write code that directly references `StandardWorkflowSteps`
2. Run with the new system
3. Verify deprecation warnings
4. Verify functionality works

#### Acceptance Criteria 4: Migration command
- Verify the migration command exists with the correct syntax
- Verify it successfully migrates state files between workflows

**Testing Scenario:**
1. Create a state file
2. Run migration command to a different workflow
3. Verify the state file is updated correctly

#### Acceptance Criteria 5: Documentation
- Verify the migration process is well-documented
- Verify documentation includes examples and troubleshooting

**Testing Scenario:**
1. Review documentation for migration instructions
2. Verify examples for common migration scenarios

## What is the Plan – Detailed Action Items

### Task 1: Implement Template Variables Support

#### 1.1 Update WorkflowStep Structure:
- Update `internal/workflow/workflow.go` to add the `Variables` field to `WorkflowStep` struct
- Update related structures like `ExternalWorkflowStep` in `internal/workflow/loader.go`
- Add proper documentation for the new field
- Update any serialization/deserialization code to handle the new field

#### 1.2 Create Template Processing System:
- Create a new file `internal/workflow/template.go` to implement template processing
- Implement `ApplyTemplateVariables` function using the text/template package
- Add custom functions including `default` for fallback values
- Implement proper error handling for template processing

#### 1.3 Add Template Execution to Workflow Manager:
- Update `WorkflowManager.ExecuteStep` in `internal/workflow/workflow.go` to process templates
- Apply variable substitution before showing prompts to users
- Add proper error handling for template execution

#### 1.4 Update External Workflow Format:
- Modify `WorkflowFileStep` in `internal/workflow/extract.go` to include variables
- Update serialization/deserialization logic for workflow files
- Ensure backward compatibility with existing workflow files

#### 1.5 Create Template Documentation:
- Add documentation in code comments
- Create a user-facing guide for template syntax
- Include examples for common template patterns

#### 1.6 Write Unit Tests:
- Create unit tests for `ApplyTemplateVariables`
- Test various template features (substitution, conditionals, iteration)
- Test error conditions and edge cases

### Task 2: Deprecate StandardWorkflowSteps

#### 2.1 Mark as Deprecated:
- Update `internal/workflow/standard_workflow.go` to add deprecation notice
- Add guidance on using `WorkflowRegistry.GetWorkflow("standard")` instead

#### 2.2 Implement Compatibility Layer:
- Create a compatibility mechanism in `internal/workflow/registry.go`
- Ensure that changes to `StandardWorkflowSteps` are reflected in the registry
- Ensure that changes via the registry are reflected in `StandardWorkflowSteps`

#### 2.3 Update Documentation:
- Create migration guide for updating code
- Document deprecation in README and user documentation
- Add examples of migrating from old to new API

#### 2.4 Write Unit Tests:
- Add tests for compatibility layer
- Verify both old and new approaches work
- Test edge cases in transition

#### 2.5 Create Removal Plan:
- Document timeline for complete removal
- Specify migration strategies for complete removal
- Include in project roadmap

### Task 3: Migrate Legacy Workflows

#### 3.1 Implement State File Migration:
- Update `WorkflowManager.LoadState` in `internal/workflow/workflow.go` to handle legacy state files
- Add detection for missing workflow information
- Auto-assign standard workflow when information is missing
- Update state file with workflow information

#### 3.2 Enhance Error Handling:
- Add specific error types for migration issues
- Implement graceful fallbacks for compatibility issues
- Add clear error messages for troubleshooting

#### 3.3 Add Warning System:
- Implement a warning system for deprecated usage
- Log warnings when legacy code paths are used
- Make warnings visible but not intrusive

#### 3.4 Write Documentation:
- Create detailed migration guide
- Add troubleshooting section for common issues

#### 3.5 Write Integration Tests:
- Test migration between different workflows
- Test backward compatibility with existing projects
- Test edge cases and error conditions

This plan ensures a smooth transition from the legacy workflow system to the new template-based system with minimal disruption to existing projects and a clear migration path for developers.

# Custom Workflow Phase 3 - Implementation Document

## Overview

This implementation document describes the architecture and technical details of Custom Workflow Phase 3, which focused on enhancing USM's workflow system with greater flexibility, reusability, and backward compatibility. The implementation introduced three major features:

1. Template variables support in workflow prompts
2. Deprecation of `StandardWorkflowSteps` with a backward-compatible transition path
3. Migration functionality for legacy workflow states

The implementation was completed in four phases: foundation, minimum viable implementation (MVI), extension, and refinement. This document outlines the final architecture, data structures, algorithms, and design decisions made throughout the implementation process.

## Data Structures

### 1. Enhanced WorkflowStep

The `WorkflowStep` struct was enhanced to support template variables:

```go
type WorkflowStep struct {
    ID          string            // Unique identifier (e.g., "01-laying-the-foundation")
    Description string            // Human-readable description
    Prompt      string            // AI agent instructions with variable interpolation
    Variables   map[string]string // Variables for template substitution
    source      promptSource      // Internal field for tracking prompt source
}
```

The new `Variables` field enables dynamic content in workflow prompts by storing key-value pairs that can be referenced in templates.

### 2. Template Context

A template context structure was implemented to manage template execution:

```go
type TemplateContext struct {
    Variables map[string]interface{} // Variables available to templates
    Functions template.FuncMap       // Custom template functions
}
```

This context structure supports both simple variables and complex data structures, allowing for nested objects and arrays in templates.

### 3. WorkflowState with Workflow Information

The `WorkflowState` struct was enhanced to track workflow information:

```go
type WorkflowState struct {
    ChangeRequestPath string    // Path to the change request file
    CurrentStepIndex  int       // Index of the current step (0-based)
    LastModified      time.Time // When the state was last updated
    CompletedSteps    []string  // List of completed step IDs
    WorkflowName      string    // Name of the workflow being used
    WorkflowPath      string    // Optional path to the workflow definition
}
```

The addition of `WorkflowName` and `WorkflowPath` fields enables support for multiple workflow definitions and migration between them.

### 4. Template Processor

A template processor was implemented to handle caching and reuse of parsed templates:

```go
type TemplateProcessor struct {
    // Cached templates to improve performance for repeated executions
    templateCache map[string]*template.Template
}
```

This processor improves performance by caching parsed templates, reducing the overhead of repeated template parsing.

## Key Algorithms

### 1. Template Processing

The core of the template variables implementation is the `ApplyTemplateVariables` function:

```go
func ApplyTemplateVariables(templateContent string, variables map[string]string) (string, error) {
    // 1. Validate template structure
    if err := ValidateTemplate(templateContent); err != nil {
        return "", err
    }

    // 2. Process variables into structured map with support for nested objects and arrays
    processedVars := processVariablesIntoStructuredMap(variables)

    // 3. Handle default values with special preprocessing
    processedTemplate := handleDefaultValues(templateContent, processedVars)

    // 4. Create template with custom functions
    tmpl, err := template.New("prompt").Funcs(customFunctions).Parse(processedTemplate)
    if err != nil {
        return "", fmt.Errorf("template parsing error: %w", err)
    }

    // 5. Execute template with variables
    var output bytes.Buffer
    err = tmpl.Execute(&output, processedVars)
    if err != nil {
        return "", fmt.Errorf("template execution error: %w", err)
    }

    // 6. Clean up output and return
    return cleanOutput(output.String()), nil
}
```

This algorithm handles complex template processing with several key features:
- Template validation to ensure proper structure
- Transformation of flat variable maps into nested structures
- Custom preprocessing for default values
- Support for conditional sections and array iteration
- Error handling at multiple stages

### 2. State File Migration

The migration system enables transitioning between different workflows while preserving progress:

```go
func MigrateStateFile(fs io.FileSystem, stateFilePath string, targetWorkflowName string, createBackup bool) ([]string, error) {
    // 1. Create backup if requested
    if createBackup {
        backupPath, err := CreateStateBackup(fs, stateFilePath)
        if err != nil {
            return warnings, fmt.Errorf("failed to create backup: %w", err)
        }
        warnings = append(warnings, fmt.Sprintf("Created backup at %s", backupPath))
    }

    // 2. Load existing state
    content, err := fs.ReadFile(stateFilePath)
    if err != nil {
        return warnings, fmt.Errorf("failed to read state file: %w", err)
    }

    // 3. Parse state
    var state WorkflowState
    err = json.Unmarshal(content, &state)
    if err != nil {
        return warnings, fmt.Errorf("failed to parse state file: %w", err)
    }

    // 4. Map progress between workflows
    originalWorkflow := state.WorkflowName
    registry := GetGlobalRegistry()
    sourceWorkflow, _ := registry.GetWorkflow(originalWorkflow)
    targetWorkflow, _ := registry.GetWorkflow(targetWorkflowName)

    // 5. Create a workflow manager and map progress
    wm := &WorkflowManager{fs: fs, registry: registry, workflow: sourceWorkflow}
    newState, mappingWarnings := wm.MapProgressBetweenWorkflows(state, targetWorkflowName)
    
    // 6. Save updated state
    updatedContent, err := json.MarshalIndent(newState, "", "  ")
    if err != nil {
        return warnings, fmt.Errorf("failed to serialize updated state: %w", err)
    }
    
    // 7. Write back to file
    err = fs.WriteFile(stateFilePath, updatedContent, 0644)
    if err != nil {
        return warnings, fmt.Errorf("failed to write updated state: %w", err)
    }

    return warnings, nil
}
```

This algorithm implements a robust migration process with several key features:
- Automatic backup creation for safety
- Validation of source and target workflows
- Step progress mapping between workflows
- Comprehensive warning collection
- Error handling with detailed error messages

### 3. Compatibility Layer for StandardWorkflowSteps

The compatibility layer synchronizes between the legacy global array and the new registry system:

```go
func initCompatibilityLayer() {
    // 1. Set up warning system
    legacyAccessWarning.logWriter = os.Stderr
    
    // 2. Create workflow definition from StandardWorkflowSteps
    registry := GetGlobalRegistry()
    workflow := &WorkflowDefinition{
        Name:        StandardWorkflowName,
        Description: "Standard workflow for implementing user stories",
        Steps:       StandardWorkflowSteps,
    }
    
    // 3. Register with the registry
    registry.RegisterBuiltInWorkflow(workflow)
    
    // 4. Set up synchronization when registry changes
    registry.AddWorkflowChangeCallback(StandardWorkflowName, func(workflow *WorkflowDefinition) {
        if workflow != nil && len(workflow.Steps) > 0 {
            StandardWorkflowSteps = workflow.Steps
        }
    })
}
```

This bidirectional synchronization ensures that changes to either the legacy array or the registry are reflected in both places, maintaining compatibility with existing code.

### 4. Progress Mapping Between Workflows

The algorithm for mapping progress between workflows preserves the user's progress when switching workflows:

```go
func (wm *WorkflowManager) MapProgressBetweenWorkflows(state WorkflowState, targetWorkflowName string) (WorkflowState, []string) {
    // 1. Create a copy of the state
    newState := state
    newState.WorkflowName = targetWorkflowName
    
    // 2. Get workflows
    sourceWorkflow := wm.workflow
    targetWorkflow, err := wm.registry.GetWorkflow(targetWorkflowName)
    if err != nil {
        return newState, []string{fmt.Sprintf("Target workflow '%s' not found", targetWorkflowName)}
    }
    
    // 3. Map completed steps by ID
    completedStepMap := make(map[string]bool)
    for _, stepID := range state.CompletedSteps {
        completedStepMap[stepID] = true
    }
    
    // 4. Build new completed steps list with matching IDs from target workflow
    newCompletedSteps := []string{}
    for _, step := range targetWorkflow.Steps {
        if completedStepMap[step.ID] {
            newCompletedSteps = append(newCompletedSteps, step.ID)
        }
    }
    
    // 5. Map current step index to equivalent position
    if sourceWorkflow != nil && state.CurrentStepIndex < len(sourceWorkflow.Steps) {
        currentStepID := sourceWorkflow.Steps[state.CurrentStepIndex].ID
        
        // Find the same step ID in target workflow
        for i, step := range targetWorkflow.Steps {
            if step.ID == currentStepID {
                newState.CurrentStepIndex = i
                break
            }
        }
    }
    
    // 6. Safety check for step index
    if newState.CurrentStepIndex >= len(targetWorkflow.Steps) {
        newState.CurrentStepIndex = 0
    }
    
    newState.CompletedSteps = newCompletedSteps
    return newState, warnings
}
```

This algorithm implements a sophisticated mapping system that preserves progress by:
- Matching step IDs between workflows
- Determining the equivalent current step in the target workflow
- Preserving completed steps that exist in both workflows
- Handling edge cases gracefully

## Design Decisions

### 1. Custom Preprocessing for Default Values

One key design decision was to implement a custom preprocessing step for default values using regular expressions:

```go
defaultPattern := regexp.MustCompile(`{{\.([a-zA-Z0-9_]+)\s*\|\s*default\s*"([^"]*)"}}`)
processedTemplate := defaultPattern.ReplaceAllStringFunc(templateContent, func(match string) string {
    // Extract variable name and default value
    submatches := defaultPattern.FindStringSubmatch(match)
    varName := submatches[1]
    defaultValue := submatches[2]
    
    // Check if the variable has a value
    if value, exists := processedVars[varName]; exists && value != "" {
        return fmt.Sprintf("{{.%s}}", varName)
    }
    
    // Add the default value to the variables
    processedVars[varName] = defaultValue
    return fmt.Sprintf("{{.%s}}", varName)
})
```

This approach was chosen because:
1. Go's built-in template default function has limitations when handling complex data types
2. Preprocessing provides more control over default value behavior
3. It allows for consistent handling of missing variables and empty strings

### 2. Bidirectional Synchronization for Backward Compatibility

A critical design decision was to implement bidirectional synchronization between `StandardWorkflowSteps` and the registry:

```go
// Register with the registry
registry.RegisterBuiltInWorkflow(workflow)

// Set up synchronization when registry changes
registry.AddWorkflowChangeCallback(StandardWorkflowName, func(workflow *WorkflowDefinition) {
    if workflow != nil && len(workflow.Steps) > 0 {
        StandardWorkflowSteps = workflow.Steps
    }
})
```

This approach ensures:
1. Changes to `StandardWorkflowSteps` are reflected in the registry
2. Changes to the registry's standard workflow are reflected in `StandardWorkflowSteps`
3. Existing code continues to work without modification
4. Proper deprecation warnings are shown for direct access

### 3. Automatic Migration for Legacy State Files

The implementation includes automatic detection and migration of legacy state files without workflow information:

```go
func AutoMigrateStateFile(fs io.FileSystem, io UserOutput, stateFilePath string) error {
    // Check if this needs migration
    needsMigration, err := needsWorkflowMigration(fs, stateFilePath)
    if err != nil {
        return fmt.Errorf("failed to check migration status: %w", err)
    }

    if !needsMigration {
        // No migration needed
        return nil
    }

    // Create a backup before migration
    backupPath, err := CreateStateBackup(fs, stateFilePath)
    if err != nil {
        return fmt.Errorf("failed to create backup before migration: %w", err)
    }

    // Update the state with the standard workflow
    state.WorkflowName = StandardWorkflowName
    
    // Save the updated state
    updatedContent, err := json.MarshalIndent(state, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to serialize updated state: %w", err)
    }

    // Write back to the file
    err = fs.WriteFile(stateFilePath, updatedContent, 0644)
    if err != nil {
        return fmt.Errorf("failed to write updated state: %w", err)
    }

    return nil
}
```

This approach provides seamless migration by:
1. Automatically detecting legacy state files
2. Creating backups before modifications
3. Assigning the standard workflow name to legacy files
4. Handling edge cases such as invalid step indices

### 4. Structured Error Handling

The implementation adopts a consistent approach to error handling with structured error messages:

```go
// Error message templates
const (
    ErrFileNotFound         = "❌ Error: File %s not found."
    ErrInvalidStateFile     = "⚠️ Warning: Invalid state file detected for %s. Starting from the beginning."
    ErrStateUpdateFailed    = "❌ Error: Failed to update workflow state: %s"
    // ... additional error constants
)
```

This approach provides:
1. Consistent error formatting throughout the codebase
2. Clear distinction between errors, warnings, and progress messages
3. Improved user experience with actionable error messages
4. Better maintainability by centralizing error message definitions

## Implementation Challenges and Solutions

### 1. Variable Scoping in Templates

A significant challenge was ensuring variable accessibility across different template scopes, particularly inside range loops:

```go
// Look for variable references in the content that aren't dot (current item)
varRefRegex := regexp.MustCompile(`{{\.([a-zA-Z0-9_]+)}}`)
varRefs := varRefRegex.FindAllStringSubmatch(content, -1)

// Create a modified content with variable references accessible in range scope
modifiedContent := content
for _, ref := range varRefs {
    if len(ref) >= 2 {
        varName := ref[1]
        // Skip if it's referencing the same variable as the range
        if varName == rangeName {
            continue
        }

        // Replace with $ prefix to access the dot context directly
        modifiedContent = strings.ReplaceAll(
            modifiedContent,
            fmt.Sprintf("{{.%s}}", varName),
            fmt.Sprintf("{{$%s}}", varName),
        )

        // Set up variable declaration at the beginning of the range
        modifiedContent = fmt.Sprintf("{{$%s := $.%s}}%s", 
            varName, varName, modifiedContent)
    }
}
```

This solution:
1. Detects variable references inside range loops
2. Converts them to use the `$` prefix for global scope access
3. Adds variable declarations at the beginning of each range block
4. Preserves the original variable value in a globally accessible scope

### 2. Maintaining Legacy Compatibility

Ensuring backward compatibility while deprecating `StandardWorkflowSteps` was a challenge:

```go
// GetWorkflowStepFromLegacy provides a compatibility function to access
// workflow steps using the registry while supporting direct indexing into
// StandardWorkflowSteps for legacy code.
func GetWorkflowStepFromLegacy(index int) (WorkflowStep, error) {
    // Log warning about legacy access (only once per process)
    logLegacyAccessWarning()
    
    // Validate index
    if index < 0 || index >= len(StandardWorkflowSteps) {
        return WorkflowStep{}, fmt.Errorf("invalid step index: %d", index)
    }
    
    // Get step from standard workflow in registry
    registry := GetGlobalRegistry()
    if registry == nil {
        // If registry isn't available, fall back to StandardWorkflowSteps
        return StandardWorkflowSteps[index], nil
    }
    
    workflow := registry.GetStandardWorkflow()
    if workflow == nil || index >= len(workflow.Steps) {
        // Fallback to legacy array if registry access fails
        return StandardWorkflowSteps[index], nil
    }
    
    return workflow.Steps[index], nil
}
```

This solution:
1. Provides a compatibility function to bridge between legacy and new approaches
2. Logs deprecation warnings with call stack information
3. Ensures graceful fallback if registry access fails
4. Maintains the same behavior for existing code

### 3. Complex Template Features

Supporting advanced template features like nested conditionals and array iteration required sophisticated parsing:

```go
// Pre-process the template to handle variable references inside range loops
rangeRegex := regexp.MustCompile(`{{range\s+\.([a-zA-Z0-9_]+)}}(.*?){{end}}`)
processedTemplate = rangeRegex.ReplaceAllStringFunc(processedTemplate, func(match string) string {
    submatches := rangeRegex.FindStringSubmatch(match)
    if len(submatches) < 3 {
        return match
    }

    rangeName := submatches[1]
    content := submatches[2]

    // Process the content to ensure variables are accessible
    // ... implementation details
    
    return fmt.Sprintf("{{range .%s}}%s{{end}}", rangeName, modifiedContent)
})
```

This solution:
1. Uses regular expressions to identify range loops
2. Preserves the template structure while modifying variable references
3. Ensures variables are accessible in all template scopes
4. Maintains compatibility with Go's template syntax

## Conclusion

The Custom Workflow Phase 3 implementation successfully delivered:

1. A robust template variables system supporting advanced features like default values, conditionals, and array iteration
2. A deprecation path for `StandardWorkflowSteps` that maintains backward compatibility
3. A comprehensive migration system for transitioning between workflows

The implementation followed a phased approach, starting with the core foundation, implementing the minimum viable functionality, extending with advanced features, and finally refining for robustness and maintainability.

The resulting architecture provides a flexible and extensible workflow system that can support diverse user needs while maintaining backward compatibility with existing projects. 
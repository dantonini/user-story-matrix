# Custom Workflow Phase 1: Implementation Details

This document details the implementation of the custom workflow functionality in the USM tool, providing a comprehensive technical overview of the architecture, data structures, algorithms, and design decisions.

## Architecture Overview

The custom workflow implementation follows a registry-based pattern combined with a file-based workflow definition system. This approach enables both programmatically defined workflows and externally defined workflows in JSON and YAML formats.

### Key Components

1. **Workflow Registry**: Central management of workflow definitions
2. **Workflow Manager**: Interface for executing and managing workflow state
3. **External Workflow Loader**: Import/export functionality for workflow definitions
4. **State Management**: Persistence of workflow execution state

## Data Structures

### Workflow Definition

The core of the implementation is the `WorkflowDefinition` structure in `internal/workflow/registry.go`:

```go
type WorkflowDefinition struct {
    // Name uniquely identifies the workflow (e.g., "standard", "custom-tutorial")
    Name string

    // Description provides a human-readable explanation of the workflow's purpose
    Description string

    // Steps contains the ordered sequence of steps that make up this workflow
    Steps []WorkflowStep
}
```

Each workflow contains a sequence of steps defined by the `WorkflowStep` structure in `internal/workflow/workflow.go`:

```go
type WorkflowStep struct {
    ID          string // Unique identifier (e.g., "01-laying-the-foundation")
    Description string // Human-readable description
    Prompt      string // AI agent instructions with variable interpolation
}
```

### Workflow Registry

The registry in `internal/workflow/registry.go` provides a centralized repository for workflow definitions:

```go
type WorkflowRegistry struct {
    // builtInWorkflows maps workflow names to their definitions
    builtInWorkflows map[string]*WorkflowDefinition
    // mutex protects concurrent access to the workflows map
    mutex sync.RWMutex
}
```

A global singleton registry is implemented to ensure consistent access across all components:

```go
var (
    globalRegistry     *WorkflowRegistry
    globalRegistryOnce sync.Once
    globalRegistryLock sync.RWMutex
)
```

### External Workflow Format

For serialization and deserialization, `internal/workflow/loader.go` defines parallel structures:

```go
type ExternalWorkflowDefinition struct {
    Name        string `json:"name" yaml:"name"`
    Description string `json:"description" yaml:"description"`
    Steps       []ExternalWorkflowStep `json:"steps" yaml:"steps"`
}

type ExternalWorkflowStep struct {
    ID          string `json:"id" yaml:"id"`
    Description string `json:"description" yaml:"description"`
    Prompt      string `json:"prompt" yaml:"prompt"`
}
```

### Workflow State

Execution state is tracked using the `WorkflowState` structure in `internal/workflow/workflow.go`:

```go
type WorkflowState struct {
    ChangeRequestPath string    // Path to the change request file
    CurrentStepIndex  int       // Index of the current step (0-based)
    LastModified      time.Time // When the state was last updated
    CompletedSteps    []string  // List of completed step IDs
}
```

## Key Algorithms and Processes

### Workflow Registration

The registry provides methods for registering and retrieving workflows:

1. **Registration**: Through `RegisterBuiltInWorkflow()`, workflows are added to the registry with their name as the key.
2. **Global Access**: `GetGlobalRegistry()` ensures a singleton pattern with thread-safe access.
3. **Standard Workflow**: `GetStandardWorkflow()` provides guaranteed access to the default workflow.

### External Workflow Loading

The workflow loader in `internal/workflow/loader.go` implements a robust algorithm for loading workflow definitions from files:

1. **Format Detection**: Based on file extension (JSON/YAML)
2. **Content Parsing**: Using the appropriate parser (json/yaml packages)
3. **Validation**: Through `validateExternalWorkflow()` to ensure all required fields are present
4. **Conversion**: From `ExternalWorkflowDefinition` to `WorkflowDefinition` via `ToWorkflowDefinition()`

Directory-based discovery is implemented in `LoadWorkflowsFromDirectory()`:

1. **Directory Scanning**: Using `ReadDir()` to list all files
2. **Format Filtering**: Via `isWorkflowFile()` to process only supported formats
3. **Bulk Loading**: Processing each file and registering valid workflows
4. **Error Tolerance**: Continuing despite individual file errors

### Workflow State Management

The `WorkflowManager` in `internal/workflow/workflow.go` implements a state machine for tracking progress:

1. **State Loading**: `LoadState()` retrieves or initializes workflow state
2. **State Validation**: Checking bounds and correcting invalid states
3. **Progress Tracking**: `UpdateState()` for advancing through workflow steps
4. **Persistence**: Via `SaveState()` to store state between sessions

### Step Execution

The `StepExecutor` (in execution-related files) handles the execution of workflow steps:

1. **Variable Interpolation**: Replacing placeholders in prompts with actual values
2. **Validation**: Checking for undefined variables and handling empty prompts
3. **Formatting**: Converting prompts to structured instructions for clarity

## Interface Design

### FileSystem Abstraction

All file operations use the `FileSystem` interface:

```go
type FileSystem interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    MkdirAll(path string, perm os.FileMode) error
    Exists(path string) bool
    ReadDir(path string) ([]os.DirEntry, error)
}
```

This allows for easy mocking in tests and flexibility in implementation.

### User Output Abstraction

User interaction is abstracted through the `UserOutput` interface:

```go
type UserOutput interface {
    Print(message string)
    PrintSuccess(message string)
    PrintError(message string)
    PrintWarning(message string)
    PrintProgress(message string)
    PrintStep(stepNumber int, totalSteps int, description string)
    IsDebugEnabled() bool
}
```

## Design Decisions

### Registry Pattern vs. Direct Loading

The registry pattern was chosen over direct loading for several reasons:

1. **Centralized Management**: Single source of truth for workflow definitions
2. **On-Demand Loading**: Workflows are loaded only when needed
3. **Name Resolution**: Consistent naming for reference and retrieval
4. **Extensibility**: Easy to add new workflows from various sources

### External Definition Format

The decision to support both JSON and YAML formats was based on:

1. **User Preference**: Different users prefer different formats
2. **Readability**: YAML is more human-readable for complex structures
3. **Tool Compatibility**: JSON integrates well with other tools
4. **Flexibility**: Different use cases may benefit from different formats

### Thread-Safety Considerations

The implementation uses several concurrency safeguards:

1. **Synchronized Registry**: Using mutexes for thread-safe access
2. **Single Initialization**: Using `sync.Once` for thread-safe singleton creation
3. **Read-Write Locks**: Allowing concurrent reads but exclusive writes

### Error Handling Strategy

The error handling approach prioritizes user experience:

1. **Contextual Errors**: Each error includes context about what failed
2. **Graceful Fallbacks**: Using default values when possible
3. **Warning System**: Non-fatal issues generate warnings, not errors
4. **Consistent Messaging**: Error messages follow a consistent pattern

## Integration Points

### Command Line Interface

The workflow manager is integrated with CLI commands through constructor functions:

```go
// For custom workflows
NewWorkflowManager(fs, io, workflowName)

// For backward compatibility
NewDefaultWorkflowManager(fs, io)
```

### File System Integration

Workflow definitions can be loaded from any location accessible through the `FileSystem` interface:

```go
// Load from a single file
LoadWorkflowFromFile(fs, filePath)

// Load from a directory
LoadWorkflowsFromDirectory(fs, directory, registry)
```

### Serialization/Deserialization

The implementation provides bidirectional conversion between internal and external formats:

```go
// Internal to external (for saving)
ExternalWorkflowDefinition{...}

// External to internal (for loading)
ToWorkflowDefinition()
```

## Testing Approach

The implementation includes comprehensive testing:

1. **Unit Tests**: For individual functions like `validateExternalWorkflow()`
2. **Integration Tests**: For file operations and registry interactions
3. **Edge Cases**: Testing error conditions, invalid formats, and boundary values
4. **Mocking**: Using `MockFileSystem` for isolated filesystem testing

## Future Considerations

The current implementation lays a solid foundation for future enhancements:

1. **Workflow Versioning**: Supporting workflow evolution over time
2. **Dynamic Variables**: More sophisticated variable resolution
3. **Conditional Steps**: Steps that execute based on conditions
4. **Parallel Execution**: Support for concurrent step execution
5. **UI Integration**: Visual representation of available workflows

## Conclusion

The custom workflow implementation delivers a flexible, extensible system for defining and executing structured workflows. By separating the workflow definition from its execution and providing robust file-based loading, the system enables users to create and share custom workflows while maintaining backward compatibility with existing code. 
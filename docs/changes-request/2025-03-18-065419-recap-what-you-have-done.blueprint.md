---
name: Recap what you have done
created-at: 2025-03-18T06:54:19+01:00
user-stories:
  - title: Recap what you have done
    file: docs/user-stories/basic-commands/04-recap-what-you-have-done.md
    content-hash: d4297aa8e8199f3407a189bbe882f0ae

---

# Blueprint

## Overview

This is a change request for implementing the following user stories:
1. Recap what you have done


## Implementation Plan

### 1. Create the `recap` Command Structure

1. Create a new file `cmd/recap.go` that implements the `recap` command using the Cobra framework
2. Add the command to the root command in the `init()` function
3. Implement the logic to identify "incomplete" change requests

### 2. Core Functionality

1. Define a function to find incomplete change requests:
   - Scan the `docs/changes-requests` directory for files with `.blueprint.md` extension
   - For each blueprint file, check if there is a corresponding `.implementation.md` file
   - If no corresponding implementation file exists, mark it as an "incomplete" change request

2. Implement the selection logic:
   - If no incomplete change requests are found, display a fancy congratulation message
   - If exactly one incomplete change request is found, use it directly
   - If multiple incomplete change requests are found, present a selection interface for the user to choose one

3. Implement the recap display:
   - When a change request is selected, display the message:
     ```
     Recap what you did in a file in docs/changes-requests/<change-request-name>.implementation.md
     ```

### 3. UI Components

1. Use the `charmbracelet/lipgloss` and `charmbracelet/bubbles` libraries to create a beautiful UI:
   - Fancy congratulation message with appropriate styling
   - Interactive change request selection interface with descriptions
   - Clear and visually appealing output messages

### 4. Testing

1. Create unit tests for the core functionality in `cmd/recap_test.go`:
   - Test finding incomplete change requests
   - Test selection logic for different scenarios (none, one, or multiple incomplete change requests)
   - Test displaying the correct output messages

2. Implement mock interfaces for testing:
   - Use the existing mock file system for testing file operations
   - Create mock user interfaces for testing user interactions

### 5. Error Handling

1. Implement robust error handling:
   - Handle case where the change requests directory doesn't exist
   - Handle file reading/parsing errors
   - Provide clear error messages to guide the user

### 6. Documentation

1. Update documentation:
   - Add command usage examples to help text
   - Ensure clear inline documentation for all functions
   - Document error cases and how they're handled

## Implementation Details

### Function Structure

```go
// findIncompleteChangeRequests finds all change requests that have a blueprint file
// but no implementation file
func findIncompleteChangeRequests(fs io.FileSystem) ([]models.ChangeRequest, error) {
    // Implementation details
}

// formatChangeRequestDescription formats a change request for display in selection list
func formatChangeRequestDescription(cr models.ChangeRequest) string {
    // Implementation details
}

// displayCongratulationMessage displays a fancy congratulation message when no
// incomplete change requests are found
func displayCongratulationMessage(term io.UserOutput) {
    // Implementation details
}

// displayRecapMessage displays the recap message for the selected change request
func displayRecapMessage(term io.UserOutput, cr models.ChangeRequest) {
    // Implementation details
}
```

### Command Definition

```go
var recapCmd = &cobra.Command{
    Use:   "recap",
    Short: "Recap what you have done",
    Long:  `Recap what you have done by displaying incomplete change requests.`,
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation details
    },
}
```

## Acceptance Criteria Validation

- ✅ Command will be named `recap`
- ✅ Implementation will search for "incomplete" change requests (blueprint file without implementation file)
- ✅ Will display congratulation message when no incomplete change requests are found
- ✅ Will allow user selection when multiple incomplete change requests are found
- ✅ Will display appropriate recap message for the selected change request

This implementation plan provides a comprehensive approach to satisfy all the acceptance criteria specified in the user story.

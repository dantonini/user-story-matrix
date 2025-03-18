---
name: Implement a change request
created-at: 2025-03-18T07:30:12+01:00
user-stories:
  - title: Implement the change request
    file: docs/user-stories/basic-commands/05-implement-the-change-request.md
    content-hash: ecc0149e53cd4719c59e60624d4ca26b

---

# Blueprint

## Overview

This blueprint provides the implementation plan for the "Implement the change request" user story. The goal is to create a new CLI command that will instruct the AI-assistant editor to implement a change request from blueprint files.

## Implementation Plan

### Phase 1: Core Command Structure

1. **Create Implement Command File**
   - Create a new file `cmd/implement.go`
   - Define a basic Cobra command structure with appropriate descriptions
   - Register the command to the root command in the `init()` function

2. **Command Configuration**
   - Configure the command with the name `implement`
   - Add appropriate short and long descriptions
   - Implement the main command execution function

### Phase 2: Finding Incomplete Change Requests

1. **Develop the Incomplete Change Request Logic**
   - Reuse the `findIncompleteChangeRequests` function approach from the recap command
   - This function should identify change requests with a blueprint file but no implementation file
   - Handle different scenarios based on the number of found change requests:
     - No incomplete change requests
     - Exactly one incomplete change request
     - Multiple incomplete change requests

2. **User Selection Interface**
   - Implement selection mechanism when multiple incomplete change requests are found
   - Display clear options to the user with adequate information about each change request
   - Process user selection correctly

### Phase 3: Generating Implementation Instructions

1. **Create Output Generation**
   - Implement logic to generate the output message with the blueprint file path
   - Format the message with markdown links that can be used by AI-powered editors
   - Include the prompt to read the user stories, validate against codebase, and proceed with implementation

2. **Error Handling and User Feedback**
   - Add appropriate error handling for all possible failure scenarios
   - Provide clear feedback messages to the user
   - Log appropriate information for debugging purposes

### Phase 4: Testing

1. **Unit Tests**
   - Create tests for the `implement` command in `cmd/implement_test.go`
   - Test different scenarios:
     - No incomplete change requests
     - Single incomplete change request
     - Multiple incomplete change requests
   - Mock dependencies using the existing mock implementations

2. **Integration Testing**
   - Test the command with actual file system interactions
   - Verify that the command correctly identifies incomplete change requests
   - Ensure the output format is correct for AI-assistant editor interaction

## Detailed Implementation Steps

1. Create a new file `cmd/implement.go` with the following structure:
   - Define the `implementCmd` command
   - Add appropriate descriptions for the command
   - Implement the `Run` function to process the command
   - Register the command in the `init()` function

2. Implement the core logic for finding incomplete change requests:
   - Use the `findIncompleteChangeRequests` function as foundation (similar to recap command)
   - This function should search in the `docs/changes-request` directory
   - Identify files with `.blueprint.md` extension
   - Check if there's a corresponding `.implementation.md` file

3. Handle the three possible outcomes:
   - If no incomplete change requests are found: Display a sad message to the user
   - If exactly one incomplete change request is found: Directly use it
   - If multiple incomplete change requests are found: Display a selection prompt

4. Generate the appropriate output message:
   ```
   Read the blueprint file in [<change-request-name-blueprint>.md](mdc:full/file/path/to/<change-request-name-blueprint>.md)
   Read all the mentioned user stories, validate the blueprint against the code base and proceed with the implementation.
   ```

5. Add proper error handling for all operations:
   - Directory access issues
   - File reading problems
   - User interaction failures

6. Create comprehensive unit tests to verify the functionality

## Technical Details

The implementation will use:
- `github.com/spf13/cobra` for the command structure
- Internal `io` package for file operations and user interaction
- Internal `models` package for working with change request data
- Mock implementations for testing

The command behavior will closely match similar commands in the application, maintaining a consistent user experience.

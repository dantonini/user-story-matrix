---
name: Recap what you have done
created-at: 2025-03-18T06:54:19+01:00
user-stories:
  - title: Recap what you have done
    file: docs/user-stories/basic-commands/04-recap-what-you-have-done.md
    content-hash: b3d7314683efcbca09f0b82082744a0c2809eec17d2662d616edf92020a5d4f7

---

# Implementation of "Recap what you have done"

## Overview

This change request implemented the `recap` command which allows users to review their work by identifying incomplete change requests (those with a blueprint file but no implementation file).

## Implemented Components

### 1. Core `recap` Command

Created a new file `cmd/recap.go` with the following components:
- Main command definition with help text and examples
- Logic to find incomplete change requests
- Logic to handle different cases (no incomplete change requests, one, or multiple)
- User-friendly output messages including a fancy congratulation message

### 2. Helper Functions

Implemented several helper functions:
- `findIncompleteChangeRequests` - Scans the `docs/changes-request` directory to find incomplete change requests
- `formatChangeRequestDescription` - Formats the details of a change request for display in the selection list
- `displayCongratulationMessage` - Displays a fancy message when no incomplete change requests are found
- `displayRecapMessage` - Displays the appropriate recap message for a selected change request

### 3. Testing

Created comprehensive tests in `cmd/recap_test.go` that cover:
- Finding incomplete change requests under various conditions
- The selection process for change requests
- The display of appropriate messages

Added a `MockUserIO` implementation in `internal/io/mock_prompt.go` to facilitate testing with testify/mock assertions.

## Files Modified

1. `cmd/recap.go` - Created new file with the recap command implementation
2. `cmd/recap_test.go` - Created new file with tests for the recap command
3. `internal/io/mock_prompt.go` - Added MockUserIO for testify/mock-based testing

## Acceptance Criteria Fulfillment

- ✅ Created command named `recap`
- ✅ Implemented logic to find "incomplete" change requests (blueprint file without implementation file)
- ✅ Added fancy congratulation message when no incomplete change requests are found
- ✅ Implemented user selection when multiple incomplete change requests are found
- ✅ Displays appropriate recap message for the selected change request

## Usage Example

```
usm recap
```

When run, this command will:
1. Look for change requests with a blueprint file but no implementation file
2. If none are found, display a congratulation message
3. If multiple are found, allow the user to select one
4. Display a message instructing the user where to write the implementation 
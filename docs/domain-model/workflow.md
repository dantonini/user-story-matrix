# USM Workflow

This document describes the typical workflow and lifecycle of entities in the User Story Matrix (USM) tool.

## Overview

The USM workflow is designed to bring structure, repeatability, and control to AI-assisted development. It follows these general steps:

1. **Define user stories** - Create structured user story files that capture requirements
2. **Create change requests** - Group user stories into change requests with implementation blueprints
3. **Implement changes** - Use AI assistance to implement the changes defined in the blueprint
4. **Track and review** - Document the implementation and review changes

## Step 1: Define User Stories

User stories are created using the `usm add user-story` command:

```bash
usm add user-story --into docs/user-stories/my-feature-area
```

This will:
1. Prompt for the user story title
2. Create a new markdown file with the correct structure
3. Generate metadata including creation time and content hash
4. Save the file in the specified directory with a sequential number prefix

The user then edits the file to add:
- A description of the user story
- The "As a... I want... so that..." format
- Acceptance criteria

## Step 2: Create Change Requests

Change requests are created using the `usm create change-request` command:

```bash
usm create change-request --from docs/user-stories/my-feature-area
```

This will:
1. List all user stories in the specified directory
2. Allow the user to select which stories to include in the change request
3. Prompt for a change request name
4. Create a blueprint file with references to the selected user stories
5. Include metadata such as creation time and content hashes of the referenced user stories

The blueprint file (with `.blueprint.md` extension) contains:
- Metadata about the change request and referenced user stories
- A structured section for the plan
- References to help track which user stories are being implemented

The command will output an ai prompt that can be used to generate the blueprint.
You can copy & paste the prompt into an AI chat (Cursor, windsurf or similar) to generate the blueprint.

## Step 3: Fill the Blueprint

The AI should:
1. Read the user stories referenced in the change request
2. Analyze the requirements and acceptance criteria
3. Generate a detailed implementation plan in the blueprint file
4. Include technical details, data structures, and implementation steps using a pseudo code style

You have to review the plan and ask for changes if needed or of course modify it by yourself.
Iterate until you are happy with the plan.

## Step 4: Implement Changes

With the implementation blueprint in place, you can start implementing the changes:
Usually you work on a change request at a time hence you would implement the blueprinted change request by running:

```bash
usm implement
```
The command will output an ai prompt that can be used to implement the blueprint.
You can copy & paste the prompt into an AI chat (Cursor, windsurf or similar) in agent mode to start the implementation.

The AI will:
1. Create new files
2. Update existing files 

## Step 5: Review Changes

Use the command `usm recap` to create a summary of the changes.


The AI will:
1. Create an implementation file (with `.implementation.md` extension) with a summary of the changes
2. Document the actual implementation details
3. Include references to code changes and implementation decisions

## Integrity Verification

USM uses content hashes to track and verify the integrity of user stories:

1. When a user story is created, a hash of its content is stored in its metadata
2. When a change request references a user story, it stores the hash at that point in time
3. If a user story changes after being referenced, the change request can detect this by comparing hashes

This ensures that:
- You're implementing the requirements that were specified at the time the change request was created
- You can detect if requirements change during implementation
- You have a clear audit trail of changes over time

## File Lifecycle

The lifecycle of files in USM is as follows:

1. **User Story Files**:
   - Created with the `add user-story` command
   - Can be updated manually to refine requirements
   - Referenced by change requests

2. **Change Request Blueprint Files**:
   - Created with the `create change-request` command
   - References user stories at a point in time
   - Can be enhanced with AI-generated implementation plans
   - Serves as a specification for implementation

3. **Change Request Implementation Files**:
   - Created when implementing a change request
   - Documents the actual implementation details
   - Can include references to code changes and decisions
   - Serves as documentation of what was actually done

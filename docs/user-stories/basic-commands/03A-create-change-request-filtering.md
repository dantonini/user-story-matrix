---
file_path: docs/user-stories/basic-commands/03A-create-change-request-filtering.md
created_at: 2025-03-24T10:00:00+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: ff77a8252e9c3b809422df613e6daa60bcccf29b4380ea76e73cd115219fce70
---

# Enhanced Change Request Creation with Filtering

As a CLI user,  
I want to filter and search through user stories when creating a change request,  
so that I can quickly find relevant stories and focus on unimplemented features.

## Acceptance Criteria

### Implementation Status Filter
- When creating a change request, the CLI should:
  - By default, only show unimplemented user stories
  - Provide a flag `--show-all` to display all user stories regardless of implementation status
  - Clearly indicate which stories are implemented and which are not in the selection UI

### Search and Filter Capabilities
- The CLI should provide an interactive search feature:
  - As the user types, the list of displayed user stories should be filtered in real-time
  - The search should match against:
    - User story titles
    - User story descriptions
    - Acceptance criteria content
  - The filtering should be case-insensitive
  - The filtering should support partial word matches

### User Interface
- The selection UI should:
  - Show the total number of stories and how many are being displayed after filtering
  - Maintain the multi-select capability from the original feature
  - Allow clearing the search filter
  - Provide clear visual distinction between implemented and unimplemented stories
  - Show a "no results" message when the filter returns no matches

### Integration
- The feature should integrate seamlessly with the existing change request creation flow
- All selected stories should be included in the change request metadata as before
- The command should maintain compatibility with the `--from` directory option 
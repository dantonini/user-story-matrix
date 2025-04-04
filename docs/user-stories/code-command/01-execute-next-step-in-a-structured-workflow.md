---
file_path: docs/user-stories/code-command/01-execute-next-step-in-a-structured-workflow.md
created_at: 2025-03-26T01:43:18+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: c98843b09b524f5d6dd5c3aafce278d99961481470980d400b3d9758c7daaf71
---

# Execute next step in a structured workflow  
As a  
developer using USM,  
I want  
the `code` command to execute the next pending step in a predefined, numbered workflow based on a change request file,  
So that  
I can process development tasks incrementally in a clear and reproducible manner.

### Acceptance Criteria
- The workflow consists of 8 numbered steps:
  1. `01-laying-the-foundation`
  2. `01-laying-the-foundation-test`
  3. `02-mvi`
  4. `02-mvi-test`
  5. `03-extend-functionalities`
  6. `03-extend-functionalities-test`
  7. `04-final-iteration`
  8. `04-final-iteration-test`
- The command detects the next pending step and executes only that step.
- Step progress is persisted in a `.step` file.
- Each completed step produces a file in the format: `<input-filename>.<step>.md`.

---
file_path: docs/user-stories/code-command/08-introduce-prompt-in-existing-step-definition.md
created_at: 2025-03-31T08:13:51+02:00
last_updated: 2025-03-31T08:13:51+02:00
_content_hash: 55419763d3973fbbe3d750e40669f163
---

# Introduce Prompt in Existing Step Definition
As a USM maintainer
I want to extend the current step definition by introducing a prompt field (in addition to description)
so that steps can provide actionable instructions to the AI agent

## Acceptance criteria
- The step definition now includes a new prompt attribute.
- The prompt supports variable interpolation, with the current implementation allowing (optinally) only the basic variable change_request_file_path
- The implementation is designed to be extendable so that additional variables can be supported in the prompt later on.

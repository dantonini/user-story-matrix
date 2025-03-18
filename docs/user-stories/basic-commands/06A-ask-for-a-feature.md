---
file_path: docs/user-stories/basic-commands/06A-ask-for-a-feature.md
created_at: 2025-03-18T09:06:34+01:00
last_updated: 2025-03-18T19:02:03+01:00
_content_hash: dec29cebc60a724852f9ed8babe37ff2
---

# Submit a Feature Request

As a user
I want to be able to submit a feature suggestion to the CLI developer
So that I can suggest a new feature for my use case

## Acceptance Criteria
- The command should allow users to submit a feature request.
- The user should provide structured input including:
  - The title of the feature
  - A description of the feature
  - The reason why this feature is important to the user
  - A structured user story following the format: "As a ... I want ... so that ..." (required)
  - Acceptance criteria for the feature
- The command should ask for confirmation before submitting the request.
- The feature request should be sent via Slack using the webhook: https://hooks.slack.com/services/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
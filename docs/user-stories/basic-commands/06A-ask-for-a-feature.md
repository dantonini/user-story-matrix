---
file_path: docs/user-stories/basic-commands/06A-ask-for-a-feature.md
created_at: 2025-03-18T09:06:34+01:00
last_updated: 2025-05-23T18:00:43+02:00
_content_hash: 7743f26f40dafa0fd9bf673c2e0eaaef8ce05ca837a0ffa05470ed30006a2604
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
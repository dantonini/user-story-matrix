---
file_path: docs/user-stories/setup/10-github.md
created_at: 2025-03-17T08:28:02+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 9fdc600b8ca421ea2ff77e6818692bbc265fe9c0afb3cb6404428481de511e97
---

# Create GitHub Actions for Automated Builds
The project should have automated builds on GitHub.

As a developer,  
I want GitHub Actions to build and test the CLI automatically,  
so that I can ensure code quality and consistency.

## Acceptance criteria

- A GitHub Actions workflow is created in `.github/workflows/build.yml`.
- The workflow builds the CLI for Linux, macOS, and Windows.
- The workflow runs tests before completing the build.
- The workflow triggers on every push to `main`.
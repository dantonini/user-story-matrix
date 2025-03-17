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
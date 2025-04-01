# USM – User Story Matrix
*A developer-first CLI to manage, generate and orchestrate AI-powered user stories.*

USM (User Story Matrix) is a command-line tool designed to bring structure, repeatability, and control to your AI-assisted development workflow.

If you're using AI tools like Cursor or Windsurf to write code, you've probably hit some limits: unclear prompts, inconsistent results, or code that kind of works... but not really.  
USM helps you **break down development into manageable, testable units** – user stories – and build a consistent flow around them:

- Define and organize user stories.
- Generate implementation blueprints.
- Apply structured prompts to AI tools.
- Track and review change requests.
- Execute implementation workflows in predictable, incremental steps.

You can think of USM as a lightweight orchestration layer between you and your AI assistant.  
It doesn't do the coding *for* you. It helps you code **with** AI – deliberately, safely, and at scale.

Whether you're working solo or in a team, USM gives you a repeatable process to make AI coding less chaotic and more productive.

## Why the name?
The name User Story Matrix reflects the idea of organizing and navigating multiple user stories thematically — like rows and columns of a matrix — to give structure and clarity to AI-assisted development.

# Installation

## Binary Releases

Download the latest binary for your platform from the [Releases](https://github.com/dantonini/usm/releases) page.

### Linux/macOS

```bash
# Download the latest release (replace X.Y.Z with the version)
curl -L https://github.com/dantonini/usm/releases/download/vX.Y.Z/usm-linux-amd64-X.Y.Z -o usm
chmod +x usm
./usm
```

### Windows

Download the executable from the [Releases](https://github.com/dantonini/usm/releases) page and run it from the command prompt.


## From Source

Prerequisites:

- Go 1.21 or higher

```bash
# Clone the repository
git clone https://github.com/dantonini/usm.git
cd usm

# Build the binary
make build

# Run the binary
./usm
```

# Shell Completion

The USM provides shell completion support for Bash, Zsh, Fish, and PowerShell.

## Zsh

Add this to your `~/.zshrc` file:

```bash
# Add usm completion to your shell
source <(usm completion zsh)

# If you have compinit disabled, you can use the following instead:
usm completion zsh > "${fpath[1]}/_usm"
```

## Bash

```bash
# Linux
usm completion bash > /etc/bash_completion.d/usm

# macOS (with homebrew)
usm completion bash > $(brew --prefix)/etc/bash_completion.d/usm

# Or directly to your ~/.bashrc
echo 'source <(usm completion bash)' >> ~/.bashrc
```

## Fish

```bash
usm completion fish > ~/.config/fish/completions/usm.fish
```

## PowerShell

```powershell
usm completion powershell > usm.ps1
```

# Usage

```bash
# Show help
usm --help

# Execute the next step in a structured implementation workflow
usm code docs/changes-request/my-change-request.blueprint.md
```

## Managing User Stories

### Adding a User Story

```bash
# Add a new user story (will be saved in docs/user-stories)
usm add user-story

# Add a user story to a specific directory
usm add user-story --into docs/user-stories/my-feature
```

### Listing User Stories

```bash
# List all user stories in the default directory
usm list user-stories

# List user stories from a specific directory
usm list user-stories --from docs/user-stories/my-feature
```

## Managing Change Requests

### Creating a Change Request

```bash
# Create a change request (interactively select user stories)
usm create change-request

# Create a change request from user stories in a specific directory
usm create change-request --from docs/user-stories/my-feature
```

### Implementing a Change Request

```bash
# Navigate through a structured implementation process for a change request
usm code docs/changes-request/my-change-request.blueprint.md

# Reset the implementation workflow and start from the beginning
usm code --reset docs/changes-request/my-change-request.blueprint.md
```

> **Note:** The `code` command is currently a proof-of-concept and will be extended with more advanced AI integration capabilities in upcoming releases. It provides a structured workflow with 4 predefined steps:
    - laying the foundation
    - minimal viable implementation
    - extend functionalities
    - final iteration

# Project Structure

- `docs/user-stories/`: Contains the user stories used to develop USM itself. This folder showcases how USM structures and manages its own development flow.
- `docs/changes-requests/`: Stores change request files generated from one or more user stories. These represent scoped implementation plans and the context for AI-assisted coding.
- `cmd/`: Entrypoint commands for the CLI. Each subcommand (e.g. add, list, create) is defined here.
- `internal/`: Internal packages and logic used by the CLI. This includes core functionalities such as user story parsing, change request generation, file handling, and prompt orchestration.


# Development

## Setup

```bash
# Install dependencies
make deps

# Install pre-commit hooks (recommended for developers)
make install-hooks
```

## Code Quality

USM uses [golangci-lint](https://golangci-lint.run/) for static code analysis to maintain code quality.

### Available Linters

The following essential linters are enabled by default:

- **unused**: Finds unused code (replaces deprecated 'deadcode' linter in newer versions)
- **errcheck**: Checks for unchecked errors
- **govet**: Reports suspicious constructs
- **staticcheck**: Applies static analysis checks

### Linting Commands

```bash
# Run linters only
make lint

# Run a full build with linting
make build-full

# Standard build (no linting)
make build

# Find and report dead/unused code
make lint-fix-deadcode
```

### Pre-commit Hook

USM includes a lightweight pre-commit hook that runs fast linters on changed files without blocking your commits.

To install it:

```bash
make install-hooks
```

### Configuration

The linting configuration is defined in `.golangci.yml` in the project root. This minimal configuration is designed to be non-intrusive while still catching important issues.

### Version Compatibility

USM handles linter compatibility automatically:
- For golangci-lint < v1.49.0: Uses the 'deadcode' linter
- For golangci-lint >= v1.49.0: Uses the 'unused' linter (which replaces the deprecated 'deadcode')

## Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Generate HTML coverage report
make coverage-html

# Show functions with less than 100% coverage
make coverage-report
```

## Code Coverage

```bash
# Basic coverage report showing function coverage percentage
make test-coverage

# Detailed report focusing on uncovered code
make coverage-report

# Generate an HTML report to visualize coverage
make coverage-html

# For a comprehensive analysis with code context
./coverage.sh
```

The `coverage.sh` script will:
1. Run all tests with coverage tracking
2. Show overall coverage percentage
3. List functions with less than 100% coverage
4. Display actual uncovered code blocks with context
5. Generate an HTML report you can open in your browser
6. Suggest areas to focus your testing efforts on

The HTML report provides a visual way to see which lines of code are covered (green), not covered (red), or not executable (gray).

## Releasing

USM uses an automated release process with GitHub Actions:

### Automated Release (Recommended)

Run the release script which will automatically:
1. Check for uncommitted changes
2. Increment the patch version (or use a specified version)
3. Commit the changes to Makefile
4. Tag and push to trigger the release workflow

```bash
# Auto-increment patch version
./release.sh

# Or specify a version
./release.sh 1.2.3
```

### Manual Process

1. Update the version in `Makefile` and any documentation
2. Create a release branch, make a PR, and merge to main
3. Tag the version on the main branch:
   ```bash
   git checkout main
   git pull
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```
4. The GitHub Actions workflow will automatically build binaries, create a GitHub release, and upload assets

For more details, see [RELEASE.md](RELEASE.md).

## Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all
```

# Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

# Feature request

Have an idea, suggestion or found something confusing?

You can contribute feedback the same way you interact with the tool: by submitting a user story 😄

```bash
usm ask feature
```

This will guide you in writing a feature request as a user story and send it to me directly.

Alternatively, feel free to open an issue or start a discussion here on GitHub.
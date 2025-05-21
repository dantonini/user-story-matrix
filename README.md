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

Download the latest binary for your platform from the [Releases](https://github.com/dantonini/user-story-matrix/releases) page.

### Linux/macOS

```bash
# Download the latest release (replace X.Y.Z with the version)
curl -L https://github.com/dantonini/user-story-matrix/releases/download/vX.Y.Z/usm-linux-amd64-X.Y.Z -o usm
chmod +x usm
./usm
```

### Windows

Download the executable from the [Releases](https://github.com/dantonini/user-story-matrix/releases) page and run it from the command prompt.


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

## Development Tools

USM includes several tools to help maintain code quality and consistency.

### Code Clone Detection with dupl

USM integrates [dupl](https://github.com/mibk/dupl), a tool for finding code clones. This helps identify potential refactoring opportunities and maintain a DRY codebase.

```bash
# Run dupl to find code clones with a threshold of 100 tokens
make dupl

# Generate an HTML report of code clones
make dupl-html

# Check only git-modified files for code clones (both modified and new files)
make dupl-modified

# Generate HTML report only for git-modified files
make dupl-modified-html

# Adjust the token threshold (higher = fewer, more significant clones)
dupl -t 200 -plumbing .
```

The HTML reports will be saved to:
- Full project: `output/reports/dupl-report.html`
- Modified files only: `output/reports/dupl-modified-report.html`

### Dead Code Detection

```bash
# Run dead code detection (dry run)
make deadcode

# Remove detected dead code
make deadcode-remove
```

### Code Linting

```bash
# Run all linters
make lint

# Run linters only on test files
make lint-tests
```

## Managing User Stories

### Adding a User Story

```bash
# Add a new user story (will be saved in docs/user-stories)
usm add user-story

# Add a user story to a specific directory
usm add user-story --into docs/user-stories/my-feature

# Add a user story without LLM processing
usm add user-story --no-llm
```

When adding a user story, USM now features intelligent processing of pasted text. You can simply paste unstructured text (like requirements, emails, or notes) into the form, and USM will use OpenAI to automatically parse and structure the content into the appropriate form fields.

### Using the LLM Paste Processor

To use the LLM paste processor, you need to configure your OpenAI API key:

```bash
# Set your OpenAI API key
usm settings api-key openai YOUR_API_KEY

# Check your current API key status
usm settings api-key openai
```

Once configured, any text you paste into the user story form will be automatically processed, saving you time and ensuring more consistent user stories.

**Note:** The LLM processing is enabled by default. Use the `--no-llm` flag to disable it if needed.

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

### Elaborating User Stories from Vague Descriptions

```bash
# Start with interactive input
usm elaborate

# Use an existing vague description file
usm elaborate docs/vague/my-feature.md

# Reset the workflow and start from the beginning
usm elaborate --reset docs/vague/my-feature.md
```

The `elaborate` command transforms vague feature descriptions into well-defined user stories following a structured 7-step workflow:

1. Analyze the description
2. Extract user personas
3. Draft initial user stories
4. Refine stories to meet INVEST criteria
5. Prioritize user stories
6. Add acceptance criteria
7. Final review and packaging

For more details, see the [Elaborate Command Documentation](docs/elaborate-command.md).

### Prompt interpolated variables

USM supports flexible prompt interpolation:

```bash
# Standard variables available in prompts:
${change_request_basename}   # The basename of the change request file (without extension)
${blueprint_basename}        # The basename of the blueprint file (without extension)
${change_request_dirname}    # The directory containing the change request file
${stepid}                    # The ID of the current workflow step
${stepname}                  # The name part of the step ID (e.g., "laying-the-foundation" from "01-laying-the-foundation")
${change_request_fullpath}   # The full path of the change request file (without extension)
${change_request_file_path}  # Full path to the change request file (includes extension)
```

These variables can also be used in workflow step prompts to create dynamic instructions for AI agents:

```
You are reviewing implementation for change request ${change_request_basename}.
Current step: ${stepname}
The full change request is located at: ${change_request_file_path}

Please focus on the following aspects...
```

For more details, see [Variable Interpolation in USM](docs/changes-request/output-file-path-interpolation.md).

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
make deadcode
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
# Run tests with coverage enabled by default
make test
```

## Code Coverage

```bash
# Basic coverage report showing function coverage percentage
make coverage

# Show coverage for a specific file
./coverage -file <path-to-file> 
```

The `coverage` command will show detailed coverage per file.

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

## Development

### Prerequisites
- Go 1.22 or newer
- Git

### Development Commands
USM offers several commands to help you during development:

```bash
# Build the tool
make build

# Run tests
make test

# Build with all tests and checks
make build-full

# Run linters to detect code issues
make lint

# Detect dead code
make deadcode

# Generate coverage report
make coverage-report
```

### Dead Code Detection
The project includes a tool for identifying and removing unused code:

```bash
# Show what would be removed (dry run)
make deadcode

# Actually remove dead code (creates backups first)
make deadcode-remove

# With more verbose output
./scripts/deadcode.sh --verbose

# Focus on a specific directory
./scripts/deadcode.sh --path ./cmd/...
```

This AST-based tool analyzes the code using Abstract Syntax Trees to properly:
- Remove entire functions, variables, and types
- Preserve code structure and formatting
- Create automatic backups of modified files
- Maintain correct Go syntax
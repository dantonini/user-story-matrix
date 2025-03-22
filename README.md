# USM – User Story Matrix
*A developer-first CLI to manage, generate and orchestrate AI-powered user stories.*

USM (User Story Matrix) is a command-line tool designed to bring structure, repeatability, and control to your AI-assisted development workflow.

If you're using AI tools like Cursor or Windsurf to write code, you’ve probably hit some limits: unclear prompts, inconsistent results, or code that kind of works... but not really.  
USM helps you **break down development into manageable, testable units** – user stories – and build a consistent flow around them:

- Define and organize user stories.
- Generate implementation blueprints.
- Apply structured prompts to AI tools.
- Track and review change requests.

You can think of USM as a lightweight orchestration layer between you and your AI assistant.  
It doesn’t do the coding *for* you. It helps you code **with** AI – deliberately, safely, and at scale.

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
```

## Testing

```bash
# Run tests
make test
```

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
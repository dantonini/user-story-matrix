# User Story Matrix CLI (USM-CLI)

A command-line tool for managing user stories and organizing them in a matrix format for better visualization and planning.

## Installation

### Prerequisites

- Go 1.21 or higher

### From Source

```bash
# Clone the repository
git clone https://github.com/user-story-matrix/usm-cli.git
cd usm-cli

# Build the binary
make build

# Run the binary
./usm
```

### Binary Releases

Download the latest binary for your platform from the [Releases](https://github.com/user-story-matrix/usm-cli/releases) page.

#### Linux/macOS

```bash
# Download the latest release (replace X.Y.Z with the version)
curl -L https://github.com/user-story-matrix/usm-cli/releases/download/vX.Y.Z/usm-linux-amd64-X.Y.Z -o usm
chmod +x usm
./usm
```

#### Windows

Download the executable from the [Releases](https://github.com/user-story-matrix/usm-cli/releases) page and run it from the command prompt.

## Shell Completion

The USM CLI provides shell completion support for Bash, Zsh, Fish, and PowerShell.

### Zsh

Add this to your `~/.zshrc` file:

```bash
# Add usm completion to your shell
source <(usm completion zsh)

# If you have compinit disabled, you can use the following instead:
usm completion zsh > "${fpath[1]}/_usm"
```

### Bash

```bash
# Linux
usm completion bash > /etc/bash_completion.d/usm

# macOS (with homebrew)
usm completion bash > $(brew --prefix)/etc/bash_completion.d/usm

# Or directly to your ~/.bashrc
echo 'source <(usm completion bash)' >> ~/.bashrc
```

### Fish

```bash
usm completion fish > ~/.config/fish/completions/usm.fish
```

### PowerShell

```powershell
usm completion powershell > usm.ps1
```

## Usage

```bash
# Show help
usm --help

# Enable debug mode
usm --debug <command>
```

### Managing User Stories

#### Adding a User Story

```bash
# Add a new user story (will be saved in docs/user-stories)
usm add user-story

# Add a user story to a specific directory
usm add user-story --into docs/user-stories/my-feature
```

#### Listing User Stories

```bash
# List all user stories in the default directory
usm list user-stories

# List user stories from a specific directory
usm list user-stories --from docs/user-stories/my-feature
```

### Managing Change Requests

#### Creating a Change Request

```bash
# Create a change request (interactively select user stories)
usm create change-request

# Create a change request from user stories in a specific directory
usm create change-request --from docs/user-stories/my-feature
```

## Project Structure

- `docs/user-stories/`: Default directory for user stories
- `docs/changes-request/`: Directory for change request files
- `cmd/`: Command-line interface commands
- `internal/`: Internal packages

## Development

### Setup

```bash
# Install dependencies
make deps
```

### Testing

```bash
# Run tests
make test
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
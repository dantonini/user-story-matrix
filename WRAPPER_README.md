# USM Wrapper (usmw)

A wrapper script for User Story Matrix CLI, similar to Maven Wrapper (`mvnw`) and Gradle Wrapper (`gradlew`). The wrapper automatically downloads and manages USM binary versions, ensuring you always have the right version without manual installation.

## Features

- **Automatic Version Management**: Downloads and updates USM binary automatically
- **Platform Detection**: Supports Linux, macOS, and Windows
- **Architecture Support**: Works with AMD64 and ARM64 architectures
- **Version Checking**: Automatically checks for updates against GitHub releases
- **Safe to Commit**: The wrapper scripts can be safely committed to git, while binaries are stored locally
- **Offline Capable**: Once downloaded, works offline until an update is needed
- **Transparent Operation**: Runs silently by default, shows debug info only when requested

## Installation

Copy the wrapper scripts to your project root:

```bash
# For Unix/Linux/macOS
curl -L -o usmw https://raw.githubusercontent.com/dantonini/user-story-matrix/main/usmw
chmod +x usmw

# For Windows
curl -L -o usmw.ps1 https://raw.githubusercontent.com/dantonini/user-story-matrix/main/usmw.ps1
```

## Usage

### Unix/Linux/macOS

```bash
# Basic usage - automatically downloads USM if needed
./usmw --help
./usmw elaborate

# Wrapper-specific commands
./usmw --usmw-version      # Show wrapper and USM version info
./usmw --usmw-help         # Show wrapper help
./usmw --usmw-update       # Force update to latest version
./usmw --usmw-clean        # Clean cache files
./usmw --usmw-debug        # Enable debug output for troubleshooting
```

### Windows

```powershell
# Basic usage - automatically downloads USM if needed
.\usmw.ps1 --help
.\usmw.ps1 elaborate

# If execution policy restricts script execution
powershell -ExecutionPolicy Bypass -File usmw.ps1 --help

# Wrapper-specific commands
.\usmw.ps1 --usmw-version      # Show wrapper and USM version info
.\usmw.ps1 --usmw-help         # Show wrapper help
.\usmw.ps1 --usmw-update       # Force update to latest version
.\usmw.ps1 --usmw-clean        # Clean cache files
.\usmw.ps1 --usmw-debug        # Enable debug output for troubleshooting
```

## How It Works

1. **First Run**: The wrapper detects your platform, downloads the latest USM binary from GitHub releases, and stores it in `~/.usm/bin/`
2. **Subsequent Runs**: The wrapper checks if you have the latest version and updates if necessary
3. **Command Execution**: All arguments are passed through to the USM binary transparently
4. **Debug Mode**: Use `--usmw-debug` to see what the wrapper is doing behind the scenes

## Directory Structure

The wrapper creates the following directory structure:

```
~/.usm/
├── bin/
│   └── usm[.exe]           # The actual USM binary
├── cache/                  # Temporary download files
└── version                 # Current version tracking file
```

## Configuration

### Environment Variables

- `USM_HOME`: Override the default home directory (default: `~/.usm`)

### Customization

```bash
# Use a custom USM home directory
export USM_HOME="/opt/usm"
./usmw --usmw-version
```

## Version Management

The wrapper automatically:

- Fetches the latest release information from GitHub API
- Downloads the appropriate binary for your platform
- Stores version information locally
- Updates when a new version is available

### Manual Version Control

```bash
# Force update to latest version
./usmw --usmw-update

# Check version information
./usmw --usmw-version

# Clean downloaded cache
./usmw --usmw-clean
```

## Platform Support

| Platform | Architecture | Binary Name | Status |
|----------|-------------|-------------|---------|
| Linux | AMD64 | `usm-linux-amd64-{version}` | ✅ Supported |
| Linux | ARM64 | `usm-linux-arm64-{version}` | ✅ Supported |
| macOS | AMD64 | `usm-darwin-amd64-{version}` | ✅ Supported |
| macOS | ARM64 | `usm-darwin-arm64-{version}` | ✅ Supported |
| Windows | AMD64 | `usm-windows-amd64-{version}.exe` | ✅ Supported |

## Troubleshooting

### Common Issues

1. **Network Issues**
   ```bash
   # Check if you can reach GitHub
   curl -I https://api.github.com/repos/dantonini/user-story-matrix/releases/latest
   ```

2. **Permission Issues**
   ```bash
   # Ensure wrapper is executable
   chmod +x usmw
   ```

3. **Windows Execution Policy**
   ```powershell
   # Allow script execution temporarily
   Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
   .\usmw.ps1 --help
   ```

### Debug Mode

The wrapper runs silently by default for a transparent experience. Use the built-in debug mode for troubleshooting:

```bash
# Show what the wrapper is doing
./usmw --usmw-debug --help

# Debug mode can be used with any command
./usmw --usmw-debug elaborate

# Or just enable debug mode alone to see wrapper help
./usmw --usmw-debug
```

Debug mode shows:
- Platform detection
- Version checking
- Download progress
- Installation steps

### Manual Installation Fallback

If the wrapper fails, you can manually install USM:

```bash
# Download manually and place in ~/.usm/bin/
mkdir -p ~/.usm/bin
curl -L -o ~/.usm/bin/usm "https://github.com/dantonini/user-story-matrix/releases/latest/download/usm-linux-amd64-0.1.8"
chmod +x ~/.usm/bin/usm
echo "0.1.8" > ~/.usm/version
```

## Contributing

The wrapper scripts are part of the main USM repository. To contribute:

1. Fork the repository
2. Make your changes to `usmw` and/or `usmw.ps1`
3. Test on your platform
4. Submit a pull request

## License

This wrapper follows the same license as the User Story Matrix project (MIT License).

## Related Links

- [User Story Matrix Repository](https://github.com/dantonini/user-story-matrix)
- [Latest Releases](https://github.com/dantonini/user-story-matrix/releases/latest)
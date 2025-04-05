# USM Scripts

This directory contains various utility scripts for the USM project.

## Dead Code Detection and Removal

The `deadcode.sh` script helps you identify and safely remove unused code from the project.

### Usage

```bash
# Basic scan (dry run)
./deadcode.sh

# Interactive mode (prompts before removing code)
./deadcode.sh --interactive

# Show detailed output
./deadcode.sh --verbose

# Dry run with detailed output
./deadcode.sh --dry-run --verbose

# Show help
./deadcode.sh --help
```

### Features

- Identifies unused variables, constants, functions, and types
- Interactive mode for safe code removal
- Creates backups of modified files
- Integrates with golangci-lint and staticcheck
- Color-coded output for better readability

### Safety Measures

1. Always runs in dry-run mode by default
2. Creates backups of all modified files
3. Interactive mode lets you review each potential removal
4. Comprehensive testing is recommended after any code removal

For large-scale refactoring, consider:
1. Running with `--dry-run` first
2. Making a git commit before removing code
3. Running tests after removal
4. Using the interactive mode for complex cases

## Other Scripts 
# AST-Based Dead Code Remover

This utility uses Go's Abstract Syntax Tree (AST) processing to safely and completely remove unused code from your Go files.

## Features

- Accurately detects unused variables, functions, types, and constants
- Removes entire unused code blocks
- Preserves code formatting, comments, and structure
- Creates backups of modified files
- Supports dry-run mode for safe testing
- Works with both staticcheck and golangci-lint

## Installation

```bash
# From the project root
go build -o usm-deadcode ./cmd/deadcode
```

## Usage

```bash
# Scan all packages in dry-run mode
./usm-deadcode --dry-run

# Verbose output
./usm-deadcode --verbose

# Specify packages to scan
./usm-deadcode --packages="./cmd/...,./internal/..."

# Actually remove dead code (creates backups first)
./usm-deadcode

# Custom backup directory
./usm-deadcode --backup-dir=./my-backups
```

## How It Works

1. Uses staticcheck/golangci-lint to identify unused code
2. Parses each file into an AST (Abstract Syntax Tree)
3. Removes unused declarations from the AST
4. Regenerates the source file from the modified AST
5. Maintains all formatting, imports, and comments

## Options

| Flag | Description |
|------|-------------|
| `--dry-run` | Don't modify files, just show what would be done |
| `--verbose` | Show detailed output about processing |
| `--packages` | Comma-separated list of packages to scan (default: "./...") |
| `--backup-dir` | Directory to store backups (default: ".deadcode-backups-TIMESTAMP") |

## Examples

Remove unused code from a specific test file:

```bash
./usm-deadcode --packages="./test/deadcode_example/..."
```

Scan all code with detailed output:

```bash
./usm-deadcode --dry-run --verbose
```

## Integration with Shell Script

The tool is integrated with the shell script for easier usage:

```bash
# Dry run mode
./scripts/deadcode.sh --dry-run

# Verbose output
./scripts/deadcode.sh --verbose

# Run on specific packages
./scripts/deadcode.sh --path="./cmd/..."
```

## Advantages over Line-by-Line Removal

1. **Whole Function/Type Removal**: Removes entire unused functions or types, not just declarations
2. **Preserves Structure**: Maintains correct Go syntax after removal
3. **AST Awareness**: Understands Go's language constructs properly
4. **Format Preservation**: Maintains code style and formatting 
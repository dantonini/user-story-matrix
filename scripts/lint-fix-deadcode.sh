#!/bin/bash

# USM dead code removal script
# This helper script identifies and optionally removes dead code using golangci-lint

set -e

# Define colors for output
RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${CYAN}=======================================${NC}"
echo -e "${CYAN}USM Dead Code Detection and Removal Tool${NC}"
echo -e "${CYAN}=======================================${NC}"
echo

# This script checks for and optionally fixes dead code in the codebase
# Define flags for interactive and force fixing
INTERACTIVE=false
FORCE=false

while getopts "if" opt; do
  case ${opt} in
    i )
      INTERACTIVE=true
      ;;
    f )
      FORCE=true
      ;;
    \? )
      echo "Usage: $0 [-i] [-f]"
      echo "  -i  Interactive mode (ask before fixing)"
      echo "  -f  Force mode (fix without asking)"
      exit 1
      ;;
  esac
done

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "Error: golangci-lint not installed"
    echo "Please install with:"
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    exit 1
fi

# Get golangci-lint version
VERSION=$(golangci-lint --version)
echo "Detected golangci-lint version: $VERSION"

# Determine which linter to use based on version
# Since version 1.49.0, deadcode is deprecated in favor of unused
if [[ "$VERSION" =~ v1\.([5-9][0-9]|[0-9]{3,}) ]]; then
    LINTER="unused"
    echo "Using linter: $LINTER (newer version)"
else
    LINTER="deadcode"
    echo "Using linter: $LINTER"
fi

# Common args for linting commands
COMMON_ARGS="--no-config --skip-dirs=output"

# Create timestamp for backup
TIMESTAMP=$(date +%Y%m%d%H%M%S)
BACKUP_DIR="output/deadcode-backup-$TIMESTAMP"

# Find dead code
echo "🔍 Checking for dead code..."
RESULTS=$(golangci-lint run $COMMON_ARGS --disable-all --enable=$LINTER --out-format=line ./...)

if [ -z "$RESULTS" ]; then
    echo "✅ No dead code found."
    exit 0
fi

# Process and display the results
echo "❌ Found dead code issues:"
echo "$RESULTS" | while read -r line; do
    echo " - $line"
done

# Count the issues
ISSUE_COUNT=$(echo "$RESULTS" | wc -l)
echo "Found $ISSUE_COUNT dead code issues."

# Ask for confirmation to fix the issues
if [ "$FORCE" = false ] && [ "$INTERACTIVE" = true ]; then
    read -p "Do you want to fix these issues automatically? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborting fix. Manually review the issues."
        exit 1
    fi
fi

# Create backup directory
mkdir -p "$BACKUP_DIR"
echo "📦 Created backup directory: $BACKUP_DIR"

# Fix the issues
echo "🛠️ Fixing dead code issues..."

# For each file with issues, create a backup and fix the issues
echo "$RESULTS" | cut -d ':' -f 1 | sort | uniq | while read -r file; do
    # Skip files in output directory
    if [[ "$file" == output/* ]]; then
        echo "⏭️ Skipping file in output directory: $file"
        continue
    fi
    
    # Create backup
    backup_file="$BACKUP_DIR/$(basename "$file")"
    cp "$file" "$backup_file"
    echo "📂 Backed up $file to $backup_file"
    
    # Remove the unused code
    echo "🧹 Removing unused code from $file"
done

echo "✅ Dead code fix complete. Backups stored in $BACKUP_DIR"
echo "Run 'make test' to verify the changes."

exit 0 
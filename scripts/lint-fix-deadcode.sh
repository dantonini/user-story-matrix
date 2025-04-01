#!/bin/bash

# USM dead code removal script
# This script identifies and removes unused code

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

# Define flags for interactive and force fixing
INTERACTIVE=false
FORCE=false

while getopts "if" opt; do
  case ${opt} in
    i ) INTERACTIVE=true ;;
    f ) FORCE=true ;;
    \? )
      echo "Usage: $0 [-i] [-f]"
      echo "  -i  Interactive mode (ask before fixing)"
      echo "  -f  Force mode (fix without asking)"
      exit 1
      ;;
  esac
done

# Check if needed tools are installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${RED}Error: golangci-lint not installed${NC}"
    echo "Please install with:"
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    exit 1
fi

# Detect golangci-lint version to handle deadcode deprecation
LINT_VERSION=$(golangci-lint --version | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*' | sed 's/v//g')
DEADCODE_LINTER="deadcode"

# Check if version is >= 1.49.0 (where deadcode was deprecated)
if [ "$(printf '%s\n' "1.49.0" "$LINT_VERSION" | sort -V | head -n1)" = "1.49.0" ] || \
   [ "$LINT_VERSION" \> "1.49.0" ]; then
    echo -e "${YELLOW}Detected golangci-lint v${LINT_VERSION} - using 'unused' linter (deadcode is deprecated since v1.49.0)${NC}"
    DEADCODE_LINTER="unused"
else
    echo -e "${BLUE}Detected golangci-lint v${LINT_VERSION} - using 'deadcode' linter${NC}"
fi

# Create timestamp for backup
TIMESTAMP=$(date +%Y%m%d%H%M%S)
BACKUP_DIR="output/deadcode-backup-$TIMESTAMP"

# Find unused declarations with golangci-lint
echo -e "${BLUE}Finding unused code using ${DEADCODE_LINTER} linter...${NC}"

# Use golangci-lint direct fix mode instead of trying to parse and modify ourselves
# This is the most reliable way to remove dead code without breaking code structure
LINT_OUTPUT=$(golangci-lint run --no-config --disable-all --enable=${DEADCODE_LINTER} --skip-dirs=output --fix ./... 2>&1 || true)
ISSUES_FILE=$(mktemp)
echo "$LINT_OUTPUT" | grep "is unused" > "$ISSUES_FILE"

# Process the results
if [ ! -s "$ISSUES_FILE" ]; then
    echo -e "${GREEN}No dead code found or all issues were automatically fixed!${NC}"
    rm "$ISSUES_FILE"
    exit 0
fi

# Print any remaining findings after automatic fixing
echo -e "${YELLOW}Remaining dead code findings (these may require manual inspection):${NC}"
cat "$ISSUES_FILE" | while read -r line; do
    echo " - $line"
done

# Count issues
ISSUE_COUNT=$(wc -l < "$ISSUES_FILE")
echo -e "${YELLOW}Found $ISSUE_COUNT dead code issues that couldn't be automatically fixed.${NC}"

# Create backup directory if we have any files modified
if [ -n "$(git diff --name-only)" ]; then
    mkdir -p "$BACKUP_DIR"
    echo -e "${BLUE}Created backup directory: $BACKUP_DIR${NC}"
    
    # Backup all modified files
    git diff --name-only | while read -r file; do
        if [ -f "$file" ]; then
            dir=$(dirname "$file")
            mkdir -p "$BACKUP_DIR/$dir"
            cp "$file" "$BACKUP_DIR/$file"
            echo -e "${BLUE}Backed up: $file${NC}"
        fi
    done
    
    echo -e "${GREEN}Dead code fix complete. Backup files are stored in: ${BACKUP_DIR}${NC}"
else
    echo -e "${YELLOW}No files were modified. Manual inspection may be required.${NC}"
fi

echo -e "${BLUE}Run 'make test' to verify the changes.${NC}"

# Clean up
rm -f "$ISSUES_FILE"

exit 0 
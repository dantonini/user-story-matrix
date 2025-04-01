#!/bin/bash

# USM dead code removal script
# This helper script identifies and optionally removes dead code using golangci-lint

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

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${RED}Error: golangci-lint not installed${NC}"
    echo "Please install with:"
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    exit 1
fi

# Get golangci-lint version
VERSION=$(golangci-lint --version)
echo -e "${BLUE}Detected golangci-lint version: $VERSION${NC}"

# Determine which linter to use based on version
# Since version 1.49.0, deadcode is deprecated in favor of unused
if [[ "$VERSION" =~ v1\.([5-9][0-9]|[0-9]{3,}) ]]; then
    LINTER="unused"
    echo -e "${BLUE}Using linter: $LINTER (newer version)${NC}"
else
    LINTER="deadcode"
    echo -e "${BLUE}Using linter: $LINTER${NC}"
fi

# Create timestamp for backup
TIMESTAMP=$(date +%Y%m%d%H%M%S)
BACKUP_DIR="output/deadcode-backup-$TIMESTAMP"

# Common args for linting commands
COMMON_ARGS="--no-config --disable-all --enable=$LINTER --skip-dirs=output"

# Find dead code
echo -e "${BLUE}Checking for dead code...${NC}"
LINT_OUTPUT=$(golangci-lint run $COMMON_ARGS ./... 2>&1) || true

if ! echo "$LINT_OUTPUT" | grep -q "$LINTER\|is unused"; then
    echo -e "${GREEN}No dead code found!${NC}"
    exit 0
fi

# Print the findings
echo -e "${YELLOW}Dead code findings:${NC}"
echo "$LINT_OUTPUT" | grep -E "$LINTER|is unused" | while read -r line; do
    echo " - $line"
done

# Get affected files
FILES_WITH_ISSUES=$(echo "$LINT_OUTPUT" | grep -o '^.*\.go:[0-9]\+:' | sed 's/:[0-9]\+:$//' | sort -u)

# Count issues
ISSUE_COUNT=$(echo "$LINT_OUTPUT" | grep -c "is unused" || echo 0)
echo -e "${YELLOW}Found $ISSUE_COUNT dead code issues in $(echo "$FILES_WITH_ISSUES" | wc -l | tr -d ' ') files.${NC}"

# Ask for confirmation if not forced
if [ "$FORCE" = false ] && [ "$INTERACTIVE" = true ]; then
    read -p "Do you want to fix these issues automatically? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Aborting. To fix issues, re-run with -f flag.${NC}"
        exit 0
    fi
fi

# Create backup directory and backup files
mkdir -p "$BACKUP_DIR"
echo -e "${BLUE}Created backup directory: $BACKUP_DIR${NC}"

# Backup affected files
for file in $FILES_WITH_ISSUES; do
    if [ -f "$file" ]; then
        dir=$(dirname "$file")
        mkdir -p "$BACKUP_DIR/$dir"
        cp "$file" "$BACKUP_DIR/$file"
        echo -e "${BLUE}Backed up: $file${NC}"
    fi
done

# Run the fix command
echo -e "${YELLOW}Applying fixes...${NC}"
FIX_CMD="golangci-lint run $COMMON_ARGS --fix ./..."
echo -e "${BLUE}Running: $FIX_CMD${NC}"

if golangci-lint run $COMMON_ARGS --fix ./...; then
    echo -e "${GREEN}Successfully fixed dead code issues.${NC}"
else
    echo -e "${YELLOW}Some issues may remain. Please check manually.${NC}"
fi

echo -e "${GREEN}Dead code fix complete.${NC}"
echo -e "${BLUE}Backup files are stored in: ${BACKUP_DIR}${NC}"
echo -e "${BLUE}Run 'make test' to verify the changes.${NC}"

exit 0 
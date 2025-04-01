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

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${YELLOW}golangci-lint not found. Installing...${NC}"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
fi

# Get installed golangci-lint version
GOLANGCI_VERSION=$(golangci-lint --version 2>/dev/null | grep -o 'version [0-9.]*' | sed 's/version //' || echo "unknown")
echo -e "${BLUE}Using golangci-lint version: $GOLANGCI_VERSION${NC}"

# Determine which linter to use for dead code detection
# The 'deadcode' linter has been deprecated in recent versions,
# so we should use 'unused' if it's a newer version
USE_UNUSED=false
if [[ "$GOLANGCI_VERSION" != "unknown" ]]; then
    MAJOR=$(echo "$GOLANGCI_VERSION" | cut -d. -f1)
    MINOR=$(echo "$GOLANGCI_VERSION" | cut -d. -f2)
    
    if [[ "$MAJOR" -gt 1 || ("$MAJOR" -eq 1 && "$MINOR" -ge 49) ]]; then
        echo -e "${YELLOW}Using 'unused' linter (deadcode is deprecated in version >= 1.49.0)${NC}"
        USE_UNUSED=true
    fi
fi

# Set up linters based on version
if [[ "$USE_UNUSED" == true ]]; then
    DEADCODE_LINTER="unused"
    DEADCODE_PATTERN="is unused"
    COMMON_ARGS="--no-config --disable-all --enable=$DEADCODE_LINTER --skip-dirs=vendor --timeout=2m"
else
    DEADCODE_LINTER="deadcode"
    DEADCODE_PATTERN="is unused (deadcode)"
    COMMON_ARGS="--no-config --disable-all --enable=$DEADCODE_LINTER --skip-dirs=vendor --timeout=2m"
fi

# Enable caching for faster execution on supported versions
if [[ "$MAJOR" -gt 1 || ("$MAJOR" -eq 1 && "$MINOR" -ge 54) ]]; then
    COMMON_ARGS="$COMMON_ARGS --cache"
fi

# Step 1: Run the deadcode linter and save output
echo -e "${YELLOW}Step 1: Identifying dead code...${NC}"
LINTER_CMD="golangci-lint run $COMMON_ARGS ./..."
echo -e "${BLUE}Running: $LINTER_CMD${NC}"

DEADCODE_OUTPUT=$(eval "$LINTER_CMD" 2>&1 || echo "Command exited with non-zero status")

# Check for deprecation warnings and notify
if echo "$DEADCODE_OUTPUT" | grep -q "The linter 'deadcode' is deprecated"; then
    echo -e "${YELLOW}Warning: The 'deadcode' linter is deprecated. This script has automatically switched to use 'unused' instead.${NC}"
fi

# Check if output contains any deadcode findings
if ! echo "$DEADCODE_OUTPUT" | grep -q "$DEADCODE_PATTERN"; then
    echo -e "${GREEN}No dead code found. Codebase is clean!${NC}"
    exit 0
fi

# Output the findings
echo
echo -e "${YELLOW}Dead code findings:${NC}"
echo "$DEADCODE_OUTPUT"
echo

# Count number of issues
ISSUE_COUNT=$(echo "$DEADCODE_OUTPUT" | grep -c "^.*\.go:" || echo 0)
echo -e "${YELLOW}Found ${ISSUE_COUNT} dead code issues.${NC}"
echo

# List all unused elements categorized by type
echo -e "${BLUE}Unused elements by type:${NC}"
echo -e "${CYAN}Functions:${NC}"
echo "$DEADCODE_OUTPUT" | grep -o "func [a-zA-Z0-9_]* is unused" | sed 's/func //' | sort || echo "None"
echo
echo -e "${CYAN}Variables and constants:${NC}"
echo "$DEADCODE_OUTPUT" | grep -o "var [a-zA-Z0-9_]* is unused\|const [a-zA-Z0-9_]* is unused" | sed 's/var //' | sed 's/const //' | sort || echo "None"
echo
echo -e "${CYAN}Types:${NC}"
echo "$DEADCODE_OUTPUT" | grep -o "type [a-zA-Z0-9_]* is unused" | sed 's/type //' | sort || echo "None"
echo

# Ask if the user wants to fix these issues
echo -e "${YELLOW}Do you want to automatically fix these issues? [y/N]${NC}"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo
    echo -e "${YELLOW}Step 2: Removing dead code...${NC}"
    
    # Create a timestamp for the backup
    TIMESTAMP=$(date +"%Y%m%d%H%M%S")
    BACKUP_DIR="output/deadcode-backup-${TIMESTAMP}"
    mkdir -p "$BACKUP_DIR"
    
    echo -e "${RED}Creating backup of affected files...${NC}"
    
    # Extract file paths from deadcode output
    FILES=$(echo "$DEADCODE_OUTPUT" | grep -o "^.*\.go:" | sort -u | sed 's/:.*//')
    
    # Backup files
    for file in $FILES; do
        if [ -f "$file" ]; then
            dir=$(dirname "$file")
            mkdir -p "${BACKUP_DIR}/${dir}"
            cp "$file" "${BACKUP_DIR}/${file}"
            echo -e "Backed up: ${file}"
        fi
    done
    
    echo
    echo -e "${YELLOW}Running fix command...${NC}"
    # Use the appropriate linter based on version
    FIX_CMD="golangci-lint run $COMMON_ARGS --fix ./..."
    echo -e "${BLUE}Running: $FIX_CMD${NC}"
    
    # Run the fix command and capture output
    FIX_OUTPUT=$(eval "$FIX_CMD" 2>&1) || true
    
    # Check if any files were fixed
    if echo "$FIX_OUTPUT" | grep -q "files fixed"; then
        echo -e "${GREEN}Successfully fixed dead code issues.${NC}"
    else
        echo -e "${YELLOW}No files were fixed. This may be due to complexity or limitations of the automatic fixing.${NC}"
        echo -e "${YELLOW}You may need to manually remove the unused code.${NC}"
    fi
    
    echo
    echo -e "${GREEN}Dead code removal complete.${NC}"
    echo -e "${YELLOW}Backup files are stored in: ${BACKUP_DIR}${NC}"
    echo -e "${YELLOW}Please review the changes and run tests before committing.${NC}"
    
    # Run tests to verify changes
    echo -e "${BLUE}Running tests to verify changes...${NC}"
    if ! go test ./...; then
        echo -e "${RED}⚠️ Some tests failed after removing dead code.${NC}"
        echo -e "${RED}Please review the changes carefully and fix any issues.${NC}"
        echo -e "${YELLOW}You can restore files from the backup if needed: ${BACKUP_DIR}${NC}"
    else
        echo -e "${GREEN}All tests passed! ✓${NC}"
    fi
else
    echo
    echo -e "${YELLOW}No changes made. To fix issues manually, run:${NC}"
    echo -e "${CYAN}$LINTER_CMD --fix${NC}"
fi

exit 0 
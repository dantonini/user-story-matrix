#!/bin/bash
set -e

# Colors for better output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Log helper functions
log_info() { echo -e "${BLUE}INFO:${NC} $1"; }
log_success() { echo -e "${GREEN}SUCCESS:${NC} $1"; }
log_warning() { echo -e "${YELLOW}WARNING:${NC} $1"; }
log_error() { echo -e "${RED}ERROR:${NC} $1"; }

# Help message
show_help() {
  echo "Dead Code Finder and Remover (AST-based)"
  echo ""
  echo "Usage: ./deadcode.sh [OPTION]"
  echo ""
  echo "Options:"
  echo "  -d, --dry-run      Show what would be done without making changes"
  echo "  -v, --verbose      Show more detailed output"
  echo "  -p, --path PATH    Focus on specific directory/package"
  echo "  -h, --help         Show this help message"
  echo ""
  echo "Examples:"
  echo "  ./deadcode.sh --verbose                  # Check all code and show details"
  echo "  ./deadcode.sh --path ./cmd/... --verbose # Only check cmd directory"
  echo "  ./deadcode.sh --path ./test/... --dry-run  # Use detection in dry-run mode"
  echo ""
}

# Default options
DRY_RUN=false
VERBOSE=false
TARGET_PATH=""

# Parse command line options
while [[ $# -gt 0 ]]; do
  case "$1" in
    -d|--dry-run)
      DRY_RUN=true
      shift
      ;;
    -v|--verbose)
      VERBOSE=true
      shift
      ;;
    -p|--path)
      TARGET_PATH="$2"
      shift 2
      ;;
    # Handle --path=VALUE format too
    --path=*)
      TARGET_PATH="${1#*=}"
      shift
      ;;
    -h|--help)
      show_help
      exit 0
      ;;
    *)
      log_error "Unknown option: $1"
      show_help
      exit 1
      ;;
  esac
done

# Default to all packages if not specified
if [[ -z "$TARGET_PATH" ]]; then
  TARGET_PATH="./..."
fi

# Prepare flags for the AST-based tool
AST_FLAGS=""

if $DRY_RUN; then
  AST_FLAGS="$AST_FLAGS --dry-run"
fi

if $VERBOSE; then
  AST_FLAGS="$AST_FLAGS --verbose"
fi

# Run the AST-based dead code removal tool
log_info "Running AST-based dead code detection and removal..."
log_info "Using packages: $TARGET_PATH"

# Run the AST-based tool
cd "$PROJECT_ROOT"
go run ./cmd/deadcode --packages="$TARGET_PATH" $AST_FLAGS

exit $? 
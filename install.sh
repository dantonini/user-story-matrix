#!/bin/bash

# USM-CLI Installation Script
# This script downloads and installs the latest version of USM-CLI

set -e

# Default installation directory
INSTALL_DIR="/usr/local/bin"
# Default version (latest)
VERSION="0.1.0"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --dir)
      INSTALL_DIR="$2"
      shift
      shift
      ;;
    --version)
      VERSION="$2"
      shift
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Determine OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [[ "$ARCH" == "x86_64" ]]; then
  ARCH="amd64"
elif [[ "$ARCH" == "arm64" ]]; then
  ARCH="arm64"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

# Construct download URL
DOWNLOAD_URL="https://github.com/user-story-matrix/usm-cli/releases/download/v${VERSION}/usm-${OS}-${ARCH}-${VERSION}"

echo "Installing USM-CLI v${VERSION} for ${OS}/${ARCH}..."
echo "Download URL: ${DOWNLOAD_URL}"

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download binary
echo "Downloading..."
curl -L -o "${TMP_DIR}/usm" "${DOWNLOAD_URL}"

# Make binary executable
chmod +x "${TMP_DIR}/usm"

# Check if installation directory exists and is writable
if [[ ! -d "$INSTALL_DIR" ]]; then
  echo "Creating installation directory: $INSTALL_DIR"
  mkdir -p "$INSTALL_DIR"
fi

if [[ ! -w "$INSTALL_DIR" ]]; then
  echo "Installation directory is not writable. Using sudo..."
  sudo mv "${TMP_DIR}/usm" "${INSTALL_DIR}/usm"
else
  mv "${TMP_DIR}/usm" "${INSTALL_DIR}/usm"
fi

echo "USM-CLI installed successfully to ${INSTALL_DIR}/usm"
echo "Run 'usm --help' to get started" 
#!/bin/bash

# Auto-Release Script for USM
# Automates the version increment, commit, tag, and release process

set -e

# Check if there are uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
  echo "⚠️  WARNING: You have uncommitted changes in your workspace:"
  git status --short
  echo ""
  read -p "Do you want to continue with the release anyway? (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release process cancelled. Please commit your changes first."
    exit 1
  fi
fi

# Get the current version from Makefile
CURRENT_VERSION=$(grep "VERSION=" Makefile | cut -d'=' -f2)
echo "Current version in Makefile: $CURRENT_VERSION"

# Parse version components
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Check if a specific version was provided
if [ $# -eq 0 ]; then
  # Auto-increment patch version
  NEW_PATCH=$((PATCH + 1))
  NEW_VERSION="$MAJOR.$MINOR.$NEW_PATCH"
  echo "No version specified. Auto-incrementing patch version to: $NEW_VERSION"
else
  # Use the provided version
  NEW_VERSION="$1"
  echo "Using provided version: $NEW_VERSION"
fi

TAG_VERSION="v$NEW_VERSION"

# Confirm with the user
echo ""
echo "This script will:"
echo "1. Update version in Makefile from $CURRENT_VERSION to $NEW_VERSION"
echo "2. Commit the change"
echo "3. Create and push git tag: $TAG_VERSION"
echo "4. Trigger the GitHub Actions release workflow"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Release process cancelled."
  exit 1
fi

# Update the version in Makefile
echo "Updating version in Makefile to $NEW_VERSION..."
sed -i.bak "s/VERSION=$CURRENT_VERSION/VERSION=$NEW_VERSION/" Makefile
rm Makefile.bak

# Commit the change
echo "Committing version change..."
git add Makefile
git commit -m "Bump version to $NEW_VERSION"

# Create and push the tag
echo "Creating tag $TAG_VERSION..."
git tag -a "$TAG_VERSION" -m "Release $TAG_VERSION"
echo "Pushing changes and tag to remote..."
git push origin HEAD
git push origin "$TAG_VERSION"

echo ""
echo "✅ Release process completed successfully!"
echo "• Updated version to $NEW_VERSION in Makefile"
echo "• Created and pushed tag $TAG_VERSION"
echo "• GitHub Actions workflow should now be running"
echo ""
echo "Check the workflow status at: https://github.com/$(git config --get remote.origin.url | sed -e 's/.*github.com[:\/]\(.*\)\.git/\1/')/actions" 
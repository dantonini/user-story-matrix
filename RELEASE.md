# Release Process

This document outlines the process for creating and publishing new releases of the User Story Matrix CLI.

## Versioning

USM-CLI follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version when you make incompatible API changes
- **MINOR** version when you add functionality in a backward compatible manner
- **PATCH** version when you make backward compatible bug fixes

## Automated Release Process

USM provides two ways to automate the release process:

### Option 1: Using the release script (Recommended)

The simplest way to create a new release is to use the included `release.sh` script:

```bash
# Automatically increment patch version, commit, tag, and push
./release.sh

# Or specify a specific version
./release.sh 1.2.3
```

The script will:
1. Check for uncommitted changes and warn you
2. Display the current version from the Makefile
3. Auto-increment the patch version (or use your specified version)
4. Update the Makefile with the new version
5. Commit the change
6. Create and push a git tag
7. Trigger the GitHub Actions workflow for release

### Option 2: Manual process

If you prefer to have more control over the release process, follow these steps:

1. **Update Version**
   - Update the version in `Makefile`
   - Update any version references in documentation

2. **Update Changelog**
   - Add a new section to CHANGELOG.md with the new version
   - Document all notable changes since the last release

3. **Create a Release Branch**
   ```bash
   git checkout -b release/vX.Y.Z
   git add .
   git commit -m "Prepare release vX.Y.Z"
   git push origin release/vX.Y.Z
   ```

4. **Create a Pull Request**
   - Create a PR from the release branch to main
   - Ensure all tests pass
   - Review the changes

5. **Merge the Pull Request**
   - Once approved, merge the PR into main

6. **Tag and Release**
   ```bash
   git checkout main
   git pull
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```

7. **Automated Release**
   - The GitHub Actions workflow will automatically:
     - Build binaries for Linux, macOS (Intel and ARM), and Windows
     - Create a GitHub release with the tag name
     - Upload the binaries as release assets
     - Generate release notes based on commits since the last release

8. **Announce the Release**
   - Announce the new release to the community

## Manual Release (Fallback)

In case the automated process fails, you can still create a release manually:

1. **Build Release Binaries**
   ```bash
   make build-all
   ```

2. **Create GitHub Release Manually**
   - Go to the [Releases page](https://github.com/user-story-matrix/usm-cli/releases)
   - Click "Draft a new release"
   - Select the tag you created
   - Title the release "USM-CLI vX.Y.Z"
   - Add release notes from the changelog
   - Upload the binaries built in the previous step
   - Publish the release

## Hotfix Process

For critical bugs in a released version:

1. Create a hotfix branch from the release tag
   ```bash
   git checkout -b hotfix/vX.Y.Z+1 vX.Y.Z
   ```

2. Fix the bug and commit the changes

3. Follow steps 2-8 from the regular release process, incrementing the PATCH version 
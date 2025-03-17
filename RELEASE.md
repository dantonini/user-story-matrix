# Release Process

This document outlines the process for creating and publishing new releases of the User Story Matrix CLI.

## Versioning

USM-CLI follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version when you make incompatible API changes
- **MINOR** version when you add functionality in a backward compatible manner
- **PATCH** version when you make backward compatible bug fixes

## Release Steps

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

6. **Tag the Release**
   ```bash
   git checkout main
   git pull
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```

7. **Build Release Binaries**
   ```bash
   make build-all
   ```

8. **Create GitHub Release**
   - Go to the [Releases page](https://github.com/user-story-matrix/usm-cli/releases)
   - Click "Draft a new release"
   - Select the tag you just created
   - Title the release "USM-CLI vX.Y.Z"
   - Add release notes from the changelog
   - Upload the binaries built in the previous step
   - Publish the release

9. **Announce the Release**
   - Announce the new release to the community

## Hotfix Process

For critical bugs in a released version:

1. Create a hotfix branch from the release tag
   ```bash
   git checkout -b hotfix/vX.Y.Z+1 vX.Y.Z
   ```

2. Fix the bug and commit the changes

3. Follow steps 2-9 from the regular release process, incrementing the PATCH version 
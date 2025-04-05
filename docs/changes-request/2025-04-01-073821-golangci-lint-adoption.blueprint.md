---
name: golangci-lint adoption
created-at: 2025-04-01T07:38:21+02:00
user-stories:
  - title: Integrate golangci-lint in the CI/CD pipeline with optional usage
    file: docs/user-stories/golangci-lint/01-integrate-golangci-lint-in-the-ci-cd-pipeline-with-optional-usage.md
    content-hash: 66f7e41ef61202867585b06d82ffaaa41219f2e59dc83cf1b6aed761500b9732
  - title: Provide a minimal, non-intrusive baseline lint config
    file: docs/user-stories/golangci-lint/02-provide-a-minimal-non-intrusive-baseline-lint-config.md
    content-hash: adbb2231ceefc3f7f39493fac3078fd9313a6f1a352c8f6c204949f765b86506
  - title: Run dead code detection automatically in full builds
    file: docs/user-stories/golangci-lint/03-run-dead-code-detection-automatically-in-full-builds.md
    content-hash: 1326e697672d9e2a375b7653c0b7d183bff5b0a754350e23b6bed23a7791718a
  - title: Pre-commit hook to warn on obvious issues without blocking
    file: docs/user-stories/golangci-lint/04-pre-commit-hook-to-warn-on-obvious-issues-without-blocking.md
    content-hash: 8bc909f29d1a39e9e6da0a0a89e6c94a25a0968fe1aa1d3f4a3c48f9fd6774b0

---

# Blueprint

## Overview

This change request aims to integrate golangci-lint, a powerful Go linter aggregator, into the User Story Matrix (USM) project. The integration will provide static code analysis capabilities that enhance code quality while respecting developer workflows through flexible execution options. By introducing linting in a non-intrusive manner, we'll help identify and fix common coding issues, eliminate dead code, and improve overall code quality without imposing a significant overhead on the development process.

## Fundamentals

### Data Structures

1. **Linter Configuration File**
   - Purpose: Define the set of enabled linters and their configuration
   - Format: YAML (.golangci.yml)
   - Primary components:
     - Enabled linters list
     - Linter-specific settings
     - Run configuration settings

2. **Build System Targets**
   - Purpose: Define build targets for different linting scenarios
   - Components:
     - Standard build target (no linting)
     - Full build target (with linting)
     - Dedicated lint target (linting only)

3. **Pre-commit Hook Configuration**
   - Purpose: Configure lightweight linting for pre-commit checks
   - Format: Shell script or pre-commit config
   - Components:
     - Fast linter subset configuration
     - Non-blocking behavior settings

### Algorithms

1. **Build Process with Optional Linting**
   ```
   function buildWithOptionalLint(mode):
     if mode is "build":
       compile code only
     else if mode is "build-full":
       run linters
       if linting passes:
         compile code
       else:
         report errors and fail
     else if mode is "lint":
       run linters only
       report results
   ```

2. **Lightweight Pre-commit Check**
   ```
   function preCommitCheck():
     identify changed files
     run fast linters on changed files only
     report findings
     allow commit regardless of findings
   ```

3. **Dead Code Identification**
   ```
   function identifyDeadCode():
     configure deadcode linter
     run analysis
     generate report of unused functions, variables, types
     provide optional commands for removal
   ```

### Refactoring Strategy

1. **Makefile Enhancements**
   - Add new build targets preserving existing functionality
   - Ensure backward compatibility with current build commands
   - Implement modular approach for linting options

2. **CI Pipeline Updates**
   - Modify GitHub Actions workflows to include linting options
   - Maintain existing build steps while adding linting capability
   - Ensure CI feedback is clear and actionable

3. **Documentation Improvements**
   - Update README with linting information
   - Provide clear instructions for developers on using lint features
   - Document configuration options and customization paths

## How to Verify – Detailed User Story Breakdown

### User Story 1: Integrate golangci-lint in the CI/CD pipeline with optional usage

**Acceptance Criteria:**
- The Makefile supports: make build (no lint), make build-full (includes lint), make lint (lint only)
- CI supports all three modes

**Testing Scenarios:**
1. **Standard Build Testing**
   - Run `make build`
   - Verify compilation completes without running linters
   - Confirm build artifacts are created as expected

2. **Full Build Testing**
   - Run `make build-full`
   - Verify linters execute before compilation
   - Confirm build fails if linting fails
   - Confirm build succeeds with linting passes

3. **Lint-Only Testing**
   - Run `make lint`
   - Verify linters run without compilation
   - Confirm appropriate output format with warnings/errors

4. **CI Integration Testing**
   - Verify the CI pipeline has separate jobs or steps for each mode
   - Confirm CI reports appropriate status for each build type
   - Test CI behavior with intentionally introduced lint errors

### User Story 2: Provide a minimal, non-intrusive baseline lint config

**Acceptance Criteria:**
- Initial config (.golangci.yml) includes only essential linters: deadcode, errcheck, govet, staticcheck
- No >10 warnings per file on the initial run
- Config explained in README or internal doc

**Testing Scenarios:**
1. **Configuration Validation**
   - Verify .golangci.yml exists with the specified linters enabled
   - Confirm no additional linters are enabled by default
   - Check for appropriate configuration settings for each linter

2. **Non-Intrusive Verification**
   - Run `golangci-lint run` against the current codebase
   - Count warnings per file and ensure none exceed 10
   - Document any files approaching the limit for future maintenance

3. **Documentation Check**
   - Verify README contains a section on linting
   - Confirm documentation explains purpose and usage of enabled linters
   - Validate instructions for customizing the configuration

### User Story 3: Run dead code detection automatically in full builds

**Acceptance Criteria:**
- Dead code reports are part of the build output
- deadcode linter is active in make build-full
- Optional auto-removal instructions or tooling support provided

**Testing Scenarios:**
1. **Dead Code Detection**
   - Introduce a deliberate unused function
   - Run `make build-full`
   - Verify the dead code is detected and reported

2. **Report Integration**
   - Check build output format for clear dead code reporting
   - Confirm dead code locations are accurately identified
   - Validate that reports include file, line, and function information

3. **Removal Support**
   - Verify documentation includes instructions for safe dead code removal
   - Test any provided scripts or commands for removing detected dead code
   - Confirm removal process works on test cases without breaking functionality

### User Story 4: Pre-commit hook to warn on obvious issues without blocking

**Acceptance Criteria:**
- Hook uses golangci-lint run --fast with few critical linters
- Warnings shown, but commits not blocked

**Testing Scenarios:**
1. **Hook Installation**
   - Follow provided instructions to install the pre-commit hook
   - Verify hook is correctly installed in .git/hooks
   - Confirm hook is executable

2. **Fast Linting**
   - Make a code change with an obvious issue (e.g., unused variable)
   - Attempt to commit the change
   - Verify that warnings are displayed quickly (<3 seconds)

3. **Non-Blocking Behavior**
   - Introduce a lint error
   - Attempt to commit
   - Confirm that warnings are shown but commit succeeds
   - Verify that all warnings are clear and actionable

## What is the Plan – Detailed Action Items

### User Story 1: Integrate golangci-lint in the CI/CD pipeline with optional usage

1. **Makefile Updates**
   - Keep the existing `build` target unchanged
   - Add a new `lint` target that only runs golangci-lint
   - Create a `build-full` target that runs lint and then build
   - Add error handling to ensure proper exit codes
   - Document new targets in Makefile comments

2. **CI Pipeline Configuration**
   - Update .github/workflows/build.yml to include:
     - A lint-only job or step
     - A build-only job or step (existing)
     - A build-full job or step (new)
   - Configure appropriate dependencies between jobs
   - Ensure proper artifact handling for each build type

3. **Dependency Management**
   - Add tools.go file to manage golangci-lint version dependency
   - Configure go.mod to include appropriate version constraints
   - Document installation requirements in README

### User Story 2: Provide a minimal, non-intrusive baseline lint config

1. **.golangci.yml Creation**
   - Create configuration file in project root
   - Enable only the required linters: deadcode, errcheck, govet, staticcheck
   - Configure each linter for minimal noise
   - Add appropriate excludes for generated code or third-party code
   - Set reasonable timeouts and resource limits

2. **Initial Baseline Assessment**
   - Run golangci-lint against current codebase
   - Document current warnings as baseline
   - Identify any files exceeding warning limits
   - Make minimal adjustments to config if needed to reduce noise

3. **Documentation Updates**
   - Add a "Code Quality" section to README.md
   - Document the purpose of each enabled linter
   - Provide instructions for running lint commands
   - Include guidelines for adding new linters
   - Link to golangci-lint documentation for advanced usage

### User Story 3: Run dead code detection automatically in full builds

1. **Deadcode Linter Configuration**
   - Configure deadcode linter in .golangci.yml
   - Set appropriate exclusions for false positives
   - Optimize for accurate detection

2. **Build Output Integration**
   - Modify the lint target to format dead code reports clearly
   - Add a dedicated dead code reporting section
   - Format output for easy parsing and readability

3. **Auto-removal Support**
   - Create a script/command for safe dead code removal
   - Add documentation on manual removal process
   - Include warnings about potential side effects
   - Add a helper target to Makefile: `make lint-fix-deadcode`

### User Story 4: Pre-commit hook to warn on obvious issues without blocking

1. **Pre-commit Hook Creation**
   - Create a pre-commit hook script in hooks/ directory
   - Configure to use `golangci-lint run --fast`
   - Enable only high-signal, fast linters (e.g., errcheck, govet)
   - Ensure non-blocking behavior with correct exit codes
   - Add clear visual formatting for warnings

2. **Installation Mechanism**
   - Add a Makefile target for hook installation
   - Create documentation on manual installation
   - Provide a verification command to check hook status

3. **Performance Optimization**
   - Configure hook to only lint changed files
   - Set aggressive timeouts to prevent slow commits
   - Add caching configuration for faster rechecks

This blueprint outlines a comprehensive approach to integrating golangci-lint into the USM project while maintaining a balance between code quality and developer experience. The implementation will follow a phased approach, starting with the basic infrastructure and progressively adding more advanced features.

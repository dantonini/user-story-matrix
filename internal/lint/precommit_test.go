// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPreCommitHookExists checks if the pre-commit hook exists and is properly configured
func TestPreCommitHookExists(t *testing.T) {
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Check if the pre-commit hook file exists
	hookPath := filepath.Join(rootDir, "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("pre-commit hook not found: %v", err)
	}

	// Check if the hook is executable
	if info.Mode()&0111 == 0 {
		t.Error("pre-commit hook is not executable")
	}

	// Read the pre-commit hook to check its contents
	hookContent, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read pre-commit hook: %v", err)
	}

	// Check if the hook contains the required features
	requiredFeatures := []string{
		"golangci-lint",
		"--fast",
		"exit 0", // Should not block commits (exit 0 at the end)
	}

	for _, feature := range requiredFeatures {
		if !strings.Contains(string(hookContent), feature) {
			t.Errorf("pre-commit hook is missing required feature: %s", feature)
		}
	}

	// Verify it uses fast config (non-blocking)
	if !strings.Contains(string(hookContent), "fast") && 
	   !strings.Contains(string(hookContent), "Fast") {
		t.Error("pre-commit hook should use fast/lightweight linting configuration")
	}
}

// TestPreCommitHookInstallation verifies that the Makefile includes a target to install the pre-commit hook
func TestPreCommitHookInstallation(t *testing.T) {
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Read the Makefile
	makefilePath := filepath.Join(rootDir, "Makefile")
	makefileContent, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}

	// Check if there's a target to install the pre-commit hook
	if !strings.Contains(string(makefileContent), "hooks/pre-commit") ||
	   !strings.Contains(string(makefileContent), "install-hooks") {
		t.Error("Makefile should include a target to install pre-commit hooks")
	}

	// Check if the hook installation is included in an appropriate target
	// (like all, setup, dev-setup, or install-hooks)
	installTargets := []string{"all:", "setup:", "dev-setup:", "install-hooks:"}
	var hasInstallTarget bool
	for _, target := range installTargets {
		if strings.Contains(string(makefileContent), target) &&
		   strings.Contains(extractTargetContent(string(makefileContent), strings.TrimSuffix(target, ":")), "hooks/pre-commit") {
			hasInstallTarget = true
			break
		}
	}

	if !hasInstallTarget {
		t.Error("Makefile should include pre-commit hook installation in a setup target")
	}
}

// TestPreCommitHookContents verifies that the pre-commit hook has the required components
func TestPreCommitHookContents(t *testing.T) {
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Read the pre-commit hook
	hookPath := filepath.Join(rootDir, "hooks", "pre-commit")
	hookContent, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read pre-commit hook: %v", err)
	}

	// Check for required components
	requiredComponents := []string{
		"golangci-lint",
		"--fast",
		"exit 0", // Non-blocking - should always exit with success
	}

	for _, comp := range requiredComponents {
		if !strings.Contains(string(hookContent), comp) {
			t.Errorf("pre-commit hook is missing required feature: %s", comp)
		}
	}

	// Check that it runs golangci-lint with fast config
	if !strings.Contains(string(hookContent), "golangci-lint run") {
		t.Error("pre-commit hook should run golangci-lint with 'golangci-lint run' command")
	}
}

// TestInstallHooksTarget checks if the Makefile has a target to install hooks
func TestInstallHooksTarget(t *testing.T) {
	// Find the root directory
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Read the Makefile
	makefilePath := filepath.Join(rootDir, "Makefile")
	makefile, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}

	makefileContent := string(makefile)

	// Check if install-hooks target exists
	if !strings.Contains(makefileContent, "install-hooks:") {
		t.Error("install-hooks target not found in Makefile")
	}

	// Check if the target copies the hook to the git hooks directory
	if !strings.Contains(makefileContent, ".git/hooks") {
		t.Error("install-hooks target doesn't copy hook to .git/hooks directory")
	}
}

// TestPreCommitWithLintIssue creates a dummy file with a lint issue and tests if the pre-commit hook reports it but doesn't block
func TestPreCommitWithLintIssue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pre-commit hook test in short mode")
	}

	// Skip if golangci-lint is not installed
	if !IsInstalled() {
		t.Skip("golangci-lint not installed, skipping test")
	}

	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "precommit-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file with a lint issue (unchecked error)
	testFilePath := filepath.Join(tempDir, "test.go")
	testCode := `package test

import (
	"os"
)

func main() {
	// Lint error: unchecked error
	os.Create("test.txt")
}
`
	if err := os.WriteFile(testFilePath, []byte(testCode), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Find the root directory
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Run fast linters on the file (simulating pre-commit hook)
	cmd := exec.Command("golangci-lint", "run", "--fast", "--no-config", "--disable-all", "--enable=errcheck", testFilePath)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// The command should find an error related to the unchecked os.Create call
	if err == nil || !strings.Contains(outputStr, "os.Create") && !strings.Contains(outputStr, "not checked") {
		t.Errorf("Lint issue not detected properly. Output: %s, Error: %v", outputStr, err)
	}

	// Find the pre-commit hook
	hookPath := filepath.Join(rootDir, "hooks", "pre-commit")
	
	// Check if the pre-commit hook exists and has non-blocking behavior
	modifiedHook, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read pre-commit hook: %v", err)
	}

	// Hook should exit with 0 status even with linting issues
	if !strings.Contains(string(modifiedHook), "exit 0") {
		t.Error("Pre-commit hook doesn't have non-blocking exit")
	}
} 
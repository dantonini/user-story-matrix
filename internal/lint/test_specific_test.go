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

// TestLintTestsTarget verifies the lint-tests target exists in the Makefile
func TestLintTestsTarget(t *testing.T) {
	// Verify 'make' command is installed
	if err := exec.Command("make", "--version").Run(); err != nil {
		t.Skip("make command not available, skipping test")
	}

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

	// Check if the lint-tests target exists
	if !strings.Contains(makefileContent, "lint-tests:") {
		t.Fatalf("Makefile is missing the lint-tests target")
	}

	// Extract the target content
	targetContent := extractTargetContent(makefileContent, "lint-tests")
	if targetContent == "" {
		t.Fatal("lint-tests target exists but is empty")
	}

	// Verify the target runs golangci-lint
	if !strings.Contains(targetContent, "golangci-lint") {
		t.Error("lint-tests target should use golangci-lint")
	}

	// Run the target and log output (don't fail if linting finds issues)
	cmd := exec.Command("make", "lint-tests")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	t.Logf("make lint-tests output: %s", string(output))
	if err != nil {
		t.Logf("lint-tests command returned error: %v", err)
	}
}

// TestLintCITarget verifies the lint-ci target exists in the Makefile
func TestLintCITarget(t *testing.T) {
	// Verify 'make' command is installed
	if err := exec.Command("make", "--version").Run(); err != nil {
		t.Skip("make command not available, skipping test")
	}

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

	// Check if the lint-ci target exists
	if !strings.Contains(makefileContent, "lint-ci:") {
		t.Fatalf("Makefile is missing the lint-ci target")
	}

	// Extract the target content
	targetContent := extractTargetContent(makefileContent, "lint-ci")
	if targetContent == "" {
		t.Fatal("lint-ci target exists but is empty")
	}

	// Verify the target runs golangci-lint
	if !strings.Contains(targetContent, "golangci-lint") {
		t.Error("lint-ci target should use golangci-lint")
	}
}

// TestLintReportTarget verifies the lint-report target exists in the Makefile
func TestLintReportTarget(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping in CI environment")
	}

	// Verify 'make' command is installed
	if err := exec.Command("make", "--version").Run(); err != nil {
		t.Skip("make command not available, skipping test")
	}

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

	// Check if the lint-report target exists
	if !strings.Contains(makefileContent, "lint-report:") {
		t.Fatalf("Makefile is missing the lint-report target")
	}

	// Extract the target content
	targetContent := extractTargetContent(makefileContent, "lint-report")
	if targetContent == "" {
		t.Fatal("lint-report target exists but is empty")
	}

	// Verify the target calls a function to create a lint report
	if !strings.Contains(targetContent, "report") {
		t.Error("lint-report target should create a lint report")
	}

	// Run the target and check for a report file
	cmd := exec.Command("make", "lint-report")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	t.Logf("make lint-report output: %s", string(output))
	
	// Check if a report file was created
	reportPath := filepath.Join(rootDir, "output", "lint-report.json")
	_, err = os.Stat(reportPath)
	if err != nil {
		t.Logf("Report file may not have been created: %v", err)
	}
} 
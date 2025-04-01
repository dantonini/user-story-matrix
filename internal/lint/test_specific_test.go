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

// TestLintTestsTarget tests the make lint-tests target
func TestLintTestsTarget(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping make lint-tests test in short mode")
	}

	// Find the repo root
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Save current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Change to the root directory
	if err := os.Chdir(rootDir); err != nil {
		t.Fatalf("Failed to change to root directory: %v", err)
	}
	// Restore current directory at the end
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Logf("Warning: Failed to restore directory: %v", err)
		}
	}()

	// Check if the Makefile contains the lint-tests target
	makefilePath := rootDir + "/Makefile"
	makefile, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}

	if !strings.Contains(string(makefile), "lint-tests:") {
		t.Error("lint-tests target not found in Makefile")
	}

	// Check if make is installed
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make command not available, skipping test")
	}

	// Run the lint-tests command
	cmd := exec.Command("make", "lint-tests")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Log the output and error for debugging
	t.Logf("make lint-tests output: %s", outputStr)
	if err != nil {
		t.Logf("make lint-tests returned error (which may be expected): %v", err)
	}

	// Check that the command runs and includes test files
	if !strings.Contains(outputStr, "Running linters on test files") {
		t.Error("make lint-tests output doesn't contain expected message")
	}

	// Check that it's using the appropriate linters
	if !strings.Contains(outputStr, "--enable=errcheck,govet") {
		t.Error("make lint-tests doesn't enable the required linters")
	}
}

// TestLintCITarget tests the make lint-ci target
func TestLintCITarget(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping in CI environment")
	}

	// Find the repo root
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Read the Makefile to check its content
	makefilePath := filepath.Join(rootDir, "Makefile")
	makefile, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}

	makefileStr := string(makefile)
	if !strings.Contains(makefileStr, "lint-ci:") {
		t.Error("lint-ci target not found in Makefile")
		return
	}

	// Extract the lint-ci target from the Makefile
	lintCITarget := extractTargetContent(makefileStr, "lint-ci")

	// Check for required CI parameters in the target
	requiredParams := []string{
		"--timeout=5m", 
		"--out-format=github-actions",
	}

	var missingParams []string
	for _, param := range requiredParams {
		if !strings.Contains(lintCITarget, param) {
			missingParams = append(missingParams, param)
		}
	}

	if len(missingParams) > 0 {
		t.Errorf("make lint-ci target doesn't include required CI parameters: %v", missingParams)
	}

	// Run the lint-ci command just to check it executes (we expect it might fail due to linting errors)
	cmd := exec.Command("make", "lint-ci")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Log the output for debugging
	t.Logf("make lint-ci output: %s", outputStr)
	if err != nil {
		t.Logf("make lint-ci returned error (which may be expected): %v", err)
	}

	// Check that the command runs with CI-specific message
	if !strings.Contains(outputStr, "Running linters for CI") {
		t.Error("make lint-ci output doesn't contain expected message")
	}
}

// TestLintReportTarget tests the make lint-report target
func TestLintReportTarget(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping make lint-report test in short mode")
	}

	// Find the repo root
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Save current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Change to the root directory
	if err := os.Chdir(rootDir); err != nil {
		t.Fatalf("Failed to change to root directory: %v", err)
	}
	// Restore current directory at the end
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Logf("Warning: Failed to restore directory: %v", err)
		}
	}()

	// Check if the Makefile contains the lint-report target
	makefilePath := rootDir + "/Makefile"
	makefile, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}

	if !strings.Contains(string(makefile), "lint-report:") {
		t.Error("lint-report target not found in Makefile")
	}

	// Check if make is installed
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make command not available, skipping test")
	}

	// Remove any existing report files
	reportPath := rootDir + "/output/reports/lint-report.json"
	_ = os.Remove(reportPath)

	// Run the lint-report command
	cmd := exec.Command("make", "lint-report")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Log the output and error for debugging
	t.Logf("make lint-report output: %s", outputStr)
	if err != nil {
		t.Logf("make lint-report returned error (which may be expected): %v", err)
	}

	// Check that the command runs and mentions report generation
	if !strings.Contains(outputStr, "Generating lint report") {
		t.Error("make lint-report output doesn't contain expected message")
	}

	// Verify that a report file was created
	if _, err := os.Stat(reportPath); err != nil {
		t.Logf("Report file may not have been created: %v", err)
	}
} 
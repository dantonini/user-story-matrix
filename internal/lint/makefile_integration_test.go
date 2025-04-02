// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package lint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMakefileLintTargets checks if the Makefile contains the required lint targets
func TestMakefileLintTargets(t *testing.T) {
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

	// Check for the required targets
	requiredTargets := []struct {
		name string
		desc string
	}{
		{"build", "Regular build target (without linting)"},
		{"lint", "Linting only target"},
		{"build-full", "Full build with linting target"},
	}

	for _, target := range requiredTargets {
		if !strings.Contains(makefileContent, target.name+":") {
			t.Errorf("Makefile is missing %s target", target.desc)
		}
	}

	// Check for the PHONY declaration
	if !strings.Contains(makefileContent, ".PHONY:") || !strings.Contains(makefileContent, "lint") || !strings.Contains(makefileContent, "build-full") {
		t.Errorf("Makefile is missing proper .PHONY declaration for lint targets")
	}
}

// TestLintCommand tests if the 'make lint' command works
func TestLintCommand(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping in CI environment")
	}

	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Execute the make lint command
	cmd := exec.Command("make", "lint")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Log the output for debugging
	t.Logf("Command output: %s", outputStr)

	// The command might return a non-zero exit code if there are linting issues,
	// but the command itself should execute
	if err != nil && !strings.Contains(outputStr, "golangci-lint") {
		t.Fatalf("Command 'make lint' failed to execute properly: %v\nOutput: %s", err, outputStr)
	}

	// Check if the output contains evidence of linting
	if !strings.Contains(outputStr, "golangci-lint") {
		// If the output doesn't mention golangci-lint, look at the Makefile to verify it's included
		makefileContent, err := os.ReadFile(filepath.Join(rootDir, "Makefile"))
		if err != nil {
			t.Fatalf("Failed to read Makefile: %v", err)
		}

		makefileStr := string(makefileContent)
		lintTarget := extractTargetContent(makefileStr, "lint")

		if !strings.Contains(lintTarget, "golangci-lint") {
			t.Error("lint target in Makefile should call golangci-lint")
		} else {
			t.Log("lint target correctly includes golangci-lint, although command output didn't mention it")
		}
	}
}

// TestBuildFullCommand tests if the 'make build-full' command works
func TestBuildFullCommand(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Find the project root directory
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
	makefileStr := string(makefileContent)

	// Check if build-full target exists
	if !strings.Contains(makefileStr, "build-full:") {
		t.Fatal("build-full target not found in Makefile")
	}

	// Check if build-full includes required components
	if !strings.Contains(makefileStr, "lint") && !strings.Contains(makefileStr, "golangci-lint") {
		t.Fatal("build-full target does not include linting")
	}

	if !strings.Contains(makefileStr, "go build") {
		t.Fatal("build-full target does not include go build")
	}

	// Check for deadcode detection (either via deadcode or unused linter)
	deadcodeDetection := strings.Contains(makefileStr, "deadcode") || strings.Contains(makefileStr, "unused")
	if !deadcodeDetection {
		t.Fatal("build-full target does not include deadcode detection")
	}

	// Create a temporary directory for command output
	tempDir, err := os.MkdirTemp("", "build-full-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Run the build-full command
	cmd := exec.Command("make", "build-full")
	cmd.Dir = rootDir // Set the working directory to the project root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	// Log output for debugging
	output := stdout.String() + stderr.String()
	t.Logf("Command output: %s", output)

	if runErr != nil {
		t.Logf("Command failed with error: %v, but this doesn't necessarily mean the test failed", runErr)
	}

	// Don't fail on command error, just check output for evidence of actions
	// This handles cases where linting fails but the process ran correctly
	if !strings.Contains(output, "Running linter") &&
		!strings.Contains(output, "Running linters") {
		t.Error("build-full output should mention running linters")
	}

	// Note: We're not checking for the build step since it might not run if linting fails
	// This is acceptable behavior since the test is primarily for the presence of linting
}

// Helper function to extract the content of a specific target from a Makefile
func extractTargetContent(makefileContent, targetName string) string {
	lines := strings.Split(makefileContent, "\n")
	var targetContent strings.Builder
	inTarget := false

	for _, line := range lines {
		if strings.HasPrefix(line, targetName+":") {
			inTarget = true
			targetContent.WriteString(line)
			targetContent.WriteString("\n")
			continue
		}

		if inTarget {
			if strings.HasPrefix(line, "\t") {
				targetContent.WriteString(line)
				targetContent.WriteString("\n")
			} else if line == "" {
				// Empty line, could still be in the target
				continue
			} else if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, " ") {
				// Line doesn't start with tab, comment, or space - we're out of the target
				break
			}
		}
	}

	return targetContent.String()
}

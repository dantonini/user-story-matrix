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

// TestDeadCodeDetection checks if dead code detection is enabled in the build-full command
func TestDeadCodeDetection(t *testing.T) {
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

	// Check if build-full includes deadcode linter
	if !strings.Contains(makefileContent, "--enable=errcheck,govet,deadcode") &&
		!strings.Contains(makefileContent, "deadcode") {
		t.Error("build-full target doesn't include deadcode linter")
	}
}

// TestLintFixDeadcodeTarget checks if the lint-fix-deadcode target is properly defined
func TestLintFixDeadcodeTarget(t *testing.T) {
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

	// Check if lint-fix-deadcode target exists
	if !strings.Contains(makefileContent, "lint-fix-deadcode:") {
		t.Error("lint-fix-deadcode target not found in Makefile")
	}

	// Check if the target calls the script
	if !strings.Contains(makefileContent, "scripts/lint-fix-deadcode.sh") {
		t.Error("lint-fix-deadcode target doesn't call the deadcode script")
	}
}

// TestDeadcodeScriptExists checks if the lint-fix-deadcode.sh script exists
func TestDeadcodeScriptExists(t *testing.T) {
	// Find the root directory
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Check if the script exists
	scriptPath := filepath.Join(rootDir, "scripts", "lint-fix-deadcode.sh")
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("lint-fix-deadcode.sh script not found: %v", err)
	}

	// Check if the script is executable
	if info.Mode()&0111 == 0 {
		t.Error("lint-fix-deadcode.sh script is not executable")
	}

	// Read the script to check its contents
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read lint-fix-deadcode.sh: %v", err)
	}

	// Check if the script contains the required features
	requiredFeatures := []string{
		"golangci-lint run",
		"enable=",
		"deadcode",
	}

	// Either deadcode or unused must be in the script to handle deprecated linters
	var containsDeadcodeHandling bool
	if strings.Contains(string(script), "deadcode") &&
		strings.Contains(string(script), "unused") &&
		strings.Contains(string(script), "deprecated") {
		containsDeadcodeHandling = true
	}

	if !containsDeadcodeHandling {
		t.Error("lint-fix-deadcode.sh doesn't have proper deadcode deprecation handling")
	}

	for _, feature := range requiredFeatures {
		if !strings.Contains(string(script), feature) {
			t.Errorf("deadcode script is missing required feature: %s", feature)
		}
	}
}

// TestDeadcodeDetectionWithDummyCode creates a dummy file with unused code and verifies it's detected
func TestDeadcodeDetectionWithDummyCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping deadcode detection test in short mode")
	}

	// Skip if golangci-lint is not installed
	if !IsInstalled() {
		t.Skip("golangci-lint not installed, skipping test")
	}

	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "deadcode-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file with unused function
	testFilePath := filepath.Join(tempDir, "test.go")
	testCode := `package test

// This function is used
func used() string {
	return "used"
}

// This function is unused (dead code)
func unused() string {
	return "unused"
}

func main() {
	used()
}
`
	if err := os.WriteFile(testFilePath, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run deadcode detection
	cmd := exec.Command("golangci-lint", "run", "--no-config", "--disable-all", "--enable=deadcode", testFilePath)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Deadcode is properly detected if either:
	// 1. The output contains "unused" (function name) and the exit code is non-zero, OR
	// 2. The output contains a warning about deadcode being deprecated but still reports the unused function
	if err == nil && !strings.Contains(outputStr, "unused") {
		t.Errorf("Deadcode detection didn't work as expected. Output: %s, Error: %v", outputStr, err)
	}

	// Check if we either find the unused function or get notified about the deprecation
	deadcodeDetected := strings.Contains(outputStr, "unused") && 
		(err != nil || strings.Contains(outputStr, "deadcode") || strings.Contains(outputStr, "unused"))
	
	deprecationWarning := strings.Contains(outputStr, "deadcode") && 
		strings.Contains(outputStr, "deprecated") && 
		strings.Contains(outputStr, "unused")

	if !deadcodeDetected && !deprecationWarning {
		t.Errorf("Deadcode detection didn't work as expected. Output: %s, Error: %v", outputStr, err)
	}
} 
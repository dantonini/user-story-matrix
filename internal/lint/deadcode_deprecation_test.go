// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package lint

import (
	"os"
	"strings"
	"testing"
)

// TestDeadcodeLinterDeprecationHandling checks if the code properly handles
// the deprecation of the deadcode linter
func TestDeadcodeLinterDeprecationHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping deadcode linter deprecation test in short mode")
	}
	
	// Check if golangci-lint is installed
	if !IsInstalled() {
		t.Skip("golangci-lint is not installed, skipping test")
	}
	
	// Get the linter version
	version, err := GetLintVersion()
	if err != nil {
		t.Fatalf("Failed to get golangci-lint version: %v", err)
	}
	
	t.Logf("Detected golangci-lint version: %s", version)
	
	// Test the getDeadCodeLinter function
	linter := getDeadCodeLinter()
	t.Logf("Using linter: %s", linter)
	
	// For golangci-lint >= 1.49.0, it should use 'unused' instead of 'deadcode'
	if strings.Contains(version, "1.4") && strings.Compare(version, "1.49.0") >= 0 {
		if linter != "unused" {
			t.Errorf("getDeadCodeLinter returned %s for version %s, expected 'unused'", linter, version)
		}
	}
	
	// Create a DeadCodeConfig and check if it uses the correct linter
	cfg := DeadCodeConfig()
	
	if len(cfg.EnabledLinters) != 1 {
		t.Errorf("DeadCodeConfig has %d linters, expected 1", len(cfg.EnabledLinters))
	} else {
		t.Logf("DeadCodeConfig uses linter: %s", cfg.EnabledLinters[0])
		
		// Make sure it matches what getDeadCodeLinter returns
		if cfg.EnabledLinters[0] != linter {
			t.Errorf("DeadCodeConfig uses %s but getDeadCodeLinter returned %s", 
				cfg.EnabledLinters[0], linter)
		}
	}
	
	// Test running the deadcode/unused linter
	output, err := Run(cfg)
	t.Logf("Linter output length: %d bytes", len(output))
	
	// Check for deprecation warnings if using 'deadcode' linter on newer versions
	if linter == "deadcode" && strings.Contains(output, "is deprecated") {
		t.Logf("Warning: deadcode linter is deprecated in this version")
	}
	
	// Make sure the linter runs without fatal errors
	if err != nil && !strings.Contains(output, "level=error") && !strings.Contains(output, "is unused") {
		// Some errors are expected (actual lint failures), but others indicate problems
		t.Errorf("Failed to run %s linter: %v", linter, err)
	}
}

// TestBuildFullContainsAlternativesToDeadcode checks if the build-full target
// correctly adapts to the deadcode linter deprecation
func TestBuildFullContainsAlternativesToDeadcode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping build-full deadcode test in short mode")
	}
	
	// Get the path to the Makefile
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}
	
	makefilePath := rootDir + "/Makefile"
	
	// Read the Makefile contents
	makefile, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}
	
	makefileContent := string(makefile)
	
	// Check if the Makefile has version-aware linter selection
	versionAware := strings.Contains(makefileContent, "DEADCODE_LINTER") ||
		strings.Contains(makefileContent, "$(DEADCODE_LINTER)") ||
		(strings.Contains(makefileContent, "deadcode") && 
		 strings.Contains(makefileContent, "unused"))
	
	if !versionAware {
		t.Error("Makefile does not appear to handle deadcode linter deprecation")
	}
	
	// Check specifically for version detection
	if !strings.Contains(makefileContent, "GOLANGCI_VERSION") {
		t.Error("Makefile does not detect golangci-lint version")
	}
	
	// If golangci-lint is installed, test if the correct linter is being used
	if IsInstalled() {
		// Get version
		version, err := GetLintVersion()
		if err != nil {
			t.Logf("Warning: Could not get golangci-lint version: %v", err)
		} else {
			t.Logf("Detected golangci-lint version: %s", version)
			
			// Determine which linter should be used
			var expectedLinter string
			if strings.Contains(version, "1.") {
				versionNumber := strings.Split(version, " ")[1]
				if strings.Compare(versionNumber, "1.49.0") >= 0 {
					expectedLinter = "unused"
				} else {
					expectedLinter = "deadcode"
				}
			} else {
				expectedLinter = "unused" // Newer major versions
			}
			
			t.Logf("Expected linter based on version: %s", expectedLinter)
			
			// The makefile should contain the correct linter for this version
			if expectedLinter == "unused" && !strings.Contains(makefileContent, "unused") {
				t.Error("Makefile does not include unused linter despite requiring it for this version")
			}
		}
	}
} 
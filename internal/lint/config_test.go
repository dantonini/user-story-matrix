// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestLintConfigContents checks if the .golangci.yml file contains the required linters
func TestLintConfigContents(t *testing.T) {
	// Find the root directory
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Read the .golangci.yml file
	configPath := filepath.Join(rootDir, ".golangci.yml")
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read .golangci.yml: %v", err)
	}

	// Parse the YAML file
	var config map[string]interface{}
	if err := yaml.Unmarshal(configContent, &config); err != nil {
		t.Fatalf("Failed to parse .golangci.yml: %v", err)
	}

	// Check if linters section exists
	lintersSection, ok := config["linters"].(map[string]interface{})
	if !ok {
		t.Fatal("Linters section not found in .golangci.yml")
	}

	// Check if disable-all is true
	disableAll, ok := lintersSection["disable-all"].(bool)
	if !ok || !disableAll {
		t.Error("disable-all should be set to true in linters section")
	}

	// Check if enable section exists
	enableSection, ok := lintersSection["enable"].([]interface{})
	if !ok {
		t.Fatal("Enable section not found in linters section")
	}

	// Required linters according to acceptance criteria
	// Note: deadcode is deprecated, so unused is an acceptable alternative
	requiredLinters := map[string]bool{
		"errcheck":    false,
		"govet":       false,
		"staticcheck": false,
	}
	
	// Track if either deadcode or unused is enabled (only one is required)
	deadcodeOrUnusedEnabled := false

	// Check if all required linters are enabled
	for _, linter := range enableSection {
		linterName, ok := linter.(string)
		if !ok {
			continue
		}
		
		if linterName == "deadcode" || linterName == "unused" {
			deadcodeOrUnusedEnabled = true
			continue
		}
		
		if _, exists := requiredLinters[linterName]; exists {
			requiredLinters[linterName] = true
		}
	}

	// Verify all required linters are enabled
	for linter, enabled := range requiredLinters {
		if !enabled {
			t.Errorf("Required linter %s is not enabled in .golangci.yml", linter)
		}
	}
	
	// Check that either deadcode or unused is enabled
	if !deadcodeOrUnusedEnabled {
		t.Error("Neither 'deadcode' nor 'unused' linter is enabled in .golangci.yml")
	}

	// Check if there are comments explaining the linters
	configStr := string(configContent)
	for linter := range requiredLinters {
		if !containsComment(configStr, linter) {
			t.Errorf("Missing comment explaining linter %s in .golangci.yml", linter)
		}
	}
	
	// Check if there's a comment for either deadcode or unused
	if !containsComment(configStr, "deadcode") && !containsComment(configStr, "unused") {
		t.Error("Missing comment explaining either 'deadcode' or 'unused' linter in .golangci.yml")
	}
}

// TestReadmeContainsLintingInfo checks if the README.md file contains linting information
func TestReadmeContainsLintingInfo(t *testing.T) {
	// Find the root directory
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Read the README.md file
	readmePath := filepath.Join(rootDir, "README.md")
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	readmeStr := string(readmeContent)

	// Check if the README contains a Code Quality section
	if !contains(readmeStr, "Code Quality") {
		t.Error("README.md does not contain a Code Quality section")
	}

	// Check if the README mentions the required linters
	requiredLinters := []string{"errcheck", "govet", "staticcheck"}
	for _, linter := range requiredLinters {
		if !contains(readmeStr, linter) {
			t.Errorf("README.md does not mention linter %s", linter)
		}
	}
	
	// Check if README mentions either deadcode or unused
	if !contains(readmeStr, "deadcode") && !contains(readmeStr, "unused") {
		t.Error("README.md does not mention either 'deadcode' or 'unused' linter")
	}

	// Check if the README contains instructions for the lint commands
	lintCommands := []string{"make lint", "make build-full"}
	for _, cmd := range lintCommands {
		if !contains(readmeStr, cmd) {
			t.Errorf("README.md does not contain instructions for %s", cmd)
		}
	}
	
	// Check for the lint-fix-deadcode command or a renamed version
	if !contains(readmeStr, "lint-fix-deadcode") && !contains(readmeStr, "lint-fix-unused") {
		t.Error("README.md does not contain instructions for removing dead/unused code")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Helper function to check if a string contains a comment about a linter
func containsComment(s, linter string) bool {
	return contains(s, linter+" ") || contains(s, "# "+linter) || contains(s, "- "+linter)
} 
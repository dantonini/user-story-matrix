// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package lint

import (
	"os"
	"testing"
)

// TestLintConfigExists checks if the .golangci.yml file exists
func TestLintConfigExists(t *testing.T) {
	// Find the root directory
	rootDir, err := FindRootDir()
	if err != nil {
		t.Fatalf("Failed to find root directory: %v", err)
	}

	// Check if the config file exists
	configPath := rootDir + "/.golangci.yml"
	_, err = os.Stat(configPath)
	if err != nil {
		t.Errorf("Config file not found: %v", err)
	}
}

// TestLintCommandWorks checks if the golangci-lint command works
func TestLintCommandWorks(t *testing.T) {
	// Skip if golangci-lint is not installed
	if !IsInstalled() {
		t.Skip("golangci-lint not installed, skipping test")
	}

	// Run a simple check
	output, err := Run(FastConfig(), ".")
	if err != nil && len(output) == 0 {
		t.Errorf("Failed to run golangci-lint: %v", err)
	}
}

// TestGetLintVersion tests if the version retrieval function works
func TestGetLintVersion(t *testing.T) {
	// Skip if golangci-lint is not installed
	if !IsInstalled() {
		t.Skip("golangci-lint not installed, skipping test")
	}

	// Get version
	version, err := GetLintVersion()
	if err != nil {
		t.Errorf("Failed to get golangci-lint version: %v", err)
	}

	// Check if version is not empty
	if version == "" {
		t.Error("Retrieved version is empty")
	}
}

// TestCreateLintReport tests if the lint report generation works
func TestCreateLintReport(t *testing.T) {
	// Skip if golangci-lint is not installed
	if !IsInstalled() {
		t.Skip("golangci-lint not installed, skipping test")
	}

	// Create a temporary file
	tempDir, err := os.MkdirTemp("", "lint-report")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	reportPath := tempDir + "/lint-report.json"

	// Generate report
	output, err := CreateLintReport(reportPath)
	if err != nil {
		t.Errorf("Failed to create lint report: %v", err)
	}

	// Check if output is not empty
	if output == "" {
		t.Error("Generated report is empty")
	}

	// Check if file was created
	_, err = os.Stat(reportPath)
	if err != nil {
		t.Errorf("Report file not created: %v", err)
	}
}

// TestConfigTypes tests different config types
func TestConfigTypes(t *testing.T) {
	// Test FastConfig
	fastCfg := FastConfig()
	if !fastCfg.Fast {
		t.Error("FastConfig.Fast should be true")
	}
	if len(fastCfg.EnabledLinters) != 2 {
		t.Errorf("FastConfig should have 2 linters, got %d", len(fastCfg.EnabledLinters))
	}

	// Test DeadCodeConfig
	deadCodeCfg := DeadCodeConfig()
	if !deadCodeCfg.VerboseOutput {
		t.Error("DeadCodeConfig.VerboseOutput should be true")
	}
	if len(deadCodeCfg.EnabledLinters) != 1 || deadCodeCfg.EnabledLinters[0] != "deadcode" {
		t.Errorf("DeadCodeConfig should only enable deadcode linter")
	}

	// Test CIConfig
	ciCfg := CIConfig()
	if !ciCfg.EnableCache {
		t.Error("CIConfig.EnableCache should be true")
	}
	if ciCfg.OutputFormat != "github-actions" {
		t.Errorf("CIConfig should use github-actions format, got %s", ciCfg.OutputFormat)
	}

	// Test TestConfig
	testCfg := TestConfig()
	if len(testCfg.Paths) != 1 || testCfg.Paths[0] != "**/*_test.go" {
		t.Errorf("TestConfig should target test files")
	}
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


// Package lint provides functionality for code linting and static analysis
package lint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config represents linting configuration options
type Config struct {
	// EnabledLinters is a list of linters to enable
	EnabledLinters []string
	// DisableAll disables all linters before enabling specific ones
	DisableAll bool
	// Fast enables only fast linters
	Fast bool
	// Fix automatically fixes issues when possible
	Fix bool
	// ConfigFile specifies the path to a config file, empty means use default
	ConfigFile string
	// VerboseOutput enables more detailed output
	VerboseOutput bool
	// Paths defines specific file paths to lint (empty = all)
	Paths []string
	// SkipDirs lists directories to skip
	SkipDirs []string
	// Timeout sets the maximum execution time
	Timeout time.Duration
	// EnableCache enables caching for faster repeated runs
	EnableCache bool
	// OnlyNewIssues only reports new issues compared to the baseline
	OnlyNewIssues bool
	// TestFiles determines whether to include test files
	TestFiles bool
	// OutputFormat sets the output format
	OutputFormat string
	// Exclude specifies patterns to exclude
	Exclude []string
}

// DefaultConfig returns the default linting configuration
func DefaultConfig() Config {
	// Check if we should use 'unused' instead of 'deadcode'
	deadcodeLinter := getDeadCodeLinter()
	
	return Config{
		EnabledLinters: []string{deadcodeLinter, "errcheck", "govet", "staticcheck"},
		DisableAll:     true,
		Fast:           false,
		Fix:            false,
		ConfigFile:     "",
		VerboseOutput:  false,
		Paths:          []string{},
		SkipDirs:       []string{"vendor"},
		Timeout:        2 * time.Minute,
		EnableCache:    true,
		OnlyNewIssues:  false,
		TestFiles:      true,
		OutputFormat:   "colored-line-number",
		Exclude:        []string{},
	}
}

// FastConfig returns a configuration for quick linting
func FastConfig() Config {
	config := DefaultConfig()
	config.EnabledLinters = []string{"errcheck", "govet"}
	config.Fast = true
	config.Timeout = 30 * time.Second
	return config
}

// DeadCodeConfig returns a configuration for dead code detection
func DeadCodeConfig() Config {
	config := DefaultConfig()
	deadcodeLinter := getDeadCodeLinter()
	config.EnabledLinters = []string{deadcodeLinter}
	config.VerboseOutput = true
	return config
}

// CIConfig returns a configuration optimized for CI environments
func CIConfig() Config {
	config := DefaultConfig()
	config.OutputFormat = "github-actions"
	config.EnableCache = true
	config.OnlyNewIssues = true
	return config
}

// TestConfig returns a configuration for linting test files
func TestConfig() Config {
	config := DefaultConfig()
	config.Paths = []string{"**/*_test.go"}
	config.EnabledLinters = []string{"errcheck", "govet", "staticcheck"}
	return config
}

// getDeadCodeLinter determines which linter to use for dead code detection
// based on the installed golangci-lint version
func getDeadCodeLinter() string {
	version, err := GetLintVersion()
	if err != nil {
		// Default to deadcode if we can't determine the version
		return "deadcode"
	}
	
	// Extract version numbers
	parts := strings.Split(version, " ")
	if len(parts) < 2 {
		return "deadcode"
	}
	
	versionStr := parts[1]
	versionParts := strings.Split(versionStr, ".")
	if len(versionParts) < 2 {
		return "deadcode"
	}
	
	// Parse major and minor versions
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return "deadcode"
	}
	
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return "deadcode"
	}
	
	// If version is >= 1.49.0, use 'unused' instead of 'deadcode'
	if major > 1 || (major == 1 && minor >= 49) {
		return "unused"
	}
	
	return "deadcode"
}

// Run executes golangci-lint with the provided configuration
// It returns the output and any error that occurred
func Run(cfg Config, paths ...string) (string, error) {
	// Check if golangci-lint is installed
	_, err := exec.LookPath("golangci-lint")
	if err != nil {
		return "", fmt.Errorf("golangci-lint not found: %w", err)
	}

	// Build command arguments
	args := []string{"run"}

	if cfg.Fast {
		args = append(args, "--fast")
	}

	if cfg.DisableAll {
		args = append(args, "--disable-all")
	}

	if len(cfg.EnabledLinters) > 0 {
		args = append(args, "--enable="+strings.Join(cfg.EnabledLinters, ","))
	}

	if cfg.Fix {
		args = append(args, "--fix")
	}

	if cfg.ConfigFile != "" {
		args = append(args, "--config", cfg.ConfigFile)
	} else if cfg.ConfigFile == "none" || cfg.ConfigFile == "false" {
		args = append(args, "--no-config")
	}

	if cfg.Timeout > 0 {
		args = append(args, fmt.Sprintf("--timeout=%s", cfg.Timeout.String()))
	}

	if cfg.EnableCache {
		args = append(args, "--cache")
	}

	if cfg.OnlyNewIssues {
		args = append(args, "--new")
	}

	if !cfg.TestFiles {
		args = append(args, "--tests=false")
	}

	if cfg.OutputFormat != "" {
		args = append(args, fmt.Sprintf("--out-format=%s", cfg.OutputFormat))
	}

	if cfg.VerboseOutput {
		args = append(args, "-v")
	}

	if len(cfg.SkipDirs) > 0 {
		for _, dir := range cfg.SkipDirs {
			args = append(args, fmt.Sprintf("--skip-dirs=%s", dir))
		}
	}

	if len(cfg.Exclude) > 0 {
		for _, pattern := range cfg.Exclude {
			args = append(args, fmt.Sprintf("--exclude=%s", pattern))
		}
	}

	// Add paths from config first
	if len(cfg.Paths) > 0 {
		args = append(args, cfg.Paths...)
	}

	// Then add explicitly specified paths
	if len(paths) > 0 {
		args = append(args, paths...)
	} else if len(cfg.Paths) == 0 {
		// Default to all packages if no path specified
		args = append(args, "./...")
	}

	// Create command
	cmd := exec.Command("golangci-lint", args...)

	// Get output
	output, err := cmd.CombinedOutput()
	
	// Return results
	return string(output), err
}

// IsInstalled checks if golangci-lint is installed and available
func IsInstalled() bool {
	_, err := exec.LookPath("golangci-lint")
	return err == nil
}

// Install attempts to install golangci-lint
func Install() error {
	fmt.Println("Installing golangci-lint...")
	cmd := exec.Command("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetLintVersion returns the installed version of golangci-lint
func GetLintVersion() (string, error) {
	if !IsInstalled() {
		return "", fmt.Errorf("golangci-lint is not installed")
	}
	
	cmd := exec.Command("golangci-lint", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get golangci-lint version: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// FindRootDir attempts to find the root directory of the project
func FindRootDir() (string, error) {
	// Start at the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Keep going up until we find a go.mod file or hit the root
	for {
		// Check if a go.mod file exists in the current directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Check if we're at the root directory
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no go.mod file found)")
		}

		// Move up to the parent directory
		dir = parent
	}
}

// CreateLintReport generates a comprehensive lint report for the project
func CreateLintReport(outputPath string) (string, error) {
	// Start with a default configuration
	cfg := DefaultConfig()
	cfg.VerboseOutput = true
	cfg.OutputFormat = "json"
	cfg.EnableCache = true
	
	// Run all linters
	output, err := Run(cfg)
	if err != nil && !strings.Contains(output, "level=error") {
		// Some lint failures are expected, but other errors should be reported
		return "", fmt.Errorf("failed to run linters: %w", err)
	}
	
	// If an output path is specified, save the report
	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return "", fmt.Errorf("failed to create output directory: %w", err)
		}
		
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return "", fmt.Errorf("failed to write lint report: %w", err)
		}
	}
	
	return output, nil
} 
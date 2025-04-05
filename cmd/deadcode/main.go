// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// UnusedResult represents the structure of staticcheck/golangci-lint unused code reports
type UnusedResult struct {
	Path string
	Line int
	Name string
	Type string // "var", "func", "type", etc.
}

// Config holds the configuration for the dead code remover
type Config struct {
	DryRun    bool
	Verbose   bool
	BackupDir string
	Packages  []string
}

func main() {
	// Parse command line flags
	var (
		dryRun    = flag.Bool("dry-run", false, "Don't modify files, just show what would be done")
		verbose   = flag.Bool("verbose", false, "Show more detailed output")
		backupDir = flag.String("backup-dir", "", "Directory to store backups (defaults to .deadcode-backups-TIMESTAMP)")
		packages  = flag.String("packages", "./...", "Comma-separated list of packages to process")
	)
	flag.Parse()

	// Create backup directory if not in dry-run mode
	timestamp := time.Now().Format("20060102150405")
	if *backupDir == "" {
		*backupDir = fmt.Sprintf(".deadcode-backups-%s", timestamp)
	}

	if !*dryRun {
		if err := os.MkdirAll(*backupDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating backup directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Backups will be stored in %s\n", *backupDir)
	}

	// Parse packages
	pkgs := strings.Split(*packages, ",")

	// Create configuration
	config := Config{
		DryRun:    *dryRun,
		Verbose:   *verbose,
		BackupDir: *backupDir,
		Packages:  pkgs,
	}

	// Find unused code using staticcheck or golangci-lint
	unusedCode, err := findUnusedCode(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding unused code: %v\n", err)
		os.Exit(1)
	}

	if len(unusedCode) == 0 {
		fmt.Println("No unused code found!")
		return
	}

	fmt.Printf("Found %d instances of unused code\n", len(unusedCode))

	// Group by file path
	fileGroups := groupByFile(unusedCode)

	// Clean up each file
	for filePath, results := range fileGroups {
		err := cleanupFile(filePath, results, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error cleaning up file %s: %v\n", filePath, err)
		}
	}
}

// findUnusedCode finds unused code in the given packages
func findUnusedCode(config Config) ([]UnusedResult, error) {
	var results []UnusedResult

	// Try different approaches to find unused code
	results, err := findUnusedCodeUsingGolangciLint(config)
	if err != nil && config.Verbose {
		fmt.Printf("golangci-lint check failed: %v\n", err)
		fmt.Println("Trying alternate methods...")
	}

	// If we found results, return them
	if len(results) > 0 {
		return results, nil
	}

	// Try using staticcheck directly
	results, err = findUnusedCodeUsingStaticcheck(config)
	if err != nil && config.Verbose {
		fmt.Printf("staticcheck check failed: %v\n", err)
	}

	// If still no results, try simple grep for "is unused" comments
	if len(results) == 0 {
		if config.Verbose {
			fmt.Println("Using go vet to find potential unused code...")
		}
		results, _ = findUnusedCodeUsingGoVet(config)
	}

	return results, nil
}

// findUnusedCodeUsingGolangciLint uses golangci-lint to find unused code
func findUnusedCodeUsingGolangciLint(config Config) ([]UnusedResult, error) {
	if config.Verbose {
		fmt.Println("Trying golangci-lint to find unused code...")
	}

	var results []UnusedResult
	var outBuf strings.Builder

	// Run golangci-lint
	cmd := exec.Command("golangci-lint", "run", "--no-config", "--disable-all", "--enable=unused", "--out-format=json")
	cmd.Args = append(cmd.Args, config.Packages...)

	if config.Verbose {
		fmt.Printf("Running: %s\n", strings.Join(cmd.Args, " "))
	}

	// Connect stdout to our buffer
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start golangci-lint: %v", err)
	}

	// Read the output
	if _, err := io.Copy(&outBuf, cmdOut); err != nil {
		return nil, fmt.Errorf("failed to read output: %v", err)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		if config.Verbose {
			fmt.Printf("golangci-lint exited with error: %v\n", err)
		}
		// Continue even with error, as we may have partial results
	}

	// Parse the JSON output
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(outBuf.String()), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse golangci-lint output: %v", err)
	}

	// Extract issues
	if issues, ok := rawData["Issues"].([]interface{}); ok {
		for _, issue := range issues {
			if m, ok := issue.(map[string]interface{}); ok {
				// Skip if not from the unused linter
				fromLinter, ok := m["FromLinter"].(string)
				if !ok || fromLinter != "unused" {
					continue
				}

				// Get position information
				pos, ok := m["Pos"].(map[string]interface{})
				if !ok {
					continue
				}

				filename, ok := pos["Filename"].(string)
				if !ok {
					continue
				}

				line, ok := pos["Line"].(float64)
				if !ok {
					continue
				}

				text, ok := m["Text"].(string)
				if !ok {
					continue
				}

				result := UnusedResult{
					Path: filename,
					Line: int(line),
					Name: text,
				}

				// Extract type and name from the message
				if strings.Contains(text, "func") {
					result.Type = "func"
				} else if strings.Contains(text, "var") {
					result.Type = "var"
				} else if strings.Contains(text, "type") {
					result.Type = "type"
				} else if strings.Contains(text, "const") {
					result.Type = "const"
				}

				// Extract the name - format is typically "xxx is unused"
				parts := strings.SplitN(text, " is ", 2)
				if len(parts) > 1 {
					nameParts := strings.Split(parts[0], " ")
					if len(nameParts) > 1 {
						result.Name = nameParts[len(nameParts)-1]
					} else {
						result.Name = parts[0]
					}
				}

				// Remove any backticks
				result.Name = strings.Trim(result.Name, "`")

				results = append(results, result)
			}
		}
	}

	if config.Verbose {
		fmt.Printf("Found %d unused code elements using golangci-lint\n", len(results))
	}

	return results, nil
}

// findUnusedCodeUsingStaticcheck uses staticcheck directly to find unused code
func findUnusedCodeUsingStaticcheck(config Config) ([]UnusedResult, error) {
	if config.Verbose {
		fmt.Println("Trying staticcheck to find unused code...")
	}

	var results []UnusedResult

	// Create temp file for output
	tempFile, err := ioutil.TempFile("", "staticcheck-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Run staticcheck - different versions have different flag formats
	cmd := exec.Command("staticcheck", "-checks=U1000", "-f=json")
	cmd.Args = append(cmd.Args, config.Packages...)

	if config.Verbose {
		fmt.Printf("Running: %s\n", strings.Join(cmd.Args, " "))
	}

	// First, try modern format with -f=json
	outBuf, err := cmd.CombinedOutput()
	if err != nil {
		if config.Verbose {
			fmt.Printf("First staticcheck attempt failed: %v\n", err)
			fmt.Printf("Output: %s\n", outBuf)
		}

		// Try older format
		cmd = exec.Command("staticcheck", "-checks=U1000")
		cmd.Args = append(cmd.Args, config.Packages...)

		if config.Verbose {
			fmt.Printf("Running: %s\n", strings.Join(cmd.Args, " "))
		}

		outBuf, err = cmd.CombinedOutput()
		if err != nil && config.Verbose {
			fmt.Printf("Second staticcheck attempt failed: %v\n", err)
			fmt.Printf("Output: %s\n", outBuf)
		}
	}

	// Parse the output - try as JSON first
	err = json.Unmarshal(outBuf, &results)
	if err != nil {
		if config.Verbose {
			fmt.Printf("Failed to parse as JSON, trying line-by-line: %v\n", err)
		}

		// Parse line by line format instead
		// Example: file.go:123:45: func foo is unused (U1000)
		lines := strings.Split(string(outBuf), "\n")
		for _, line := range lines {
			if strings.Contains(line, "U1000") {
				parts := strings.Split(line, ":")
				if len(parts) < 4 {
					continue
				}

				filename := parts[0]
				lineNum := 0
				fmt.Sscanf(parts[1], "%d", &lineNum)

				// Extract the message part
				msgParts := strings.SplitN(line, ": ", 2)
				if len(msgParts) < 2 {
					continue
				}

				message := msgParts[1]
				// Strip " (U1000)" suffix
				message = strings.TrimSuffix(message, " (U1000)")

				result := UnusedResult{
					Path: filename,
					Line: lineNum,
					Name: message,
				}

				// Extract type and name
				if strings.HasPrefix(message, "func ") {
					result.Type = "func"
					result.Name = strings.TrimPrefix(message, "func ")
					result.Name = strings.Split(result.Name, " ")[0]
				} else if strings.HasPrefix(message, "var ") {
					result.Type = "var"
					result.Name = strings.TrimPrefix(message, "var ")
					result.Name = strings.Split(result.Name, " ")[0]
				} else if strings.HasPrefix(message, "type ") {
					result.Type = "type"
					result.Name = strings.TrimPrefix(message, "type ")
					result.Name = strings.Split(result.Name, " ")[0]
				} else if strings.HasPrefix(message, "const ") {
					result.Type = "const"
					result.Name = strings.TrimPrefix(message, "const ")
					result.Name = strings.Split(result.Name, " ")[0]
				}

				// Remove any backticks
				result.Name = strings.Trim(result.Name, "`")

				results = append(results, result)
			}
		}
	}

	if config.Verbose {
		fmt.Printf("Found %d unused code elements using staticcheck\n", len(results))
	}

	return results, nil
}

// findUnusedCodeUsingGoVet uses go vet to find unused code
func findUnusedCodeUsingGoVet(config Config) ([]UnusedResult, error) {
	if config.Verbose {
		fmt.Println("Trying go vet to find unused code...")
	}

	var results []UnusedResult

	// Run go vet
	cmd := exec.Command("go", "vet", "-unusedresult")
	cmd.Args = append(cmd.Args, config.Packages...)

	if config.Verbose {
		fmt.Printf("Running: %s\n", strings.Join(cmd.Args, " "))
	}

	outBuf, _ := cmd.CombinedOutput()
	lines := strings.Split(string(outBuf), "\n")

	for _, line := range lines {
		if strings.Contains(line, "is unused") || strings.Contains(line, "not used") {
			parts := strings.Split(line, ":")
			if len(parts) < 3 {
				continue
			}

			filename := parts[0]
			lineNum := 0
			fmt.Sscanf(parts[1], "%d", &lineNum)

			// Extract the message part
			msgParts := strings.SplitN(line, ": ", 2)
			if len(msgParts) < 2 {
				continue
			}

			message := msgParts[1]

			result := UnusedResult{
				Path: filename,
				Line: lineNum,
				Name: message,
			}

			// Try to extract the name from message
			if strings.Contains(message, "is unused") {
				parts := strings.Split(message, " is unused")
				if len(parts) > 0 {
					result.Name = parts[0]
				}
			}

			// Remove any backticks
			result.Name = strings.Trim(result.Name, "`")

			results = append(results, result)
		}
	}

	if config.Verbose {
		fmt.Printf("Found %d unused code elements using go vet\n", len(results))
	}

	return results, nil
}

// groupByFile groups the unused results by file path for efficient processing
func groupByFile(results []UnusedResult) map[string][]UnusedResult {
	fileGroups := make(map[string][]UnusedResult)
	for _, result := range results {
		fileGroups[result.Path] = append(fileGroups[result.Path], result)
	}
	return fileGroups
}

// cleanupFile removes unused code from a single file using AST parsing and modification
func cleanupFile(filePath string, unusedResults []UnusedResult, config Config) error {
	if config.Verbose {
		fmt.Printf("Processing file: %s\n", filePath)
	}

	// Create a map of unused symbols for faster lookup
	unusedSymbols := make(map[string]bool)
	for _, result := range unusedResults {
		// Extract the actual name
		symbolName := strings.TrimSpace(result.Name)
		// Remove any "is unused" or similar suffixes
		symbolName = strings.Split(symbolName, " is ")[0]
		
		if config.Verbose {
			fmt.Printf("  - Found unused symbol: %s (type: %s)\n", symbolName, result.Type)
		}
		unusedSymbols[symbolName] = true
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %v", err)
	}

	// Create a new list of declarations, omitting the unused ones
	var newDecls []ast.Decl
	var removedCount int

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Check if this function is unused
			if unusedSymbols[d.Name.Name] {
				if config.Verbose {
					fmt.Printf("  - Removing unused function: %s\n", d.Name.Name)
				}
				removedCount++
				continue
			}
		case *ast.GenDecl:
			// Handle var, const, and type declarations
			var newSpecs []ast.Spec
			removedFromDecl := false

			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					// Remove unused variables
					var newNames []*ast.Ident
					var newValues []ast.Expr

					for i, name := range s.Names {
						if unusedSymbols[name.Name] {
							if config.Verbose {
								fmt.Printf("  - Removing unused variable: %s\n", name.Name)
							}
							removedCount++
							removedFromDecl = true
							continue
						}
						newNames = append(newNames, name)
						if i < len(s.Values) {
							newValues = append(newValues, s.Values[i])
						}
					}

					if len(newNames) > 0 {
						s.Names = newNames
						s.Values = newValues
						newSpecs = append(newSpecs, s)
					}

				case *ast.TypeSpec:
					// Remove unused types
					if unusedSymbols[s.Name.Name] {
						if config.Verbose {
							fmt.Printf("  - Removing unused type: %s\n", s.Name.Name)
						}
						removedCount++
						removedFromDecl = true
						continue
					}
					newSpecs = append(newSpecs, s)

				default:
					newSpecs = append(newSpecs, s)
				}
			}

			// If we removed some specs but not all, update the declaration
			if removedFromDecl && len(newSpecs) > 0 {
				d.Specs = newSpecs
				newDecls = append(newDecls, d)
			} else if !removedFromDecl {
				// If nothing was removed, keep the entire declaration
				newDecls = append(newDecls, d)
			}
			// If all specs were removed, drop the entire declaration

		default:
			// Keep other declaration types
			newDecls = append(newDecls, decl)
		}
	}

	// No unused code found in this file
	if removedCount == 0 {
		if config.Verbose {
			fmt.Printf("  - No unused code to remove in this file\n")
		}
		return nil
	}

	// Update the AST with the new declarations
	node.Decls = newDecls

	// Create a backup of the original file if not in dry-run mode
	if !config.DryRun {
		backupPath := filepath.Join(config.BackupDir, filepath.Base(filePath))
		if err := copyFile(filePath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
	}

	// Format and write the modified AST back to the file
	if !config.DryRun {
		var buf strings.Builder
		if err := format.Node(&buf, fset, node); err != nil {
			return fmt.Errorf("failed to format modified code: %v", err)
		}

		if err := ioutil.WriteFile(filePath, []byte(buf.String()), 0644); err != nil {
			return fmt.Errorf("failed to write modified file: %v", err)
		}

		fmt.Printf("Removed %d unused elements from %s (backup created)\n", removedCount, filePath)
	} else {
		fmt.Printf("Would remove %d unused elements from %s (dry run)\n", removedCount, filePath)
	}

	return nil
}

// copyFile creates a copy of a file
func copyFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
} 
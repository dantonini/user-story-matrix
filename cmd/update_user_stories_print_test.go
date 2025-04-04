// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// captureOutput captures standard output for testing
func captureOutput(fn func()) string {
	// Save original output
	originalStdout := os.Stdout
	
	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Run the function that produces output
	fn()
	
	// Close the writer and restore original stdout
	w.Close()
	os.Stdout = originalStdout
	
	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	
	return buf.String()
}

// TestPrintGroupedFiles_EmptyList tests printGroupedFiles with an empty list
func TestPrintGroupedFiles_EmptyList(t *testing.T) {
	output := captureOutput(func() {
		printGroupedFiles([]string{}, "")
	})
	
	// Verify no output is produced for empty list
	assert.Equal(t, "", output, "Empty list should produce no output")
}

// TestPrintGroupedFiles_SingleDirectory tests printGroupedFiles with files in one directory
func TestPrintGroupedFiles_SingleDirectory(t *testing.T) {
	// Files in a single directory
	files := []string{
		"docs/user-stories/story1.md",
		"docs/user-stories/story2.md",
		"docs/user-stories/epic1.md",
	}
	
	output := captureOutput(func() {
		printGroupedFiles(files, "")
	})
	
	// Verify output contains expected directory and files
	assert.Contains(t, output, "📁 docs/user-stories/")
	assert.Contains(t, output, "• story1.md")
	assert.Contains(t, output, "• story2.md")
	assert.Contains(t, output, "• epic1.md")
	
	// Verify order (directory first, then files)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.GreaterOrEqual(t, len(lines), 4, "Should have at least 4 lines of output")
	assert.Contains(t, lines[0], "📁 docs/user-stories/", "First line should be the directory")
	
	// Verify files are indented
	for i := 1; i < len(lines); i++ {
		assert.Contains(t, lines[i], "  • ", "File lines should be indented with bullets")
	}
}

// TestPrintGroupedFiles_MultipleDirectories tests printGroupedFiles with files in multiple directories
func TestPrintGroupedFiles_MultipleDirectories(t *testing.T) {
	// Files in multiple directories
	files := []string{
		"docs/user-stories/story1.md",
		"docs/user-stories/story2.md",
		"docs/epics/epic1.md",
		"docs/epics/epic2.md",
		"features/feature1.md",
	}
	
	output := captureOutput(func() {
		printGroupedFiles(files, "  ") // With indent
	})
	
	// Verify output contains all directories and files
	assert.Contains(t, output, "docs/user-stories/")
	assert.Contains(t, output, "docs/epics/")
	assert.Contains(t, output, "features/")
	
	assert.Contains(t, output, "story1.md")
	assert.Contains(t, output, "story2.md")
	assert.Contains(t, output, "epic1.md")
	assert.Contains(t, output, "epic2.md")
	assert.Contains(t, output, "feature1.md")
	
	// Verify the output is not empty and has sufficient lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.GreaterOrEqual(t, len(lines), 8, "Output should have at least 8 lines (3 directories + 5 files)")
}

// TestPrintGroupedFiles_EdgeCases tests printGroupedFiles with special cases
func TestPrintGroupedFiles_EdgeCases(t *testing.T) {
	// Edge cases: root files, nested directories
	files := []string{
		"root-file.md",                           // File in root
		"deeply/nested/directory/structure.md",   // Deeply nested file
		".hidden/file.md",                        // Hidden directory
	}
	
	output := captureOutput(func() {
		printGroupedFiles(files, "")
	})
	
	// Verify all directories and files are present
	assert.Contains(t, output, "📁 .")
	assert.Contains(t, output, "📁 deeply/nested/directory")
	assert.Contains(t, output, "📁 .hidden")
	
	assert.Contains(t, output, "• root-file.md")
	assert.Contains(t, output, "• structure.md")
	assert.Contains(t, output, "• file.md")
} 
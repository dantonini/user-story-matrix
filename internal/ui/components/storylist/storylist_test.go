// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storylist

import (
	"fmt"
	"testing"
	"time"
)

func TestCalculateCommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name:     "empty paths",
			paths:    []string{},
			expected: "",
		},
		{
			name:     "single path",
			paths:    []string{"docs/user-stories/dir1/file1.md"},
			expected: "docs/user-stories/dir1/file1.md",
		},
		{
			name: "multiple paths with common prefix",
			paths: []string{
				"docs/user-stories/dir1/file1.md",
				"docs/user-stories/dir1/file2.md",
				"docs/user-stories/dir1/file3.md",
			},
			expected: "docs/user-stories/dir1",
		},
		{
			name: "multiple paths with partial common prefix",
			paths: []string{
				"docs/user-stories/dir1/file1.md",
				"docs/user-stories/dir2/file2.md",
				"docs/user-stories/dir3/file3.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "no common prefix",
			paths: []string{
				"docs/user-stories/file1.md",
				"src/components/file2.md",
				"test/file3.md",
			},
			expected: "",
		},
		{
			name: "common prefix at root",
			paths: []string{
				"docs/user-stories/file1.md",
				"docs/code/file2.md",
				"docs/tests/file3.md",
			},
			expected: "docs",
		},
		{
			name: "exact same paths",
			paths: []string{
				"docs/user-stories/file.md",
				"docs/user-stories/file.md",
				"docs/user-stories/file.md",
			},
			expected: "docs/user-stories/file.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateCommonPrefix(tc.paths)
			if result != tc.expected {
				t.Errorf("Expected common prefix to be %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestShortenPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		commonPrefix string
		expected     string
	}{
		{
			name:         "empty path",
			path:         "",
			commonPrefix: "docs/user-stories",
			expected:     "",
		},
		{
			name:         "empty common prefix",
			path:         "docs/user-stories/file.md",
			commonPrefix: "",
			expected:     "docs/user-stories/file.md",
		},
		{
			name:         "path contains common prefix",
			path:         "docs/user-stories/dir1/file.md",
			commonPrefix: "docs/user-stories",
			expected:     "…/dir1/file.md",
		},
		{
			name:         "path does not contain common prefix",
			path:         "src/components/file.md",
			commonPrefix: "docs/user-stories",
			expected:     "src/components/file.md",
		},
		{
			name:         "path equals common prefix",
			path:         "docs/user-stories",
			commonPrefix: "docs/user-stories",
			expected:     "docs/user-stories",
		},
		{
			name:         "common prefix is the entire path",
			path:         "docs/user-stories/dir1/file.md",
			commonPrefix: "docs/user-stories/dir1/file.md",
			expected:     "docs/user-stories/dir1/file.md", // No shortening as it would be empty
		},
		{
			name:         "path with trailing slash in common prefix",
			path:         "docs/user-stories/dir1/file.md",
			commonPrefix: "docs/user-stories/",
			expected:     "…/dir1/file.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := shortenPath(tc.path, tc.commonPrefix)
			if result != tc.expected {
				t.Errorf("Expected shortened path to be %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestCalculateCommonPrefixEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name: "paths with mixed case should be case sensitive",
			paths: []string{
				"docs/User-Stories/file1.md",
				"docs/user-stories/file2.md",
			},
			expected: "docs",
		},
		{
			name: "paths with trailing slashes",
			paths: []string{
				"docs/user-stories/dir1/",
				"docs/user-stories/dir2/",
			},
			expected: "docs/user-stories",
		},
		{
			name: "paths with varying depths",
			paths: []string{
				"docs/user-stories/dir1/subdir/file1.md",
				"docs/user-stories/dir1/file2.md",
				"docs/user-stories/file3.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "subset paths",
			paths: []string{
				"docs/user-stories",
				"docs/user-stories/dir1/file.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "paths with special characters",
			paths: []string{
				"docs/user-stories/test-file-1.md",
				"docs/user-stories/test_file_2.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "paths with numbers",
			paths: []string{
				"docs/user-stories/1-introduction.md",
				"docs/user-stories/2-setup.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "absolute paths",
			paths: []string{
				"/home/user/docs/user-stories/file1.md",
				"/home/user/docs/user-stories/file2.md",
			},
			expected: "/home/user/docs/user-stories",
		},
		{
			name: "mixture of absolute and relative paths",
			paths: []string{
				"/docs/user-stories/file1.md",
				"docs/user-stories/file2.md",
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateCommonPrefix(tc.paths)
			if result != tc.expected {
				t.Errorf("Expected common prefix to be %q, got %q", tc.expected, result)
			}
		})
	}
}

// TestCalculateCommonPrefixBenchmark is a helper to validate the performance
// of the common prefix calculation with a large number of paths
func TestCalculateCommonPrefixBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark test in short mode")
	}
	
	// Create a large number of paths with a common prefix
	paths := make([]string, 1000)
	for i := range paths {
		paths[i] = fmt.Sprintf("docs/user-stories/dir%d/file%d.md", i%10, i)
	}
	
	start := time.Now()
	result := calculateCommonPrefix(paths)
	duration := time.Since(start)
	
	expected := "docs/user-stories"
	if result != expected {
		t.Errorf("Expected common prefix to be %q, got %q", expected, result)
	}
	
	t.Logf("Calculated common prefix for %d paths in %v", len(paths), duration)
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user-story-matrix/usm/internal/io"
)

// withTempDir creates a temporary directory for testing and cleans it up afterward
func withTempDir(t *testing.T, fn func(dir string, fs io.FileSystem)) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "usm-test-*")
	require.NoError(t, err, "Failed to create temporary directory")
	
	// Clean up after test
	defer os.RemoveAll(tempDir)
	
	// Run test with real filesystem
	fs := io.NewOSFileSystem()
	fn(tempDir, fs)
}

// TestIntegration_SimplestCase tests the most basic update scenario
func TestIntegration_SimplestCase(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Log the temp directory
		fmt.Printf("Using temporary directory: %s\n", tempDir)
		
		// Create a single file with empty metadata block
		docsDir := filepath.Join(tempDir, "docs")
		require.NoError(t, fs.MkdirAll(docsDir, 0755))
		
		storyPath := filepath.Join(docsDir, "story.md")
		emptyMetadata := "---\n\n---\n"
		content := emptyMetadata + "# Test Story\n\nThis is a test story."
		require.NoError(t, fs.WriteFile(storyPath, []byte(content), 0644))
		
		// Get the content before update for verification
		originalContent, err := fs.ReadFile(storyPath)
		require.NoError(t, err)
		fmt.Println("Original content:", string(originalContent))
		
		// Run UpdateFileMetadata directly to test just that function
		updated, hashMap, err := UpdateFileMetadata(storyPath, tempDir, fs)
		require.NoError(t, err)
		
		// Check results
		assert.True(t, updated, "File should be marked as updated")
		assert.NotEmpty(t, hashMap.NewHash, "New hash should not be empty")
		
		// Verify file content was updated properly
		newContent, err := fs.ReadFile(storyPath)
		require.NoError(t, err)
		fmt.Println("New content:", string(newContent))
		
		assert.True(t, strings.HasPrefix(string(newContent), "---"), "Content should start with metadata delimiter")
		assert.Contains(t, string(newContent), "file_path:", "Metadata should contain file_path")
		assert.Contains(t, string(newContent), "created_at:", "Metadata should contain created_at")
		assert.Contains(t, string(newContent), "last_updated:", "Metadata should contain last_updated")
		assert.Contains(t, string(newContent), "_content_hash:", "Metadata should contain content_hash")
		assert.Contains(t, string(newContent), "# Test Story", "Original content should be preserved")
	})
}

// Simple integration test for UpdateAllUserStoryMetadata
func TestIntegration_UpdateAllUserStoryMetadata(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		fmt.Printf("Using temporary directory: %s\n", tempDir)
		
		// Create directory structure
		docsDir := filepath.Join(tempDir, "docs")
		userStoriesDir := filepath.Join(docsDir, "user-stories")
		require.NoError(t, fs.MkdirAll(userStoriesDir, 0755))
		
		// Create two types of files
		// 1. A file with no metadata 
		story1Path := filepath.Join(userStoriesDir, "story1.md")
		story1Content := "# Test Story 1\n\nThis is a story without metadata."
		require.NoError(t, fs.WriteFile(story1Path, []byte(story1Content), 0644))
		
		// 2. A file with existing metadata that needs updating
		story2Path := filepath.Join(userStoriesDir, "story2.md")
		story2Content := `---
file_path: old/path/story2.md
created_at: 2022-05-15T10:30:00Z
last_updated: 2022-05-16T10:30:00Z
_content_hash: oldhash123
---

# Test Story 2
This story has metadata but needs updating.
`
		require.NoError(t, fs.WriteFile(story2Path, []byte(story2Content), 0644))
		
		// Test initial state
		fmt.Println("===== BEFORE UPDATE =====")
		content1, _ := fs.ReadFile(story1Path)
		fmt.Println("Story1 content:", string(content1))
		content2, _ := fs.ReadFile(story2Path)
		fmt.Println("Story2 content:", string(content2))
		
		// Key point: UpdateAllUserStoryMetadata expects:
		// - userStoriesDir: The directory to scan for markdown files
		// - rootDir: The base directory for relative paths in metadata
		fmt.Println("\n===== RUNNING UpdateAllUserStoryMetadata =====")
		updated, unchanged, hashMap, err := UpdateAllUserStoryMetadata(docsDir, tempDir, fs)
		require.NoError(t, err)
		
		// Print detailed info for debugging
		fmt.Println("\n===== RESULTS =====")
		fmt.Println("Updated files:", updated)
		fmt.Println("Unchanged files:", unchanged)
		fmt.Println("Hash map:", hashMap)
		
		// Get relative paths
		relPath1, err := filepath.Rel(tempDir, story1Path)
		require.NoError(t, err)
		relPath2, err := filepath.Rel(tempDir, story2Path)
		require.NoError(t, err)
		
		// Test final state
		fmt.Println("\n===== AFTER UPDATE =====")
		newContent1, _ := fs.ReadFile(story1Path)
		fmt.Println("Story1 content:", string(newContent1))
		newContent2, _ := fs.ReadFile(story2Path)
		fmt.Println("Story2 content:", string(newContent2))
		
		// Check results for both files
		assert.Contains(t, updated, relPath1, "Story1 should be in updated list")
		assert.Contains(t, updated, relPath2, "Story2 should be in updated list")
		assert.Equal(t, 2, len(updated), "Should have 2 updated files")
		assert.Equal(t, 0, len(unchanged), "Should have 0 unchanged files")
		assert.Contains(t, hashMap, relPath1, "Story1 should be in hash map")
		assert.Contains(t, hashMap, relPath2, "Story2 should be in hash map")
		assert.Equal(t, 2, len(hashMap), "Hash map should have 2 entries")
		
		// Verify file content was updated with metadata
		// Story1
		assert.True(t, strings.HasPrefix(string(newContent1), "---"), "Story1 should start with metadata delimiter")
		assert.Contains(t, string(newContent1), "file_path: "+relPath1, "Metadata should contain correct relative path")
		assert.Contains(t, string(newContent1), "created_at:", "Metadata should contain creation date")
		assert.Contains(t, string(newContent1), "last_updated:", "Metadata should contain last updated date")
		assert.Contains(t, string(newContent1), "_content_hash:", "Metadata should contain content hash")
		assert.Contains(t, string(newContent1), "# Test Story 1", "Original content should be preserved")
		
		// Story2
		assert.True(t, strings.HasPrefix(string(newContent2), "---"), "Story2 should start with metadata delimiter")
		assert.Contains(t, string(newContent2), "file_path: "+relPath2, "Metadata should contain updated relative path")
		assert.Contains(t, string(newContent2), "created_at: 2022-05-15T10:30:00Z", "Metadata should preserve original creation date")
		assert.Contains(t, string(newContent2), "last_updated:", "Metadata should contain updated last_updated date")
		assert.Contains(t, string(newContent2), "_content_hash:", "Metadata should contain updated content hash")
		assert.NotContains(t, string(newContent2), "old/path/story2.md", "Old path should be updated")
		assert.NotContains(t, string(newContent2), "_content_hash: oldhash123", "Old hash should be updated")
		assert.Contains(t, string(newContent2), "# Test Story 2", "Original content should be preserved")
		
		// Run again - files should be unchanged this time
		fmt.Println("\n===== RUNNING SECOND TIME =====")
		updated2, unchanged2, hashMap2, err := UpdateAllUserStoryMetadata(docsDir, tempDir, fs)
		require.NoError(t, err)
		
		// Print detailed info again
		fmt.Println("\n===== SECOND RUN RESULTS =====")
		fmt.Println("Updated files:", updated2)
		fmt.Println("Unchanged files:", unchanged2)
		fmt.Println("Hash map:", hashMap2)
		
		assert.Equal(t, 0, len(updated2), "Should have 0 updated files on second run")
		assert.Equal(t, 2, len(unchanged2), "Should have 2 unchanged files on second run")
		assert.Contains(t, unchanged2, relPath1, "Story1 should be in unchanged list on second run")
		assert.Contains(t, unchanged2, relPath2, "Story2 should be in unchanged list on second run")
		assert.Equal(t, 0, len(hashMap2), "Hash map should be empty on second run")
	})
}

// TestIntegration_UpdateAllUserStoryMetadata_EmptyDirectory tests the empty directory case
func TestIntegration_UpdateAllUserStoryMetadata_EmptyDirectory(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create empty directory
		userStoriesDir := filepath.Join(tempDir, "docs", "user-stories")
		require.NoError(t, fs.MkdirAll(userStoriesDir, 0755))
		
		// Run UpdateAllUserStoryMetadata on empty directory
		updated, unchanged, hashMap, err := UpdateAllUserStoryMetadata(
			userStoriesDir,
			tempDir,
			fs,
		)
		
		// Verify results
		require.NoError(t, err)
		assert.Empty(t, updated, "No files should be updated in empty directory")
		assert.Empty(t, unchanged, "No files should be unchanged in empty directory")
		assert.Empty(t, hashMap, "Hash map should be empty for empty directory")
	})
}

// TestIntegration_UpdateAllUserStoryMetadata_NonexistentDirectory tests handling of non-existent directories
func TestIntegration_UpdateAllUserStoryMetadata_NonexistentDirectory(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Try to use a directory that doesn't exist
		nonexistentDir := filepath.Join(tempDir, "does-not-exist")
		
		// Run UpdateAllUserStoryMetadata on non-existent directory
		updated, unchanged, hashMap, err := UpdateAllUserStoryMetadata(
			nonexistentDir,
			tempDir,
			fs,
		)
		
		// Verify results - should return error
		assert.Error(t, err, "Should return error for non-existent directory")
		assert.Empty(t, updated, "No files should be updated for non-existent directory")
		assert.Empty(t, unchanged, "No files should be unchanged for non-existent directory")
		assert.Empty(t, hashMap, "Hash map should be empty for non-existent directory")
	})
}

// Multiple file test
func TestIntegration_UpdateAllUserStoryMetadata_Multiple(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		docsDir := filepath.Join(tempDir, "docs")
		userStoriesDir := filepath.Join(docsDir, "user-stories")
		epicsDir := filepath.Join(userStoriesDir, "epics")
		skipDir := filepath.Join(tempDir, "node_modules") // Should be skipped
		
		// Create directories
		require.NoError(t, fs.MkdirAll(userStoriesDir, 0755))
		require.NoError(t, fs.MkdirAll(epicsDir, 0755))
		require.NoError(t, fs.MkdirAll(skipDir, 0755))
		
		// Create test files
		// 1. User story with no metadata (should add metadata)
		story1Path := filepath.Join(userStoriesDir, "story1.md")
		story1Content := "# Story 1\n\nThis is a story without metadata."
		require.NoError(t, fs.WriteFile(story1Path, []byte(story1Content), 0644))
		
		// 2. User story with outdated metadata (should update)
		story2Path := filepath.Join(userStoriesDir, "story2.md")
		story2Content := `---
file_path: docs/user-stories/story2.md
created_at: 2022-05-15T10:30:00Z
last_updated: 2022-05-16T10:30:00Z
_content_hash: oldhash123
---

# Story 2
This story has metadata but with an outdated hash.
`
		require.NoError(t, fs.WriteFile(story2Path, []byte(story2Content), 0644))
		
		// 3. Story in subdirectory (should process too)
		epicPath := filepath.Join(epicsDir, "epic1.md")
		epicContent := "# Epic 1\n\nThis is an epic without metadata."
		require.NoError(t, fs.WriteFile(epicPath, []byte(epicContent), 0644))
		
		// 4. File in skipped directory (should be ignored)
		skipFilePath := filepath.Join(skipDir, "skip.md")
		skipContent := "# Skip\n\nThis file should be skipped."
		require.NoError(t, fs.WriteFile(skipFilePath, []byte(skipContent), 0644))
		
		// 5. Non-markdown file (should be ignored)
		txtPath := filepath.Join(userStoriesDir, "notes.txt")
		txtContent := "Just some notes, not markdown."
		require.NoError(t, fs.WriteFile(txtPath, []byte(txtContent), 0644))
		
		// Run UpdateAllUserStoryMetadata
		updated, unchanged, hashMap, err := UpdateAllUserStoryMetadata(
			docsDir, // directory to scan
			tempDir, // root directory for relative paths
			fs,
		)
		
		// Verify basic results
		require.NoError(t, err)
		
		// Expected relative paths
		relStory1, _ := filepath.Rel(tempDir, story1Path)
		relStory2, _ := filepath.Rel(tempDir, story2Path)
		relEpic, _ := filepath.Rel(tempDir, epicPath)
		
		// Verify files were processed correctly
		assert.Contains(t, updated, relStory1, "Story1 should be in the updated list")
		assert.Contains(t, updated, relStory2, "Story2 should be in the updated list")
		assert.Contains(t, updated, relEpic, "Epic should be in the updated list")
		assert.Len(t, updated, 3, "Three files should be updated")
		assert.Len(t, unchanged, 0, "No files should be unchanged")
		
		// Check the hashMap (keys are relative paths)
		assert.Contains(t, hashMap, relStory1, "Story1 should be in the hash map")
		assert.Contains(t, hashMap, relStory2, "Story2 should be in the hash map")
		assert.Contains(t, hashMap, relEpic, "Epic should be in the hash map")
		assert.Len(t, hashMap, 3, "Hash map should have 3 entries")
		
		// Verify file contents were updated correctly
		// 1. Story1
		content1, err := fs.ReadFile(story1Path)
		require.NoError(t, err)
		contentStr1 := string(content1)
		assert.True(t, strings.HasPrefix(contentStr1, "---"), "Story1 should start with metadata delimiter")
		assert.Contains(t, contentStr1, "file_path: "+relStory1)
		assert.Contains(t, contentStr1, "created_at:")
		assert.Contains(t, contentStr1, "last_updated:")
		assert.Contains(t, contentStr1, "_content_hash:")
		assert.Contains(t, contentStr1, "# Story 1")
		
		// 2. Story2 - should preserve original creation date
		content2, err := fs.ReadFile(story2Path)
		require.NoError(t, err)
		contentStr2 := string(content2)
		assert.True(t, strings.HasPrefix(contentStr2, "---"), "Story2 should start with metadata delimiter")
		assert.Contains(t, contentStr2, "file_path: "+relStory2)
		assert.Contains(t, contentStr2, "created_at: 2022-05-15T10:30:00Z", "Should preserve original creation date")
		assert.Contains(t, contentStr2, "last_updated:")
		assert.Contains(t, contentStr2, "_content_hash:")
		assert.NotContains(t, contentStr2, "_content_hash: oldhash123", "Old hash should be replaced")
		assert.Contains(t, contentStr2, "# Story 2")
		
		// Run again - all files should be unchanged
		updated2, unchanged2, hashMap2, err := UpdateAllUserStoryMetadata(
			docsDir,
			tempDir,
			fs,
		)
		require.NoError(t, err)
		
		assert.Len(t, updated2, 0, "No files should be updated on second run")
		assert.Len(t, unchanged2, 3, "All 3 files should be unchanged on second run")
		assert.Contains(t, unchanged2, relStory1)
		assert.Contains(t, unchanged2, relStory2)
		assert.Contains(t, unchanged2, relEpic)
		assert.Empty(t, hashMap2, "Hash map should be empty for unchanged files")
	})
} 
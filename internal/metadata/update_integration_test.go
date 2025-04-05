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
	"time"

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

// TestIntegration_UpdateFileMetadata_AddsMetadataToNewFile tests that UpdateFileMetadata
// correctly adds metadata to a new file that doesn't have any metadata yet.
func TestIntegration_UpdateFileMetadata_AddsMetadataToNewFile(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		usDir := filepath.Join(tempDir, "docs", "user-stories")
		require.NoError(t, fs.MkdirAll(usDir, 0755))
		
		// Create test file without metadata
		testPath := filepath.Join(usDir, "test.md")
		testContent := "# Test File\n\nThis is a test file."
		require.NoError(t, fs.WriteFile(testPath, []byte(testContent), 0644))
		
		// Update metadata using tempDir as root
		updated, hashMap, err := UpdateFileMetadata(testPath, tempDir, fs)
		require.NoError(t, err, "Should update metadata without error")
		
		// Verify the function returned the expected values
		assert.True(t, updated, "The file should have been updated")
		assert.NotEmpty(t, hashMap.NewHash, "A new hash should have been calculated")
		assert.Empty(t, hashMap.OldHash, "Old hash should be empty for a new file")
		assert.True(t, hashMap.Changed, "Content should be marked as changed")
		
		// Read the updated content
		updatedContent, err := fs.ReadFile(testPath)
		require.NoError(t, err, "Should be able to read the updated file")
		
		// Get the content as a string
		updatedContentStr := string(updatedContent)
		
		// Verify that metadata was added
		assert.Contains(t, updatedContentStr, "---", "Content should contain metadata delimiter")
		assert.Contains(t, updatedContentStr, "file_path:", "Content should contain file_path field")
		assert.Contains(t, updatedContentStr, "docs/user-stories/test.md", "Content should contain the right file path component")
		assert.Contains(t, updatedContentStr, "created_at:", "Content should contain created_at field")
		assert.Contains(t, updatedContentStr, "last_updated:", "Content should contain last_updated field")
		assert.Contains(t, updatedContentStr, "_content_hash:", "Content should contain _content_hash field")
		
		// Verify that the original content was preserved
		assert.Contains(t, updatedContentStr, "# Test File", "Original title should be preserved")
		assert.Contains(t, updatedContentStr, "This is a test file.", "Original content should be preserved")
		
		// Extract metadata to verify it properly
		metadata, err := ExtractMetadata(updatedContentStr)
		require.NoError(t, err, "Should be able to extract metadata")
		
		// Verify metadata fields
		assert.NotEmpty(t, metadata.FilePath, "FilePath should not be empty")
		assert.Contains(t, metadata.FilePath, "docs/user-stories/test.md", "FilePath should contain the right path component")
		assert.False(t, metadata.CreatedAt.IsZero(), "CreatedAt should not be zero")
		assert.False(t, metadata.LastUpdated.IsZero(), "LastUpdated should not be zero")
		assert.Equal(t, hashMap.NewHash, metadata.ContentHash, "ContentHash should match calculated hash")
		
		// Calculate the hash on the original content and verify it matches
		expectedHash := CalculateContentHash(testContent)
		assert.Equal(t, expectedHash, metadata.ContentHash, "ContentHash should match calculated hash")
	})
}

// TestIntegration_UpdateFileMetadata_UpdatesExistingMetadata tests that UpdateFileMetadata
// correctly updates metadata in a file that already has metadata.
func TestIntegration_UpdateFileMetadata_UpdatesExistingMetadata(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		usDir := filepath.Join(tempDir, "docs", "user-stories")
		require.NoError(t, fs.MkdirAll(usDir, 0755))
		
		// Create test file with outdated metadata
		testPath := filepath.Join(usDir, "test.md")
		
		oldHash := "oldhash123"
		oldContent := fmt.Sprintf(`---
file_path: old/path/test.md
created_at: 2022-05-15T10:30:00Z
last_updated: 2022-05-16T10:30:00Z
_content_hash: %s
---

# Test File

This is a test file with outdated metadata.`, oldHash)
		
		require.NoError(t, fs.WriteFile(testPath, []byte(oldContent), 0644))
		
		// Update metadata
		updated, hashMap, err := UpdateFileMetadata(testPath, tempDir, fs)
		require.NoError(t, err)
		
		// Verify the function returned the expected values
		assert.True(t, updated, "The file should have been updated")
		assert.NotEmpty(t, hashMap.NewHash, "A new hash should have been calculated")
		assert.Equal(t, oldHash, hashMap.OldHash, "Old hash should match the one in the file")
		assert.True(t, hashMap.Changed, "Content should be marked as changed")
		
		// Read the updated content
		updatedContent, err := fs.ReadFile(testPath)
		require.NoError(t, err)
		
		// Get the content as a string
		updatedContentStr := string(updatedContent)
		
		// Verify that metadata was updated
		assert.Contains(t, updatedContentStr, "---")
		assert.Contains(t, updatedContentStr, "file_path:", "File path field should be present")
		assert.Contains(t, updatedContentStr, "docs/user-stories/test.md", "Path should include correct components")
		assert.Contains(t, updatedContentStr, "created_at: 2022-05-15T10:30:00Z", "Creation date should be preserved")
		assert.Contains(t, updatedContentStr, "last_updated:", "Last updated should be present")
		assert.NotContains(t, updatedContentStr, "last_updated: 2022-05-16T10:30:00Z", "Last updated should be changed")
		assert.Contains(t, updatedContentStr, "_content_hash:", "Content hash should be present")
		assert.NotContains(t, updatedContentStr, "_content_hash: "+oldHash, "Old hash should be replaced")
		
		// Verify that the original content was preserved
		assert.Contains(t, updatedContentStr, "# Test File")
		assert.Contains(t, updatedContentStr, "This is a test file with outdated metadata.")
		
		// Extract metadata to verify it properly
		metadata, err := ExtractMetadata(updatedContentStr)
		require.NoError(t, err)
		
		// Verify metadata fields
		assert.Contains(t, metadata.FilePath, "docs/user-stories/test.md", "File path should contain correct components")
		assert.Equal(t, "2022-05-15T10:30:00Z", metadata.CreatedAt.Format(time.RFC3339), "CreatedAt should be preserved")
		assert.NotEqual(t, "2022-05-16T10:30:00Z", metadata.LastUpdated.Format(time.RFC3339), "LastUpdated should be updated")
		assert.Equal(t, hashMap.NewHash, metadata.ContentHash, "ContentHash should match the new hash")
	})
}

// TestIntegration_UpdateFileMetadata_PreservesMetadata tests that UpdateFileMetadata
// correctly preserves metadata when the content hash hasn't changed.
func TestIntegration_UpdateFileMetadata_PreservesMetadata(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		usDir := filepath.Join(tempDir, "docs", "user-stories")
		require.NoError(t, fs.MkdirAll(usDir, 0755))
		
		// Create test file with current metadata
		testPath := filepath.Join(usDir, "test.md")
		relPath, err := filepath.Rel(tempDir, testPath)
		require.NoError(t, err)
		
		originalContent := "# Test File\n\nThis is a test file with current metadata."
		currentHash := CalculateContentHash(originalContent)
		
		initialContent := fmt.Sprintf(`---
file_path: %s
created_at: 2022-05-15T10:30:00Z
last_updated: 2022-05-16T10:30:00Z
_content_hash: %s
---

%s`, relPath, currentHash, originalContent)
		
		require.NoError(t, fs.WriteFile(testPath, []byte(initialContent), 0644))
		
		// Update metadata
		updated, hashMap, err := UpdateFileMetadata(testPath, tempDir, fs)
		require.NoError(t, err)
		
		// Verify the function returned the expected values
		assert.False(t, updated, "The file should not have been updated")
		assert.Equal(t, currentHash, hashMap.NewHash, "New hash should match current hash")
		assert.Equal(t, currentHash, hashMap.OldHash, "Old hash should match current hash")
		assert.False(t, hashMap.Changed, "Content should not be marked as changed")
		
		// Read the file content
		currentContent, err := fs.ReadFile(testPath)
		require.NoError(t, err)
		
		// Verify content hasn't changed
		assert.Equal(t, initialContent, string(currentContent), "File content should be unchanged")
	})
} 
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
		assert.Contains(t, contentStr1, "last_updated:", "Story1 should have last updated date")
		assert.Contains(t, contentStr1, "_content_hash:", "Story1 should have content hash")
		assert.Contains(t, contentStr1, "# Story 1")
		
		// 2. Story2 - should preserve original creation date
		content2, err := fs.ReadFile(story2Path)
		require.NoError(t, err)
		contentStr2 := string(content2)
		assert.True(t, strings.HasPrefix(contentStr2, "---"), "Story2 should start with metadata delimiter")
		assert.Contains(t, contentStr2, "file_path: "+relStory2)
		assert.Contains(t, contentStr2, "created_at: 2022-05-15T10:30:00Z", "Should preserve original creation date")
		assert.Contains(t, contentStr2, "last_updated:", "Story2 should have last updated date")
		assert.Contains(t, contentStr2, "_content_hash:", "Story2 should have content hash")
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

// TestIntegration_UpdateAllUserStoryMetadata_Complex tests comprehensive scenarios for UpdateAllUserStoryMetadata
// This is an integration test that replaces the skipped mock-based test in update_test.go
func TestIntegration_UpdateAllUserStoryMetadata_Complex(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		docsDir := filepath.Join(tempDir, "docs")
		userStoriesDir := filepath.Join(docsDir, "user-stories")
		epicsDir := filepath.Join(userStoriesDir, "epics")
		ignoreDir := filepath.Join(tempDir, "docs", "ignore-me") // Will be scanned because it's not in SkippedDirectories
		nodeModulesDir := filepath.Join(tempDir, "node_modules") // Should be skipped
		
		// Create all directories
		require.NoError(t, fs.MkdirAll(userStoriesDir, 0755))
		require.NoError(t, fs.MkdirAll(epicsDir, 0755))
		require.NoError(t, fs.MkdirAll(ignoreDir, 0755))
		require.NoError(t, fs.MkdirAll(nodeModulesDir, 0755))
		
		// Create test files - some with metadata, some without
		// 1. File with no metadata, needs adding
		story1Path := filepath.Join(userStoriesDir, "story1.md")
		story1Content := "# Story 1\n\nThis is a story without metadata."
		require.NoError(t, fs.WriteFile(story1Path, []byte(story1Content), 0644))
		
		// 2. File with metadata but outdated content hash
		story2Path := filepath.Join(userStoriesDir, "story2.md")
		oldHash := "oldhash123"
		story2Content := fmt.Sprintf(`---
file_path: docs/user-stories/story2.md
created_at: 2022-05-15T10:30:00Z
last_updated: 2022-05-16T10:30:00Z
_content_hash: %s
---

# Story 2
This story has metadata but the content hash is outdated.
`, oldHash)
		require.NoError(t, fs.WriteFile(story2Path, []byte(story2Content), 0644))
		
		// 3. File with correct metadata and hash, shouldn't change
		story3Path := filepath.Join(userStoriesDir, "story3.md")
		unchangedContent := "# Story 3\nThis story won't change."
		currentHash := CalculateContentHash(unchangedContent)
		story3Content := fmt.Sprintf(`---
file_path: docs/user-stories/story3.md
created_at: 2022-05-15T10:30:00Z
last_updated: 2022-05-16T10:30:00Z
_content_hash: %s
---

%s`, currentHash, unchangedContent)
		require.NoError(t, fs.WriteFile(story3Path, []byte(story3Content), 0644))
		
		// 4. File in subdirectory without metadata
		epicPath := filepath.Join(epicsDir, "epic1.md")
		epicContent := "# Epic 1\n\nThis is an epic without metadata."
		require.NoError(t, fs.WriteFile(epicPath, []byte(epicContent), 0644))
		
		// 5. File in 'ignore-me' directory (should be processed, it's not in the SkippedDirectories list)
		ignorePath := filepath.Join(ignoreDir, "ignore-story.md")
		ignoreContent := "# Ignore Story\n\nThis should still be processed."
		require.NoError(t, fs.WriteFile(ignorePath, []byte(ignoreContent), 0644))
		
		// 6. File in node_modules directory (should be skipped)
		nodePath := filepath.Join(nodeModulesDir, "readme.md")
		nodeContent := "# Node Module\n\nThis should be skipped."
		require.NoError(t, fs.WriteFile(nodePath, []byte(nodeContent), 0644))
		
		// 7. Non-markdown file (should be skipped)
		textPath := filepath.Join(userStoriesDir, "notes.txt")
		textContent := "Just some notes, not markdown."
		require.NoError(t, fs.WriteFile(textPath, []byte(textContent), 0644))
		
		// Calculate relative paths for verification
		relStory1, _ := filepath.Rel(tempDir, story1Path)
		relStory2, _ := filepath.Rel(tempDir, story2Path)
		relStory3, _ := filepath.Rel(tempDir, story3Path)
		relEpic, _ := filepath.Rel(tempDir, epicPath)
		relIgnore, _ := filepath.Rel(tempDir, ignorePath)
		
		// Before update - read and verify content
		beforeStory1, err := fs.ReadFile(story1Path)
		require.NoError(t, err)
		assert.Equal(t, story1Content, string(beforeStory1), "Story1 content should match before update")
		
		beforeStory2, err := fs.ReadFile(story2Path)
		require.NoError(t, err)
		assert.Equal(t, story2Content, string(beforeStory2), "Story2 content should match before update")
		
		beforeStory3, err := fs.ReadFile(story3Path)
		require.NoError(t, err)
		assert.Equal(t, story3Content, string(beforeStory3), "Story3 content should match before update")
		
		// Run UpdateAllUserStoryMetadata
		updated, unchanged, hashMap, err := UpdateAllUserStoryMetadata(docsDir, tempDir, fs)
		require.NoError(t, err, "UpdateAllUserStoryMetadata should not return an error")
		
		// Verify results - specific files that should be updated
		assert.Equal(t, 4, len(updated), "Four files should be updated")
		assert.Contains(t, updated, relStory1, "Story1 should be in updated list")
		assert.Contains(t, updated, relStory2, "Story2 should be in updated list")
		assert.Contains(t, updated, relEpic, "Epic should be in updated list")
		assert.Contains(t, updated, relIgnore, "Ignore-story should be in updated list")
		
		// Verify unchanged files
		assert.Equal(t, 1, len(unchanged), "One file should remain unchanged")
		assert.Contains(t, unchanged, relStory3, "Story3 should be in unchanged list")
		
		// Verify hash map contains correct entries
		assert.Equal(t, 4, len(hashMap), "Hash map should have 4 entries")
		assert.Contains(t, hashMap, relStory1, "Story1 should be in hash map")
		assert.Contains(t, hashMap, relStory2, "Story2 should be in hash map")
		assert.Contains(t, hashMap, relEpic, "Epic should be in hash map")
		assert.Contains(t, hashMap, relIgnore, "Ignore-story should be in hash map")
		
		// Verify the hash details for Story2
		assert.Equal(t, oldHash, hashMap[relStory2].OldHash, "Story2 old hash should match")
		assert.NotEqual(t, oldHash, hashMap[relStory2].NewHash, "Story2 new hash should be different")
		assert.True(t, hashMap[relStory2].Changed, "Story2 should be marked as changed")
		
		// Verify file content has been updated correctly
		// Story1 - check metadata was added
		afterStory1, err := fs.ReadFile(story1Path)
		require.NoError(t, err)
		story1After := string(afterStory1)
		assert.True(t, strings.HasPrefix(story1After, "---"), "Story1 should start with metadata delimiter")
		assert.Contains(t, story1After, "file_path: "+relStory1, "Story1 should have correct file path")
		assert.Contains(t, story1After, "created_at:", "Story1 should have creation date")
		assert.Contains(t, story1After, "last_updated:", "Story1 should have last updated date")
		assert.Contains(t, story1After, "_content_hash:", "Story1 should have content hash")
		assert.Contains(t, story1After, "# Story 1", "Story1 should preserve original content")
		
		// Story2 - check metadata was updated (but creation date preserved)
		afterStory2, err := fs.ReadFile(story2Path)
		require.NoError(t, err)
		story2After := string(afterStory2)
		assert.True(t, strings.HasPrefix(story2After, "---"), "Story2 should start with metadata delimiter")
		assert.Contains(t, story2After, "file_path: "+relStory2, "Story2 should have correct file path")
		assert.Contains(t, story2After, "created_at: 2022-05-15T10:30:00Z", "Story2 should preserve original creation date")
		assert.Contains(t, story2After, "last_updated:", "Story2 should have last updated date")
		assert.NotContains(t, story2After, "last_updated: 2022-05-16T10:30:00Z", "Story2 should have updated last_updated date")
		assert.Contains(t, story2After, "_content_hash:", "Story2 should have content hash")
		assert.NotContains(t, story2After, "_content_hash: "+oldHash, "Story2 should have updated content hash")
		assert.Contains(t, story2After, "# Story 2", "Story2 should preserve original content")
		
		// Story3 - should remain unchanged
		afterStory3, err := fs.ReadFile(story3Path)
		require.NoError(t, err)
		assert.Equal(t, string(beforeStory3), string(afterStory3), "Story3 should remain unchanged")
		
		// Epic - check metadata was added
		afterEpic, err := fs.ReadFile(epicPath)
		require.NoError(t, err)
		epicAfter := string(afterEpic)
		assert.True(t, strings.HasPrefix(epicAfter, "---"), "Epic should start with metadata delimiter")
		assert.Contains(t, epicAfter, "file_path: "+relEpic, "Epic should have correct file path")
		assert.Contains(t, epicAfter, "created_at:", "Epic should have creation date")
		assert.Contains(t, epicAfter, "last_updated:", "Epic should have last updated date")
		assert.Contains(t, epicAfter, "_content_hash:", "Epic should have content hash")
		assert.Contains(t, epicAfter, "# Epic 1", "Epic should preserve original content")
		
		// Ignore-story - check metadata was added
		afterIgnore, err := fs.ReadFile(ignorePath)
		require.NoError(t, err)
		ignoreAfter := string(afterIgnore)
		assert.True(t, strings.HasPrefix(ignoreAfter, "---"), "Ignore-story should start with metadata delimiter")
		assert.Contains(t, ignoreAfter, "file_path: "+relIgnore, "Ignore-story should have correct file path")
		assert.Contains(t, ignoreAfter, "created_at:", "Ignore-story should have creation date")
		assert.Contains(t, ignoreAfter, "last_updated:", "Ignore-story should have last updated date")
		assert.Contains(t, ignoreAfter, "_content_hash:", "Ignore-story should have content hash")
		assert.Contains(t, ignoreAfter, "# Ignore Story", "Ignore-story should preserve original content")
		
		// Skipped files should not be modified
		afterNode, err := fs.ReadFile(nodePath)
		require.NoError(t, err)
		assert.Equal(t, nodeContent, string(afterNode), "Node module file should not be modified")
		
		afterText, err := fs.ReadFile(textPath)
		require.NoError(t, err)
		assert.Equal(t, textContent, string(afterText), "Text file should not be modified")
		
		// Run again - all files should remain unchanged this time
		updated2, unchanged2, hashMap2, err := UpdateAllUserStoryMetadata(docsDir, tempDir, fs)
		require.NoError(t, err)
		
		assert.Equal(t, 0, len(updated2), "No files should be updated on second run")
		assert.Equal(t, 5, len(unchanged2), "All 5 files should be unchanged on second run")
		assert.Contains(t, unchanged2, relStory1, "Story1 should be in unchanged list on second run")
		assert.Contains(t, unchanged2, relStory2, "Story2 should be in unchanged list on second run")
		assert.Contains(t, unchanged2, relStory3, "Story3 should be in unchanged list on second run")
		assert.Contains(t, unchanged2, relEpic, "Epic should be in unchanged list on second run")
		assert.Contains(t, unchanged2, relIgnore, "Ignore-story should be in unchanged list on second run")
		assert.Empty(t, hashMap2, "Hash map should be empty on second run")
	})
}

// TestIntegration_UpdateAllUserStoryMetadata_UpdatesAllFiles verifies that all files
// in a complex directory structure are properly updated with metadata
func TestIntegration_UpdateAllUserStoryMetadata_UpdatesAllFiles(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		docsDir := filepath.Join(tempDir, "docs")
		userStoriesDir := filepath.Join(docsDir, "user-stories")
		feature1Dir := filepath.Join(userStoriesDir, "feature1")
		feature2Dir := filepath.Join(userStoriesDir, "feature2")
		changeRequestDir := filepath.Join(docsDir, "changes-request")
		
		// Create all directories
		require.NoError(t, fs.MkdirAll(feature1Dir, 0755))
		require.NoError(t, fs.MkdirAll(feature2Dir, 0755))
		require.NoError(t, fs.MkdirAll(changeRequestDir, 0755))
		
		// Create test files in various directories
		testFiles := []struct {
			path    string
			content string
		}{
			{
				path:    filepath.Join(feature1Dir, "story1.md"),
				content: "# Story 1\n\nThis is story 1.",
			},
			{
				path:    filepath.Join(feature1Dir, "story2.md"),
				content: "# Story 2\n\nThis is story 2.",
			},
			{
				path:    filepath.Join(feature2Dir, "story3.md"),
				content: "# Story 3\n\nThis is story 3.",
			},
			{
				path:    filepath.Join(changeRequestDir, "cr1.md"),
				content: "# Change Request 1\n\nThis is a change request that should not be affected.",
			},
		}
		
		// Create all test files
		for _, file := range testFiles {
			require.NoError(t, fs.WriteFile(file.path, []byte(file.content), 0644))
		}
		
		// Calculate relative paths for verification
		var userStoryPaths []string
		var changeRequestPaths []string
		for _, file := range testFiles {
			relPath, err := filepath.Rel(tempDir, file.path)
			require.NoError(t, err)
			
			if strings.Contains(relPath, "changes-request") {
				changeRequestPaths = append(changeRequestPaths, relPath)
			} else {
				userStoryPaths = append(userStoryPaths, relPath)
			}
		}
		
		// Verify files exist and have expected content before update
		for _, file := range testFiles {
			content, err := fs.ReadFile(file.path)
			require.NoError(t, err)
			assert.Equal(t, file.content, string(content), "File should contain expected content")
		}
		
		// Update all metadata - only targeting the user-stories directory
		updated, unchanged, hashMap, err := UpdateAllUserStoryMetadata(userStoriesDir, tempDir, fs)
		require.NoError(t, err, "UpdateAllUserStoryMetadata should not return an error")
		
		// Verify results
		assert.Equal(t, len(userStoryPaths), len(updated), "All user story files should be updated")
		assert.Equal(t, 0, len(unchanged), "No files should remain unchanged in initial run")
		assert.Equal(t, len(userStoryPaths), len(hashMap), "Hash map should have entry for each user story")
		
		// Verify each user story path is in the updated list
		for _, path := range userStoryPaths {
			assert.Contains(t, updated, path, "User story should be in updated list")
			assert.Contains(t, hashMap, path, "User story should be in hash map")
		}
		
		// Verify that change request files were not updated (not in the target directory)
		for _, crPath := range changeRequestPaths {
			assert.NotContains(t, updated, crPath, "Change request should not be in updated list")
			assert.NotContains(t, hashMap, crPath, "Change request should not be in hash map")
			
			absolutePath := filepath.Join(tempDir, crPath)
			content, err := fs.ReadFile(absolutePath)
			require.NoError(t, err)
			
			for _, file := range testFiles {
				if file.path == absolutePath {
					assert.Equal(t, file.content, string(content), "Change request content should be unchanged")
					break
				}
			}
		}
		
		// Verify each user story file was updated with metadata
		for _, path := range userStoryPaths {
			absolutePath := filepath.Join(tempDir, path)
			
			// Read the updated content
			content, err := fs.ReadFile(absolutePath)
			require.NoError(t, err)
			updatedContent := string(content)
			
			// Verify metadata was added correctly
			assert.True(t, strings.HasPrefix(updatedContent, "---"), "File should start with metadata delimiter")
			assert.Contains(t, updatedContent, "file_path: "+path, "Metadata should contain correct file path")
			assert.Contains(t, updatedContent, "created_at:", "Metadata should contain creation date")
			assert.Contains(t, updatedContent, "last_updated:", "Metadata should contain last updated date")
			assert.Contains(t, updatedContent, "_content_hash:", "Metadata should contain content hash")
			
			// Verify original content was preserved
			for _, file := range testFiles {
				if filepath.Join(tempDir, path) == file.path {
					assert.Contains(t, updatedContent, file.content, "Original content should be preserved")
					break
				}
			}
			
			// Extract and verify metadata
			metadata, err := ExtractMetadata(updatedContent)
			require.NoError(t, err)
			
			assert.Equal(t, path, metadata.FilePath, "File path in metadata should match relative path")
			assert.False(t, metadata.CreatedAt.IsZero(), "Created at should not be zero")
			assert.False(t, metadata.LastUpdated.IsZero(), "Last updated should not be zero")
			assert.NotEmpty(t, metadata.ContentHash, "Content hash should not be empty")
			
			// Verify content hash matches expected value
			for _, file := range testFiles {
				if filepath.Join(tempDir, path) == file.path {
					expectedHash := CalculateContentHash(file.content)
					assert.Equal(t, expectedHash, metadata.ContentHash, "Content hash should match expected value")
					break
				}
			}
		}
		
		// Run again - all files should be unchanged
		updated2, unchanged2, hashMap2, err := UpdateAllUserStoryMetadata(userStoriesDir, tempDir, fs)
		require.NoError(t, err)
		
		assert.Equal(t, 0, len(updated2), "No files should be updated on second run")
		assert.Equal(t, len(userStoryPaths), len(unchanged2), "All user story files should be unchanged on second run")
		assert.Empty(t, hashMap2, "Hash map should be empty for unchanged files")
		
		// Verify each path is in the unchanged list
		for _, path := range userStoryPaths {
			assert.Contains(t, unchanged2, path, "User story should be in unchanged list on second run")
		}
	})
} 
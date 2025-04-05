// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user-story-matrix/usm/internal/io"
)

// TestUpdateFileMetadata_PreservesCreationDate verifies that the creation date is preserved when updating metadata
func TestUpdateFileMetadata_PreservesCreationDate(t *testing.T) {
	fs := io.NewMockFileSystem()
	
	// Create a file with existing metadata
	originalCreationDate := "2022-05-15T10:30:00Z"
	fs.AddFile("test.md", []byte(`---
file_path: test.md
created_at: ` + originalCreationDate + `
last_updated: 2022-05-16T10:30:00Z
_content_hash: oldhash
---

# Test
This is a test file that will have its content changed.
`))

	// Update the file with changed content
	updated, hashMap, err := UpdateFileMetadata("test.md", "", fs)
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.True(t, hashMap.Changed) // Content hash changed
	
	// Check that original creation date is preserved
	content, err := fs.ReadFile("test.md")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "created_at: "+originalCreationDate)
}

// TestUpdateFileMetadata_UpdatesLastUpdatedOnlyOnContentChange verifies that last_updated is only changed when content changes
func TestUpdateFileMetadata_UpdatesLastUpdatedOnlyOnContentChange(t *testing.T) {
	fs := io.NewMockFileSystem()
	
	// Create a file with existing metadata and set the current time
	lastUpdated := "2022-06-20T15:45:00Z"
	
	// Test case 1: Content hasn't changed - content hash matches
	contentWithoutMetadata := "# Unchanged Content\nThis content will not change.\n"
	expectedHash := CalculateContentHash(contentWithoutMetadata)
	
	fs.AddFile("unchanged.md", []byte(`---
file_path: unchanged.md
created_at: 2022-06-19T15:45:00Z
last_updated: ` + lastUpdated + `
_content_hash: ` + expectedHash + `
---

# Unchanged Content
This content will not change.
`))
	
	// Update the file metadata
	updated, hashMap, err := UpdateFileMetadata("unchanged.md", "", fs)
	assert.NoError(t, err)
	assert.False(t, updated) // No update needed when hash matches
	assert.False(t, hashMap.Changed) // Content hasn't changed
	
	// Check that last_updated remains the same
	content, err := fs.ReadFile("unchanged.md")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "last_updated: "+lastUpdated)
	
	// Test case 2: Content has changed - hash doesn't match
	fs.AddFile("changed.md", []byte(`---
file_path: changed.md
created_at: 2022-06-19T15:45:00Z
last_updated: ` + lastUpdated + `
_content_hash: oldhashvalue
---

# Changed Content
This content will change.
`))
	
	// Update the file
	updated, hashMap, err = UpdateFileMetadata("changed.md", "", fs)
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.True(t, hashMap.Changed)
	
	// Check that last_updated is updated to a newer time
	content, err = fs.ReadFile("changed.md")
	assert.NoError(t, err)
	assert.NotContains(t, string(content), "last_updated: "+lastUpdated)
	// The exact timestamp will be different, so just check it contains last_updated
	assert.Contains(t, string(content), "last_updated:")
}

// WriteTrackingMockFileSystem is a mock file system that tracks writes
type WriteTrackingMockFileSystem struct {
	*io.MockFileSystem
	writesCalled     int
	writtenPaths     []string
	writtenData      map[string][]byte
	writtenCallbacks []func(path string, data []byte)
}

// NewWriteTrackingMockFileSystem creates a new write-tracking mock file system
func NewWriteTrackingMockFileSystem() *WriteTrackingMockFileSystem {
	return &WriteTrackingMockFileSystem{
		MockFileSystem: io.NewMockFileSystem(),
		writtenPaths:   []string{},
		writtenData:    make(map[string][]byte),
	}
}

// AddWriteCallback adds a callback to be called when WriteFile is called
func (fs *WriteTrackingMockFileSystem) AddWriteCallback(callback func(path string, data []byte)) {
	fs.writtenCallbacks = append(fs.writtenCallbacks, callback)
}

// WriteFile overrides the mock's WriteFile to track writes
func (fs *WriteTrackingMockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	err := fs.MockFileSystem.WriteFile(path, data, perm)
	if err != nil {
		return err
	}
	
	fs.writesCalled++
	fs.writtenPaths = append(fs.writtenPaths, path)
	fs.writtenData[path] = append([]byte{}, data...) // Make a copy of the data
	
	// Call any registered callbacks
	for _, callback := range fs.writtenCallbacks {
		callback(path, data)
	}
	
	return nil
}

// GetWriteCount returns the number of times WriteFile was called
func (fs *WriteTrackingMockFileSystem) GetWriteCount() int {
	return fs.writesCalled
}

// GetWrittenPaths returns the paths that were written to
func (fs *WriteTrackingMockFileSystem) GetWrittenPaths() []string {
	return fs.writtenPaths
}

// GetWrittenData returns the data that was written to a path
func (fs *WriteTrackingMockFileSystem) GetWrittenData(path string) []byte {
	return fs.writtenData[path]
}

// TestUpdateFileMetadata_AddsMetadataToNewFile verifies that metadata is added to a file without metadata
func TestUpdateFileMetadata_AddsMetadataToNewFile(t *testing.T) {
	// This test has been implemented as an integration test using a real filesystem
	// See TestIntegration_UpdateFileMetadata_AddsMetadataToNewFile in update_integration_test.go
	t.Skip("Implemented as an integration test with real filesystem in update_integration_test.go")
}

// TestFindMarkdownFiles_FindsAllMarkdownFiles verifies that FindMarkdownFiles finds all markdown files in a directory
func TestFindMarkdownFiles_FindsAllMarkdownFiles(t *testing.T) {
	fs := io.NewMockFileSystem()
	
	// Add test directories
	fs.AddDirectory("docs")
	fs.AddDirectory("docs/user-stories")
	fs.AddDirectory("node_modules")
	fs.AddDirectory(".git")
	
	// Add markdown files
	fs.AddFile("docs/user-stories/story1.md", []byte("# Story 1"))
	fs.AddFile("docs/user-stories/story2.md", []byte("# Story 2"))
	
	// Add non-markdown file
	fs.AddFile("docs/user-stories/not-markdown.txt", []byte("Not markdown"))
	
	// Add file in directory that should be skipped
	fs.AddFile("node_modules/test.md", []byte("# Test"))
	
	// Find markdown files
	files, err := FindMarkdownFiles("docs/user-stories", fs)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))
	assert.Contains(t, files, "docs/user-stories/story1.md")
	assert.Contains(t, files, "docs/user-stories/story2.md")
	assert.NotContains(t, files, "docs/user-stories/not-markdown.txt")
	assert.NotContains(t, files, "node_modules/test.md")
}

// TestFindMarkdownFiles_SkipsIgnoredDirectories verifies that FindMarkdownFiles skips ignored directories
func TestFindMarkdownFiles_SkipsIgnoredDirectories(t *testing.T) {
	fs := io.NewMockFileSystem()
	
	// Create test directories
	fs.AddDirectory("docs")
	fs.AddDirectory("docs/node_modules")
	fs.AddDirectory("docs/.git")
	fs.AddDirectory("docs/dist")
	fs.AddDirectory("docs/build")
	
	// Add markdown files
	fs.AddFile("docs/file.md", []byte("# File"))
	fs.AddFile("docs/node_modules/node.md", []byte("# Node"))
	fs.AddFile("docs/.git/git.md", []byte("# Git"))
	fs.AddFile("docs/dist/dist.md", []byte("# Dist"))
	fs.AddFile("docs/build/build.md", []byte("# Build"))
	
	// Find markdown files
	files, err := FindMarkdownFiles("docs", fs)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Contains(t, files, "docs/file.md")
	assert.NotContains(t, files, "docs/node_modules/node.md")
	assert.NotContains(t, files, "docs/.git/git.md")
	assert.NotContains(t, files, "docs/dist/dist.md")
	assert.NotContains(t, files, "docs/build/build.md")
}

// TestShouldSkipDirectory tests that the function correctly identifies directories to skip
func TestShouldSkipDirectory(t *testing.T) {
	// Test directories that should be skipped
	for _, dir := range SkippedDirectories {
		assert.True(t, ShouldSkipDirectory(dir), fmt.Sprintf("%s should be skipped", dir))
	}
	
	// Test directories that should not be skipped
	for _, dir := range []string{
		"docs",
		"user-stories",
		"src",
		"content",
		"images",
		"non-standard-name",
	} {
		assert.False(t, ShouldSkipDirectory(dir), fmt.Sprintf("%s should not be skipped", dir))
	}
	
	// Test case sensitivity (directory names should match exactly)
	if len(SkippedDirectories) > 0 {
		// Convert first skipped directory to uppercase
		upperDir := strings.ToUpper(SkippedDirectories[0])
		if upperDir != SkippedDirectories[0] { // Only test if case is different
			assert.False(t, ShouldSkipDirectory(upperDir), 
				fmt.Sprintf("%s should not be skipped (case-sensitive match)", upperDir))
		}
	}
}

// TestUpdateAllUserStoryMetadata tests the basic functionality of updating multiple markdown files
func TestUpdateAllUserStoryMetadata(t *testing.T) {
	// This test has been implemented as a comprehensive integration test 
	// See TestIntegration_UpdateAllUserStoryMetadata_Complex in update_integration_test.go
	t.Skip("Implemented as an integration test with real filesystem in update_integration_test.go")
}

// TestUpdateAllUserStoryMetadata_UpdatesAllFiles verifies that all files in a directory are updated
func TestUpdateAllUserStoryMetadata_UpdatesAllFiles(t *testing.T) {
	// This test has been implemented as an integration test using a real filesystem
	// See TestIntegration_UpdateAllUserStoryMetadata_UpdatesAllFiles in update_integration_test.go
	t.Skip("Implemented as an integration test with real filesystem in update_integration_test.go")
}

func TestUpdateFileMetadata_PreservesOriginalCreationDate(t *testing.T) {
	// Create mock filesystem
	fs := io.NewMockFileSystem()
	
	// Create time values
	originalTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// Create a file with existing metadata
	existingMetadata := fmt.Sprintf("---\nfile_path: docs/user-stories/test.md\ncreated_at: %s\nlast_updated: %s\n_content_hash: original-hash\n---\n\n",
		originalTime.Format(time.RFC3339),
		originalTime.Format(time.RFC3339))
	
	content := existingMetadata + "# Test File\n\nThis is a test file."
	fs.AddFile("docs/user-stories/test.md", []byte(content))
	
	// Update metadata
	updated, hashMap, err := UpdateFileMetadata("docs/user-stories/test.md", ".", fs)
	require.NoError(t, err)
	
	// Verify the function returned the expected values
	assert.True(t, updated, "The file should have been updated")
	assert.NotEqual(t, "original-hash", hashMap.NewHash, "A new hash should have been calculated")
	assert.Equal(t, "original-hash", hashMap.OldHash, "Old hash should match the original")
	assert.True(t, hashMap.Changed, "Content should be marked as changed")
	
	// Get the last write operation
	writeOp, exists := fs.GetLastWrite("docs/user-stories/test.md")
	require.True(t, exists, "Expected a write operation to occur")
	
	// Extract metadata from updated content
	updatedContent := string(writeOp.Content)
	updatedMetadata, err := ExtractMetadata(updatedContent)
	require.NoError(t, err)
	
	// Verify that creation date is preserved
	assert.Equal(t, originalTime.Format(time.RFC3339), updatedMetadata.CreatedAt.Format(time.RFC3339), 
		"Creation date should be preserved")
}

func TestUpdateFileMetadata_UpdatesLastUpdatedForChangedContent(t *testing.T) {
	// Create mock filesystem
	fs := io.NewMockFileSystem()
	
	// Create time values
	originalTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// Create a file with existing metadata
	existingMetadata := fmt.Sprintf("---\nfile_path: docs/user-stories/test.md\ncreated_at: %s\nlast_updated: %s\n_content_hash: original-hash\n---\n\n",
		originalTime.Format(time.RFC3339),
		originalTime.Format(time.RFC3339))
	
	content := existingMetadata + "# Test File\n\nThis is a test file with updated content."
	fs.AddFile("docs/user-stories/test.md", []byte(content))
	
	// Update metadata
	updated, hashMap, err := UpdateFileMetadata("docs/user-stories/test.md", ".", fs)
	require.NoError(t, err)
	
	// Verify the function returned the expected values
	assert.True(t, updated, "The file should have been updated")
	assert.NotEqual(t, "original-hash", hashMap.NewHash, "A new hash should have been calculated")
	assert.Equal(t, "original-hash", hashMap.OldHash, "Old hash should match the original")
	assert.True(t, hashMap.Changed, "Content should be marked as changed")
	
	// Get the last write operation
	writeOp, exists := fs.GetLastWrite("docs/user-stories/test.md")
	require.True(t, exists, "Expected a write operation to occur")
	
	// Extract metadata from updated content
	updatedContent := string(writeOp.Content)
	updatedMetadata, err := ExtractMetadata(updatedContent)
	require.NoError(t, err)
	
	// Verify that last updated is changed
	assert.NotEqual(t, originalTime.Format(time.RFC3339), updatedMetadata.LastUpdated.Format(time.RFC3339), 
		"Last updated date should be changed for content changes")
}

func TestUpdateFileMetadata_SkipsUpdateForUnchangedContent(t *testing.T) {
	// Create mock filesystem
	fs := io.NewMockFileSystem()
	
	// Create test content and calculate its hash
	testContent := "# Test File\n\nThis is test content."
	contentHash := CalculateContentHash(testContent)
	
	// Create existing metadata with the correct hash
	existingMetadata := fmt.Sprintf("---\nfile_path: docs/user-stories/test.md\ncreated_at: %s\nlast_updated: %s\n_content_hash: %s\n---\n\n",
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
		contentHash)
	
	// Create full file content
	fullContent := existingMetadata + testContent
	fs.AddFile("docs/user-stories/test.md", []byte(fullContent))
	
	// Record initial write operations count
	initialWriteOps := len(fs.WriteOps)
	
	// Update metadata
	updated, hashMap, err := UpdateFileMetadata("docs/user-stories/test.md", ".", fs)
	require.NoError(t, err)
	
	// Verify the function returned the expected values
	assert.False(t, updated, "The file should not have been updated")
	assert.Equal(t, contentHash, hashMap.NewHash, "New hash should match the original")
	assert.Equal(t, contentHash, hashMap.OldHash, "Old hash should match the original")
	assert.False(t, hashMap.Changed, "Content should not be marked as changed")
	
	// Check if any new write operations occurred
	assert.Equal(t, initialWriteOps, len(fs.WriteOps), 
		"No write operations should happen for unchanged content")
} 
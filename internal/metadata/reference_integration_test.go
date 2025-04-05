// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user-story-matrix/usm/internal/io"
)

// TestIntegration_UpdateChangeRequestReferences tests the UpdateChangeRequestReferences function
// using a real filesystem instead of a mock to avoid the issues that caused the original test to be skipped.
func TestIntegration_UpdateChangeRequestReferences(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}
	
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-update-change-request-references")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test file
	changeRequestContent := `
---
title: Test Change Request
created_at: 2025-03-17T12:00:00Z
---

This is a test change request that references a user story.

## User Stories
- title: Test User Story
  file: docs/user-stories/test/test-story.md
  content-hash: oldhash123
- title: Another Test User Story
  file: docs/user-stories/test/another-story.md
  content-hash: oldhash456
`
	changeRequestFile := filepath.Join(tempDir, "test-change-request.md")
	err = os.WriteFile(changeRequestFile, []byte(changeRequestContent), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Set up hash map
	hashMap := ContentChangeMap{
		"docs/user-stories/test/test-story.md": {
			FilePath: "docs/user-stories/test/test-story.md",
			OldHash:  "oldhash123",
			NewHash:  "newhash789",
			Changed:  true,
		},
	}
	
	// Initialize filesystem
	fs := io.NewOSFileSystem()
	
	// Update references
	updated, refsUpdated, mismatches, err := UpdateChangeRequestReferences(changeRequestFile, hashMap, fs)
	
	// Check results
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, 1, refsUpdated)
	assert.Equal(t, 0, len(mismatches))
	
	// Read back the file content
	updatedContent, err := os.ReadFile(changeRequestFile)
	assert.NoError(t, err)
	
	// Verify content was updated
	assert.Contains(t, string(updatedContent), "content-hash: newhash789")
	assert.Contains(t, string(updatedContent), "content-hash: oldhash456") // Should not change
}

func TestIntegration_UpdateChangeRequestReferences_MismatchedHash(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}
	
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-update-change-request-mismatched")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test file
	changeRequestContent := `
---
title: Test Change Request
created_at: 2025-03-17T12:00:00Z
---

This is a test change request that references a user story with a mismatched hash.

## User Stories
- title: Test User Story With Mismatched Hash
  file: docs/user-stories/test/test-story.md
  content-hash: differenthash123
`
	changeRequestFile := filepath.Join(tempDir, "test-mismatched-change-request.md")
	err = os.WriteFile(changeRequestFile, []byte(changeRequestContent), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Set up hash map
	hashMap := ContentChangeMap{
		"docs/user-stories/test/test-story.md": {
			FilePath: "docs/user-stories/test/test-story.md",
			OldHash:  "oldhash123", // Different from what's in the file
			NewHash:  "newhash789",
			Changed:  true,
		},
	}
	
	// Initialize filesystem
	fs := io.NewOSFileSystem()
	
	// Update references
	updated, refsUpdated, mismatches, err := UpdateChangeRequestReferences(changeRequestFile, hashMap, fs)
	
	// Check results
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, 1, refsUpdated)
	assert.Equal(t, 1, len(mismatches))
	
	// Verify the mismatched reference
	if len(mismatches) > 0 {
		assert.Equal(t, "docs/user-stories/test/test-story.md", mismatches[0].FilePath)
		assert.Equal(t, "differenthash123", mismatches[0].ReferenceHash)
		assert.Equal(t, "oldhash123", mismatches[0].OldHash)
	}
	
	// Read back the file content
	updatedContent, err := os.ReadFile(changeRequestFile)
	assert.NoError(t, err)
	
	// Verify content was updated despite the mismatch
	assert.Contains(t, string(updatedContent), "content-hash: newhash789")
}

// TestIntegration_UpdateAllChangeRequestReferences tests the UpdateAllChangeRequestReferences
// function using a real filesystem.
func TestIntegration_UpdateAllChangeRequestReferences(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}
	
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-update-all-change-requests")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create directory structure
	changeRequestDir := filepath.Join(tempDir, "docs", "changes-request")
	err = os.MkdirAll(changeRequestDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create test files
	changeRequest1 := `
---
title: Change Request 1
created_at: 2025-03-17T12:00:00Z
---

## User Stories
- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: old-hash-1
`
	
	changeRequest2 := `
---
title: Change Request 2
created_at: 2025-03-17T12:00:00Z
---

## User Stories
- title: Story 2
  file: docs/user-stories/story2.md
  content-hash: old-hash-2
`
	
	// Write files
	err = os.WriteFile(filepath.Join(changeRequestDir, "change-request-1.md"), []byte(changeRequest1), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	err = os.WriteFile(filepath.Join(changeRequestDir, "change-request-2.md"), []byte(changeRequest2), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Set up hash map
	hashMap := ContentChangeMap{
		"docs/user-stories/story1.md": {
			FilePath: "docs/user-stories/story1.md",
			OldHash:  "old-hash-1",
			NewHash:  "new-hash-1",
			Changed:  true,
		},
		// Story 2 unchanged
	}
	
	// Initialize filesystem
	fs := io.NewOSFileSystem()
	
	// Update all references
	updatedFiles, unchangedFiles, refsUpdated, mismatches, err := UpdateAllChangeRequestReferences(tempDir, hashMap, fs)
	
	// Check results
	assert.NoError(t, err)
	assert.Equal(t, 1, len(updatedFiles))
	assert.Equal(t, 1, len(unchangedFiles))
	assert.Equal(t, 1, refsUpdated)
	assert.Equal(t, 0, len(mismatches))
}

func TestIntegration_UpdateAllChangeRequestReferences_NoChanges(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}
	
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-update-all-change-requests-no-changes")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create directory structure
	changeRequestDir := filepath.Join(tempDir, "docs", "changes-request")
	err = os.MkdirAll(changeRequestDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create test file
	changeRequest := `
---
title: Change Request
created_at: 2025-03-17T12:00:00Z
---

## User Stories
- title: Story
  file: docs/user-stories/story.md
  content-hash: old-hash
`
	
	// Write file
	err = os.WriteFile(filepath.Join(changeRequestDir, "change-request.md"), []byte(changeRequest), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Empty hash map (no changes)
	hashMap := ContentChangeMap{}
	
	// Initialize filesystem
	fs := io.NewOSFileSystem()
	
	// Update all references
	updatedFiles, unchangedFiles, refsUpdated, mismatches, err := UpdateAllChangeRequestReferences(tempDir, hashMap, fs)
	
	// Check results
	assert.NoError(t, err)
	assert.Nil(t, updatedFiles)
	assert.Nil(t, unchangedFiles)
	assert.Equal(t, 0, refsUpdated)
	assert.Nil(t, mismatches)
}

func TestIntegration_UpdateAllChangeRequestReferences_WithMismatchedHashes(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}
	
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-update-all-change-requests-mismatched")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create directory structure
	changeRequestDir := filepath.Join(tempDir, "docs", "changes-request")
	err = os.MkdirAll(changeRequestDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create test files
	changeRequest1 := `
---
title: Change Request 1
created_at: 2025-03-17T12:00:00Z
---

## User Stories
- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: different-hash-1
`
	
	changeRequest2 := `
---
title: Change Request 2
created_at: 2025-03-17T12:00:00Z
---

## User Stories
- title: Story 2
  file: docs/user-stories/story2.md
  content-hash: old-hash-2
`
	
	// Write files
	err = os.WriteFile(filepath.Join(changeRequestDir, "change-request-1.md"), []byte(changeRequest1), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	err = os.WriteFile(filepath.Join(changeRequestDir, "change-request-2.md"), []byte(changeRequest2), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Set up hash map
	hashMap := ContentChangeMap{
		"docs/user-stories/story1.md": {
			FilePath: "docs/user-stories/story1.md",
			OldHash:  "old-hash-1", // Different from what's in the file
			NewHash:  "new-hash-1",
			Changed:  true,
		},
		"docs/user-stories/story2.md": {
			FilePath: "docs/user-stories/story2.md",
			OldHash:  "old-hash-2",
			NewHash:  "new-hash-2",
			Changed:  true,
		},
	}
	
	// Initialize filesystem
	fs := io.NewOSFileSystem()
	
	// Update all references
	updatedFiles, unchangedFiles, refsUpdated, mismatches, err := UpdateAllChangeRequestReferences(tempDir, hashMap, fs)
	
	// Check results
	assert.NoError(t, err)
	assert.Equal(t, 2, len(updatedFiles))
	assert.Equal(t, 0, len(unchangedFiles))
	assert.Equal(t, 2, refsUpdated)
	assert.Equal(t, 1, len(mismatches))
	
	// Verify the mismatched reference
	if len(mismatches) > 0 {
		assert.Equal(t, "docs/user-stories/story1.md", mismatches[0].FilePath)
		assert.Equal(t, "different-hash-1", mismatches[0].ReferenceHash)
		assert.Equal(t, "old-hash-1", mismatches[0].OldHash)
	}
}

// Helper function to create a temporary directory for testing
func setupTempDir(t *testing.T) (string, func()) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-metadata-references")
	if err != nil {
		t.Fatal(err)
	}
	
	// Create required subdirectories
	changeRequestDir := filepath.Join(tempDir, "docs", "changes-request")
	err = os.MkdirAll(changeRequestDir, 0755)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatal(err)
	}
	
	// Return directory and cleanup function
	return tempDir, func() {
		os.RemoveAll(tempDir)
	}
}

func TestIntegration_UpdateChangeRequestReferences_PreventFilePathCorruption(t *testing.T) {
	// Skip if we're not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	// Setup temporary directory
	tempDir, cleanup := setupTempDir(t)
	defer cleanup()
	
	// Use real file system
	fs := io.NewOSFileSystem()
	
	// Create a change request file with multiple user story references
	// Similar to the format in the corrupted files we observed
	changeRequestContent := `---
name: full tui
created-at: 2025-03-24T20:03:55+01:00
user-stories:
  - title: Initial View of Change Request Selection UI
    file: docs/user-stories/create-change-request-tui/01-initial-view-of-change-request-selection-ui.md
    content-hash: oldhash111
  - title: Live Search Filtering
    file: docs/user-stories/create-change-request-tui/02-live-search-filtering.md
    content-hash: oldhash222
  - title: Entering Search Mode Separates Typing from Selection
    file: docs/user-stories/create-change-request-tui/03-entering-search-mode-separates-typing-from-selection.md
    content-hash: oldhash333
  - title: Keyboard Navigation and Selection
    file: docs/user-stories/create-change-request-tui/06-keyboard-navigation-and-selection.md
    content-hash: oldhash444
---
`
	changeRequestFile := filepath.Join(tempDir, "2025-03-24-200355-full-tui.blueprint.md")
	err := fs.WriteFile(changeRequestFile, []byte(changeRequestContent), 0644)
	require.NoError(t, err)
	
	// Create a hash map with changes for multiple files, using long hashes
	hashMap := ContentChangeMap{
		"docs/user-stories/create-change-request-tui/01-initial-view-of-change-request-selection-ui.md": {
			FilePath: "docs/user-stories/create-change-request-tui/01-initial-view-of-change-request-selection-ui.md",
			OldHash:  "oldhash111",
			NewHash:  "e7896fb05c2c6c218b772146cd753f125d3e666f8bd0288a545f0d5d0ed42ed2",
			Changed:  true,
		},
		"docs/user-stories/create-change-request-tui/02-live-search-filtering.md": {
			FilePath: "docs/user-stories/create-change-request-tui/02-live-search-filtering.md",
			OldHash:  "oldhash222",
			NewHash:  "448981a2d2918b6bb7bfbc6015ef86e9dff5e1c0a944aa53d652ae3371ce40f2",
			Changed:  true,
		},
		"docs/user-stories/create-change-request-tui/03-entering-search-mode-separates-typing-from-selection.md": {
			FilePath: "docs/user-stories/create-change-request-tui/03-entering-search-mode-separates-typing-from-selection.md",
			OldHash:  "oldhash333",
			NewHash:  "1f9a34087be1e027edf4ef0b979b3a846a9c17fb722f176bbb6439561279c663",
			Changed:  true,
		},
		"docs/user-stories/create-change-request-tui/06-keyboard-navigation-and-selection.md": {
			FilePath: "docs/user-stories/create-change-request-tui/06-keyboard-navigation-and-selection.md",
			OldHash:  "oldhash444",
			NewHash:  "feeb2080784b92262b59d45aed619d0b7980b7d3905532d52b779a88de31203d",
			Changed:  true,
		},
	}
	
	// Update the references
	updated, refsUpdated, mismatches, err := UpdateChangeRequestReferences(changeRequestFile, hashMap, fs)
	require.NoError(t, err)
	require.True(t, updated)
	require.Equal(t, 4, refsUpdated)
	require.Empty(t, mismatches)
	
	// Read the file back and check for correct updates
	content, err := fs.ReadFile(changeRequestFile)
	require.NoError(t, err)
	
	updatedContent := string(content)
	
	// Check that file paths are intact
	assert.Contains(t, updatedContent, "file: docs/user-stories/create-change-request-tui/01-initial-view-of-change-request-selection-ui.md")
	assert.Contains(t, updatedContent, "file: docs/user-stories/create-change-request-tui/02-live-search-filtering.md")
	assert.Contains(t, updatedContent, "file: docs/user-stories/create-change-request-tui/03-entering-search-mode-separates-typing-from-selection.md")
	assert.Contains(t, updatedContent, "file: docs/user-stories/create-change-request-tui/06-keyboard-navigation-and-selection.md")
	
	// Check that content hashes are updated
	assert.Contains(t, updatedContent, "content-hash: e7896fb05c2c6c218b772146cd753f125d3e666f8bd0288a545f0d5d0ed42ed2")
	assert.Contains(t, updatedContent, "content-hash: 448981a2d2918b6bb7bfbc6015ef86e9dff5e1c0a944aa53d652ae3371ce40f2")
	assert.Contains(t, updatedContent, "content-hash: 1f9a34087be1e027edf4ef0b979b3a846a9c17fb722f176bbb6439561279c663")
	assert.Contains(t, updatedContent, "content-hash: feeb2080784b92262b59d45aed619d0b7980b7d3905532d52b779a88de31203d")
	
	// Check for corruption patterns - these should NOT be present
	corruptionPatterns := []string{
		"live-search448981a2d2918b6bb7bfbc6015ef86e9dff5e1c0a944aa53d652ae3371ce40f2",
		"03-entering-s1f9a34087be1e027edf4ef0b979b3a846a9c17fb722f176bbb6439561279c663",
		"06-kfeeb2080784b92262b59d45aed619d0b7980b7d3905532d52b779a88de31203dvigation",
	}
	
	for _, pattern := range corruptionPatterns {
		assert.NotContains(t, updatedContent, pattern, "Found corruption pattern: %s", pattern)
	}
} 
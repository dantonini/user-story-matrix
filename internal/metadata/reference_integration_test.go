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
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestFormatMetadata verifies that metadata is formatted correctly
func TestFormatMetadata(t *testing.T) {
	// Create a test metadata object
	createdAt, _ := time.Parse(time.RFC3339, "2022-05-15T10:30:00Z")
	lastUpdated, _ := time.Parse(time.RFC3339, "2022-05-16T10:30:00Z")
	filePath := "docs/user-stories/test.md"
	contentHash := "testhash123"
	
	metadata := Metadata{
		FilePath:    filePath,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdated,
		ContentHash: "oldhash", // This should be ignored, using the new contentHash param
		RawMetadata: map[string]string{
			"file_path":    filePath,
			"created_at":   createdAt.Format(time.RFC3339),
			"last_updated": lastUpdated.Format(time.RFC3339),
			"_content_hash": "oldhash",
		},
	}
	
	// Format the metadata
	formatted := FormatMetadata(metadata, contentHash)
	
	// Verify the result
	assert.Contains(t, formatted, "---")
	assert.Contains(t, formatted, "file_path: "+filePath)
	assert.Contains(t, formatted, "created_at: 2022-05-15T10:30:00Z")
	assert.Contains(t, formatted, "last_updated: 2022-05-16T10:30:00Z")
	assert.Contains(t, formatted, "_content_hash: "+contentHash) // Should use new hash
	assert.Contains(t, formatted, "---\n\n") // Should end with separator and newlines
}

// MockFileInfo implements os.FileInfo for testing
type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m MockFileInfo) Name() string       { return m.name }
func (m MockFileInfo) Size() int64        { return m.size }
func (m MockFileInfo) Mode() os.FileMode  { return m.mode }
func (m MockFileInfo) ModTime() time.Time { return m.modTime }
func (m MockFileInfo) IsDir() bool        { return m.isDir }
func (m MockFileInfo) Sys() interface{}   { return nil }

// TestGenerateMetadata_NewFile verifies metadata generation for a new file
func TestGenerateMetadata_NewFile(t *testing.T) {
	// Setup test data
	filePath := "docs/user-stories/test.md"
	contentHash := "newhash123"
	modTime := time.Now().Add(-24 * time.Hour) // Yesterday
	
	// Create a mock file info
	fileInfo := MockFileInfo{
		name:    "test.md",
		size:    100,
		mode:    0644,
		modTime: modTime,
		isDir:   false,
	}
	
	// Empty metadata for a new file
	emptyMetadata := Metadata{}
	
	// Generate metadata
	result := GenerateMetadata(filePath, ".", fileInfo, emptyMetadata, contentHash)
	
	// Verify the result
	assert.Contains(t, result, "file_path: "+filePath)
	assert.Contains(t, result, "created_at: "+modTime.Format(time.RFC3339)) // Should use modTime for new file
	assert.Contains(t, result, "last_updated:") // Should be current time
	assert.Contains(t, result, "_content_hash: "+contentHash)
}

// TestGenerateMetadata_ExistingFile verifies metadata generation preserves values for existing files
func TestGenerateMetadata_ExistingFile(t *testing.T) {
	// Setup test data
	filePath := "docs/user-stories/test.md"
	oldHash := "oldhash123"
	newHash := "newhash456" // Different hash to trigger last_updated change
	
	createdAt, _ := time.Parse(time.RFC3339, "2022-05-15T10:30:00Z")
	lastUpdated, _ := time.Parse(time.RFC3339, "2022-05-16T10:30:00Z")
	
	// Create a mock file info with more recent mod time
	fileInfo := MockFileInfo{
		name:    "test.md",
		size:    100,
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}
	
	// Existing metadata
	existingMetadata := Metadata{
		FilePath:    filePath,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdated,
		ContentHash: oldHash,
		RawMetadata: map[string]string{
			"file_path":    filePath,
			"created_at":   createdAt.Format(time.RFC3339),
			"last_updated": lastUpdated.Format(time.RFC3339),
			"_content_hash": oldHash,
		},
	}
	
	// Generate metadata with changed content
	result := GenerateMetadata(filePath, ".", fileInfo, existingMetadata, newHash)
	
	// Verify the result
	assert.Contains(t, result, "file_path: "+filePath)
	assert.Contains(t, result, "created_at: "+createdAt.Format(time.RFC3339)) // Should preserve original creation date
	assert.NotContains(t, result, "last_updated: "+lastUpdated.Format(time.RFC3339)) // Should update since hash changed
	assert.Contains(t, result, "_content_hash: "+newHash)
}

// TestGenerateMetadata_UnchangedContent verifies last_updated is preserved when content hasn't changed
func TestGenerateMetadata_UnchangedContent(t *testing.T) {
	// Setup test data
	filePath := "docs/user-stories/test.md"
	sameHash := "samehash123" // Same hash to indicate content hasn't changed
	
	createdAt, _ := time.Parse(time.RFC3339, "2022-05-15T10:30:00Z")
	lastUpdated, _ := time.Parse(time.RFC3339, "2022-05-16T10:30:00Z")
	
	// Create a mock file info with more recent mod time
	fileInfo := MockFileInfo{
		name:    "test.md",
		size:    100,
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}
	
	// Existing metadata with same content hash
	existingMetadata := Metadata{
		FilePath:    filePath,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdated,
		ContentHash: sameHash,
		RawMetadata: map[string]string{
			"file_path":    filePath,
			"created_at":   createdAt.Format(time.RFC3339),
			"last_updated": lastUpdated.Format(time.RFC3339),
			"_content_hash": sameHash,
		},
	}
	
	// Generate metadata with unchanged content
	result := GenerateMetadata(filePath, ".", fileInfo, existingMetadata, sameHash)
	
	// Verify the result
	assert.Contains(t, result, "file_path: "+filePath)
	assert.Contains(t, result, "created_at: "+createdAt.Format(time.RFC3339)) // Should preserve original creation date
	assert.Contains(t, result, "last_updated: "+lastUpdated.Format(time.RFC3339)) // Should preserve last_updated
	assert.Contains(t, result, "_content_hash: "+sameHash)
} 
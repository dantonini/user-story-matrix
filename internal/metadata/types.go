// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"time"
)

// Metadata represents the metadata section in a file
type Metadata struct {
	FilePath     string    `yaml:"file_path"`
	CreatedAt    time.Time `yaml:"created_at"`
	LastUpdated  time.Time `yaml:"last_updated"`
	ContentHash  string    `yaml:"_content_hash"`
	RawMetadata  map[string]string
}

// ContentHashMap represents the changes in a file's content hash
type ContentHashMap struct {
	FilePath  string
	OldHash   string
	NewHash   string
	Changed   bool // Whether the actual content changed (not just metadata)
}

// ContentChangeMap maps file paths to their ContentHashMap
type ContentChangeMap map[string]ContentHashMap

// MetadataOptions provides configuration options for metadata operations
type MetadataOptions struct {
	SkipReferences bool // Whether to skip updating references in change requests
	Debug          bool // Whether to enable debug logging
}

// NewDefaultMetadataOptions creates a new MetadataOptions with default values
func NewDefaultMetadataOptions() MetadataOptions {
	return MetadataOptions{
		SkipReferences: false,
		Debug:          false,
	}
} 
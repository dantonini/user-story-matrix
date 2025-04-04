// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"regexp"
	"strings"
	"time"
)

var (
	// Regex pattern to match metadata section
	metadataRegex = regexp.MustCompile(`(?m)^---\s*\n([\s\S]*?)\n---\s*\n`)

	// Regex pattern to match specific metadata key-value pairs
	metadataKeyValueRegex = regexp.MustCompile(`(?m)^([^:]+):\s*(.*)$`)
)

// ExtractMetadata extracts metadata from file content
func ExtractMetadata(content string) (Metadata, error) {
	metadata := Metadata{
		RawMetadata: make(map[string]string),
	}

	// Extract raw metadata key-value pairs
	rawMetadata := extractRawMetadata(content)
	metadata.RawMetadata = rawMetadata

	// Parse specific fields
	if filePath, ok := rawMetadata["file_path"]; ok {
		metadata.FilePath = filePath
	}

	if contentHash, ok := rawMetadata["_content_hash"]; ok {
		metadata.ContentHash = contentHash
	}

	// Parse timestamps
	if createdAt, ok := rawMetadata["created_at"]; ok {
		t, err := time.Parse(time.RFC3339, createdAt)
		if err == nil {
			metadata.CreatedAt = t
		}
	}

	if lastUpdated, ok := rawMetadata["last_updated"]; ok {
		t, err := time.Parse(time.RFC3339, lastUpdated)
		if err == nil {
			metadata.LastUpdated = t
		}
	}

	return metadata, nil
}

// extractRawMetadata extracts the raw metadata key-value pairs from content
func extractRawMetadata(content string) map[string]string {
	rawMetadata := make(map[string]string)

	matches := metadataRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return rawMetadata
	}

	metadataText := matches[1]
	kvMatches := metadataKeyValueRegex.FindAllStringSubmatch(metadataText, -1)

	for _, kv := range kvMatches {
		if len(kv) >= 3 {
			key := strings.TrimSpace(kv[1])
			value := strings.TrimSpace(kv[2])
			if key != "" && value != "" {
				rawMetadata[key] = value
			}
		}
	}

	return rawMetadata
}

// GetContentWithoutMetadata removes metadata section from content
func GetContentWithoutMetadata(content string) string {
	return metadataRegex.ReplaceAllString(content, "")
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package utils

import (
	"strings"
)

// FilterContentHash removes content hash lines from markdown content
// If includeContentHash is true, the content hash will be preserved
func FilterContentHash(content string, includeContentHash bool) string {
	if includeContentHash {
		return content
	}
	
	// Split the content into lines
	lines := strings.Split(content, "\n")
	// Pre-allocate the filtered lines slice to avoid the prealloc lint issue
	filteredLines := make([]string, 0, len(lines))
	inMetadata := false
	isFirstMetadataSection := true // Tracks if we're in the first metadata section
	metadataSectionCount := 0      // Counts how many metadata sections we've seen
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			// Toggle metadata section flag
			if !inMetadata {
				// Starting a metadata section
				inMetadata = true
				metadataSectionCount++
				isFirstMetadataSection = (metadataSectionCount == 1)
			} else {
				// Ending a metadata section
				inMetadata = false
			}
			
			filteredLines = append(filteredLines, line)
			continue
		}
		
		// Skip content hash lines, but only in the first metadata section
		if inMetadata && isFirstMetadataSection && strings.Contains(line, "_content_hash:") {
			continue
		}
		
		filteredLines = append(filteredLines, line)
	}
	
	return strings.Join(filteredLines, "\n")
} 
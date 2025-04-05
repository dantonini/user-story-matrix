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
	
	// TODO: Implement a more sophisticated filtering algorithm
	// This is a placeholder implementation using a simple string split and join
	// In the actual implementation, we will:
	// 1. Split the content into lines
	// 2. Track whether we are in the metadata section
	// 3. Skip content hash lines when in metadata
	// 4. Join everything back together
	
	// For now, we'll use a basic implementation
	lines := strings.Split(content, "\n")
	// Pre-allocate the filtered lines slice to avoid the prealloc lint issue
	filteredLines := make([]string, 0, len(lines))
	inMetadata := false
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			// Toggle metadata section flag
			inMetadata = !inMetadata
			filteredLines = append(filteredLines, line)
			continue
		}
		
		// Skip content hash lines in metadata section
		if inMetadata && strings.Contains(line, "_content_hash:") {
			continue
		}
		
		filteredLines = append(filteredLines, line)
	}
	
	return strings.Join(filteredLines, "\n")
} 
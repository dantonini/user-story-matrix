// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package utils

import (
	"strings"
)

// FilterContentHash removes content hash lines from markdown content.
//
// This function processes YAML frontmatter in markdown files to selectively
// remove content hash metadata. Key behaviors:
//
// - If includeContentHash is true, returns the original content unchanged
// - Only removes the _content_hash line from the first metadata section (between --- markers)
// - Preserves any content hash entries in subsequent metadata sections
// - Maintains exact formatting and whitespace of the document
//
// This is particularly useful for generating clean output when displaying
// user stories without exposing implementation details like content hashes.
func FilterContentHash(content string, includeContentHash bool) string {
	if includeContentHash {
		return content
	}
	
	lines := strings.Split(content, "\n")
	filteredLines := make([]string, 0, len(lines))
	
	inMetadata := false
	isFirstMetadataSection := false // Will be set to true when first section starts
	
	for _, line := range lines {
		lineContent := strings.TrimSpace(line)
		
		// Handle metadata section markers
		if lineContent == "---" {
			if !inMetadata {
				// Starting a metadata section
				inMetadata = true
				// First --- marker we encounter starts the first metadata section
				if !isFirstMetadataSection {
					isFirstMetadataSection = true
				}
			} else {
				// Ending a metadata section
				inMetadata = false
				// Stop filtering after first section ends
				isFirstMetadataSection = false
			}
			
			filteredLines = append(filteredLines, line)
			continue
		}
		
		// Only filter content hash in the first metadata section
		if inMetadata && isFirstMetadataSection && strings.HasPrefix(lineContent, "_content_hash:") {
			continue // Skip this line
		}
		
		filteredLines = append(filteredLines, line)
	}
	
	return strings.Join(filteredLines, "\n")
} 
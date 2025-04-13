// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"go.uber.org/zap"
)

// Reference represents a user story reference in a change request
type Reference struct {
	Title       string // Title of the user story
	FilePath    string // Path to the file
	ContentHash string // Content hash of the user story
}

// MismatchedReference represents a reference with a hash mismatch
type MismatchedReference struct {
	FilePath      string
	Title         string
	ReferenceHash string
	OldHash       string
}

// ChangeRequestInfo contains information about a change request file
type ChangeRequestInfo struct {
	FilePath   string
	References []Reference
}

// ContentHashPair represents a pair of old and new content hashes
type ContentHashPair struct {
	OldHash string
	NewHash string
}

// FindChangeRequestFiles finds all change request files in a directory
func FindChangeRequestFiles(root string, fs io.FileSystem) ([]string, error) {
	changeRequestDir := filepath.Join(root, "docs", "changes-request")
	
	// Check if the directory exists
	if !fs.Exists(changeRequestDir) {
		return nil, fmt.Errorf("change request directory not found: %s", changeRequestDir)
	}
	
	// Get all files in the directory
	entries, err := fs.ReadDir(changeRequestDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	
	var files []string
	
	// Look for all change request files, not just blueprint files
	for _, entry := range entries {
		if entry.IsDir() {
			// Recursively search subdirectories
			subdir := filepath.Join(changeRequestDir, entry.Name())
			subfiles, err := FindChangeRequestFiles(subdir, fs)
			if err != nil {
				logger.Warn("Error scanning subdirectory for change requests",
					zap.String("dir", subdir),
					zap.Error(err))
				// Continue with other directories even if one fails
				continue
			}
			files = append(files, subfiles...)
			continue
		}
		
		filename := entry.Name()
		// Include all markdown files in the change request directory
		if strings.HasSuffix(filename, ".md") {
			files = append(files, filepath.Join(changeRequestDir, filename))
		}
	}
	
	return files, nil
}

// ExtractReferences extracts references to user stories from content
func ExtractReferences(content string) []Reference {
	// Find all user stories references in the content
	// Pre-allocate the slice with a small initial capacity to avoid reallocation
	references := make([]Reference, 0, 10)
	
	// First find all title entries to establish the stories
	titleMatches := regexp.MustCompile(`(?m)^\s*-\s*title:\s*(.*?)(\s*\n)`).FindAllStringSubmatch(content, -1)
	
	for _, titleMatch := range titleMatches {
		if len(titleMatch) < 2 {
			continue
		}
		
		title := strings.TrimSpace(titleMatch[1])
		
		// Find the corresponding file and content-hash entries
		// This approach is more robust as it finds the actual file reference
		// Use the start of the title match to limit the search range
		titleIdx := strings.Index(content, titleMatch[0])
		if titleIdx == -1 {
			continue
		}
		
		// Find the next item or the end of the YAML section
		endIdx := strings.Index(content[titleIdx+len(titleMatch[0]):], "- title:")
		if endIdx == -1 {
			endIdx = strings.Index(content[titleIdx+len(titleMatch[0]):], "---")
		}
		
		var itemText string
		if endIdx == -1 {
			itemText = content[titleIdx:]
		} else {
			itemText = content[titleIdx : titleIdx+len(titleMatch[0])+endIdx]
		}
		
		// Extract file path and content hash from the item text
		fileMatch := regexp.MustCompile(`file:\s*(.*?)(\s*\n)`).FindStringSubmatch(itemText)
		if len(fileMatch) < 2 {
			continue
		}
		
		filePath := strings.TrimSpace(fileMatch[1])
		
		// Extract content hash
		hashMatch := regexp.MustCompile(`content-hash:\s*(.*?)(\s*\n)`).FindStringSubmatch(itemText)
		var contentHash string
		if len(hashMatch) >= 2 {
			contentHash = strings.TrimSpace(hashMatch[1])
		}
		
		// Add to references
		references = append(references, Reference{
			Title:       title,
			FilePath:    filePath,
			ContentHash: contentHash,
		})
	}
	
	return references
}

// ValidateChangedReferences checks all references against the hash map and reports any that need updating
func ValidateChangedReferences(references []Reference, hashMap ContentChangeMap) ([]Reference, []MismatchedReference) {
	changedReferences := []Reference{}
	mismatchedReferences := []MismatchedReference{}
	
	for _, ref := range references {
		if hashInfo, ok := hashMap[ref.FilePath]; ok && hashInfo.Changed {
			if hashInfo.OldHash == ref.ContentHash {
				changedReferences = append(changedReferences, ref)
			} else {
				// Reference hash doesn't match the old hash - might indicate a problem
				// Don't log here as we'll display the mismatches in a more user-friendly way
				
				// Add to mismatched references collection
				mismatchedReferences = append(mismatchedReferences, MismatchedReference{
					FilePath:      ref.FilePath,
					Title:         ref.Title,
					ReferenceHash: ref.ContentHash,
					OldHash:       hashInfo.OldHash,
				})
				
				changedReferences = append(changedReferences, ref)
			}
		}
	}
	
	return changedReferences, mismatchedReferences
}

// UpdateChangeRequestReferencesInContent updates references to change requests in the given content.
// It takes a map of file paths to content hashes, and replaces occurrences of the old
// hashes in the content with the new hashes. It returns the updated content and a slice
// of mismatched references.
func UpdateChangeRequestReferencesInContent(content string, changedFiles map[string]ContentHashPair) (string, []MismatchedReference) {
	var mismatchedReferences []MismatchedReference

	// Extract the existing references from the content
	existingRefs := ExtractReferences(content)

	// Split the content into lines for easier manipulation
	contentLines := strings.Split(content, "\n")
	
	// Process each existing reference
	for _, ref := range existingRefs {
		if pair, ok := changedFiles[ref.FilePath]; ok {
			// The file has changed, check if the content hash matches
			if ref.ContentHash != pair.OldHash {
				// Mismatch detected
				mismatchedReferences = append(mismatchedReferences, MismatchedReference{
					FilePath:      ref.FilePath,
					Title:         ref.Title,
					ReferenceHash: ref.ContentHash,
					OldHash:       pair.OldHash,
				})
			}
			
			// Update the hash in the content regardless of whether there was a mismatch
			// Need to find where this particular file reference is in the content
			for i, line := range contentLines {
				trimmedLine := strings.TrimSpace(line)
				if strings.HasPrefix(trimmedLine, "file:") && strings.Contains(line, ref.FilePath) {
					// Found the file reference, now find the content hash line that follows
					for j := i + 1; j < len(contentLines) && j < i + 5; j++ {
						// Look in the next few lines (max 5) for the content-hash line
						hashLine := strings.TrimSpace(contentLines[j])
						if strings.HasPrefix(hashLine, "content-hash:") {
							// Found the hash line, update it
							// Maintain the same indentation
							indentation := strings.Repeat(" ", len(contentLines[j]) - len(strings.TrimLeft(contentLines[j], " ")))
							contentLines[j] = indentation + "content-hash: " + pair.NewHash
							break
						}
					}
					break
				}
			}
		}
	}
	
	// Reconstruct the content with the updated lines
	updatedContent := strings.Join(contentLines, "\n")
	
	return updatedContent, mismatchedReferences
}

// FilterChangedContent filters the hash map to include only files with changed content
func FilterChangedContent(hashMap ContentChangeMap) ContentChangeMap {
	filteredMap := make(ContentChangeMap)
	
	for path, info := range hashMap {
		if info.Changed {
			filteredMap[path] = info
		}
	}
	
	return filteredMap
}

// ConvertToContentHashPairMap converts a ContentChangeMap to map[string]ContentHashPair
func ConvertToContentHashPairMap(changeMap ContentChangeMap) map[string]ContentHashPair {
	pairMap := make(map[string]ContentHashPair)
	
	for path, info := range changeMap {
		pairMap[path] = ContentHashPair{
			OldHash: info.OldHash,
			NewHash: info.NewHash,
		}
	}
	
	return pairMap
}

// UpdateAllChangeRequestReferences updates references in all change request files
// Returns:
// - []string: list of updated files
// - []string: list of unchanged files
// - int: total number of references updated
// - []MismatchedReference: list of references with mismatched hashes
// - error: any error that occurred
func UpdateAllChangeRequestReferences(root string, hashMap ContentChangeMap, fs io.FileSystem) ([]string, []string, int, []MismatchedReference, error) {
	// Filter the hash map to include only files with changed content
	changedMap := FilterChangedContent(hashMap)
	
	// If no content has changed, no need to update references
	if len(changedMap) == 0 {
		logger.Debug("No content changes detected, skipping reference updates")
		return nil, nil, 0, nil, nil
	}
	
	// Find all change request files
	files, err := FindChangeRequestFiles(root, fs)
	if err != nil {
		return nil, nil, 0, nil, fmt.Errorf("failed to find change request files: %w", err)
	}
	
	updatedFiles := make([]string, 0, len(files))
	unchangedFiles := make([]string, 0, len(files))
	allMismatchedRefs := make([]MismatchedReference, 0)
	totalReferencesUpdated := 0
	errors := make([]string, 0) // Track any errors during processing
	
	// Check and update references in each file
	for _, file := range files {
		logger.Debug("Processing change request", zap.String("file", file))
		
		updated, refsUpdated, mismatchedReferences, err := UpdateChangeRequestReferences(file, changedMap, fs)
		if err != nil {
			logger.Error("Failed to update references", 
				zap.String("file", file), 
				zap.Error(err))
			errors = append(errors, fmt.Sprintf("%s: %s", file, err.Error()))
			continue
		}
		
		// Collect all mismatched references
		allMismatchedRefs = append(allMismatchedRefs, mismatchedReferences...)
		
		relPath, err := filepath.Rel(root, file)
		if err != nil {
			relPath = file // Use full path if relative path can't be determined
		}
		
		if updated {
			updatedFiles = append(updatedFiles, relPath)
			totalReferencesUpdated += refsUpdated
		} else {
			unchangedFiles = append(unchangedFiles, relPath)
		}
	}
	
	// If there were any errors, log a summary
	if len(errors) > 0 {
		logger.Warn("Some files could not be updated", 
			zap.Int("error_count", len(errors)),
			zap.Strings("errors", errors))
	}
	
	// Stats for logging
	stats := map[string]int{
		"total": len(files),
		"updated": len(updatedFiles),
		"unchanged": len(unchangedFiles),
		"errors": len(errors),
		"references_updated": totalReferencesUpdated,
	}
	
	logger.Debug("Completed change request reference update", 
		zap.Int("total", stats["total"]),
		zap.Int("updated", stats["updated"]),
		zap.Int("unchanged", stats["unchanged"]),
		zap.Int("errors", stats["errors"]),
		zap.Int("references_updated", stats["references_updated"]))
	
	return updatedFiles, unchangedFiles, totalReferencesUpdated, allMismatchedRefs, nil
}

// UpdateChangeRequestReferences updates references in a single change request file
// Returns:
// - bool: whether the file was updated
// - int: number of references updated
// - []MismatchedReference: list of references with mismatched hashes
// - error: any error that occurred
func UpdateChangeRequestReferences(filePath string, hashMap ContentChangeMap, fs io.FileSystem) (bool, int, []MismatchedReference, error) {
	// Read the file content
	content, err := fs.ReadFile(filePath)
	if err != nil {
		return false, 0, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	// Convert ContentChangeMap to map[string]ContentHashPair
	pairMap := ConvertToContentHashPairMap(hashMap)
	
	// Update references in the content
	updatedContent, mismatchedReferences := UpdateChangeRequestReferencesInContent(string(content), pairMap)
	
	// Count the number of references that were updated
	referencesUpdated := 0
	
	// Extract the existing references from the content to determine how many were updated
	existingRefs := ExtractReferences(string(content))
	for _, ref := range existingRefs {
		if _, ok := pairMap[ref.FilePath]; ok {
			// The file is in the change map - we'll update it regardless of hash match
			referencesUpdated++
		}
	}
	
	// If content was updated, write it back to the file
	if updatedContent != string(content) {
		err = fs.WriteFile(filePath, []byte(updatedContent), 0644)
		if err != nil {
			return false, referencesUpdated, mismatchedReferences, fmt.Errorf("failed to write updated file %s: %w", filePath, err)
		}
		return true, referencesUpdated, mismatchedReferences, nil
	}
	
	return false, 0, mismatchedReferences, nil
} 
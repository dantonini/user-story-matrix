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

// Regular expression to match user story references in change request files
var userStoryReferenceRegex = regexp.MustCompile(`(?m)^(\s*-\s*title:\s*.+\n\s*file:\s*)([^\n]+)(\n\s*content-hash:\s*)([^\n]+)(\n)`)

// Reference represents a user story reference in a change request
type Reference struct {
	Title       string
	FilePath    string
	ContentHash string
	Line        int // Line number in the change request file
}

// ChangeRequestInfo contains information about a change request file
type ChangeRequestInfo struct {
	FilePath   string
	References []Reference
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

// ExtractReferences extracts all user story references from a change request file
func ExtractReferences(content string) []Reference {
	references := []Reference{}
	matches := userStoryReferenceRegex.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		// The match array should contain:
		// [0]: full match
		// [1]: prefix (spaces + "- title:" + content + newline + spaces + "file:")
		// [2]: file path
		// [3]: newline + spaces + "content-hash:"
		// [4]: content hash
		// [5]: newline
		if len(match) < 6 {
			continue
		}
		
		filePath := match[2]
		contentHash := match[4]
		
		// Extract title from the previous line
		titleStart := strings.LastIndex(match[1], "title:")
		if titleStart == -1 {
			continue
		}
		titleLine := match[1][titleStart:]
		titleEnd := strings.Index(titleLine, "\n")
		if titleEnd == -1 {
			continue
		}
		title := strings.TrimSpace(strings.TrimPrefix(titleLine[:titleEnd], "title:"))
		
		references = append(references, Reference{
			Title:       title,
			FilePath:    filePath,
			ContentHash: contentHash,
			Line:        0, // TODO: Calculate actual line number
		})
	}
	
	return references
}

// ValidateChangedReferences checks all references against the hash map and reports any that need updating
func ValidateChangedReferences(references []Reference, hashMap ContentChangeMap) []Reference {
	changedReferences := []Reference{}
	
	for _, ref := range references {
		if hashInfo, ok := hashMap[ref.FilePath]; ok && hashInfo.Changed {
			if hashInfo.OldHash == ref.ContentHash {
				changedReferences = append(changedReferences, ref)
			} else {
				// Reference hash doesn't match the old hash - might indicate a problem
				logger.Warn("Reference hash doesn't match stored old hash",
					zap.String("file", ref.FilePath),
					zap.String("reference_hash", ref.ContentHash),
					zap.String("old_hash", hashInfo.OldHash))
				changedReferences = append(changedReferences, ref)
			}
		}
	}
	
	return changedReferences
}

// UpdateChangeRequestReferences updates references in change request files
// Returns:
// - bool: whether the file was updated
// - int: number of references updated
// - error: any error that occurred
func UpdateChangeRequestReferences(filePath string, hashMap ContentChangeMap, fs io.FileSystem) (bool, int, error) {
	// Read file content
	content, err := fs.ReadFile(filePath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read change request file: %w", err)
	}
	
	contentStr := string(content)
	
	changesMade := false
	updatedReferences := 0
	
	// Extract all references
	references := ExtractReferences(contentStr)
	
	// Validate which references need updating
	changedReferences := ValidateChangedReferences(references, hashMap)
	
	if len(changedReferences) == 0 {
		return false, 0, nil
	}
	
	// Find all user story references
	matches := userStoryReferenceRegex.FindAllStringSubmatch(contentStr, -1)
	matchIndices := userStoryReferenceRegex.FindAllStringSubmatchIndex(contentStr, -1)
	
	// Process matches one by one
	for i, match := range matches {
		matchIndex := matchIndices[i]
		
		// Extract the file path and current hash
		filePath := match[2]
		currentHash := match[4]
		
		// Check if this file is in our hash map
		if hashInfo, ok := hashMap[filePath]; ok && hashInfo.Changed {
			// Update the content hash in the document
			// We need to find where in the string the content hash starts and ends
			hashStartPos := matchIndex[8]
			hashEndPos := matchIndex[9]
			
			// Update the content hash
			newContent := contentStr[:hashStartPos] + hashInfo.NewHash + contentStr[hashEndPos:]
			contentStr = newContent
			changesMade = true
			updatedReferences++
			
			logger.Debug("Updated reference hash", 
				zap.String("file", filePath),
				zap.String("old_hash", currentHash),
				zap.String("new_hash", hashInfo.NewHash))
		}
	}
	
	// Write the updated content back to the file if changes were made
	if changesMade {
		fileInfo, err := fs.Stat(filePath)
		if err != nil {
			return false, updatedReferences, fmt.Errorf("failed to get file info: %w", err)
		}
		
		err = fs.WriteFile(filePath, []byte(contentStr), fileInfo.Mode())
		if err != nil {
			return false, updatedReferences, fmt.Errorf("failed to write updated content: %w", err)
		}
	}
	
	return changesMade, updatedReferences, nil
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

// UpdateAllChangeRequestReferences updates references in all change request files
// Returns:
// - []string: list of updated files
// - []string: list of unchanged files
// - int: total number of references updated
// - error: any error that occurred
func UpdateAllChangeRequestReferences(root string, hashMap ContentChangeMap, fs io.FileSystem) ([]string, []string, int, error) {
	// Filter the hash map to include only files with changed content
	changedMap := FilterChangedContent(hashMap)
	
	// If no content has changed, no need to update references
	if len(changedMap) == 0 {
		logger.Debug("No content changes detected, skipping reference updates")
		return nil, nil, 0, nil
	}
	
	// Find all change request files
	files, err := FindChangeRequestFiles(root, fs)
	if err != nil {
		return nil, nil, 0, err
	}

	if len(files) == 0 {
		logger.Warn("No change request files found", zap.String("root", root))
		return nil, nil, 0, nil
	}
	
	updatedFiles := make([]string, 0, len(files))
	unchangedFiles := make([]string, 0, len(files))
	errors := make([]string, 0)
	totalReferencesUpdated := 0
	
	// Update references in each file
	for _, file := range files {
		logger.Debug("Processing change request", zap.String("file", file))
		
		updated, referencesUpdated, err := UpdateChangeRequestReferences(file, changedMap, fs)
		if err != nil {
			logger.Error("Failed to update references", 
				zap.String("file", file), 
				zap.Error(err))
			errors = append(errors, fmt.Sprintf("%s: %s", file, err.Error()))
			continue
		}
		
		relPath, err := filepath.Rel(root, file)
		if err != nil {
			relPath = file // Use full path if relative path can't be determined
		}
		
		if updated {
			updatedFiles = append(updatedFiles, relPath)
			totalReferencesUpdated += referencesUpdated
			logger.Debug("Updated references", 
				zap.String("file", relPath),
				zap.Int("references_updated", referencesUpdated))
		} else {
			unchangedFiles = append(unchangedFiles, relPath)
		}
	}
	
	// If there were any errors, log a summary
	if len(errors) > 0 {
		logger.Warn("Some change requests could not be updated", 
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
	
	logger.Info("Completed change request reference update", 
		zap.Int("total", stats["total"]),
		zap.Int("updated", stats["updated"]),
		zap.Int("unchanged", stats["unchanged"]),
		zap.Int("errors", stats["errors"]),
		zap.Int("references_updated", stats["references_updated"]))
	
	return updatedFiles, unchangedFiles, totalReferencesUpdated, nil
} 
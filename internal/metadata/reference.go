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
	
	// Filter for blueprint files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		filename := entry.Name()
		if strings.HasSuffix(filename, ".blueprint.md") {
			files = append(files, filepath.Join(changeRequestDir, filename))
		}
	}
	
	return files, nil
}

// UpdateChangeRequestReferences updates references in change request files
// Returns:
// - bool: whether the file was updated
// - error: any error that occurred
func UpdateChangeRequestReferences(filePath string, hashMap ContentChangeMap, fs io.FileSystem) (bool, error) {
	// Read file content
	content, err := fs.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read change request file: %w", err)
	}
	
	contentStr := string(content)
	changesMade := false
	
	// Find all user story references
	matches := userStoryReferenceRegex.FindAllStringSubmatchIndex(contentStr, -1)
	
	// Process matches in reverse order to avoid index issues
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		
		// Extract the file path and current hash
		filePath := contentStr[match[4]:match[5]]
		currentHash := contentStr[match[8]:match[9]]
		
		// Check if this file is in our hash map
		if hashInfo, ok := hashMap[filePath]; ok && hashInfo.Changed {
			// Update the content hash
			newContent := contentStr[:match[8]] + hashInfo.NewHash + contentStr[match[9]:]
			contentStr = newContent
			changesMade = true
			
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
			return false, fmt.Errorf("failed to get file info: %w", err)
		}
		
		err = fs.WriteFile(filePath, []byte(contentStr), fileInfo.Mode())
		if err != nil {
			return false, fmt.Errorf("failed to write updated content: %w", err)
		}
	}
	
	return changesMade, nil
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
// - error: any error that occurred
func UpdateAllChangeRequestReferences(root string, hashMap ContentChangeMap, fs io.FileSystem) ([]string, []string, error) {
	// Filter the hash map to include only files with changed content
	changedMap := FilterChangedContent(hashMap)
	
	// If no content has changed, no need to update references
	if len(changedMap) == 0 {
		logger.Debug("No content changes detected, skipping reference updates")
		return nil, nil, nil
	}
	
	// Find all change request files
	files, err := FindChangeRequestFiles(root, fs)
	if err != nil {
		return nil, nil, err
	}
	
	updatedFiles := make([]string, 0, len(files))
	unchangedFiles := make([]string, 0, len(files))
	
	// Update references in each file
	for _, file := range files {
		logger.Debug("Processing change request", zap.String("file", file))
		
		updated, err := UpdateChangeRequestReferences(file, changedMap, fs)
		if err != nil {
			logger.Error("Failed to update references", 
				zap.String("file", file), 
				zap.Error(err))
			fmt.Printf("Error updating references in %s: %s\n", file, err)
			continue
		}
		
		relPath, err := filepath.Rel(root, file)
		if err != nil {
			relPath = file // Use full path if relative path can't be determined
		}
		
		if updated {
			updatedFiles = append(updatedFiles, relPath)
			logger.Debug("Updated references", zap.String("file", relPath))
		} else {
			unchangedFiles = append(unchangedFiles, relPath)
		}
	}
	
	return updatedFiles, unchangedFiles, nil
} 
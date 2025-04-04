// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"go.uber.org/zap"
)

// UpdateFileMetadata updates the metadata section of a file
// Returns:
// - bool: whether the file was updated
// - ContentHashMap: information about content hash changes
// - error: any error that occurred
func UpdateFileMetadata(filePath, root string, fs io.FileSystem) (bool, ContentHashMap, error) {
	hashMap := ContentHashMap{
		FilePath: filePath,
	}

	// Get file info
	fileInfo, err := fs.Stat(filePath)
	if err != nil {
		return false, hashMap, err
	}

	// Read file content
	content, err := fs.ReadFile(filePath)
	if err != nil {
		return false, hashMap, err
	}

	// Extract existing metadata
	existingMetadata, err := ExtractMetadata(string(content))
	if err != nil {
		return false, hashMap, err
	}

	// Calculate content hash
	contentWithoutMetadata := GetContentWithoutMetadata(string(content))
	contentHash := CalculateContentHash(contentWithoutMetadata)

	// Store old and new hash in the hash map
	hashMap.OldHash = existingMetadata.ContentHash
	hashMap.NewHash = contentHash
	
	// Flag whether content has actually changed
	hashMap.Changed = existingMetadata.ContentHash != contentHash

	// Generate new metadata
	newMetadata := GenerateMetadata(filePath, root, fileInfo, existingMetadata, contentHash)

	// Check if metadata has changed (to avoid unnecessary updates)
	currentMetadataBytes := metadataRegex.Find(content)
	
	if string(currentMetadataBytes) == newMetadata || 
		(len(currentMetadataBytes) == 0 && len(existingMetadata.RawMetadata) == 0 && contentWithoutMetadata == string(content)) {
		// No changes needed
		return false, hashMap, nil
	}

	// Update the file with new metadata
	newContent := newMetadata + contentWithoutMetadata
	err = fs.WriteFile(filePath, []byte(newContent), fileInfo.Mode())
	if err != nil {
		return false, hashMap, err
	}

	return true, hashMap, nil
}

// FindMarkdownFiles recursively finds all markdown files in a directory
func FindMarkdownFiles(dir string, fs io.FileSystem) ([]string, error) {
	var files []string

	entries, err := fs.ReadDir(dir)
	if err != nil {
		return files, err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		// Skip ignored directories
		if entry.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || base == "dist" || base == "build" {
				logger.Debug("Skipping directory", zap.String("dir", path))
				continue
			}

			// Recursively process subdirectories
			subfiles, err := FindMarkdownFiles(path, fs)
			if err != nil {
				return files, err
			}
			files = append(files, subfiles...)
		} else if strings.HasSuffix(strings.ToLower(path), ".md") {
			// Add markdown files
			files = append(files, path)
			logger.Debug("Found markdown file", zap.String("file", path))
		}
	}

	return files, nil
}

// UpdateAllUserStoryMetadata updates metadata for all user story files
// Returns:
// - []string: list of updated files
// - []string: list of unchanged files
// - ContentChangeMap: map of file paths to hash change information
// - error: any error that occurred
func UpdateAllUserStoryMetadata(userStoriesDir, root string, fs io.FileSystem) ([]string, []string, ContentChangeMap, error) {
	// Find all markdown files in the user stories directory
	files, err := FindMarkdownFiles(userStoriesDir, fs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to find markdown files: %w", err)
	}

	updatedFiles := make([]string, 0, len(files))
	unchangedFiles := make([]string, 0, len(files))
	hashMap := make(ContentChangeMap)

	// Update metadata for each file
	for _, file := range files {
		logger.Debug("Processing file", zap.String("file", file))

		updated, fileHashMap, err := UpdateFileMetadata(file, root, fs)
		if err != nil {
			logger.Error("Failed to update metadata", 
				zap.String("file", file), 
				zap.Error(err))
			fmt.Printf("Error updating %s: %s\n", file, err)
			continue
		}

		relPath, err := filepath.Rel(root, file)
		if err != nil {
			relPath = file // Use full path if relative path can't be determined
		}

		if updated {
			updatedFiles = append(updatedFiles, relPath)
			hashMap[relPath] = fileHashMap
		} else {
			unchangedFiles = append(unchangedFiles, relPath)
		}
	}

	return updatedFiles, unchangedFiles, hashMap, nil
} 
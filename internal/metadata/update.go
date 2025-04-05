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

// SkippedDirectories is a list of directories to skip when scanning for markdown files
var SkippedDirectories = []string{
	"node_modules",
	".git",
	"dist",
	"build",
	"vendor",  // Added vendor directory to skip
	"tmp",     // Added tmp directory to skip
	"temp",    // Added temp directory to skip
	".cache",  // Added .cache directory to skip
	".github", // Added .github directory to skip
}

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
		return false, hashMap, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	// Read file content
	content, err := fs.ReadFile(filePath)
	if err != nil {
		return false, hashMap, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	logger.Debug("Read file content", 
		zap.String("file", filePath),
		zap.Int("content_length", len(content)))

	// Extract existing metadata
	existingMetadata, err := ExtractMetadata(string(content))
	if err != nil {
		return false, hashMap, fmt.Errorf("failed to extract metadata from %s: %w", filePath, err)
	}

	// Calculate content hash
	contentWithoutMetadata := GetContentWithoutMetadata(string(content))
	contentHash := CalculateContentHash(contentWithoutMetadata)
	
	logger.Debug("Calculated content hash", 
		zap.String("file", filePath),
		zap.String("hash", contentHash),
		zap.String("old_hash", existingMetadata.ContentHash))

	// Store old and new hash in the hash map
	hashMap.OldHash = existingMetadata.ContentHash
	hashMap.NewHash = contentHash
	
	// Flag whether content has actually changed
	hashMap.Changed = existingMetadata.ContentHash != contentHash

	// Generate new metadata
	newMetadata := GenerateMetadata(filePath, root, fileInfo, existingMetadata, contentHash)
	
	logger.Debug("Generated new metadata", 
		zap.String("file", filePath),
		zap.String("metadata", newMetadata))

	// Check if metadata has changed (to avoid unnecessary updates)
	currentMetadataBytes := metadataRegex.Find(content)
	
	// FIXED CONDITION: A file needs updating if any of these conditions are true:
	// 1. The file has no metadata section at all (len(currentMetadataBytes) == 0)
	// 2. The existing metadata doesn't match the new metadata
	needsUpdate := len(currentMetadataBytes) == 0 || string(currentMetadataBytes) != newMetadata
	
	if !needsUpdate {
		// No changes needed
		logger.Debug("No metadata changes needed", 
			zap.String("file", filePath),
			zap.Bool("content_changed", hashMap.Changed))
		return false, hashMap, nil
	}

	// Update the file with new metadata
	newContent := newMetadata + contentWithoutMetadata
	
	logger.Debug("Writing updated content", 
		zap.String("file", filePath),
		zap.Int("content_length", len(newContent)))
	
	err = fs.WriteFile(filePath, []byte(newContent), fileInfo.Mode())
	if err != nil {
		return false, hashMap, fmt.Errorf("failed to write updated file %s: %w", filePath, err)
	}
	
	// Verify the file was updated - read it back for validation
	verifyContent, verifyErr := fs.ReadFile(filePath)
	if verifyErr != nil {
		logger.Warn("Could not verify file update", 
			zap.String("file", filePath),
			zap.Error(verifyErr))
	} else if string(verifyContent) != newContent {
		logger.Warn("File content verification failed",
			zap.String("file", filePath),
			zap.Int("expected_length", len(newContent)),
			zap.Int("actual_length", len(verifyContent)))
	}

	logger.Debug("Updated file metadata", 
		zap.String("file", filePath),
		zap.Bool("content_changed", hashMap.Changed),
		zap.String("new_hash", contentHash))

	return true, hashMap, nil
}

// ShouldSkipDirectory checks if the directory should be skipped
func ShouldSkipDirectory(dirName string) bool {
	for _, skipDir := range SkippedDirectories {
		if dirName == skipDir {
			return true
		}
	}
	return false
}

// FindMarkdownFiles recursively finds all markdown files in a directory
func FindMarkdownFiles(dir string, fs io.FileSystem) ([]string, error) {
	var files []string

	// Check if the directory exists
	if !fs.Exists(dir) {
		return files, fmt.Errorf("directory not found: %s", dir)
	}

	entries, err := fs.ReadDir(dir)
	if err != nil {
		return files, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		// Skip ignored directories
		if entry.IsDir() {
			base := filepath.Base(path)
			if ShouldSkipDirectory(base) {
				logger.Debug("Skipping directory", zap.String("dir", path))
				continue
			}

			// Recursively process subdirectories
			subfiles, err := FindMarkdownFiles(path, fs)
			if err != nil {
				logger.Warn("Error scanning subdirectory", 
					zap.String("dir", path), 
					zap.Error(err))
				// Continue scanning other directories even if one fails
				continue
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

	if len(files) == 0 {
		logger.Warn("No markdown files found in directory", zap.String("dir", userStoriesDir))
		return nil, nil, nil, nil
	}

	updatedFiles := make([]string, 0, len(files))
	unchangedFiles := make([]string, 0, len(files))
	hashMap := make(ContentChangeMap)
	errors := make([]string, 0) // Track any errors during processing

	// Update metadata for each file
	for _, file := range files {
		logger.Debug("Processing file", zap.String("file", file))

		updated, fileHashMap, err := UpdateFileMetadata(file, root, fs)
		if err != nil {
			logger.Error("Failed to update metadata", 
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
			hashMap[relPath] = fileHashMap
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
	}

	logger.Debug("Completed user story metadata update", 
		zap.Int("total", stats["total"]),
		zap.Int("updated", stats["updated"]),
		zap.Int("unchanged", stats["unchanged"]),
		zap.Int("errors", stats["errors"]))

	return updatedFiles, unchangedFiles, hashMap, nil
} 
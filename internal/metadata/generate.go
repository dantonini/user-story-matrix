// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user-story-matrix/usm/internal/logger"
	"go.uber.org/zap"
)

// CalculateContentHash calculates the SHA-256 hash of content
func CalculateContentHash(content string) string {
	hash := sha256.New()
	hash.Write([]byte(content))
	return hex.EncodeToString(hash.Sum(nil))
}

// GenerateMetadata creates a metadata section for a file
func GenerateMetadata(filePath, root string, fileInfo os.FileInfo, existingMetadata Metadata, contentHash string) string {
	// Get the relative path
	relativePath, err := filepath.Rel(root, filePath)
	if err != nil {
		relativePath = filePath // Use full path if relative path can't be determined
	}
	
	// Use existing creation date if available, otherwise use file modification time
	// This preserves the original creation date as required by the user story
	var creationDate string
	if !existingMetadata.CreatedAt.IsZero() {
		creationDate = existingMetadata.CreatedAt.Format(time.RFC3339)
	} else if createdAt, ok := existingMetadata.RawMetadata["created_at"]; ok && createdAt != "" {
		creationDate = createdAt
	} else {
		creationDate = fileInfo.ModTime().Format(time.RFC3339) // Use mod time as fallback
	}
	
	// Check if content has changed by comparing hashes
	storedHash := existingMetadata.ContentHash
	contentChanged := storedHash != contentHash
	
	// Only update last_updated date if content has changed or it doesn't exist
	var modifiedDate string
	if !existingMetadata.LastUpdated.IsZero() && !contentChanged {
		modifiedDate = existingMetadata.LastUpdated.Format(time.RFC3339)
	} else if lastUpdated, ok := existingMetadata.RawMetadata["last_updated"]; ok && lastUpdated != "" && !contentChanged {
		modifiedDate = lastUpdated
	} else {
		modifiedDate = time.Now().Format(time.RFC3339)
		logger.Debug("Updating modified date", 
			zap.String("file", relativePath), 
			zap.String("old_hash", storedHash), 
			zap.String("new_hash", contentHash),
			zap.Bool("content_changed", contentChanged))
	}
	
	// Build the metadata section
	metadata := fmt.Sprintf("---\nfile_path: %s\ncreated_at: %s\nlast_updated: %s\n_content_hash: %s\n---\n\n", 
		relativePath, creationDate, modifiedDate, contentHash)
	
	return metadata
}

// FormatMetadata formats a Metadata struct into a string representation
func FormatMetadata(metadata Metadata, contentHash string) string {
	creationDate := metadata.CreatedAt.Format(time.RFC3339)
	modifiedDate := metadata.LastUpdated.Format(time.RFC3339)
	
	return fmt.Sprintf("---\nfile_path: %s\ncreated_at: %s\nlast_updated: %s\n_content_hash: %s\n---\n\n", 
		metadata.FilePath, creationDate, modifiedDate, contentHash)
} 
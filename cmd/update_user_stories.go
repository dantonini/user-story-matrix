// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/logger"
	"go.uber.org/zap"
)

// Regex pattern to match metadata section
var metadataRegex = regexp.MustCompile(`(?m)^---\s*\n([\s\S]*?)\n---\s*\n`)

// Regex pattern to match specific metadata key-value pairs
var metadataKeyValueRegex = regexp.MustCompile(`(?m)^([^:]+):\s*(.*)$`)

// updateUserStoriesCmd represents the update user-stories metadata command
var updateUserStoriesCmd = &cobra.Command{
	Use:   "update user-stories metadata",
	Short: "Update metadata in user story markdown files",
	Long: `Update metadata in user story markdown files.
	
This command recursively scans for Markdown files in the docs/user-stories directory
and ensures each has an up-to-date metadata section containing:
- File path (relative to project root)
- Creation date
- Last edited date
- Content hash (hidden with underscore prefix)`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Updating user story metadata")
		
		// Get the project root directory
		root, err := os.Getwd()
		if err != nil {
			logger.Error("Failed to get current directory", zap.Error(err))
			fmt.Fprintf(os.Stderr, "Error: Failed to get current directory: %s\n", err)
			return
		}
		
		// Check for the --test-root flag (only used in tests)
		var userStoriesDir string
		testRoot, err := cmd.Flags().GetString("test-root")
		if err != nil {
			logger.Error("Failed to get test-root flag", zap.Error(err))
			fmt.Fprintf(os.Stderr, "Error: Failed to get test-root flag: %s\n", err)
			return
		}
		if testRoot != "" {
			// For testing, use the specified directory
			userStoriesDir = filepath.Join(testRoot, "docs", "user-stories")
			logger.Debug("Using test root directory",
				zap.String("test_root", testRoot),
				zap.String("user_stories_dir", userStoriesDir))
		} else {
			// Normal operation: use current directory
			docsDir := filepath.Join(root, "docs")
			userStoriesDir = filepath.Join(docsDir, "user-stories")
		}
		
		logger.Debug("Scanning for user stories", 
			zap.String("dir", userStoriesDir),
			zap.String("root", root))
		
		// Scan for markdown files
		files, err := findMarkdownFiles(userStoriesDir)
		if err != nil {
			logger.Error("Failed to find markdown files", zap.Error(err))
			fmt.Fprintf(os.Stderr, "Error: Failed to find markdown files: %s\n", err)
			return
		}
		
		logger.Info(fmt.Sprintf("Found %d markdown files", len(files)))
		
		// Update metadata for each file
		updatedCount := 0
		unchangedCount := 0
		
		for _, file := range files {
			logger.Debug("Processing file", zap.String("file", file))
			
			updated, hash, err := updateFileMetadata(file, root)
			if err != nil {
				logger.Error("Failed to update metadata", zap.String("file", file), zap.Error(err))
				fmt.Fprintf(os.Stderr, "Error updating %s: %s\n", file, err)
				continue
			}
			
			relPath, err := filepath.Rel(root, file)
			if err != nil {
				relPath = file // Use full path if relative path can't be determined
			}
			
			if updated {
				updatedCount++
				logger.Debug("Updated metadata", 
					zap.String("file", relPath), 
					zap.String("content_hash", hash))
				fmt.Printf("✅ Updated metadata for: %s\n", relPath)
			} else {
				unchangedCount++
				logger.Debug("No changes needed", 
					zap.String("file", relPath), 
					zap.String("content_hash", hash))
				fmt.Printf("ℹ️ No changes needed for: %s\n", relPath)
			}
		}
		
		logger.Debug("Processing complete", 
			zap.Int("total", len(files)), 
			zap.Int("updated", updatedCount), 
			zap.Int("unchanged", unchangedCount))
		
		fmt.Printf("✨ Processed %d files (%d updated, %d unchanged)\n", len(files), updatedCount, unchangedCount)
	},
}

// findMarkdownFiles recursively finds all markdown files in a directory
func findMarkdownFiles(dir string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip ignored directories
		base := filepath.Base(path)
		if info.IsDir() && (base == "node_modules" || base == ".git" || base == "dist" || base == "build") {
			logger.Debug("Skipping directory", zap.String("dir", path))
			return filepath.SkipDir
		}
		
		// Add markdown files
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			files = append(files, path)
			logger.Debug("Found markdown file", zap.String("file", path))
		}
		
		return nil
	})
	
	return files, err
}

// extractExistingMetadata extracts metadata from file content
func extractExistingMetadata(content string) map[string]string {
	metadata := make(map[string]string)
	
	matches := metadataRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return metadata
	}
	
	metadataText := matches[1]
	kvMatches := metadataKeyValueRegex.FindAllStringSubmatch(metadataText, -1)
	
	for _, kv := range kvMatches {
		if len(kv) >= 3 {
			key := strings.TrimSpace(kv[1])
			value := strings.TrimSpace(kv[2])
			if key != "" && value != "" {
				metadata[key] = value
			}
		}
	}
	
	return metadata
}

// getContentWithoutMetadata removes metadata section from content
func getContentWithoutMetadata(content string) string {
	return metadataRegex.ReplaceAllString(content, "")
}

// calculateContentHash calculates SHA-256 hash of content
func calculateContentHash(content string) string {
	hash := sha256.New()
	hash.Write([]byte(content))
	return hex.EncodeToString(hash.Sum(nil))
}

// generateMetadata creates a metadata section for a file
func generateMetadata(filePath, root string, fileInfo os.FileInfo, existingMetadata map[string]string, contentHash string) string {
	relativePath, err := filepath.Rel(root, filePath)
	if err != nil {
		relativePath = filePath // Use full path if relative path can't be determined
	}
	
	// Use existing creation date if available, otherwise use file creation time
	creationDate := existingMetadata["created_at"]
	if creationDate == "" {
		creationDate = fileInfo.ModTime().Format(time.RFC3339) // Use mod time as fallback for birthtime
	}
	
	// Check if content has changed
	storedHash := existingMetadata["_content_hash"]
	contentChanged := storedHash != contentHash
	
	// Only update last_updated date if content has changed or it doesn't exist
	modifiedDate := existingMetadata["last_updated"]
	if modifiedDate == "" || contentChanged {
		modifiedDate = fileInfo.ModTime().Format(time.RFC3339)
		logger.Debug("Updating modified date", 
			zap.String("file", relativePath), 
			zap.String("old_hash", storedHash), 
			zap.String("new_hash", contentHash),
			zap.Bool("content_changed", contentChanged))
	}
	
	// Build metadata section
	metadata := fmt.Sprintf("---\nfile_path: %s\ncreated_at: %s\nlast_updated: %s\n_content_hash: %s\n---\n\n", 
		relativePath, creationDate, modifiedDate, contentHash)
	
	return metadata
}

// updateFileMetadata updates the metadata section of a file
func updateFileMetadata(filePath, root string) (bool, string, error) {
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, "", err
	}
	
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, "", err
	}
	
	// Extract existing metadata
	existingMetadata := extractExistingMetadata(string(content))
	
	// Calculate content hash
	contentWithoutMetadata := getContentWithoutMetadata(string(content))
	contentHash := calculateContentHash(contentWithoutMetadata)
	
	// Generate new metadata
	metadata := generateMetadata(filePath, root, fileInfo, existingMetadata, contentHash)
	
	// Check if file already has metadata
	var newContent string
	if metadataRegex.MatchString(string(content)) {
		// Replace existing metadata
		newContent = metadataRegex.ReplaceAllString(string(content), metadata)
	} else {
		// Add metadata to the beginning
		newContent = metadata + string(content)
	}
	
	// Only write if content or metadata has changed
	if newContent != string(content) {
		logger.Debug("Content or metadata changed, updating file", 
			zap.String("file", filePath))
		
		err = os.WriteFile(filePath, []byte(newContent), 0600)
		if err != nil {
			return false, contentHash, err
		}
		return true, contentHash, nil
	}
	
	return false, contentHash, nil
}

func init() {
	rootCmd.AddCommand(updateUserStoriesCmd)
	// Add a hidden flag for testing
	updateUserStoriesCmd.Flags().String("test-root", "", "Root directory for testing (hidden)")
	if err := updateUserStoriesCmd.Flags().MarkHidden("test-root"); err != nil {
		logger.Error("Failed to mark flag as hidden", zap.Error(err))
	}
	logger.Debug("Update user-stories command added to root command")
} 
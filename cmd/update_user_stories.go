// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/metadata"
	"go.uber.org/zap"
)

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
- Content hash (hidden with underscore prefix)

By default, it also updates content hash references in change request files when user story
content changes. Use the --skip-references flag to disable this behavior.

Directories like node_modules, .git, dist, build, vendor, tmp, .cache, and .github are automatically skipped.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Updating user story metadata")
		
		// Get command options
		skipReferences, _ := cmd.Flags().GetBool("skip-references")
		debug, _ := cmd.Flags().GetBool("debug")
		
		// If debug mode is enabled, adjust the logger level
		if debug {
			logger.SetDebugMode(true)
			logger.Debug("Debug mode enabled")
		}
		
		// Get the project root directory
		root, err := os.Getwd()
		if err != nil {
			logger.Error("Failed to get current directory", zap.Error(err))
			fmt.Fprintf(os.Stderr, "Error: Failed to get current directory: %s\n", err)
			return
		}
		
		// Initialize the file system
		fs := io.NewOSFileSystem()
		
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
			root = testRoot
		} else {
			// Normal operation: use current directory
			docsDir := filepath.Join(root, "docs")
			userStoriesDir = filepath.Join(docsDir, "user-stories")
		}
		
		logger.Debug("Scanning for user stories", 
			zap.String("dir", userStoriesDir),
			zap.String("root", root))
		
		// Update all user story metadata
		updatedFiles, unchangedFiles, hashMap, err := metadata.UpdateAllUserStoryMetadata(userStoriesDir, root, fs)
		if err != nil {
			logger.Error("Failed to update user story metadata", zap.Error(err))
			fmt.Fprintf(os.Stderr, "Error: Failed to update user story metadata: %s\n", err)
			return
		}
		
		// Print summary of user story updates
		if len(updatedFiles) > 0 {
			fmt.Println("📋 Updated user story metadata:")
			for _, file := range updatedFiles {
				fmt.Printf("   ✅ %s\n", file)
			}
		}
		
		if debug && len(unchangedFiles) > 0 {
			fmt.Println("📋 Unchanged user stories:")
			for _, file := range unchangedFiles {
				fmt.Printf("   ℹ️ %s\n", file)
			}
		}
		
		logger.Debug("Processing of user stories complete", 
			zap.Int("total", len(updatedFiles) + len(unchangedFiles)), 
			zap.Int("updated", len(updatedFiles)), 
			zap.Int("unchanged", len(unchangedFiles)))
		
		// If references shouldn't be skipped and we have content changes, update references
		if !skipReferences && len(hashMap) > 0 {
			logger.Info("Updating change request references")
			
			// Update change request references
			updatedRefs, unchangedRefs, referencesUpdated, err := metadata.UpdateAllChangeRequestReferences(root, hashMap, fs)
			if err != nil {
				logger.Error("Failed to update change request references", zap.Error(err))
				fmt.Fprintf(os.Stderr, "Error: Failed to update change request references: %s\n", err)
			} else {
				// Print summary of reference updates
				if len(updatedRefs) > 0 {
					fmt.Println("📋 Updated change request references:")
					for _, file := range updatedRefs {
						fmt.Printf("   ✅ %s\n", file)
					}
					fmt.Printf("   📊 Total references updated: %d\n", referencesUpdated)
				}
				
				if len(updatedRefs) > 0 || len(unchangedRefs) > 0 {
					logger.Debug("Processing of change requests complete", 
						zap.Int("total", len(updatedRefs) + len(unchangedRefs)), 
						zap.Int("updated", len(updatedRefs)), 
						zap.Int("unchanged", len(unchangedRefs)),
						zap.Int("references_updated", referencesUpdated))
					
					fmt.Printf("✨ Processed %d change request files (%d updated, %d unchanged)\n", 
						len(updatedRefs) + len(unchangedRefs),
						len(updatedRefs),
						len(unchangedRefs))
				} else {
					logger.Debug("No change requests were processed")
					fmt.Println("ℹ️ No change requests needed updating")
				}
			}
		} else if skipReferences {
			logger.Debug("Skipping change request reference updates")
			fmt.Println("ℹ️ Skipped change request reference updates (--skip-references flag used)")
		} else if len(hashMap) == 0 {
			logger.Debug("No content changes detected, skipping reference updates")
			fmt.Println("ℹ️ No content changes detected, skipping reference updates")
		}
		
		fmt.Printf("✨ Processed %d user story files (%d updated, %d unchanged)\n", 
			len(updatedFiles) + len(unchangedFiles),
			len(updatedFiles),
			len(unchangedFiles))
	},
}

func init() {
	rootCmd.AddCommand(updateUserStoriesCmd)
	
	// Add flags
	updateUserStoriesCmd.Flags().Bool("skip-references", false, "Skip updating references in change request files")
	updateUserStoriesCmd.Flags().Bool("debug", false, "Enable debug mode with detailed logging")
	
	// Hidden flag for testing
	updateUserStoriesCmd.Flags().String("test-root", "", "Test root directory (for testing only)")
	updateUserStoriesCmd.Flags().MarkHidden("test-root")
}

// For testing
func resetUpdateUserStoriesCmd() {
	updateUserStoriesCmd = &cobra.Command{
		Use:   "update user-stories metadata",
		Short: "Update metadata in user story markdown files",
		Long:  `Update metadata in user story markdown files.`,
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	// Reinitialize the command with flags
	rootCmd.AddCommand(updateUserStoriesCmd)
	
	// Add flags
	updateUserStoriesCmd.Flags().Bool("skip-references", false, "Skip updating references in change request files")
	updateUserStoriesCmd.Flags().Bool("debug", false, "Enable debug mode with detailed logging")
	
	// Hidden flag for testing
	updateUserStoriesCmd.Flags().String("test-root", "", "Test root directory (for testing only)")
	updateUserStoriesCmd.Flags().MarkHidden("test-root")
} 
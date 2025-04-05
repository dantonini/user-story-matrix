// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/metadata"
	"github.com/user-story-matrix/usm/internal/ui/styles"
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

Directories like node_modules, .git, dist, build, vendor, tmp, .cache, and .github are automatically skipped.

The command preserves original creation dates if they exist, and only updates last_updated dates
when content has actually changed, making it safe to run as part of automated workflows.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Updating user story metadata")
		
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
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		
		// Initialize the file system
		fs := io.NewOSFileSystem()
		
		// Check for the --test-root flag (only used in tests)
		var userStoriesDir string
		testRoot, err := cmd.Flags().GetString("test-root")
		if err != nil {
			return fmt.Errorf("failed to get test-root flag: %w", err)
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
		
		// Verify user stories directory exists
		if !fs.Exists(userStoriesDir) {
			return fmt.Errorf("user stories directory not found: %s", userStoriesDir)
		}
		
		logger.Debug("Scanning for user stories", 
			zap.String("dir", userStoriesDir),
			zap.String("root", root))
		
		// Update all user story metadata
		updatedFiles, unchangedFiles, hashMap, err := metadata.UpdateAllUserStoryMetadata(userStoriesDir, root, fs)
		if err != nil {
			return fmt.Errorf("failed to update user story metadata: %w", err)
		}
		
		// Print summary of user story updates
		if len(updatedFiles) > 0 {
			fmt.Println("📋 Updated user story metadata:")
			// Group files by directory for better readability
			printGroupedFiles(updatedFiles, "  ")
		} else {
			fmt.Println("📋 No user story files needed updating")
		}
		
		if debug && len(unchangedFiles) > 0 {
			fmt.Println("📋 Unchanged user stories:")
			printGroupedFiles(unchangedFiles, "  ")
		}
		
		logger.Debug("Processing of user stories complete", 
			zap.Int("total", len(updatedFiles) + len(unchangedFiles)), 
			zap.Int("updated", len(updatedFiles)), 
			zap.Int("unchanged", len(unchangedFiles)))
		
		// If references shouldn't be skipped and we have content changes, update references
		updatedRefs := []string{}
		unchangedRefs := []string{}
		referencesUpdated := 0
		var mismatchedReferences []metadata.MismatchedReference
		
		if !skipReferences && len(hashMap) > 0 {
			// Only update references if there are actually content changes (not just metadata changes)
			changedHashMap := metadata.FilterChangedContent(hashMap)
			
			if len(changedHashMap) > 0 {
				logger.Debug("Updating change request references",
					zap.Int("changed_files", len(changedHashMap)))
				fmt.Println("🔄 Updating references in change requests...")
				
				// Update change request references
				updatedRefs, unchangedRefs, referencesUpdated, mismatchedReferences, err = metadata.UpdateAllChangeRequestReferences(root, changedHashMap, fs)
				if err != nil {
					return fmt.Errorf("failed to update change request references: %w", err)
				}
				
				// Print mismatched references with nice formatting
				if len(mismatchedReferences) > 0 {
					printMismatchedReferences(mismatchedReferences)
				}
				
				// Print summary of reference updates
				if len(updatedRefs) > 0 {
					fmt.Println("✅ Updated references in these change requests:")
					printGroupedFiles(updatedRefs, "  ")
					fmt.Printf("   📊 Total references updated: %d\n", referencesUpdated)
				} else {
					fmt.Println("ℹ️ No change requests needed reference updates")
				}
			} else {
				logger.Debug("No content changes detected, skipping reference updates")
				fmt.Println("ℹ️ No content changes detected, skipping reference updates")
			}
		} else if skipReferences {
			logger.Debug("Skipping change request reference updates")
			fmt.Println("ℹ️ Skipped change request reference updates (--skip-references flag used)")
		}
		
		// Print final summary
		fmt.Println("\n✨ Summary:")
		fmt.Printf("   User stories: %d processed (%d updated, %d unchanged)\n", 
			len(updatedFiles) + len(unchangedFiles),
			len(updatedFiles),
			len(unchangedFiles))
		
		if !skipReferences {
			fmt.Printf("   Change requests: %d processed (%d updated, %d unchanged, %d references updated)\n", 
				len(updatedRefs) + len(unchangedRefs),
				len(updatedRefs),
				len(unchangedRefs),
				referencesUpdated)
		}
		
		return nil
	},
}

// printMismatchedReferences prints a nicely formatted list of mismatched references
func printMismatchedReferences(mismatchedRefs []metadata.MismatchedReference) {
	if len(mismatchedRefs) == 0 {
		return
	}
	
	// Initialize UI styles
	s := styles.DefaultStyles()
	
	// Group mismatched references by file path
	mismatchesByFile := make(map[string]int)
	for _, ref := range mismatchedRefs {
		mismatchesByFile[ref.FilePath]++
	}
	
	// Create header with warning
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	warningHeader := warningStyle.Render("⚠️  Hash Mismatch Detected")
	
	fmt.Println("\n" + warningHeader)
	fmt.Println(s.Normal.Render("Some user story hashes in change requests don't match the expected values:"))
	
	// Show a count by file
	for filePath, count := range mismatchesByFile {
		fileName := filepath.Base(filePath)
		dirPath := filepath.Dir(filePath)
		
		fileInfo := fmt.Sprintf("%s (%d %s)", 
			fileName, 
			count, 
			pluralize("reference", count))
		
		fmt.Printf("  %s in %s\n", 
			s.Error.Render(fileInfo),
			s.Normal.Render(dirPath))
	}
	
	// Explanation of what this means
	fmt.Println()
	fmt.Println(s.Subtle.Render("This usually happens when:"))
	fmt.Println(s.Subtle.Render("• A change request was created with an user story"))
	fmt.Println(s.Subtle.Render("• The user story was later modified without updating the change request"))
	fmt.Println()
	fmt.Println(s.Normal.Render("All references have been updated to the current hash values."))
	fmt.Println(s.Normal.Render("You may want to review the updated change requests to ensure they're still valid."))
	fmt.Println()
}

// pluralize returns a pluralized version of a word based on count
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

// printGroupedFiles prints files grouped by their directory for better readability
func printGroupedFiles(files []string, indent string) {
	if len(files) == 0 {
		return
	}
	
	// Group files by directory
	filesByDir := make(map[string][]string)
	for _, file := range files {
		dir := filepath.Dir(file)
		filesByDir[dir] = append(filesByDir[dir], filepath.Base(file))
	}
	
	// Print each directory with its files
	for dir, fileList := range filesByDir {
		fmt.Printf("%s📁 %s/\n", indent, dir)
		for _, file := range fileList {
			fmt.Printf("%s  • %s\n", indent, file)
		}
	}
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
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
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
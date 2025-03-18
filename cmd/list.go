package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/utils"
)

var (
	// Directory to list user stories from
	fromDir string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources",
	Long:  `List resources like user stories or change requests.`,
}

// listUserStoriesCmd represents the list user-stories command
var listUserStoriesCmd = &cobra.Command{
	Use:   "user-stories",
	Short: "List all user stories",
	Long: `List all user stories in the specified directory or in the default directory (docs/user-stories).

Example:
  usm list user-stories
  usm list user-stories --from docs/user-stories/my-feature
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create filesystem and IO interfaces
		fs := io.NewOSFileSystem()
		terminal := io.NewTerminalIO()
		
		// Get the target directory
		targetDir := "docs/user-stories"
		if fromDir != "" {
			targetDir = fromDir
		}
		
		// Check if the directory exists
		if !fs.Exists(targetDir) {
			terminal.PrintError(fmt.Sprintf("Directory not found: %s", targetDir))
			return
		}
		
		// Collect all user stories
		var userStories []models.UserStory
		
		err := fs.WalkDir(targetDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			
			// Skip directories
			if d.IsDir() {
				return nil
			}
			
			// Skip non-markdown files
			if filepath.Ext(path) != ".md" {
				return nil
			}
			
			// Read the file
			content, err := fs.ReadFile(path)
			if err != nil {
				logger.Debug("Failed to read file: " + err.Error())
				return nil
			}
			
			// Parse the user story
			userStory, err := models.LoadUserStoryFromFile(path, content)
			if err != nil {
				logger.Debug("Failed to parse user story: " + err.Error())
				return nil
			}
			
			userStories = append(userStories, userStory)
			return nil
		})
		
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to walk directory: %s", err))
			return
		}
		
		// Check if any user stories were found
		if len(userStories) == 0 {
			terminal.Print(fmt.Sprintf("No user stories found in: %s", targetDir))
			return
		}
		
		// Format and print the table
		headers, rows := utils.FormatUserStoryTable(userStories)
		terminal.PrintTable(headers, rows)
		
		// Print summary
		terminal.Print(fmt.Sprintf("\nTotal: %d user stories", len(userStories)))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	
	// Add user-stories subcommand
	listCmd.AddCommand(listUserStoriesCmd)
	
	// Add flags
	listUserStoriesCmd.Flags().StringVar(&fromDir, "from", "", "Directory to list user stories from (default is docs/user-stories)")
} 
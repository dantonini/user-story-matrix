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
	// Directory to read user stories from
	fromUserStoriesDir string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new change requests",
	Long:  `Create a new change request based on existing user stories.`,
}

// createChangeRequestCmd represents the create change-request command
var createChangeRequestCmd = &cobra.Command{
	Use:   "change-request",
	Short: "Create a new change request",
	Long: `Create a new change request based on existing user stories.

The command will show a list of available user stories and allow you to select one or more.
The selected user stories will be included in the change request.

Example:
  usm create change-request
  usm create change-request --from docs/user-stories/my-feature
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create filesystem and IO interfaces
		fs := io.NewOSFileSystem()
		terminal := io.NewTerminalIO()
		
		// Get the source directory for user stories
		userStoriesDir := "docs/user-stories"
		if fromUserStoriesDir != "" {
			userStoriesDir = fromUserStoriesDir
		}
		
		// Check if the source directory exists
		if !fs.Exists(userStoriesDir) {
			terminal.PrintError(fmt.Sprintf("Directory not found: %s", userStoriesDir))
			return
		}
		
		// Collect all user stories
		var userStories []models.UserStory
		
		err := fs.WalkDir(userStoriesDir, func(path string, d os.DirEntry, err error) error {
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
			terminal.PrintError(fmt.Sprintf("No user stories found in: %s", userStoriesDir))
			return
		}
		
		// Print available user stories
		terminal.Print("Available user stories:")
		
		// Create options for selection
		options := make([]string, len(userStories))
		for i, story := range userStories {
			options[i] = utils.FormatUserStoryListItem(story, i)
		}
		
		// Ask the user to select one or more user stories
		selected, err := terminal.MultiSelect("Select user stories for the change request:", options)
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to select user stories: %s", err))
			return
		}
		
		// Check if any user stories were selected
		if len(selected) == 0 {
			terminal.PrintError("No user stories selected")
			return
		}
		
		// Ask for the change request name
		name, err := terminal.Prompt("Enter the change request name:")
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to read input: %s", err))
			return
		}
		
		if name == "" {
			terminal.PrintError("Name cannot be empty")
			return
		}
		
		// Create references to the selected user stories
		references := make([]models.UserStoryReference, len(selected))
		for i, idx := range selected {
			us := userStories[idx]
			references[i] = models.UserStoryReference{
				Title:       us.Title,
				FilePath:    us.FilePath,
				ContentHash: us.ContentHash,
			}
		}
		
		// Generate the change request template
		template := models.GenerateChangeRequestTemplate(name, references)
		
		// Ensure the change requests directory exists
		changeRequestsDir := "docs/changes-request"
		if !fs.Exists(changeRequestsDir) {
			if err := fs.MkdirAll(changeRequestsDir, 0755); err != nil {
				terminal.PrintError(fmt.Sprintf("Failed to create directory: %s", err))
				return
			}
		}
		
		// Generate the filename
		filename := models.GenerateChangeRequestFilename(name)
		
		// Generate the file path
		filePath := filepath.Join(changeRequestsDir, filename)
		
		// Check if the file already exists
		if fs.Exists(filePath) {
			terminal.PrintError(fmt.Sprintf("File already exists: %s", filePath))
			return
		}
		
		// Save the file
		if err := fs.WriteFile(filePath, []byte(template), 0644); err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to write file: %s", err))
			return
		}
		
		// Success message
		terminal.PrintSuccess(fmt.Sprintf("Change request created: %s", filePath))
		
		// Show next steps
		promptInstruction := models.GetPromptInstruction(filePath, len(selected))
		terminal.Print("\nNext steps:")
		terminal.Print("The change request file has been created. You can now edit it with the following prompt:")
		terminal.Print("\n" + promptInstruction + "\n")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	
	// Add change-request subcommand
	createCmd.AddCommand(createChangeRequestCmd)
	
	// Add flags
	createChangeRequestCmd.Flags().StringVar(&fromUserStoriesDir, "from", "", "Directory to read user stories from (default is docs/user-stories)")
} 
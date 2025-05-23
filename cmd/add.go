// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/implementation"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui"
	"github.com/user-story-matrix/usm/internal/ui/contracts"
)

var (
	// Directory to save the user story
	intoDir string
	
	// Directory to read user stories from for change request creation
	fromUserStoriesDir string
	
	// Show all user stories, including implemented ones
	showAll bool
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new item",
	Long:  `Add a new item like user story or change request.`,
}

// addUserStoryCmd represents the add user-story command
var addUserStoryCmd = &cobra.Command{
	Use:   "user-story",
	Short: "Add a new user story",
	Long: `Add a new user story in markdown format.

The story will be saved in the specified directory (using --into)
or in the default directory (docs/user-stories) if not specified.

Example:
  usm add user-story
  usm add user-story --into docs/user-stories/my-feature
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create filesystem and IO interfaces
		fs := io.NewOSFileSystem()
		terminal := io.NewTerminalIO()
		
		// Get the target directory
		targetDir := "docs/user-stories"
		if intoDir != "" {
			targetDir = intoDir
		}
		
		// Ensure the target directory exists
		if !fs.Exists(targetDir) {
			if err := fs.MkdirAll(targetDir, 0755); err != nil {
				terminal.PrintError(fmt.Sprintf("Failed to create directory: %s", err))
				return
			}
		}
		
		// Get entries from the target directory to determine next sequential number
		entries, err := fs.ReadDir(targetDir)
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to read directory: %s", err))
			return
		}
		
		// Get the next sequential number
		sequentialNumber := models.GetNextSequentialNumber(entries)
		
		// Create an empty user story with current time
		us := models.UserStory{
			CreatedAt: time.Now(),
			LastUpdated: time.Now(),
		}
		
		// Create the form
		formModel := ui.CreateUserStoryFormWithLLM(us)
		
		// Run the form
		p := tea.NewProgram(formModel)
		result, err := p.Run()
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Error running form: %s", err))
			return
		}
		
		// Process the form result
		userStory, shouldContinue := processFormResult(result, terminal)
		if !shouldContinue {
			return
		}
		
		// Generate the filename
		filename := models.GenerateFilename(sequentialNumber, userStory.Title)
		
		// Generate the file path
		filePath := filepath.Join(targetDir, filename)
		
		// Check if the file already exists
		if fs.Exists(filePath) {
			terminal.PrintError(fmt.Sprintf("File already exists: %s", filePath))
			return
		}
		
		// Set the file path in the user story
		relativePath, err := filepath.Rel(filepath.Dir(os.Args[0]), filePath)
		if err != nil {
			// If we can't get the relative path, use the absolute path
			relativePath = filePath
		}
		
		// Update the file path
		userStory.FilePath = relativePath
		
		// Save the file
		if err := fs.WriteFile(filePath, []byte(userStory.Content), 0644); err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to write file: %s", err))
			return
		}
		
		// Success message
		terminal.PrintSuccess(fmt.Sprintf("User story created: %s", filePath))
		
		logger.Debug("User story created with sequential number: " + sequentialNumber)
	},
}

// addChangeRequestCmd represents the add change-request command
var addChangeRequestCmd = &cobra.Command{
	Use:   "change-request",
	Short: "Add a new change request",
	Long: `Add a new change request based on existing user stories.

The command will show a list of available user stories and allow you to select one or more.
The selected user stories will be included in the change request.

Example:
  usm add change-request
  usm add change-request --from docs/user-stories/my-feature
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

			// Check if the user story is implemented
			if err := implementation.UpdateImplementationStatus(&userStory, fs); err != nil {
				logger.Debug("Failed to check implementation status: " + err.Error())
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

		// Create a selection UI with the showAll flag
		selectionUI := ui.CurrentNewSelectionUI(userStories, showAll)

		// Create a program with more options
		p := tea.NewProgram(selectionUI,
			// Add option to capture the terminal window size on startup
			tea.WithAltScreen(),
			// Send an initial window size event to ensure the UI is properly sized
			tea.WithMouseCellMotion(),
		)

		// Run the program
		model, err := p.Run()
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to run selection UI: %s", err))
			return
		}

		// Get the selected stories
		selAdapter, ok := model.(*ui.SelectionAdapter)
		if !ok {
			terminal.PrintError("Error: could not get selection result")
			return
		}
		selected := selAdapter.GetSelected()

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
		if err := fs.WriteFile(filePath, []byte(template), 0600); err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to write file: %s", err))
			return
		}

		// Success message
		terminal.PrintSuccess(fmt.Sprintf("Change request created: %s", filePath))

		// Show next steps
		nextStepsInstruction := models.GetNextStepsInstruction(filePath)
		terminal.Print("\nNext steps:")
		terminal.PrintProgress(nextStepsInstruction)
	},
}

// processFormResult handles the form result and returns the user story and a boolean indicating
// if the process should continue. This function is extracted to make it more testable.
func processFormResult(result interface{}, terminal io.UserOutput) (models.UserStory, bool) {
	// Handle different form types
	if formResult, ok := result.(contracts.UserStorySubmitter); ok {
		// New form implementation using the explicit interface
		userStory := formResult.GetUserStory()
		confirmSubmission := formResult.GetConfirmSubmission()
		
		if !confirmSubmission {
			terminal.Print("User story empty, creation cancelled")
			return models.UserStory{}, false
		}
		
		return userStory, true
	}
	
	// Fallback to type assertions for backward compatibility
	var userStory models.UserStory
	var confirmSubmission bool
	
	if getter, ok := result.(interface{ GetUserStory() models.UserStory }); ok {
		userStory = getter.GetUserStory()
	} else {
		terminal.PrintError("Error: could not get user story from form")
		return models.UserStory{}, false
	}
	
	if confirmGetter, ok := result.(interface{ GetConfirmSubmission() bool }); ok {
		confirmSubmission = confirmGetter.GetConfirmSubmission()
	} else {
		terminal.PrintError("Error: could not determine if submission was confirmed")
		return models.UserStory{}, false
	}
	
	if !confirmSubmission {
		terminal.Print("User story empty, creation cancelled")
		return models.UserStory{}, false
	}
	
	return userStory, true
}

func init() {
	rootCmd.AddCommand(addCmd)
	
	// Add user-story subcommand
	addCmd.AddCommand(addUserStoryCmd)
	
	// Add change-request subcommand
	addCmd.AddCommand(addChangeRequestCmd)
	
	// Add flags for user-story command
	addUserStoryCmd.Flags().StringVar(&intoDir, "into", "", "Directory to save the user story (default is docs/user-stories)")
	
	// Add flags for change-request command
	addChangeRequestCmd.Flags().StringVar(&fromUserStoriesDir, "from", "", "Directory to read user stories from (default is docs/user-stories)")
	addChangeRequestCmd.Flags().BoolVar(&showAll, "show-all", false, "Show all user stories, including implemented ones")
	
	// Register the new selection UI implementation
	ui.RegisterNewSelectionUIMaker()
} 
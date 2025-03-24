// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/implementation"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui"
)

// Program interface for testing
type program interface {
	Run() (tea.Model, error)
}

// Default tea.Program wrapper
type teaProgram struct {
	*tea.Program
}

func (p *teaProgram) Run() (tea.Model, error) {
	return p.Program.Run()
}

// Program creator type
type programCreator func(m tea.Model, opts ...tea.ProgramOption) program

var (
	// Directory to read user stories from
	fromUserStoriesDir string
	// Show all user stories, including implemented ones
	showAll bool
	// Terminal interface for testing
	terminal interface {
		io.UserInput
		io.UserOutput
	}
	// Program creator for testing
	newProgram programCreator = func(m tea.Model, opts ...tea.ProgramOption) program {
		return &teaProgram{tea.NewProgram(m, opts...)}
	}
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
		p := newProgram(selectionUI, 
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
		selected := model.(*ui.SelectionUI).GetSelected()
		
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
	createChangeRequestCmd.Flags().BoolVar(&showAll, "show-all", false, "Show all user stories, including implemented ones")

	// Initialize terminal
	terminal = io.NewTerminalIO()
} 
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
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui"
	"github.com/user-story-matrix/usm/internal/ui/contracts"
)

var (
	// Directory to save the user story
	intoDir string
	
	// Enable LLM processing for user stories (true by default)
	enableLLM bool
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user story",
	Long:  `Add a new user story in markdown format.`,
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
  usm add user-story --no-llm  (disable LLM processing)
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
		
		// Create the form, using the LLM-enabled version if available and run it
		formModel := ui.CreateUserStoryFormWithLLM(us, enableLLM)
		
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
	
	// Legacy form
	if ptrForm, ok := result.(*io.UserStoryForm); ok {
		if !ptrForm.ConfirmSubmission {
			terminal.Print("User story empty, creation cancelled")
			return models.UserStory{}, false
		}
		return ptrForm.GetUserStory(), true
	}
	
	// Fallback to type assertions for backward compatibility
	// This branch should be removed once all forms implement the contracts.FormResult interface
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
	
	// Add flags
	addUserStoryCmd.Flags().StringVar(&intoDir, "into", "", "Directory to save the user story (default is docs/user-stories)")
	addUserStoryCmd.Flags().BoolVar(&enableLLM, "no-llm", false, "Disable LLM processing of pasted text")
	
	// Invert the flag meaning (--no-llm sets enableLLM to false)
	enableLLM = !enableLLM
} 
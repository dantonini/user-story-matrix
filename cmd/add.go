package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
)

var (
	// Directory to save the user story
	intoDir string
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
		
		// Prompt the user for the title
		title, err := terminal.Prompt("Enter the user story title:")
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to read input: %s", err))
			return
		}
		
		if title == "" {
			terminal.PrintError("Title cannot be empty")
			return
		}
		
		// Generate the filename
		filename := models.GenerateFilename(sequentialNumber, title)
		
		// Generate the file path
		filePath := filepath.Join(targetDir, filename)
		
		// Check if the file already exists
		if fs.Exists(filePath) {
			terminal.PrintError(fmt.Sprintf("File already exists: %s", filePath))
			return
		}
		
		// Generate the template
		template := models.GenerateUserStoryTemplate(title)
		
		// Fill in the remaining template fields
		relativePath, err := filepath.Rel(filepath.Dir(os.Args[0]), filePath)
		if err != nil {
			// If we can't get the relative path, use the absolute path
			relativePath = filePath
		}
		finalTemplate := models.FinalizeUserStoryTemplate(template, relativePath)
		
		// Save the file
		if err := fs.WriteFile(filePath, []byte(finalTemplate), 0644); err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to write file: %s", err))
			return
		}
		
		// Success message
		terminal.PrintSuccess(fmt.Sprintf("User story created: %s", filePath))
		
		logger.Debug("User story created with sequential number: " + sequentialNumber)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	
	// Add user-story subcommand
	addCmd.AddCommand(addUserStoryCmd)
	
	// Add flags
	addUserStoryCmd.Flags().StringVar(&intoDir, "into", "", "Directory to save the user story (default is docs/user-stories)")
} 
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
)

// recapCmd represents the recap command
var recapCmd = &cobra.Command{
	Use:   "recap",
	Short: "Recap what you have done",
	Long: `Recap what you have done by displaying incomplete change requests.

The command looks for "incomplete" change requests in the docs/changes-requests directory.
An incomplete change request is one that has a blueprint file but no implementation file.

If no incomplete change requests are found, a congratulation message is displayed.
If multiple incomplete change requests are found, you can select one from a list.

Example:
  usm recap
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create filesystem and IO interfaces
		fs := io.NewOSFileSystem()
		terminal := io.NewTerminalIO()

		// Find incomplete change requests
		incompleteChangeRequests, err := findIncompleteChangeRequests(fs)
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to find incomplete change requests: %s", err))
			return
		}

		// Handle based on number of incomplete change requests found
		switch len(incompleteChangeRequests) {
		case 0:
			// No incomplete change requests found
			displayCongratulationMessage(terminal)
		case 1:
			// Exactly one incomplete change request found, use it directly
			displayRecapMessage(terminal, incompleteChangeRequests[0])
		default:
			// Multiple incomplete change requests found, let the user select one
			options := make([]string, len(incompleteChangeRequests))
			for i, cr := range incompleteChangeRequests {
				options[i] = formatChangeRequestDescription(cr)
			}

			// Prompt the user to select a change request
			selectedIndex, err := terminal.Select("Select a change request to recap:", options)
			if err != nil {
				terminal.PrintError(fmt.Sprintf("Failed to select change request: %s", err))
				return
			}

			// Display the recap message for the selected change request
			displayRecapMessage(terminal, incompleteChangeRequests[selectedIndex])
		}
	},
}

// findIncompleteChangeRequests finds all change requests that have a blueprint file
// but no implementation file
func findIncompleteChangeRequests(fs io.FileSystem) ([]models.ChangeRequest, error) {
	var incompleteChangeRequests []models.ChangeRequest

	// Define the change requests directory
	changeRequestsDir := "docs/changes-request"

	// Check if the directory exists
	if !fs.Exists(changeRequestsDir) {
		return incompleteChangeRequests, fmt.Errorf("change requests directory not found: %s", changeRequestsDir)
	}

	// Get all files in the directory
	entries, err := fs.ReadDir(changeRequestsDir)
	if err != nil {
		return incompleteChangeRequests, fmt.Errorf("failed to read directory: %s", err)
	}

	// Iterate through the files to find blueprint files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".blueprint.md") {
			continue
		}

		// Load the blueprint file
		blueprintPath := filepath.Join(changeRequestsDir, filename)
		content, err := fs.ReadFile(blueprintPath)
		if err != nil {
			logger.Debug(fmt.Sprintf("Failed to read blueprint file %s: %s", blueprintPath, err))
			continue
		}

		// Parse the change request
		changeRequest, err := models.LoadChangeRequestFromContent(blueprintPath, content)
		if err != nil {
			logger.Debug(fmt.Sprintf("Failed to parse change request from %s: %s", blueprintPath, err))
			continue
		}

		// Check if there is a corresponding implementation file
		implementationFilename := strings.Replace(filename, ".blueprint.md", ".implementation.md", 1)
		implementationPath := filepath.Join(changeRequestsDir, implementationFilename)
		
		if !fs.Exists(implementationPath) {
			// This is an incomplete change request (has blueprint but no implementation)
			incompleteChangeRequests = append(incompleteChangeRequests, changeRequest)
		}
	}

	return incompleteChangeRequests, nil
}

// formatChangeRequestDescription formats a change request for display in selection list
func formatChangeRequestDescription(cr models.ChangeRequest) string {
	// Format the creation date
	createdAt := cr.CreatedAt.Format("2006-01-02 15:04:05")
	
	// Count the number of user stories
	userStoryCount := len(cr.UserStories)
	
	// Create a description
	return fmt.Sprintf("%s (Created: %s, Stories: %d)", cr.Name, createdAt, userStoryCount)
}

// displayCongratulationMessage displays a fancy congratulation message when no
// incomplete change requests are found
func displayCongratulationMessage(term io.UserOutput) {
	message := `
🎉 Congratulations! 🎉

All change requests have been completed.
There are no pending implementation files to create.
`
	term.PrintSuccess(message)
}

// displayRecapMessage displays the recap message for the selected change request
func displayRecapMessage(term io.UserOutput, cr models.ChangeRequest) {
	// Get the base filename without the extension
	baseFilename := filepath.Base(cr.FilePath)
	baseFilename = strings.TrimSuffix(baseFilename, ".blueprint.md")
	
	// Create the implementation filename
	implementationFilename := fmt.Sprintf("docs/changes-request/%s.implementation.md", baseFilename)
	
	// Display the message
	message := fmt.Sprintf("Recap what you did in a file in %s", implementationFilename)
	term.Print(message)
}

func init() {
	rootCmd.AddCommand(recapCmd)
} 
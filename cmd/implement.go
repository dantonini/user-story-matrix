// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/changerequest"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
)

// implementCmd represents the implement command
var implementCmd = &cobra.Command{
	Use:   "implement",
	Short: "Implement a change request",
	Long: `Implement a change request using the blueprint file.

The command looks for "incomplete" change requests in the docs/changes-requests directory.
An incomplete change request is one that has a blueprint file but no implementation file.

If no incomplete change requests are found, a sad message is displayed.
If multiple incomplete change requests are found, you can select one from a list.

Example:
  usm implement
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create filesystem and IO interfaces
		fs := io.NewOSFileSystem()
		terminal := io.NewTerminalIO()

		// Find incomplete change requests
		incompleteChangeRequests, err := changerequest.FindIncomplete(fs)
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to find incomplete change requests: %s", err))
			return
		}

		// Handle based on number of incomplete change requests found
		switch len(incompleteChangeRequests) {
		case 0:
			// No incomplete change requests found
			displayNoChangeRequestsMessage(terminal)
		case 1:
			// Exactly one incomplete change request found, use it directly
			displayImplementationMessage(terminal, incompleteChangeRequests[0])
		default:
			// Multiple incomplete change requests found, let the user select one
			options := make([]string, len(incompleteChangeRequests))
			for i, cr := range incompleteChangeRequests {
				options[i] = changerequest.FormatDescription(cr)
			}

			// Prompt the user to select a change request
			selectedIndex, err := terminal.Select("Select a change request to implement:", options)
			if err != nil {
				terminal.PrintError(fmt.Sprintf("Failed to select change request: %s", err))
				return
			}

			// Display the implementation message for the selected change request
			displayImplementationMessage(terminal, incompleteChangeRequests[selectedIndex])
		}
	},
}

// displayNoChangeRequestsMessage displays a message when no incomplete change requests are found
func displayNoChangeRequestsMessage(term io.UserOutput) {
	message := `
😢 No change requests to implement.

All change requests have been completed or there are no change requests at all.
You can create a new change request using the 'usm create change-request' command.
`
	term.Print(message)
}

// displayImplementationMessage displays the implementation message for the selected change request
func displayImplementationMessage(term io.UserOutput, cr models.ChangeRequest) {

	
	// Create the message
	message := fmt.Sprintf(
		"Read the blueprint file in %s:\n- Read all the mentioned user stories.\n- Validate the blueprint against the code base.\n- Proceed with the implementation.\n",
		cr.FilePath,
	)
	
	// Display the message
	term.Print(message)
}

func init() {
	rootCmd.AddCommand(implementCmd)
	logger.Debug("Implement command added to root command")
} 
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
)

const slackWebhookURL = "https://hooks.slack.com/services/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"

// askCmd represents the ask command
var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Submit feature requests and suggestions",
	Long:  `Submit feature requests or suggestions to the CLI developers.`,
}

// askFeatureCmd represents the feature command
var askFeatureCmd = &cobra.Command{
	Use:   "feature",
	Short: "Submit a feature request",
	Long: `Submit a feature request to the CLI developers.

The command provides an interactive form to enter details about the feature request.
You can navigate between fields using Tab and Shift+Tab. Your draft will be saved
automatically, so you can resume it later if you interrupt the process.

Example:
  usm ask feature
`,
	Run: func(cmd *cobra.Command, args []string) {
		fs := io.NewOSFileSystem()
		terminal := io.NewTerminalIO()
		draftManager := io.NewDraftManager(fs)
		
		// Load any existing draft
		fr, err := draftManager.LoadDraft()
		if err != nil {
			logger.Debug("Failed to load draft: " + err.Error())
			fr = models.NewFeatureRequest()
		}
		
		// Create and configure the form
		form := io.NewFeatureForm(fr)
		
		// Setup signal handling to save draft on interrupt
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			draftRequest := form.SaveDraft()
			if err := draftManager.SaveDraft(draftRequest); err != nil {
				logger.Error("Failed to save draft: " + err.Error())
			} else {
				terminal.Print("\nDraft saved. You can resume later with 'usm ask feature'.")
			}
			os.Exit(0)
		}()
		
		// Run the form UI
		p := tea.NewProgram(form)
		result, err := p.Run()
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Error running form: %s", err))
			return
		}
		
		finalForm, ok := result.(*io.FeatureForm)
		if !ok {
			terminal.PrintError("Error: could not get form result")
			return
		}
		
		finalRequest := finalForm.GetFeatureRequest()
		
		// If the form was completed and confirmed, submit the request
		if finalRequest.IsComplete() {
			terminal.Print("Submitting feature request...")
			
			slackClient := io.NewSlackClient(slackWebhookURL)
			if err := slackClient.SendFeatureRequest(finalRequest); err != nil {
				terminal.PrintError(fmt.Sprintf("Failed to send feature request: %s", err))
				return
			}
			
			// Delete the draft after successful submission
			if err := draftManager.DeleteDraft(); err != nil {
				logger.Debug("Failed to delete draft: " + err.Error())
			}
			
			terminal.PrintSuccess("Feature request submitted successfully!")
		} else {
			// Save the draft for later
			if err := draftManager.SaveDraft(finalRequest); err != nil {
				terminal.PrintError(fmt.Sprintf("Failed to save draft: %s", err))
				return
			}
			
			terminal.Print("Feature request saved as draft. You can resume later with 'usm ask feature'.")
		}
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
	askCmd.AddCommand(askFeatureCmd)
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/llm"
	"github.com/user-story-matrix/usm/internal/logger"
)

// settingsCmd represents the settings command
var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Manage application settings",
	Long:  `Manage application settings including API keys for external services.`,
}

// settingsApiKeyCmd represents the settings api-key command
var settingsApiKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "Manage API keys",
	Long:  `Manage API keys for external services like OpenAI.`,
}

// settingsApiKeyOpenaiCmd represents the settings api-key openai command
var settingsApiKeyOpenaiCmd = &cobra.Command{
	Use:   "openai [key]",
	Short: "Configure OpenAI API key",
	Long: `Configure or view the OpenAI API key.

If no key is provided, the current key will be displayed (if configured).
If a key is provided, it will be validated and stored if valid.

Example:
  usm settings api-key openai sk-YOURAPIKEY
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create file system and terminal IO
		fs := io.NewOSFileSystem()
		terminal := io.NewTerminalIO()
		
		// Create config manager
		configManager := llm.NewConfigManager(fs)
		err := configManager.LoadConfig()
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to load configuration: %s", err))
			return
		}
		
		// If no arguments, show current key status
		if len(args) == 0 {
			if configManager.IsOpenAIKeyConfigured() {
				// Only show a masked version of the key for security
				key := configManager.GetOpenAIKey()
				masked := maskAPIKey(key)
				terminal.PrintSuccess(fmt.Sprintf("OpenAI API key is configured: %s", masked))
			} else {
				terminal.PrintWarning("OpenAI API key is not configured")
				terminal.Print("Use 'usm settings api-key openai YOUR-API-KEY' to configure")
			}
			return
		}
		
		// Set the new API key
		key := args[0]
		
		// Create a processor to validate the key
		processor := llm.NewOpenAIProcessor(llm.WithAPIKey(key))
		
		// Create a context with timeout for validation
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Validate the key
		terminal.Print("Validating OpenAI API key...")
		err = processor.ValidateConfiguration(ctx)
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Invalid API key: %s", err))
			return
		}
		
		// Save the key
		err = configManager.SetOpenAIKey(key, processor)
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Failed to save API key: %s", err))
			return
		}
		
		terminal.PrintSuccess("OpenAI API key configured successfully")
		logger.Debug("OpenAI API key configured")
	},
}

// maskAPIKey masks an API key for display, showing only the first 4 and last 4 characters
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	
	return key[:4] + "..." + key[len(key)-4:]
}

func init() {
	rootCmd.AddCommand(settingsCmd)
	settingsCmd.AddCommand(settingsApiKeyCmd)
	settingsApiKeyCmd.AddCommand(settingsApiKeyOpenaiCmd)
} 
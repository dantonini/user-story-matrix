// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/logger"
)

var (
	debug bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "usm",
	Short: "User Story Matrix CLI",
	Long: `User Story Matrix CLI (usm-cli) is a tool for managing user stories fully integrated with
any AI-powered coding assistant.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger based on debug flag
		if err := logger.Initialize(debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing logger: %s\n", err)
			os.Exit(1)
		}
		
		if debug {
			logger.Debug("Debug mode enabled")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	defer func() {
		err := logger.Sync()
		if err != nil {
			// We can only log this error since we're in cleanup
			fmt.Fprintf(os.Stderr, "Error syncing logger: %v\n", err)
		}
	}()
	return rootCmd.Execute()
}

func init() {
	// Add persistent flags that will be available to all commands
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode with verbose logging")
} 
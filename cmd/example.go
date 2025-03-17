package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/logger"
)

// exampleCmd represents the example command
var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "A simple example command",
	Long:  `This is a simple example command that prints "Hello, USM!"`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Example command executed")
		fmt.Println("Hello, USM!")
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
	logger.Debug("Example command added to root command")
} 
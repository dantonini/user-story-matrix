// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// extractWorkflowCmd represents the extract-workflow command
var extractWorkflowCmd = &cobra.Command{
	Use:   "extract-workflow",
	Short: "Extract standard workflow to filesystem",
	Long: `Extract the built-in standard workflow to the filesystem.
This creates template files that can be customized to create new workflows.

The workflow is extracted to the specified output directory, creating:
1. A workflow.yaml file with step metadata
2. A prompts/ subdirectory with individual prompt files for each step

Example:
  usm extract-workflow --output ./my-workflow
`,
	Run: func(cmd *cobra.Command, args []string) {
		output := io.NewTerminalIO()
		
		outputDir, err := cmd.Flags().GetString("output")
		if err != nil {
			output.PrintError(fmt.Sprintf("Error getting output directory flag: %v", err))
			return
		}

		if outputDir == "" {
			outputDir = workflow.StandardTemplateDir
		}

		if !filepath.IsAbs(outputDir) {
			var err error
			outputDir, err = filepath.Abs(outputDir)
			if err != nil {
				log.Fatalf("Failed to resolve absolute path: %v", err)
			}
		}

		fs := io.NewOSFileSystem()
		
		output.PrintProgress("Extracting standard workflow to " + outputDir)
		
		err = workflow.ExtractStandardWorkflow(fs, outputDir)
		if err != nil {
			output.PrintError("Failed to extract workflow: " + err.Error())
			return
		}
		
		output.PrintSuccess("Workflow extracted successfully to: " + outputDir)
	},
}

func init() {
	rootCmd.AddCommand(extractWorkflowCmd)

	// Add flags
	extractWorkflowCmd.Flags().StringP("output", "o", "", "Output directory for extracted workflow")
} 
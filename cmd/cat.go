// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/models"
)

// CatOptions contains the options for the cat command
type CatOptions struct {
	ShowContentHash   bool
	ChangeRequestPath string
	FilterPattern     string   // Regular expression to filter stories by title or content
	ColorOutput       bool     // Whether to use colorized output
	CompactMode       bool     // Whether to use compact output format
	ExcludeStories    []string // Stories to exclude (by title or path)
}

// UserStoryOutput defines the interface for outputting user stories
type UserStoryOutput interface {
	Print(message string)
	PrintError(message string)
	PrintSuccess(message string)
	PrintWarning(message string)
	IsDebugEnabled() bool
}

// catCmd represents the cat command
var catCmd = &cobra.Command{
	Use:   "cat [change-request-file]",
	Short: "Display content of user stories in a change request",
	Long: `Display the content of all user stories referenced in a change request.
	
The command reads the change request file and displays the content of each user story.
By default, the content hash is not displayed.

Examples:
  usm cat docs/changes-request/my-feature.md
  usm cat docs/changes-request/my-feature.md --show-content-hash
  usm cat docs/changes-request/my-feature.md --filter "auth|login"
  usm cat docs/changes-request/my-feature.md --compact`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments are provided, show usage
		if len(args) < 1 {
			cmd.Help()
			return
		}

		// Get the change request file path
		changeRequestPath := args[0]

		// Get command flags
		showContentHash, _ := cmd.Flags().GetBool("show-content-hash")
		filterPattern, _ := cmd.Flags().GetString("filter")
		colorOutput, _ := cmd.Flags().GetBool("color")
		compactMode, _ := cmd.Flags().GetBool("compact")
		excludeStories, _ := cmd.Flags().GetStringSlice("exclude")

		// Create options
		options := CatOptions{
			ShowContentHash:   showContentHash,
			ChangeRequestPath: changeRequestPath,
			FilterPattern:     filterPattern,
			ColorOutput:       colorOutput,
			CompactMode:       compactMode,
			ExcludeStories:    excludeStories,
		}

		// Create the file system
		fs := io.NewOSFileSystem()

		// Create the terminal
		terminal := io.NewTerminalIO()

		// Check if the change request file exists
		if !fs.Exists(changeRequestPath) {
			terminal.PrintError(fmt.Sprintf("Change request file not found: %s", changeRequestPath))
			return
		}

		// Read the change request
		content, err := fs.ReadFile(changeRequestPath)
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Error reading change request file: %s", err))
			return
		}

		// Parse the change request
		changeRequest, err := models.LoadChangeRequestFromContent(changeRequestPath, content)
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Error parsing change request: %s", err))
			return
		}

		// Process and print the user stories
		err = processAndPrintUserStories(fs, terminal, options, changeRequest)
		if err != nil {
			terminal.PrintError(fmt.Sprintf("Error processing user stories: %s", err))
			return
		}
	},
}

// processAndPrintUserStories processes and prints the content of user stories in a change request
func processAndPrintUserStories(fs io.FileSystem, output UserStoryOutput, options CatOptions, changeRequest models.ChangeRequest) error {
	// Compile filter regex if provided
	var filterRegex *regexp.Regexp
	var err error
	if options.FilterPattern != "" {
		filterRegex, err = regexp.Compile(options.FilterPattern)
		if err != nil {
			return fmt.Errorf("invalid filter pattern: %s", err)
		}
	}

	// Count stories
	totalStories := len(changeRequest.UserStories)
	displayedStories := 0
	skippedStories := 0
	errorStories := 0

	// Process each user story
	for _, userStory := range changeRequest.UserStories {
		// Check if the story should be excluded
		if shouldExcludeStory(userStory, options.ExcludeStories) {
			skippedStories++
			continue
		}

		// Print the file path as a markdown comment
		if options.ColorOutput {
			output.PrintSuccess(fmt.Sprintf("[//]: # (%s)\n", userStory.FilePath))
		} else {
			output.Print(fmt.Sprintf("[//]: # (%s)\n", userStory.FilePath))
		}

		// Check if the user story file exists
		if !fs.Exists(userStory.FilePath) {
			output.PrintError(fmt.Sprintf("User story file not found: %s", userStory.FilePath))
			errorStories++
			continue
		}

		// Read the user story file
		contentBytes, err := fs.ReadFile(userStory.FilePath)
		if err != nil {
			output.PrintError(fmt.Sprintf("Error reading user story file: %s", err))
			errorStories++
			continue
		}

		content := string(contentBytes)

		// Skip if content doesn't match filter pattern
		if filterRegex != nil && !filterRegex.MatchString(content) {
			// Also check if title matches
			if !filterRegex.MatchString(userStory.Title) {
				skippedStories++
				continue
			}
		}

		// Process the content
		processedContent := processUserStoryContent(content, options.ShowContentHash)

		// If compact mode is enabled, use a more condensed format
		if options.CompactMode {
			processedContent = createCompactOutput(userStory, processedContent)
		}

		// Print the content
		output.Print(processedContent)
		
		// Add a separator between user stories
		if !options.CompactMode {
			output.Print("\n---\n")
		} else {
			output.Print("\n")
		}

		displayedStories++
	}

	// Print summary if debug is enabled or filtering was applied
	if options.FilterPattern != "" || skippedStories > 0 || output.IsDebugEnabled() {
		summary := fmt.Sprintf(
			"Summary: %d of %d stories displayed, %d skipped, %d errors",
			displayedStories, totalStories, skippedStories, errorStories,
		)
		
		if options.ColorOutput {
			output.PrintWarning(summary)
		} else {
			output.Print(summary)
		}
	}

	return nil
}

// shouldExcludeStory checks if a story should be excluded based on the exclude patterns
func shouldExcludeStory(story models.UserStoryReference, excludePatterns []string) bool {
	if len(excludePatterns) == 0 {
		return false
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(story.Title, pattern) || strings.Contains(story.FilePath, pattern) {
			return true
		}
	}
	return false
}

// createCompactOutput creates a compact representation of a user story
func createCompactOutput(story models.UserStoryReference, content string) string {
	// Extract title from content
	titleRegex := regexp.MustCompile(`(?m)^# (.*)$`)
	matches := titleRegex.FindStringSubmatch(content)
	
	title := story.Title
	if len(matches) > 1 {
		title = matches[1]
	}
	
	// Extract just the content part (no frontmatter)
	contentParts := strings.Split(content, "---")
	if len(contentParts) >= 3 {
		mainContent := strings.TrimSpace(contentParts[2])
		// Keep only first paragraph after title
		paragraphs := strings.SplitN(mainContent, "\n\n", 3)
		if len(paragraphs) >= 2 {
			// Skip the title paragraph, return just the first content paragraph
			firstPara := strings.TrimSpace(paragraphs[1])
			return fmt.Sprintf("# %s\n%s", title, firstPara)
		}
		return mainContent
	}
	
	return content
}

// processUserStoryContent processes the content of a user story
// If showContentHash is false, it removes the content hash line from the frontmatter
func processUserStoryContent(content string, showContentHash bool) string {
	if !showContentHash {
		// Remove the content hash line from all metadata sections
		re := regexp.MustCompile(`(?m)^_content_hash:.*\n`)
		content = re.ReplaceAllString(content, "")
	}

	return content
}

func init() {
	rootCmd.AddCommand(catCmd)

	// Add flags
	catCmd.Flags().BoolP("show-content-hash", "c", false, "Show content hash in the output")
	catCmd.Flags().StringP("filter", "f", "", "Filter stories by regular expression pattern")
	catCmd.Flags().BoolP("color", "l", false, "Use colorized output")
	catCmd.Flags().BoolP("compact", "m", false, "Use compact output format (title and first paragraph only)")
	catCmd.Flags().StringSliceP("exclude", "e", []string{}, "Exclude stories matching these patterns (comma-separated)")
} 
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
	"github.com/user-story-matrix/usm/internal/utils"
)

// CatOptions contains the configuration options for the cat command.
// This struct centralizes all parameters that affect the command's behavior.
type CatOptions struct {
	ShowContentHash   bool     // Whether to include content hash in output
	ChangeRequestPath string   // Path to change request file
	FilterPattern     string   // Regular expression to filter stories by title or content
	ColorOutput       bool     // Whether to use colorized output
	CompactMode       bool     // Whether to use compact output format
	ExcludeStories    []string // Stories to exclude (by title or path)
}

// UserStoryOutput defines the interface for outputting user stories.
// This interface decouples the command from specific terminal implementations
// and facilitates testing.
type UserStoryOutput interface {
	Print(message string)
	PrintError(message string)
	PrintSuccess(message string)
	PrintWarning(message string)
	IsDebugEnabled() bool
}

// ProcessingResult contains statistics about the user story processing operation.
// Used to generate summaries and track operation outcomes.
type ProcessingResult struct {
	TotalStories     int
	DisplayedStories int
	SkippedStories   int
	ErrorStories     int
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
  usm cat docs/changes-request/my-feature.md --compact
  usm cat docs/changes-request/my-feature.md --exclude "auth,registration"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments are provided, show usage
		if len(args) < 1 {
			if err := cmd.Help(); err != nil {
				fmt.Println("Error displaying help:", err)
			}
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

		// Create dependencies
		fs := io.NewOSFileSystem()
		terminal := io.NewTerminalIO()

		// Execute the command
		if err := executeCatCommand(fs, terminal, options); err != nil {
			terminal.PrintError(err.Error())
		}
	},
}

// executeCatCommand orchestrates the execution of the cat command.
// It loads the change request, processes the user stories, and handles errors.
func executeCatCommand(fs io.FileSystem, output UserStoryOutput, options CatOptions) error {
	// Check if the change request file exists
	if !fs.Exists(options.ChangeRequestPath) {
		return fmt.Errorf("change request file not found: %s", options.ChangeRequestPath)
	}

	// Read the change request
	content, err := fs.ReadFile(options.ChangeRequestPath)
	if err != nil {
		return fmt.Errorf("error reading change request file: %s", err)
	}

	// Parse the change request
	changeRequest, err := models.LoadChangeRequestFromContent(options.ChangeRequestPath, content)
	if err != nil {
		return fmt.Errorf("error parsing change request: %s", err)
	}

	// Process and print the user stories
	if err := processAndPrintUserStories(fs, output, options, changeRequest); err != nil {
		return fmt.Errorf("error processing user stories: %s", err)
	}

	return nil
}

// processAndPrintUserStories processes and prints the content of user stories in a change request.
// It handles filtering, formatting, and displaying the stories according to the provided options.
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

	// Initialize results
	result := ProcessingResult{
		TotalStories: len(changeRequest.UserStories),
	}

	// Process each user story
	for _, userStory := range changeRequest.UserStories {
		// Check if the story should be excluded
		if shouldExcludeStory(userStory, options.ExcludeStories) {
			result.SkippedStories++
			continue
		}

		// Print the file path as a markdown comment
		printFilePath(output, userStory.FilePath, options.ColorOutput)

		// Check if the user story file exists
		if !fs.Exists(userStory.FilePath) {
			output.PrintError(fmt.Sprintf("User story file not found: %s", userStory.FilePath))
			result.ErrorStories++
			continue
		}

		// Read the user story file
		contentBytes, err := fs.ReadFile(userStory.FilePath)
		if err != nil {
			output.PrintError(fmt.Sprintf("Error reading user story file: %s", err))
			result.ErrorStories++
			continue
		}

		content := string(contentBytes)

		// Skip if content doesn't match filter pattern
		if filterRegex != nil && !matchesFilter(filterRegex, content, userStory.Title) {
			result.SkippedStories++
			continue
		}

		// Process the content
		processedContent := processUserStoryContent(content, options.ShowContentHash)

		// Format content based on output mode
		if options.CompactMode {
			processedContent = createCompactOutput(userStory, processedContent)
		}

		// Print the content
		output.Print(processedContent)
		
		// Add a separator between user stories
		printSeparator(output, options.CompactMode)

		result.DisplayedStories++
	}

	// Print summary if needed
	printSummary(output, result, options)

	return nil
}

// printFilePath prints the file path as a markdown comment,
// with optional color formatting.
func printFilePath(output UserStoryOutput, filePath string, useColor bool) {
	comment := fmt.Sprintf("[//]: # (%s)\n", filePath)
	if useColor {
		output.PrintSuccess(comment)
	} else {
		output.Print(comment)
	}
}

// printSeparator adds a separator between user stories based on output mode.
func printSeparator(output UserStoryOutput, compactMode bool) {
	if !compactMode {
		output.Print("\n---\n")
	} else {
		output.Print("\n")
	}
}

// printSummary prints a summary of the processing results if appropriate.
func printSummary(output UserStoryOutput, result ProcessingResult, options CatOptions) {
	// Print summary if debug is enabled, filtering was applied, or stories were skipped
	if options.FilterPattern != "" || result.SkippedStories > 0 || output.IsDebugEnabled() {
		summary := fmt.Sprintf(
			"Summary: %d of %d stories displayed, %d skipped, %d errors",
			result.DisplayedStories, result.TotalStories, result.SkippedStories, result.ErrorStories,
		)
		
		if options.ColorOutput {
			output.PrintWarning(summary)
		} else {
			output.Print(summary)
		}
	}
}

// matchesFilter checks if a story's content or title matches the given regex pattern.
func matchesFilter(regex *regexp.Regexp, content, title string) bool {
	return regex.MatchString(content) || regex.MatchString(title)
}

// shouldExcludeStory checks if a story should be excluded based on the exclude patterns.
// Returns true if any pattern matches the story's title or file path.
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

// createCompactOutput creates a compact representation of a user story.
// It extracts the title and first paragraph for a more concise view.
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

// processUserStoryContent processes the content of a user story.
// If showContentHash is false, it removes the content hash line from the frontmatter.
func processUserStoryContent(content string, showContentHash bool) string {
	return utils.FilterContentHash(content, showContentHash)
}

func init() {
	rootCmd.AddCommand(catCmd)

	// Add flags with descriptions
	catCmd.Flags().BoolP("show-content-hash", "c", false, "Show content hash in the output")
	catCmd.Flags().StringP("filter", "f", "", "Filter stories by regular expression pattern")
	catCmd.Flags().BoolP("color", "l", false, "Use colorized output")
	catCmd.Flags().BoolP("compact", "m", false, "Use compact output format (title and first paragraph only)")
	catCmd.Flags().StringSliceP("exclude", "e", []string{}, "Exclude stories matching these patterns (comma-separated)")
} 
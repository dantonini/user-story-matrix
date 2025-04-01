// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user-story-matrix/usm/internal/models"
)

// Colors and styles
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	filePathStyle = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("8"))
	hashStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	dateStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("14"))
	numberStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	subtitleStyle = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("13"))
)

// FormatUserStoryListItem formats a user story as a list item
func FormatUserStoryListItem(story models.UserStory, index int) string {
	// Format: [01] Title (./path/to/file.md)
	title := titleStyle.Render(story.Title)
	number := numberStyle.Render(fmt.Sprintf("[%s]", story.SequentialNumber))
	filePath := filePathStyle.Render(fmt.Sprintf("(%s)", shortPath(story.FilePath)))

	return fmt.Sprintf("%s %s %s", number, title, filePath)
}

// FormatUserStoryDetail formats a user story with full details
func FormatUserStoryDetail(story models.UserStory) string {
	// Create a styled representation of a user story
	var builder strings.Builder

	// Title
	builder.WriteString(titleStyle.Render(fmt.Sprintf("# %s\n", story.Title)))

	// Metadata
	builder.WriteString(filePathStyle.Render(fmt.Sprintf("Path: %s\n", story.FilePath)))
	builder.WriteString(hashStyle.Render(fmt.Sprintf("Hash: %s\n", story.ContentHash)))

	// Dates
	if !story.CreatedAt.IsZero() {
		builder.WriteString(dateStyle.Render(fmt.Sprintf("Created: %s\n", story.CreatedAt.Format("2006-01-02 15:04:05"))))
	}
	if !story.LastUpdated.IsZero() {
		builder.WriteString(dateStyle.Render(fmt.Sprintf("Updated: %s\n", story.LastUpdated.Format("2006-01-02 15:04:05"))))
	}

	// Content preview (first few lines)
	if story.Content != "" {
		lines := strings.Split(story.Content, "\n")
		contentPreview := lines
		if len(lines) > 10 {
			contentPreview = lines[:10]
			contentPreview = append(contentPreview, "...")
		}

		builder.WriteString("\nContent Preview:\n")
		for _, line := range contentPreview {
			builder.WriteString(fmt.Sprintf("%s\n", line))
		}
	}

	return builder.String()
}

// FormatChangeRequestListItem formats a change request as a list item
func FormatChangeRequestListItem(cr models.ChangeRequest, index int) string {
	// Format: [index] Name (created: date) [3 user stories]
	name := titleStyle.Render(cr.Name)
	number := numberStyle.Render(fmt.Sprintf("[%d]", index+1))
	date := dateStyle.Render(cr.CreatedAt.Format("2006-01-02"))
	storiesCount := subtitleStyle.Render(fmt.Sprintf("[%d user stories]", len(cr.UserStories)))

	return fmt.Sprintf("%s %s (created: %s) %s", number, name, date, storiesCount)
}

// FormatChangeRequestDetail formats a change request with full details
func FormatChangeRequestDetail(cr models.ChangeRequest) string {
	// Create a styled representation of a change request
	var builder strings.Builder

	// Title
	builder.WriteString(titleStyle.Render(fmt.Sprintf("# %s\n", cr.Name)))

	// Metadata
	builder.WriteString(filePathStyle.Render(fmt.Sprintf("Path: %s\n", cr.FilePath)))

	// Date
	if !cr.CreatedAt.IsZero() {
		builder.WriteString(dateStyle.Render(fmt.Sprintf("Created: %s\n", cr.CreatedAt.Format("2006-01-02 15:04:05"))))
	}

	// User Stories
	builder.WriteString(headerStyle.Render("\nUser Stories:\n"))
	for i, us := range cr.UserStories {
		number := fmt.Sprintf("%d.", i+1)
		title := us.Title
		filePath := shortPath(us.FilePath)

		builder.WriteString(fmt.Sprintf("%s %s (%s)\n", number, title, filePath))
	}

	return builder.String()
}

// FormatUserStoryTable formats user stories as a table for the CLI
func FormatUserStoryTable(stories []models.UserStory) ([]string, [][]string) {
	headers := []string{"#", "Title", "Created At", "Path"}
	rows := make([][]string, len(stories))

	for i, story := range stories {
		rows[i] = []string{
			story.SequentialNumber,
			story.Title,
			story.CreatedAt.Format("2006-01-02"),
			shortPath(story.FilePath),
		}
	}

	return headers, rows
}

// FormatChangeRequestTable formats change requests as a table for the CLI
func FormatChangeRequestTable(requests []models.ChangeRequest) ([]string, [][]string) {
	headers := []string{"#", "Name", "Created At", "User Stories", "Path"}
	rows := make([][]string, len(requests))

	for i, cr := range requests {
		rows[i] = []string{
			fmt.Sprintf("%d", i+1),
			cr.Name,
			cr.CreatedAt.Format("2006-01-02"),
			fmt.Sprintf("%d", len(cr.UserStories)),
			shortPath(cr.FilePath),
		}
	}

	return headers, rows
}

// shortPath returns a shortened version of a file path for display
func shortPath(path string) string {
	// If the path is not too long, return it as is
	if len(path) < 40 {
		return path
	}

	// Otherwise, keep the filename and a few parent directories
	dir, file := filepath.Split(path)
	dirs := strings.Split(dir, string(filepath.Separator))

	// If we have only a few directories, return the full path
	if len(dirs) <= 3 {
		return path
	}

	// Return ".../<dir1>/<dir2>/<filename>"
	return fmt.Sprintf("...%c%s%c%s", filepath.Separator,
		strings.Join(dirs[len(dirs)-3:len(dirs)-1], string(filepath.Separator)),
		filepath.Separator, file)
}

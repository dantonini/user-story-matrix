// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestFormatUserStoryListItem(t *testing.T) {
	story := models.UserStory{
		Title:            "Test Story",
		FilePath:         "docs/user-stories/01-test-story.md",
		SequentialNumber: "01",
	}

	result := FormatUserStoryListItem(story, 0)
	
	// Since the output is styled with lipgloss, we can't do exact string matches
	// Instead, check that the expected content is present
	assert.Contains(t, result, "01")
	assert.Contains(t, result, "Test Story")
	assert.Contains(t, result, "docs/user-stories/01-test-story.md")
}

func TestFormatUserStoryDetail(t *testing.T) {
	// Set up a fixed time for testing
	testTime, _ := time.Parse(time.RFC3339, "2025-01-02T15:04:05Z")
	
	tests := []struct {
		name         string
		story        models.UserStory
		expectedParts []string
	}{
		{
			name: "Full story with all fields",
			story: models.UserStory{
				Title:            "Test Story",
				FilePath:         "docs/user-stories/01-test-story.md",
				SequentialNumber: "01",
				ContentHash:      "abc123",
				CreatedAt:        testTime,
				LastUpdated:      testTime.Add(24 * time.Hour),
				Content:          "# Test Story\n\nThis is a test story\nwith multiple lines.",
			},
			expectedParts: []string{
				"Test Story",
				"docs/user-stories/01-test-story.md",
				"abc123",
				"2025-01-02 15:04:05", // Created
				"2025-01-03 15:04:05", // Updated
				"This is a test story",
				"with multiple lines",
			},
		},
		{
			name: "Story with no content",
			story: models.UserStory{
				Title:            "Empty Story",
				FilePath:         "docs/user-stories/empty.md",
				SequentialNumber: "02",
				ContentHash:      "empty123",
				CreatedAt:        testTime,
			},
			expectedParts: []string{
				"Empty Story",
				"docs/user-stories/empty.md",
				"empty123",
				"2025-01-02 15:04:05", // Created
			},
		},
		{
			name: "Story with very long content",
			story: models.UserStory{
				Title:            "Long Story",
				FilePath:         "docs/user-stories/long.md",
				SequentialNumber: "03",
				Content:          strings.Repeat("Line of content\n", 20),
			},
			expectedParts: []string{
				"Long Story",
				"docs/user-stories/long.md",
				"Line of content",
				"...", // Should indicate truncation
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatUserStoryDetail(tt.story)
			
			for _, part := range tt.expectedParts {
				assert.Contains(t, result, part)
			}
		})
	}
}

func TestFormatChangeRequestListItem(t *testing.T) {
	testTime, _ := time.Parse(time.RFC3339, "2025-01-02T15:04:05Z")
	
	cr := models.ChangeRequest{
		Name:      "Test Change Request",
		FilePath:  "docs/change-requests/test.md",
		CreatedAt: testTime,
		UserStories: []models.UserStoryReference{
			{Title: "Story 1", FilePath: "docs/user-stories/01.md"},
			{Title: "Story 2", FilePath: "docs/user-stories/02.md"},
		},
	}

	result := FormatChangeRequestListItem(cr, 0)
	
	assert.Contains(t, result, "Test Change Request")
	assert.Contains(t, result, "2025-01-02") // Date
	assert.Contains(t, result, "2 user stories") // Count
}

func TestFormatChangeRequestDetail(t *testing.T) {
	testTime, _ := time.Parse(time.RFC3339, "2025-01-02T15:04:05Z")
	
	tests := []struct {
		name string
		cr   models.ChangeRequest
	}{
		{
			name: "Change request with multiple stories",
			cr: models.ChangeRequest{
				Name:      "Test Change Request",
				FilePath:  "docs/change-requests/test.md",
				CreatedAt: testTime,
				UserStories: []models.UserStoryReference{
					{Title: "Story 1", FilePath: "docs/user-stories/01.md"},
					{Title: "Story 2", FilePath: "docs/user-stories/02.md"},
				},
			},
		},
		{
			name: "Change request with no stories",
			cr: models.ChangeRequest{
				Name:      "Empty Change Request",
				FilePath:  "docs/change-requests/empty.md",
				CreatedAt: testTime,
				UserStories: []models.UserStoryReference{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatChangeRequestDetail(tt.cr)
			
			// Simply verify that the result is not empty and contains the basic information
			assert.NotEmpty(t, result)
			assert.Contains(t, result, tt.cr.Name)
			assert.Contains(t, result, tt.cr.FilePath)
			
			// Check that the date is correctly formatted
			dateStr := testTime.Format("2025-01-02")
			assert.Contains(t, result, dateStr)
			
			// Check user stories if any
			for i, story := range tt.cr.UserStories {
				assert.Contains(t, result, story.Title)
				assert.Contains(t, result, story.FilePath)
				assert.Contains(t, result, fmt.Sprintf("%d.", i+1)) // Check numbering
			}
		})
	}
}

func TestFormatUserStoryTable(t *testing.T) {
	testTime, _ := time.Parse(time.RFC3339, "2025-01-02T15:04:05Z")
	
	stories := []models.UserStory{
		{
			Title:            "Story 1",
			FilePath:         "docs/user-stories/01.md",
			SequentialNumber: "01",
			CreatedAt:        testTime,
		},
		{
			Title:            "Story 2",
			FilePath:         "docs/user-stories/02.md",
			SequentialNumber: "02",
			CreatedAt:        testTime.Add(24 * time.Hour),
		},
	}

	headers, rows := FormatUserStoryTable(stories)
	
	// Verify headers
	expectedHeaders := []string{"#", "Title", "Created At", "Path"}
	assert.Equal(t, expectedHeaders, headers)
	
	// Verify row contents
	assert.Equal(t, 2, len(rows))
	
	// Check first row
	assert.Equal(t, "01", rows[0][0])
	assert.Equal(t, "Story 1", rows[0][1])
	assert.Equal(t, "2025-01-02", rows[0][2])
	assert.Equal(t, "docs/user-stories/01.md", rows[0][3])
	
	// Check second row
	assert.Equal(t, "02", rows[1][0])
	assert.Equal(t, "Story 2", rows[1][1])
	assert.Equal(t, "2025-01-03", rows[1][2])
	assert.Equal(t, "docs/user-stories/02.md", rows[1][3])
}

func TestFormatChangeRequestTable(t *testing.T) {
	testTime, _ := time.Parse(time.RFC3339, "2025-01-02T15:04:05Z")
	
	requests := []models.ChangeRequest{
		{
			Name:      "CR 1",
			FilePath:  "docs/change-requests/cr1.md",
			CreatedAt: testTime,
			UserStories: []models.UserStoryReference{
				{Title: "Story 1"},
				{Title: "Story 2"},
			},
		},
		{
			Name:      "CR 2",
			FilePath:  "docs/change-requests/cr2.md",
			CreatedAt: testTime.Add(24 * time.Hour),
			UserStories: []models.UserStoryReference{
				{Title: "Story 3"},
			},
		},
	}

	headers, rows := FormatChangeRequestTable(requests)
	
	// Verify headers
	expectedHeaders := []string{"#", "Name", "Created At", "User Stories", "Path"}
	assert.Equal(t, expectedHeaders, headers)
	
	// Verify row contents
	assert.Equal(t, 2, len(rows))
	
	// Check first row
	assert.Equal(t, "1", rows[0][0])
	assert.Equal(t, "CR 1", rows[0][1])
	assert.Equal(t, "2025-01-02", rows[0][2])
	assert.Equal(t, "2", rows[0][3]) // Number of user stories
	assert.Equal(t, "docs/change-requests/cr1.md", rows[0][4])
	
	// Check second row
	assert.Equal(t, "2", rows[1][0])
	assert.Equal(t, "CR 2", rows[1][1])
	assert.Equal(t, "2025-01-03", rows[1][2])
	assert.Equal(t, "1", rows[1][3]) // Number of user stories
	assert.Equal(t, "docs/change-requests/cr2.md", rows[1][4])
}

func TestShortPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Short path unchanged",
			path:     "docs/file.md",
			expected: "docs/file.md",
		},
		{
			name:     "Just under threshold",
			path:     strings.Repeat("a", 39),
			expected: strings.Repeat("a", 39),
		},
		{
			name:     "Path with few directories",
			path:     "dir1/dir2/file.md",
			expected: "dir1/dir2/file.md",
		},
		{
			name:     "Long path with many directories",
			path:     "/very/long/path/with/many/subdirectories/to/a/file.md",
			expected: ".../to/a/file.md",
		},
		{
			name:     "Very long path with deep nesting",
			path:     "/root/dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/file.md",
			expected: ".../dir7/dir8/file.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shortPath(tt.path)
			
			// For Windows compatibility, normalize separators for comparison
			normalizedExpected := strings.ReplaceAll(tt.expected, "/", fmt.Sprintf("%c", filepath.Separator))
			assert.Equal(t, normalizedExpected, result)
		})
	}
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package pages

import (
	"testing"

	"github.com/user-story-matrix/usm/internal/models"
)

func TestSelectionPage(t *testing.T) {
	// Create test stories
	stories := []models.UserStory{
		{
			Title:         "User Story 1",
			FilePath:      "docs/user-stories/user-story-1.md",
			IsImplemented: false,
		},
		{
			Title:         "User Story 2",
			FilePath:      "docs/user-stories/user-story-2.md",
			IsImplemented: true,
		},
		{
			Title:         "User Story 3",
			FilePath:      "docs/user-stories/user-story-3.md",
			IsImplemented: false,
		},
	}

	// Test showing only unimplemented stories
	t.Run("ShowUnimplementedOnly", func(t *testing.T) {
		page := New(stories, false)
		page.Init()

		// Check that only unimplemented stories are shown
		if len(page.state.VisibleStories) != 2 {
			t.Errorf("Expected 2 visible stories, got %d", len(page.state.VisibleStories))
		}
	})

	// Test showing all stories
	t.Run("ShowAll", func(t *testing.T) {
		page := New(stories, true)
		page.Init()

		// Check that all stories are shown
		if len(page.state.VisibleStories) != 3 {
			t.Errorf("Expected 3 visible stories, got %d", len(page.state.VisibleStories))
		}
	})

	// Test toggling implementation filter
	t.Run("ToggleImplementationFilter", func(t *testing.T) {
		page := New(stories, false)
		page.Init()

		// Initially only unimplemented stories are shown
		if len(page.state.VisibleStories) != 2 {
			t.Errorf("Expected 2 visible stories, got %d", len(page.state.VisibleStories))
		}

		// Toggle to show all stories
		page.state.ToggleImplementationFilter()
		page.updateResults()

		// Now all stories should be shown
		if len(page.state.VisibleStories) != 3 {
			t.Errorf("Expected 3 visible stories, got %d", len(page.state.VisibleStories))
		}
	})

	// Test selection
	t.Run("Selection", func(t *testing.T) {
		page := New(stories, true)
		page.Init()

		// Initially no stories are selected
		if page.state.SelectedCount() != 0 {
			t.Errorf("Expected 0 selected stories, got %d", page.state.SelectedCount())
		}

		// Select a story
		page.state.ToggleSelection(stories[0].FilePath)

		// Now one story should be selected
		if page.state.SelectedCount() != 1 {
			t.Errorf("Expected 1 selected story, got %d", page.state.SelectedCount())
		}

		// Toggle the same story to deselect it
		page.state.ToggleSelection(stories[0].FilePath)

		// Now no stories should be selected
		if page.state.SelectedCount() != 0 {
			t.Errorf("Expected 0 selected stories, got %d", page.state.SelectedCount())
		}
	})

	// Test search
	t.Run("Search", func(t *testing.T) {
		page := New(stories, true)
		page.Init()

		// Initially all stories are shown
		if len(page.state.VisibleStories) != 3 {
			t.Errorf("Expected 3 visible stories, got %d", len(page.state.VisibleStories))
		}

		// Set search text
		page.searchBox = page.searchBox.SetValue("User Story 1")
		page.updateResults()

		// Now only one story should be shown
		if len(page.state.VisibleStories) != 1 {
			t.Errorf("Expected 1 visible story, got %d", len(page.state.VisibleStories))
		}

		// Clear search text
		page.searchBox = page.searchBox.SetValue("")
		page.updateResults()

		// Now all stories should be shown again
		if len(page.state.VisibleStories) != 3 {
			t.Errorf("Expected 3 visible stories, got %d", len(page.state.VisibleStories))
		}
	})

	// Test focus switching
	t.Run("FocusSwitching", func(t *testing.T) {
		page := New(stories, true)
		page.Init()

		// Initially search box is focused
		if !page.state.SearchFocused {
			t.Errorf("Expected search box to be focused")
		}

		// Switch focus to list
		page.state.FocusList()

		// Now list should be focused
		if page.state.SearchFocused {
			t.Errorf("Expected list to be focused")
		}

		// Switch focus back to search
		page.state.FocusSearch()

		// Now search box should be focused
		if !page.state.SearchFocused {
			t.Errorf("Expected search box to be focused")
		}
	})
} 
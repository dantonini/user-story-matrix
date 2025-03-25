// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package models

import (
	"github.com/user-story-matrix/usm/internal/models"
)

// UIState represents the current state of the TUI
type UIState struct {
	// Focus state
	SearchFocused bool

	// Filter state
	FilterText     string
	ShowImplemented bool

	// Selection state
	SelectedIDs map[string]bool // Map of story IDs to selection state

	// Current view
	VisibleStories  []models.UserStory
	CursorPosition  int
	TotalStories    int
	FilteredStories int
}

// NewUIState creates a new UI state
func NewUIState() *UIState {
	return &UIState{
		SearchFocused:   true, // Start with search focused
		ShowImplemented: false, // Default to showing only unimplemented stories
		SelectedIDs:     make(map[string]bool),
		CursorPosition:  0,
	}
}

// FocusSearch sets the focus to the search box
func (s *UIState) FocusSearch() {
	s.SearchFocused = true
}

// FocusList sets the focus to the story list
func (s *UIState) FocusList() {
	s.SearchFocused = false
}

// ToggleImplementationFilter toggles whether to show implemented stories
func (s *UIState) ToggleImplementationFilter() {
	s.ShowImplemented = !s.ShowImplemented
}

// SetFilterText updates the filter text
func (s *UIState) SetFilterText(text string) {
	s.FilterText = text
}

// ToggleSelection toggles whether the specified story is selected
func (s *UIState) ToggleSelection(id string) {
	if _, exists := s.SelectedIDs[id]; exists {
		delete(s.SelectedIDs, id)
	} else {
		s.SelectedIDs[id] = true
	}
}

// IsSelected returns whether the specified story is selected
func (s *UIState) IsSelected(id string) bool {
	_, exists := s.SelectedIDs[id]
	return exists
}

// SelectedCount returns the number of selected stories
func (s *UIState) SelectedCount() int {
	return len(s.SelectedIDs)
}

// SetVisibleStories updates the visible stories
func (s *UIState) SetVisibleStories(stories []models.UserStory, totalStories int) {
	s.VisibleStories = stories
	s.FilteredStories = len(stories)
	s.TotalStories = totalStories
	
	// Reset cursor position if it's out of bounds
	if s.CursorPosition >= len(stories) {
		if len(stories) > 0 {
			s.CursorPosition = 0
		} else {
			s.CursorPosition = -1
		}
	}
}

// GetSelectedStoryIndices returns the indices of all selected stories
func (s *UIState) GetSelectedStoryIndices(allStories []models.UserStory) []int {
	var selected []int
	
	for i, story := range allStories {
		if s.IsSelected(story.FilePath) {
			selected = append(selected, i)
		}
	}
	
	return selected
} 
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
	if id == "" {
		return // Safety check for empty ID
	}
	
	if _, exists := s.SelectedIDs[id]; exists {
		delete(s.SelectedIDs, id)
	} else {
		s.SelectedIDs[id] = true
	}
}

// IsSelected returns whether the specified story is selected
func (s *UIState) IsSelected(id string) bool {
	if id == "" {
		return false // Safety check for empty ID
	}
	_, exists := s.SelectedIDs[id]
	return exists
}

// SelectedCount returns the number of selected stories
func (s *UIState) SelectedCount() int {
	return len(s.SelectedIDs)
}

// SetVisibleStories updates the visible stories
func (s *UIState) SetVisibleStories(stories []models.UserStory, totalStories int) {
	if stories == nil {
		stories = []models.UserStory{} // Convert nil to empty slice for safety
	}
	
	s.VisibleStories = stories
	s.FilteredStories = len(stories)
	s.TotalStories = totalStories
	
	// Reset cursor position if it's out of bounds
	if len(stories) == 0 {
		s.CursorPosition = -1 // No items to select
	} else if s.CursorPosition >= len(stories) || s.CursorPosition < 0 {
		s.CursorPosition = 0 // Reset to first item
	}
}

// GetSelectedStoryIndices returns the indices of all selected stories
func (s *UIState) GetSelectedStoryIndices(allStories []models.UserStory) []int {
	if allStories == nil {
		return []int{} // Return empty slice if no stories
	}
	
	var selected []int
	
	for i, story := range allStories {
		if story.FilePath != "" && s.IsSelected(story.FilePath) {
			selected = append(selected, i)
		}
	}
	
	return selected
}

// HiddenSelectedCount returns the number of selected stories that are not currently visible
func (s *UIState) HiddenSelectedCount() int {
	if len(s.SelectedIDs) == 0 {
		return 0 // Quick return if nothing is selected
	}
	
	// Count selected stories that are not in the visible stories
	visibleIDs := make(map[string]bool, len(s.VisibleStories))
	
	// Add all visible story IDs to the map
	for _, story := range s.VisibleStories {
		if story.FilePath != "" {
			visibleIDs[story.FilePath] = true
		}
	}
	
	// Count selected stories that are not in the visible stories
	hiddenCount := 0
	for id := range s.SelectedIDs {
		if !visibleIDs[id] {
			hiddenCount++
		}
	}
	
	return hiddenCount
} 
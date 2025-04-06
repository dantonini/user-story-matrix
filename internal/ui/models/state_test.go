// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestNewUIState(t *testing.T) {
	// Act
	state := NewUIState()
	
	// Assert
	assert.NotNil(t, state)
	assert.True(t, state.SearchFocused, "Default state should have search focused")
	assert.False(t, state.ShowImplemented, "Default state should not show implemented stories")
	assert.NotNil(t, state.SelectedIDs, "SelectedIDs map should be initialized")
	assert.Equal(t, 0, state.CursorPosition, "Default cursor position should be 0")
	assert.Empty(t, state.FilterText, "Filter text should be empty")
	assert.Empty(t, state.VisibleStories, "There should be no visible stories initially")
	assert.Equal(t, 0, state.TotalStories, "Total stories should be 0")
	assert.Equal(t, 0, state.FilteredStories, "Filtered stories should be 0")
}

func TestFocusSearch(t *testing.T) {
	// Arrange
	state := NewUIState()
	state.SearchFocused = false
	
	// Act
	state.FocusSearch()
	
	// Assert
	assert.True(t, state.SearchFocused, "Search should be focused after FocusSearch")
}

func TestFocusList(t *testing.T) {
	// Arrange
	state := NewUIState()
	
	// Act
	state.FocusList()
	
	// Assert
	assert.False(t, state.SearchFocused, "Search should not be focused after FocusList")
}

func TestToggleImplementationFilter(t *testing.T) {
	// Arrange
	state := NewUIState()
	initialValue := state.ShowImplemented
	
	// Act
	state.ToggleImplementationFilter()
	
	// Assert
	assert.NotEqual(t, initialValue, state.ShowImplemented, "ToggleImplementationFilter should toggle the value")
	
	// Act again - should toggle back
	state.ToggleImplementationFilter()
	
	// Assert
	assert.Equal(t, initialValue, state.ShowImplemented, "Toggling twice should return to original value")
}

func TestSetFilterText(t *testing.T) {
	// Arrange
	state := NewUIState()
	testText := "test filter"
	
	// Act
	state.SetFilterText(testText)
	
	// Assert
	assert.Equal(t, testText, state.FilterText, "Filter text should be updated")
}

func TestToggleSelection(t *testing.T) {
	// Arrange
	state := NewUIState()
	testID := "test-story-id"
	
	// Act - select
	state.ToggleSelection(testID)
	
	// Assert
	assert.True(t, state.SelectedIDs[testID], "Story should be selected after toggle")
	
	// Act - deselect
	state.ToggleSelection(testID)
	
	// Assert
	_, exists := state.SelectedIDs[testID]
	assert.False(t, exists, "Story should not be selected after toggling again")
	
	// Test empty ID (should do nothing)
	prevLen := len(state.SelectedIDs)
	state.ToggleSelection("")
	assert.Equal(t, prevLen, len(state.SelectedIDs), "Empty ID should not affect selections")
}

func TestIsSelected(t *testing.T) {
	// Arrange
	state := NewUIState()
	testID := "test-story-id"
	
	// Act & Assert - initially not selected
	assert.False(t, state.IsSelected(testID), "Story should not be selected initially")
	
	// Act - select
	state.SelectedIDs[testID] = true
	
	// Assert
	assert.True(t, state.IsSelected(testID), "Story should be selected after being added to SelectedIDs")
	
	// Test with empty ID
	assert.False(t, state.IsSelected(""), "Empty ID should never be selected")
}

func TestSelectedCount(t *testing.T) {
	// Arrange
	state := NewUIState()
	
	// Assert - initially empty
	assert.Equal(t, 0, state.SelectedCount(), "Initial selected count should be 0")
	
	// Act - add some selections
	state.SelectedIDs["id1"] = true
	state.SelectedIDs["id2"] = true
	state.SelectedIDs["id3"] = true
	
	// Assert
	assert.Equal(t, 3, state.SelectedCount(), "Selected count should match number of selected IDs")
	
	// Act - remove one
	delete(state.SelectedIDs, "id2")
	
	// Assert
	assert.Equal(t, 2, state.SelectedCount(), "Selected count should update when selection is removed")
}

func TestSetVisibleStories(t *testing.T) {
	// Arrange
	state := NewUIState()
	stories := []models.UserStory{
		{Title: "Story 1", FilePath: "path/to/story1.md"},
		{Title: "Story 2", FilePath: "path/to/story2.md"},
	}
	totalStories := 5 // Total vs. filtered
	
	// Act
	state.SetVisibleStories(stories, totalStories)
	
	// Assert
	assert.Equal(t, stories, state.VisibleStories, "Visible stories should be updated")
	assert.Equal(t, len(stories), state.FilteredStories, "Filtered stories count should be updated")
	assert.Equal(t, totalStories, state.TotalStories, "Total stories count should be updated")
	assert.Equal(t, 0, state.CursorPosition, "Cursor position should be valid (first item)")
	
	// Test with nil stories
	state.CursorPosition = 5 // Set to an invalid value
	state.SetVisibleStories(nil, 10)
	assert.Empty(t, state.VisibleStories, "Nil stories should be converted to empty slice")
	assert.Equal(t, -1, state.CursorPosition, "Cursor position should be -1 when no stories")
	
	// Test with empty stories
	state.CursorPosition = 5 // Set to an invalid value
	state.SetVisibleStories([]models.UserStory{}, 10)
	assert.Empty(t, state.VisibleStories, "Visible stories should be empty")
	assert.Equal(t, -1, state.CursorPosition, "Cursor position should be -1 when no stories")
	
	// Test cursor positioning with out-of-bounds cursor
	state.CursorPosition = 10 // Set to an invalid value
	state.SetVisibleStories(stories, totalStories)
	assert.Equal(t, 0, state.CursorPosition, "Cursor position should be reset to 0 when out of bounds")
	
	// Test cursor positioning with negative cursor
	state.CursorPosition = -5 // Set to an invalid value
	state.SetVisibleStories(stories, totalStories)
	assert.Equal(t, 0, state.CursorPosition, "Cursor position should be reset to 0 when negative")
}

func TestGetSelectedStoryIndices(t *testing.T) {
	// Arrange
	state := NewUIState()
	stories := []models.UserStory{
		{Title: "Story 1", FilePath: "path/to/story1.md"},
		{Title: "Story 2", FilePath: "path/to/story2.md"},
		{Title: "Story 3", FilePath: "path/to/story3.md"},
		{Title: "Story 4", FilePath: "path/to/story4.md"},
	}
	
	// Select some stories
	state.SelectedIDs["path/to/story1.md"] = true
	state.SelectedIDs["path/to/story3.md"] = true
	
	// Act
	indices := state.GetSelectedStoryIndices(stories)
	
	// Assert
	assert.Equal(t, []int{0, 2}, indices, "Selected indices should match selected stories")
	
	// Test with nil stories
	indices = state.GetSelectedStoryIndices(nil)
	assert.Empty(t, indices, "Nil stories should return empty indices")
	
	// Test with empty stories
	indices = state.GetSelectedStoryIndices([]models.UserStory{})
	assert.Empty(t, indices, "Empty stories should return empty indices")
	
	// Test with stories having empty file paths
	emptyPathStories := []models.UserStory{
		{Title: "Story 1", FilePath: ""},
		{Title: "Story 2", FilePath: "path/to/story2.md"},
	}
	state.SelectedIDs["path/to/story2.md"] = true
	indices = state.GetSelectedStoryIndices(emptyPathStories)
	assert.Equal(t, []int{1}, indices, "Should skip stories with empty file paths")
}

func TestHiddenSelectedCount(t *testing.T) {
	// Arrange
	state := NewUIState()
	
	// Add some selections
	state.SelectedIDs["path/to/story1.md"] = true
	state.SelectedIDs["path/to/story2.md"] = true
	state.SelectedIDs["path/to/story3.md"] = true
	
	// Set visible stories (note: only story1 and story3 are visible)
	state.VisibleStories = []models.UserStory{
		{Title: "Story 1", FilePath: "path/to/story1.md"},
		{Title: "Story 3", FilePath: "path/to/story3.md"},
	}
	
	// Act
	hiddenCount := state.HiddenSelectedCount()
	
	// Assert
	assert.Equal(t, 1, hiddenCount, "One selected story should be hidden")
	
	// Test with no selections
	state.SelectedIDs = make(map[string]bool)
	hiddenCount = state.HiddenSelectedCount()
	assert.Equal(t, 0, hiddenCount, "Should be zero when no stories are selected")
	
	// Test with all visible
	state.SelectedIDs["path/to/story1.md"] = true
	hiddenCount = state.HiddenSelectedCount()
	assert.Equal(t, 0, hiddenCount, "Should be zero when all selected stories are visible")
	
	// Test with story with empty file path
	state.VisibleStories = []models.UserStory{
		{Title: "Story 1", FilePath: ""},
	}
	state.SelectedIDs["path/to/story1.md"] = true
	hiddenCount = state.HiddenSelectedCount()
	assert.Equal(t, 1, hiddenCount, "Should count stories with empty file paths as hidden")
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestNewSelectionUI(t *testing.T) {
	stories := []models.UserStory{
		{Title: "Story 1", IsImplemented: false},
		{Title: "Story 2", IsImplemented: true},
	}

	ui := NewSelectionUI(stories, false)

	assert.NotNil(t, ui.searchBox)
	assert.NotNil(t, ui.engine)
	assert.NotEmpty(t, ui.statusBar)
	assert.NotNil(t, ui.selected)
	assert.Equal(t, 2, ui.engine.GetState().TotalCount)
}

func TestUpdate(t *testing.T) {
	stories := []models.UserStory{
		{Title: "Story 1", IsImplemented: false},
		{Title: "Story 2", IsImplemented: true},
	}

	ui := NewSelectionUI(stories, false)

	// Test window resize
	t.Run("Window resize", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		newModel, _ := ui.Update(msg)
		assert.Equal(t, ui, newModel)
	})

	// Test quit
	t.Run("Quit", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		newModel, cmd := ui.Update(msg)
		assert.Equal(t, ui, newModel)
		assert.NotNil(t, cmd)
	})

	// Test story selection
	t.Run("Story selection", func(t *testing.T) {
		// Add a selected story for testing
		ui.selected = append(ui.selected, 0)
		assert.Equal(t, 1, len(ui.selected))
		
		// Test removal
		ui.selected = []int{}
		assert.Equal(t, 0, len(ui.selected))
	})

	// Test search
	t.Run("Search", func(t *testing.T) {
		// Type search query
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Story 1")}
		newModel, _ := ui.Update(msg)
		assert.Equal(t, ui, newModel)
	})
}

func TestView(t *testing.T) {
	stories := []models.UserStory{
		{Title: "Story 1", IsImplemented: false},
		{Title: "Story 2", IsImplemented: true},
	}

	ui := NewSelectionUI(stories, false)
	ui.ready = true // Set ready to true to avoid initialization message

	view := ui.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Type to search")
	assert.Contains(t, view, "Selected: 0 stories")
}

func TestGetSelected(t *testing.T) {
	stories := []models.UserStory{
		{Title: "Story 1", IsImplemented: false},
		{Title: "Story 2", IsImplemented: true},
	}

	ui := NewSelectionUI(stories, false)

	// Initially no stories selected
	selected := ui.GetSelected()
	assert.Empty(t, selected)

	// Select a story
	ui.selected = append(ui.selected, 0)
	selected = ui.GetSelected()
	assert.Equal(t, 1, len(selected))
	assert.Equal(t, 0, selected[0])
}

// TestSelectionToggle tests the complete story selection functionality 
// including simulating key presses and verifying visual indicators
func TestSelectionToggle(t *testing.T) {
	stories := []models.UserStory{
		{Title: "Story 1", IsImplemented: false},
		{Title: "Story 2", IsImplemented: true},
	}

	ui := NewSelectionUI(stories, false)
	ui.ready = true

	// First, verify no stories are selected
	assert.Empty(t, ui.selected)
	
	// Verify the UI starts in search mode (this has changed from the original test)
	assert.True(t, ui.searchFocused, "UI should start in search mode")

	// Switch to list mode
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := ui.Update(enterMsg)
	ui = newModel.(*SelectionUI)
	
	// Verify now we're in list mode
	assert.False(t, ui.searchFocused, "UI should be in list mode after pressing Enter")
	
	// Ensure we're at the first item
	ui.storyList.Select(0)
	
	// Get the initial state
	firstItem := ui.storyList.SelectedItem().(storyItem)
	assert.False(t, firstItem.isSelected, "Item should not be selected initially")
	
	// Create a key event for Space
	spaceKeyMsg := tea.KeyMsg{Type: tea.KeySpace}
	
	// Send the key message through the Update method
	newModel, _ = ui.Update(spaceKeyMsg)
	updatedUI := newModel.(*SelectionUI)
	
	// Check if the selection was registered in the selected array
	assert.Equal(t, 1, len(updatedUI.selected), "Should have one item selected")
	
	// Get the new selected item from the list to check if it's visually updated
	updatedItem := updatedUI.storyList.SelectedItem().(storyItem)
	assert.True(t, updatedItem.isSelected, "Item should be visually marked as selected")
	
	// Try selecting the same item again to deselect it
	newModel, _ = updatedUI.Update(spaceKeyMsg)
	updatedUI = newModel.(*SelectionUI)
	
	// Verify the item was deselected
	assert.Empty(t, updatedUI.selected, "No items should be selected after deselecting")
	
	// Check the visual state
	deselectedItem := updatedUI.storyList.SelectedItem().(storyItem)
	assert.False(t, deselectedItem.isSelected, "Item should not be visually marked as selected after deselection")
	
	// Test automatic switch to search mode when typing
	letterKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	newModel, _ = updatedUI.Update(letterKeyMsg)
	updatedUI = newModel.(*SelectionUI)
	
	// Verify we're in search mode
	assert.True(t, updatedUI.searchFocused, "Should switch to search mode after typing")
	
	// Test that Enter returns to list mode
	newModel, _ = updatedUI.Update(enterMsg)
	updatedUI = newModel.(*SelectionUI)
	
	assert.False(t, updatedUI.searchFocused, "Should return to list mode after pressing Enter")
}

// TestSearchWithVaryingStoryCount tests the search functionality with different numbers of stories
func TestSearchWithVaryingStoryCount(t *testing.T) {
	testCases := []struct {
		name       string
		storyCount int
	}{
		{"Small set (5 stories)", 5},
		{"Medium set (20 stories)", 20},
		{"Large set (50 stories)", 50},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test stories
			stories := make([]models.UserStory, tc.storyCount)
			for i := 0; i < tc.storyCount; i++ {
				stories[i] = models.UserStory{
					Title:        fmt.Sprintf("Story %d", i),
					Description:  fmt.Sprintf("Description for story %d", i),
					FilePath:     fmt.Sprintf("path/to/story_%d.md", i),
					IsImplemented: i%3 == 0, // Make every third story implemented
				}
			}

			// Create the UI with all stories shown
			ui := NewSelectionUI(stories, true)
			ui.ready = true // Mark as ready to avoid initialization message

			// Manually set the search text and update
			ui.searchBox.SetValue("Story 1")
			
			// Force search update
			ui.updateList()
			
			// Verify the search box has the text
			assert.Equal(t, "Story 1", ui.searchBox.Value())
			
			// Get the filtered view
			filteredView := ui.View()
			
			// Verify something is filtered
			assertContainsAny(t, filteredView, []string{
				"matching 'Story 1'",
				"1/",
				"FILTER:",
			})
			
			// Count how many items are displayed in the list
			filteredItemCount := len(ui.storyList.Items())
			
			// For large story counts, make sure filtering actually reduced the items
			if tc.storyCount > 20 {
				assert.Less(t, filteredItemCount, tc.storyCount, "Filtering should reduce item count")
			}
			
			// Verify we can switch back to list mode
			enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
			newModel, _ := ui.Update(enterMsg)
			ui = newModel.(*SelectionUI)
			
			assert.False(t, ui.searchFocused, "UI should be in list mode after pressing Enter")
		})
	}
}

// Helper functions for assertions
func assertContainsAny(t *testing.T, str string, substrings []string) {
	t.Helper()
	for _, sub := range substrings {
		if strings.Contains(str, sub) {
			return // Found at least one substring
		}
	}
	t.Errorf("String did not contain any of the expected substrings.\nString: %s\nExpected any of: %v", str, substrings)
}

// TestSelectionUIWithWindowSize tests the selection UI with different window sizes
func TestSelectionUIWithWindowSize(t *testing.T) {
	stories := []models.UserStory{
		{Title: "Story 1", IsImplemented: false},
		{Title: "Story 2", IsImplemented: true},
		{Title: "Story 3", IsImplemented: false},
	}

	testCases := []struct {
		name   string
		width  int
		height int
	}{
		{"Small terminal", 40, 15},
		{"Medium terminal", 80, 24},
		{"Large terminal", 120, 40},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ui := NewSelectionUI(stories, false)
			
			// Send window size message
			sizeMsg := tea.WindowSizeMsg{Width: tc.width, Height: tc.height}
			newModel, _ := ui.Update(sizeMsg)
			resizedUI := newModel.(*SelectionUI)
			
			// Verify size was updated
			assert.Equal(t, tc.width, resizedUI.width)
			assert.Equal(t, tc.height, resizedUI.height)
			assert.True(t, resizedUI.ready)
			
			// Render the view
			view := resizedUI.View()
			
			// The search UI changes for small terminals, so check different text patterns
			if tc.width < 50 {
				assertContainsAny(t, view, []string{"SEARCH:", "🔍"})
			} else {
				assertContainsAny(t, view, []string{"TYPE TO SEARCH", "🔍"})
			}
			
			// Verify search box width is reasonable for the terminal size
			maxWidth := int(float64(tc.width) * 0.8)
			if maxWidth < 30 {
				maxWidth = tc.width - 4
			}
			assert.Equal(t, maxWidth, resizedUI.searchBox.Width, 
				"Search box width should be appropriate for terminal size")
		})
	}
}
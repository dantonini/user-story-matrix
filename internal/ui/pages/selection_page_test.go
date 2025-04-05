// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package pages

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

// Test data
func getTestStories() []models.UserStory {
	return []models.UserStory{
		{
			Title:         "Add login functionality",
			FilePath:      "docs/user-stories/auth/01-add-login-functionality.md",
			Description:   "Users should be able to log in with their credentials",
			IsImplemented: false,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		},
		{
			Title:         "Integrate payment provider",
			FilePath:      "docs/user-stories/payment/01-integrate-payment-provider.md",
			Description:   "Users should be able to pay for services",
			IsImplemented: false,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		},
		{
			Title:         "Export user data to CSV",
			FilePath:      "docs/user-stories/export/01-export-user-data-to-csv.md",
			Description:   "Users should be able to export their data",
			IsImplemented: true,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		},
	}
}

// Test creating with nil stories
func TestCreateWithNilStories(t *testing.T) {
	// Create with nil stories (should not panic)
	page := New(nil, false)
	assert.NotNil(t, page, "Page should be created even with nil stories")
	assert.Equal(t, 0, len(page.stories), "Stories should be initialized to empty slice")
	
	// Initialize the page
	page.Init()
	
	// Get the view
	view := page.View()
	
	// Check if no stories message is shown
	assert.Contains(t, view, "No matching user stories found")
}

// Test initial view
func TestInitialView(t *testing.T) {
	page := New(getTestStories(), false)

	// Initialize the page
	page.Init()

	// Get the view
	view := page.View()

	// Check if key elements are present
	assert.Contains(t, view, "Search")
	assert.Contains(t, view, "Add login functionality")
	assert.Contains(t, view, "Integrate payment provider")
	assert.NotContains(t, view, "Export user data to CSV") // Implemented stories should not be shown by default
}

// Test implementation filter toggle
func TestToggleImplementationFilter(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()

	// Toggle the filter to show all stories
	page.state.ToggleImplementationFilter()
	page.updateResults()

	// Get the view
	view := page.View()

	// Check if implemented story is now shown
	assert.Contains(t, view, "Export user data to CSV")
}

// Test search filtering
func TestSearchFiltering(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()

	// Set search text
	page.searchBox = page.searchBox.SetValue("login")
	page.updateResults()

	// Get the view
	view := page.View()

	// Check if only login story is shown
	assert.Contains(t, view, "Add login functionality")
	assert.NotContains(t, view, "Integrate payment provider")
}

// Test focus switching
func TestFocusSwitching(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()

	// Check initial focus is on search (based on NewUIState setting SearchFocused to true)
	assert.True(t, page.state.SearchFocused, "Search should be focused initially")

	// Simulate tab key press to switch to list
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)

	// Check focus switched to list
	assert.False(t, page.state.SearchFocused, "List should be focused after Tab key press")

	// Simulate tab key press again to switch back to search
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)

	// Check focus switched back to search
	assert.True(t, page.state.SearchFocused, "Search should be focused after second Tab key press")
}

// Test selection
func TestSelection(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()

	// Switch focus to list
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)

	// Select an item with space
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeySpace})
	page = model.(*SelectionPage)

	// Check if item is selected
	selected := page.GetSelected()
	assert.GreaterOrEqual(t, len(selected), 1, "At least one item should be selected")
	
	// Verify the selected value in the state
	assert.Equal(t, 1, page.state.SelectedCount(), "One story should be selected")
}

// Test exiting
func TestExiting(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()

	// Simulate escape key press
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyEscape})
	page = model.(*SelectionPage)

	// Check if we're quitting
	assert.True(t, page.quitting)
}

// Test auto-focus first result after search
func TestAutoFocusFirstResultAfterSearch(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()

	// Set search text
	page.searchBox = page.searchBox.SetValue("login")
	page.updateResults()
	
	// Switch to list mode
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = model.(*SelectionPage)
	
	// Get the current item and verify it's the first result
	item, found := page.storyList.CurrentItem()
	assert.True(t, found, "Should have a current item")
	assert.Contains(t, item.Story.Title, "login", "First result should be focused")
	
	// Try another search term
	page.state.FocusSearch()
	page.searchBox = page.searchBox.Focus()
	page.searchBox = page.searchBox.SetValue("payment")
	page.updateResults()
	
	// Switch to list mode again
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = model.(*SelectionPage)
	
	// Verify the first payment-related story is now focused
	item, found = page.storyList.CurrentItem()
	assert.True(t, found, "Should have a current item")
	assert.Contains(t, item.Story.Title, "payment", "First payment result should be focused")
}

// Test clear search filter
func TestClearSearchFilter(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()
	
	// Set search text to filter results
	page.searchBox = page.searchBox.SetValue("login")
	page.updateResults()
	
	// Verify filtering is applied - only login story should be visible
	assert.Equal(t, 1, len(page.state.VisibleStories), "Only one story should be visible")
	assert.Equal(t, "Add login functionality", page.state.VisibleStories[0].Title, "Login story should be visible")
	
	// Simulate pressing Esc to clear the search
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyEscape})
	page = model.(*SelectionPage)
	
	// Verify search text is cleared
	assert.Equal(t, "", page.searchBox.Value(), "Search text should be cleared")
	
	// Verify filter text in state is cleared
	assert.Equal(t, "", page.state.FilterText, "Filter text in state should be cleared")
	
	// Verify unimplemented stories are visible again
	assert.Equal(t, 2, len(page.state.VisibleStories), "All unimplemented stories should be visible")
	
	// Verify search box is still focused
	assert.True(t, page.state.SearchFocused, "Search box should remain focused")
}

// Test persist selections across searches
func TestPersistSelectionsAcrossSearches(t *testing.T) {
	page := New(getTestStories(), true) // Show all stories including implemented ones
	page.Init()
	
	// Switch to list mode
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)
	
	// Select the first story (login)
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeySpace})
	page = model.(*SelectionPage)
	
	// Verify selection state
	firstID := page.stories[0].FilePath
	assert.True(t, page.state.IsSelected(firstID), "First story should be selected")
	
	// Move to second story (payment)
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = model.(*SelectionPage)
	
	// Select the second story (payment)
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeySpace})
	page = model.(*SelectionPage)
	
	// Verify both stories are selected
	secondID := page.stories[1].FilePath
	assert.True(t, page.state.IsSelected(firstID), "First story should still be selected")
	assert.True(t, page.state.IsSelected(secondID), "Second story should be selected")
	assert.Equal(t, 2, page.state.SelectedCount(), "Two stories should be selected")
	
	// Switch to search mode
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)
	
	// Search for "export" to filter out the selected stories
	page.searchBox = page.searchBox.SetValue("export")
	page.updateResults()
	
	// Verify that the selected stories are hidden but still selected
	assert.Equal(t, 1, len(page.state.VisibleStories), "Only export story should be visible")
	assert.Equal(t, 2, page.state.SelectedCount(), "Two stories should still be selected")
	assert.Equal(t, 2, page.state.HiddenSelectedCount(), "Two selected stories should be hidden")
	
	// Clear the search
	page.searchBox = page.searchBox.SetValue("")
	page.updateResults()
	
	// Verify all selections are still maintained
	assert.Equal(t, 2, page.state.SelectedCount(), "All selections should be maintained")
	assert.Equal(t, 0, page.state.HiddenSelectedCount(), "No selections should be hidden anymore")
	assert.True(t, page.state.IsSelected(firstID), "First story should still be selected")
	assert.True(t, page.state.IsSelected(secondID), "Second story should still be selected")
}

// Test show selection count while typing
func TestShowSelectionCountWhileTyping(t *testing.T) {
	page := New(getTestStories(), true) // Show all stories
	page.Init()
	
	// Switch to list mode
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)
	
	// Select the first story
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeySpace})
	page = model.(*SelectionPage)
	
	// Verify selection state
	firstID := page.stories[0].FilePath
	assert.True(t, page.state.IsSelected(firstID), "First story should be selected")
	assert.Equal(t, 1, page.state.SelectedCount(), "One story should be selected")
	
	// Switch back to search mode
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)
	
	// Verify the selection count is maintained
	assert.Equal(t, 1, page.state.SelectedCount(), "Selection count should be maintained after switching to search")
	assert.Equal(t, 0, page.state.HiddenSelectedCount(), "No selected stories should be hidden")
	
	// Type in search box
	page.searchBox = page.searchBox.SetValue("ex")
	page.updateResults()
	
	// Verify the selection count is still maintained while typing
	assert.Equal(t, 1, page.state.SelectedCount(), "Selection count should be maintained while typing")
	assert.Equal(t, 1, page.state.HiddenSelectedCount(), "Selected login story should be hidden")
	
	// Type a search term that doesn't match the selected story
	page.searchBox = page.searchBox.SetValue("export")
	page.updateResults()
	
	// Verify the selection count shows hidden items
	assert.Equal(t, 1, page.state.SelectedCount(), "Selection count should be maintained")
	assert.Equal(t, 1, page.state.HiddenSelectedCount(), "Selected story should be hidden")
	assert.Equal(t, 1, len(page.state.VisibleStories), "Only the export story should be visible")
}

// Test all keys available in the keymap
func TestAllKeybindings(t *testing.T) {
	page := New(getTestStories(), true) // Show all stories
	page.Init()
	
	// Test each key binding in search mode
	keys := []tea.KeyType{
		tea.KeyTab,       // Switch to list
		tea.KeyEnter,     // Confirm
		tea.KeyCtrlA,     // Toggle filter
		tea.KeyCtrlL,     // Clear search
		tea.KeyRunes,     // Type in search
		tea.KeyEscape,    // Clear search or quit
	}
	
	// Test each key in search mode
	for _, k := range keys {
		var msg tea.Msg
		if k == tea.KeyRunes {
			msg = tea.KeyMsg{Type: k, Runes: []rune("test")}
		} else {
			msg = tea.KeyMsg{Type: k}
		}
		
		model, _ := page.Update(msg)
		assert.NotNil(t, model, "Model should not be nil after key press")
	}
	
	// Switch to list mode
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)
	
	// Test each key binding in list mode
	listKeys := []tea.KeyType{
		tea.KeyTab,       // Switch to search
		tea.KeyUp,        // Move up
		tea.KeyDown,      // Move down
		tea.KeyPgUp,      // Page up
		tea.KeyPgDown,    // Page down
		tea.KeySpace,     // Select
		tea.KeyEnter,     // Confirm
		tea.KeyCtrlA,     // Toggle filter
		tea.KeyEscape,    // Quit
	}
	
	// Test each key in list mode
	for _, k := range listKeys {
		model, _ := page.Update(tea.KeyMsg{Type: k})
		// For Esc/Enter, the model will be quitting so we don't need to assert
		if k != tea.KeyEscape && k != tea.KeyEnter {
			assert.NotNil(t, model, "Model should not be nil after key press")
		}
	}
}

// Test window resize handling
func TestWindowResize(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()
	
	// Initial size
	assert.Equal(t, 80, page.width)
	assert.Equal(t, 24, page.height)
	
	// Simulate window resize
	model, _ := page.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	page = model.(*SelectionPage)
	
	// Check new size
	assert.Equal(t, 120, page.width)
	assert.Equal(t, 40, page.height)
	
	// Verify components were resized
	view := page.View()
	assert.NotEmpty(t, view, "View should not be empty after resize")
}

// Test edge case: Consecutive search text changes
func TestConsecutiveSearchTextChanges(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()
	
	// Set initial search
	page.searchBox = page.searchBox.SetValue("l")
	page.updateResults()
	
	// Set another search immediately (cached value optimization should kick in)
	page.searchBox = page.searchBox.SetValue("lo")
	page.updateResults()
	
	// Set another search immediately
	page.searchBox = page.searchBox.SetValue("log")
	page.updateResults()
	
	// Final search
	page.searchBox = page.searchBox.SetValue("login")
	page.updateResults()
	
	// Verify final state
	view := page.View()
	assert.Contains(t, view, "Add login functionality")
	assert.NotContains(t, view, "Integrate payment provider")
}

// Test help toggle
func TestHelpToggle(t *testing.T) {
	page := New(getTestStories(), false)
	page.Init()
	
	// Get initial length
	initialView := page.View()
	initialLines := strings.Split(initialView, "\n")
	initialLineCount := len(initialLines)
	
	// Toggle help off by simulating the ? key
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	page = model.(*SelectionPage)
	
	// Get the view with help toggled off
	toggledView := page.View()
	toggledLines := strings.Split(toggledView, "\n")
	toggledLineCount := len(toggledLines)
	
	// We can't check specific text as it might vary, but after toggling help,
	// the view should be different
	assert.NotEqual(t, initialView, toggledView, "View should change after toggling help")
	
	// Toggle help back on
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	page = model.(*SelectionPage)
	
	// Get the final view
	finalView := page.View()
	finalLines := strings.Split(finalView, "\n")
	finalLineCount := len(finalLines)
	
	// After toggling help back on, we should be back to the initial state
	// (though content might differ due to rendering)
	assert.NotEqual(t, toggledView, finalView, "View should change again after toggling help back on")
	
	// Either the line count or content must be different between toggled states
	assert.True(t, initialLineCount != toggledLineCount || 
		finalLineCount != toggledLineCount || 
		initialView != toggledView || 
		finalView != toggledView,
		"Toggling help should cause a visible difference in the UI")
} 
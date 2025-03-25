// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package pages

import (
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

	// Check initial focus
	assert.True(t, page.state.SearchFocused)

	// Simulate tab key press
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)

	// Check focus switched to list
	assert.False(t, page.state.SearchFocused)

	// Simulate tab key press again
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)

	// Check focus switched back to search
	assert.True(t, page.state.SearchFocused)
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
	assert.Equal(t, 1, len(selected))
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
	
	// Verify filtering is applied
	view := page.View()
	assert.Contains(t, view, "Add login functionality")
	assert.NotContains(t, view, "Integrate payment provider")
	
	// Simulate pressing Esc to clear the search
	model, _ := page.Update(tea.KeyMsg{Type: tea.KeyEscape})
	page = model.(*SelectionPage)
	
	// Verify search is cleared
	assert.Equal(t, "", page.searchBox.Value(), "Search text should be cleared")
	
	// Verify all unimplemented stories are shown again
	view = page.View()
	assert.Contains(t, view, "Add login functionality")
	assert.Contains(t, view, "Integrate payment provider")
	assert.NotContains(t, view, "Export user data to CSV") // Still shouldn't show implemented stories
	
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
	
	// Move to second story (payment)
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = model.(*SelectionPage)
	
	// Select the second story (payment)
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeySpace})
	page = model.(*SelectionPage)
	
	// Verify two stories are selected
	assert.Equal(t, 2, page.state.SelectedCount(), "Two stories should be selected")
	
	// Switch to search mode
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)
	
	// Search for "export" to filter out the selected stories
	page.searchBox = page.searchBox.SetValue("export")
	page.updateResults()
	
	// Verify that the selected stories are hidden but still selected
	assert.Equal(t, 2, page.state.SelectedCount(), "Two stories should still be selected")
	assert.Equal(t, 2, page.state.HiddenSelectedCount(), "Two selected stories should be hidden")
	
	// Check status bar shows hidden selections
	view := page.View()
	assert.Contains(t, view, "2 selected (2 hidden)", "Status bar should show hidden selections")
	
	// Clear the search
	page.searchBox = page.searchBox.SetValue("")
	page.updateResults()
	
	// Verify all selections are still maintained
	assert.Equal(t, 2, page.state.SelectedCount(), "All selections should be maintained")
	assert.Equal(t, 0, page.state.HiddenSelectedCount(), "No selections should be hidden anymore")
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
	
	// Switch back to search mode
	model, _ = page.Update(tea.KeyMsg{Type: tea.KeyTab})
	page = model.(*SelectionPage)
	
	// Verify the selection count is visible in search mode
	view := page.View()
	assert.Contains(t, view, "✔ 1 selected", "Selection count should be visible in search mode")
	
	// Type in search box
	page.searchBox = page.searchBox.SetValue("ex")
	page.updateResults()
	
	// Verify the selection count is still visible while typing
	view = page.View()
	assert.Contains(t, view, "✔ 1 selected", "Selection count should still be visible while typing")
	assert.Contains(t, view, "Filter:", "Filter status should be visible")
	
	// Type a search term that doesn't match the selected story
	page.searchBox = page.searchBox.SetValue("export")
	page.updateResults()
	
	// Verify the selection count shows hidden items
	view = page.View()
	assert.Contains(t, view, "1 selected (1 hidden)", "Should show hidden selection count")
} 
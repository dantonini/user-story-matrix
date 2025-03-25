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
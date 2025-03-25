// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui/pages"
)

// SelectionAdapter adapts the new POM-based selection page to the existing interface
type SelectionAdapter struct {
	page *pages.SelectionPage
}

// NewSelectionAdapter creates a new selection adapter
func NewSelectionAdapter(stories []models.UserStory, showAll bool) *SelectionAdapter {
	return &SelectionAdapter{
		page: pages.New(stories, showAll),
	}
}

// Init initializes the adapter
func (a *SelectionAdapter) Init() tea.Cmd {
	return a.page.Init()
}

// Update handles messages and updates the adapter
func (a *SelectionAdapter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return a.page.Update(msg)
}

// View renders the adapter
func (a *SelectionAdapter) View() string {
	return a.page.View()
}

// GetSelected returns the selected story indices
func (a *SelectionAdapter) GetSelected() []int {
	return a.page.GetSelected()
}

// RegisterNewSelectionUIMaker registers a function to create a new selection UI
// This function allows us to switch between the old and new implementations
func RegisterNewSelectionUIMaker() {
	// Store the current implementation
	OldNewSelectionUI := CurrentNewSelectionUI
	
	// Register the new implementation
	CurrentNewSelectionUI = func(stories []models.UserStory, showAll bool) tea.Model {
		return NewSelectionAdapter(stories, showAll)
	}
	
	// Save the old implementation for reference
	DefaultNewSelectionUI = OldNewSelectionUI
} 
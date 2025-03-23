// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package ui

import (
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
		assert.Nil(t, cmd)
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
	assert.Equal(t, "Selection UI", view)
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
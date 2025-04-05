// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/models"

	// "github.com/user-story-matrix/usm/internal/ui/components/userstoryform" - Will be used in MVI phase
	"github.com/user-story-matrix/usm/internal/ui/pages"
)

// CurrentNewSelectionUI is a function type for creating a selection UI
var CurrentNewSelectionUI = func(stories []models.UserStory, showAll bool) tea.Model {
	return NewSelectionAdapter(stories, showAll)
}

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
	model, cmd := a.page.Update(msg)
	if pageModel, ok := model.(*pages.SelectionPage); ok {
		a.page = pageModel
		return a, cmd
	}
	return a, cmd
}

// View renders the adapter
func (a *SelectionAdapter) View() string {
	return a.page.View()
}

// GetSelected returns the selected story indices
func (a *SelectionAdapter) GetSelected() []int {
	return a.page.GetSelected()
}

// RegisterNewSelectionUIMaker registers the new selection UI implementation
// For backward compatibility - this function now does nothing since we
// permanently use the new implementation
func RegisterNewSelectionUIMaker() {
	// The new implementation is already set as default in CurrentNewSelectionUI
}

// CreateUserStoryFormWithLLM creates a new user story form with LLM processing support
// TODO: This is a placeholder implementation for the foundation phase
// In the MVI phase, this will be replaced with a real implementation using
// the userstoryform package and LLM processor
func CreateUserStoryFormWithLLM(us models.UserStory, enableLLM bool) tea.Model {
	// For the foundation phase, just return the original form
	// In MVI phase, this will be replaced with the enhanced form 
	// that uses LLM processing
	return io.NewUserStoryForm(us)
}

// FormatUIMessage formats a UI message with style
func FormatUIMessage(message string, style string) string {
	switch style {
	case "success":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(fmt.Sprintf("✓ %s", message))
	case "error":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("✗ %s", message))
	case "warning":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(fmt.Sprintf("! %s", message))
	case "info":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(fmt.Sprintf("ℹ %s", message))
	default:
		return message
	}
} 
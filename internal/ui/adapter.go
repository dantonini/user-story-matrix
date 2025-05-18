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
	"github.com/user-story-matrix/usm/internal/llm"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui/components/userstoryform"
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

// fileSystemFactory is a function to create a file system
// This can be overridden in tests to inject a mock
var fileSystemFactory = func() io.FileSystem {
	return io.NewOSFileSystem()
}

// Helper function to explicitly ignore errors when we know they're non-critical
// This is used to satisfy the linter while documenting our intent
func ignoreError(err error) {
	// Intentionally empty - we're explicitly ignoring the error
}

// CreateUserStoryFormWithLLM creates a new user story form with LLM processing support
func CreateUserStoryFormWithLLM(us models.UserStory) tea.Model {
	// Create the filesystem for configuration
	fs := fileSystemFactory()
	
	// Create the configuration manager
	configManager := llm.NewConfigManager(fs)
	
	// Load the configuration - errors are expected if config doesn't exist yet
	// and will be handled gracefully by the processor
	ignoreError(configManager.LoadConfig())
	
	// Create the LLM processor regardless of whether an API key exists
	// This ensures the processor is always properly initialized
	processor := llm.NewOpenAIProcessor(
		llm.WithAPIKey(configManager.GetOpenAIKey()),
		llm.WithModel("gpt-4o-mini"),
		llm.WithMaxTokens(1000),
		llm.WithTemperature(0.2),
	)
	
	// No need to check API key configuration here - the form handling code will
	// check processor.IsConfigured() when needed
	return userstoryform.New(us, processor, configManager)
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
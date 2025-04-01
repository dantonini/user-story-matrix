// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/search"
)

// Styles for the UI components
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")). // Bright white
			Background(lipgloss.Color("63")). // Bright blue
			Bold(true)

	implementedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")). // Bright white
			Background(lipgloss.Color("25")). // Blue background
			Bold(true).
			Width(100).
			Padding(0, 1)

	searchBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).   // Use rounded borders for visibility
			BorderForeground(lipgloss.Color("205")). // Brighter border
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1).
			Width(100)

	// Adding a more prominent label style for the search box
	searchLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				MarginBottom(0).
				MarginTop(1)

	// NewSelectionUIFunc is a function type for creating a new selection UI
	NewSelectionUIFunc func(stories []models.UserStory, showAll bool) tea.Model

	// DefaultNewSelectionUI is the default implementation of NewSelectionUIFunc
	DefaultNewSelectionUI = func(stories []models.UserStory, showAll bool) tea.Model {
		return NewSelectionUI(stories, showAll)
	}

	// CurrentNewSelectionUI is the current implementation of NewSelectionUIFunc
	CurrentNewSelectionUI = DefaultNewSelectionUI
)

// KeyMap defines the keybindings for the selection UI
type KeyMap struct {
	Select    key.Binding
	Done      key.Binding
	Quit      key.Binding
	ToggleAll key.Binding
	Search    key.Binding // New keybinding for search
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select/deselect"),
		),
		Done: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm selection"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "toggle all/unimplemented"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
	}
}

// storyItem represents a user story list item
type storyItem struct {
	story      models.UserStory
	index      int
	isSelected bool
}

func (i storyItem) Title() string {
	// Define cursor - show an arrow for the selected item in the list
	var cursor string
	if i.isSelected {
		checkbox := "[✓]"
		cursor = checkbox + " "
	} else {
		checkbox := "[ ]"
		cursor = checkbox + " "
	}

	title := i.story.Title

	// Add file path for disambiguation if available, trimming docs/user-stories prefix
	if i.story.FilePath != "" {
		path := i.story.FilePath
		// Remove docs/user-stories prefix if present
		path = strings.TrimPrefix(path, "docs/user-stories/")
		pathInfo := implementedStyle.Render(" (" + path + ")")
		title = title + pathInfo
	}

	// Different styling based on selection and implementation status
	if i.isSelected {
		title = selectedStyle.Render(title)
	} else if i.story.IsImplemented {
		title = implementedStyle.Render(title + " [implemented]")
	}

	return cursor + title
}

func (i storyItem) Description() string {
	// Return empty string to hide description section
	return ""
}

func (i storyItem) FilterValue() string {
	return i.story.Title + " " + i.story.Description + " " + i.story.FilePath
}

// SelectionUI represents the enhanced selection UI
type SelectionUI struct {
	searchBox     textinput.Model
	storyList     list.Model
	statusBar     string
	engine        *search.Engine
	ready         bool
	selected      []int
	keyMap        KeyMap
	width         int
	height        int
	stories       []models.UserStory
	showAll       bool
	quitting      bool
	searchFocused bool // Track whether search box is focused
}

// NewSelectionUI creates a new selection UI
func NewSelectionUI(stories []models.UserStory, showAll bool) *SelectionUI {
	// Create search box
	searchBox := textinput.New()
	searchBox.Placeholder = "Type to search user stories..."
	searchBox.CharLimit = 100
	searchBox.Width = 50 // Make it wider to be more visible
	// Make the cursor more visible
	searchBox.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("236")).
		Bold(true)
	// Set a prompt to make it clearer this is a search box
	searchBox.Prompt = "🔍 "
	// Make the text more visible
	searchBox.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")). // Bright pink
		Bold(true)
	// Make placeholder text visible but subdued
	searchBox.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Gray
		Italic(true)
	// Change style when focused
	searchBox.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")). // Bright pink
		Bold(true)
	// Start with search box blurred (list mode)
	searchBox.Blur()

	// Create search engine
	engine := search.NewEngine(stories)
	engine.SetShowAll(showAll)

	// Setup list
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0) // Remove space between items
	delegate.SetHeight(1)  // Use height of 1 to avoid division by zero

	// Customize the selection indicator to make it more visible
	delegate.Styles.SelectedTitle = selectedStyle.Copy().Background(lipgloss.Color("63")).Foreground(lipgloss.Color("15"))
	delegate.Styles.SelectedDesc = selectedStyle

	// Set a custom style for normal items to ensure they're visible
	delegate.Styles.NormalTitle = lipgloss.NewStyle()

	filteredStories := engine.Filter("")

	// Create list items
	items := make([]list.Item, len(filteredStories))
	for i, story := range filteredStories {
		// Find the original index in the stories array
		originalIndex := -1
		for j, s := range stories {
			if s.Title == story.Title {
				originalIndex = j
				break
			}
		}

		items[i] = storyItem{
			story:      story,
			index:      originalIndex, // Use original index for consistent selection
			isSelected: false,
		}
	}

	// Set initial size to non-zero values to prevent divide-by-zero error
	storyList := list.New(items, delegate, 30, 20) // Use safe default dimensions
	storyList.Title = ""                           // Remove "User Stories" title
	storyList.SetShowStatusBar(false)              // Hide the status bar with "x items"
	storyList.SetFilteringEnabled(false)           // We'll handle filtering ourselves
	storyList.SetShowHelp(false)                   // Hide the navigation help
	storyList.SetShowPagination(true)              // Show pagination to help with navigation
	storyList.DisableQuitKeybindings()

	// Initialize the UI with a clear status message about filter mode
	initialStatus := "MODE: LIST"
	if showAll {
		initialStatus += " | FILTER: ALL STORIES SHOWN"
	} else {
		initialStatus += " | FILTER: UNIMPLEMENTED ONLY"
	}

	// Initialize the UI
	ui := &SelectionUI{
		searchBox:     searchBox,
		storyList:     storyList,
		engine:        engine,
		statusBar:     initialStatus,
		ready:         false,
		selected:      []int{},
		keyMap:        DefaultKeyMap(),
		stories:       stories,
		showAll:       showAll,
		searchFocused: true, // Start with search mode active by default
		width:         80,   // Ensure width is non-zero initially
		height:        25,   // Ensure height is non-zero initially
	}

	return ui
}

// Init initializes the selection UI
func (ui *SelectionUI) Init() tea.Cmd {
	// Initial update to ensure the list is populated
	ui.updateList()

	// Select the first item to make sure cursor is visible initially
	if len(ui.storyList.Items()) > 0 {
		ui.storyList.Select(0)
	}

	// Start with search box focused
	ui.searchFocused = true
	ui.searchBox.Focus()
	ui.statusBar = "SEARCH MODE: Type to filter stories, press Enter when done"

	return textinput.Blink
}

// Update updates the UI based on messages
func (ui *SelectionUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check for the dedicated search key (/) to enter search mode
		if key.Matches(msg, ui.keyMap.Search) && !ui.searchFocused {
			ui.searchFocused = true
			ui.searchBox.Focus()
			ui.statusBar = "SEARCH MODE: Type to filter stories, press Enter when done"
			ui.updateList() // Update the list to reflect the new mode
			return ui, textinput.Blink
		}

		// First, check if any typing key is pressed when not in search mode
		// to automatically enter search mode
		if !ui.searchFocused {
			// If a letter, number, or common symbol is typed, enter search mode
			if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
				// Auto-switch to search mode on typing
				ui.searchFocused = true
				ui.searchBox.Focus()
				ui.statusBar = "SEARCH MODE: Type to filter stories, press Enter when done"

				// Forward the typed key to the search box
				newSearchBox, searchBoxCmd := ui.searchBox.Update(msg)
				ui.searchBox = newSearchBox
				if ui.searchBox.Value() != ui.engine.GetState().SearchQuery {
					ui.updateList()
				}
				return ui, searchBoxCmd
			}
		}

		// Handle global key presses
		switch {
		case key.Matches(msg, ui.keyMap.Quit):
			ui.quitting = true
			return ui, tea.Quit

		case key.Matches(msg, ui.keyMap.Done):
			// If in search mode, Enter returns to list mode
			if ui.searchFocused {
				ui.searchFocused = false
				ui.searchBox.Blur()
				ui.statusBar = "LIST MODE: Use arrow keys to navigate, space to select"
				return ui, nil
			}
			// Otherwise, Enter completes the selection
			return ui, tea.Quit

		case key.Matches(msg, ui.keyMap.Select):
			// If in search mode and space is pressed, handle it as a space character
			if ui.searchFocused && msg.String() == " " {
				// Forward space key to search box
				newSearchBox, searchBoxCmd := ui.searchBox.Update(msg)
				ui.searchBox = newSearchBox
				if ui.searchBox.Value() != ui.engine.GetState().SearchQuery {
					ui.updateList()
				}
				return ui, searchBoxCmd
			}

			// Only handle selection when in list mode
			if !ui.searchFocused && ui.storyList.SelectedItem() != nil {
				item := ui.storyList.SelectedItem().(storyItem)
				// Check if the item is already selected
				found := false
				for i, idx := range ui.selected {
					if idx == item.index {
						// Remove from selection
						ui.statusBar = fmt.Sprintf("Deselected: %s", item.story.Title)
						ui.selected = append(ui.selected[:i], ui.selected[i+1:]...)
						found = true
						break
					}
				}
				if !found {
					// Add to selection
					ui.statusBar = fmt.Sprintf("Selected: %s", item.story.Title)
					ui.selected = append(ui.selected, item.index)
				}

				// Update the item in the list
				items := ui.storyList.Items()
				for i, listItem := range items {
					currentItem := listItem.(storyItem)
					if currentItem.index == item.index {
						currentItem.isSelected = !found
						items[i] = currentItem
						break
					}
				}
				ui.storyList.SetItems(items)

				// Update status bar with selection status
				if len(ui.selected) > 0 {
					ui.statusBar = fmt.Sprintf("%s | %d stories selected", ui.statusBar, len(ui.selected))
				}
			}
			return ui, nil // Return early to prevent list handling this key event

		case key.Matches(msg, ui.keyMap.ToggleAll):
			// Toggle between showing all stories and only unimplemented ones
			ui.showAll = !ui.showAll
			ui.engine.SetShowAll(ui.showAll)

			// Update the filter display with animation
			if ui.showAll {
				ui.statusBar = "✓ SHOWING ALL STORIES (including implemented)"
			} else {
				ui.statusBar = "✓ SHOWING UNIMPLEMENTED STORIES ONLY"
			}

			// Update the list to reflect the new filter
			ui.updateList()

			return ui, nil // Return early to prevent list handling this key event

		// Escape always exits search mode
		case msg.Type == tea.KeyEsc:
			if ui.searchFocused {
				ui.searchFocused = false
				ui.searchBox.Blur()
				ui.statusBar = "LIST MODE: Use arrow keys to navigate, space to select"
				return ui, nil
			}
		}

	case tea.WindowSizeMsg:
		// Handle window resize
		ui.width = msg.Width
		ui.height = msg.Height
		ui.ready = true

		// Update searchbox width (max 80% of window width)
		maxWidth := int(float64(msg.Width) * 0.8)
		if maxWidth < 30 {
			maxWidth = msg.Width - 4 // For very small windows
		}
		ui.searchBox.Width = maxWidth

		// Update list height, accounting for search box height
		searchBoxHeight := 8     // Search section with borders and padding
		statusBarHeight := 2     // Status bar + spacing
		helpHeight := 2          // Help text + spacing
		selectionInfoHeight := 2 // Selected count + spacing

		listHeight := msg.Height - searchBoxHeight - statusBarHeight - helpHeight - selectionInfoHeight
		if listHeight < 5 {
			listHeight = 5 // Ensure a minimum visibility
		}

		// Update the list size
		ui.storyList.SetSize(msg.Width, listHeight)
	}

	// Handle searchbox input if it's focused
	if ui.searchFocused && !ui.quitting {
		newSearchBox, searchBoxCmd := ui.searchBox.Update(msg)
		ui.searchBox = newSearchBox

		// If the search query changed, update the list
		if ui.searchBox.Value() != ui.engine.GetState().SearchQuery {
			ui.updateList()
		}

		cmds = append(cmds, searchBoxCmd)
	} else {
		// Handle list input if we're not focused on the search box
		newList, listCmd := ui.storyList.Update(msg)
		ui.storyList = newList
		cmds = append(cmds, listCmd)
	}

	return ui, tea.Batch(cmds...)
}

// updateList updates the story list based on current filter settings
func (ui *SelectionUI) updateList() {
	// Get filtered stories
	filteredStories := ui.engine.Filter(ui.searchBox.Value())

	// Create new items
	items := make([]list.Item, len(filteredStories))
	for i, story := range filteredStories {
		// Check if this story is selected
		isSelected := false
		originalIndex := -1 // Store the original index in the full stories array

		// Find the story's original index by title (more reliable than position)
		for j, s := range ui.stories {
			if s.Title == story.Title {
				originalIndex = j
				break
			}
		}

		// Check if the original index is in the selected slice
		for _, idx := range ui.selected {
			if idx == originalIndex {
				isSelected = true
				break
			}
		}

		items[i] = storyItem{
			story:      story,
			index:      originalIndex, // Use the original index for selection
			isSelected: isSelected,
		}
	}

	ui.storyList.SetItems(items)

	// Update filter status
	state := ui.engine.GetState()

	// Show number of filtered stories more prominently
	filterStatus := fmt.Sprintf("FILTER: %d/%d stories", state.FilteredCount, state.TotalCount)

	// Check if we have an active search query and show it
	if ui.searchBox.Value() != "" {
		filterStatus = fmt.Sprintf("FILTER: %d/%d stories matching '%s'",
			state.FilteredCount, state.TotalCount, ui.searchBox.Value())
	}

	// Add a clear notification when no results match the filter
	if state.FilteredCount == 0 && ui.searchBox.Value() != "" {
		filterStatus = fmt.Sprintf("NO RESULTS FOUND for '%s' - try a different search", ui.searchBox.Value())
	}

	// Add implementation filter status
	if !state.ShowAll {
		filterStatus += " [ UNIMPLEMENTED ONLY ]"
	} else {
		filterStatus += " [ ALL STORIES ]"
	}

	// Update status bar based on mode, but always include filter info
	if ui.searchFocused {
		ui.statusBar = fmt.Sprintf("MODE: SEARCH | %s", filterStatus)
	} else {
		ui.statusBar = fmt.Sprintf("MODE: LIST | %s", filterStatus)
	}
}

// View renders the UI
func (ui *SelectionUI) View() string {
	if !ui.ready {
		return "Initializing..."
	}

	// Build the UI
	var b strings.Builder

	// Status bar showing mode and filter info at the very top
	statusBarText := ui.statusBar
	if ui.width > 0 { // Make sure we have a valid width
		statusBarStyle = statusBarStyle.Copy().Width(ui.width)
	}
	b.WriteString(statusBarStyle.Render(statusBarText) + "\n\n")

	// Calculate search section width based on terminal width
	searchWidth := ui.width - 4
	if searchWidth < 20 {
		searchWidth = ui.width // For very small terminals, use full width
	}

	// IMPORTANT: Make search section more visible with background highlight
	// For small terminals, adjust the layout to be simpler
	var searchSectionContent string
	if ui.width < 50 {
		// Simplified search section for small terminals
		searchSectionContent = searchLabelStyle.Render("🔍 SEARCH:") + "\n" + ui.searchBox.View()
	} else {
		// Full search section for larger terminals
		searchSectionContent = searchLabelStyle.Render("🔍 TYPE TO SEARCH OR PRESS '/' TO SEARCH") +
			"\n\n" + ui.searchBox.View()
	}

	searchSection := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).       // Dark background
		Border(lipgloss.NormalBorder()).         // Add border
		BorderForeground(lipgloss.Color("205")). // Pink border
		Padding(1, 2).                           // Add padding
		Margin(1, 0, 1, 0).                      // Add margin top/bottom
		Width(searchWidth).                      // Use calculated width
		Align(lipgloss.Center).                  // Center the content
		Render(searchSectionContent)

	b.WriteString(searchSection + "\n")

	// Story list right after search box
	b.WriteString(ui.storyList.View())

	// Footer
	selected := fmt.Sprintf("\nSelected: %d stories", len(ui.selected))
	b.WriteString(selected)

	// Help text that changes based on mode
	var help string
	if ui.searchFocused {
		help = "\nEnter/Esc: Return to list | Ctrl+A: Toggle All/Unimplemented | Ctrl+C: Quit"
	} else {
		help = "\nSpace: Select/Deselect | / to search | Enter: Done | Ctrl+A: Toggle All/Unimplemented | Ctrl+C: Quit"
	}
	b.WriteString(help)

	return b.String()
}

// GetSelected returns the selected stories
func (ui *SelectionUI) GetSelected() []int {
	return ui.selected
}

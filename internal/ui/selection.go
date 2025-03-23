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
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("240"))

	implementedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	statusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("238")).
		Padding(0, 1)

	searchBoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("241")).
		Padding(0, 1)

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
	Select   key.Binding
	Done     key.Binding
	Quit     key.Binding
	ToggleAll key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Select: key.NewBinding(
			key.WithKeys("ctrl+space"),
			key.WithHelp("ctrl+space", "select/deselect"),
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
	}
}

// storyItem represents a user story list item
type storyItem struct {
	story      models.UserStory
	index      int
	isSelected bool
}

func (i storyItem) Title() string {
	title := i.story.Title
	if i.story.IsImplemented {
		title = implementedStyle.Render(title + " [implemented]")
	}
	if i.isSelected {
		return selectedStyle.Render(title)
	}
	return title
}

func (i storyItem) Description() string {
	return i.story.Description
}

func (i storyItem) FilterValue() string {
	return i.story.Title + " " + i.story.Description
}

// SelectionUI represents the enhanced selection UI
type SelectionUI struct {
	searchBox    textinput.Model
	storyList    list.Model
	statusBar    string
	engine       *search.Engine
	ready        bool
	selected     []int
	keyMap       KeyMap
	width        int
	height       int
	stories      []models.UserStory
	showAll      bool
	quitting     bool
}

// NewSelectionUI creates a new selection UI
func NewSelectionUI(stories []models.UserStory, showAll bool) *SelectionUI {
	// Create search box
	searchBox := textinput.New()
	searchBox.Placeholder = "Type to search..."
	searchBox.Focus()
	searchBox.CharLimit = 100
	searchBox.Width = 30

	// Create search engine
	engine := search.NewEngine(stories)
	engine.SetShowAll(showAll)

	// Setup list
	delegate := list.NewDefaultDelegate()
	filteredStories := engine.Filter("")
	
	// Create list items
	items := make([]list.Item, len(filteredStories))
	for i, story := range filteredStories {
		items[i] = storyItem{
			story: story,
			index: i,
			isSelected: false,
		}
	}
	
	storyList := list.New(items, delegate, 0, 0)
	storyList.Title = "User Stories"
	storyList.SetShowStatusBar(true)
	storyList.SetFilteringEnabled(false) // We'll handle filtering ourselves
	storyList.SetShowHelp(true)
	storyList.SetShowPagination(true)
	storyList.DisableQuitKeybindings()

	// Initialize the UI
	ui := &SelectionUI{
		searchBox: searchBox,
		storyList: storyList,
		engine:    engine,
		statusBar: "Ready",
		ready:     false,
		selected:  []int{},
		keyMap:    DefaultKeyMap(),
		stories:   stories,
		showAll:   showAll,
	}
	
	return ui
}

// Init initializes the selection UI
func (ui *SelectionUI) Init() tea.Cmd {
	return textinput.Blink
}

// Update updates the UI based on messages
func (ui *SelectionUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global key presses
		switch {
		case key.Matches(msg, ui.keyMap.Quit):
			ui.quitting = true
			return ui, tea.Quit
		
		case key.Matches(msg, ui.keyMap.Done):
			return ui, tea.Quit

		case key.Matches(msg, ui.keyMap.Select):
			// Toggle selection of the current item
			if ui.storyList.SelectedItem() != nil {
				item := ui.storyList.SelectedItem().(storyItem)
				// Check if the item is already selected
				found := false
				for i, idx := range ui.selected {
					if idx == item.index {
						// Remove from selection
						ui.selected = append(ui.selected[:i], ui.selected[i+1:]...)
						found = true
						break
					}
				}
				if !found {
					// Add to selection
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
			}

		case key.Matches(msg, ui.keyMap.ToggleAll):
			// Toggle between showing all stories and only unimplemented ones
			ui.showAll = !ui.showAll
			ui.engine.SetShowAll(ui.showAll)
			ui.updateList()
		}
	
	case tea.WindowSizeMsg:
		// Handle window resize
		ui.width = msg.Width
		ui.height = msg.Height
		ui.ready = true
		
		// Update searchbox width
		ui.searchBox.Width = msg.Width - 10
		
		// Update list height
		headerHeight := 3 // Title + status bar + help
		searchBoxHeight := 3 // Input + padding
		ui.storyList.SetSize(msg.Width, msg.Height-headerHeight-searchBoxHeight)
	}
	
	// Handle searchbox input
	if !ui.quitting {
		newSearchBox, searchBoxCmd := ui.searchBox.Update(msg)
		ui.searchBox = newSearchBox
		
		// If the search query changed, update the list
		if ui.searchBox.Value() != ui.engine.GetState().SearchQuery {
			ui.updateList()
		}
		
		cmds = append(cmds, searchBoxCmd)
	}
	
	// Handle list input if we're not focused on the search box
	newList, listCmd := ui.storyList.Update(msg)
	ui.storyList = newList
	cmds = append(cmds, listCmd)
	
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
		for _, idx := range ui.selected {
			if idx == i {
				isSelected = true
				break
			}
		}
		
		items[i] = storyItem{
			story:      story,
			index:      i,
			isSelected: isSelected,
		}
	}
	
	ui.storyList.SetItems(items)
	
	// Update status bar
	state := ui.engine.GetState()
	filterStatus := fmt.Sprintf("Showing %d of %d stories", state.FilteredCount, state.TotalCount)
	if state.ShowAll {
		filterStatus += " (including implemented)"
	} else {
		filterStatus += " (unimplemented only)"
	}
	ui.statusBar = filterStatus
}

// View renders the UI
func (ui *SelectionUI) View() string {
	if !ui.ready {
		return "Initializing..."
	}
	
	// Build the UI
	var b strings.Builder
	
	// Search box
	b.WriteString(searchBoxStyle.Render(ui.searchBox.View()) + "\n\n")
	
	// Status bar
	b.WriteString(statusBarStyle.Render(ui.statusBar) + "\n")
	
	// Story list
	b.WriteString(ui.storyList.View())
	
	// Footer
	selected := fmt.Sprintf("\nSelected: %d stories", len(ui.selected))
	b.WriteString(selected)
	
	// Help text for commands
	help := "\nCtrl+Space: Select/Deselect | Enter: Done | Ctrl+A: Toggle All/Unimplemented | Ctrl+C: Quit"
	b.WriteString(help)
	
	return b.String()
}

// GetSelected returns the selected stories
func (ui *SelectionUI) GetSelected() []int {
	return ui.selected
}
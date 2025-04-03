// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package pages

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/search"
	"github.com/user-story-matrix/usm/internal/ui/components/searchbox"
	"github.com/user-story-matrix/usm/internal/ui/components/statusbar"
	"github.com/user-story-matrix/usm/internal/ui/components/storylist"
	uimodels "github.com/user-story-matrix/usm/internal/ui/models"
	"github.com/user-story-matrix/usm/internal/ui/styles"
)

// SelectionPage represents the main user story selection page
type SelectionPage struct {
	// Components
	searchBox searchbox.SearchBox
	storyList storylist.StoryList
	statusBar statusbar.StatusBar
	
	// State
	state      *uimodels.UIState
	keyMap     uimodels.KeyMap
	styles     *styles.Styles
	
	// Data
	stories    []models.UserStory
	engine     *search.Engine
	
	// UI state
	width      int
	height     int
	quitting   bool
	ready      bool
	
	// Cache fields for performance
	lastView   string
	needsRender bool
	lastSearchValue string
}

// New creates a new selection page
func New(stories []models.UserStory, showAll bool) *SelectionPage {
	if stories == nil {
		stories = []models.UserStory{} // Convert nil to empty slice for safety
	}
	
	// Create state
	state := uimodels.NewUIState()
	state.ShowImplemented = showAll
	
	// Create search engine
	engine := search.NewEngine(stories)
	engine.SetShowAll(showAll)
	
	// Create styles
	styleSet := styles.DefaultStyles()
	
	// Create key map
	keyMap := uimodels.DefaultKeyMap()
	
	// Create components
	searchbox := searchbox.New(styleSet)
	storylist := storylist.New(styleSet)
	statusbar := statusbar.New(styleSet, keyMap)
	
	// Set initial focus
	if state.SearchFocused {
		searchbox = searchbox.Focus()
		storylist = storylist.Blur()
	} else {
		searchbox = searchbox.Blur()
		storylist = storylist.Focus()
	}
	
	return &SelectionPage{
		searchBox: searchbox,
		storyList: storylist,
		statusBar: statusbar,
		state:     state,
		keyMap:    keyMap,
		styles:    styleSet,
		stories:   stories,
		engine:    engine,
		width:     80,
		height:    24,
		quitting:  false,
		ready:     true,
		needsRender: true,
	}
}

// Init initializes the page
func (p *SelectionPage) Init() tea.Cmd {
	// Start with the search box focused
	return p.updateResults()
}

// updateResults updates the filtered results based on the current search text
func (p *SelectionPage) updateResults() tea.Cmd {
	// Get the current search text
	searchText := p.searchBox.Value()
	
	// Skip updating if search text hasn't changed and filter hasn't changed
	if searchText == p.lastSearchValue && !p.needsRender {
		return nil
	}
	
	p.lastSearchValue = searchText
	
	// Update the state
	p.state.SetFilterText(searchText)
	
	// Set the show all flag in the engine
	p.engine.SetShowAll(p.state.ShowImplemented)
	
	// Get filtered stories
	filtered := p.engine.Filter(searchText)
	
	// Update visible stories in state
	p.state.SetVisibleStories(filtered, len(p.stories))
	
	// Update story list
	p.storyList = p.storyList.SetItems(filtered, p.state.SelectedIDs)
	
	// Ensure the first item is focused if there are any results
	if len(filtered) > 0 && p.state.CursorPosition != 0 {
		// Set cursor to the first item
		p.storyList = p.storyList.SetCursor(0)
	}
	
	p.needsRender = true
	
	return nil
}

// GetSelected returns the indices of the selected stories
func (p *SelectionPage) GetSelected() []int {
	return p.state.GetSelectedStoryIndices(p.stories)
}

// Update handles messages and updates the page
func (p *SelectionPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle window resize
		p.width = msg.Width
		p.height = msg.Height
		p.ready = true
		p.needsRender = true
		
		// Update component sizes
		p.searchBox = p.searchBox.SetWidth(msg.Width - 4)
		p.storyList = p.storyList.SetSize(msg.Width, msg.Height-10) // Adjust for search box and status bar
		p.statusBar = p.statusBar.SetWidth(msg.Width)
		
	case tea.KeyMsg:
		// Handle key presses
		switch {
		case p.state.SearchFocused:
			// Handle search mode key bindings
			switch {
			case key.Matches(msg, p.keyMap.Quit):
				// In search mode, Esc clears the search or exits if already empty
				if p.searchBox.Value() != "" {
					// Clear the search text
					p.searchBox = p.searchBox.SetValue("")
					p.needsRender = true
					
					// Update the results with empty filter
					cmds = append(cmds, p.updateResults())
					
					// Keep focus in search box
					return p, tea.Batch(cmds...)
				} else {
					// If search is already empty, quit the application
					p.quitting = true
					p.needsRender = true
					return p, tea.Quit
				}
			
			case key.Matches(msg, p.keyMap.Tab):
				// Switch to list mode
				p.state.FocusList()
				p.searchBox = p.searchBox.Blur()
				p.storyList = p.storyList.Focus()
				p.needsRender = true
				
			case key.Matches(msg, p.keyMap.Done):
				// Apply search and switch to list mode
				p.state.FocusList()
				p.searchBox = p.searchBox.Blur()
				p.storyList = p.storyList.Focus()
				p.needsRender = true
				
			case key.Matches(msg, p.keyMap.ToggleFilter):
				// Toggle implementation filter
				p.state.ToggleImplementationFilter()
				p.needsRender = true
				cmds = append(cmds, p.updateResults())
				
			case key.Matches(msg, p.keyMap.Clear):
				// Clear search text
				p.searchBox = p.searchBox.SetValue("")
				p.needsRender = true
				cmds = append(cmds, p.updateResults())
				
			case key.Matches(msg, p.keyMap.Help):
				// Toggle help display
				p.statusBar = p.statusBar.ToggleHelp()
				p.needsRender = true
				
			default:
				// Update search box
				var cmd tea.Cmd
				p.searchBox, cmd = p.searchBox.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				
				// Update results if search text changed
				if p.state.FilterText != p.searchBox.Value() {
					cmds = append(cmds, p.updateResults())
					p.needsRender = true
				}
			}
		
		case !p.state.SearchFocused:
			// Handle list mode key bindings
			switch {
			case key.Matches(msg, p.keyMap.Quit):
				// Quit the application
				p.quitting = true
				p.needsRender = true
				return p, tea.Quit
				
			case key.Matches(msg, p.keyMap.Tab), key.Matches(msg, p.keyMap.Search):
				// Switch to search mode
				p.state.FocusSearch()
				p.searchBox = p.searchBox.Focus()
				p.storyList = p.storyList.Blur()
				p.needsRender = true
				
			case key.Matches(msg, p.keyMap.Select):
				// Toggle selection of current item
				var id string
				p.storyList, id = p.storyList.ToggleSelection()
				if id != "" {
					p.state.ToggleSelection(id)
					p.needsRender = true
				}
				
			case key.Matches(msg, p.keyMap.Up):
				// Move cursor up
				p.storyList = p.storyList.MoveUp()
				p.needsRender = true
				
			case key.Matches(msg, p.keyMap.Down):
				// Move cursor down
				p.storyList = p.storyList.MoveDown()
				p.needsRender = true
				
			case key.Matches(msg, p.keyMap.PageUp):
				// Page up
				p.storyList = p.storyList.PageUp()
				p.needsRender = true
				
			case key.Matches(msg, p.keyMap.PageDown):
				// Page down
				p.storyList = p.storyList.PageDown()
				p.needsRender = true
				
			case key.Matches(msg, p.keyMap.ToggleFilter):
				// Toggle implementation filter
				p.state.ToggleImplementationFilter()
				p.needsRender = true
				cmds = append(cmds, p.updateResults())
				
			case key.Matches(msg, p.keyMap.Help):
				// Toggle help display
				p.statusBar = p.statusBar.ToggleHelp()
				p.needsRender = true
				
			case key.Matches(msg, p.keyMap.Done):
				// Complete selection
				p.quitting = true
				p.needsRender = true
				return p, tea.Quit
			}
		}
	}
	
	// Return the updated model
	return p, tea.Batch(cmds...)
}

// View renders the page
func (p *SelectionPage) View() string {
	if !p.ready {
		return "Loading..."
	}
	
	if p.quitting {
		return "Change request creation canceled by user."
	}
	
	// If nothing has changed, return the cached view
	if !p.needsRender && p.lastView != "" {
		return p.lastView
	}
	
	var sb strings.Builder
	
	// Render search box
	sb.WriteString(p.searchBox.View())
	sb.WriteString("\n")
	
	// Render divider
	divider := strings.Repeat("─", p.width)
	sb.WriteString(divider)
	sb.WriteString("\n")
	
	// Render story list
	listView := p.storyList.View()
	if len(p.state.VisibleStories) == 0 {
		// Show no results message
		noResults := p.styles.Error.Render("⚠️  No matching user stories found.")
		sb.WriteString(noResults)
	} else {
		sb.WriteString(listView)
	}
	sb.WriteString("\n")
	
	// Render divider
	sb.WriteString(divider)
	sb.WriteString("\n")
	
	// Render status bar
	sb.WriteString(p.statusBar.View(p.state))
	
	// Cache the view
	p.lastView = sb.String()
	p.needsRender = false
	
	return p.lastView
} 
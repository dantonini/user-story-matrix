// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package searchbox

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/user-story-matrix/usm/internal/ui/styles"
)

// SearchBox represents a search input component
type SearchBox struct {
	textInput textinput.Model
	styles    *styles.Styles
	focused   bool
	width     int
}

// New creates a new SearchBox component
func New(styles *styles.Styles) SearchBox {
	ti := textinput.New()
	ti.Placeholder = "Type to search user stories..."
	ti.CharLimit = 100
	ti.Width = 50
	
	// Configure cursor style
	ti.Cursor.Style = styles.SearchCursor
	
	// Configure text style
	ti.TextStyle = styles.SearchText
	
	// Configure placeholder style
	ti.PlaceholderStyle = styles.SearchPlaceholder
	
	// Add prompt emoji
	ti.Prompt = "🔍 "
	ti.PromptStyle = styles.SearchText
	
	// Start unfocused
	ti.Blur()
	
	return SearchBox{
		textInput: ti,
		styles:    styles,
		focused:   false,
		width:     50,
	}
}

// Focus focuses the search box
func (s SearchBox) Focus() SearchBox {
	s.focused = true
	s.textInput.Focus()
	return s
}

// Blur blurs the search box
func (s SearchBox) Blur() SearchBox {
	s.focused = false
	s.textInput.Blur()
	return s
}

// SetValue sets the search box value
func (s SearchBox) SetValue(value string) SearchBox {
	s.textInput.SetValue(value)
	return s
}

// Value returns the current value of the search box
func (s SearchBox) Value() string {
	return s.textInput.Value()
}

// Focused returns whether the search box is focused
func (s SearchBox) Focused() bool {
	return s.focused
}

// SetWidth sets the width of the search box
func (s SearchBox) SetWidth(width int) SearchBox {
	s.width = width
	s.textInput.Width = width
	return s
}

// Update handles messages and updates the search box
func (s SearchBox) Update(msg tea.Msg) (SearchBox, tea.Cmd) {
	// Only process messages when focused
	if !s.focused {
		return s, nil
	}
	
	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)
	
	return s, cmd
}

// View renders the search box
func (s SearchBox) View() string {
	// Create a label based on focus state
	var label string
	if s.focused {
		label = s.styles.SearchLabel.Render("🔍 Search [typing]:")
	} else {
		label = s.styles.SearchLabel.Render("🔍 Search:")
	}
	
	// Create search input with styled border
	searchView := s.styles.SearchBox.Copy().Width(s.width).Render(s.textInput.View())
	
	// Combine the label and input
	return label + "\n" + searchView
} 
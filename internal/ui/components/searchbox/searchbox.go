// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package searchbox

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user-story-matrix/usm/internal/ui/styles"
)

// SearchBox represents a search input component
type SearchBox struct {
		textInput textinput.Model
	styles    *styles.Styles
	focused   bool
	width     int
	value     string // Cache the current value to detect changes
}

// New creates a new SearchBox component
func New(styles *styles.Styles) SearchBox {
	ti := textinput.New()
	ti.Placeholder = "Type to search user stories, CTRL+a to toggle all/unimplemented ..."
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
		value:     "",
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
	if s.value == value {
		return s // Skip setting the same value to avoid unnecessary renders
	}
	
	s.value = value
	s.textInput.SetValue(value)
	return s
}

// Value returns the current value of the search box
func (s SearchBox) Value() string {
	return s.value
}

// Focused returns whether the search box is focused
func (s SearchBox) Focused() bool {
	return s.focused
}

// SetWidth sets the width of the search box
func (s SearchBox) SetWidth(width int) SearchBox {
	if width <= 0 {
		width = 50 // Ensure a minimum width
	}
	
	s.width = width
	
	// Adjust for prompt and styling, ensuring a reasonable minimum
	inputWidth := width - 20
	if inputWidth < 30 {
		inputWidth = 30
	}
	
	s.textInput.Width = inputWidth
	return s
}

// Update handles messages and updates the search box
func (s SearchBox) Update(msg tea.Msg) (SearchBox, tea.Cmd) {
	// Only process messages when focused
	if !s.focused {
		return s, nil
	}
	
	// Update the text input
	var cmd tea.Cmd
	newTextInput, cmd := s.textInput.Update(msg)
	
	// Update our model
	s.textInput = newTextInput
	
	// Check if the value has changed
	newValue := s.textInput.Value()
	if s.value != newValue {
		s.value = newValue
	}
	
	return s, cmd
}

// View renders the search box
func (s SearchBox) View() string {
	// Create a label based on focus state
	labelContent := "Search:"
	
	// Apply different styles based on focus state
	var label string
	var textView string
	
	if s.focused {
		// Focused state - bright/highlight colors
		label = s.styles.SearchLabel.Copy().
			PaddingRight(1).
			Render(labelContent)
			
		// Show active text input
		textView = s.textInput.View()
	} else {
		// Unfocused state - muted colors
		label = s.styles.SearchLabel.Copy().
			PaddingRight(1).
			Foreground(lipgloss.Color("240")). // Dimmer color when unfocused
			Render(labelContent)
			
		// Create a dimmed version of the text input
		// Get current value to display
		value := s.textInput.Value()
		
		// Create a custom dimmed version when unfocused
		if value != "" {
			// Show the value in a dimmed style with tag hints
			textView = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")). // Dimmer color
				Render(s.textInput.Prompt + value + " (tab to edit, CTRL+a to toggle)")
		} else {
			// Show instruction instead of placeholder when unfocused and empty
			textView = lipgloss.NewStyle().
				Foreground(lipgloss.Color("237")). // Even dimmer for instruction
				Render(s.textInput.Prompt + "Tab to search user stories")
		}
	}
	
	// Calculate width for the input field portion
	labelWidth := lipgloss.Width(label)
	inputWidth := s.width - labelWidth - 2 // Account for spacing
	if inputWidth < 40 {
		inputWidth = 40 // Minimum reasonable width
	}
	
	// Set width for text input that will be used when focused
	s.textInput.Width = inputWidth
	
	// Join components horizontally with proper alignment
	return lipgloss.JoinHorizontal(lipgloss.Center, label, textView)
} 
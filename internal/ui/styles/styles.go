// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles for the UI components
type Styles struct {
	// General text styles
	Title        lipgloss.Style
	Selected     lipgloss.Style
	Highlighted  lipgloss.Style
	Normal       lipgloss.Style
	Implemented  lipgloss.Style
	Unimplemented lipgloss.Style
	Error        lipgloss.Style
	Subtle       lipgloss.Style
	Success      lipgloss.Style
	
	// Component styles
	SearchBox    lipgloss.Style
	SearchLabel  lipgloss.Style
	SearchCursor lipgloss.Style
	SearchText   lipgloss.Style
	SearchPlaceholder lipgloss.Style
	
	StatusBar    lipgloss.Style
	Checkbox     lipgloss.Style
	CheckboxChecked lipgloss.Style
	
	// Containers
	Container    lipgloss.Style
	Border       lipgloss.Style
	FocusedBorder lipgloss.Style
}

// DefaultStyles returns the default styles
func DefaultStyles() *Styles {
	return &Styles{
		// General text styles
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true),
			
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")). // Bright white
			Background(lipgloss.Color("4")). // Dark blue
			Bold(true),
			
		Highlighted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")). // White text
			Background(lipgloss.Color("8")). // Dark gray background
			Bold(false),
			
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
			
		Implemented: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")), // Dim gray
			
		Unimplemented: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")), // Bright white
			
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Bright red
			Bold(true),
			
		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")), // Dim gray
			
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")), // Green
			
		// Component styles
		SearchBox: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()). // Use rounded borders
			BorderForeground(lipgloss.Color("205")). // Pink border
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1),
			
		SearchLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")). // Pink
			Bold(true).
			MarginBottom(0).
			MarginTop(1),
			
		SearchCursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("236")).
			Bold(true),
			
		SearchText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")). // Pink
			Bold(true),
			
		SearchPlaceholder: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Gray
			Italic(true),
			
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")). // Bright white
			Background(lipgloss.Color("25")). // Blue background
			Bold(true).
			Padding(0, 1),
			
		Checkbox: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
			
		CheckboxChecked: lipgloss.NewStyle().
			Foreground(lipgloss.Color("43")). // Green
			Bold(true),
			
		// Containers
		Container: lipgloss.NewStyle().
			Padding(1, 2),
			
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")),
			
		FocusedBorder: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")), // Pink color to match search focus
	}
}

// ItemStyles returns styles for specific indices
func (s *Styles) ItemStyles(selected, implemented, focused bool) lipgloss.Style {
	switch {
	case selected && focused:
		return s.Selected
	case focused:
		return s.Highlighted
	case selected:
		return s.Selected
	case implemented:
		return s.Implemented
	default:
		return s.Normal
	}
}

// GetCheckbox returns a styled checkbox based on state
func (s *Styles) GetCheckbox(checked bool) string {
	if checked {
		return "[✓]"
	}
	return "[ ]"
}

// GetImplementationStatus returns a styled implementation status
func (s *Styles) GetImplementationStatus(implemented bool) string {
	if implemented {
		return "I"
	}
	return "U"
} 
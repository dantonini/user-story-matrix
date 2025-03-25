// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package models

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines keybindings for the TUI
type KeyMap struct {
	// Navigation
	Up         key.Binding
	Down       key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	
	// Mode switching
	Tab        key.Binding
	Search     key.Binding
	
	// Actions
	Select     key.Binding
	Done       key.Binding
	Quit       key.Binding
	ToggleFilter key.Binding
	Clear      key.Binding
	Help       key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("PgUp", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("PgDn", "page down"),
		),
		
		// Mode switching
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("Tab", "switch focus"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		
		// Actions
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("Space", "select/deselect"),
		),
		Done: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "confirm"),
		),
		Quit: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("Esc/Ctrl+C", "quit"),
		),
		ToggleFilter: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("Ctrl+A", "toggle all/unimplemented"),
		),
		Clear: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("Ctrl+L", "clear search"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}

// ListModeHelpView returns help view text for list mode
func (k KeyMap) ListModeHelpView() string {
	return "↑/↓: navigate | Space: select | Tab: search | Enter: confirm | Esc: quit"
}

// SearchModeHelpView returns help view text for search mode
func (k KeyMap) SearchModeHelpView() string {
	return "Type to search | Esc: cancel | Enter: apply | Tab: list"
} 
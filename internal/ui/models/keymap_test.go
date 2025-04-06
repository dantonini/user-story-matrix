// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultKeyMap(t *testing.T) {
	// Act
	keymap := DefaultKeyMap()
	
	// Assert
	
	// Navigation bindings
	assert.NotEmpty(t, keymap.Up.Help().Key, "Up key binding should have a help key")
	assert.NotEmpty(t, keymap.Up.Help().Desc, "Up key binding should have a help description")
	assert.Contains(t, keymap.Up.Keys(), "up", "Up key binding should include 'up' key")
	assert.Contains(t, keymap.Up.Keys(), "k", "Up key binding should include 'k' key")
	
	assert.NotEmpty(t, keymap.Down.Help().Key, "Down key binding should have a help key")
	assert.NotEmpty(t, keymap.Down.Help().Desc, "Down key binding should have a help description")
	assert.Contains(t, keymap.Down.Keys(), "down", "Down key binding should include 'down' key")
	assert.Contains(t, keymap.Down.Keys(), "j", "Down key binding should include 'j' key")
	
	assert.NotEmpty(t, keymap.PageUp.Help().Key, "PageUp key binding should have a help key")
	assert.NotEmpty(t, keymap.PageUp.Help().Desc, "PageUp key binding should have a help description")
	assert.Contains(t, keymap.PageUp.Keys(), "pgup", "PageUp key binding should include 'pgup' key")
	
	assert.NotEmpty(t, keymap.PageDown.Help().Key, "PageDown key binding should have a help key")
	assert.NotEmpty(t, keymap.PageDown.Help().Desc, "PageDown key binding should have a help description")
	assert.Contains(t, keymap.PageDown.Keys(), "pgdown", "PageDown key binding should include 'pgdown' key")
	
	// Mode switching bindings
	assert.NotEmpty(t, keymap.Tab.Help().Key, "Tab key binding should have a help key")
	assert.NotEmpty(t, keymap.Tab.Help().Desc, "Tab key binding should have a help description")
	assert.Contains(t, keymap.Tab.Keys(), "tab", "Tab key binding should include 'tab' key")
	
	assert.NotEmpty(t, keymap.Search.Help().Key, "Search key binding should have a help key")
	assert.NotEmpty(t, keymap.Search.Help().Desc, "Search key binding should have a help description")
	assert.Contains(t, keymap.Search.Keys(), "/", "Search key binding should include '/' key")
	
	// Action bindings
	assert.NotEmpty(t, keymap.Select.Help().Key, "Select key binding should have a help key")
	assert.NotEmpty(t, keymap.Select.Help().Desc, "Select key binding should have a help description")
	assert.Contains(t, keymap.Select.Keys(), " ", "Select key binding should include space key")
	
	assert.NotEmpty(t, keymap.Done.Help().Key, "Done key binding should have a help key")
	assert.NotEmpty(t, keymap.Done.Help().Desc, "Done key binding should have a help description")
	assert.Contains(t, keymap.Done.Keys(), "enter", "Done key binding should include 'enter' key")
	
	assert.NotEmpty(t, keymap.Quit.Help().Key, "Quit key binding should have a help key")
	assert.NotEmpty(t, keymap.Quit.Help().Desc, "Quit key binding should have a help description")
	assert.Contains(t, keymap.Quit.Keys(), "esc", "Quit key binding should include 'esc' key")
	assert.Contains(t, keymap.Quit.Keys(), "ctrl+c", "Quit key binding should include 'ctrl+c' key")
	
	assert.NotEmpty(t, keymap.ToggleFilter.Help().Key, "ToggleFilter key binding should have a help key")
	assert.NotEmpty(t, keymap.ToggleFilter.Help().Desc, "ToggleFilter key binding should have a help description")
	assert.Contains(t, keymap.ToggleFilter.Keys(), "ctrl+a", "ToggleFilter key binding should include 'ctrl+a' key")
	
	assert.NotEmpty(t, keymap.Clear.Help().Key, "Clear key binding should have a help key")
	assert.NotEmpty(t, keymap.Clear.Help().Desc, "Clear key binding should have a help description")
	assert.Contains(t, keymap.Clear.Keys(), "ctrl+l", "Clear key binding should include 'ctrl+l' key")
	
	assert.NotEmpty(t, keymap.Help.Help().Key, "Help key binding should have a help key")
	assert.NotEmpty(t, keymap.Help.Help().Desc, "Help key binding should have a help description")
	assert.Contains(t, keymap.Help.Keys(), "?", "Help key binding should include '?' key")
}

func TestListModeHelpView(t *testing.T) {
	// Arrange
	keymap := DefaultKeyMap()
	
	// Act
	helpText := keymap.ListModeHelpView()
	
	// Assert
	assert.NotEmpty(t, helpText, "List mode help view should not be empty")
	
	// Check for expected key descriptions
	expectedTerms := []string{
		"navigate", 
		"select", 
		"search", 
		"confirm", 
		"quit",
	}
	
	for _, term := range expectedTerms {
		assert.Contains(t, helpText, term, "List mode help view should mention '%s'", term)
	}
	
	// Check for expected navigation keys
	assert.Contains(t, helpText, "↑/↓", "Help view should mention up/down navigation")
	assert.Contains(t, helpText, "Space", "Help view should mention Space key")
	assert.Contains(t, helpText, "Tab", "Help view should mention Tab key")
	assert.Contains(t, helpText, "Enter", "Help view should mention Enter key")
	assert.Contains(t, helpText, "Esc", "Help view should mention Esc key")
	
	// Check formatting
	assert.True(t, strings.Contains(helpText, ":"), "Help view should use colon format for key descriptions")
	assert.True(t, strings.Contains(helpText, "|"), "Help view should separate commands with pipe character")
}

func TestSearchModeHelpView(t *testing.T) {
	// Arrange
	keymap := DefaultKeyMap()
	
	// Act
	helpText := keymap.SearchModeHelpView()
	
	// Assert
	assert.NotEmpty(t, helpText, "Search mode help view should not be empty")
	
	// Check for expected key descriptions
	expectedTerms := []string{
		"Type to search", 
		"cancel", 
		"apply", 
		"list",
	}
	
	for _, term := range expectedTerms {
		assert.Contains(t, helpText, term, "Search mode help view should mention '%s'", term)
	}
	
	// Check for expected keys
	assert.Contains(t, helpText, "Esc", "Help view should mention Esc key")
	assert.Contains(t, helpText, "Enter", "Help view should mention Enter key")
	assert.Contains(t, helpText, "Tab", "Help view should mention Tab key")
	
	// Check formatting
	assert.True(t, strings.Contains(helpText, ":"), "Help view should use colon format for key descriptions")
	assert.True(t, strings.Contains(helpText, "|"), "Help view should separate commands with pipe character")
} 
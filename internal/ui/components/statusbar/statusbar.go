// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package statusbar

import (
	"fmt"
	"strings"

	"github.com/user-story-matrix/usm/internal/ui/models"
	"github.com/user-story-matrix/usm/internal/ui/styles"
)

// StatusBar represents a status bar component
type StatusBar struct {
	styles       *styles.Styles
	width        int
	keyMap       models.KeyMap
	showHelp     bool
	lastFilterStatus string
}

// New creates a new StatusBar component
func New(styles *styles.Styles, keyMap models.KeyMap) StatusBar {
	return StatusBar{
		styles:       styles,
		width:        80,
		keyMap:       keyMap,
		showHelp:     true,
		lastFilterStatus: "",
	}
}

// SetWidth sets the width of the status bar
func (s StatusBar) SetWidth(width int) StatusBar {
	s.width = width
	return s
}

// ToggleHelp toggles whether to show help
func (s StatusBar) ToggleHelp() StatusBar {
	s.showHelp = !s.showHelp
	return s
}

// View renders the status bar
func (s StatusBar) View(state *models.UIState) string {
	var sb strings.Builder
	
	// Mode indicator
	var modeInfo string
	if state.SearchFocused {
		modeInfo = "MODE: SEARCH"
	} else {
		modeInfo = "MODE: LIST"
	}
	
	// Filter status
	var filterStatus string
	if state.ShowImplemented {
		filterStatus = "FILTER: ALL STORIES"
	} else {
		filterStatus = "FILTER: UNIMPLEMENTED ONLY"
	}
	
	// If there's a search query, include it
	if state.FilterText != "" {
		filterStatus = fmt.Sprintf("SEARCH: '%s' | %s", state.FilterText, filterStatus)
	}
	
	// Store the filter status for height calculations
	s.lastFilterStatus = filterStatus
	
	// Selection status
	selectionStatus := fmt.Sprintf("✔ %d selected | %d visible / %d total", 
		state.SelectedCount(), state.FilteredStories, state.TotalStories)
	
	// Combine the status elements
	status := fmt.Sprintf("%s | %s | %s", modeInfo, filterStatus, selectionStatus)
	
	// Render the status bar
	statusBar := s.styles.StatusBar.Copy().Width(s.width).Render(status)
	sb.WriteString(statusBar + "\n")
	
	// Add help text if enabled
	if s.showHelp {
		var helpText string
		if state.SearchFocused {
			helpText = s.keyMap.SearchModeHelpView()
		} else {
			helpText = s.keyMap.ListModeHelpView()
		}
		sb.WriteString(helpText + "\n")
	}
	
	return sb.String()
}

// Height returns the height of the status bar
func (s StatusBar) Height() int {
	if s.showHelp {
		return 2 // Status bar + help text
	}
	return 1 // Just status bar
} 
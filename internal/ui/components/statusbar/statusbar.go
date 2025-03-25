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
	
	// Selection status with hidden selections if any
	selectionStatus := fmt.Sprintf("✔ %d selected", state.SelectedCount())
	
	// Add hidden selection count if there are any
	if hiddenCount := state.HiddenSelectedCount(); hiddenCount > 0 {
		selectionStatus += fmt.Sprintf(" (%d hidden)", hiddenCount)
	}
	
	// Visible status
	visibleStatus := fmt.Sprintf("%d visible / %d total", state.FilteredStories, state.TotalStories)
	
	// Filter status
	var filterStatus string
	if state.ShowImplemented {
		filterStatus = "Filter: All"
	} else {
		filterStatus = "Filter: Unimplemented"
	}
	
	// Combine the status elements
	status := fmt.Sprintf("%s | %s | %s", selectionStatus, visibleStatus, filterStatus)
	
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
		sb.WriteString(helpText)
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
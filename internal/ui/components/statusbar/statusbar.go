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
	styles          *styles.Styles
	width           int
	keyMap          models.KeyMap
	showHelp        bool
	lastFilterStatus string
	
	// Cache fields for performance
	lastState       *models.UIState
	cachedStatusBar string
	cachedHelpText  string
	stateChanged    bool
}

// New creates a new StatusBar component
func New(styles *styles.Styles, keyMap models.KeyMap) StatusBar {
	return StatusBar{
		styles:          styles,
		width:           80,
		keyMap:          keyMap,
		showHelp:        true,
		lastFilterStatus: "",
		stateChanged:    true, // Force initial render
	}
}

// SetWidth sets the width of the status bar
func (s StatusBar) SetWidth(width int) StatusBar {
	if s.width != width {
		s.width = width
		s.stateChanged = true // Width changed, need to re-render
	}
	return s
}

// ToggleHelp toggles whether to show help
func (s StatusBar) ToggleHelp() StatusBar {
	s.showHelp = !s.showHelp
	s.stateChanged = true // Help visibility changed, need to re-render
	return s
}

// shouldUpdate checks if the status bar needs to be re-rendered
func (s *StatusBar) shouldUpdate(state *models.UIState) bool {
	if s.stateChanged {
		return true
	}
	
	if s.lastState == nil {
		return true
	}
	
	// Check if any relevant state changed
	return s.lastState.SearchFocused != state.SearchFocused ||
		s.lastState.SelectedCount() != state.SelectedCount() ||
		s.lastState.HiddenSelectedCount() != state.HiddenSelectedCount() ||
		s.lastState.FilteredStories != state.FilteredStories ||
		s.lastState.TotalStories != state.TotalStories ||
		s.lastState.ShowImplemented != state.ShowImplemented
}

// View renders the status bar
func (s StatusBar) View(state *models.UIState) string {
	// Check if we can use the cached view
	if !s.shouldUpdate(state) {
		// Build final output from cache
		if !s.showHelp {
			return s.cachedStatusBar
		}
		return s.cachedStatusBar + "\n" + s.cachedHelpText
	}
	
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
	sb.WriteString(statusBar)
	
	// Update cache and state tracking
	s.cachedStatusBar = statusBar
	s.lastFilterStatus = filterStatus
	
	// Cache the help text separately
	if state.SearchFocused {
		s.cachedHelpText = s.keyMap.SearchModeHelpView()
	} else {
		s.cachedHelpText = s.keyMap.ListModeHelpView()
	}
	
	// Create a copy of the state for comparison
	stateCopy := *state
	s.lastState = &stateCopy
	s.stateChanged = false
	
	// Add help text if enabled
	if s.showHelp {
		sb.WriteString("\n" + s.cachedHelpText)
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
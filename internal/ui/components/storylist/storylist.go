// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storylist

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui/styles"
)

// StoryItem represents a user story in the list
type StoryItem struct {
	Story      models.UserStory
	Index      int
	IsSelected bool
}

// StoryList represents a list of user stories
type StoryList struct {
	items         []StoryItem
	cursor        int
	styles        *styles.Styles
	focused       bool
	width         int
	height        int
	visibleStart  int
	visibleEnd    int
	totalCount    int
	selectedCount int
	// Cache fields for performance
	lastRender    string
	needsRender   bool
}

// New creates a new StoryList component
func New(styles *styles.Styles) StoryList {
	return StoryList{
		items:         []StoryItem{},
		cursor:        0,
		styles:        styles,
		focused:       false,
		width:         80,
		height:        10,
		visibleStart:  0,
		visibleEnd:    0,
		totalCount:    0,
		selectedCount: 0,
		needsRender:   true,
	}
}

// Focus focuses the story list
func (l StoryList) Focus() StoryList {
	if !l.focused {
		l.focused = true
		l.needsRender = true
	}
	return l
}

// Blur blurs the story list
func (l StoryList) Blur() StoryList {
	if l.focused {
		l.focused = false
		l.needsRender = true
	}
	return l
}

// SetItems sets the items in the story list
func (l StoryList) SetItems(stories []models.UserStory, selectedIDs map[string]bool) StoryList {
	if stories == nil {
		stories = []models.UserStory{} // Convert nil to empty slice for safety
	}
	
	// Create new story items
	items := make([]StoryItem, len(stories))
	
	// Count selected items
	selectedCount := 0
	
	for i, story := range stories {
		// Check if this story is selected
		isSelected := selectedIDs[story.FilePath]
		if isSelected {
			selectedCount++
		}
		
		items[i] = StoryItem{
			Story:      story,
			Index:      i,
			IsSelected: isSelected,
		}
	}
	
	l.items = items
	l.totalCount = len(stories)
	l.selectedCount = selectedCount
	l.needsRender = true
	
	// Ensure cursor is still valid
	if len(items) == 0 {
		l.cursor = 0
	} else if l.cursor >= len(items) {
		l.cursor = len(items) - 1
	} else if l.cursor < 0 {
		l.cursor = 0
	}
	
	// Update visible range
	l.updateVisibleRange()
	
	return l
}

// SetSize sets the dimensions of the story list
func (l StoryList) SetSize(width, height int) StoryList {
	if width <= 0 {
		width = 80 // Ensure minimum width
	}
	if height <= 0 {
		height = 10 // Ensure minimum height
	}
	
	if l.width != width || l.height != height {
		l.width = width
		l.height = height
		l.needsRender = true
		
		// Update visible range
		l.updateVisibleRange()
	}
	
	return l
}

// updateVisibleRange updates the range of visible items
func (l *StoryList) updateVisibleRange() {
	if len(l.items) == 0 {
		l.visibleStart = 0
		l.visibleEnd = 0
		return
	}
	
	// Ensure cursor is always visible
	if l.cursor < l.visibleStart {
		l.visibleStart = l.cursor
	} else if l.cursor >= l.visibleEnd {
		// Move the window so that cursor is at the end
		l.visibleStart = l.cursor - l.height + 1
		if l.visibleStart < 0 {
			l.visibleStart = 0
		}
	}
	
	// Calculate visible end based on height
	l.visibleEnd = l.visibleStart + l.height
	if l.visibleEnd > len(l.items) {
		l.visibleEnd = len(l.items)
	}
	
	l.needsRender = true
}

// ToggleSelection toggles the selection of the currently selected item
func (l StoryList) ToggleSelection() (StoryList, string) {
	if len(l.items) == 0 || l.cursor < 0 || l.cursor >= len(l.items) {
		return l, ""
	}
	
	// Toggle the selected status
	l.items[l.cursor].IsSelected = !l.items[l.cursor].IsSelected
	
	// Update selected count
	if l.items[l.cursor].IsSelected {
		l.selectedCount++
	} else {
		l.selectedCount--
	}
	
	l.needsRender = true
	
	// Get the toggled story ID
	return l, l.items[l.cursor].Story.FilePath
}

// MoveUp moves the cursor up
func (l StoryList) MoveUp() StoryList {
	if len(l.items) == 0 {
		return l
	}
	
	l.cursor--
	if l.cursor < 0 {
		l.cursor = 0
	} else {
		l.needsRender = true
	}
	
	// Update visible range
	l.updateVisibleRange()
	
	return l
}

// MoveDown moves the cursor down
func (l StoryList) MoveDown() StoryList {
	if len(l.items) == 0 {
		return l
	}
	
	l.cursor++
	if l.cursor >= len(l.items) {
		l.cursor = len(l.items) - 1
	} else {
		l.needsRender = true
	}
	
	// Update visible range
	l.updateVisibleRange()
	
	return l
}

// PageUp scrolls one page up
func (l StoryList) PageUp() StoryList {
	if len(l.items) == 0 {
		return l
	}
	
	l.cursor -= l.height
	if l.cursor < 0 {
		l.cursor = 0
	}
	
	l.needsRender = true
	
	// Update visible range
	l.updateVisibleRange()
	
	return l
}

// PageDown scrolls one page down
func (l StoryList) PageDown() StoryList {
	if len(l.items) == 0 {
		return l
	}
	
	l.cursor += l.height
	if l.cursor >= len(l.items) {
		l.cursor = len(l.items) - 1
	}
	
	l.needsRender = true
	
	// Update visible range
	l.updateVisibleRange()
	
	return l
}

// CurrentItem returns the currently selected item
func (l StoryList) CurrentItem() (StoryItem, bool) {
	if len(l.items) == 0 || l.cursor < 0 || l.cursor >= len(l.items) {
		return StoryItem{}, false
	}
	
	return l.items[l.cursor], true
}

// Update handles messages and updates the story list
func (l StoryList) Update(msg tea.Msg) (StoryList, tea.Cmd) {
	// Only process messages when focused
	if !l.focused {
		return l, nil
	}
	
	// Handle key presses
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			return l.MoveUp(), nil
		case "down", "j":
			return l.MoveDown(), nil
		case "pgup":
			return l.PageUp(), nil
		case "pgdown":
			return l.PageDown(), nil
		case " ":
			newList, _ := l.ToggleSelection()
			return newList, nil
		}
	}
	
	return l, nil
}

// calculateCommonPrefix finds the common directory prefix across a set of paths
func calculateCommonPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	
	// Start with the first path as the reference
	reference := strings.Split(paths[0], "/")
	
	// Compare with all other paths
	for _, path := range paths[1:] {
		parts := strings.Split(path, "/")
		
		// Find how many segments match
		var i int
		for i = 0; i < len(reference) && i < len(parts); i++ {
			if reference[i] != parts[i] {
				break
			}
		}
		
		// Update reference to only keep matching parts
		reference = reference[:i]
		if len(reference) == 0 {
			break
		}
	}
	
	// Convert back to path string
	if len(reference) == 0 {
		return ""
	}
	
	return strings.Join(reference, "/")
}

// shortenPath removes common prefix from a path
func shortenPath(path string, commonPrefix string) string {
	if commonPrefix == "" || path == "" {
		return path
	}
	
	// If the commonPrefix is the entire path, don't shorten
	if path == commonPrefix || commonPrefix == path+"/" {
		return path
	}
	
	// If path starts with common prefix, remove it
	if strings.HasPrefix(path, commonPrefix) {
		shortened := path[len(commonPrefix):]
		// Remove leading slash if present
		if strings.HasPrefix(shortened, "/") {
			shortened = shortened[1:]
		}
		// Special case: if the path is exactly the common prefix
		if shortened == "" {
			return "…/"
		}
		return "…/" + shortened
	}
	
	return path
}

// View renders the story list
func (l StoryList) View() string {
	if len(l.items) == 0 {
		return l.styles.Normal.Render("No stories to display.")
	}
	
	// Return cached view if nothing has changed
	if !l.needsRender && l.lastRender != "" {
		return l.lastRender
	}
	
	var sb strings.Builder
	
	// Calculate common prefix for all visible items with paths
	var paths []string
	for i := l.visibleStart; i < l.visibleEnd && i < len(l.items); i++ {
		if path := l.items[i].Story.FilePath; path != "" {
			paths = append(paths, path)
		}
	}
	commonPrefix := calculateCommonPrefix(paths)
	
	// Display only visible items
	for i := l.visibleStart; i < l.visibleEnd && i < len(l.items); i++ {
		item := l.items[i]
		
		// Build the raw line content without any styling first
		checkbox := "[ ]"
		if item.IsSelected {
			checkbox = "[✓]"
		}
		
		impStatus := "U"
		if item.Story.IsImplemented {
			impStatus = "I"
		}
		
		// Create the title (truncate if too long)
		title := item.Story.Title
		maxTitleWidth := l.width - 15
		if len(title) > maxTitleWidth {
			title = title[:maxTitleWidth-3] + "..."
		}
		
		// Create the full raw line
		rawLine := fmt.Sprintf(" %s %s %s", checkbox, impStatus, title)
		
		// Simple style selection based on conditions
		var renderedLine string
		switch {
		case l.focused && i == l.cursor && item.IsSelected:
			// Selected and focused item (cursor)
			renderedLine = l.styles.Selected.Render(rawLine)
		case l.focused && i == l.cursor:
			// Focused but not selected item (cursor)
			renderedLine = l.styles.Highlighted.Render(rawLine)
		case item.IsSelected:
			// Selected but not focused item
			renderedLine = l.styles.Selected.Render(rawLine)
		case item.Story.IsImplemented:
			// Implemented item
			renderedLine = l.styles.Implemented.Render(rawLine)
		default:
			// Default case
			renderedLine = l.styles.Normal.Render(rawLine)
		}
		
		// Add the rendered line to output
		sb.WriteString(renderedLine)
		sb.WriteString("\n")
		
		// Only show shortened filepath on the currently focused item for less visual noise
		if l.focused && i == l.cursor && item.Story.FilePath != "" {
			filePath := shortenPath(item.Story.FilePath, commonPrefix)
			pathLine := fmt.Sprintf("       %s", filePath)
			sb.WriteString(l.styles.Implemented.Render(pathLine))
			sb.WriteString("\n")
		}
	}
	
	// Show simple indicator for navigation
	if len(l.items) > l.height {
		sb.WriteString(l.styles.Implemented.Render(" ↑/↓ to navigate"))
	}
	
	// Cache the rendered view
	l.lastRender = sb.String()
	l.needsRender = false
	
	return l.lastRender
}

// SetCursor sets the cursor position
func (l StoryList) SetCursor(position int) StoryList {
	if len(l.items) == 0 {
		return l
	}
	
	// Set the cursor to the specified position
	if l.cursor != position {
		l.cursor = position
		l.needsRender = true
		
		// Ensure the cursor is within bounds
		if l.cursor < 0 {
			l.cursor = 0
		} else if l.cursor >= len(l.items) {
			l.cursor = len(l.items) - 1
		}
		
		// Update visible range
		l.updateVisibleRange()
	}
	
	return l
} 
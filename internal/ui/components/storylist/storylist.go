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
	}
}

// Focus focuses the story list
func (l StoryList) Focus() StoryList {
	l.focused = true
	return l
}

// Blur blurs the story list
func (l StoryList) Blur() StoryList {
	l.focused = false
	return l
}

// SetItems sets the items in the story list
func (l StoryList) SetItems(stories []models.UserStory, selectedIDs map[string]bool) StoryList {
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
	
	// Ensure cursor is still valid
	if l.cursor >= len(items) {
		if len(items) > 0 {
			l.cursor = len(items) - 1
		} else {
			l.cursor = 0
		}
	}
	
	// Update visible range
	l.updateVisibleRange()
	
	return l
}

// SetSize sets the dimensions of the story list
func (l StoryList) SetSize(width, height int) StoryList {
	l.width = width
	l.height = height
	
	// Update visible range
	l.updateVisibleRange()
	
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

// View renders the story list
func (l StoryList) View() string {
	if len(l.items) == 0 {
		return "No stories to display."
	}
	
	var sb strings.Builder
	
	// Display only visible items
	for i := l.visibleStart; i < l.visibleEnd && i < len(l.items); i++ {
		item := l.items[i]
		
		// Create the checkbox
		checkbox := l.styles.GetCheckbox(item.IsSelected)
		
		// Create the implementation status
		impStatus := l.styles.GetImplementationStatus(item.Story.IsImplemented)
		
		// Create the title (truncate if too long)
		title := item.Story.Title
		if len(title) > l.width - 30 {
			title = title[:l.width-33] + "..."
		}
		
		// Create the file path (truncate and show only relevant parts)
		filePath := item.Story.FilePath
		if len(filePath) > 30 {
			parts := strings.Split(filePath, "/")
			if len(parts) > 2 {
				// Show last two parts only
				filePath = ".../" + strings.Join(parts[len(parts)-2:], "/")
			}
		}
		
		// Determine style based on selection and cursor
		var style = l.styles.ItemStyles(
			item.IsSelected,
			item.Story.IsImplemented,
			l.focused && i == l.cursor,
		)
		
		// Create the full line
		line := fmt.Sprintf("%s %s %s (%s)", checkbox, impStatus, title, filePath)
		
		// Apply style to line
		sb.WriteString(style.Width(l.width).Render(line))
		sb.WriteString("\n")
	}
	
	// Add scrolling indicator if needed
	if len(l.items) > l.height {
		indicator := fmt.Sprintf("Showing %d-%d of %d", 
			l.visibleStart+1, l.visibleEnd, len(l.items))
		sb.WriteString(l.styles.Normal.Render(indicator))
	}
	
	return sb.String()
}

// SetCursor sets the cursor position
func (l StoryList) SetCursor(position int) StoryList {
	if len(l.items) == 0 {
		return l
	}
	
	// Set the cursor to the specified position
	l.cursor = position
	
	// Ensure the cursor is within bounds
	if l.cursor < 0 {
		l.cursor = 0
	} else if l.cursor >= len(l.items) {
		l.cursor = len(l.items) - 1
	}
	
	// Update visible range
	l.updateVisibleRange()
	
	return l
} 
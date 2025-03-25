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
	
	// Handle window resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		l = l.SetSize(msg.Width, msg.Height-6) // Adjust for header/footer
		return l, nil
	}
	
	return l, nil
}

// View renders the story list
func (l StoryList) View() string {
	if len(l.items) == 0 {
		return l.styles.Error.Render("⚠️  No matching user stories found.")
	}
	
	// Create a string builder for the list
	var sb strings.Builder
	
	// Draw a separator line
	sb.WriteString(strings.Repeat("─", l.width) + "\n")
	
	// Draw visible items
	for i := l.visibleStart; i < l.visibleEnd; i++ {
		item := l.items[i]
		
		// Get styled checkbox
		checkbox := l.styles.GetCheckbox(item.IsSelected)
		
		// Get implementation status
		implStatus := l.styles.GetImplementationStatus(item.Story.IsImplemented)
		
		// Get the path display
		path := strings.TrimPrefix(item.Story.FilePath, "docs/user-stories/")
		pathInfo := " (" + path + ")"
		
		// Prepare the line
		line := fmt.Sprintf("%s %s %s%s", checkbox, implStatus, item.Story.Title, pathInfo)
		
		// Apply the appropriate style based on item state
		isCurrent := i == l.cursor
		styledLine := l.styles.ItemStyles(
			item.IsSelected, 
			item.Story.IsImplemented, 
			isCurrent && l.focused,
		).Render(line)
		
		// Add the line to the view
		sb.WriteString(styledLine + "\n")
	}
	
	// Draw a separator line
	sb.WriteString(strings.Repeat("─", l.width) + "\n")
	
	// Draw paginator
	if len(l.items) > l.height {
		// Show current range and total
		paginator := fmt.Sprintf("Stories %d-%d / %d | ↑↓ scroll | PgUp/PgDn fast scroll", 
			l.visibleStart+1, l.visibleEnd, len(l.items))
		sb.WriteString(paginator + "\n")
	}
	
	return sb.String()
} 
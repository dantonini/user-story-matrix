// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storylist

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui/styles"
)

func TestCalculateCommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name:     "empty paths",
			paths:    []string{},
			expected: "",
		},
		{
			name:     "single path",
			paths:    []string{"docs/user-stories/dir1/file1.md"},
			expected: "docs/user-stories/dir1/file1.md",
		},
		{
			name: "multiple paths with common prefix",
			paths: []string{
				"docs/user-stories/dir1/file1.md",
				"docs/user-stories/dir1/file2.md",
				"docs/user-stories/dir1/file3.md",
			},
			expected: "docs/user-stories/dir1",
		},
		{
			name: "multiple paths with partial common prefix",
			paths: []string{
				"docs/user-stories/dir1/file1.md",
				"docs/user-stories/dir2/file2.md",
				"docs/user-stories/dir3/file3.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "no common prefix",
			paths: []string{
				"docs/user-stories/file1.md",
				"src/components/file2.md",
				"test/file3.md",
			},
			expected: "",
		},
		{
			name: "common prefix at root",
			paths: []string{
				"docs/user-stories/file1.md",
				"docs/code/file2.md",
				"docs/tests/file3.md",
			},
			expected: "docs",
		},
		{
			name: "exact same paths",
			paths: []string{
				"docs/user-stories/file.md",
				"docs/user-stories/file.md",
				"docs/user-stories/file.md",
			},
			expected: "docs/user-stories/file.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateCommonPrefix(tc.paths)
			if result != tc.expected {
				t.Errorf("Expected common prefix to be %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestShortenPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		commonPrefix string
		expected     string
	}{
		{
			name:         "empty path",
			path:         "",
			commonPrefix: "docs/user-stories",
			expected:     "",
		},
		{
			name:         "empty common prefix",
			path:         "docs/user-stories/file.md",
			commonPrefix: "",
			expected:     "docs/user-stories/file.md",
		},
		{
			name:         "path contains common prefix",
			path:         "docs/user-stories/dir1/file.md",
			commonPrefix: "docs/user-stories",
			expected:     "…/dir1/file.md",
		},
		{
			name:         "path does not contain common prefix",
			path:         "src/components/file.md",
			commonPrefix: "docs/user-stories",
			expected:     "src/components/file.md",
		},
		{
			name:         "path equals common prefix",
			path:         "docs/user-stories",
			commonPrefix: "docs/user-stories",
			expected:     "docs/user-stories",
		},
		{
			name:         "common prefix is the entire path",
			path:         "docs/user-stories/dir1/file.md",
			commonPrefix: "docs/user-stories/dir1/file.md",
			expected:     "docs/user-stories/dir1/file.md", // No shortening as it would be empty
		},
		{
			name:         "path with trailing slash in common prefix",
			path:         "docs/user-stories/dir1/file.md",
			commonPrefix: "docs/user-stories/",
			expected:     "…/dir1/file.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := shortenPath(tc.path, tc.commonPrefix)
			if result != tc.expected {
				t.Errorf("Expected shortened path to be %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestCalculateCommonPrefixEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name: "paths with mixed case should be case sensitive",
			paths: []string{
				"docs/User-Stories/file1.md",
				"docs/user-stories/file2.md",
			},
			expected: "docs",
		},
		{
			name: "paths with trailing slashes",
			paths: []string{
				"docs/user-stories/dir1/",
				"docs/user-stories/dir2/",
			},
			expected: "docs/user-stories",
		},
		{
			name: "paths with varying depths",
			paths: []string{
				"docs/user-stories/dir1/subdir/file1.md",
				"docs/user-stories/dir1/file2.md",
				"docs/user-stories/file3.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "subset paths",
			paths: []string{
				"docs/user-stories",
				"docs/user-stories/dir1/file.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "paths with special characters",
			paths: []string{
				"docs/user-stories/test-file-1.md",
				"docs/user-stories/test_file_2.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "paths with numbers",
			paths: []string{
				"docs/user-stories/1-introduction.md",
				"docs/user-stories/2-setup.md",
			},
			expected: "docs/user-stories",
		},
		{
			name: "absolute paths",
			paths: []string{
				"/home/user/docs/user-stories/file1.md",
				"/home/user/docs/user-stories/file2.md",
			},
			expected: "/home/user/docs/user-stories",
		},
		{
			name: "mixture of absolute and relative paths",
			paths: []string{
				"/docs/user-stories/file1.md",
				"docs/user-stories/file2.md",
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateCommonPrefix(tc.paths)
			if result != tc.expected {
				t.Errorf("Expected common prefix to be %q, got %q", tc.expected, result)
			}
		})
	}
}

// TestCalculateCommonPrefixBenchmark is a helper to validate the performance
// of the common prefix calculation with a large number of paths
func TestCalculateCommonPrefixBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark test in short mode")
	}
	
	// Create a large number of paths with a common prefix
	paths := make([]string, 1000)
	for i := range paths {
		paths[i] = fmt.Sprintf("docs/user-stories/dir%d/file%d.md", i%10, i)
	}
	
	start := time.Now()
	result := calculateCommonPrefix(paths)
	duration := time.Since(start)
	
	expected := "docs/user-stories"
	if result != expected {
		t.Errorf("Expected common prefix to be %q, got %q", expected, result)
	}
	
	t.Logf("Calculated common prefix for %d paths in %v", len(paths), duration)
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList
	list := New(s)
	
	// Check default values
	if list.items == nil || len(list.items) != 0 {
		t.Errorf("Expected empty items slice, got %v", list.items)
	}
	if list.cursor != 0 {
		t.Errorf("Expected cursor to be 0, got %d", list.cursor)
	}
	if list.styles != s {
		t.Errorf("Expected styles to be set correctly")
	}
	if list.focused != false {
		t.Errorf("Expected focused to be false, got %t", list.focused)
	}
	if list.width != 80 {
		t.Errorf("Expected width to be 80, got %d", list.width)
	}
	if list.height != 10 {
		t.Errorf("Expected height to be 10, got %d", list.height)
	}
	if list.needsRender != true {
		t.Errorf("Expected needsRender to be true, got %t", list.needsRender)
	}
}

// TestFocus tests the Focus function
func TestFocus(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList
	list := New(s)
	
	// Initially the list should not be focused
	if list.focused != false {
		t.Errorf("Expected focused to be false initially, got %t", list.focused)
	}
	
	// Focus the list
	list = list.Focus()
	
	// Check that the list is now focused
	if list.focused != true {
		t.Errorf("Expected focused to be true after focus, got %t", list.focused)
	}
	if list.needsRender != true {
		t.Errorf("Expected needsRender to be true after focus, got %t", list.needsRender)
	}
	
	// Focus again should not change the state
	list.needsRender = false
	list = list.Focus()
	if list.needsRender != false {
		t.Errorf("Expected needsRender to remain false when already focused, got %t", list.needsRender)
	}
}

// TestBlur tests the Blur function
func TestBlur(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList and focus it
	list := New(s).Focus()
	
	// Initially the list should be focused
	if list.focused != true {
		t.Errorf("Expected focused to be true initially, got %t", list.focused)
	}
	
	// Blur the list
	list = list.Blur()
	
	// Check that the list is now not focused
	if list.focused != false {
		t.Errorf("Expected focused to be false after blur, got %t", list.focused)
	}
	if list.needsRender != true {
		t.Errorf("Expected needsRender to be true after blur, got %t", list.needsRender)
	}
	
	// Blur again should not change the state
	list.needsRender = false
	list = list.Blur()
	if list.needsRender != false {
		t.Errorf("Expected needsRender to remain false when already blurred, got %t", list.needsRender)
	}
}

// createTestUserStory creates a test user story for testing
func createTestUserStory(index int) models.UserStory {
	return models.UserStory{
		Title:            fmt.Sprintf("Test User Story %d", index),
		FilePath:         fmt.Sprintf("docs/user-stories/%02d-test-user-story-%d.md", index, index),
		ContentHash:      fmt.Sprintf("hash%d", index),
		SequentialNumber: fmt.Sprintf("%02d", index),
		CreatedAt:        time.Now(),
		LastUpdated:      time.Now(),
		Content:          fmt.Sprintf("# Test User Story %d\n\nThis is a test user story.", index),
		Description:      fmt.Sprintf("Description %d", index),
		IsImplemented:    index%2 == 0, // Even indices are implemented
	}
}

// TestSetItems tests the SetItems function
func TestSetItems(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList
	list := New(s)
	
	// Create test stories
	stories := []models.UserStory{
		createTestUserStory(1),
		createTestUserStory(2),
		createTestUserStory(3),
	}
	
	// Create selected IDs map
	selectedIDs := map[string]bool{
		stories[0].FilePath: true,
		stories[2].FilePath: true,
	}
	
	// Set the items
	list = list.SetItems(stories, selectedIDs)
	
	// Check that the items are set correctly
	if len(list.items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(list.items))
	}
	
	// Check that the selected items are correct
	if list.selectedCount != 2 {
		t.Errorf("Expected 2 selected items, got %d", list.selectedCount)
	}
	
	// Check specific items
	if !list.items[0].IsSelected {
		t.Errorf("Expected item 0 to be selected")
	}
	if list.items[1].IsSelected {
		t.Errorf("Expected item 1 to not be selected")
	}
	if !list.items[2].IsSelected {
		t.Errorf("Expected item 2 to be selected")
	}
	
	// Test with nil stories
	list = list.SetItems(nil, selectedIDs)
	if len(list.items) != 0 {
		t.Errorf("Expected 0 items when setting nil stories, got %d", len(list.items))
	}
	
	// Test cursor handling with empty list
	if list.cursor != 0 {
		t.Errorf("Expected cursor to be 0 for empty list, got %d", list.cursor)
	}
	
	// Test cursor handling with fewer items than current cursor
	stories = []models.UserStory{
		createTestUserStory(1),
	}
	list = New(s)
	list.cursor = 5
	list = list.SetItems(stories, nil)
	if list.cursor != 0 {
		t.Errorf("Expected cursor to be reset to 0 when fewer items, got %d", list.cursor)
	}
}

// TestSetSize tests the SetSize function
func TestSetSize(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList
	list := New(s)
	
	// Set a valid size
	list = list.SetSize(100, 20)
	if list.width != 100 || list.height != 20 {
		t.Errorf("Expected width=100, height=20, got width=%d, height=%d", list.width, list.height)
	}
	
	// Test minimum size handling
	list = list.SetSize(0, 0)
	if list.width != 80 || list.height != 10 {
		t.Errorf("Expected minimum width=80, height=10, got width=%d, height=%d", list.width, list.height)
	}
	
	// Test that setting the same size doesn't trigger a render
	list.needsRender = false
	list = list.SetSize(80, 10)
	if list.needsRender != false {
		t.Errorf("Expected needsRender to remain false when size unchanged, got %t", list.needsRender)
	}
}

// TestToggleSelection tests the ToggleSelection function
func TestToggleSelection(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList
	list := New(s)
	
	// Test with empty list
	var newList StoryList
	newList, id := list.ToggleSelection()
	if id != "" {
		t.Errorf("Expected empty ID for empty list, got %s", id)
	}
	if newList.selectedCount != list.selectedCount {
		t.Errorf("Expected selectedCount to remain unchanged for empty list")
	}
	
	// Create test stories
	stories := []models.UserStory{
		createTestUserStory(1),
		createTestUserStory(2),
	}
	
	// Set the items with none selected
	list = list.SetItems(stories, nil)
	if list.selectedCount != 0 {
		t.Errorf("Expected 0 selected items initially, got %d", list.selectedCount)
	}
	
	// Toggle the first item
	list, id = list.ToggleSelection()
	if id != stories[0].FilePath {
		t.Errorf("Expected ID %s, got %s", stories[0].FilePath, id)
	}
	if !list.items[0].IsSelected {
		t.Errorf("Expected item 0 to be selected after toggle")
	}
	if list.selectedCount != 1 {
		t.Errorf("Expected 1 selected item after toggle, got %d", list.selectedCount)
	}
	
	// Toggle it again to unselect
	list, id = list.ToggleSelection()
	if id != stories[0].FilePath {
		t.Errorf("Expected ID %s, got %s", stories[0].FilePath, id)
	}
	if list.items[0].IsSelected {
		t.Errorf("Expected item 0 to be unselected after second toggle")
	}
	if list.selectedCount != 0 {
		t.Errorf("Expected 0 selected items after unselect, got %d", list.selectedCount)
	}
}

// TestMovementFunctions tests the movement functions (MoveUp, MoveDown, PageUp, PageDown)
func TestMovementFunctions(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a large list of stories for testing pagination
	var stories []models.UserStory
	for i := 1; i <= 20; i++ {
		stories = append(stories, createTestUserStory(i))
	}
	
	// Create a new StoryList and set the items
	list := New(s).SetItems(stories, nil)
	list = list.SetSize(100, 10) // Set height to 10 to test pagination
	
	// Test MoveUp with cursor at 0
	originalCursor := list.cursor
	list = list.MoveUp()
	if list.cursor != originalCursor {
		t.Errorf("Expected cursor to remain at %d when already at top, got %d", originalCursor, list.cursor)
	}
	
	// Test MoveDown
	list = list.MoveDown()
	if list.cursor != 1 {
		t.Errorf("Expected cursor to move to 1, got %d", list.cursor)
	}
	
	// Move to the middle
	list.cursor = 10
	list.updateVisibleRange()
	
	// Test PageUp
	list = list.PageUp()
	if list.cursor != 0 {
		t.Errorf("Expected cursor to move to 0 after PageUp from 10, got %d", list.cursor)
	}
	
	// Move to the middle again
	list.cursor = 10
	list.updateVisibleRange()
	
	// Test PageDown
	list = list.PageDown()
	if list.cursor != 19 {
		t.Errorf("Expected cursor to move to 19 after PageDown from 10, got %d", list.cursor)
	}
	
	// Test MoveDown at the bottom
	list = list.MoveDown()
	if list.cursor != 19 {
		t.Errorf("Expected cursor to remain at 19 when already at bottom, got %d", list.cursor)
	}
	
	// Test with empty list
	list = New(s)
	list = list.MoveUp()
	list = list.MoveDown()
	list = list.PageUp()
	list = list.PageDown()
	if list.cursor != 0 {
		t.Errorf("Expected cursor to remain at 0 for empty list, got %d", list.cursor)
	}
}

// TestCurrentItem tests the CurrentItem function
func TestCurrentItem(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList
	list := New(s)
	
	// Test with empty list
	item, ok := list.CurrentItem()
	if ok {
		t.Errorf("Expected ok to be false for empty list")
	}
	
	// Create test stories
	stories := []models.UserStory{
		createTestUserStory(1),
		createTestUserStory(2),
	}
	
	// Set the items
	list = list.SetItems(stories, nil)
	
	// Get the current item
	item, ok = list.CurrentItem()
	if !ok {
		t.Errorf("Expected ok to be true for non-empty list")
	}
	if item.Story.Title != stories[0].Title {
		t.Errorf("Expected title %s, got %s", stories[0].Title, item.Story.Title)
	}
	
	// Move to the second item
	list = list.MoveDown()
	item, ok = list.CurrentItem()
	if !ok {
		t.Errorf("Expected ok to be true for non-empty list")
	}
	if item.Story.Title != stories[1].Title {
		t.Errorf("Expected title %s, got %s", stories[1].Title, item.Story.Title)
	}
}

// TestSetCursor tests the SetCursor function
func TestSetCursor(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList
	list := New(s)
	
	// Create test stories
	stories := []models.UserStory{
		createTestUserStory(1),
		createTestUserStory(2),
		createTestUserStory(3),
	}
	
	// Set the items
	list = list.SetItems(stories, nil)
	
	// Set cursor to valid value
	list = list.SetCursor(1)
	if list.cursor != 1 {
		t.Errorf("Expected cursor to be 1, got %d", list.cursor)
	}
	
	// Set cursor to negative value
	list = list.SetCursor(-1)
	if list.cursor != 0 {
		t.Errorf("Expected cursor to be clamped to 0, got %d", list.cursor)
	}
	
	// Set cursor beyond the end
	list = list.SetCursor(10)
	if list.cursor != 2 {
		t.Errorf("Expected cursor to be clamped to 2, got %d", list.cursor)
	}
	
	// Set cursor on empty list
	list = New(s)
	list = list.SetCursor(5)
	if list.cursor != 0 {
		t.Errorf("Expected cursor to be 0 for empty list, got %d", list.cursor)
	}
}

// TestUpdate tests the Update function
func TestUpdate(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList
	list := New(s)
	
	// Create test stories
	stories := []models.UserStory{
		createTestUserStory(1),
		createTestUserStory(2),
		createTestUserStory(3),
	}
	
	// Set the items
	list = list.SetItems(stories, nil)
	
	// Focus the list so it will process messages
	list = list.Focus()
	
	// Test key message
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyUp})
	if list.cursor != 0 {
		t.Errorf("Expected cursor to remain at 0 after up key at top, got %d", list.cursor)
	}
	
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyDown})
	if list.cursor != 1 {
		t.Errorf("Expected cursor to be 1 after down key, got %d", list.cursor)
	}
	
	// Test window resize message - note that the Update method doesn't handle this
	// so we need to manually call SetSize for testing
	originalWidth, originalHeight := list.width, list.height
	list = list.SetSize(100, 30)
	if list.width != 100 || list.height != 30 {
		t.Errorf("Expected width=100, height=30, got width=%d, height=%d", list.width, list.height)
	}
	if list.width == originalWidth && list.height == originalHeight {
		t.Errorf("Expected dimensions to change after SetSize")
	}
	
	// Test key message with unfocused list - should not change cursor
	list = list.Blur()
	originalCursor := list.cursor
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyDown})
	if list.cursor != originalCursor {
		t.Errorf("Expected cursor to remain at %d when unfocused, got %d", originalCursor, list.cursor)
	}
	
	// Test ignored message
	list = list.Focus()
	originalCursor = list.cursor
	list, _ = list.Update("random message")
	if list.cursor != originalCursor {
		t.Errorf("Expected cursor to remain unchanged after unhandled message")
	}
}

// TestView tests the View function
func TestView(t *testing.T) {
	// Create a styles instance for testing
	s := styles.DefaultStyles()
	
	// Create a new StoryList
	list := New(s)
	
	// Test view with empty list
	view := list.View()
	if view == "" {
		t.Errorf("Expected non-empty view for empty list")
	}
	if !strings.Contains(view, "No stories") {
		t.Errorf("Expected view to contain 'No stories' for empty list, got: %s", view)
	}
	
	// Create test stories
	stories := []models.UserStory{
		createTestUserStory(1),
		createTestUserStory(2),
	}
	
	// Set the items and focus the list
	list = list.SetItems(stories, nil).Focus()
	
	// Test view with items
	view = list.View()
	if view == "" {
		t.Errorf("Expected non-empty view for list with items")
	}
	
	// Check basic content
	// We don't check exact formatting because the styling can change
	if !strings.Contains(view, "Test User Story 1") {
		t.Errorf("Expected view to contain the title of the first story, got: %s", view)
	}
	
	// Check caching works
	list.needsRender = false
	cachedView := list.View()
	if cachedView != view {
		t.Errorf("Expected cached view to be returned when needsRender is false")
	}
} 
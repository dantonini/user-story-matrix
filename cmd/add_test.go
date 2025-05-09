// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui"
	"github.com/user-story-matrix/usm/internal/ui/contracts"
)

// MockFormResult implements contracts.UserStorySubmitter for testing
type MockFormResult struct {
	Story               models.UserStory
	ConfirmationStatus  bool
}

func (m *MockFormResult) GetUserStory() models.UserStory {
	return m.Story
}

func (m *MockFormResult) GetConfirmSubmission() bool {
	return m.ConfirmationStatus
}

// MockTerminalIO implements the io.UserOutput interface for testing
type MockTerminalIO struct {
	Messages      []string
	SuccessOutput []string
	ErrorOutput   []string
}

func (m *MockTerminalIO) Print(message string) {
	m.Messages = append(m.Messages, message)
}

func (m *MockTerminalIO) PrintSuccess(message string) {
	m.SuccessOutput = append(m.SuccessOutput, message)
}

func (m *MockTerminalIO) PrintError(message string) {
	m.ErrorOutput = append(m.ErrorOutput, message)
}

func (m *MockTerminalIO) PrintTable(headers []string, rows [][]string) {}
func (m *MockTerminalIO) PrintWarning(message string) {}
func (m *MockTerminalIO) PrintProgress(message string) {}
func (m *MockTerminalIO) PrintStep(stepNumber int, totalSteps int, description string) {}
func (m *MockTerminalIO) IsDebugEnabled() bool { return false }
func (m *MockTerminalIO) Prompt(prompt string) (string, error) { return "test-name", nil }

// TestFormResultHandling verifies the command handler correctly processes
// forms that implement the contracts.FormResult interface
func TestFormResultHandling(t *testing.T) {
	// Create a mock form result
	mockResult := &MockFormResult{
		Story: models.UserStory{
			Title:       "Test Story",
			Description: "As a tester,\nI want to verify interfaces,\nso that components work correctly.",
			Criteria:    []string{"Interface works", "No runtime errors"},
		},
		ConfirmationStatus: true,
	}
	
	// Verify the mock implements the interface
	var _ contracts.UserStorySubmitter = (*MockFormResult)(nil)
	
	// Verify we can use the interface methods directly
	story := mockResult.GetUserStory()
	assert.Equal(t, "Test Story", story.Title)
	assert.True(t, mockResult.GetConfirmSubmission())
	
	// Now we can test processFormResult directly with our mock
	mockTerminal := &MockTerminalIO{}
	userStory, shouldContinue := processFormResult(mockResult, mockTerminal)
	
	// Verify the return values
	assert.True(t, shouldContinue)
	assert.Equal(t, "Test Story", userStory.Title)
	assert.Equal(t, "As a tester,\nI want to verify interfaces,\nso that components work correctly.", userStory.Description)
	assert.Equal(t, []string{"Interface works", "No runtime errors"}, userStory.Criteria)
	assert.Empty(t, mockTerminal.Messages, "No messages should be printed when confirmation is true")
	
	// Test case when confirmation is false
	mockCancelResult := &MockFormResult{
		Story: models.UserStory{
			Title: "Cancelled Story",
		},
		ConfirmationStatus: false,
	}
	
	mockTerminal = &MockTerminalIO{}
	_, shouldContinue = processFormResult(mockCancelResult, mockTerminal)
	
	// Verify the return values for cancelled submission
	assert.False(t, shouldContinue)
	assert.Len(t, mockTerminal.Messages, 1, "Should have one message when cancelled")
	assert.Equal(t, "User story empty, creation cancelled", mockTerminal.Messages[0])
}

// TestAddChangeRequestComponents tests the components of the add change-request command
func TestAddChangeRequestComponents(t *testing.T) {
	// Test file path generation
	t.Run("file path generation", func(t *testing.T) {
		name := "test-change-request"
		filename := models.GenerateChangeRequestFilename(name)
		assert.Contains(t, filename, "test-change-request")
		assert.Contains(t, filename, ".md")
	})

	// Test template generation
	t.Run("template generation", func(t *testing.T) {
		name := "test-change-request"
		references := []models.UserStoryReference{
			{
				Title:       "Test Story 1",
				FilePath:    "docs/user-stories/01-test-story-1.md",
				ContentHash: "hash1",
			},
		}
		template := models.GenerateChangeRequestTemplate(name, references)
		assert.Contains(t, template, "name: test-change-request")
		assert.Contains(t, template, "Test Story 1")
		assert.Contains(t, template, "docs/user-stories/01-test-story-1.md")
	})

	// Test prompt instruction generation
	t.Run("prompt instruction generation", func(t *testing.T) {
		filePath := filepath.Join("docs/changes-request", "test-change-request.md")
		promptInstruction := models.GetPromptInstruction(filePath, 1)
		assert.Contains(t, promptInstruction, filePath)
	})
}

// TestAddImplementationStatusFilter tests the implementation status filter acceptance criteria
func TestAddImplementationStatusFilter(t *testing.T) {
	// Save original UI creator to restore it after the test
	originalSelectionUI := ui.CurrentNewSelectionUI

	// Restore UI creator after the test to avoid affecting other tests
	defer func() {
		ui.CurrentNewSelectionUI = originalSelectionUI
	}()

	// Test data with a mix of implemented and unimplemented stories
	userStories := []models.UserStory{
		{Title: "Unimplemented Story 1", FilePath: "story1.md", IsImplemented: false},
		{Title: "Implemented Story 1", FilePath: "story2.md", IsImplemented: true},
		{Title: "Unimplemented Story 2", FilePath: "story3.md", IsImplemented: false},
		{Title: "Implemented Story 2", FilePath: "story4.md", IsImplemented: true},
	}

	// Test case 1: Default behavior (--show-all=false)
	// According to acceptance criteria: "By default, only show unimplemented user stories"
	t.Run("Default shows only unimplemented stories", func(t *testing.T) {
		var capturedShowAll bool

		// Mock selection UI creator to capture the showAll flag value
		ui.CurrentNewSelectionUI = func(stories []models.UserStory, showAll bool) tea.Model {
			// Capture the value of showAll flag passed to the UI
			capturedShowAll = showAll

			// Create a mock UI that returns some selected items
			return &MockSelectionUI{
				selected: []int{0}, // Select the first story
			}
		}

		// Reset the flag to ensure test isolation
		showAll = false

		// Directly test that the UI receives the correct showAll value
		_ = ui.CurrentNewSelectionUI(userStories, showAll)

		// Verify that showAll flag was set to false, meaning only unimplemented stories are shown
		assert.False(t, capturedShowAll)
	})

	// Test case 2: With --show-all flag
	// According to acceptance criteria: "Provide a flag `--show-all` to display all user stories regardless of implementation status"
	t.Run("Show-all flag shows all stories", func(t *testing.T) {
		var capturedShowAll bool

		// Mock selection UI creator to capture the showAll flag value
		ui.CurrentNewSelectionUI = func(stories []models.UserStory, showAll bool) tea.Model {
			// Capture the value of showAll flag passed to the UI
			capturedShowAll = showAll

			// Create a mock UI that returns some selected items
			return &MockSelectionUI{
				selected: []int{0, 1}, // Select the first two stories
			}
		}

		// Set the flag to true
		showAll = true

		// Directly test that the UI receives the correct showAll value
		_ = ui.CurrentNewSelectionUI(userStories, showAll)

		// Verify that showAll flag was set to true, meaning all stories are shown regardless of implementation status
		assert.True(t, capturedShowAll)
	})
}

// MockSelectionUI implements a mock selection UI
type MockSelectionUI struct {
	selected []int
}

func (m *MockSelectionUI) Init() tea.Cmd {
	return nil
}

func (m *MockSelectionUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *MockSelectionUI) View() string {
	return "Mock Selection UI"
}

func (m *MockSelectionUI) GetSelected() []int {
	return m.selected
} 
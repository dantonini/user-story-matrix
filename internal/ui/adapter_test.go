// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package ui

import (
	"context"
	"errors"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/llm"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui/components/userstoryform"
)

// MockLLMProcessor is a mock implementation of the LLMProcessor interface for testing
type MockLLMProcessor struct {
	mock.Mock
}

func (m *MockLLMProcessor) ProcessUnstructuredText(ctx context.Context, text string) (llm.UserStoryData, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(llm.UserStoryData), args.Error(1)
}

func (m *MockLLMProcessor) GetConfidenceScores() map[string]float64 {
	args := m.Called()
	return args.Get(0).(map[string]float64)
}

func (m *MockLLMProcessor) IsConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLLMProcessor) ValidateConfiguration(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockLLMProcessor) Configure(config llm.APIKeyConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockLLMProcessor) GetProcessingState() llm.ProcessingState {
	args := m.Called()
	return args.Get(0).(llm.ProcessingState)
}

// TestCreateUserStoryFormWithLLMDisabled tests that the original form is returned when LLM is disabled
func TestCreateUserStoryFormWithLLMDisabled(t *testing.T) {
	// Arrange
	us := models.UserStory{
		Title: "Test User Story",
	}
	
	// Act
	form := CreateUserStoryFormWithLLM(us, false) // enableLLM = false
	
	// Assert
	assert.NotNil(t, form)
	
	// Verify it's the original form type by checking the type
	_, isOriginalForm := form.(*io.UserStoryForm)
	assert.True(t, isOriginalForm, "Expected original UserStoryForm when LLM is disabled")
}

// TestCreateUserStoryFormWithConfigError tests that the original form is returned when config loading fails
func TestCreateUserStoryFormWithConfigError(t *testing.T) {
	// Arrange
	us := models.UserStory{
		Title: "Test User Story",
	}
	
	// Create a mock filesystem that will return an error when LoadConfig is called
	mockFS := io.NewMockFileSystem()
	mockFS.SetReadFileError(".usm/config/llm_config.json", errors.New("simulated config error"))
	
	// Override the filesystem factory to return our mock
	oldFactory := fileSystemFactory
	fileSystemFactory = func() io.FileSystem {
		return mockFS
	}
	
	// Restore the original function when we're done
	defer func() {
		fileSystemFactory = oldFactory
	}()
	
	// Act
	form := CreateUserStoryFormWithLLM(us, true) // enableLLM = true, but config loading will fail
	
	// Assert
	assert.NotNil(t, form)
	
	// Print debugging info
	fmt.Printf("Form type: %T\n", form)
	
	// Verify it's the original form type by checking the type
	_, isOriginalForm := form.(*io.UserStoryForm)
	assert.True(t, isOriginalForm, "Expected original UserStoryForm when config loading fails")
}

// createTestConfigFileWithContent creates a config file with the specified content in the mock filesystem
func createTestConfigFileWithContent(mockFS *io.MockFileSystem, content string) {
	mockFS.AddFile(".usm/config/llm_config.json", []byte(content))
}

// TestCreateUserStoryFormWithValidConfig tests that the enhanced form is returned when everything works correctly
func TestCreateUserStoryFormWithValidConfig(t *testing.T) {
	// Arrange
	us := models.UserStory{
		Title: "Test User Story",
	}
	
	// Create a mock filesystem with a valid config
	mockFS := io.NewMockFileSystem()
	validConfig := `{"openai_key":"valid-api-key","is_valid":true,"last_validated":"2025-01-01T00:00:00Z"}`
	createTestConfigFileWithContent(mockFS, validConfig)
	
	// Override the filesystem factory
	oldFactory := fileSystemFactory
	fileSystemFactory = func() io.FileSystem {
		return mockFS
	}
	defer func() {
		fileSystemFactory = oldFactory
	}()
	
	// Act
	form := CreateUserStoryFormWithLLM(us, true)
	
	// Assert
	assert.NotNil(t, form)
	
	// Verify it's the enhanced form type by checking the type
	_, isEnhancedForm := form.(*userstoryform.UserStoryForm)
	assert.True(t, isEnhancedForm, "Expected enhanced UserStoryForm when config is valid")
}

// TestCreateUserStoryFormWithEmptyAPIKey tests that the original form is returned when the API key is empty
func TestCreateUserStoryFormWithEmptyAPIKey(t *testing.T) {
	// Arrange
	us := models.UserStory{
		Title: "Test User Story",
	}
	
	// Create a mock filesystem with an empty API key config
	mockFS := io.NewMockFileSystem()
	emptyConfig := `{"openai_key":"","is_valid":false,"last_validated":"2025-01-01T00:00:00Z"}`
	createTestConfigFileWithContent(mockFS, emptyConfig)
	
	// Override the filesystem factory
	oldFactory := fileSystemFactory
	fileSystemFactory = func() io.FileSystem {
		return mockFS
	}
	defer func() {
		fileSystemFactory = oldFactory
	}()
	
	// Act
	form := CreateUserStoryFormWithLLM(us, true)
	
	// Assert
	assert.NotNil(t, form)
	
	// With our updated logic, an empty API key should result in the original form
	_, isOriginalForm := form.(*io.UserStoryForm)
	assert.True(t, isOriginalForm, "Expected original UserStoryForm when API key is empty")
}

// TestNewSelectionAdapter tests the creation of a new selection adapter
func TestNewSelectionAdapter(t *testing.T) {
	// Arrange
	stories := []models.UserStory{
		{Title: "Test Story 1", FilePath: "path/to/story1.md"},
		{Title: "Test Story 2", FilePath: "path/to/story2.md"},
	}
	
	// Act
	adapter := NewSelectionAdapter(stories, true)
	
	// Assert
	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.page)
}

// TestSelectionAdapterInit tests the initialization of the selection adapter
func TestSelectionAdapterInit(t *testing.T) {
	// Arrange
	adapter := NewSelectionAdapter([]models.UserStory{}, true)
	
	// Act - we don't need to capture the return value
	adapter.Init()
	
	// Assert - We just check that the method runs without errors
	assert.NotPanics(t, func() {
		adapter.Init()
	})
}

// TestSelectionAdapterView tests the view method of the selection adapter
func TestSelectionAdapterView(t *testing.T) {
	// Arrange
	adapter := NewSelectionAdapter([]models.UserStory{}, true)
	
	// Act
	view := adapter.View()
	
	// Assert
	assert.NotEmpty(t, view, "View should return a non-empty string")
}

// TestSelectionAdapterGetSelected tests the GetSelected method of the selection adapter
func TestSelectionAdapterGetSelected(t *testing.T) {
	// Arrange
	stories := []models.UserStory{
		{Title: "Test Story 1", FilePath: "path/to/story1.md"},
		{Title: "Test Story 2", FilePath: "path/to/story2.md"},
	}
	adapter := NewSelectionAdapter(stories, true)
	
	// Act
	selected := adapter.GetSelected()
	
	// Assert - it's ok for the result to be nil or empty
	// We just check that the method runs without errors
	assert.IsType(t, []int{}, selected, "GetSelected should return a slice of integers")
}

// TestRegisterNewSelectionUIMaker tests the RegisterNewSelectionUIMaker function
func TestRegisterNewSelectionUIMaker(t *testing.T) {
	// Save current implementation
	prevImpl := CurrentNewSelectionUI
	
	// Set a test implementation
	testImpl := func(stories []models.UserStory, showAll bool) tea.Model {
		return nil
	}
	CurrentNewSelectionUI = testImpl
	
	// Call the function (which should do nothing)
	RegisterNewSelectionUIMaker()
	
	// Verify the implementation wasn't changed
	assert.NotNil(t, CurrentNewSelectionUI, "CurrentNewSelectionUI should not be nil")
	
	// Extra check that helps ensure code coverage
	assert.Equal(t, fmt.Sprintf("%p", testImpl), fmt.Sprintf("%p", CurrentNewSelectionUI), 
		"CurrentNewSelectionUI should remain unchanged")
	
	// Restore previous implementation
	CurrentNewSelectionUI = prevImpl
}

// TestFormatUIMessage tests the FormatUIMessage function
func TestFormatUIMessage(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		message  string
		style    string
		expected string
		contains string
	}{
		{
			name:     "Success style",
			message:  "Success message",
			style:    "success",
			contains: "✓ Success message",
		},
		{
			name:     "Error style",
			message:  "Error message",
			style:    "error",
			contains: "✗ Error message",
		},
		{
			name:     "Warning style",
			message:  "Warning message",
			style:    "warning",
			contains: "! Warning message",
		},
		{
			name:     "Info style",
			message:  "Info message",
			style:    "info",
			contains: "ℹ Info message",
		},
		{
			name:     "Default style",
			message:  "Default message",
			style:    "unknown",
			expected: "Default message",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := FormatUIMessage(tc.message, tc.style)
			
			// Assert
			if tc.expected != "" {
				assert.Equal(t, tc.expected, result)
			}
			if tc.contains != "" {
				assert.Contains(t, result, tc.contains)
			}
		})
	}
}

// TestSelectionAdapterUpdateWithMsgHandling tests the Update method with various message types
func TestSelectionAdapterUpdateWithMsgHandling(t *testing.T) {
	// Create an actual SelectionAdapter with real stories
	stories := []models.UserStory{
		{Title: "Test Story 1", FilePath: "path/to/story1.md"},
		{Title: "Test Story 2", FilePath: "path/to/story2.md"},
	}
	adapter := NewSelectionAdapter(stories, true)
	
	// Test with various message types to achieve full coverage
	testCases := []struct {
		name string
		msg  tea.Msg
	}{
		{
			name: "WindowSizeMsg",
			msg:  tea.WindowSizeMsg{Width: 100, Height: 50},
		},
		{
			name: "KeyMsg",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
		},
		{
			name: "MouseMsg",
			msg:  tea.MouseMsg{Type: tea.MouseLeft, X: 10, Y: 10},
		},
		{
			name: "Custom string message",
			msg:  "test message",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act 
			result, _ := adapter.Update(tc.msg)
			
			// Assert - we just need to make sure it doesn't panic
			// and returns a valid model
			assert.NotNil(t, result)
			
			// Make sure it's still our adapter type
			_, ok := result.(*SelectionAdapter)
			assert.True(t, ok)
		})
	}
} 
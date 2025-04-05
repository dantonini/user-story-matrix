package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user-story-matrix/usm/internal/llm"
	"github.com/user-story-matrix/usm/internal/models"
)

// TestGetFieldConfidenceExtended tests the GetFieldConfidence method
func TestGetFieldConfidenceExtended(t *testing.T) {
	// Arrange
	model := createTestModelExtended()
	model.ConfidenceScores = map[string]float64{
		"title":               0.9,
		"description":         0.8,
		"as_a":                0.7,
		"i_want":              0.6,
		"so_that":             0.5,
		"acceptance_criteria": 0.4,
	}

	// Test cases
	testCases := []struct {
		name           string
		fieldName      string
		expectedScore  float64
	}{
		{
			name:           "Title field",
			fieldName:      "title",
			expectedScore:  0.9,
		},
		{
			name:           "Description field",
			fieldName:      "description",
			expectedScore:  0.8,
		},
		{
			name:           "AsA field",
			fieldName:      "as_a",
			expectedScore:  0.7,
		},
		{
			name:           "IWant field",
			fieldName:      "i_want",
			expectedScore:  0.6,
		},
		{
			name:           "SoThat field",
			fieldName:      "so_that",
			expectedScore:  0.5,
		},
		{
			name:           "AcceptanceCriteria field",
			fieldName:      "acceptance_criteria",
			expectedScore:  0.4,
		},
		{
			name:           "Acceptance criteria 1 field",
			fieldName:      "acceptance_criteria_1",
			expectedScore:  0.4, // Should map to "acceptance_criteria"
		},
		{
			name:           "Unknown field",
			fieldName:      "unknown",
			expectedScore:  0.0, // Default score for unknown fields
		},
	}

	// Act & Assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := model.GetFieldConfidence(tc.fieldName)
			assert.Equal(t, tc.expectedScore, score)
		})
	}
}

// TestIsFieldAutoPopulatedExtended tests the IsFieldAutoPopulated method
func TestIsFieldAutoPopulatedExtended(t *testing.T) {
	// Arrange
	model := createTestModelExtended()
	model.AutoPopulatedFields = map[string]bool{
		"title":               true,
		"as_a":                true,
		"acceptance_criteria": true,
	}

	// Test cases
	testCases := []struct {
		name          string
		fieldName     string
		expectedValue bool
	}{
		{
			name:          "Auto-populated field",
			fieldName:     "title",
			expectedValue: true,
		},
		{
			name:          "Not auto-populated field",
			fieldName:     "description",
			expectedValue: false,
		},
		{
			name:          "Another auto-populated field",
			fieldName:     "as_a",
			expectedValue: true,
		},
		{
			name:          "Auto-populated acceptance criteria",
			fieldName:     "acceptance_criteria",
			expectedValue: true,
		},
		{
			name:          "Unknown field",
			fieldName:     "unknown",
			expectedValue: false,
		},
	}

	// Act & Assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isAutoPopulated := model.IsFieldAutoPopulated(tc.fieldName)
			assert.Equal(t, tc.expectedValue, isAutoPopulated)
		})
	}
}

// TestMarkFieldEditedExtended tests the MarkFieldEdited method
func TestMarkFieldEditedExtended(t *testing.T) {
	// Arrange
	model := createTestModelExtended()
	model.AutoPopulatedFields = map[string]bool{
		"title":               true,
		"description":         true,
		"as_a":                true,
		"i_want":              true,
		"so_that":             true,
		"acceptance_criteria": true,
	}

	// Test cases
	testCases := []struct {
		name             string
		fieldToEdit      string
		expectedFieldMap map[string]bool
	}{
		{
			name:        "Edit title field",
			fieldToEdit: "title",
			expectedFieldMap: map[string]bool{
				"description":         true,
				"as_a":                true,
				"i_want":              true,
				"so_that":             true,
				"acceptance_criteria": true,
			},
		},
		{
			name:        "Edit acceptance criteria field",
			fieldToEdit: "acceptance_criteria",
			expectedFieldMap: map[string]bool{
				"title":       true,
				"description": true,
				"as_a":        true,
				"i_want":      true,
				"so_that":     true,
			},
		},
		{
			name:        "Edit multiple fields",
			fieldToEdit: "i_want",
			expectedFieldMap: map[string]bool{
				"title":               true,
				"description":         true,
				"as_a":                true,
				"so_that":             true,
				"acceptance_criteria": true,
			},
		},
		{
			name:        "Edit unknown field",
			fieldToEdit: "unknown",
			expectedFieldMap: map[string]bool{
				"title":               true,
				"description":         true,
				"as_a":                true,
				"i_want":              true,
				"so_that":             true,
				"acceptance_criteria": true,
			},
		},
	}

	// Act & Assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset auto-populated fields for each test
			model.AutoPopulatedFields = map[string]bool{
				"title":               true,
				"description":         true,
				"as_a":                true,
				"i_want":              true,
				"so_that":             true,
				"acceptance_criteria": true,
			}
			
			// Mark the field as edited
			model.MarkFieldEdited(tc.fieldToEdit)
			
			// Verify the auto-populated fields
			assert.Equal(t, tc.expectedFieldMap, model.AutoPopulatedFields)
			assert.False(t, model.IsFieldAutoPopulated(tc.fieldToEdit))
		})
	}
}

// TestGetTimeoutMessageExtended tests the GetTimeoutMessage method
func TestGetTimeoutMessageExtended(t *testing.T) {
	// Arrange
	model := createTestModelExtended()
	model.TimeoutThreshold = 1 * time.Second

	// Test cases
	testCases := []struct {
		name            string
		processingState llm.ProcessingState
		timeSinceStart  time.Duration
		expectedMessage string
	}{
		{
			name:            "Not processing",
			processingState: llm.ProcessingIdle,
			timeSinceStart:  0,
			expectedMessage: "",
		},
		{
			name:            "Processing, no timeout",
			processingState: llm.ProcessingActive,
			timeSinceStart:  500 * time.Millisecond,
			expectedMessage: "",
		},
		{
			name:            "Processing, timeout reached",
			processingState: llm.ProcessingActive,
			timeSinceStart:  2 * time.Second,
			expectedMessage: "Processing is taking longer than expected. You can press ESC to cancel.",
		},
	}

	// Act & Assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model.ProcessingState = tc.processingState
			model.ProcessingStartTime = time.Now().Add(-tc.timeSinceStart)
			
			message := model.GetTimeoutMessage()
			assert.Equal(t, tc.expectedMessage, message)
		})
	}
}

// TestMarkAllFieldsEditedExtended tests the MarkAllFieldsEdited method
func TestMarkAllFieldsEditedExtended(t *testing.T) {
	// Arrange
	model := createTestModelExtended()
	model.AutoPopulatedFields = map[string]bool{
		"title":               true,
		"description":         true,
		"as_a":                true,
		"i_want":              true,
		"so_that":             true,
		"acceptance_criteria": true,
	}

	// Act
	model.MarkAllFieldsEdited()

	// Assert
	assert.Empty(t, model.AutoPopulatedFields)
	assert.False(t, model.IsFieldAutoPopulated("title"))
	assert.False(t, model.IsFieldAutoPopulated("description"))
	assert.False(t, model.IsFieldAutoPopulated("as_a"))
	assert.False(t, model.IsFieldAutoPopulated("i_want"))
	assert.False(t, model.IsFieldAutoPopulated("so_that"))
	assert.False(t, model.IsFieldAutoPopulated("acceptance_criteria"))
}

// Helper method to create a test model
func createTestModelExtended() *UserStoryFormModel {
	mockProcessor := new(MockLLMProcessorExtended)
	mockProcessor.On("IsConfigured").Return(true)
	
	return NewUserStoryFormModel(models.UserStory{
		Title:       "Test Story",
		Description: "Test Description",
		Criteria:    []string{"Criteria 1", "Criteria 2"},
	}, mockProcessor, nil)
}

// MockLLMProcessorExtended implements llm.LLMProcessor for testing
type MockLLMProcessorExtended struct {
	mock.Mock
}

func (m *MockLLMProcessorExtended) ProcessUnstructuredText(ctx context.Context, text string) (llm.UserStoryData, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(llm.UserStoryData), args.Error(1)
}

func (m *MockLLMProcessorExtended) GetConfidenceScores() map[string]float64 {
	args := m.Called()
	return args.Get(0).(map[string]float64)
}

func (m *MockLLMProcessorExtended) IsConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLLMProcessorExtended) ValidateConfiguration(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockLLMProcessorExtended) Configure(config llm.APIKeyConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockLLMProcessorExtended) GetProcessingState() llm.ProcessingState {
	args := m.Called()
	return args.Get(0).(llm.ProcessingState)
}

// TestShouldShowAPIKeyMessage tests the ShouldShowAPIKeyMessage method
func TestShouldShowAPIKeyMessage(t *testing.T) {
	// Arrange
	model := createTestModelExtended()
	
	// Test cases
	testCases := []struct {
		name           string
		showAPIKeyFlag bool
		expected       bool
	}{
		{
			name:           "API key message should be shown",
			showAPIKeyFlag: true,
			expected:       true,
		},
		{
			name:           "API key message should not be shown",
			showAPIKeyFlag: false,
			expected:       false,
		},
	}
	
	// Act & Assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model.ShowAPIKeyMessage = tc.showAPIKeyFlag
			result := model.ShouldShowAPIKeyMessage()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGetAPIKeyMessage tests the GetAPIKeyMessage method
func TestGetAPIKeyMessage(t *testing.T) {
	// Arrange
	model := createTestModelExtended()
	
	// Test that the method returns the expected message
	expectedMessage := "API key not configured. Please set your OpenAI API key in the settings to enable auto-formatting."
	
	// Act
	message := model.GetAPIKeyMessage()
	
	// Assert
	assert.Equal(t, expectedMessage, message)
}

// TestGetProcessingStateMethod tests the GetProcessingState method
func TestGetProcessingStateMethod(t *testing.T) {
	// Arrange
	model := createTestModelExtended()
	
	// Test cases for different processing states
	testCases := []struct {
		name            string
		processingState llm.ProcessingState
		expected        llm.ProcessingState
	}{
		{
			name:            "Idle state",
			processingState: llm.ProcessingIdle,
			expected:        llm.ProcessingIdle,
		},
		{
			name:            "Active state",
			processingState: llm.ProcessingActive,
			expected:        llm.ProcessingActive,
		},
		{
			name:            "Success state",
			processingState: llm.ProcessingSuccess,
			expected:        llm.ProcessingSuccess,
		},
		{
			name:            "Error state",
			processingState: llm.ProcessingError,
			expected:        llm.ProcessingError,
		},
		{
			name:            "Timeout state",
			processingState: llm.ProcessingTimeout,
			expected:        llm.ProcessingTimeout,
		},
		{
			name:            "Cancelled state",
			processingState: llm.ProcessingCancelled,
			expected:        llm.ProcessingCancelled,
		},
		{
			name:            "Not configured state",
			processingState: llm.ProcessingNotConfigured,
			expected:        llm.ProcessingNotConfigured,
		},
		{
			name:            "Partial success state",
			processingState: llm.ProcessingPartialSuccess,
			expected:        llm.ProcessingPartialSuccess,
		},
	}
	
	// Act & Assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model.ProcessingState = tc.processingState
			result := model.GetProcessingState()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsProcessingActive(t *testing.T) {
	t.Run("Active processing", func(t *testing.T) {
		// Setup
		model := UserStoryFormModel{
			ProcessingState: llm.ProcessingActive,
		}

		// Test
		isActive := model.IsProcessingActive()

		// Assert
		assert.True(t, isActive, "Processing should be active")
	})

	t.Run("Inactive processing", func(t *testing.T) {
		tests := []struct {
			name           string
			processingState llm.ProcessingState
		}{
			{"Idle state", llm.ProcessingIdle},
			{"Success state", llm.ProcessingSuccess},
			{"Error state", llm.ProcessingError},
			{"Not configured state", llm.ProcessingNotConfigured},
			{"Timeout state", llm.ProcessingTimeout},
			{"Cancelled state", llm.ProcessingCancelled},
			{"Partial success state", llm.ProcessingPartialSuccess},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				// Setup
				model := UserStoryFormModel{
					ProcessingState: tc.processingState,
				}

				// Test
				isActive := model.IsProcessingActive()

				// Assert
				assert.False(t, isActive, "Processing should not be active")
			})
		}
	})
}

func TestGetFormData(t *testing.T) {
	// Setup
	expectedFormData := FormData{
		Title:              "Test Title",
		Description:        "Test Description",
		AsA:                "Test User",
		IWant:              "Test Feature",
		SoThat:             "Test Benefit",
		AcceptanceCriteria: []string{"Criteria 1", "Criteria 2"},
	}

	model := UserStoryFormModel{
		FormData: expectedFormData,
	}

	// Test
	actualFormData := model.GetFormData()

	// Assert
	assert.Equal(t, expectedFormData, actualFormData, "GetFormData should return the form data from the model")
}

func TestGetLastError(t *testing.T) {
	t.Run("With error", func(t *testing.T) {
		// Setup
		testError := fmt.Errorf("test error")
		model := UserStoryFormModel{
			LastError: testError,
		}

		// Test
		errorString := model.GetLastError()

		// Assert
		assert.Equal(t, "test error", errorString, "GetLastError should return the error message as a string")
	})

	t.Run("No error", func(t *testing.T) {
		// Setup
		model := UserStoryFormModel{
			LastError: nil,
		}

		// Test
		errorString := model.GetLastError()

		// Assert
		assert.Equal(t, "", errorString, "GetLastError should return an empty string when there is no error")
	})
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestNewOpenAIProcessor(t *testing.T) {
	// Act
	processor := NewOpenAIProcessor()
	
	// Assert
	assert.NotNil(t, processor)
	assert.Equal(t, "gpt-4o-mini", processor.model)
	assert.Equal(t, 1000, processor.maxTokens)
	assert.Equal(t, float32(0.2), processor.temperature)
	assert.Empty(t, processor.apiKey)
	assert.False(t, processor.isConfigured)
	assert.Equal(t, ProcessingIdle, processor.processingState)
}

func TestNewOpenAIProcessorWithOptions(t *testing.T) {
	// Act
	processor := NewOpenAIProcessor(
		WithModel("test-model"),
		WithMaxTokens(500),
		WithTemperature(0.5),
		WithAPIKey("test-key"),
		WithTimeout(time.Second * 10),
	)
	
	// Assert
	assert.NotNil(t, processor)
	assert.Equal(t, "test-model", processor.model)
	assert.Equal(t, 500, processor.maxTokens)
	assert.Equal(t, float32(0.5), processor.temperature)
	assert.Equal(t, "test-key", processor.apiKey)
	assert.True(t, processor.isConfigured)
	assert.Equal(t, ProcessingIdle, processor.processingState)
}

func TestProcessUnstructuredTextWhenNotConfigured(t *testing.T) {
	// Arrange
	processor := NewOpenAIProcessor()
	
	// Act
	_, err := processor.ProcessUnstructuredText(context.Background(), "Test text")
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
	assert.Equal(t, ProcessingNotConfigured, processor.processingState)
}

func TestProcessUnstructuredTextWithValidConfiguration(t *testing.T) {
	// This test was failing because it was making a real API call
	// Skip this test as it's now covered by TestOpenAIProcessorWithMocks
	t.Skip("Skip this test as it's now covered by TestOpenAIProcessorWithMocks")
}

func TestGetConfidenceScores(t *testing.T) {
	// This test was failing because it was making a real API call
	// Skip this test as it's now covered by TestOpenAIProcessorWithMocks
	t.Skip("Skip this test as it's now covered by TestOpenAIProcessorWithMocks")
}

func TestIsConfigured(t *testing.T) {
	// Arrange
	processor1 := NewOpenAIProcessor()
	processor2 := NewOpenAIProcessor(WithAPIKey("test-key"))
	
	// Act & Assert
	assert.False(t, processor1.IsConfigured())
	assert.True(t, processor2.IsConfigured())
}

func TestValidateConfiguration(t *testing.T) {
	// Test with empty API key
	t.Run("Empty API Key", func(t *testing.T) {
		// Create processor with empty API key
		processor := NewOpenAIProcessor()
		
		// Call ValidateConfiguration method
		err := processor.ValidateConfiguration(context.Background())
		
		// Assert that it returns an error for missing API key
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key not provided")
		assert.False(t, processor.isConfigured)
	})
	
	// Test with cancelled context
	t.Run("Context Cancellation", func(t *testing.T) {
		// Create processor with a test API key
		processor := NewOpenAIProcessor(WithAPIKey("test-key"))
		
		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		
		// Call ValidateConfiguration with cancelled context
		err := processor.ValidateConfiguration(ctx)
		
		// Assert it returns context cancellation error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	// Skip actual API call test
	t.Run("Real API Call", func(t *testing.T) {
		t.Skip("Skipping test that would make a real API call")
	})
}

// Test all cases for the Configure method
func TestConfigureMethod(t *testing.T) {
	t.Run("Configure with valid API key", func(t *testing.T) {
		// Arrange
		processor := NewOpenAIProcessor()
		config := APIKeyConfig{
			OpenAIKey:     "valid-key",
			IsValid:       true,
			LastValidated: time.Now(),
		}
		
		// Act
		err := processor.Configure(config)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "valid-key", processor.apiKey)
		assert.True(t, processor.isConfigured)
		assert.NotNil(t, processor.client)
	})
	
	t.Run("Configure with empty API key", func(t *testing.T) {
		// Arrange
		processor := NewOpenAIProcessor()
		config := APIKeyConfig{
			OpenAIKey:     "",
			IsValid:       false,
			LastValidated: time.Time{},
		}
		
		// Act
		err := processor.Configure(config)
		
		// Assert
		assert.NoError(t, err) // Configure doesn't return errors
		assert.Equal(t, "", processor.apiKey)
		assert.False(t, processor.isConfigured)
		assert.Nil(t, processor.client)
	})
	
	t.Run("Reconfigure from valid to empty", func(t *testing.T) {
		// Arrange - start with valid config
		processor := NewOpenAIProcessor(WithAPIKey("initial-key"))
		assert.True(t, processor.isConfigured)
		assert.NotNil(t, processor.client)
		
		// Act - reconfigure to empty
		config := APIKeyConfig{
			OpenAIKey:     "",
			IsValid:       false,
			LastValidated: time.Time{},
		}
		err := processor.Configure(config)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "", processor.apiKey)
		assert.False(t, processor.isConfigured)
		assert.Nil(t, processor.client)
	})
	
	t.Run("Reconfigure from empty to valid", func(t *testing.T) {
		// Arrange - start with no config
		processor := NewOpenAIProcessor()
		assert.False(t, processor.isConfigured)
		assert.Nil(t, processor.client)
		
		// Act - reconfigure to valid
		config := APIKeyConfig{
			OpenAIKey:     "new-valid-key",
			IsValid:       true,
			LastValidated: time.Now(),
		}
		err := processor.Configure(config)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "new-valid-key", processor.apiKey)
		assert.True(t, processor.isConfigured)
		assert.NotNil(t, processor.client)
	})
}

func TestGetProcessingState(t *testing.T) {
	// Arrange
	processor := NewOpenAIProcessor()
	
	// Initial state
	assert.Equal(t, ProcessingIdle, processor.GetProcessingState())
	
	// Change state
	processor.processingState = ProcessingActive
	assert.Equal(t, ProcessingActive, processor.GetProcessingState())
	
	// Process text with no API key
	processor.ProcessUnstructuredText(context.Background(), "Test text")
	assert.Equal(t, ProcessingNotConfigured, processor.GetProcessingState())
}

// MockOpenAIAPIClient implements the OpenAIClientInterface for testing
type MockOpenAIAPIClient struct {
	ReturnError     error
	ReturnResponse  string
	ReturnedChoices []openai.ChatCompletionChoice
	CallCount       int
}

// CreateChatCompletion implements the OpenAIClientInterface for testing
func (m *MockOpenAIAPIClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	m.CallCount++
	
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return openai.ChatCompletionResponse{}, ctx.Err()
	default:
		// Continue processing
	}
	
	if m.ReturnError != nil {
		return openai.ChatCompletionResponse{}, m.ReturnError
	}
	
	choices := m.ReturnedChoices
	if len(choices) == 0 && m.ReturnResponse != "" {
		choices = []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: m.ReturnResponse,
				},
				FinishReason: openai.FinishReasonStop,
			},
		}
	}
	
	return openai.ChatCompletionResponse{
		ID:      "test-response-id",
		Choices: choices,
	}, nil
}

func TestProcessUnstructuredTextWithValidResponse(t *testing.T) {
	// Sample valid JSON response from OpenAI
	validResponse := `{
		"title": "User Authentication System",
		"description": "Implement a secure authentication system",
		"as_a": "system administrator",
		"i_want": "to manage user access securely",
		"so_that": "sensitive data remains protected",
		"acceptance_criteria": ["Secure password storage", "Two-factor authentication", "Failed login tracking"],
		"confidence": {
			"title": 0.9,
			"description": 0.8,
			"as_a": 0.95,
			"i_want": 0.85,
			"so_that": 0.9,
			"acceptance_criteria": 0.85
		}
	}`

	// Create mock client
	mockClient := &MockOpenAIAPIClient{
		ReturnResponse: validResponse,
	}

	// Create processor with mock client
	processor := NewOpenAIProcessor(WithAPIKey("test-api-key"))
	processor.client = mockClient
	processor.isConfigured = true

	// Process text
	ctx := context.Background()
	result, err := processor.ProcessUnstructuredText(ctx, "Sample unstructured text")

	// Verify there was no error
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify call count
	if mockClient.CallCount != 1 {
		t.Errorf("Expected 1 API call, got %d", mockClient.CallCount)
	}

	// Verify results
	if result.Title != "User Authentication System" {
		t.Errorf("Expected title 'User Authentication System', got '%s'", result.Title)
	}
	if len(result.AcceptanceCriteria) != 3 {
		t.Errorf("Expected 3 acceptance criteria, got %d", len(result.AcceptanceCriteria))
	}

	// Verify confidence scores were stored
	scores := processor.GetConfidenceScores()
	if scores["title"] != 0.9 {
		t.Errorf("Expected title confidence 0.9, got %f", scores["title"])
	}

	// Verify processing state
	if processor.GetProcessingState() != ProcessingSuccess {
		t.Errorf("Expected processing state Success, got %v", processor.GetProcessingState())
	}
}

func TestProcessUnstructuredTextWithPartialResponse(t *testing.T) {
	// Sample partial JSON response from OpenAI
	partialResponse := `{
		"title": "User Authentication",
		"description": "",
		"as_a": "",
		"i_want": "to implement authentication",
		"so_that": "",
		"acceptance_criteria": ["Secure login"],
		"confidence": {
			"title": 0.7,
			"description": 0.1,
			"as_a": 0.1,
			"i_want": 0.6,
			"so_that": 0.1,
			"acceptance_criteria": 0.5
		}
	}`

	// Create mock client
	mockClient := &MockOpenAIAPIClient{
		ReturnResponse: partialResponse,
	}

	// Create processor with mock client
	processor := NewOpenAIProcessor(WithAPIKey("test-api-key"))
	processor.client = mockClient
	processor.isConfigured = true

	// Process text
	ctx := context.Background()
	_, err := processor.ProcessUnstructuredText(ctx, "Sample unstructured text")

	// Verify there was no error (partial success is still success)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify processing state
	if processor.GetProcessingState() != ProcessingPartialSuccess {
		t.Errorf("Expected processing state PartialSuccess, got %v", processor.GetProcessingState())
	}
}

func TestProcessUnstructuredTextWithEmptyResponse(t *testing.T) {
	// Create mock client that returns empty choices
	mockClient := &MockOpenAIAPIClient{
		ReturnedChoices: []openai.ChatCompletionChoice{},
	}

	// Create processor with mock client
	processor := NewOpenAIProcessor(WithAPIKey("test-api-key"))
	processor.client = mockClient
	processor.isConfigured = true

	// Process text
	ctx := context.Background()
	_, err := processor.ProcessUnstructuredText(ctx, "Sample unstructured text")

	// Verify there was an error
	if err == nil {
		t.Error("Expected error for empty response, got nil")
	}

	// Verify processing state
	if processor.GetProcessingState() != ProcessingError {
		t.Errorf("Expected processing state Error, got %v", processor.GetProcessingState())
	}
}

func TestProcessUnstructuredTextWithInvalidJSON(t *testing.T) {
	// Sample invalid JSON response from OpenAI
	invalidResponse := `{"title": "Invalid JSON response missing closing bracket`

	// Create mock client
	mockClient := &MockOpenAIAPIClient{
		ReturnResponse: invalidResponse,
	}

	// Create processor with mock client
	processor := NewOpenAIProcessor(WithAPIKey("test-api-key"))
	processor.client = mockClient
	processor.isConfigured = true

	// Process text
	ctx := context.Background()
	_, err := processor.ProcessUnstructuredText(ctx, "Sample unstructured text")

	// Verify there was an error
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	// Verify processing state
	if processor.GetProcessingState() != ProcessingError {
		t.Errorf("Expected processing state Error, got %v", processor.GetProcessingState())
	}
}

func TestProcessUnstructuredTextWithAPIError(t *testing.T) {
	// Create mock client that returns an error
	mockClient := &MockOpenAIAPIClient{
		ReturnError: errors.New("API rate limit exceeded"),
	}

	// Create processor with mock client
	processor := NewOpenAIProcessor(WithAPIKey("test-api-key"))
	processor.client = mockClient
	processor.isConfigured = true

	// Process text
	ctx := context.Background()
	_, err := processor.ProcessUnstructuredText(ctx, "Sample unstructured text")

	// Verify there was an error
	if err == nil {
		t.Error("Expected error from API, got nil")
	}

	// Verify processing state
	if processor.GetProcessingState() != ProcessingError {
		t.Errorf("Expected processing state Error, got %v", processor.GetProcessingState())
	}
}

func TestProcessUnstructuredTextWithCancellation(t *testing.T) {
	// Create a processor with mock client
	processor := NewOpenAIProcessor(WithAPIKey("test-api-key"))
	processor.client = &MockOpenAIAPIClient{
		// Use a sleep in the mock to simulate a long-running request
		ReturnResponse: `{"title": "This would be returned if not cancelled"}`,
	}
	processor.isConfigured = true

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	
	// Cancel the context before the API call completes
	cancel()

	// Process text with cancelled context
	_, err := processor.ProcessUnstructuredText(ctx, "Sample unstructured text")

	// Verify there was a cancellation error
	if err == nil || !strings.Contains(err.Error(), "canceled") {
		t.Errorf("Expected context cancellation error, got %v", err)
	}

	// Verify processing state
	if processor.GetProcessingState() != ProcessingCancelled {
		t.Errorf("Expected processing state Cancelled, got %v", processor.GetProcessingState())
	}
}

// Test isRetryableError function
func TestIsRetryableError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection reset error",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      errors.New("request timeout after 10s"),
			expected: true,
		},
		{
			name:     "rate limit error",
			err:      errors.New("rate limit exceeded, try again later"),
			expected: true,
		},
		{
			name:     "server error",
			err:      errors.New("internal server error occurred"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      errors.New("invalid request format"),
			expected: false,
		},
		{
			name:     "service unavailable error",
			err:      errors.New("service unavailable - try again later"),
			expected: true,
		},
		{
			name:     "bad gateway error",
			err:      errors.New("bad gateway error occurred"),
			expected: true,
		},
		{
			name:     "mixed case error message",
			err:      errors.New("CONNECTION RESET by peer"),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isRetryableError(tc.err)
			if result != tc.expected {
				t.Errorf("Expected isRetryableError(%v) to be %v, got %v", tc.err, tc.expected, result)
			}
		})
	}
}

func TestValidateConfigurationWithMock(t *testing.T) {
	// Test valid configuration
	t.Run("Valid Configuration", func(t *testing.T) {
		// Create a mock client that returns success
		mockClient := &MockOpenAIAPIClient{}
		
		// Create a mock processor with the existing pattern
		processor := newMockProcessor()
		
		// Set our mock client
		processor.client = mockClient
		
		// Set up the mock behavior
		processor.validateConfigFunc = func(ctx context.Context) error {
			// This will be called instead of the real implementation
			mockClient.CallCount++
			processor.isConfigured = true
			return nil
		}
		
		// Test the validation
		err := processor.ValidateConfiguration(context.Background())
		
		// Verify no error was returned
		assert.NoError(t, err)
		
		// Verify the processor is now configured
		assert.True(t, processor.IsConfigured())
		
		// Verify the mock was called
		assert.Equal(t, 1, mockClient.CallCount)
	})
	
	// Test empty API key
	t.Run("Empty API Key", func(t *testing.T) {
		// Create a mock processor
		processor := newMockProcessor()
		processor.apiKey = "" // Set API key to empty
		
		// Set up the mock behavior
		processor.validateConfigFunc = func(ctx context.Context) error {
			// Should return error without calling the mock client
			processor.isConfigured = false
			return errors.New("OpenAI API key not provided")
		}
		
		// Test the validation
		err := processor.ValidateConfiguration(context.Background())
		
		// Verify the error was returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key not provided")
		
		// Verify the processor is not configured
		assert.False(t, processor.IsConfigured())
	})
	
	// Test API validation error
	t.Run("API Validation Error", func(t *testing.T) {
		// Create a mock client that returns an error
		mockClient := &MockOpenAIAPIClient{
			ReturnError: errors.New("invalid API key"),
		}
		
		// Create a mock processor
		processor := newMockProcessor()
		processor.client = mockClient
		
		// Set up the mock behavior
		processor.validateConfigFunc = func(ctx context.Context) error {
			// Simulate API error
			mockClient.CallCount++
			processor.isConfigured = false
			return errors.New("failed to validate OpenAI API key: invalid API key")
		}
		
		// Test the validation
		err := processor.ValidateConfiguration(context.Background())
		
		// Verify the error was returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate OpenAI API key")
		
		// Verify the processor is not configured
		assert.False(t, processor.IsConfigured())
		
		// Verify the mock was called
		assert.Equal(t, 1, mockClient.CallCount)
	})
	
	// Test validation with context cancellation
	t.Run("Context Cancellation", func(t *testing.T) {
		// Create a mock processor
		processor := newMockProcessor()
		
		// Set up the mock behavior to simulate context cancellation
		processor.validateConfigFunc = func(ctx context.Context) error {
			// Check context cancellation
			select {
			case <-ctx.Done():
				// Make sure to set isConfigured to false when cancellation occurs
				processor.isConfigured = false
				return ctx.Err()
			default:
				t.Error("Context should be cancelled")
				return nil
			}
		}
		
		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		
		// Attempt validation with cancelled context
		err := processor.ValidateConfiguration(ctx)
		
		// Error should contain reference to context cancellation
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
		
		// Processor should not be configured after error
		assert.False(t, processor.isConfigured)
	})
}

// Test for missing confidence scores in response
func TestProcessUnstructuredTextWithMissingConfidence(t *testing.T) {
	// Sample response with missing confidence scores
	response := `{
		"title": "Feature Request",
		"description": "Add new feature",
		"as_a": "user",
		"i_want": "to use new feature",
		"so_that": "I can be more productive",
		"acceptance_criteria": ["It works"]
	}`

	mockClient := &MockOpenAIAPIClient{
		ReturnResponse: response,
	}

	processor := NewOpenAIProcessor(WithAPIKey("test-api-key"))
	processor.client = mockClient
	processor.isConfigured = true

	// Process text
	ctx := context.Background()
	result, err := processor.ProcessUnstructuredText(ctx, "Sample text")

	// Should not error even with missing confidence
	assert.NoError(t, err)
	
	// Should add default confidence scores
	scores := processor.GetConfidenceScores()
	assert.Equal(t, 0.0, scores["title"])
	assert.Equal(t, 0.0, scores["description"])
	assert.Equal(t, 0.0, scores["as_a"])
	assert.Equal(t, 0.0, scores["i_want"])
	assert.Equal(t, 0.0, scores["so_that"])
	assert.Equal(t, 0.0, scores["acceptance_criteria"])
	
	// Verify the response was parsed correctly
	assert.Equal(t, "Feature Request", result.Title)
	assert.Equal(t, "user", result.AsA)
}

// Test for processing with no meaningful data extracted
func TestProcessUnstructuredTextWithNoMeaningfulData(t *testing.T) {
	// Sample response with empty fields
	response := `{
		"title": "",
		"description": "",
		"as_a": "",
		"i_want": "",
		"so_that": "",
		"acceptance_criteria": [],
		"confidence": {
			"title": 0.1,
			"description": 0.1,
			"as_a": 0.1,
			"i_want": 0.1,
			"so_that": 0.1,
			"acceptance_criteria": 0.1
		}
	}`

	mockClient := &MockOpenAIAPIClient{
		ReturnResponse: response,
	}

	processor := NewOpenAIProcessor(WithAPIKey("test-api-key"))
	processor.client = mockClient
	processor.isConfigured = true

	// Process text
	ctx := context.Background()
	_, err := processor.ProcessUnstructuredText(ctx, "Sample text")

	// Should error with no meaningful data
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract any meaningful data")
	assert.Equal(t, ProcessingError, processor.processingState)
} 
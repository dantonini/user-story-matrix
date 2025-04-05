// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package llm

import (
	"context"
	"testing"
	"time"

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
	// Arrange
	processor := NewOpenAIProcessor(WithAPIKey("test-key"))
	
	// Act
	result, err := processor.ProcessUnstructuredText(context.Background(), "Test text")
	
	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Title)
	assert.NotEmpty(t, result.Description)
	assert.NotEmpty(t, result.AsA)
	assert.NotEmpty(t, result.IWant)
	assert.NotEmpty(t, result.SoThat)
	assert.NotEmpty(t, result.AcceptanceCriteria)
	assert.NotEmpty(t, result.Confidence)
	assert.Equal(t, ProcessingSuccess, processor.processingState)
}

func TestGetConfidenceScores(t *testing.T) {
	// Arrange
	processor := NewOpenAIProcessor(WithAPIKey("test-key"))
	
	// Process some text to populate confidence scores
	processor.ProcessUnstructuredText(context.Background(), "Test text")
	
	// Act
	scores := processor.GetConfidenceScores()
	
	// Assert
	assert.NotEmpty(t, scores)
	assert.Contains(t, scores, "title")
	assert.Contains(t, scores, "description")
	assert.Contains(t, scores, "as_a")
	assert.Contains(t, scores, "i_want")
	assert.Contains(t, scores, "so_that")
	assert.Contains(t, scores, "acceptance_criteria")
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
	// Arrange
	processor1 := NewOpenAIProcessor()
	processor2 := NewOpenAIProcessor(WithAPIKey("test-key"))
	
	// Act
	err1 := processor1.ValidateConfiguration(context.Background())
	err2 := processor2.ValidateConfiguration(context.Background())
	
	// Assert
	assert.Error(t, err1)
	assert.NoError(t, err2)
	assert.False(t, processor1.isConfigured)
	assert.True(t, processor2.isConfigured)
}

func TestConfigure(t *testing.T) {
	// Arrange
	processor := NewOpenAIProcessor()
	config := APIKeyConfig{
		OpenAIKey: "test-key",
		IsValid:   true,
		LastValidated: time.Now(),
	}
	
	// Act
	err := processor.Configure(config)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test-key", processor.apiKey)
	assert.True(t, processor.isConfigured)
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
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package llm

import (
	"context"
	"time"
)

// ProcessingState represents the current state of LLM processing
type ProcessingState int

const (
	// ProcessingIdle indicates no active processing
	ProcessingIdle ProcessingState = iota
	// ProcessingActive indicates active processing
	ProcessingActive
	// ProcessingSuccess indicates successful processing
	ProcessingSuccess
	// ProcessingError indicates an error during processing
	ProcessingError
	// ProcessingTimeout indicates a timeout during processing
	ProcessingTimeout
	// ProcessingCancelled indicates the processing was cancelled
	ProcessingCancelled
	// ProcessingNotConfigured indicates the processor is not configured (e.g., missing API key)
	ProcessingNotConfigured
	// ProcessingPartialSuccess indicates the processing was partially successful (some fields may be missing or incomplete)
	ProcessingPartialSuccess
)

// UserStoryData represents the structured data parsed from unstructured text
type UserStoryData struct {
	// Title of the user story
	Title string `json:"title"`
	
	// Description of the user story
	Description string `json:"description"`
	
	// AsA is the user type in the user story format "As a..."
	AsA string `json:"as_a"`
	
	// IWant is the capability in the user story format "I want..."
	IWant string `json:"i_want"`
	
	// SoThat is the benefit in the user story format "so that..."
	SoThat string `json:"so_that"`
	
	// AcceptanceCriteria is a list of acceptance criteria
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	
	// Confidence scores for each field
	Confidence map[string]float64 `json:"confidence"`
}

// APIKeyConfig represents the configuration for API keys
type APIKeyConfig struct {
	// OpenAIKey is the API key for OpenAI
	OpenAIKey string `json:"openai_key"`
	
	// IsValid indicates whether the key has been validated
	IsValid bool `json:"is_valid"`
	
	// LastValidated is the time when the key was last validated
	LastValidated time.Time `json:"last_validated"`
}

// LLMProcessor defines the interface for LLM processing capabilities
type LLMProcessor interface {
	// ProcessUnstructuredText takes unstructured text and returns structured user story data
	ProcessUnstructuredText(ctx context.Context, text string) (UserStoryData, error)
	
	// GetConfidenceScores returns confidence scores for each parsed field
	GetConfidenceScores() map[string]float64
	
	// IsConfigured returns whether the processor has been properly configured (API key, etc.)
	IsConfigured() bool
	
	// ValidateConfiguration validates the current configuration
	ValidateConfiguration(ctx context.Context) error
	
	// Configure sets up the processor with the provided configuration
	Configure(config APIKeyConfig) error
	
	// GetProcessingState returns the current processing state
	GetProcessingState() ProcessingState
}

// LLMConfig represents the configuration options for the LLM processor
type LLMConfig struct {
	// APIKey is the API key for the LLM service
	APIKey string
	
	// Model is the model name to use
	Model string
	
	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int
	
	// Temperature controls the randomness of the model output
	Temperature float32
	
	// Timeout is the maximum time to wait for a response
	Timeout time.Duration
}

// LLMProcessor factories and constructor
// These will be implemented by specific providers (e.g., OpenAI)

// WithConfiguration sets the configuration for the processor
type WithConfiguration func(*LLMConfig)

// WithModel sets the model name
func WithModel(model string) WithConfiguration {
	return func(c *LLMConfig) {
		c.Model = model
	}
}

// WithMaxTokens sets the maximum number of tokens
func WithMaxTokens(maxTokens int) WithConfiguration {
	return func(c *LLMConfig) {
		c.MaxTokens = maxTokens
	}
}

// WithTemperature sets the temperature
func WithTemperature(temperature float32) WithConfiguration {
	return func(c *LLMConfig) {
		c.Temperature = temperature
	}
}

// WithTimeout sets the timeout
func WithTimeout(timeout time.Duration) WithConfiguration {
	return func(c *LLMConfig) {
		c.Timeout = timeout
	}
}

// WithAPIKey sets the API key
func WithAPIKey(apiKey string) WithConfiguration {
	return func(c *LLMConfig) {
		c.APIKey = apiKey
	}
} 
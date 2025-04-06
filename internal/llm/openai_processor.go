// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProcessor implements the LLMProcessor interface for OpenAI
type OpenAIProcessor struct {
	// client represents the OpenAI API client
	client *openai.Client
	
	// apiKey is the OpenAI API key
	apiKey string
	
	// model is the OpenAI model to use (default: "gpt-4o-mini")
	model string
	
	// maxTokens is the maximum number of tokens to generate (default: 1000)
	maxTokens int
	
	// temperature controls the randomness of the model output (default: 0.2)
	temperature float32
	
	// isConfigured indicates whether the processor has been properly configured
	isConfigured bool
	
	// processingState tracks the current state of LLM processing
	processingState ProcessingState
	
	// confidenceScores stores the confidence scores for each parsed field
	confidenceScores map[string]float64
}

// NewOpenAIProcessor creates a new OpenAI processor with the specified configuration
func NewOpenAIProcessor(options ...WithConfiguration) *OpenAIProcessor {
	// Default configuration
	config := &LLMConfig{
		Model:       "gpt-4o-mini",
		MaxTokens:   1000,
		Temperature: 0.2,
		Timeout:     time.Second * 30,
	}
	
	// Apply options
	for _, option := range options {
		option(config)
	}
	
	// Create the processor
	processor := &OpenAIProcessor{
		model:           config.Model,
		maxTokens:       config.MaxTokens,
		temperature:     config.Temperature,
		apiKey:          config.APIKey,
		isConfigured:    config.APIKey != "",
		processingState: ProcessingIdle,
		confidenceScores: make(map[string]float64),
	}
	
	// Initialize OpenAI client if API key is provided
	if processor.apiKey != "" {
		processor.client = openai.NewClient(processor.apiKey)
	}
	
	return processor
}

// ProcessUnstructuredText takes unstructured text and returns structured user story data
func (p *OpenAIProcessor) ProcessUnstructuredText(ctx context.Context, text string) (UserStoryData, error) {
	// If not configured, return error
	if !p.isConfigured || p.client == nil {
		p.processingState = ProcessingNotConfigured
		return UserStoryData{}, errors.New("OpenAI processor not configured: missing API key")
	}
	
	// Set processing state to active
	p.processingState = ProcessingActive
	
	// Define the structured output format
	responseFormat := &openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONObject,
	}
	
	// Create the structured output schema prompt
	systemPrompt := `You are an AI trained to extract user story information from unstructured text.
Extract the following components in JSON format:
- title: A concise title for the user story (25 words or less)
- description: A detailed description (if available)
- as_a: The user type or role (from "As a...")
- i_want: The capability or feature (from "I want...")
- so_that: The benefit or reason (from "So that...")
- acceptance_criteria: Array of acceptance criteria (bullet points or separate items)

Also include a confidence score (0.0 to 1.0) for each field to indicate how confident you are in the extraction.

Your response must be a valid JSON object that matches this structure:
{
  "title": "string",
  "description": "string",
  "as_a": "string",
  "i_want": "string",
  "so_that": "string",
  "acceptance_criteria": ["string", "string", ...],
  "confidence": {
    "title": float,
    "description": float,
    "as_a": float,
    "i_want": float,
    "so_that": float,
    "acceptance_criteria": float
  }
}

If you can't extract a field, provide an empty string for that field or empty array for acceptance criteria, with a low confidence score.`

	// Create the chat completion request
	req := openai.ChatCompletionRequest{
		Model:          p.model,
		Messages:       []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		MaxTokens:     p.maxTokens,
		Temperature:   p.temperature,
		ResponseFormat: responseFormat,
	}
	
	// Try to send the request with retries and exponential backoff
	var resp openai.ChatCompletionResponse
	var err error
	
	maxRetries := 3
	retryDelay := 500 * time.Millisecond
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check if the context has been canceled
		select {
		case <-ctx.Done():
			p.processingState = ProcessingCancelled
			return UserStoryData{}, ctx.Err()
		default:
			// Continue with the request
		}
		
		// Send the request to the OpenAI API
		resp, err = p.client.CreateChatCompletion(ctx, req)
		
		// If successful or non-retryable error, break the loop
		if err == nil {
			break
		}
		
		// Check if this is a retryable error (rate limiting, network issues, etc.)
		if isRetryableError(err) && attempt < maxRetries-1 {
			// Exponential backoff
			select {
			case <-ctx.Done():
				p.processingState = ProcessingCancelled
				return UserStoryData{}, ctx.Err()
			case <-time.After(retryDelay):
				// Increase delay for next retry
				retryDelay *= 2
				continue
			}
		}
		
		// If we get here, it's a non-retryable error or we've exhausted retries
		p.processingState = ProcessingError
		return UserStoryData{}, fmt.Errorf("OpenAI API error after %d attempts: %w", attempt+1, err)
	}
	
	// Check if we got a response
	if len(resp.Choices) == 0 {
		p.processingState = ProcessingError
		return UserStoryData{}, errors.New("OpenAI API returned empty response")
	}
	
	// Parse the response
	content := resp.Choices[0].Message.Content
	
	// Unmarshal the JSON response
	var userData UserStoryData
	err = json.Unmarshal([]byte(content), &userData)
	if err != nil {
		p.processingState = ProcessingError
		return UserStoryData{}, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}
	
	// Validate the response has required fields
	if userData.Confidence == nil {
		userData.Confidence = make(map[string]float64)
	}
	
	// Add default confidence values for missing fields
	requiredFields := []string{"title", "description", "as_a", "i_want", "so_that", "acceptance_criteria"}
	for _, field := range requiredFields {
		if _, ok := userData.Confidence[field]; !ok {
			userData.Confidence[field] = 0.0
		}
	}
	
	// Clean up the parsed data
	userData.Title = strings.TrimSpace(userData.Title)
	userData.Description = strings.TrimSpace(userData.Description)
	userData.AsA = strings.TrimSpace(userData.AsA)
	userData.IWant = strings.TrimSpace(userData.IWant)
	userData.SoThat = strings.TrimSpace(userData.SoThat)
	
	// Clean acceptance criteria
	for i, ac := range userData.AcceptanceCriteria {
		userData.AcceptanceCriteria[i] = strings.TrimSpace(ac)
	}
	
	// Filter empty acceptance criteria
	filteredAC := make([]string, 0, len(userData.AcceptanceCriteria))
	for _, ac := range userData.AcceptanceCriteria {
		if ac != "" {
			filteredAC = append(filteredAC, ac)
		}
	}
	userData.AcceptanceCriteria = filteredAC
	
	// Store confidence scores
	p.confidenceScores = userData.Confidence
	
	// Check if we have at least some valid data
	if userData.Title == "" && userData.AsA == "" && userData.IWant == "" && len(userData.AcceptanceCriteria) == 0 {
		p.processingState = ProcessingError
		return userData, errors.New("failed to extract any meaningful data from the text")
	}
	
	// Check if we have partial data
	if userData.Title == "" || userData.AsA == "" || userData.IWant == "" {
		p.processingState = ProcessingPartialSuccess
	} else {
		// Set processing state to success
		p.processingState = ProcessingSuccess
	}
	
	return userData, nil
}

// isRetryableError determines if an error should be retried
func isRetryableError(err error) bool {
	// Check for network errors, timeouts, and rate limiting
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	
	// Check for common retryable error patterns
	retryablePatterns := []string{
		"connection reset", 
		"timeout", 
		"too many requests",
		"rate limit",
		"internal server error",
		"service unavailable",
		"bad gateway",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}
	
	return false
}

// GetConfidenceScores returns confidence scores for each parsed field
func (p *OpenAIProcessor) GetConfidenceScores() map[string]float64 {
	return p.confidenceScores
}

// IsConfigured returns whether the processor has been properly configured (API key, etc.)
func (p *OpenAIProcessor) IsConfigured() bool {
	return p.isConfigured && p.client != nil
}

// ValidateConfiguration validates the current configuration
func (p *OpenAIProcessor) ValidateConfiguration(ctx context.Context) error {
	if p.apiKey == "" {
		p.isConfigured = false
		return errors.New("OpenAI API key not provided")
	}
	
	// Create or update the client
	p.client = openai.NewClient(p.apiKey)
	
	// Make a test request to validate the API key
	req := openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello, this is a test request to validate the API key.",
			},
		},
		MaxTokens:   10,
		Temperature: 0,
	}
	
	_, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		p.isConfigured = false
		return fmt.Errorf("failed to validate OpenAI API key: %w", err)
	}
	
	p.isConfigured = true
	return nil
}

// Configure sets up the processor with the provided configuration
func (p *OpenAIProcessor) Configure(config APIKeyConfig) error {
	p.apiKey = config.OpenAIKey
	p.isConfigured = config.IsValid
	
	// Initialize/reinitialize OpenAI client with the new API key if provided
	if p.apiKey != "" {
		p.client = openai.NewClient(p.apiKey)
	} else {
		p.client = nil
		p.isConfigured = false
	}
	
	return nil
}

// GetProcessingState returns the current processing state
func (p *OpenAIProcessor) GetProcessingState() ProcessingState {
	return p.processingState
} 
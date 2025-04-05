// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package models

import (
	"context"
	"strings"
	"time"

	"github.com/user-story-matrix/usm/internal/llm"
	"github.com/user-story-matrix/usm/internal/models"
)

// UserStoryFormModel represents the data model for the user story form
type UserStoryFormModel struct {
	// UserStory holds the current user story being edited
	UserStory models.UserStory
	
	// LLMProcessor is the processor for unstructured text
	LLMProcessor llm.LLMProcessor
	
	// ConfigManager manages API key configuration
	ConfigManager *llm.ConfigManager
	
	// AutoPopulatedFields tracks which fields were auto-populated by the LLM
	AutoPopulatedFields map[string]bool
	
	// ConfidenceScores stores the confidence score for each field
	ConfidenceScores map[string]float64
	
	// ProcessingState tracks the current state of LLM processing
	ProcessingState llm.ProcessingState
	
	// ProcessingStartTime tracks when the LLM processing started
	ProcessingStartTime time.Time
	
	// TimeoutThreshold is the threshold for displaying a timeout warning (default: 5s)
	TimeoutThreshold time.Duration
	
	// ProcessingCancelled indicates whether the processing was cancelled
	ProcessingCancelled bool
	
	// ShowAPIKeyMessage indicates whether to show a message about missing API key
	ShowAPIKeyMessage bool
	
	// LastError stores the last error from the LLM processor
	LastError error
	
	// FormData stores the current form field values
	FormData FormData
	
	// UIState stores the current state of the UI
	UIState FormUIState
}

// FormData stores the current form field values
type FormData struct {
	// Title of the user story
	Title string
	
	// Description of the user story
	Description string
	
	// AsA is the user type in the user story format "As a..."
	AsA string
	
	// IWant is the capability in the user story format "I want..."
	IWant string
	
	// SoThat is the benefit in the user story format "so that..."
	SoThat string
	
	// AcceptanceCriteria is a list of acceptance criteria
	AcceptanceCriteria []string
}

// FormUIState tracks the current state of the UI
type FormUIState struct {
	// ActiveField is the currently focused field
	ActiveField string
	
	// ActiveACIndex is the index of the currently focused acceptance criteria
	ActiveACIndex int
	
	// SubmissionConfirmed indicates whether the user confirmed the submission
	SubmissionConfirmed bool
	
	// FormCancelled indicates whether the form was cancelled
	FormCancelled bool
	
	// FormWidth tracks the width of the form
	FormWidth int
	
	// FormHeight tracks the height of the form
	FormHeight int
}

// NewUserStoryFormModel creates a new user story form model
func NewUserStoryFormModel(userStory models.UserStory, processor llm.LLMProcessor, configManager *llm.ConfigManager) *UserStoryFormModel {
	// Initialize the model with default values
	model := &UserStoryFormModel{
		UserStory:          userStory,
		LLMProcessor:       processor,
		ConfigManager:      configManager,
		AutoPopulatedFields: make(map[string]bool),
		ConfidenceScores:   make(map[string]float64),
		ProcessingState:    llm.ProcessingIdle,
		TimeoutThreshold:   5 * time.Second, // 5 second default timeout threshold
		ProcessingCancelled: false,
		ShowAPIKeyMessage:  false,
		FormData: FormData{
			Title:              userStory.Title,
			Description:        userStory.Description,
			AcceptanceCriteria: userStory.Criteria,
		},
		UIState: FormUIState{
			ActiveField:        "title",
			ActiveACIndex:      0,
			SubmissionConfirmed: false,
			FormCancelled:      false,
			FormWidth:          80,
			FormHeight:         24,
		},
	}
	
	return model
}

// ProcessClipboardContent processes the clipboard content with the LLM processor
func (m *UserStoryFormModel) ProcessClipboardContent(ctx context.Context, content string) {
	// Check if the LLM processor is configured
	if !m.LLMProcessor.IsConfigured() {
		m.ProcessingState = llm.ProcessingNotConfigured
		m.ShowAPIKeyMessage = true
		return
	}
	
	// Reset processing state
	m.ProcessingState = llm.ProcessingActive
	m.ProcessingStartTime = time.Now()
	m.ProcessingCancelled = false
	m.LastError = nil
	
	// Process the content asynchronously
	go func() {
		// Create a cancellable context
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		
		// Process the unstructured text
		data, err := m.LLMProcessor.ProcessUnstructuredText(ctx, content)
		
		// Check if the processing was cancelled
		if m.ProcessingCancelled {
			m.ProcessingState = llm.ProcessingCancelled
			return
		}
		
		// Handle errors
		if err != nil {
			m.LastError = err
			m.ProcessingState = llm.ProcessingError
			return
		}
		
		// Update the form data
		m.updateFormDataFromLLM(data)
		
		// Update processing state
		m.ProcessingState = llm.ProcessingSuccess
	}()
}

// updateFormDataFromLLM updates the form data with the processed LLM data
func (m *UserStoryFormModel) updateFormDataFromLLM(data llm.UserStoryData) {
	// Store the confidence scores
	m.ConfidenceScores = data.Confidence
	
	// Populate fields based on confidence thresholds
	if score, ok := data.Confidence["title"]; ok && score > 0.5 {
		m.FormData.Title = data.Title
		m.AutoPopulatedFields["title"] = true
	}
	
	if score, ok := data.Confidence["description"]; ok && score > 0.5 {
		m.FormData.Description = data.Description
		m.AutoPopulatedFields["description"] = true
	}
	
	if score, ok := data.Confidence["as_a"]; ok && score > 0.5 {
		m.FormData.AsA = data.AsA
		m.AutoPopulatedFields["as_a"] = true
	}
	
	if score, ok := data.Confidence["i_want"]; ok && score > 0.5 {
		m.FormData.IWant = data.IWant
		m.AutoPopulatedFields["i_want"] = true
	}
	
	if score, ok := data.Confidence["so_that"]; ok && score > 0.5 {
		m.FormData.SoThat = data.SoThat
		m.AutoPopulatedFields["so_that"] = true
	}
	
	// Populate acceptance criteria
	if score, ok := data.Confidence["acceptance_criteria"]; ok && score > 0.5 {
		// Only copy as many criteria as we have fields for
		m.FormData.AcceptanceCriteria = data.AcceptanceCriteria
		m.AutoPopulatedFields["acceptance_criteria"] = true
	}
}

// CancelProcessing cancels any ongoing LLM processing
func (m *UserStoryFormModel) CancelProcessing() {
	if m.ProcessingState == llm.ProcessingActive {
		m.ProcessingCancelled = true
		m.ProcessingState = llm.ProcessingCancelled
	}
}

// GetTimeoutMessage returns an appropriate timeout message based on the processing duration
func (m *UserStoryFormModel) GetTimeoutMessage() string {
	if m.ProcessingState != llm.ProcessingActive {
		return ""
	}
	
	duration := time.Since(m.ProcessingStartTime)
	if duration >= m.TimeoutThreshold {
		return "Processing is taking longer than expected. You can press ESC to cancel."
	}
	
	return ""
}

// IsProcessingActive returns whether LLM processing is currently active
func (m *UserStoryFormModel) IsProcessingActive() bool {
	return m.ProcessingState == llm.ProcessingActive
}

// ShouldShowAPIKeyMessage returns whether to show the API key message
func (m *UserStoryFormModel) ShouldShowAPIKeyMessage() bool {
	return m.ShowAPIKeyMessage
}

// GetAPIKeyMessage returns the API key message
func (m *UserStoryFormModel) GetAPIKeyMessage() string {
	return "API key not configured. Please set your OpenAI API key in the settings to enable auto-formatting."
}

// MarkFieldEdited marks a field as manually edited by the user
func (m *UserStoryFormModel) MarkFieldEdited(fieldName string) {
	if fieldName != "" {
		delete(m.AutoPopulatedFields, fieldName)
	}
}

// MarkAllFieldsEdited marks all fields as manually edited by the user
func (m *UserStoryFormModel) MarkAllFieldsEdited() {
	m.AutoPopulatedFields = make(map[string]bool)
}

// GetFieldConfidence returns the confidence score for a field
func (m *UserStoryFormModel) GetFieldConfidence(fieldName string) float64 {
	confKey := fieldNameToConfidenceKey(fieldName)
	if score, ok := m.ConfidenceScores[confKey]; ok {
		return score
	}
	
	// Default confidence for fields not found
	return 0.0
}

// IsFieldAutoPopulated returns whether a field was auto-populated by the LLM
func (m *UserStoryFormModel) IsFieldAutoPopulated(fieldName string) bool {
	populated, ok := m.AutoPopulatedFields[fieldName]
	return ok && populated
}

// GetProcessingState returns the current processing state
func (m *UserStoryFormModel) GetProcessingState() llm.ProcessingState {
	return m.ProcessingState
}

// GetLastError returns the last error encountered during processing
func (m *UserStoryFormModel) GetLastError() string {
	if m.LastError != nil {
		return m.LastError.Error()
	}
	return ""
}

// GetFormData returns the current form data
func (m *UserStoryFormModel) GetFormData() FormData {
	return m.FormData
}

// Private helper functions

// fieldNameToConfidenceKey converts a field name to its confidence key
func fieldNameToConfidenceKey(fieldName string) string {
	if fieldName == "acceptance_criteria" || strings.HasPrefix(fieldName, "acceptance_criteria_") {
		return "acceptance_criteria"
	}
	return fieldName
} 
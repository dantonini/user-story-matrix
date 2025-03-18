package models

import (
	"fmt"
	"time"
)

// FeatureRequest represents a feature request from a user
type FeatureRequest struct {
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Importance         string    `json:"importance"`
	UserStory          string    `json:"user_story"`
	AcceptanceCriteria []string  `json:"acceptance_criteria"`
	CreatedAt          time.Time `json:"created_at"`
}

// NewFeatureRequest creates a new feature request with default values
func NewFeatureRequest() FeatureRequest {
	return FeatureRequest{
		CreatedAt: time.Now(),
	}
}

// FormatForSubmission formats the feature request for submission
func (fr *FeatureRequest) FormatForSubmission() string {
	formatted := fmt.Sprintf("*Feature Request: %s*\n\n", fr.Title)
	formatted += fmt.Sprintf("*Description:*\n%s\n\n", fr.Description)
	formatted += fmt.Sprintf("*Importance:*\n%s\n\n", fr.Importance)
	formatted += fmt.Sprintf("*User Story:*\n%s\n\n", fr.UserStory)
	
	formatted += "*Acceptance Criteria:*\n"
	for i, criteria := range fr.AcceptanceCriteria {
		formatted += fmt.Sprintf("%d. %s\n", i+1, criteria)
	}
	
	return formatted
}

// IsComplete checks if all required fields are filled
func (fr *FeatureRequest) IsComplete() bool {
	return fr.Title != "" && 
		   fr.Description != "" && 
		   fr.Importance != "" && 
		   fr.UserStory != "" && 
		   len(fr.AcceptanceCriteria) > 0
} 
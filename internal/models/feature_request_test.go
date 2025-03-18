package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFeatureRequest(t *testing.T) {
	fr := NewFeatureRequest()
	
	assert.NotZero(t, fr.CreatedAt)
	assert.Empty(t, fr.Title)
	assert.Empty(t, fr.Description)
	assert.Empty(t, fr.Importance)
	assert.Empty(t, fr.UserStory)
	assert.Empty(t, fr.AcceptanceCriteria)
}

func TestFormatForSubmission(t *testing.T) {
	fr := FeatureRequest{
		Title:       "Test Feature",
		Description: "This is a test feature",
		Importance:  "It's very important",
		UserStory:   "As a user, I want to test features, so that I can verify they work",
		AcceptanceCriteria: []string{
			"Feature should be testable",
			"Feature should work correctly",
		},
		CreatedAt: time.Now(),
	}
	
	formatted := fr.FormatForSubmission()
	
	assert.Contains(t, formatted, "*Feature Request: Test Feature*")
	assert.Contains(t, formatted, "*Description:*\nThis is a test feature")
	assert.Contains(t, formatted, "*Importance:*\nIt's very important")
	assert.Contains(t, formatted, "*User Story:*\nAs a user, I want to test features, so that I can verify they work")
	assert.Contains(t, formatted, "*Acceptance Criteria:*")
	assert.Contains(t, formatted, "1. Feature should be testable")
	assert.Contains(t, formatted, "2. Feature should work correctly")
}

func TestIsComplete(t *testing.T) {
	tests := []struct {
		name     string
		fr       FeatureRequest
		expected bool
	}{
		{
			name: "Complete feature request",
			fr: FeatureRequest{
				Title:              "Test Feature",
				Description:        "This is a test feature",
				Importance:         "It's very important",
				UserStory:          "As a user, I want to test features, so that I can verify they work",
				AcceptanceCriteria: []string{"Feature should work correctly"},
				CreatedAt:          time.Now(),
			},
			expected: true,
		},
		{
			name: "Missing title",
			fr: FeatureRequest{
				Description:        "This is a test feature",
				Importance:         "It's very important",
				UserStory:          "As a user, I want to test features, so that I can verify they work",
				AcceptanceCriteria: []string{"Feature should work correctly"},
				CreatedAt:          time.Now(),
			},
			expected: false,
		},
		{
			name: "Missing description",
			fr: FeatureRequest{
				Title:              "Test Feature",
				Importance:         "It's very important",
				UserStory:          "As a user, I want to test features, so that I can verify they work",
				AcceptanceCriteria: []string{"Feature should work correctly"},
				CreatedAt:          time.Now(),
			},
			expected: false,
		},
		{
			name: "Missing importance",
			fr: FeatureRequest{
				Title:              "Test Feature",
				Description:        "This is a test feature",
				UserStory:          "As a user, I want to test features, so that I can verify they work",
				AcceptanceCriteria: []string{"Feature should work correctly"},
				CreatedAt:          time.Now(),
			},
			expected: false,
		},
		{
			name: "Missing user story",
			fr: FeatureRequest{
				Title:              "Test Feature",
				Description:        "This is a test feature",
				Importance:         "It's very important",
				AcceptanceCriteria: []string{"Feature should work correctly"},
				CreatedAt:          time.Now(),
			},
			expected: false,
		},
		{
			name: "Missing acceptance criteria",
			fr: FeatureRequest{
				Title:       "Test Feature",
				Description: "This is a test feature",
				Importance:  "It's very important",
				UserStory:   "As a user, I want to test features, so that I can verify they work",
				CreatedAt:   time.Now(),
			},
			expected: false,
		},
		{
			name: "Empty acceptance criteria",
			fr: FeatureRequest{
				Title:              "Test Feature",
				Description:        "This is a test feature",
				Importance:         "It's very important",
				UserStory:          "As a user, I want to test features, so that I can verify they work",
				AcceptanceCriteria: []string{},
				CreatedAt:          time.Now(),
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.fr.IsComplete())
		})
	}
} 
package io

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestNewSlackClient(t *testing.T) {
	webhookURL := "https://example.com/webhook"
	client := NewSlackClient(webhookURL)
	
	assert.NotNil(t, client)
	assert.Equal(t, webhookURL, client.webhookURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestSendFeatureRequest(t *testing.T) {
	fr := models.FeatureRequest{
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
	
	// Test successful request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	client := NewSlackClient(server.URL)
	err := client.SendFeatureRequest(fr)
	
	assert.NoError(t, err)
	
	// Test unsuccessful request
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	
	client = NewSlackClient(server.URL)
	err = client.SendFeatureRequest(fr)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code 400")
} 
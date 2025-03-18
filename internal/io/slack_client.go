package io

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user-story-matrix/usm/internal/models"
)

// SlackClient handles sending messages to Slack
type SlackClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewSlackClient creates a new Slack client
func NewSlackClient(webhookURL string) *SlackClient {
	return &SlackClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SlackMessage represents a message to be sent to Slack
type SlackMessage struct {
	Text string `json:"text"`
}

// SendFeatureRequest sends a feature request to Slack
func (sc *SlackClient) SendFeatureRequest(fr models.FeatureRequest) error {
	formattedMsg := fr.FormatForSubmission()
	
	msg := SlackMessage{
		Text: formattedMsg,
	}
	
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	resp, err := sc.httpClient.Post(sc.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status code %d", resp.StatusCode)
	}
	
	return nil
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package io

import (
	"bytes"
	"context"
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
	return sc.SendFeatureRequestWithContext(context.Background(), fr)
}

// SendFeatureRequestWithContext sends a feature request to Slack with context
func (sc *SlackClient) SendFeatureRequestWithContext(ctx context.Context, fr models.FeatureRequest) error {
	formattedMsg := fr.FormatForSubmission()
	
	msg := SlackMessage{
		Text: formattedMsg,
	}
	
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sc.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status code %d", resp.StatusCode)
	}
	
	return nil
} 
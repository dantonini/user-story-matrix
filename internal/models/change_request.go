// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// UserStoryReference represents a reference to a user story in a change request
type UserStoryReference struct {
	Title       string `json:"title" yaml:"title"`
	FilePath    string `json:"file" yaml:"file"`
	ContentHash string `json:"content_hash" yaml:"content-hash"`
}

// ChangeRequest represents a change request document
type ChangeRequest struct {
	Name        string              `json:"name" yaml:"name"`
	CreatedAt   time.Time           `json:"created_at" yaml:"created-at"`
	UserStories []UserStoryReference `json:"user_stories" yaml:"user-stories"`
	FilePath    string              `json:"file_path" yaml:"-"`
}

// GenerateChangeRequestTemplate generates a template for a new change request
func GenerateChangeRequestTemplate(name string, userStories []UserStoryReference) string {
	template := `---
name: {{name}}
created-at: {{created_at}}
user-stories:
{{user_stories}}
---

# Blueprint

## Overview

This is a change request for implementing the following user stories:
{{user_story_titles}}

Please provide a detailed implementation plan.
`
	
	// Fill in the name
	template = strings.ReplaceAll(template, "{{name}}", name)
	
	// Fill in the creation date
	now := time.Now().Format(time.RFC3339)
	template = strings.ReplaceAll(template, "{{created_at}}", now)
	
	// Fill in user stories
	var userStoriesBuilder strings.Builder
	var userStoryTitlesBuilder strings.Builder
	
	for i, us := range userStories {
		userStoriesBuilder.WriteString(fmt.Sprintf("  - title: %s\n", us.Title))
		userStoriesBuilder.WriteString(fmt.Sprintf("    file: %s\n", us.FilePath))
		userStoriesBuilder.WriteString(fmt.Sprintf("    content-hash: %s\n", us.ContentHash))
		
		userStoryTitlesBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, us.Title))
	}
	
	template = strings.ReplaceAll(template, "{{user_stories}}", userStoriesBuilder.String())
	template = strings.ReplaceAll(template, "{{user_story_titles}}", userStoryTitlesBuilder.String())
	
	return template
}

// GenerateChangeRequestFilename generates a filename for a change request
func GenerateChangeRequestFilename(name string) string {
	// Format: yyyy-mm-dd-HHMMSS-<change-request-name>.blueprint.md
	now := time.Now()
	date := now.Format("2006-01-02")
	timeStr := now.Format("150405")
	
	slug := SlugifyTitle(name)
	
	return fmt.Sprintf("%s-%s-%s.blueprint.md", date, timeStr, slug)
}

// LoadChangeRequestFromContent loads a change request from content
func LoadChangeRequestFromContent(filePath string, content []byte) (ChangeRequest, error) {
	cr := ChangeRequest{
		FilePath: filePath,
	}
	
	contentStr := string(content)
	
	// Extract metadata
	metadata, err := ExtractMetadataFromContent(contentStr)
	if err != nil {
		return cr, err
	}
	
	// Get name
	if name, ok := metadata["name"]; ok {
		cr.Name = name
	}
	
	// Parse creation date
	if createdAt, ok := metadata["created-at"]; ok {
		t, err := time.Parse(time.RFC3339, createdAt)
		if err == nil {
			cr.CreatedAt = t
		}
	}
	
	// Parse user stories - this is more complex and would need YAML parsing
	// For simplicity, we'll use a regex approach for now
	userStoriesRegex := regexp.MustCompile(`(?m)^  - title: (.*)$\n^    file: (.*)$\n^    content-hash: (.*)$`)
	matches := userStoriesRegex.FindAllStringSubmatch(contentStr, -1)
	
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		
		cr.UserStories = append(cr.UserStories, UserStoryReference{
			Title:       match[1],
			FilePath:    match[2],
			ContentHash: match[3],
		})
	}
	
	return cr, nil
}


// GetNextStepsInstruction generates next steps instruction for the change request
func GetNextStepsInstruction(changeRequestPath string) string {
	return fmt.Sprintf("To continue, run: ./usm code %s", changeRequestPath)
} 
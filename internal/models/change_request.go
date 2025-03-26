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

// GetPromptInstruction generates a prompt instruction for the change request
func GetPromptInstruction(changeRequestPath string, userStoryCount int) string {
	return fmt.Sprintf(
		`Your task is to produce a blueprint file for the change request.

A blueprint is a technical design document that outlines proposed codebase changes without actual implementation. It helps:
- Understand the proposed changes before coding
- Create a clear roadmap for upcoming development tasks

General Guidelines:
- The blueprint has a metadata section referencing a set of user stories. Each user story has a title and a filename. Read all the user stories at once using the command ./cat-user-stories-in-change-request.sh <change_request_path>.
- The document is not for writing code but for transmitting ideas, concepts, and plans.
- Follow a top-down (or break-down) approach: start with a high-level overview and progressively drill down into specifics.

# Overview
**Purpose:**  
Provide a brief summary that captures the essence of all user stories.  
- Highlight common themes and relationships among the user stories.
- Summarize overall objectives without detailing individual acceptance criteria.

## Foudamentals
**Purpose:**  
Outline the key technical concepts necessary to address the user stories:
- **Data Structures:** Define any high-level data structures, including their purposes.
- **Algorithms:** Describe key algorithms using pseudo-code, outlining their intended functionality.
- **Refactoring Strategy:** Summarize any broad refactoring plans for the existing codebase.

# How to verify – Detailed User Story Breakdown
**Purpose:**  
For each user story, detail how the changes will be verified:
- **Acceptance Criteria:** Break down each user story into its individual acceptance criteria.
- **Testing Scenarios:** For each criterion, provide clear, concise testing scenarios that are tangible and automatable.
- **Bottom-Up Detailing:** Start with basic criteria and work toward more complex conditions.

# What is the Plan – Detailed Action Items
**Purpose:**  
For each user story, outline a detailed plan for what needs to be done. Take into account the user story verification process described earlier so to make the verification process easy to implement.
- **Task Breakdown:** Describe each implementation step without writing actual code.
- **Specific Data Structures:** List any data structures that need to be defined or modified, along with their purposes.
- **Specific Algorithms:** Provide pseudo-code for any specific algorithms, explaining their function.
- **Targeted Refactoring:** Detail any precise refactoring steps required for the existing codebase.
- **Validation:** Ensure the plan is validated against the current codebase, ensuring feasibility and completeness.

**Note:**  
Remember, the blueprint should be a planning and communication tool. Do not include any actual code – only high-level pseudo-code and detailed action items that make the subsequent verification and development process straightforward.

validate them against the codebase, and define a detailed plan for the change. Don't do any implementation, just describe what needs to be done. You can describe data structures, algorithm in pseudo code, refactoring steps, etc. 
Store the plan in the change request file %s in markdown format in a section called \"Blueprint\". Ensure to include the steps required to satisfy the acceptance criteria of all mentioned user stories.`,
		changeRequestPath,
	)
} 
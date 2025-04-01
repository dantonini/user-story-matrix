// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package implementation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestIsUserStoryImplemented(t *testing.T) {
	// Create mock filesystem
	mockFS := io.NewMockFileSystem()
	
	// Create test user story
	userStoryPath := "docs/user-stories/01-test-user-story.md"
	err := mockFS.WriteFile(userStoryPath, []byte(`---
file_path: docs/user-stories/01-test-user-story.md
created_at: 2025-01-01T00:00:00Z
last_updated: 2025-01-01T00:00:00Z
_content_hash: abcdef123456
---

# Test User Story

As a tester,
I want to test if a user story is implemented,
so that I can verify the implementation status feature.

## Acceptance criteria

- Check if user story is marked as implemented when referenced in an implemented change request
`), 0644)
	assert.NoError(t, err)
	
	// Create test change request blueprint
	changeRequestPath := "docs/changes-request/2025-01-01-000000-test-change-request.blueprint.md"
	err = mockFS.WriteFile(changeRequestPath, []byte(`---
name: test-change-request
created-at: 2025-01-01T00:00:00Z
user-stories:
  - title: Test User Story
    file: docs/user-stories/01-test-user-story.md
    content-hash: abcdef123456
---

# Blueprint

## Overview

This is a change request for implementing the following user stories:
1. Test User Story

Please provide a detailed implementation plan.
`), 0644)
	assert.NoError(t, err)
	
	// Create test directory structure
	err = mockFS.MkdirAll("docs/user-stories", 0755)
	assert.NoError(t, err)
	err = mockFS.MkdirAll("docs/changes-request", 0755)
	assert.NoError(t, err)
	
	// Create test user story model
	userStory := models.UserStory{
		Title:        "Test User Story",
		FilePath:     userStoryPath,
		ContentHash:  "abcdef123456",
		IsImplemented: false,
	}
	
	// Test 1: User story should not be implemented yet
	isImplemented, err := IsUserStoryImplemented(userStory, mockFS)
	assert.NoError(t, err)
	assert.False(t, isImplemented, "User story should not be marked as implemented before change request implementation exists")
	
	// Test 2: Add implementation file and check again, user story should be implemented
	implementationPath := "docs/changes-request/2025-01-01-000000-test-change-request.implementation.md"
	err = mockFS.WriteFile(implementationPath, []byte(`# Implementation

This is the implementation for the user story.
`), 0644)
	assert.NoError(t, err)
	
	isImplemented, err = IsUserStoryImplemented(userStory, mockFS)
	assert.NoError(t, err)
	assert.True(t, isImplemented, "User story should be marked as implemented when referenced in an implemented change request")
	
	// Test 3: Test UpdateImplementationStatus function
	err = UpdateImplementationStatus(&userStory, mockFS)
	assert.NoError(t, err)
	assert.True(t, userStory.IsImplemented, "UpdateImplementationStatus should set IsImplemented flag to true")
} 
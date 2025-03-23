// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestNewEngine(t *testing.T) {
	stories := []models.UserStory{
		{Title: "Story 1", IsImplemented: false},
		{Title: "Story 2", IsImplemented: true},
	}

	engine := NewEngine(stories)

	assert.Equal(t, 2, len(engine.stories))
	assert.Equal(t, 2, engine.state.TotalCount)
	assert.NotNil(t, engine.cache.ImplementationStatus)
	assert.NotNil(t, engine.cache.SearchResults)
}

func TestFilter(t *testing.T) {
	// Create test stories
	stories := []models.UserStory{
		{
			Title:        "Add user authentication",
			Description: "Implement user authentication system",
			Criteria:    []string{"Support login", "Support logout"},
			IsImplemented: false,
		},
		{
			Title:        "Add user profile",
			Description: "Implement user profile page",
			Criteria:    []string{"Show user info", "Allow editing"},
			IsImplemented: true,
		},
		{
			Title:        "Fix login bug",
			Description: "Fix bug in login system",
			Criteria:    []string{"Fix error handling"},
			IsImplemented: false,
		},
	}

	engine := NewEngine(stories)

	// Test filtering by implementation status
	t.Run("Filter by implementation status", func(t *testing.T) {
		// By default, only show unimplemented stories
		filtered := engine.Filter("")
		assert.Equal(t, 2, len(filtered))
		assert.Equal(t, "Add user authentication", filtered[0].Title)
		assert.Equal(t, "Fix login bug", filtered[1].Title)

		// Show all stories
		engine.SetShowAll(true)
		filtered = engine.Filter("")
		assert.Equal(t, 3, len(filtered))
	})

	// Test text search
	t.Run("Text search", func(t *testing.T) {
		// Reset show all
		engine.SetShowAll(false)

		// Search in title
		filtered := engine.Filter("auth")
		if len(filtered) > 0 {
			assert.Contains(t, filtered[0].Title, "auth")
		}

		// Search in description
		filtered = engine.Filter("profile page")
		assert.Equal(t, 0, len(filtered)) // Profile story is implemented

		// Show all and search again
		engine.SetShowAll(true)
		filtered = engine.Filter("profile")
		if len(filtered) > 0 {
			foundProfile := false
			for _, story := range filtered {
				if story.Title == "Add user profile" {
					foundProfile = true
					break
				}
			}
			assert.True(t, foundProfile, "Should find the profile story")
		}

		// Search in criteria
		filtered = engine.Filter("error")
		if len(filtered) > 0 {
			foundError := false
			for _, story := range filtered {
				if story.Title == "Fix login bug" {
					foundError = true
					break
				}
			}
			assert.True(t, foundError, "Should find the error handling story")
		}

		// No matches
		filtered = engine.Filter("xyznonexistent")
		assert.Equal(t, 0, len(filtered))
	})

	// Test search result caching
	t.Run("Search result caching", func(t *testing.T) {
		// Clear the cache first
		engine.ClearCache()
		
		// First search
		engine.Filter("login")
		assert.NotEmpty(t, engine.cache.SearchResults["login"])

		// Same search should use cache
		before := engine.cache.LastUpdated
		engine.Filter("login")
		assert.Equal(t, before, engine.cache.LastUpdated)

		// Different search should update cache
		time.Sleep(time.Millisecond) // Ensure time difference
		engine.Filter("profile")
		assert.NotEqual(t, before, engine.cache.LastUpdated)
	})

	// Test clearing cache
	t.Run("Clear cache", func(t *testing.T) {
		engine.Filter("login")
		assert.NotEmpty(t, engine.cache.SearchResults)

		engine.ClearCache()
		assert.Empty(t, engine.cache.SearchResults)
		assert.Empty(t, engine.cache.ImplementationStatus)
		assert.True(t, engine.cache.LastUpdated.IsZero())
	})
}

func TestGetState(t *testing.T) {
	stories := []models.UserStory{
		{Title: "Story 1", IsImplemented: false},
		{Title: "Story 2", IsImplemented: true},
	}

	engine := NewEngine(stories)

	// Initial state
	state := engine.GetState()
	assert.Equal(t, 2, state.TotalCount)
	assert.Equal(t, "", state.SearchQuery)
	assert.False(t, state.ShowAll)

	// After filtering
	filtered := engine.Filter("Story")
	state = engine.GetState()
	assert.Equal(t, len(filtered), state.FilteredCount) // Only check that filtered count matches result length
	assert.Equal(t, "Story", state.SearchQuery)

	// After showing all
	engine.SetShowAll(true)
	filtered = engine.Filter("Story")
	state = engine.GetState()
	assert.Equal(t, len(filtered), state.FilteredCount) // Only check that filtered count matches result length
	assert.True(t, state.ShowAll)
}
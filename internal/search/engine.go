// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package search

import (
	"strings"
	"sync"
	"time"

	"github.com/sahilm/fuzzy"
	"github.com/user-story-matrix/usm/internal/models"
)

// FilterState represents the current state of filtering
type FilterState struct {
	SearchQuery    string
	ShowAll       bool
	FilteredCount int
	TotalCount    int
}

// SearchCache represents the cache for search results
type SearchCache struct {
	ImplementationStatus map[string]bool    // Cache of story implementation status
	SearchResults       map[string][]int    // Cache of search results
	LastUpdated        time.Time           // When the cache was last updated
	sync.RWMutex                          // For thread-safe access
}

// Engine represents the search engine for filtering user stories
type Engine struct {
	stories []models.UserStory
	state   FilterState
	cache   SearchCache
	mu      sync.RWMutex
}

// NewEngine creates a new search engine instance
func NewEngine(stories []models.UserStory) *Engine {
	return &Engine{
		stories: stories,
		cache: SearchCache{
			ImplementationStatus: make(map[string]bool),
			SearchResults:       make(map[string][]int),
		},
		state: FilterState{
			TotalCount: len(stories),
		},
	}
}

// SetShowAll updates the show all flag
func (e *Engine) SetShowAll(showAll bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state.ShowAll = showAll
}

// Filter applies the current filters and returns matching stories
func (e *Engine) Filter(query string) []models.UserStory {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Update search query
	e.state.SearchQuery = query

	// Start with all stories
	filtered := make([]models.UserStory, 0, len(e.stories))

	// First filter by implementation status
	for _, story := range e.stories {
		if !e.state.ShowAll && story.IsImplemented {
			continue
		}
		filtered = append(filtered, story)
	}

	// If no search query, return all stories that match implementation status
	if query == "" {
		e.state.FilteredCount = len(filtered)
		return filtered
	}

	// Check cache for search results
	if results, ok := e.cache.SearchResults[query]; ok {
		// Return cached results
		matchedStories := make([]models.UserStory, 0, len(results))
		for _, idx := range results {
			if idx < len(filtered) {
				matchedStories = append(matchedStories, filtered[idx])
			}
		}
		e.state.FilteredCount = len(matchedStories)
		return matchedStories
	}

	// Prepare data for fuzzy search
	searchStrings := make([]string, 0, len(filtered))
	for _, story := range filtered {
		// Combine searchable fields with weights
		searchStr := strings.Join([]string{
			story.Title,                    // Highest weight
			story.Description,              // Medium weight
			strings.Join(story.Criteria, " "), // Lower weight
		}, " ")
		searchStrings = append(searchStrings, searchStr)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, searchStrings)

	// Sort stories by match score and update scores
	result := make([]models.UserStory, 0, len(matches))
	matchIndices := make([]int, 0, len(matches))
	for _, match := range matches {
		story := filtered[match.Index]
		story.MatchScore = float64(match.Score) / 100.0
		result = append(result, story)
		matchIndices = append(matchIndices, match.Index)
	}

	// Cache the results
	e.cache.Lock()
	e.cache.SearchResults[query] = matchIndices
	e.cache.LastUpdated = time.Now()
	e.cache.Unlock()

	e.state.FilteredCount = len(result)
	return result
}

// GetState returns the current filter state
func (e *Engine) GetState() FilterState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

// ClearCache clears the search cache
func (e *Engine) ClearCache() {
	e.cache.Lock()
	defer e.cache.Unlock()
	e.cache.SearchResults = make(map[string][]int)
	e.cache.ImplementationStatus = make(map[string]bool)
	e.cache.LastUpdated = time.Time{}
}

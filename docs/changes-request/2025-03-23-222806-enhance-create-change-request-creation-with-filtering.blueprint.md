---
name: enhance create change request creation with filtering
created-at: 2025-03-23T22:28:06+01:00
user-stories:
  - title: Enhanced Change Request Creation with Filtering
    file: docs/user-stories/basic-commands/03A-create-change-request-filtering.md
    content-hash: ff77a8252e9c3b809422df613e6daa60bcccf29b4380ea76e73cd115219fce70

---

# Blueprint

## Overview

This blueprint outlines the implementation plan for enhancing the change request creation functionality with advanced filtering and search capabilities. The goal is to improve the user experience when selecting user stories by providing real-time filtering, implementation status indicators, and better visibility into the selection process.

## Implementation Plan

### Phase 1: Data Model and State Management

1. **Extend UserStory Model**
   ```go
   type UserStory struct {
       // ... existing fields ...
       IsImplemented bool      // Whether the story has been implemented
       MatchScore   float64   // Search relevance score (for filtering)
   }
   ```

2. **Create Filter State Structure**
   ```go
   type FilterState struct {
       SearchQuery     string
       ShowAll        bool
       FilteredCount  int
       TotalCount    int
   }
   ```

### Phase 2: Implementation Status Detection

1. **Create Implementation Status Detector**
   - Create new package `internal/userstory/status`
   - Implement logic to determine if a user story is implemented:
     ```go
     func IsImplemented(story models.UserStory) bool {
         // Check if story is referenced in any implementation files
         // Return true if found in at least one implementation file
     }
     ```
   - Add caching mechanism to avoid re-checking files unnecessarily

2. **Update User Story Loading**
   - Modify the existing user story loading logic to include implementation status
   - Add implementation status to the story metadata when loading
   - Cache results to improve performance

### Phase 3: Search and Filter Implementation

1. **Create Search Engine Package**
   ```go
   package search

   type Engine struct {
       stories []models.UserStory
       state   FilterState
   }

   func (e *Engine) Filter(query string) []models.UserStory {
       // Apply filters:
       // 1. Implementation status (if !showAll)
       // 2. Text search across title/description/criteria
       // 3. Update match scores
       // 4. Sort by relevance
   }
   ```

2. **Implement Text Search Algorithm**
   - Use fuzzy matching for partial word matches
   - Search across multiple fields:
     - Title (highest weight)
     - Description (medium weight)
     - Acceptance criteria (lower weight)
   - Calculate relevance scores
   - Support case-insensitive matching

### Phase 4: UI Components Enhancement

1. **Create Enhanced Selection UI**
   ```go
   type SelectionUI struct {
       searchBox    *textinput.Model
       storyList    *list.Model
       statusBar    *statusbar.Model
       filterState  FilterState
   }
   ```

2. **Implement Real-time Search**
   - Add search box above story list
   - Update filtered results as user types
   - Debounce search updates (prevent too frequent updates)
   - Highlight matching text in results

3. **Status Bar Component**
   - Show total/filtered story counts
   - Display current filter state
   - Show keyboard shortcuts

4. **Visual Enhancements**
   - Add color coding for implementation status
   - Show search match score indicators
   - Improve spacing and layout
   - Add keyboard shortcut hints

### Phase 5: Command Integration

1. **Update Command Flags**
   ```go
   var (
       showAll bool
       fromDir string
   )

   func init() {
       createChangeRequestCmd.Flags().BoolVar(&showAll, "show-all", false, "Show all user stories, including implemented ones")
       // ... existing flags ...
   }
   ```

2. **Modify Command Flow**
   ```go
   func createChangeRequest() {
       // 1. Load user stories with implementation status
       // 2. Initialize search engine
       // 3. Create and run enhanced selection UI
       // 4. Process selected stories
       // 5. Create change request as before
   }
   ```

### Phase 6: Testing Strategy

1. **Unit Tests**
   - Test implementation status detection
   - Test search and filter algorithms
   - Test UI component behavior
   - Test command integration

2. **Integration Tests**
   - Test end-to-end flow with various filter combinations
   - Test performance with large sets of stories
   - Test edge cases in filtering and selection

3. **Test Data**
   - Create comprehensive test fixtures
   - Include various story states and content types
   - Test with real-world-like data

## Technical Details

### Dependencies
- `github.com/charmbracelet/bubbles` - For enhanced TUI components
- `github.com/sahilm/fuzzy` - For fuzzy text matching
- Existing internal packages

### Performance Considerations
- Cache implementation status results
- Debounce search updates
- Optimize search algorithm for large sets
- Lazy load story content

### Error Handling
- Graceful degradation of search features
- Clear error messages for status detection issues
- Recovery mechanisms for UI state

### Domain Model Documentation Updates

The search engine represents a significant addition to the project's domain model. The following updates will be made to `docs/domain-model/domain-model.md`:

1. **Add Search Engine Section**
   ```markdown
   ### SearchEngine

   The `SearchEngine` is responsible for filtering and ranking user stories based on search criteria and implementation status.

   ```go
   type SearchEngine struct {
       Stories []UserStory    // Stories to search through
       State   FilterState    // Current filter state
       Cache   SearchCache    // Cache for search results
   }

   type FilterState struct {
       SearchQuery    string   // Current search term
       ShowAll       bool     // Whether to show implemented stories
       FilteredCount int      // Number of stories after filtering
       TotalCount   int      // Total number of stories
   }

   type SearchCache struct {
       ImplementationStatus map[string]bool    // Cache of story implementation status
       SearchResults       map[string][]int    // Cache of search results
       LastUpdated        time.Time           // When the cache was last updated
   }
   ```

   The SearchEngine provides:
   - Real-time filtering of user stories
   - Implementation status tracking
   - Relevance scoring
   - Result caching for performance
   ```

2. **Update Core Entities**
   ```markdown
   ### UserStory

   The `UserStory` model is extended with search-related fields:

   ```go
   type UserStory struct {
       // ... existing fields ...
       IsImplemented bool      // Whether the story has been implemented
       MatchScore   float64   // Search relevance score
   }
   ```
   ```

3. **Add Relationships Section**
   ```markdown
   ## Search-Related Relationships

   ```
   ┌────────────┐      filters       ┌────────────────┐
   │            │ 0..*          1    │                │
   │ UserStory  ├─────────────────►  │ SearchEngine   │
   │            │                    │                │
   └────────────┘                    └────────────────┘
   ```

   - SearchEngine maintains a collection of UserStories
   - SearchEngine tracks implementation status
   - SearchEngine calculates and assigns match scores
   ```

4. **Add Search Behavior Documentation**
   ```markdown
   ## Search Behavior

   The search functionality follows these principles:

   1. **Relevance Scoring**
      - Title matches: highest weight (1.0)
      - Description matches: medium weight (0.7)
      - Acceptance criteria matches: lower weight (0.5)

   2. **Implementation Status**
      - By default, only unimplemented stories are shown
      - `--show-all` flag reveals all stories
      - Implementation status is cached for performance

   3. **Search Algorithm**
      - Case-insensitive matching
      - Partial word matching using fuzzy search
      - Real-time updates as user types
      - Results sorted by relevance score

   4. **Performance Optimizations**
      - Implementation status caching
      - Search result caching
      - Debounced updates
      - Lazy content loading
   ```

## Migration Plan

1. **Implement Changes Incrementally**
   - Add implementation status detection first
   - Integrate basic filtering
   - Enhance UI components
   - Add real-time search
   - Polish and optimize

2. **Backward Compatibility**
   - Maintain support for existing command flags
   - Preserve current change request format
   - Keep existing selection UI as fallback

## Acceptance Criteria Validation

✓ Implementation Status Filter
  - Default shows only unimplemented stories
  - `--show-all` flag displays all stories
  - Clear visual indicators for implementation status

✓ Search and Filter Capabilities
  - Real-time filtering as user types
  - Case-insensitive matching
  - Partial word matches
  - Search across all story content

✓ User Interface
  - Shows story counts and filter status
  - Multi-select capability preserved
  - Clear visual feedback
  - "No results" handling

✓ Integration
  - Works with existing `--from` directory option
  - Maintains change request metadata format
  - Seamless integration with existing workflow

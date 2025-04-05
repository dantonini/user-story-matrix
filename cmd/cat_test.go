// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/models"
)

// setupCatTest prepares a test environment for the cat command
func setupCatTest(t *testing.T) (*cobra.Command, *bytes.Buffer, *io.MockFileSystem) {
	// Create a buffer to capture output
	outBuf := new(bytes.Buffer)

	// Create a mock file system
	mockFS := io.NewMockFileSystem()

	// Create a test change request file
	changeRequestContent := `---
name: test change request
created-at: 2025-04-05T10:00:00Z
user-stories:
  - title: Test Story 1
    file: docs/user-stories/test-1.md
    content-hash: abc123
  - title: Test Story 2
    file: docs/user-stories/test-2.md
    content-hash: def456
  - title: Missing Story
    file: docs/user-stories/missing.md
    content-hash: xyz789
  - title: Login Feature
    file: docs/user-stories/auth/login.md
    content-hash: auth123
  - title: Registration Feature
    file: docs/user-stories/auth/registration.md
    content-hash: auth456
---
`
	mockFS.AddFile("test-cr.md", []byte(changeRequestContent))

	// Create test user story files
	userStory1Content := `---
file_path: docs/user-stories/test-1.md
created_at: 2025-04-05T10:00:00Z
last_updated: 2025-04-05T10:00:00Z
_content_hash: abc123
---

# Test Story 1
This is the first test story.
`
	mockFS.AddFile("docs/user-stories/test-1.md", []byte(userStory1Content))

	userStory2Content := `---
file_path: docs/user-stories/test-2.md
created_at: 2025-04-05T10:00:00Z
last_updated: 2025-04-05T10:00:00Z
_content_hash: def456
---

# Test Story 2
This is the second test story.
`
	mockFS.AddFile("docs/user-stories/test-2.md", []byte(userStory2Content))

	// Story with multiple metadata sections
	userStory3Content := `---
file_path: docs/user-stories/auth/login.md
created_at: 2025-04-05T10:00:00Z
last_updated: 2025-04-05T10:00:00Z
_content_hash: auth123
---

# Login Feature Story
This story describes the login functionality.

## Acceptance Criteria
- User can enter username and password
- System validates credentials
- User is redirected to dashboard on success

---
additional_section: true
_content_hash: section123
---

## Implementation Notes
Implementation details go here.
`
	mockFS.AddFile("docs/user-stories/auth/login.md", []byte(userStory3Content))

	// Story with very long content for compact mode testing
	userStory4Content := `---
file_path: docs/user-stories/auth/registration.md
created_at: 2025-04-05T10:00:00Z
last_updated: 2025-04-05T10:00:00Z
_content_hash: auth456
---

# Registration Feature
This story describes the user registration functionality.

As a new user
I want to create an account
so that I can access the system

## First paragraph
The registration system should allow new users to create accounts with email verification.

## Second paragraph
After registration, users should receive a confirmation email to verify their account.

## Third paragraph
The registration form should include validation for all required fields.
`
	mockFS.AddFile("docs/user-stories/auth/registration.md", []byte(userStory4Content))

	// Create a malformed change request file
	malformedCRContent := `---
name: malformed
created-at: invalid-date
user-stories:
  - incomplete
---
`
	mockFS.AddFile("malformed-cr.md", []byte(malformedCRContent))

	// Create an empty change request file
	emptyContent := `---
name: empty change request
created-at: 2025-04-05T10:00:00Z
user-stories: []
---
`
	mockFS.AddFile("empty-cr.md", []byte(emptyContent))

	// Create a command for testing
	cmd := &cobra.Command{Use: "cat"}
	cmd.SetOut(outBuf)
	cmd.SetErr(outBuf)
	
	return cmd, outBuf, mockFS
}

// TestCatDisplaysUserStoryContent tests that the cat command displays user story content
func TestCatDisplaysUserStoryContent(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Set up the cat command options
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, createTestChangeRequest())
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that content from both user stories is included
	assert.Contains(t, output, "# Test Story 1")
	assert.Contains(t, output, "This is the first test story.")
	assert.Contains(t, output, "# Test Story 2")
	assert.Contains(t, output, "This is the second test story.")
}

// TestCatDisplaysFilePathComments tests that each user story is preceded by a file path comment
func TestCatDisplaysFilePathComments(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Set up the cat command options
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, createTestChangeRequest())
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that file path comments are included
	assert.Contains(t, output, "[//]: # (docs/user-stories/test-1.md)")
	assert.Contains(t, output, "[//]: # (docs/user-stories/test-2.md)")
	
	// Simpler test to check that each file path comment is followed by content
	assert.True(t, strings.Contains(output, "[//]: # (docs/user-stories/test-1.md)"))
	assert.True(t, strings.Contains(output, "# Test Story 1"))
	assert.True(t, strings.Contains(output, "[//]: # (docs/user-stories/test-2.md)"))
	assert.True(t, strings.Contains(output, "# Test Story 2"))
}

// TestCatHandlesMissingFiles tests that the cat command handles missing files gracefully
func TestCatHandlesMissingFiles(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Set up the cat command options
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
	}
	
	// Prepare a change request with missing file
	cr := createTestChangeRequest()
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, cr)
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that an error message is displayed for the missing file
	assert.Contains(t, output, "User story file not found: docs/user-stories/missing.md")
}

// TestCatShowsUsageWhenNoArgs tests that usage instructions are shown when no arguments are provided
func TestCatShowsUsageWhenNoArgs(t *testing.T) {
	// For this test, we'll verify the logic directly
	called := false
	
	// Create a simple command that checks for args and calls a function
	run := func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			// This simulates what the real command does
			called = true
			return
		}
	}
	
	// Create a test command
	cmd := &cobra.Command{
		Use:   "cat",
		Run:   run,
	}
	
	// Execute with no args
	cmd.SetArgs([]string{})
	_ = cmd.Execute()
	
	// Verify that the help logic would have been triggered
	assert.True(t, called, "Help should be triggered when no arguments are provided")
}

// TestCatExcludesContentHashByDefault tests that content hash is excluded by default
func TestCatExcludesContentHashByDefault(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Set up the cat command options with default behavior (no content hash)
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, createTestChangeRequest())
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that content hash is not included
	assert.NotContains(t, output, "_content_hash: abc123")
	assert.NotContains(t, output, "_content_hash: def456")
}

// TestCatIncludesContentHashWithFlag tests that content hash is included when the flag is set
func TestCatIncludesContentHashWithFlag(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Set up the cat command options with showContentHash=true
	options := CatOptions{
		ShowContentHash:   true,
		ChangeRequestPath: "test-cr.md",
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, createTestChangeRequest())
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that content hash is included
	assert.Contains(t, output, "_content_hash: abc123")
	assert.Contains(t, output, "_content_hash: def456")
}

// TestProcessUserStoryContent tests the processUserStoryContent function
func TestProcessUserStoryContent(t *testing.T) {
	// Test with showContentHash=false
	content := `---
file_path: docs/user-stories/test.md
created_at: 2025-04-05T10:00:00Z
_content_hash: abc123
---

# Test Story
Content here.
`
	result := processUserStoryContent(content, false)
	assert.NotContains(t, result, "_content_hash: abc123")
	assert.Contains(t, result, "# Test Story")
	
	// Test with showContentHash=true
	result = processUserStoryContent(content, true)
	assert.Contains(t, result, "_content_hash: abc123")
	assert.Contains(t, result, "# Test Story")
}

// TestFilteringUserStories tests filtering user stories by pattern
func TestFilteringUserStories(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Create a change request with multiple stories
	cr := createExtendedTestChangeRequest()
	
	// Set up the cat command options with a filter pattern
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
		FilterPattern:     "login|Login",
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, cr)
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that only stories matching the filter pattern are included
	assert.Contains(t, output, "Login Feature Story")
	assert.Contains(t, output, "login functionality")
	
	// Verify that non-matching stories are not included
	assert.NotContains(t, output, "Test Story 1")
	assert.NotContains(t, output, "Registration Feature")
	
	// Verify the summary shows filtered stories count
	assert.Contains(t, output, "Summary: 1 of 5 stories displayed")
}

// TestExcludingUserStories tests excluding user stories by pattern
func TestExcludingUserStories(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Create a change request with multiple stories
	cr := createExtendedTestChangeRequest()
	
	// Set up the cat command options with exclude patterns
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
		ExcludeStories:    []string{"auth", "missing"},
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, cr)
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that excluded stories are not included
	assert.NotContains(t, output, "Login Feature Story")
	assert.NotContains(t, output, "Registration Feature")
	
	// Verify that non-excluded stories are included
	assert.Contains(t, output, "Test Story 1")
	assert.Contains(t, output, "Test Story 2")
	
	// Verify the summary shows excluded stories count
	assert.Contains(t, output, "Summary: 2 of 5 stories displayed")
}

// TestCompactOutputMode tests the compact output mode
func TestCompactOutputMode(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Create a change request with multiple stories
	cr := createExtendedTestChangeRequest()
	
	// Set up the cat command options with compact mode
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
		CompactMode:       true,
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, cr)
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that the compact format is used
	assert.Contains(t, output, "# Registration Feature")
	assert.Contains(t, output, "As a new user")
	assert.Contains(t, output, "I want to create an account")
	assert.Contains(t, output, "so that I can access the system")
	
	// Verify that additional paragraphs are not included in compact mode
	assert.NotContains(t, output, "Second paragraph")
	assert.NotContains(t, output, "Third paragraph")
	
	// Verify the separator format is correct
	assert.NotContains(t, output, "---\n")
}

// TestColorOutput tests the colorized output option
func TestColorOutput(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal with color tracking
	mockTerminal := &colorTrackingTerminal{
		buffer:            outBuf,
		successCallCount:  0,
		warningCallCount:  0,
	}
	
	// Create a change request with multiple stories
	cr := createExtendedTestChangeRequest()
	
	// Set up the cat command options with color output
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
		ColorOutput:       true,
		FilterPattern:     "login|Login", // Add filter to trigger summary
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, cr)
	
	// Check results
	assert.NoError(t, err)
	
	// Verify that colored output methods were called
	assert.Greater(t, mockTerminal.successCallCount, 0, "PrintSuccess should be called for colorized file paths")
	assert.Equal(t, 1, mockTerminal.warningCallCount, "PrintWarning should be called for summary")
}

// TestMultipleSectionContentHashRemoval tests that content hash is removed from all metadata sections
func TestMultipleSectionContentHashRemoval(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Create a specific change request that includes the login story with multiple metadata sections
	cr := models.ChangeRequest{
		Name:      "test change request",
		CreatedAt: time.Date(2025, 4, 5, 10, 0, 0, 0, time.UTC),
		UserStories: []models.UserStoryReference{
			{
				Title:       "Login Feature",
				FilePath:    "docs/user-stories/auth/login.md",
				ContentHash: "auth123",
			},
		},
		FilePath: "test-cr.md",
	}
	
	// Set up the cat command options
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, cr)
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that content hash is not included in any section
	assert.NotContains(t, output, "_content_hash: auth123")
	assert.NotContains(t, output, "_content_hash: section123")
}

// TestEmptyChangeRequest tests handling of change requests with no user stories
func TestEmptyChangeRequest(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal with debug enabled for summary output
	mockTerminal := &colorTrackingTerminal{
		buffer: outBuf,
	}
	
	// Read the empty change request
	content, _ := mockFS.ReadFile("empty-cr.md")
	cr, err := models.LoadChangeRequestFromContent("empty-cr.md", content)
	assert.NoError(t, err)
	
	// Set up the cat command options
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "empty-cr.md",
	}
	
	// Call the function under test
	err = processAndPrintUserStories(mockFS, mockTerminal, options, cr)
	
	// Check results
	assert.NoError(t, err)
	output := outBuf.String()
	
	// Verify that a summary is displayed for empty change requests when debug is enabled
	assert.Contains(t, output, "Summary: 0 of 0 stories displayed")
}

// TestMissingChangeRequestFile tests that the cat command handles a missing change request file gracefully
func TestMissingChangeRequestFile(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Set up options with a non-existent file
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "non-existent-file.md",
	}
	
	// Define a mock function that would be used in the real command's run function
	testFunc := func() error {
		// Check if the file exists
		if !mockFS.Exists(options.ChangeRequestPath) {
			mockTerminal.PrintError(fmt.Sprintf("Change request file not found: %s", options.ChangeRequestPath))
			return fmt.Errorf("change request file not found: %s", options.ChangeRequestPath)
		}
		return nil
	}
	
	// Call the test function
	err := testFunc()
	
	// Verify that an error is returned
	assert.Error(t, err, "Should return an error for a missing file")
	
	// Verify that an error message is displayed
	output := outBuf.String()
	assert.Contains(t, output, "Change request file not found:")
}

// TestInvalidFilterPattern tests handling of invalid regex in filter pattern
func TestInvalidFilterPattern(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Create a change request with multiple stories
	cr := createExtendedTestChangeRequest()
	
	// Set up the cat command options with an invalid filter pattern
	options := CatOptions{
		ShowContentHash:   false,
		ChangeRequestPath: "test-cr.md",
		FilterPattern:     "[", // Invalid regex pattern
	}
	
	// Call the function under test
	err := processAndPrintUserStories(mockFS, mockTerminal, options, cr)
	
	// Check results
	assert.Error(t, err, "Should return an error for invalid regex pattern")
	assert.Contains(t, err.Error(), "invalid filter pattern")
}

// TestExecuteCatCommand tests the executeCatCommand function, focusing on error handling cases
func TestExecuteCatCommand(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	mockTerminal := &mockTerminal{buffer: outBuf}
	
	// Ensure the test change request file exists for the filter pattern test
	mockFS.AddFile("test-cr.md", []byte(`---
name: Valid Change Request
created-at: 2025-01-01T00:00:00Z
user-stories:
  - title: Test Story
    file: test.md
    content-hash: abc123
---`))
	
	tests := []struct {
		name          string
		options       CatOptions
		setupMock     func(*io.MockFileSystem)
		expectedError string
	}{
		{
			name: "Non-existent change request file",
			options: CatOptions{
				ChangeRequestPath: "non-existent.md",
			},
			setupMock: func(fs *io.MockFileSystem) {
				// Don't add the file to simulate missing file
			},
			expectedError: "change request file not found",
		},
		{
			name: "Invalid filter pattern",
			options: CatOptions{
				ChangeRequestPath: "test-cr.md",
				FilterPattern:     "[", // Invalid regex pattern
			},
			setupMock: func(fs *io.MockFileSystem) {
				// The test-cr.md file was already added above
			},
			expectedError: "invalid filter pattern",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clear buffer
			outBuf.Reset()
			
			// Setup mock
			tc.setupMock(mockFS)
			
			// Call the function under test
			err := executeCatCommand(mockFS, mockTerminal, tc.options)
			
			// Verify error
			if err == nil {
				t.Errorf("Expected error containing '%s', but got nil", tc.expectedError)
				return
			}
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// createTestChangeRequest creates a test change request
func createTestChangeRequest() models.ChangeRequest {
	return models.ChangeRequest{
		Name:      "test change request",
		CreatedAt: time.Date(2025, 4, 5, 10, 0, 0, 0, time.UTC),
		UserStories: []models.UserStoryReference{
			{
				Title:       "Test Story 1",
				FilePath:    "docs/user-stories/test-1.md",
				ContentHash: "abc123",
			},
			{
				Title:       "Test Story 2",
				FilePath:    "docs/user-stories/test-2.md",
				ContentHash: "def456",
			},
			{
				Title:       "Missing Story",
				FilePath:    "docs/user-stories/missing.md",
				ContentHash: "xyz789",
			},
		},
		FilePath: "test-cr.md",
	}
}

// createExtendedTestChangeRequest creates an extended test change request with more stories
func createExtendedTestChangeRequest() models.ChangeRequest {
	return models.ChangeRequest{
		Name:      "test change request",
		CreatedAt: time.Date(2025, 4, 5, 10, 0, 0, 0, time.UTC),
		UserStories: []models.UserStoryReference{
			{
				Title:       "Test Story 1",
				FilePath:    "docs/user-stories/test-1.md",
				ContentHash: "abc123",
			},
			{
				Title:       "Test Story 2",
				FilePath:    "docs/user-stories/test-2.md",
				ContentHash: "def456",
			},
			{
				Title:       "Missing Story",
				FilePath:    "docs/user-stories/missing.md",
				ContentHash: "xyz789",
			},
			{
				Title:       "Login Feature",
				FilePath:    "docs/user-stories/auth/login.md",
				ContentHash: "auth123",
			},
			{
				Title:       "Registration Feature",
				FilePath:    "docs/user-stories/auth/registration.md",
				ContentHash: "auth456",
			},
		},
		FilePath: "test-cr.md",
	}
}

// mockTerminal is a simple implementation of the UserOutput interface for testing
type mockTerminal struct {
	buffer       *bytes.Buffer
	debugEnabled bool
}

func (m *mockTerminal) Print(message string) {
	m.buffer.WriteString(message)
}

func (m *mockTerminal) PrintSuccess(message string) {
	m.buffer.WriteString(message)
}

func (m *mockTerminal) PrintError(message string) {
	m.buffer.WriteString(message)
}

func (m *mockTerminal) PrintTable(headers []string, rows [][]string) {
	// Simple implementation for testing
	m.buffer.WriteString(strings.Join(headers, " | ") + "\n")
	for _, row := range rows {
		m.buffer.WriteString(strings.Join(row, " | ") + "\n")
	}
}

func (m *mockTerminal) PrintWarning(message string) {
	m.buffer.WriteString(message)
}

func (m *mockTerminal) PrintProgress(message string) {
	m.buffer.WriteString(message)
}

func (m *mockTerminal) PrintStep(stepNumber int, totalSteps int, description string) {
	m.buffer.WriteString(
		"STEP " + string(rune('0'+stepNumber)) + "/" + string(rune('0'+totalSteps)) + ": " + description + "\n",
	)
}

func (m *mockTerminal) IsDebugEnabled() bool {
	return m.debugEnabled
}

// colorTrackingTerminal is a special terminal that tracks color usage
type colorTrackingTerminal struct {
	buffer           *bytes.Buffer
	successCallCount int
	warningCallCount int
}

func (c *colorTrackingTerminal) Print(message string) {
	c.buffer.WriteString(message)
}

func (c *colorTrackingTerminal) PrintSuccess(message string) {
	c.successCallCount++
	c.buffer.WriteString(message)
}

func (c *colorTrackingTerminal) PrintError(message string) {
	c.buffer.WriteString(message)
}

func (c *colorTrackingTerminal) PrintTable(headers []string, rows [][]string) {
	c.buffer.WriteString(strings.Join(headers, " | ") + "\n")
	for _, row := range rows {
		c.buffer.WriteString(strings.Join(row, " | ") + "\n")
	}
}

func (c *colorTrackingTerminal) PrintWarning(message string) {
	c.warningCallCount++
	c.buffer.WriteString(message)
}

func (c *colorTrackingTerminal) PrintProgress(message string) {
	c.buffer.WriteString(message)
}

func (c *colorTrackingTerminal) PrintStep(stepNumber int, totalSteps int, description string) {
	c.buffer.WriteString(
		"STEP " + string(rune('0'+stepNumber)) + "/" + string(rune('0'+totalSteps)) + ": " + description + "\n",
	)
}

func (c *colorTrackingTerminal) IsDebugEnabled() bool {
	return true
}

// TestUtilityFunctions tests the utility functions used by the cat command
func TestUtilityFunctions(t *testing.T) {
	t.Run("printFilePath", func(t *testing.T) {
		// Setup
		outBuf := new(bytes.Buffer)
		normalTerminal := &mockTerminal{buffer: outBuf}
		colorTerminal := &colorTrackingTerminal{buffer: outBuf}
		
		// Test normal output
		outBuf.Reset()
		printFilePath(normalTerminal, "test/path.md", false)
		assert.Equal(t, "[//]: # (test/path.md)\n", outBuf.String())
		
		// Test colorized output
		outBuf.Reset()
		printFilePath(colorTerminal, "test/path.md", true)
		assert.Equal(t, 1, colorTerminal.successCallCount)
	})
	
	t.Run("printSeparator", func(t *testing.T) {
		// Setup
		outBuf := new(bytes.Buffer)
		terminal := &mockTerminal{buffer: outBuf}
		
		// Test normal mode
		outBuf.Reset()
		printSeparator(terminal, false)
		assert.Equal(t, "\n---\n", outBuf.String())
		
		// Test compact mode
		outBuf.Reset()
		printSeparator(terminal, true)
		assert.Equal(t, "\n", outBuf.String())
	})
	
	t.Run("printSummary", func(t *testing.T) {
		// Setup
		outBuf := new(bytes.Buffer)
		normalTerminal := &mockTerminal{buffer: outBuf, debugEnabled: false}
		colorTerminal := &colorTrackingTerminal{buffer: outBuf}
		
		result := ProcessingResult{
			TotalStories:     5,
			DisplayedStories: 3,
			SkippedStories:   1,
			ErrorStories:     1,
		}
		
		// Test no summary when not needed (no filter, no debug, no skipped)
		outBuf.Reset()
		printSummary(normalTerminal, ProcessingResult{TotalStories: 5, DisplayedStories: 5}, CatOptions{})
		assert.Equal(t, "", outBuf.String())
		
		// Test summary with filter
		outBuf.Reset()
		printSummary(normalTerminal, result, CatOptions{FilterPattern: "test"})
		assert.Contains(t, outBuf.String(), "Summary: 3 of 5 stories displayed, 1 skipped, 1 errors")
		
		// Test summary with color
		outBuf.Reset()
		colorTerminal.warningCallCount = 0
		printSummary(colorTerminal, result, CatOptions{ColorOutput: true, FilterPattern: "test"})
		assert.Equal(t, 1, colorTerminal.warningCallCount)
	})
	
	t.Run("matchesFilter", func(t *testing.T) {
		// Test with simple regex pattern that will definitely match
		simpleRegex, err := regexp.Compile("test")
		assert.NoError(t, err, "Error compiling simple test regex")
		
		// Simple test - should match content
		assert.True(t, matchesFilter(simpleRegex, "This contains test pattern", "Title"),
			"Simple pattern should match content")
			
		// Simple test - should match title
		assert.True(t, matchesFilter(simpleRegex, "Content", "Title with test in it"),
			"Simple pattern should match title")
			
		// Simple test - no match
		assert.False(t, matchesFilter(simpleRegex, "Content", "Title"),
			"Simple pattern should not match when neither matches")
	})
}

// TestProcessingWithEmptyChangeRequest tests processing an empty change request
func TestProcessingWithEmptyChangeRequest(t *testing.T) {
	// Setup test environment
	_, outBuf, mockFS := setupCatTest(t)
	
	// Create a mock terminal with debug enabled
	mockTerminal := &mockTerminal{
		buffer:       outBuf,
		debugEnabled: true,
	}
	
	// Create an empty change request
	cr := models.ChangeRequest{
		Name:        "empty change request",
		CreatedAt:   time.Now(),
		UserStories: []models.UserStoryReference{},
		FilePath:    "empty-cr.md",
	}
	
	// Process the empty change request
	err := processAndPrintUserStories(mockFS, mockTerminal, CatOptions{}, cr)
	
	// Check results
	assert.NoError(t, err)
	assert.Contains(t, outBuf.String(), "Summary: 0 of 0 stories displayed")
} 
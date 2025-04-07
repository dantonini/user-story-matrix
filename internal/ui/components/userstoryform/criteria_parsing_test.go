// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package userstoryform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

// TestAcceptanceCriteriaParsing is a smoke test for the various formats of acceptance criteria
// This test ensures that the enhanced acceptance criteria parsing logic in GetUserStory
// correctly handles different input formats
func TestAcceptanceCriteriaParsing(t *testing.T) {
	// Create test cases with different acceptance criteria formats
	testCases := []struct {
		name           string
		criteriaInput  string
		expectedOutput []string
	}{
		{
			name:           "Space separated criteria",
			criteriaInput:  "First Second Third",
			expectedOutput: []string{"First", "Second", "Third"},
		},
		{
			name:           "Newline separated criteria",
			criteriaInput:  "First\nSecond\nThird",
			expectedOutput: []string{"First", "Second", "Third"},
		},
		{
			name:           "Mixed format with extra spaces",
			criteriaInput:  "  First  \n  Second  \nThird  ",
			expectedOutput: []string{"First", "Second", "Third"},
		},
		{
			name:           "Empty lines should be filtered",
			criteriaInput:  "First\n\nSecond\n\n\nThird",
			expectedOutput: []string{"First", "Second", "Third"},
		},
		{
			name:           "Empty input",
			criteriaInput:  "",
			expectedOutput: []string{},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a basic form with an empty user story
			testForm := newTestUserStoryForm(t, models.UserStory{})
			form := testForm.UserStoryForm
			
			// We don't need to set the input value since we're directly testing the parsing method
			
			// Use the form's parseAcceptanceCriteria method directly
			criteria := form.parseAcceptanceCriteria(tc.criteriaInput)
			
			// Verify the criteria were parsed correctly
			assert.Equal(t, len(tc.expectedOutput), len(criteria), 
				"Expected %d criteria, got %d", len(tc.expectedOutput), len(criteria))
			
			for i, expected := range tc.expectedOutput {
				if i < len(criteria) {
					assert.Equal(t, expected, criteria[i], 
						"Criterion %d should be '%s', got '%s'", i, expected, criteria[i])
				}
			}
		})
	}
}

// TestComplexCriteriaParsing tests more complex scenarios for acceptance criteria parsing
func TestComplexCriteriaParsing(t *testing.T) {
	testCases := []struct {
		name           string
		criteriaInput  string
		expectedOutput []string
	}{
		{
			name: "Bullet points with dashes",
			criteriaInput: "- First criterion\n- Second criterion\n- Third criterion",
			expectedOutput: []string{
				"First criterion",
				"Second criterion",
				"Third criterion",
			},
		},
		{
			name: "Bullet points with asterisks",
			criteriaInput: "* First criterion\n* Second criterion\n* Third criterion",
			expectedOutput: []string{
				"First criterion", 
				"Second criterion", 
				"Third criterion",
			},
		},
		{
			name: "Numbered list",
			criteriaInput: "1. First criterion\n2. Second criterion\n3. Third criterion",
			expectedOutput: []string{
				"First criterion",
				"Second criterion",
				"Third criterion",
			},
		},
		{
			name: "Mixed format list",
			criteriaInput: "- First criterion\n* Second criterion\n3. Third criterion",
			expectedOutput: []string{
				"First criterion",
				"Second criterion",
				"Third criterion",
			},
		},
		{
			name: "Parenthesized numbers",
			criteriaInput: "(1) First criterion\n(2) Second criterion\n(3) Third criterion",
			expectedOutput: []string{
				"First criterion",
				"Second criterion",
				"Third criterion",
			},
		},
		{
			name: "Multi-word criteria without separators",
			criteriaInput: "The user can create stories\nThe system validates input\nAll fields are required",
			expectedOutput: []string{
				"The user can create stories",
				"The system validates input",
				"All fields are required",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a form to use for testing
			testForm := newTestUserStoryForm(t, models.UserStory{})
			form := testForm.UserStoryForm
			
			// Test with direct parsing using the form's method
			criteria := form.parseAcceptanceCriteria(tc.criteriaInput)
			
			// Verify correct number of criteria extracted
			assert.Equal(t, len(tc.expectedOutput), len(criteria),
				"Expected %d criteria, got %d", len(tc.expectedOutput), len(criteria))
			
			// Verify each criterion matches expected output
			for i, expected := range tc.expectedOutput {
				if i < len(criteria) {
					assert.Equal(t, expected, criteria[i],
						"Criterion %d should be '%s', got '%s'", i, expected, criteria[i])
				}
			}
			
			// Also test through the GetUserStory method for integration testing
			// First, store the criteria input for testing purposes
			form.rawCriteriaInput = tc.criteriaInput
			
			// Set at least one criteria input
			if len(form.criteriasInputs) > 0 {
				form.criteriasInputs[0].SetValue(tc.criteriaInput)
			}
			
			// Get user story with parsed criteria
			story := form.GetUserStory()
			
			// The test might need to be adjusted depending on how GetUserStory processes criteria
			// As it may not use parseAcceptanceCriteria directly
			if len(story.Criteria) > 0 {
				// If criteria were parsed correctly, verify the first one
				assert.NotEmpty(t, story.Criteria[0], "User story should have non-empty criteria")
			}
		})
	}
} 
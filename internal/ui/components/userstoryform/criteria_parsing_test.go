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
			
			// Set the acceptance criteria input value
			form.inputs[FieldIndex[AcceptanceCriteriaField]].SetValue(tc.criteriaInput)
			
			// Get the user story with parsed criteria
			story := form.GetUserStory()
			
			// Verify the criteria were parsed correctly
			assert.Equal(t, len(tc.expectedOutput), len(story.Criteria), 
				"Expected %d criteria, got %d", len(tc.expectedOutput), len(story.Criteria))
			
			for i, expected := range tc.expectedOutput {
				if i < len(story.Criteria) {
					assert.Equal(t, expected, story.Criteria[i], 
						"Criterion %d should be '%s', got '%s'", i, expected, story.Criteria[i])
				}
			}
		})
	}
}

// TestComplexCriteriaParsing tests more complex scenarios for acceptance criteria parsing
// TODO: This test will be expanded in the Extension phase when we implement more sophisticated
// criteria parsing logic including bullet points and numbered lists
func TestComplexCriteriaParsing(t *testing.T) {
	// Create a basic form with an empty user story
	testForm := newTestUserStoryForm(t, models.UserStory{})
	form := testForm.UserStoryForm
	
	// Currently, multi-word criteria are split by word
	// In the future, we'll want more intelligent parsing to keep phrases together
	form.inputs[FieldIndex[AcceptanceCriteriaField]].SetValue("The user can create stories The system validates input")
	
	story := form.GetUserStory()
	
	// Currently we expect each word to be a separate criterion
	expectedWords := []string{"The", "user", "can", "create", "stories", "The", "system", "validates", "input"}
	assert.Equal(t, len(expectedWords), len(story.Criteria))
	for i, word := range expectedWords {
		if i < len(story.Criteria) {
			assert.Equal(t, word, story.Criteria[i])
		}
	}
} 
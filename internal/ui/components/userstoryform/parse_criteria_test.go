// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package userstoryform

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestParseAcceptanceCriteriaEdgeCases(t *testing.T) {
	// Create a form instance for testing
	form := New(models.UserStory{}, nil, nil)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Mixed bullet styles",
			input:    "- First item\n* Second item\n• Third item",
			expected: []string{"First item", "Second item", "\xa2 Third item"},
		},
		{
			name:     "Numbered list with extra whitespace",
			input:    "1.   First numbered item   \n   2.   Second numbered item  ",
			expected: []string{"First numbered item", "Second numbered item"},
		},
		{
			name:     "Mixed bullet and numbered items",
			input:    "• First bullet\n2. Second numbered\n- Third dash",
			expected: []string{"\xa2 First bullet", "Second numbered", "Third dash"},
		},
		{
			name:     "Parenthesized numbers",
			input:    "(1) First item\n(2) Second item\n(3) Third item",
			expected: []string{"First item", "Second item", "Third item"},
		},
		{
			name:     "Multi-line with inconsistent formatting",
			input:    "First criterion\n  - Second criterion with indent\n    * Third criterion with more indent",
			expected: []string{"First criterion", "Second criterion with indent", "Third criterion with more indent"},
		},
		{
			name:     "Extra empty lines and whitespace",
			input:    "\n\n   First criterion   \n\n   Second criterion   \n\n",
			expected: []string{"First criterion", "Second criterion"},
		},
		{
			name:     "Single criterion with newlines",
			input:    "This is\na single criterion\nspanning multiple lines",
			expected: []string{"This is", "a single criterion", "spanning multiple lines"},
		},
		{
			name:     "Comma-separated criteria",
			input:    "First, Second, Third",
			expected: []string{"First,", "Second,", "Third"},
		},
		{
			name:     "Long multi-word criterion",
			input:    "This is a very long acceptance criterion that should be treated as a single item",
			expected: []string{"This is a very long acceptance criterion that should be treated as a single item"},
		},
		{
			name:     "Criteria with special characters",
			input:    "• First: validate user input\n• Second: check for edge-cases\n• Third: handle API responses",
			expected: []string{"\xa2 First: validate user input", "\xa2 Second: check for edge-cases", "\xa2 Third: handle API responses"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set the raw input for testing
			form.rawCriteriaInput = tc.input
			
			// Parse the criteria
			result := form.parseAcceptanceCriteria(tc.input)
			
			// Assert that the result matches the expected output
			assert.Equal(t, tc.expected, result, "Criteria parsing didn't match expected output")
		})
	}
}

func TestParseAcceptanceCriteriaBoundaries(t *testing.T) {
	// Create a form instance for testing
	form := New(models.UserStory{}, nil, nil)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Very short criteria",
			input:    "A B C",
			expected: []string{"A", "B", "C"},
		},
		{
			name:     "Very long criteria item",
			input:    strings.Repeat("This is a very long acceptance criterion. ", 10),
			expected: []string{strings.Repeat("This is a very long acceptance criterion. ", 10)[0:419]},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Only whitespace",
			input:    "     \n   \t   \n    ",
			expected: []string{},
		},
		{
			name:     "Many criteria",
			input:    strings.Join(generateNumberedList(50), "\n"),
			expected: generateNumberedListExpected(50),
		},
		{
			name:     "Single word",
			input:    "Criterion",
			expected: []string{"Criterion"},
		},
		{
			name:     "Single letter criteria",
			input:    "A B C D E",
			expected: []string{"A B C D E"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set the raw input for testing
			form.rawCriteriaInput = tc.input
			
			// Parse the criteria
			result := form.parseAcceptanceCriteria(tc.input)
			
			// Assert that the result matches the expected output
			assert.Equal(t, tc.expected, result, "Criteria parsing didn't match expected output")
		})
	}
}

// Helper function to generate a numbered list of criteria
func generateNumberedList(count int) []string {
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = fmt.Sprintf("%d. Criterion number %d", i+1, i+1)
	}
	return result
}

// Helper function to generate the expected output from a numbered list
func generateNumberedListExpected(count int) []string {
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = fmt.Sprintf("Criterion number %d", i+1)
	}
	return result
}

func TestGetConfidenceIndicator(t *testing.T) {
	tests := []struct {
		name       string
		confidence float64
		contains   string // We check if the result contains this string
	}{
		{
			name:       "High confidence",
			confidence: 0.9,
			contains:   "✓",
		},
		{
			name:       "Medium confidence",
			confidence: 0.6,
			contains:   "◎",
		},
		{
			name:       "Low confidence",
			confidence: 0.3,
			contains:   "?",
		},
		{
			name:       "Exact boundary - high",
			confidence: 0.8,
			contains:   "✓",
		},
		{
			name:       "Exact boundary - medium",
			confidence: 0.5,
			contains:   "◎",
		},
		{
			name:       "Zero confidence",
			confidence: 0.0,
			contains:   "?",
		},
		{
			name:       "Full confidence",
			confidence: 1.0,
			contains:   "✓",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getConfidenceIndicator(tc.confidence)
			assert.Contains(t, result, tc.contains)
			assert.Contains(t, result, fmt.Sprintf("%.0f%%", tc.confidence*100))
		})
	}
}

func TestGetFieldLabel(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{
			name:     "Title field",
			field:    TitleField,
			expected: "Title:",
		},
		{
			name:     "As a field",
			field:    AsAField,
			expected: "As a:",
		},
		{
			name:     "I want field",
			field:    IWantField,
			expected: "I want:",
		},
		{
			name:     "So that field",
			field:    SoThatField,
			expected: "So that:",
		},
		{
			name:     "Acceptance Criteria field",
			field:    AcceptanceCriteriaField,
			expected: "Acceptance Criteria:",
		},
		{
			name:     "Unknown field",
			field:    "unknown_field",
			expected: "unknown_field:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getFieldLabel(tc.field)
			assert.Equal(t, tc.expected, result)
		})
	}
} 
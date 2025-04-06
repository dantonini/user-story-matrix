// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package clipboard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRegressionMiddlePasteScenarios verifies that the enhanced paste detection
// correctly identifies pastes in the middle of existing text
func TestRegressionMiddlePasteScenarios(t *testing.T) {
	testCases := []struct {
		name         string
		previousText string
		currentText  string
		shouldDetect bool
		minLength    int // Minimum length the detected paste should be
	}{
		{
			name:         "Short user story to full user story with components",
			previousText: "As a user",
			currentText: "As a user, I want to create a component that displays a list of items, " +
				"so that I can show data in a structured format.\n\n" +
				"Acceptance criteria:\n" +
				"- The component should accept an array of items\n" +
				"- Each item should have a title and description\n" +
				"- The component should handle empty arrays gracefully",
			shouldDetect: true,
			minLength:    100,
		},
		{
			name:         "Code snippet with function declaration to complete function",
			previousText: "func parseData(input string) {",
			currentText: "func parseData(input string) {\n" +
				"    if strings.TrimSpace(input) == \"\" {\n" +
				"        return nil, errors.New(\"empty input\")\n" +
				"    }\n\n" +
				"    var result []string\n" +
				"    lines := strings.Split(input, \"\\n\")\n" +
				"    for _, line := range lines {\n" +
				"        if strings.TrimSpace(line) != \"\" {\n" +
				"            result = append(result, strings.TrimSpace(line))\n" +
				"        }\n" +
				"    }\n" +
				"    return result, nil\n" +
				"}",
			shouldDetect: true,
			minLength:    50,
		},
		{
			name:         "Nested text insertion",
			previousText: "This text has [placeholder] that needs to be filled.",
			currentText:  "This text has [a complex structure with multiple levels of information] that needs to be filled.",
			shouldDetect: true,
			minLength:    10,
		},
		{
			name:         "Text with bullet points and special characters",
			previousText: "Requirements:",
			currentText: "Requirements:\n" +
				"• Should handle UTF-8 characters: ö, ñ, 漢字, ひらがな\n" +
				"• Must work with lists and bullet points\n" +
				"• Should detect code snippets like `code.function()`\n" +
				"• Must handle formatted text with *emphasis* and **strong emphasis**",
			shouldDetect: true,
			minLength:    50,
		},
		{
			name:         "Very similar beginning and end with large middle insertion",
			previousText: "The quick brown fox jumps over the lazy dog.",
			currentText: "The quick brown fox jumps over the fence, runs through the meadow, " +
				"chases the rabbit, crosses the stream, climbs the hill, and finally jumps over the lazy dog.",
			shouldDetect: true,
			minLength:    50,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Try both detection methods
			content, detected := GetActiveFieldValue(tc.currentText, tc.previousText)
			
			// Assert detection matches expectation
			assert.Equal(t, tc.shouldDetect, detected, 
				"Detection should be %v but got %v", tc.shouldDetect, detected)
			
			if tc.shouldDetect {
				// If detection is expected, the content should be of sufficient length
				assert.GreaterOrEqual(t, len(content), tc.minLength,
					"Detected content length %d should be at least %d", len(content), tc.minLength)
				
				// The detected content should be different from the previous text
				assert.NotEqual(t, tc.previousText, content,
					"Detected content should differ from previous text")
				
				// The content should make sense in context
				if tc.name == "Code snippet with function declaration to complete function" {
					assert.Contains(t, content, "strings.Split")
				} else if tc.name == "Text with bullet points and special characters" {
					assert.Contains(t, content, "UTF-8")
				}
			} else {
				// If we don't expect detection, content should be empty
				assert.Empty(t, content)
			}
			
			// Try the more direct implementation
			if tc.shouldDetect {
				directResult, directDetected := detectMiddlePaste(tc.previousText, tc.currentText)
				assert.True(t, directDetected, 
					"Direct middle paste detection should succeed for %s", tc.name)
				assert.NotEmpty(t, directResult,
					"Direct middle paste detection should return content for %s", tc.name)
			}
		})
	}
}

// TestRegressionLargeInsertScenarios verifies the enhanced large insertion
// detection logic correctly identifies large chunks of text inserted into existing content
func TestRegressionLargeInsertScenarios(t *testing.T) {
	testCases := []struct {
		name         string
		previousText string
		currentText  string
		shouldDetect bool
		contentCheck string // A substring that should be in the detected content
	}{
		{
			name:         "Insert markdown list in the middle of a paragraph",
			previousText: "Here's a list of requirements for the feature. Let me know if you need more details.",
			currentText: "Here's a list of requirements for the feature:\n\n" +
				"1. User authentication\n" +
				"2. Data persistence\n" +
				"3. Responsive design\n" +
				"4. Accessibility compliance\n" +
				"5. Internationalization support\n\n" +
				"Let me know if you need more details.",
			shouldDetect: true,
			contentCheck: "Data persistence",
		},
		{
			name:         "Insert JSON data in a code comment",
			previousText: "// Configuration options:",
			currentText: "// Configuration options:\n" +
				"/*\n" +
				"{\n" +
				"  \"api_key\": \"YOUR_API_KEY_HERE\",\n" +
				"  \"endpoint\": \"https://api.example.com/v1\",\n" +
				"  \"timeout_ms\": 5000,\n" +
				"  \"retry_count\": 3,\n" +
				"  \"log_level\": \"info\"\n" +
				"}\n" +
				"*/",
			shouldDetect: true,
			contentCheck: "api_key",
		},
		{
			name:         "Add complete function implementation",
			previousText: "// TODO: Implement the parseAcceptanceCriteria function",
			currentText: "// TODO: Implement the parseAcceptanceCriteria function\n" +
				"func parseAcceptanceCriteria(input string) []string {\n" +
				"    if strings.TrimSpace(input) == \"\" {\n" +
				"        return []string{}\n" +
				"    }\n\n" +
				"    var criteria []string\n" +
				"    lines := strings.Split(input, \"\\n\")\n\n" +
				"    for _, line := range lines {\n" +
				"        line = strings.TrimSpace(line)\n" +
				"        if line != \"\" {\n" +
				"            criteria = append(criteria, line)\n" +
				"        }\n" +
				"    }\n\n" +
				"    return criteria\n" +
				"}",
			shouldDetect: true,
			contentCheck: "TrimSpace",
		},
		{
			name:         "Replace placeholder with actual content",
			previousText: "User story: [PLACEHOLDER]",
			currentText: "User story: As a developer, I want to implement the clipboard detection feature, " +
				"so that I can automatically process pasted content through the LLM service and " +
				"populate form fields with structured data.",
			shouldDetect: true,
			contentCheck: "clipboard detection",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with detectLargeInsert
			insertedContent := detectLargeInsert(tc.previousText, tc.currentText)
			
			if tc.shouldDetect {
				// Content should be detected
				assert.NotEmpty(t, insertedContent, 
					"Should detect inserted content for %s", tc.name)
				
				// The detected content should contain the expected substring
				assert.Contains(t, insertedContent, tc.contentCheck,
					"Detected content should contain %q", tc.contentCheck)
				
				// The detected content should be large enough
				assert.GreaterOrEqual(t, len(insertedContent), len(tc.contentCheck),
					"Detected content should be at least as long as the check string")
			} else {
				// No content should be detected
				assert.Empty(t, insertedContent,
					"Should not detect any inserted content for %s", tc.name)
			}
			
			// Test with GetActiveFieldValue
			content, detected := GetActiveFieldValue(tc.currentText, tc.previousText)
			
			// Verification for GetActiveFieldValue
			if tc.shouldDetect {
				assert.True(t, detected,
					"GetActiveFieldValue should detect the paste for %s", tc.name)
				assert.NotEmpty(t, content,
					"GetActiveFieldValue should return content for %s", tc.name)
				assert.Contains(t, content, tc.contentCheck,
					"GetActiveFieldValue content should contain %q", tc.contentCheck)
			}
		})
	}
}

// TestRegressionComplexEdgeCases verifies edge cases that might challenge
// the paste detection algorithms
func TestRegressionComplexEdgeCases(t *testing.T) {
	testCases := []struct {
		name           string
		previousText   string
		currentText    string
		shouldDetect   bool
		expectedResult string
	}{
		{
			name:           "Repeated text patterns",
			previousText:   "This is a test. This is a test. This is a test.",
			currentText:    "This is a test. This is a INSERTED CONTENT test. This is a test.",
			shouldDetect:   true,
			expectedResult: "INSERTED CONTENT ",
		},
		{
			name:           "Very similar strings with small difference",
			previousText:   "The application should process text input and extract structured data.",
			currentText:    "The application should process text input efficiently and extract structured data.",
			shouldDetect:   false,
			expectedResult: "",
		},
		{
			name:           "Text with special characters and formatting",
			previousText:   "Title: Test Document",
			currentText:    "Title: Test Document\n\n## Heading 1\n\n* Bullet 1\n* Bullet 2\n\n```code block```\n\n> Blockquote",
			shouldDetect:   true,
			expectedResult: "\n\n## Heading",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Direct middle paste detection
			content, detected := detectMiddlePaste(tc.previousText, tc.currentText)
			
			assert.Equal(t, tc.shouldDetect, detected,
				"detectMiddlePaste should %v for %s", 
				map[bool]string{true: "detect", false: "not detect"}[tc.shouldDetect], 
				tc.name)
			
			if tc.shouldDetect && detected {
				if tc.expectedResult != "" {
					assert.Contains(t, content, tc.expectedResult,
						"Detected content should contain %q", tc.expectedResult)
				}
			} else if !tc.shouldDetect {
				assert.Empty(t, content, 
					"Content should be empty for non-detected case %s", tc.name)
			}
			
			// Test through the main public interface
			apiContent, apiDetected := GetActiveFieldValue(tc.currentText, tc.previousText)
			
			assert.Equal(t, tc.shouldDetect, apiDetected && len(apiContent) > 0,
				"GetActiveFieldValue should %v for %s", 
				map[bool]string{true: "detect", false: "not detect"}[tc.shouldDetect], 
				tc.name)
		})
	}
} 
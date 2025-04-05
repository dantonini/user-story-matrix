// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package clipboard

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestIsPasteEvent(t *testing.T) {
	tests := []struct {
		name     string
		keyMsg   tea.KeyMsg
		expected bool
	}{
		{
			name:     "Ctrl+V",
			keyMsg:   tea.KeyMsg{Type: tea.KeyCtrlV},
			expected: true,
		},
		{
			name:     "SYN character",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{22}},
			expected: true,
		},
		{
			name:     "Regular key press",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			expected: false,
		},
		{
			name:     "Enter key press",
			keyMsg:   tea.KeyMsg{Type: tea.KeyEnter},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsPasteEvent(test.keyMsg)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsLongEnoughForProcessing(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "Empty string",
			content:  "",
			expected: false,
		},
		{
			name:     "Short string",
			content:  "Short text",
			expected: false,
		},
		{
			name:     "Text at exactly threshold",
			content:  generateStringOfLength(PasteThresholdLength),
			expected: true,
		},
		{
			name:     "Long text",
			content:  generateStringOfLength(PasteThresholdLength + 20),
			expected: true,
		},
		{
			name:     "Long text with whitespace",
			content:  "   " + generateStringOfLength(PasteThresholdLength) + "   ",
			expected: true,
		},
		{
			name:     "Text below threshold with whitespace",
			content:  "   " + generateStringOfLength(PasteThresholdLength-10) + "   ",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsLongEnoughForProcessing(test.content)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestExtractPastedText(t *testing.T) {
	tests := []struct {
		name     string
		keyMsg   tea.KeyMsg
		expected string
	}{
		{
			name:     "Key runes",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Pasted text")},
			expected: "Pasted text",
		},
		{
			name:     "Ctrl+V",
			keyMsg:   tea.KeyMsg{Type: tea.KeyCtrlV},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractPastedText(test.keyMsg)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetActiveFieldValue(t *testing.T) {
	tests := []struct {
		name          string
		currentValue  string
		previousValue string
		expectedText  string
		expectedFound bool
	}{
		{
			name:          "No change",
			currentValue:  "Original text",
			previousValue: "Original text",
			expectedText:  "",
			expectedFound: false,
		},
		{
			name:          "Small change",
			currentValue:  "Original texta",
			previousValue: "Original text",
			expectedText:  "",
			expectedFound: false,
		},
		{
			name:          "Large paste",
			currentValue:  "Original text" + generateStringOfLength(PasteThresholdLength),
			previousValue: "Original text",
			expectedText:  generateStringOfLength(PasteThresholdLength),
			expectedFound: true,
		},
		{
			name:          "Paste at start",
			currentValue:  generateStringOfLength(PasteThresholdLength) + "Original text",
			previousValue: "Original text",
			expectedText:  generateStringOfLength(PasteThresholdLength),
			expectedFound: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, found := GetActiveFieldValue(test.currentValue, test.previousValue)
			assert.Equal(t, test.expectedText, text)
			assert.Equal(t, test.expectedFound, found)
		})
	}
}

// Helper to generate a string of specific length
func generateStringOfLength(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "a"
	}
	return result
} 
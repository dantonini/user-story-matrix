// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package clipboard

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PasteThresholdLength is the minimum length of text that will be considered
// as a potential user story to process.
const PasteThresholdLength = 50

// PasteMsg represents a clipboard paste event
type PasteMsg struct {
	Content string
}

// IsPasteEvent checks if a keyboard message is a paste event
func IsPasteEvent(msg tea.KeyMsg) bool {
	// Check for common paste key combinations
	return (msg.Type == tea.KeyCtrlV) ||
		(msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 22) // ASCII 22 (SYN) - Ctrl+V on some terminals
}

// IsLongEnoughForProcessing checks if the content is long enough to be processed as a user story
func IsLongEnoughForProcessing(content string) bool {
	return len(strings.TrimSpace(content)) >= PasteThresholdLength
}

// ExtractPastedText extracts text from a paste event
// In a real terminal, we can't actually get the clipboard content directly
// from a tea.KeyMsg in all cases, but we can handle some common scenarios
func ExtractPastedText(msg tea.KeyMsg) string {
	// Get pasted text from key runes if available
	if msg.Type == tea.KeyRunes {
		return string(msg.Runes)
	}
	
	// In most other cases, we can't extract the clipboard content directly
	// and will need to rely on the actual pasted text in the input element
	return ""
}

// GetActiveFieldValue returns the value of the currently active field
// This helps detect when a field suddenly has a lot more content
func GetActiveFieldValue(currentValue, previousValue string) (string, bool) {
	currentLen := len(currentValue)
	previousLen := len(previousValue)
	
	// If the current value is significantly longer than the previous value,
	// it might be a paste event
	if currentLen > previousLen && (currentLen-previousLen) >= PasteThresholdLength {
		// Check if the paste was at the end
		if strings.HasPrefix(currentValue, previousValue) {
			// Extract the newly added content at the end
			newContent := currentValue[previousLen:]
			return newContent, true
		}
		
		// Check if the paste was at the beginning
		if strings.HasSuffix(currentValue, previousValue) {
			// Extract the newly added content at the beginning
			newContent := currentValue[:currentLen-previousLen]
			return newContent, true
		}
		
		// TODO: In the MVI phase, implement detection for paste in the middle
		// of text and more complex paste scenarios
	}
	
	return "", false
} 
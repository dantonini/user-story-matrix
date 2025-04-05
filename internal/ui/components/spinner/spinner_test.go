// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package spinner

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestNewSpinner(t *testing.T) {
	// Act
	spinner := New()
	
	// Assert
	assert.Equal(t, "Processing...", spinner.Message)
	assert.Equal(t, "", spinner.AdditionalMessage)
	assert.Equal(t, false, spinner.Visible)
	assert.Equal(t, 80, spinner.Width)
	assert.Equal(t, lipgloss.Color("205"), spinner.ForegroundColor)
}

func TestSpinnerVisibility(t *testing.T) {
	// Arrange
	spinner := New()
	
	// Act & Assert
	
	// Initially invisible
	assert.Equal(t, "", spinner.View())
	
	// Set visible
	spinner.SetVisible(true)
	assert.True(t, spinner.Visible)
	assert.NotEqual(t, "", spinner.View())
	
	// Set invisible again
	spinner.SetVisible(false)
	assert.False(t, spinner.Visible)
	assert.Equal(t, "", spinner.View())
}

func TestSpinnerMessages(t *testing.T) {
	// Arrange
	spinner := New()
	spinner.SetVisible(true)
	
	// Act
	spinner.SetMessage("Loading data...")
	spinner.SetAdditionalMessage("This might take a few seconds")
	
	// Assert
	view := spinner.View()
	assert.Contains(t, view, "Loading data...")
	assert.Contains(t, view, "This might take a few seconds")
}

func TestSpinnerUpdate(t *testing.T) {
	// Arrange
	spinner := New()
	spinner.SetVisible(true)
	
	// Act
	initialView := spinner.View()
	updatedSpinner, _ := spinner.Update(tea.Tick(time.Millisecond, func(t time.Time) tea.Msg { return nil }))
	updatedView := updatedSpinner.View()
	
	// Assert
	assert.NotNil(t, updatedSpinner)
	// The view should change after an update due to spinner animation
	// But both should contain the message
	assert.Contains(t, initialView, "Processing...")
	assert.Contains(t, updatedView, "Processing...")
}

func TestSpinnerWidth(t *testing.T) {
	// Arrange
	spinner := New()
	spinner.SetVisible(true)
	spinner.SetWidth(20)
	
	// Act - Set a long additional message that should be truncated
	longMessage := "This is a very long message that should be truncated based on the width setting"
	spinner.SetAdditionalMessage(longMessage)
	
	// Assert
	view := spinner.View()
	lines := strings.Split(view, "\n")
	
	// Should have two lines (spinner+message and additional message)
	if len(lines) > 1 {
		additionalMessageLine := lines[1]
		// The rendered line should end with "..." and be <= width
		// Note: Exact length check is tricky because of ANSI color codes
		assert.True(t, strings.Contains(additionalMessageLine, "..."), "Long message should be truncated")
		assert.True(t, len(stripANSI(additionalMessageLine)) <= 20, "Truncated message should respect width")
	} else {
		t.Fail()
	}
}

func TestSpinnerForegroundColor(t *testing.T) {
	// Arrange
	spinner := New()
	
	// Act
	initialColor := spinner.ForegroundColor
	spinner.SetForegroundColor(lipgloss.Color("12")) // Set to blue
	
	// Assert
	assert.Equal(t, lipgloss.Color("205"), initialColor)
	assert.Equal(t, lipgloss.Color("12"), spinner.ForegroundColor)
}

func TestCreateTimeoutTicker(t *testing.T) {
	// Act
	cmd := CreateTimeoutTicker()
	
	// Assert
	assert.NotNil(t, cmd)
}

// Helper function to strip ANSI color codes for length checking
func stripANSI(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]+)*|[a-zA-Z\\d]+(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]*)*)?\\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-ntqry=><~]))"
	r := strings.NewReplacer("\u001B", "", "\u009B", "")
	return r.Replace(str)
} 
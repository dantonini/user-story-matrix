// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package spinner

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents a spinner component that displays a loading animation
type Model struct {
	// Spinner is the underlying spinner component
	Spinner spinner.Model
	
	// Message is the text to display next to the spinner
	Message string
	
	// AdditionalMessage is a secondary message displayed below the spinner
	AdditionalMessage string
	
	// ForegroundColor sets the color of the spinner
	ForegroundColor lipgloss.Color
	
	// MessageStyle is the style for the message text
	MessageStyle lipgloss.Style
	
	// AdditionalMessageStyle is the style for the additional message text
	AdditionalMessageStyle lipgloss.Style
	
	// Visible controls whether the spinner is displayed
	Visible bool
	
	// Width is the maximum width of the spinner component
	Width int
}

// New creates a new spinner model with default settings
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	return Model{
		Spinner:                s,
		Message:                "Processing...",
		AdditionalMessage:      "",
		ForegroundColor:        lipgloss.Color("205"),
		MessageStyle:           lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		AdditionalMessageStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
		Visible:                false,
		Width:                  80,
	}
}

// Init initializes the spinner
func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

// Update updates the spinner model
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}
	
	var cmd tea.Cmd
	m.Spinner, cmd = m.Spinner.Update(msg)
	return m, cmd
}

// View renders the spinner component
func (m Model) View() string {
	if !m.Visible {
		return ""
	}
	
	var sb strings.Builder
	
	// Create the spinner with message text
	spinnerText := fmt.Sprintf("%s %s", m.Spinner.View(), m.Message)
	sb.WriteString(m.MessageStyle.Render(spinnerText))
	
	// Add additional message if present
	if m.AdditionalMessage != "" {
		sb.WriteString("\n")
		truncated := m.AdditionalMessage
		if len(truncated) > m.Width {
			truncated = truncated[:m.Width-3] + "..."
		}
		sb.WriteString(m.AdditionalMessageStyle.Render(truncated))
	}
	
	return sb.String()
}

// SetMessage sets the spinner message
func (m *Model) SetMessage(message string) {
	m.Message = message
}

// SetAdditionalMessage sets the additional message displayed below the spinner
func (m *Model) SetAdditionalMessage(message string) {
	m.AdditionalMessage = message
}

// SetVisible controls whether the spinner is displayed
func (m *Model) SetVisible(visible bool) {
	m.Visible = visible
}

// SetWidth sets the maximum width of the spinner component
func (m *Model) SetWidth(width int) {
	m.Width = width
}

// SetForegroundColor sets the color of the spinner
func (m *Model) SetForegroundColor(color lipgloss.Color) {
	m.ForegroundColor = color
	m.Spinner.Style = m.Spinner.Style.Copy().Foreground(color)
	m.MessageStyle = m.MessageStyle.Copy().Foreground(color)
}

// CreateTimeoutTicker creates a command that sends time update messages
// This is useful for updating the spinner with timeout information
func CreateTimeoutTicker() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return timeoutTickMsg(t)
	})
}

// TimeoutTickMsg represents a timeout tick message
type timeoutTickMsg time.Time 
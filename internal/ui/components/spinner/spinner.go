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

// SpinnerStyle represents predefined spinner animation styles
type SpinnerStyle string

// Predefined spinner styles
const (
	StyleDot        SpinnerStyle = "dot"
	StyleLine       SpinnerStyle = "line"
	StyleMinimalDot SpinnerStyle = "minimal_dot"
	StyleJump       SpinnerStyle = "jump"
	StylePulse      SpinnerStyle = "pulse"
	StylePoints     SpinnerStyle = "points"
	StyleGlobe      SpinnerStyle = "globe"
	StyleMoon       SpinnerStyle = "moon"
	StyleMeter      SpinnerStyle = "meter"
	StyleMonkey     SpinnerStyle = "monkey"
	StyleHamburger  SpinnerStyle = "hamburger"
	StyleEllipsis   SpinnerStyle = "ellipsis"
	StyleBars       SpinnerStyle = "bars"         // Custom style
	StyleClock      SpinnerStyle = "clock"        // Custom style
	StyleCircle     SpinnerStyle = "circle"       // Custom style
	StyleGrowingBar SpinnerStyle = "growing_bar"  // Custom style
)

// Custom spinner definitions
var (
	// Bars spinner (vertical loading bars)
	barsSpinner = spinner.Spinner{
		Frames: []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▂", "▁"},
		FPS:    time.Second / 5,
	}
	
	// Clock spinner (like an analog clock face)
	clockSpinner = spinner.Spinner{
		Frames: []string{"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚"},
		FPS:    time.Second / 2,
	}
	
	// Circle spinner (smooth circular animation)
	circleSpinner = spinner.Spinner{
		Frames: []string{"◐", "◓", "◑", "◒"},
		FPS:    time.Second / 3,
	}
	
	// Growing Bar spinner
	growingBarSpinner = spinner.Spinner{
		Frames: []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
		FPS:    time.Second / 4,
	}
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
	
	// Style represents the current spinner style
	Style SpinnerStyle
	
	// FPS is the frames per second for the spinner animation
	FPS time.Duration
}

// New creates a new spinner model with default settings
func New() Model {
	s := spinner.New()
	
	// Use our custom circle spinner as the default
	s.Spinner = circleSpinner
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
		Style:                  StyleCircle,
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
	
	// If we're visible but no command was returned, make sure to keep ticking
	if cmd == nil {
		return m, m.Spinner.Tick
	}
	
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
// Returns a tea.Cmd if visibility is being enabled to start animation
func (m *Model) SetVisible(visible bool) tea.Cmd {
	// Only return a command if we're enabling visibility
	if visible && !m.Visible {
		m.Visible = true
		return m.Spinner.Tick
	}
	
	m.Visible = visible
	return nil
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

// SetStyle sets the spinner animation style
func (m *Model) SetStyle(style SpinnerStyle) {
	m.Style = style
	
	// Select the spinner type based on style
	var spinnerType spinner.Spinner
	
	switch style {
	case StyleDot:
		spinnerType = spinner.Dot
	case StyleLine:
		spinnerType = spinner.Line
	case StyleMinimalDot:
		spinnerType = spinner.MiniDot
	case StyleJump:
		spinnerType = spinner.Jump
	case StylePulse:
		spinnerType = spinner.Pulse
	case StylePoints:
		spinnerType = spinner.Points
	case StyleGlobe:
		spinnerType = spinner.Globe
	case StyleMoon:
		spinnerType = spinner.Moon
	case StyleMeter:
		spinnerType = spinner.Meter
	case StyleMonkey:
		spinnerType = spinner.Monkey
	case StyleHamburger:
		spinnerType = spinner.Hamburger
	case StyleEllipsis:
		spinnerType = spinner.Ellipsis
	case StyleBars:
		spinnerType = barsSpinner
	case StyleClock:
		spinnerType = clockSpinner
	case StyleCircle:
		spinnerType = circleSpinner
	case StyleGrowingBar:
		spinnerType = growingBarSpinner
	default:
		spinnerType = spinner.Dot
	}
	
	// Slow down the animation by applying a custom frame rate
	spinnerType.FPS = time.Second / 3  // 3 frames per second for a more pleasant pace
	
	// Apply the customized spinner
	m.Spinner.Spinner = spinnerType
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
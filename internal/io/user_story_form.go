// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package io

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user-story-matrix/usm/internal/models"
)

// Field represents a form field type
type UserStoryFieldType int

const (
	USTitleField UserStoryFieldType = iota
	USDescriptionField
	USAsField
	USWantField
	USSoThatField
	USAcceptanceCriteriaField
)

// UserStoryForm is a tea.Model for the user story form
type UserStoryForm struct {
	us                models.UserStory
	titleInput        textinput.Model
	descInput         textinput.Model
	asInput          textinput.Model
	wantInput        textinput.Model
	soThatInput      textinput.Model
	acInputs         []textinput.Model
	activeField      UserStoryFieldType
	activeACIndex    int
	ConfirmSubmission bool
	cancel           bool
	focused          bool
	width            int
	height           int
	err              error
}

// NewUserStoryForm creates a new user story form
func NewUserStoryForm(us models.UserStory) *UserStoryForm {
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter title"
	titleInput.Focus()
	titleInput.Width = 80
	titleInput.CharLimit = 100
	titleInput.SetValue(us.Title)

	descInput := textinput.New()
	descInput.Placeholder = "Enter description"
	descInput.Width = 80
	descInput.CharLimit = 200

	asInput := textinput.New()
	asInput.Placeholder = "Enter user type (As a ...)"
	asInput.Width = 80
	asInput.CharLimit = 100

	wantInput := textinput.New()
	wantInput.Placeholder = "Enter desired capability (I want ...)"
	wantInput.Width = 80
	wantInput.CharLimit = 100

	soThatInput := textinput.New()
	soThatInput.Placeholder = "Enter benefit (so that ...)"
	soThatInput.Width = 80
	soThatInput.CharLimit = 100

	// Create 5 acceptance criteria inputs
	acInputs := make([]textinput.Model, 5)
	for i := 0; i < 5; i++ {
		acInputs[i] = textinput.New()
		acInputs[i].Placeholder = fmt.Sprintf("Enter acceptance criteria %d", i+1)
		acInputs[i].Width = 80
		acInputs[i].CharLimit = 200
	}

	form := &UserStoryForm{
		us:               us,
		titleInput:       titleInput,
		descInput:        descInput,
		asInput:         asInput,
		wantInput:       wantInput,
		soThatInput:     soThatInput,
		acInputs:        acInputs,
		activeField:     USTitleField,
		activeACIndex:   0,
		ConfirmSubmission: false,
		cancel:          false,
		focused:         true,
		width:           80,
		height:          24,
	}

	return form
}

// Init initializes the form
func (f *UserStoryForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles user input events
func (f *UserStoryForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			// Only cancel if at least one field has content
			if f.hasContent() {
				f.cancel = true
				return f, tea.Quit
			}
			// If no content, just quit without setting cancel flag
			return f, tea.Quit

		case tea.KeyTab:
			// Move to the next field
			cmd = f.nextField()
			if cmd != nil {
				return f, cmd
			}

		case tea.KeyShiftTab:
			// Move to the previous field
			f.prevField()

		case tea.KeyEnter:
			// Move to the next field
			cmd = f.nextField()
			if cmd != nil {
				return f, cmd
			}

		default:
			// Handle other keys based on active field
			switch f.activeField {
			case USTitleField:
				f.titleInput, cmd = f.titleInput.Update(msg)
				cmds = append(cmds, cmd)
			case USDescriptionField:
				f.descInput, cmd = f.descInput.Update(msg)
				cmds = append(cmds, cmd)
			case USAsField:
				f.asInput, cmd = f.asInput.Update(msg)
				cmds = append(cmds, cmd)
			case USWantField:
				f.wantInput, cmd = f.wantInput.Update(msg)
				cmds = append(cmds, cmd)
			case USSoThatField:
				f.soThatInput, cmd = f.soThatInput.Update(msg)
				cmds = append(cmds, cmd)
			case USAcceptanceCriteriaField:
				if f.activeACIndex < len(f.acInputs) {
					f.acInputs[f.activeACIndex], cmd = f.acInputs[f.activeACIndex].Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		}

	case tea.WindowSizeMsg:
		f.width = msg.Width
		f.height = msg.Height
	}

	return f, tea.Batch(cmds...)
}

// View renders the form
func (f *UserStoryForm) View() string {
	var b strings.Builder

	// Form title - aligned to the left with no extra spaces
	formTitleStyle := lipgloss.NewStyle().Bold(true).AlignHorizontal(lipgloss.Left)
	b.WriteString(formTitleStyle.Render("User Story Form") + "\n\n")

	// Show all fields
	// Highlight the active field with different styling
	titleStyle := lipgloss.NewStyle()
	descStyle := lipgloss.NewStyle()
	asStyle := lipgloss.NewStyle()
	wantStyle := lipgloss.NewStyle()
	soThatStyle := lipgloss.NewStyle()
	ac1Style := lipgloss.NewStyle()
	ac2Style := lipgloss.NewStyle()
	ac3Style := lipgloss.NewStyle()
	ac4Style := lipgloss.NewStyle()
	ac5Style := lipgloss.NewStyle()
	
	switch f.activeField {
	case USTitleField:
		titleStyle = titleStyle.Bold(true).Foreground(lipgloss.Color("5"))
	case USDescriptionField:
		descStyle = descStyle.Bold(true).Foreground(lipgloss.Color("5"))
	case USAsField:
		asStyle = asStyle.Bold(true).Foreground(lipgloss.Color("5"))
	case USWantField:
		wantStyle = wantStyle.Bold(true).Foreground(lipgloss.Color("5"))
	case USSoThatField:
		soThatStyle = soThatStyle.Bold(true).Foreground(lipgloss.Color("5"))
	case USAcceptanceCriteriaField:
		switch f.activeACIndex {
		case 0:
			ac1Style = ac1Style.Bold(true).Foreground(lipgloss.Color("5"))
		case 1:
			ac2Style = ac2Style.Bold(true).Foreground(lipgloss.Color("5"))
		case 2:
			ac3Style = ac3Style.Bold(true).Foreground(lipgloss.Color("5"))
		case 3:
			ac4Style = ac4Style.Bold(true).Foreground(lipgloss.Color("5"))
		case 4:
			ac5Style = ac5Style.Bold(true).Foreground(lipgloss.Color("5"))
		}
	}
	
	// Define label settings
	labelWidth := 12
	
	// Title field
	b.WriteString(titleStyle.Width(labelWidth).Render("Title"))
	b.WriteString(f.titleInput.View() + "\n\n")
	
	// Description field
	b.WriteString(descStyle.Width(labelWidth).Render("Description"))
	b.WriteString(f.descInput.View() + "\n\n")
	
	// User Story fields
	headerStyle := lipgloss.NewStyle().Bold(true).AlignHorizontal(lipgloss.Left)
	b.WriteString(headerStyle.Render("User Story") + "\n")
	b.WriteString(asStyle.Width(labelWidth).Render("As a"))
	b.WriteString(f.asInput.View() + "\n")
	
	b.WriteString(wantStyle.Width(labelWidth).Render("I want"))
	b.WriteString(f.wantInput.View() + "\n")
	
	b.WriteString(soThatStyle.Width(labelWidth).Render("So that"))
	b.WriteString(f.soThatInput.View() + "\n\n")
	
	// Acceptance Criteria fields
	b.WriteString(headerStyle.Render("Acceptance Criteria") + "\n")
	
	b.WriteString(ac1Style.Width(labelWidth).Render("1."))
	b.WriteString(f.acInputs[0].View() + "\n")
	
	b.WriteString(ac2Style.Width(labelWidth).Render("2."))
	b.WriteString(f.acInputs[1].View() + "\n")
	
	b.WriteString(ac3Style.Width(labelWidth).Render("3."))
	b.WriteString(f.acInputs[2].View() + "\n")
	
	b.WriteString(ac4Style.Width(labelWidth).Render("4."))
	b.WriteString(f.acInputs[3].View() + "\n")
	
	b.WriteString(ac5Style.Width(labelWidth).Render("5."))
	b.WriteString(f.acInputs[4].View() + "\n")

	// Help text
	b.WriteString("\n" + lipgloss.NewStyle().Faint(true).Render("Tab: Next • Shift+Tab: Previous • Enter: Next • Ctrl+C: Quit"))

	return b.String()
}

// nextField moves to the next field
func (f *UserStoryForm) nextField() tea.Cmd {
	switch f.activeField {
	case USTitleField:
		f.titleInput.Blur()
		f.activeField = USDescriptionField
		f.descInput.Focus()
	case USDescriptionField:
		f.descInput.Blur()
		f.activeField = USAsField
		f.asInput.Focus()
	case USAsField:
		f.asInput.Blur()
		f.activeField = USWantField
		f.wantInput.Focus()
	case USWantField:
		f.wantInput.Blur()
		f.activeField = USSoThatField
		f.soThatInput.Focus()
	case USSoThatField:
		f.soThatInput.Blur()
		f.activeField = USAcceptanceCriteriaField
		f.activeACIndex = 0
		f.acInputs[0].Focus()
	case USAcceptanceCriteriaField:
		f.acInputs[f.activeACIndex].Blur()
		if f.activeACIndex < len(f.acInputs)-1 {
			f.activeACIndex++
			f.acInputs[f.activeACIndex].Focus()
		} else {
			// Only auto-submit if there's content
			if f.hasContent() {
				f.ConfirmSubmission = true
				return tea.Quit
			}
			// If no content, quit without confirming submission
			return tea.Quit
		}
	}
	return nil
}

// prevField moves to the previous field
func (f *UserStoryForm) prevField() {
	switch f.activeField {
	case USDescriptionField:
		f.descInput.Blur()
		f.activeField = USTitleField
		f.titleInput.Focus()
	case USAsField:
		f.asInput.Blur()
		f.activeField = USDescriptionField
		f.descInput.Focus()
	case USWantField:
		f.wantInput.Blur()
		f.activeField = USAsField
		f.asInput.Focus()
	case USSoThatField:
		f.soThatInput.Blur()
		f.activeField = USWantField
		f.wantInput.Focus()
	case USAcceptanceCriteriaField:
		f.acInputs[f.activeACIndex].Blur()
		if f.activeACIndex > 0 {
			f.activeACIndex--
			f.acInputs[f.activeACIndex].Focus()
		} else {
			f.activeField = USSoThatField
			f.soThatInput.Focus()
		}
	}
}

// hasContent checks if any field has content
func (f *UserStoryForm) hasContent() bool {
	if f.titleInput.Value() != "" ||
		f.descInput.Value() != "" ||
		f.asInput.Value() != "" ||
		f.wantInput.Value() != "" ||
		f.soThatInput.Value() != "" {
		return true
	}

	// Check acceptance criteria
	for _, input := range f.acInputs {
		if input.Value() != "" {
			return true
		}
	}

	return false
}

// GetTitle returns the current title value
func (f *UserStoryForm) GetTitle() string {
	return f.titleInput.Value()
}

// SetFilePath sets the file path in the user story
func (f *UserStoryForm) SetFilePath(path string) {
	f.us.FilePath = path
}

// GetUserStory returns the final user story
func (f *UserStoryForm) GetUserStory() models.UserStory {
	us := f.us
	us.Title = f.titleInput.Value()

	// Build the content without metadata
	var contentWithoutMetadata strings.Builder

	// Add title
	contentWithoutMetadata.WriteString(fmt.Sprintf("# %s\n", us.Title))

	// Add description
	if desc := f.descInput.Value(); desc != "" {
		contentWithoutMetadata.WriteString(desc + "\n\n")
	}

	// Add user story
	contentWithoutMetadata.WriteString(fmt.Sprintf("As a %s\nI want %s\nso that %s\n\n",
		f.asInput.Value(),
		f.wantInput.Value(),
		f.soThatInput.Value()))

	// Add acceptance criteria
	contentWithoutMetadata.WriteString("## Acceptance criteria\n")
	for _, input := range f.acInputs {
		if value := input.Value(); value != "" {
			contentWithoutMetadata.WriteString(fmt.Sprintf("- %s\n", value))
		}
	}

	// Calculate content hash from content without metadata
	var contentHash string
	if f.hasContent() {
		contentHash = models.GenerateContentHash(contentWithoutMetadata.String())
	} else {
		contentHash = "d41d8cd98f00b204e9800998ecf8427e" // MD5 hash of empty string
	}

	// Build final content with metadata and hash
	var finalContent strings.Builder
	finalContent.WriteString("---\n")
	finalContent.WriteString(fmt.Sprintf("file_path: %s\n", us.FilePath))
	finalContent.WriteString(fmt.Sprintf("created_at: %s\n", us.CreatedAt.Format("2006-01-02T15:04:05Z07:00")))
	finalContent.WriteString(fmt.Sprintf("last_updated: %s\n", us.LastUpdated.Format("2006-01-02T15:04:05Z07:00")))
	finalContent.WriteString(fmt.Sprintf("_content_hash: %s\n", contentHash))
	finalContent.WriteString("---\n\n")
	finalContent.WriteString(contentWithoutMetadata.String())

	us.Content = finalContent.String()
	us.ContentHash = contentHash

	return us
} 
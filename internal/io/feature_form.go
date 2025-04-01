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
type FieldType int

const (
	TitleField FieldType = iota
	DescriptionField
	UserStoryAsField
	UserStoryWantField
	UserStorySoThatField
	AcceptanceCriteria1Field
	AcceptanceCriteria2Field
	AcceptanceCriteria3Field
	AcceptanceCriteria4Field
	AcceptanceCriteria5Field
	ReviewField
)

// FeatureForm is a tea.Model for the feature request form
type FeatureForm struct {
	fr                   models.FeatureRequest
	titleInput           textinput.Model
	descInput            textinput.Model
	userStoryAsInput     textinput.Model
	userStoryWantInput   textinput.Model
	userStorySoThatInput textinput.Model
	acInputs             []textinput.Model
	activeField          FieldType
	activeACIndex        int
	reviewMode           bool
	ConfirmSubmission    bool // User confirmed submission
	editMode             bool
	cancel               bool
	focused              bool
	width                int
	height               int
	err                  error
}

// NewFeatureForm creates a new feature request form
func NewFeatureForm(fr models.FeatureRequest) *FeatureForm {
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter title"
	titleInput.Focus()
	titleInput.Width = 80
	titleInput.CharLimit = 100
	titleInput.SetValue(fr.Title)

	descInput := textinput.New()
	descInput.Placeholder = "Enter description"
	descInput.Width = 80
	descInput.CharLimit = 200
	descInput.SetValue(fr.Description)

	// Parse existing user story if available
	userStoryAs := ""
	userStoryWant := ""
	userStorySoThat := ""

	if fr.UserStory != "" {
		parts := strings.Split(fr.UserStory, " I want ")
		if len(parts) > 1 {
			userStoryAs = strings.TrimPrefix(parts[0], "As a ")
			remainingParts := strings.Split(parts[1], " so that ")
			if len(remainingParts) > 1 {
				userStoryWant = remainingParts[0]
				userStorySoThat = remainingParts[1]
			} else {
				userStoryWant = parts[1]
			}
		} else {
			userStoryAs = strings.TrimPrefix(fr.UserStory, "As a ")
		}
	}

	userStoryAsInput := textinput.New()
	userStoryAsInput.Placeholder = " Enter user type"
	userStoryAsInput.Width = 80
	userStoryAsInput.CharLimit = 100
	userStoryAsInput.SetValue(userStoryAs)

	userStoryWantInput := textinput.New()
	userStoryWantInput.Placeholder = "Enter desired capability"
	userStoryWantInput.Width = 80
	userStoryWantInput.CharLimit = 100
	userStoryWantInput.SetValue(userStoryWant)

	userStorySoThatInput := textinput.New()
	userStorySoThatInput.Placeholder = "Enter benefit"
	userStorySoThatInput.Width = 80
	userStorySoThatInput.CharLimit = 100
	userStorySoThatInput.SetValue(userStorySoThat)

	// Create 5 acceptance criteria inputs
	acInputs := make([]textinput.Model, 5)
	for i := 0; i < 5; i++ {
		acInputs[i] = textinput.New()
		acInputs[i].Placeholder = fmt.Sprintf("Enter acceptance criteria %d", i+1)
		acInputs[i].Width = 80
		acInputs[i].CharLimit = 200

		// Set values from existing AC if available
		if i < len(fr.AcceptanceCriteria) {
			acInputs[i].SetValue(fr.AcceptanceCriteria[i])
		}
	}

	form := &FeatureForm{
		fr:                   fr,
		titleInput:           titleInput,
		descInput:            descInput,
		userStoryAsInput:     userStoryAsInput,
		userStoryWantInput:   userStoryWantInput,
		userStorySoThatInput: userStorySoThatInput,
		acInputs:             acInputs,
		activeField:          TitleField,
		activeACIndex:        0,
		reviewMode:           false,
		ConfirmSubmission:    false,
		editMode:             false,
		cancel:               false,
		focused:              true,
		width:                80,
		height:               24,
	}

	return form
}

// Init initializes the form
func (f *FeatureForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles user input events
func (f *FeatureForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			f.cancel = true
			return f, tea.Quit

		case tea.KeyTab:
			// Move to the next field
			if !f.reviewMode {
				f.nextField()
			}

		case tea.KeyShiftTab:
			// Move to the previous field
			if !f.reviewMode {
				f.prevField()
			}

		case tea.KeyEnter:
			if !f.reviewMode {
				// Move to the next field
				f.nextField()
			} else {
				// In review mode, Enter confirms submission (Y is default)
				f.ConfirmSubmission = true
				return f, tea.Quit
			}

		case tea.KeyEsc:
			if f.reviewMode {
				// Exit review mode and go back to editing
				f.reviewMode = false
				f.editMode = true
				f.activeField = TitleField
				f.titleInput.Focus()
			}

		default:
			// Handle other keys based on active field
			if !f.reviewMode {
				switch f.activeField {
				case TitleField:
					f.titleInput, cmd = f.titleInput.Update(msg)
					cmds = append(cmds, cmd)
				case DescriptionField:
					f.descInput, cmd = f.descInput.Update(msg)
					cmds = append(cmds, cmd)
				case UserStoryAsField:
					f.userStoryAsInput, cmd = f.userStoryAsInput.Update(msg)
					cmds = append(cmds, cmd)
				case UserStoryWantField:
					f.userStoryWantInput, cmd = f.userStoryWantInput.Update(msg)
					cmds = append(cmds, cmd)
				case UserStorySoThatField:
					f.userStorySoThatInput, cmd = f.userStorySoThatInput.Update(msg)
					cmds = append(cmds, cmd)
				case AcceptanceCriteria1Field:
					f.acInputs[0], cmd = f.acInputs[0].Update(msg)
					cmds = append(cmds, cmd)
				case AcceptanceCriteria2Field:
					f.acInputs[1], cmd = f.acInputs[1].Update(msg)
					cmds = append(cmds, cmd)
				case AcceptanceCriteria3Field:
					f.acInputs[2], cmd = f.acInputs[2].Update(msg)
					cmds = append(cmds, cmd)
				case AcceptanceCriteria4Field:
					f.acInputs[3], cmd = f.acInputs[3].Update(msg)
					cmds = append(cmds, cmd)
				case AcceptanceCriteria5Field:
					f.acInputs[4], cmd = f.acInputs[4].Update(msg)
					cmds = append(cmds, cmd)
				}
			} else {
				// In review mode
				switch msg.String() {
				case "y", "Y", "":
					f.ConfirmSubmission = true
					return f, tea.Quit
				case "n", "N":
					f.reviewMode = false
					f.editMode = true
					f.activeField = TitleField
					f.titleInput.Focus()
				}
			}
		}

	case tea.WindowSizeMsg:
		f.width = msg.Width
		f.height = msg.Height
	}

	// Update feature request from form fields
	f.updateFeatureRequest()

	// Return command
	return f, tea.Batch(cmds...)
}

// View renders the form UI
func (f *FeatureForm) View() string {
	var b strings.Builder

	if f.reviewMode {
		// If all acceptance criteria are empty and we're in review mode, go directly to confirmation
		allEmpty := true
		for _, ac := range f.fr.AcceptanceCriteria {
			if ac != "" {
				allEmpty = false
				break
			}
		}

		if allEmpty {
			return f.renderConfirmationOnly()
		}

		return f.renderReviewMode()
	}

	// Form title - aligned to the left with no extra spaces
	formTitleStyle := lipgloss.NewStyle().Bold(true).AlignHorizontal(lipgloss.Left)
	b.WriteString(formTitleStyle.Render("Feature Request Form") + "\n\n")

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
	case TitleField:
		titleStyle = titleStyle.Bold(true).Foreground(lipgloss.Color("12"))
	case DescriptionField:
		descStyle = descStyle.Bold(true).Foreground(lipgloss.Color("12"))
	case UserStoryAsField:
		asStyle = asStyle.Bold(true).Foreground(lipgloss.Color("12"))
	case UserStoryWantField:
		wantStyle = wantStyle.Bold(true).Foreground(lipgloss.Color("12"))
	case UserStorySoThatField:
		soThatStyle = soThatStyle.Bold(true).Foreground(lipgloss.Color("12"))
	case AcceptanceCriteria1Field:
		ac1Style = ac1Style.Bold(true).Foreground(lipgloss.Color("12"))
	case AcceptanceCriteria2Field:
		ac2Style = ac2Style.Bold(true).Foreground(lipgloss.Color("12"))
	case AcceptanceCriteria3Field:
		ac3Style = ac3Style.Bold(true).Foreground(lipgloss.Color("12"))
	case AcceptanceCriteria4Field:
		ac4Style = ac4Style.Bold(true).Foreground(lipgloss.Color("12"))
	case AcceptanceCriteria5Field:
		ac5Style = ac5Style.Bold(true).Foreground(lipgloss.Color("12"))
	}

	// Define label settings
	labelWidth := 12

	// Title field
	b.WriteString(titleStyle.Width(labelWidth).Render("Title:"))
	b.WriteString(" " + f.titleInput.View() + "\n")

	// Description field
	b.WriteString(descStyle.Width(labelWidth).Render("Description:"))
	b.WriteString(" " + f.descInput.View() + "\n")

	// User Story fields
	headerStyle := lipgloss.NewStyle().Bold(true).AlignHorizontal(lipgloss.Left)
	b.WriteString(headerStyle.Render("User Story") + "\n")
	b.WriteString(asStyle.Width(labelWidth).Render("As a:"))
	b.WriteString(" " + f.userStoryAsInput.View() + "\n")

	b.WriteString(wantStyle.Width(labelWidth).Render("I want:"))
	b.WriteString(" " + f.userStoryWantInput.View() + "\n")

	b.WriteString(soThatStyle.Width(labelWidth).Render("So that:"))
	b.WriteString(" " + f.userStorySoThatInput.View() + "\n")

	// Acceptance Criteria fields
	b.WriteString(headerStyle.Render("Acceptance Criteria") + "\n")

	b.WriteString(ac1Style.Width(labelWidth).Render("1:"))
	b.WriteString(" " + f.acInputs[0].View() + "\n")

	b.WriteString(ac2Style.Width(labelWidth).Render("2:"))
	b.WriteString(" " + f.acInputs[1].View() + "\n")

	b.WriteString(ac3Style.Width(labelWidth).Render("3:"))
	b.WriteString(" " + f.acInputs[2].View() + "\n")

	b.WriteString(ac4Style.Width(labelWidth).Render("4:"))
	b.WriteString(" " + f.acInputs[3].View() + "\n")

	b.WriteString(ac5Style.Width(labelWidth).Render("5:"))
	b.WriteString(" " + f.acInputs[4].View() + "\n\n")

	// Navigation help
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).AlignHorizontal(lipgloss.Left)
	b.WriteString(helpStyle.Render(
		"Tab: next field, Shift+Tab: previous field, Enter: confirm field\n" +
			"Press Tab after filling all fields to submit\n" +
			"Press Ctrl+C to cancel and save as draft\n"))

	return b.String()
}

// renderConfirmationOnly renders just the confirmation prompt without summary
func (f *FeatureForm) renderConfirmationOnly() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Submit Feature Request?\n"))
	b.WriteString("Press Enter to confirm (Y) or N to go back to editing\n")

	return b.String()
}

// renderReviewMode renders the review mode view
func (f *FeatureForm) renderReviewMode() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Review Feature Request\n"))

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Title: "))
	b.WriteString(f.fr.Title + "\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Description: "))
	b.WriteString(f.fr.Description + "\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("User Story:\n"))
	b.WriteString(f.fr.UserStory + "\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Acceptance Criteria:\n"))
	for i, criteria := range f.fr.AcceptanceCriteria {
		if criteria != "" {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, criteria))
		}
	}

	b.WriteString("\nSubmit this feature request? [Y/n]\n")
	b.WriteString("Press Enter to confirm or Esc to go back to editing\n")

	return b.String()
}

// RenderThankYouMessage returns a warm thank you message after submission
func (f *FeatureForm) RenderThankYouMessage() string {
	var b strings.Builder

	// Add a decorative element
	thanksStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")). // Green color
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("10")).
		Padding(1, 2).
		Align(lipgloss.Center)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // White color
		Width(60).
		Align(lipgloss.Center)

	b.WriteString("\n\n")
	b.WriteString(thanksStyle.Render("Feature Request Submitted!"))
	b.WriteString("\n\n")

	message := "Thank you for taking the time to submit a feature request! " +
		"Your feedback is incredibly valuable and helps make this tool better for everyone. " +
		"I'll review your request carefully and do my best to implement it soon."

	b.WriteString(messageStyle.Render(message))
	b.WriteString("\n\n")

	return b.String()
}

// nextField advances to the next field
func (f *FeatureForm) nextField() {
	// Update feature request from current field
	f.updateFeatureRequest()

	// Unfocus current field
	switch f.activeField {
	case TitleField:
		f.titleInput.Blur()
	case DescriptionField:
		f.descInput.Blur()
	case UserStoryAsField:
		f.userStoryAsInput.Blur()
	case UserStoryWantField:
		f.userStoryWantInput.Blur()
	case UserStorySoThatField:
		f.userStorySoThatInput.Blur()
	case AcceptanceCriteria1Field:
		f.acInputs[0].Blur()
	case AcceptanceCriteria2Field:
		f.acInputs[1].Blur()
	case AcceptanceCriteria3Field:
		f.acInputs[2].Blur()
	case AcceptanceCriteria4Field:
		f.acInputs[3].Blur()
	case AcceptanceCriteria5Field:
		f.acInputs[4].Blur()
	}

	// Move to next field
	switch f.activeField {
	case TitleField:
		f.activeField = DescriptionField
		f.descInput.Focus()
	case DescriptionField:
		f.activeField = UserStoryAsField
		f.userStoryAsInput.Focus()
	case UserStoryAsField:
		f.activeField = UserStoryWantField
		f.userStoryWantInput.Focus()
	case UserStoryWantField:
		f.activeField = UserStorySoThatField
		f.userStorySoThatInput.Focus()
	case UserStorySoThatField:
		f.activeField = AcceptanceCriteria1Field
		f.acInputs[0].Focus()
	case AcceptanceCriteria1Field:
		f.activeField = AcceptanceCriteria2Field
		f.acInputs[1].Focus()
	case AcceptanceCriteria2Field:
		f.activeField = AcceptanceCriteria3Field
		f.acInputs[2].Focus()
	case AcceptanceCriteria3Field:
		f.activeField = AcceptanceCriteria4Field
		f.acInputs[3].Focus()
	case AcceptanceCriteria4Field:
		f.activeField = AcceptanceCriteria5Field
		f.acInputs[4].Focus()
	case AcceptanceCriteria5Field:
		// Move to review mode when all fields are complete
		f.activeField = ReviewField
		f.reviewMode = true
	}
}

// prevField goes back to the previous field
func (f *FeatureForm) prevField() {
	// Update feature request from current field
	f.updateFeatureRequest()

	// Unfocus current field
	switch f.activeField {
	case TitleField:
		f.titleInput.Blur()
	case DescriptionField:
		f.descInput.Blur()
	case UserStoryAsField:
		f.userStoryAsInput.Blur()
	case UserStoryWantField:
		f.userStoryWantInput.Blur()
	case UserStorySoThatField:
		f.userStorySoThatInput.Blur()
	case AcceptanceCriteria1Field:
		f.acInputs[0].Blur()
	case AcceptanceCriteria2Field:
		f.acInputs[1].Blur()
	case AcceptanceCriteria3Field:
		f.acInputs[2].Blur()
	case AcceptanceCriteria4Field:
		f.acInputs[3].Blur()
	case AcceptanceCriteria5Field:
		f.acInputs[4].Blur()
	}

	// Move to previous field
	switch f.activeField {
	case DescriptionField:
		f.activeField = TitleField
		f.titleInput.Focus()
	case UserStoryAsField:
		f.activeField = DescriptionField
		f.descInput.Focus()
	case UserStoryWantField:
		f.activeField = UserStoryAsField
		f.userStoryAsInput.Focus()
	case UserStorySoThatField:
		f.activeField = UserStoryWantField
		f.userStoryWantInput.Focus()
	case AcceptanceCriteria1Field:
		f.activeField = UserStorySoThatField
		f.userStorySoThatInput.Focus()
	case AcceptanceCriteria2Field:
		f.activeField = AcceptanceCriteria1Field
		f.acInputs[0].Focus()
	case AcceptanceCriteria3Field:
		f.activeField = AcceptanceCriteria2Field
		f.acInputs[1].Focus()
	case AcceptanceCriteria4Field:
		f.activeField = AcceptanceCriteria3Field
		f.acInputs[2].Focus()
	case AcceptanceCriteria5Field:
		f.activeField = AcceptanceCriteria4Field
		f.acInputs[3].Focus()
	case ReviewField:
		f.activeField = AcceptanceCriteria5Field
		f.acInputs[4].Focus()
		f.reviewMode = false
	}
}

// updateFeatureRequest updates the feature request from form fields
func (f *FeatureForm) updateFeatureRequest() {
	f.fr.Title = f.titleInput.Value()
	f.fr.Description = f.descInput.Value()

	// Combine user story parts
	asValue := strings.TrimSpace(f.userStoryAsInput.Value())
	wantValue := strings.TrimSpace(f.userStoryWantInput.Value())
	soThatValue := strings.TrimSpace(f.userStorySoThatInput.Value())

	userStory := ""
	if asValue != "" {
		userStory = "As a " + asValue
		if wantValue != "" {
			userStory += " I want " + wantValue
			if soThatValue != "" {
				userStory += " so that " + soThatValue
			}
		}
	}

	f.fr.UserStory = userStory

	// For backwards compatibility, store the combined user story in the importance field
	f.fr.Importance = userStory

	// Collect non-empty acceptance criteria
	var criteria []string
	for _, input := range f.acInputs {
		value := strings.TrimSpace(input.Value())
		if value != "" {
			criteria = append(criteria, value)
		}
	}
	f.fr.AcceptanceCriteria = criteria
}

// SaveDraft returns the current state of the feature request
func (f *FeatureForm) SaveDraft() models.FeatureRequest {
	f.updateFeatureRequest()
	return f.fr
}

// GetFeatureRequest returns the completed feature request
func (f *FeatureForm) GetFeatureRequest() models.FeatureRequest {
	return f.fr
}

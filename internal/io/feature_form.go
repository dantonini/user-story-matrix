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
	ImportanceField
	UserStoryField
	AcceptanceCriteria1Field
	AcceptanceCriteria2Field
	AcceptanceCriteria3Field
	ReviewField
)

// FeatureForm is a tea.Model for the feature request form
type FeatureForm struct {
	fr                  models.FeatureRequest
	titleInput          textinput.Model
	descInput           textinput.Model
	importanceInput     textinput.Model
	userStoryInput      textinput.Model
	acInputs            []textinput.Model
	activeField         FieldType
	activeACIndex       int
	reviewMode          bool
	ConfirmSubmission   bool // User confirmed submission
	editMode            bool
	cancel              bool
	focused             bool
	width               int
	height              int
	err                 error
}

// NewFeatureForm creates a new feature request form
func NewFeatureForm(fr models.FeatureRequest) *FeatureForm {
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter the title of the feature"
	titleInput.Focus()
	titleInput.Width = 80
	titleInput.CharLimit = 100
	titleInput.SetValue(fr.Title)

	descInput := textinput.New()
	descInput.Placeholder = "Enter a detailed description of the feature"
	descInput.Width = 80
	descInput.CharLimit = 200
	descInput.SetValue(fr.Description)

	importanceInput := textinput.New()
	importanceInput.Placeholder = "Explain why this feature is important to you"
	importanceInput.Width = 80
	importanceInput.CharLimit = 200
	importanceInput.SetValue(fr.Importance)

	userStoryInput := textinput.New()
	userStoryInput.Placeholder = "Format: As a ... I want ... so that ..."
	userStoryInput.Width = 80
	userStoryInput.CharLimit = 200
	userStoryInput.SetValue(fr.UserStory)

	// Create 3 acceptance criteria inputs
	acInputs := make([]textinput.Model, 3)
	for i := 0; i < 3; i++ {
		acInputs[i] = textinput.New()
		acInputs[i].Placeholder = fmt.Sprintf("Acceptance criteria %d", i+1)
		acInputs[i].Width = 80
		acInputs[i].CharLimit = 200
		
		// Set values from existing AC if available
		if i < len(fr.AcceptanceCriteria) {
			acInputs[i].SetValue(fr.AcceptanceCriteria[i])
		}
	}

	form := &FeatureForm{
		fr:                fr,
		titleInput:        titleInput,
		descInput:         descInput,
		importanceInput:   importanceInput,
		userStoryInput:    userStoryInput,
		acInputs:          acInputs,
		activeField:       TitleField,
		activeACIndex:     0,
		reviewMode:        false,
		ConfirmSubmission: false,
		editMode:          false,
		cancel:            false,
		focused:           true,
		width:             80,
		height:            24,
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
				case ImportanceField:
					f.importanceInput, cmd = f.importanceInput.Update(msg)
					cmds = append(cmds, cmd)
				case UserStoryField:
					f.userStoryInput, cmd = f.userStoryInput.Update(msg)
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

	// Title
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Feature Request Form\n\n"))

	// Show all fields
	// Highlight the active field with different styling
	titleStyle := lipgloss.NewStyle()
	descStyle := lipgloss.NewStyle()
	importanceStyle := lipgloss.NewStyle()
	userStoryStyle := lipgloss.NewStyle()
	ac1Style := lipgloss.NewStyle()
	ac2Style := lipgloss.NewStyle()
	ac3Style := lipgloss.NewStyle()
	
	switch f.activeField {
	case TitleField:
		titleStyle = titleStyle.Bold(true).Foreground(lipgloss.Color("12"))
	case DescriptionField:
		descStyle = descStyle.Bold(true).Foreground(lipgloss.Color("12"))
	case ImportanceField:
		importanceStyle = importanceStyle.Bold(true).Foreground(lipgloss.Color("12"))
	case UserStoryField:
		userStoryStyle = userStoryStyle.Bold(true).Foreground(lipgloss.Color("12"))
	case AcceptanceCriteria1Field:
		ac1Style = ac1Style.Bold(true).Foreground(lipgloss.Color("12"))
	case AcceptanceCriteria2Field:
		ac2Style = ac2Style.Bold(true).Foreground(lipgloss.Color("12"))
	case AcceptanceCriteria3Field:
		ac3Style = ac3Style.Bold(true).Foreground(lipgloss.Color("12"))
	}
	
	// Title field
	b.WriteString(titleStyle.Render("Title") + " (required):\n")
	b.WriteString(f.titleInput.View() + "\n\n")
	
	// Description field
	b.WriteString(descStyle.Render("Description") + " (required):\n")
	b.WriteString(f.descInput.View() + "\n\n")
	
	// Importance field
	b.WriteString(importanceStyle.Render("Why it is important") + " (required):\n")
	b.WriteString(f.importanceInput.View() + "\n\n")
	
	// User Story field
	b.WriteString(userStoryStyle.Render("User Story") + " (required):\n")
	b.WriteString("Format: As a ... I want ... so that ...\n")
	b.WriteString(f.userStoryInput.View() + "\n\n")
	
	// Acceptance Criteria fields
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Acceptance Criteria") + " (at least one required):\n")
	b.WriteString(ac1Style.Render("Acceptance criteria 1") + ":\n")
	b.WriteString(f.acInputs[0].View() + "\n\n")
	
	b.WriteString(ac2Style.Render("Acceptance criteria 2") + ":\n")
	b.WriteString(f.acInputs[1].View() + "\n\n")
	
	b.WriteString(ac3Style.Render("Acceptance criteria 3") + ":\n")
	b.WriteString(f.acInputs[2].View() + "\n\n")

	// Navigation help
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
		"Tab: next field, Shift+Tab: previous field, Enter: confirm field\n" +
		"Press Tab after filling all fields to submit\n" +
		"Press Ctrl+C to cancel and save as draft\n"))

	return b.String()
}

// renderConfirmationOnly renders just the confirmation prompt without summary
func (f *FeatureForm) renderConfirmationOnly() string {
	var b strings.Builder
	
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Submit Feature Request?\n\n"))
	b.WriteString("Press Enter to confirm (Y) or N to go back to editing\n")
	
	return b.String()
}

// renderReviewMode renders the review mode view
func (f *FeatureForm) renderReviewMode() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Review Feature Request\n\n"))

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Title: "))
	b.WriteString(f.fr.Title + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Description:\n"))
	b.WriteString(f.fr.Description + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Why it is important:\n"))
	b.WriteString(f.fr.Importance + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("User Story:\n"))
	b.WriteString(f.fr.UserStory + "\n\n")

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
	case ImportanceField:
		f.importanceInput.Blur()
	case UserStoryField:
		f.userStoryInput.Blur()
	case AcceptanceCriteria1Field:
		f.acInputs[0].Blur()
	case AcceptanceCriteria2Field:
		f.acInputs[1].Blur()
	case AcceptanceCriteria3Field:
		f.acInputs[2].Blur()
	}

	// Move to next field
	switch f.activeField {
	case TitleField:
		f.activeField = DescriptionField
		f.descInput.Focus()
	case DescriptionField:
		f.activeField = ImportanceField
		f.importanceInput.Focus()
	case ImportanceField:
		f.activeField = UserStoryField
		f.userStoryInput.Focus()
	case UserStoryField:
		f.activeField = AcceptanceCriteria1Field
		f.acInputs[0].Focus()
	case AcceptanceCriteria1Field:
		f.activeField = AcceptanceCriteria2Field
		f.acInputs[1].Focus()
	case AcceptanceCriteria2Field:
		f.activeField = AcceptanceCriteria3Field
		f.acInputs[2].Focus()
	case AcceptanceCriteria3Field:
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
	case ImportanceField:
		f.importanceInput.Blur()
	case UserStoryField:
		f.userStoryInput.Blur()
	case AcceptanceCriteria1Field:
		f.acInputs[0].Blur()
	case AcceptanceCriteria2Field:
		f.acInputs[1].Blur()
	case AcceptanceCriteria3Field:
		f.acInputs[2].Blur()
	}

	// Move to previous field
	switch f.activeField {
	case DescriptionField:
		f.activeField = TitleField
		f.titleInput.Focus()
	case ImportanceField:
		f.activeField = DescriptionField
		f.descInput.Focus()
	case UserStoryField:
		f.activeField = ImportanceField
		f.importanceInput.Focus()
	case AcceptanceCriteria1Field:
		f.activeField = UserStoryField
		f.userStoryInput.Focus()
	case AcceptanceCriteria2Field:
		f.activeField = AcceptanceCriteria1Field
		f.acInputs[0].Focus()
	case AcceptanceCriteria3Field:
		f.activeField = AcceptanceCriteria2Field
		f.acInputs[1].Focus()
	case ReviewField:
		f.activeField = AcceptanceCriteria3Field
		f.acInputs[2].Focus()
		f.reviewMode = false
	}
}

// updateFeatureRequest updates the feature request from form fields
func (f *FeatureForm) updateFeatureRequest() {
	f.fr.Title = f.titleInput.Value()
	f.fr.Description = f.descInput.Value()
	f.fr.Importance = f.importanceInput.Value()
	f.fr.UserStory = f.userStoryInput.Value()

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
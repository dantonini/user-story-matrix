package io

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
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
	AcceptanceCriteriaField
	ReviewField
)

// FeatureForm is a tea.Model for the feature request form
type FeatureForm struct {
	fr                models.FeatureRequest
	titleInput        textinput.Model
	descInput         textarea.Model
	importanceInput   textarea.Model
	userStoryInput    textarea.Model
	acInput           textarea.Model
	acItems           []string
	activeField       FieldType
	reviewMode        bool
	confirmSubmission bool
	editMode          bool
	cancel            bool
	focused           bool
	width             int
	height            int
	err               error
}

// NewFeatureForm creates a new feature request form
func NewFeatureForm(fr models.FeatureRequest) *FeatureForm {
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter the title of the feature"
	titleInput.Focus()
	titleInput.Width = 80
	titleInput.CharLimit = 100
	titleInput.SetValue(fr.Title)

	descInput := textarea.New()
	descInput.Placeholder = "Enter a detailed description of the feature"
	descInput.SetValue(fr.Description)
	descInput.SetWidth(80)
	descInput.SetHeight(10)

	importanceInput := textarea.New()
	importanceInput.Placeholder = "Explain why this feature is important to you"
	importanceInput.SetValue(fr.Importance)
	importanceInput.SetWidth(80)
	importanceInput.SetHeight(5)

	userStoryInput := textarea.New()
	userStoryInput.Placeholder = "Format: As a ... I want ... so that ..."
	userStoryInput.SetValue(fr.UserStory)
	userStoryInput.SetWidth(80)
	userStoryInput.SetHeight(5)

	acInput := textarea.New()
	acInput.Placeholder = "Enter acceptance criteria (one per line)"
	acInput.SetWidth(80)
	acInput.SetHeight(5)
	
	// Join acceptance criteria for display
	if len(fr.AcceptanceCriteria) > 0 {
		acInput.SetValue(strings.Join(fr.AcceptanceCriteria, "\n"))
	}

	form := &FeatureForm{
		fr:                fr,
		titleInput:        titleInput,
		descInput:         descInput,
		importanceInput:   importanceInput,
		userStoryInput:    userStoryInput,
		acInput:           acInput,
		acItems:           fr.AcceptanceCriteria,
		activeField:       TitleField,
		reviewMode:        false,
		confirmSubmission: false,
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
				if msg.Alt {
					// Alt+Enter adds a newline in textareas
					switch f.activeField {
					case DescriptionField:
						f.descInput, cmd = f.descInput.Update(msg)
						cmds = append(cmds, cmd)
					case ImportanceField:
						f.importanceInput, cmd = f.importanceInput.Update(msg)
						cmds = append(cmds, cmd)
					case UserStoryField:
						f.userStoryInput, cmd = f.userStoryInput.Update(msg)
						cmds = append(cmds, cmd)
					case AcceptanceCriteriaField:
						f.acInput, cmd = f.acInput.Update(msg)
						cmds = append(cmds, cmd)
					}
				} else {
					// Move to the next field
					f.nextField()
				}
			} else {
				// In review mode, Enter confirms submission
				f.confirmSubmission = true
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
				case AcceptanceCriteriaField:
					f.acInput, cmd = f.acInput.Update(msg)
					cmds = append(cmds, cmd)
				}
			} else {
				// In review mode
				switch msg.String() {
				case "y", "Y":
					f.confirmSubmission = true
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
		return f.renderReviewMode()
	}

	// Title
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Feature Request Form\n\n"))

	// Render active field with label and input
	switch f.activeField {
	case TitleField:
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Title") + " (required)\n")
		b.WriteString(f.titleInput.View() + "\n\n")
		b.WriteString("Press Tab to move to the next field\n")
	case DescriptionField:
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Description") + " (required)\n")
		b.WriteString(f.descInput.View() + "\n\n")
		b.WriteString("Press Tab to move to the next field, Alt+Enter for new line\n")
	case ImportanceField:
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Importance") + " (required)\n")
		b.WriteString("Explain why this feature is important to you\n")
		b.WriteString(f.importanceInput.View() + "\n\n")
		b.WriteString("Press Tab to move to the next field, Alt+Enter for new line\n")
	case UserStoryField:
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("User Story") + " (required)\n")
		b.WriteString("Format: As a ... I want ... so that ...\n")
		b.WriteString(f.userStoryInput.View() + "\n\n")
		b.WriteString("Press Tab to move to the next field, Alt+Enter for new line\n")
	case AcceptanceCriteriaField:
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Acceptance Criteria") + " (required)\n")
		b.WriteString("Enter one criterion per line\n")
		b.WriteString(f.acInput.View() + "\n\n")
		b.WriteString("Press Tab when done, Alt+Enter for new line\n")
	case ReviewField:
		// This case shouldn't be reached as review mode is handled separately
		f.reviewMode = true
		return f.renderReviewMode()
	}

	// Footer
	b.WriteString("\nPress Ctrl+C to cancel and save as draft")

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

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Importance:\n"))
	b.WriteString(f.fr.Importance + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("User Story:\n"))
	b.WriteString(f.fr.UserStory + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Acceptance Criteria:\n"))
	for i, criteria := range f.fr.AcceptanceCriteria {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, criteria))
	}

	b.WriteString("\nSubmit this feature request? [y/n]\n")
	b.WriteString("Press Esc to go back to editing\n")

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
	case AcceptanceCriteriaField:
		f.acInput.Blur()
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
		f.activeField = AcceptanceCriteriaField
		f.acInput.Focus()
	case AcceptanceCriteriaField:
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
	case AcceptanceCriteriaField:
		f.acInput.Blur()
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
	case AcceptanceCriteriaField:
		f.activeField = UserStoryField
		f.userStoryInput.Focus()
	case ReviewField:
		f.activeField = AcceptanceCriteriaField
		f.acInput.Focus()
		f.reviewMode = false
	}
}

// updateFeatureRequest updates the feature request from form fields
func (f *FeatureForm) updateFeatureRequest() {
	f.fr.Title = f.titleInput.Value()
	f.fr.Description = f.descInput.Value()
	f.fr.Importance = f.importanceInput.Value()
	f.fr.UserStory = f.userStoryInput.Value()

	// Split acceptance criteria by newlines
	acText := f.acInput.Value()
	if acText != "" {
		criteria := strings.Split(acText, "\n")
		var nonEmptyCriteria []string
		for _, c := range criteria {
			if strings.TrimSpace(c) != "" {
				nonEmptyCriteria = append(nonEmptyCriteria, c)
			}
		}
		f.fr.AcceptanceCriteria = nonEmptyCriteria
	} else {
		f.fr.AcceptanceCriteria = []string{}
	}
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
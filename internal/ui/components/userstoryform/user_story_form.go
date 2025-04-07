// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package userstoryform

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user-story-matrix/usm/internal/llm"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui/clipboard"
	"github.com/user-story-matrix/usm/internal/ui/components/spinner"
	formModels "github.com/user-story-matrix/usm/internal/ui/models"
)

// Field labels
const (
	TitleField            = "title"
	AsAField              = "as_a"
	IWantField            = "i_want"
	SoThatField           = "so_that"
	AcceptanceCriteriaField = "acceptance_criteria"
)

// FieldIndex maps field names to their index in the inputs slice
var FieldIndex = map[string]int{
	TitleField:              0,
	AsAField:                1,
	IWantField:              2,
	SoThatField:             3,
	AcceptanceCriteriaField: 4,
}

// UserStoryForm represents the form for creating a user story with LLM processing
type UserStoryForm struct {
	model             *formModels.UserStoryFormModel
	story             models.UserStory
	inputs            []textinput.Model
	criteriasInputs   []textinput.Model // Add separate inputs for acceptance criteria
	cursor            cursor.Model
	focused           int
	focusedCriteria   int // Track which criteria field has focus
	submitted         bool
	width             int
	height            int
	spinner           spinner.Model
	fieldPrevValues   map[int]string
	criteriaPrevValues map[int]string // Track previous values for criteria fields
	processor         llm.LLMProcessor
	configManager     *llm.ConfigManager
	processingCtx     context.Context
	processingCancel  context.CancelFunc
	lastTimeoutCheck  time.Time
	rawCriteriaInput  string // Used for testing to preserve exact input format
	inCriteriaSection bool   // Track if we're in the criteria section
}

// New creates a new UserStoryForm
func New(us models.UserStory, processor llm.LLMProcessor, configManager *llm.ConfigManager) *UserStoryForm {
	form := &UserStoryForm{
		story:             us,
		cursor:            cursor.New(),
		focused:           0,
		focusedCriteria:   0,
		processor:         processor,
		configManager:     configManager,
		spinner:           spinner.New(),
		fieldPrevValues:   make(map[int]string),
		criteriaPrevValues: make(map[int]string),
		lastTimeoutCheck:  time.Now(),
		inCriteriaSection: false,
	}

	// Initialize the model
	form.model = formModels.NewUserStoryFormModel(us, processor, configManager)

	// Initialize inputs
	form.inputs = make([]textinput.Model, 4) // Basic fields (not including criteria)
	
	// Initialize title input
	form.inputs[0] = textinput.New()
	form.inputs[0].Placeholder = "User Story Title"
	form.inputs[0].Focus()
	form.inputs[0].Width = 60
	form.inputs[0].Prompt = ""
	
	// Initialize as-a input
	form.inputs[1] = textinput.New()
	form.inputs[1].Placeholder = "As a..."
	form.inputs[1].Width = 60
	form.inputs[1].Prompt = ""
	
	// Initialize i-want input
	form.inputs[2] = textinput.New()
	form.inputs[2].Placeholder = "I want..."
	form.inputs[2].Width = 60
	form.inputs[2].Prompt = ""
	
	// Initialize so-that input
	form.inputs[3] = textinput.New()
	form.inputs[3].Placeholder = "So that..."
	form.inputs[3].Width = 60
	form.inputs[3].Prompt = ""
	
	// Initialize five separate criteria inputs
	form.criteriasInputs = make([]textinput.Model, 5)
	for i := 0; i < 5; i++ {
		form.criteriasInputs[i] = textinput.New()
		form.criteriasInputs[i].Placeholder = fmt.Sprintf("Enter acceptance criteria %d", i+1)
		form.criteriasInputs[i].Width = 60
		form.criteriasInputs[i].Prompt = ""
	}
	
	// Set existing criteria if any
	for i, criteria := range us.Criteria {
		if i < len(form.criteriasInputs) {
			form.criteriasInputs[i].SetValue(criteria)
		}
	}

	// Set values from user story
	if us.Title != "" {
		form.inputs[0].SetValue(us.Title)
	}
	
	// Store initial values for paste detection
	for i, input := range form.inputs {
		form.fieldPrevValues[i] = input.Value()
	}
	
	for i, input := range form.criteriasInputs {
		form.criteriaPrevValues[i] = input.Value()
	}
	
	return form
}

// Init initializes the form
func (f *UserStoryForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the form
func (f *UserStoryForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle spinner updates
	updatedSpinner, spinnerCmd := f.spinner.Update(msg)
	f.spinner = updatedSpinner
	if spinnerCmd != nil {
		cmds = append(cmds, spinnerCmd)
	}

	// Handle different message types
	switch msg := msg.(type) {
	case hideSpinnerMsg:
		// Hide the spinner
		f.spinner.SetVisible(false)
		return f, nil
		
	case tea.KeyMsg:
		// Check for paste event
		if clipboard.IsPasteEvent(msg) {
			// Get pasted content
			pastedText := clipboard.ExtractPastedText(msg)
			
			// If we can extract it directly
			if pastedText != "" && clipboard.IsLongEnoughForProcessing(pastedText) {
				f.processClipboardContent(pastedText)
				// Return here to prevent the paste from being processed by the text input
				return f, tea.Batch(cmds...)
			}
		}
		
		// Handle key messages
		switch msg.String() {
		case "ctrl+a":
			// Confirm all auto-populated fields if there are any
			if f.model != nil && f.model.HasAutoPopulatedFields() {
				f.model.ConfirmAllFields()
				// Update UI to reflect the changes
				f.updateUIFromModel()
				
				// Show a confirmation message
				confirmMsg := fmt.Sprintf("Confirmed %d auto-populated fields", f.model.GetAutoPopulatedFieldCount())
				confirmStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Italic(true)
				
				// Set a message in the spinner and show it briefly
				f.spinner.SetMessage(confirmStyle.Render(confirmMsg))
				f.spinner.SetVisible(true)
				
				// Schedule the spinner to be hidden after a short delay
				cmds = append(cmds, tea.Tick(time.Second*2, func(time.Time) tea.Msg {
					return hideSpinnerMsg{}
				}))
				
				return f, tea.Batch(cmds...)
			}
			return f, nil
			
		case "esc":
			// Cancel processing if active
			if f.model.IsProcessingActive() {
				f.model.CancelProcessing()
				if f.processingCancel != nil {
					f.processingCancel()
				}
				f.spinner.SetVisible(false)
				return f, nil
			}
			
			// If not processing, ESC is used to cancel the form
			return f, tea.Quit
			
		case "tab", "shift+tab":
			// When in criteria section, tab navigates between criteria fields
			if f.inCriteriaSection {
				if msg.String() == "tab" {
					f.focusedCriteria = (f.focusedCriteria + 1) % len(f.criteriasInputs)
					// If we wrapped around back to 0, move to next main section
					if f.focusedCriteria == 0 {
						f.inCriteriaSection = false
						f.focused = 0 // Move back to the first field
						
						// Update focus for all fields
						for i := range f.inputs {
							if i == f.focused {
								f.inputs[i].Focus()
							} else {
								f.inputs[i].Blur()
							}
						}
						
						for i := range f.criteriasInputs {
							f.criteriasInputs[i].Blur()
						}
					} else {
						// Update focus within criteria fields
						for i := range f.criteriasInputs {
							if i == f.focusedCriteria {
								f.criteriasInputs[i].Focus()
							} else {
								f.criteriasInputs[i].Blur()
							}
						}
					}
				} else { // shift+tab
					f.focusedCriteria = (f.focusedCriteria - 1 + len(f.criteriasInputs)) % len(f.criteriasInputs)
					// If we wrapped around to the last criteria field, go back to main fields
					if f.focusedCriteria == len(f.criteriasInputs) - 1 && !f.inCriteriaSection {
						f.inCriteriaSection = true
						f.focused = len(f.inputs) - 1 // Last normal field
						
						// Update focus for all fields
						for i := range f.inputs {
							if i == f.focused {
								f.inputs[i].Focus()
							} else {
								f.inputs[i].Blur()
							}
						}
						
						for i := range f.criteriasInputs {
							f.criteriasInputs[i].Blur()
						}
					} else {
						// Update focus within criteria fields
						for i := range f.criteriasInputs {
							if i == f.focusedCriteria {
								f.criteriasInputs[i].Focus()
							} else {
								f.criteriasInputs[i].Blur()
							}
						}
					}
				}
				return f, nil
			}
			
			// Normal tab navigation between main fields
			if msg.String() == "tab" {
				f.focused = (f.focused + 1) % len(f.inputs)
				// If we're at the last field, move to criteria section
				if f.focused == 0 {
					// We wrapped around, so go to criteria section
					f.inCriteriaSection = true
					f.focusedCriteria = 0
					
					// Update focus
					for i := range f.inputs {
						f.inputs[i].Blur()
					}
					
					f.criteriasInputs[0].Focus()
					for i := 1; i < len(f.criteriasInputs); i++ {
						f.criteriasInputs[i].Blur()
					}
					
					return f, nil
				}
			} else { // shift+tab
				f.focused = (f.focused - 1 + len(f.inputs)) % len(f.inputs)
				// If we're at the last field coming backwards, go to criteria
				if f.focused == len(f.inputs) - 1 && msg.String() == "shift+tab" {
					f.inCriteriaSection = true
					f.focusedCriteria = len(f.criteriasInputs) - 1 // Focus last criteria
					
					// Update focus
					for i := range f.inputs {
						f.inputs[i].Blur()
					}
					
					f.criteriasInputs[f.focusedCriteria].Focus()
					for i := 0; i < len(f.criteriasInputs); i++ {
						if i != f.focusedCriteria {
							f.criteriasInputs[i].Blur()
						}
					}
					
					return f, nil
				}
			}
			
			// Update field focus for main fields
			for i := range f.inputs {
				if i == f.focused {
					f.inputs[i].Focus()
				} else {
					f.inputs[i].Blur()
				}
			}
			
			return f, nil
			
		case "up", "down":
			// Handle up/down navigation
			if f.inCriteriaSection {
				// When in criteria section, up/down navigates between criteria fields
				if msg.String() == "up" {
					f.focusedCriteria = (f.focusedCriteria - 1 + len(f.criteriasInputs)) % len(f.criteriasInputs)
				} else {
					f.focusedCriteria = (f.focusedCriteria + 1) % len(f.criteriasInputs)
				}
				
				// Update focus
				for i := range f.criteriasInputs {
					if i == f.focusedCriteria {
						f.criteriasInputs[i].Focus()
					} else {
						f.criteriasInputs[i].Blur()
					}
				}
				
				return f, nil
			}
			
				// Otherwise navigate between main fields
			if msg.String() == "up" {
				f.focused = (f.focused - 1 + len(f.inputs)) % len(f.inputs)
			} else {
				f.focused = (f.focused + 1) % len(f.inputs)
			}
			
			// Update field focus
			for i := range f.inputs {
				if i == f.focused {
					f.inputs[i].Focus()
				} else {
					f.inputs[i].Blur()
				}
			}
			
			return f, nil
		
		case "enter":
			// When in criteria section, enter moves to next criteria or submits
			if f.inCriteriaSection {
				if f.focusedCriteria < len(f.criteriasInputs) - 1 {
					// Move to next criteria field
					f.criteriasInputs[f.focusedCriteria].Blur()
					f.focusedCriteria++
					f.criteriasInputs[f.focusedCriteria].Focus()
				} else {
					// On last criteria field, submit the form
					f.submitted = true
					return f, tea.Quit
				}
				return f, nil
			}
			
			// When on last regular field, move to criteria section
			if f.focused == len(f.inputs) - 1 {
				f.inputs[f.focused].Blur()
				f.inCriteriaSection = true
				f.focusedCriteria = 0
				f.criteriasInputs[f.focusedCriteria].Focus()
				return f, nil
			}
			
			// Otherwise move to next field
			f.focused = (f.focused + 1) % len(f.inputs)
			
			// Update field focus
			for i := range f.inputs {
				if i == f.focused {
					f.inputs[i].Focus()
				} else {
					f.inputs[i].Blur()
				}
			}
			
			return f, nil
		}
	
	case tea.WindowSizeMsg:
		// Update form dimensions
		f.width = msg.Width
		f.height = msg.Height
		
		// Update spinner width
		f.spinner.SetWidth(msg.Width)
	}
	
	// If processing is active, check for timeout
	if f.model.IsProcessingActive() {
		// Only check every 500ms to avoid unnecessary overhead
		if time.Since(f.lastTimeoutCheck) > 500*time.Millisecond {
			timeoutMsg := f.model.GetTimeoutMessage()
			if timeoutMsg != "" {
				f.spinner.SetAdditionalMessage(timeoutMsg)
			}
			f.lastTimeoutCheck = time.Now()
		}
	}
	
	// Update the active input if focused
	if !f.inCriteriaSection && f.focused >= 0 && f.focused < len(f.inputs) {
		// Get the current value before the update
		prevValue := f.fieldPrevValues[f.focused]
		
		// Update the input
		newInput, inputCmd := f.inputs[f.focused].Update(msg)
		f.inputs[f.focused] = newInput
		cmds = append(cmds, inputCmd)
		
		// Get the current value after the update
		currentValue := f.inputs[f.focused].Value()
		
		// If the value changed, check if it might be a paste event
		if currentValue != prevValue {
			// Check if the change is large enough to be a paste event
			newContent, isPaste := clipboard.GetActiveFieldValue(currentValue, prevValue)
			if isPaste && clipboard.IsLongEnoughForProcessing(newContent) {
				f.processClipboardContent(newContent)
			} else {
				// If user is typing, mark the field as manually edited
				f.model.MarkFieldEdited(f.getFieldNameByIndex(f.focused))
			}
		}
		
		// Store the new value for paste detection
		f.fieldPrevValues[f.focused] = currentValue
	} else if f.inCriteriaSection && f.focusedCriteria >= 0 && f.focusedCriteria < len(f.criteriasInputs) {
		// Update the criteria input
		newInput, inputCmd := f.criteriasInputs[f.focusedCriteria].Update(msg)
		f.criteriasInputs[f.focusedCriteria] = newInput
		cmds = append(cmds, inputCmd)
		
		// Get the current value after the update
		currentValue := f.criteriasInputs[f.focusedCriteria].Value()
		
		// Store the new value for paste detection
		f.criteriaPrevValues[f.focusedCriteria] = currentValue
		
		// Since criteria have changed, mark the field as manually edited
		f.model.MarkFieldEdited(AcceptanceCriteriaField)
	}
	
	return f, tea.Batch(cmds...)
}

// View renders the form
func (f *UserStoryForm) View() string {
	var b strings.Builder

	// Define styles
	labelStyle := lipgloss.NewStyle().Width(12)
	focusedLabelStyle := lipgloss.NewStyle().Width(12).Foreground(lipgloss.Color("205")).Bold(true)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))

	// Title field (special handling - first field)
	titleLabel := getFieldLabel(TitleField)
	var titleLabelStyle lipgloss.Style
	if f.focused == 0 && !f.inCriteriaSection {
		titleLabelStyle = focusedLabelStyle
	} else {
		titleLabelStyle = labelStyle
	}
	b.WriteString(titleLabelStyle.Render(titleLabel))
	
	// Add confidence indicator if needed
	if f.model != nil && f.model.IsFieldAutoPopulated(TitleField) {
		confidence := f.model.GetFieldConfidence(TitleField)
		confidenceIndicator := getConfidenceIndicator(confidence)
		b.WriteString(" " + confidenceIndicator)
	}
	
	b.WriteString(" > " + f.inputs[0].View() + "\n\n")

	// User Story section with header
	b.WriteString(headerStyle.Render("User Story") + "\n")
	
	// As a field
	asALabel := getFieldLabel(AsAField)
	var asALabelStyle lipgloss.Style
	if f.focused == 1 && !f.inCriteriaSection {
		asALabelStyle = focusedLabelStyle
	} else {
		asALabelStyle = labelStyle
	}
	b.WriteString(asALabelStyle.Render(asALabel))
	
	if f.model != nil && f.model.IsFieldAutoPopulated(AsAField) {
		confidence := f.model.GetFieldConfidence(AsAField)
		confidenceIndicator := getConfidenceIndicator(confidence)
		b.WriteString(" " + confidenceIndicator)
	}
	
	b.WriteString(" > " + f.inputs[1].View() + "\n")
	
	// I want field
	iWantLabel := getFieldLabel(IWantField)
	var iWantLabelStyle lipgloss.Style
	if f.focused == 2 && !f.inCriteriaSection {
		iWantLabelStyle = focusedLabelStyle
	} else {
		iWantLabelStyle = labelStyle
	}
	b.WriteString(iWantLabelStyle.Render(iWantLabel))
	
	if f.model != nil && f.model.IsFieldAutoPopulated(IWantField) {
		confidence := f.model.GetFieldConfidence(IWantField)
		confidenceIndicator := getConfidenceIndicator(confidence)
		b.WriteString(" " + confidenceIndicator)
	}
	
	b.WriteString(" > " + f.inputs[2].View() + "\n")
	
	// So that field
	soThatLabel := getFieldLabel(SoThatField)
	var soThatLabelStyle lipgloss.Style
	if f.focused == 3 && !f.inCriteriaSection {
		soThatLabelStyle = focusedLabelStyle
	} else {
		soThatLabelStyle = labelStyle
	}
	b.WriteString(soThatLabelStyle.Render(soThatLabel))
	
	if f.model != nil && f.model.IsFieldAutoPopulated(SoThatField) {
		confidence := f.model.GetFieldConfidence(SoThatField)
		confidenceIndicator := getConfidenceIndicator(confidence)
		b.WriteString(" " + confidenceIndicator)
	}
	
	b.WriteString(" > " + f.inputs[3].View() + "\n\n")
	
	// Acceptance Criteria section
	b.WriteString(headerStyle.Render("Acceptance Criteria") + "\n")
	
	// Render all five criteria fields with numbers
	for i := 0; i < 5; i++ {
		var criteriaLabelStyle lipgloss.Style
		if f.inCriteriaSection && f.focusedCriteria == i {
			criteriaLabelStyle = focusedLabelStyle
		} else {
			criteriaLabelStyle = labelStyle
		}
		
		b.WriteString(criteriaLabelStyle.Render(fmt.Sprintf("%d.", i+1)))
		
		// Add confidence indicator if needed
		if f.model != nil && f.model.IsFieldAutoPopulated(AcceptanceCriteriaField) {
			confidence := f.model.GetFieldConfidence(AcceptanceCriteriaField)
			confidenceIndicator := getConfidenceIndicator(confidence)
			b.WriteString(" " + confidenceIndicator)
		}
		
		b.WriteString(" > " + f.criteriasInputs[i].View() + "\n")
	}

	// Processing indicator
	if f.model != nil && f.model.IsProcessingActive() {
		if timeoutMsg := f.model.GetTimeoutMessage(); timeoutMsg != "" {
			f.spinner.SetAdditionalMessage(timeoutMsg)
		}
		
		f.spinner.SetVisible(true)
		b.WriteString("\n" + f.spinner.View() + "\n")
	}
	
	// API key message if needed
	if f.model != nil && f.model.ShouldShowAPIKeyMessage() {
		apiKeyMsg := f.model.GetAPIKeyMessage()
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Italic(true)
		b.WriteString("\n" + msgStyle.Render(apiKeyMsg) + "\n")
	}
	
	// Error message if there was an error
	if f.model != nil && f.model.GetProcessingState() == llm.ProcessingError {
		errMsg := f.model.GetLastError()
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Italic(true)
		b.WriteString("\n" + errStyle.Render("Error: "+errMsg) + "\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().Faint(true)
	
	if f.model != nil && f.model.IsProcessingActive() {
		b.WriteString("\n" + helpStyle.Render("Press ESC to cancel processing"))
	} else {
		helpText := "Tab/Shift+Tab: Navigate • Enter: Next • Esc: Cancel"
		
		// Add confirmation shortcut if there are auto-populated fields
		if f.model != nil && f.model.HasAutoPopulatedFields() {
			helpText += " • Ctrl+A: Confirm All Auto-populated Fields"
		}
		
		b.WriteString("\n" + helpStyle.Render(helpText))
		
		// Add LLM paste help text if available
		if f.processor != nil && f.processor.IsConfigured() {
			pasteHelpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Italic(true)
			// Mac systems use ⌘ (Command) instead of Ctrl
			if runtime.GOOS == "darwin" {
				b.WriteString("\n" + pasteHelpStyle.Render("Tip: Use ⌘+P to process text with LLM"))
			} else {
				b.WriteString("\n" + pasteHelpStyle.Render("Tip: Use Ctrl+P to process text with LLM"))
			}
		}
	}

	return b.String()
}

// getFieldLabel returns a user-friendly label for a field
func getFieldLabel(fieldName string) string {
	switch fieldName {
	case TitleField:
		return "Title:"
	case AsAField:
		return "As a:"
	case IWantField:
		return "I want:"
	case SoThatField:
		return "So that:"
	case AcceptanceCriteriaField:
		return "Acceptance Criteria:"
	default:
		return fieldName + ":"
	}
}

// getConfidenceIndicator returns a visual indicator based on confidence level
func getConfidenceIndicator(confidence float64) string {
	var indicator string
	
	// Choose indicator and style based on confidence level
	var style lipgloss.Style
	if confidence >= 0.8 {
		// High confidence - Green checkmark
		indicator = "✓"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
	} else if confidence >= 0.5 {
		// Medium confidence - Yellow circled checkmark
		indicator = "◎"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	} else {
		// Low confidence - Orange question mark
		indicator = "?"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	}
	
	// Add confidence percentage
	return style.Render(fmt.Sprintf("%s %.0f%%", indicator, confidence*100))
}

// GetUserStory returns a user story from the form input
func (f *UserStoryForm) GetUserStory() models.UserStory {
	// Get field values
	title := strings.TrimSpace(f.inputs[0].Value())
	asA := strings.TrimSpace(f.inputs[1].Value())
	iWant := strings.TrimSpace(f.inputs[2].Value())
	soThat := strings.TrimSpace(f.inputs[3].Value())
	
	// Build the description with the as-a, i-want, so-that format
	description := fmt.Sprintf("As a %s,\nI want %s,\nso that %s.", asA, iWant, soThat)
	
	// Collect acceptance criteria from all criteria inputs
	var criteria []string
	for _, input := range f.criteriasInputs {
		if value := strings.TrimSpace(input.Value()); value != "" {
			criteria = append(criteria, value)
		}
	}
	
	// Update the story and return it
	f.story.Title = title
	f.story.Description = description
	f.story.Criteria = criteria
	
	return f.story
}

// processClipboardContent processes clipboard content with the LLM
func (f *UserStoryForm) processClipboardContent(content string) {
	// Show spinner
	f.spinner.SetVisible(true)
	f.spinner.SetMessage("Processing pasted text...")
	
	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	f.processingCtx = ctx
	f.processingCancel = cancel
	
	// Process the content
	f.model.ProcessClipboardContent(ctx, content)
	
	// Start polling for updates to update the UI with results
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-f.processingCtx.Done():
				return
			case <-ticker.C:
				if f.model.GetProcessingState() != llm.ProcessingActive {
					// Processing finished, update UI
					f.updateUIFromModel()
					return
				}
			}
		}
	}()
}

// updateUIFromModel updates the form fields from the model data
func (f *UserStoryForm) updateUIFromModel() {
	formData := f.model.GetFormData()
	
	// Update the title field
	if f.model.IsFieldAutoPopulated(TitleField) {
		f.inputs[0].SetValue(formData.Title)
		f.fieldPrevValues[0] = formData.Title
	}
	
	// Update the as-a field
	if f.model.IsFieldAutoPopulated(AsAField) {
		f.inputs[1].SetValue(formData.AsA)
		f.fieldPrevValues[1] = formData.AsA
	}
	
	// Update the i-want field
	if f.model.IsFieldAutoPopulated(IWantField) {
		f.inputs[2].SetValue(formData.IWant)
		f.fieldPrevValues[2] = formData.IWant
	}
	
	// Update the so-that field
	if f.model.IsFieldAutoPopulated(SoThatField) {
		f.inputs[3].SetValue(formData.SoThat)
		f.fieldPrevValues[3] = formData.SoThat
	}
	
	// Update the acceptance criteria fields
	if f.model.IsFieldAutoPopulated(AcceptanceCriteriaField) {
		// Clear all criteria fields first
		for i := range f.criteriasInputs {
			f.criteriasInputs[i].SetValue("")
		}
		
		// Set the new criteria values
		for i, criterion := range formData.AcceptanceCriteria {
			if i < len(f.criteriasInputs) {
				f.criteriasInputs[i].SetValue(criterion)
				f.criteriaPrevValues[i] = criterion
			}
		}
	}
	
	// Hide spinner
	f.spinner.SetVisible(false)
}

// getFieldNameByIndex returns the field name for a given index
func (f *UserStoryForm) getFieldNameByIndex(index int) string {
	switch index {
	case 0:
		return TitleField
	case 1:
		return AsAField
	case 2:
		return IWantField
	case 3:
		return SoThatField
	default:
		return ""
	}
}

// hideSpinnerMsg is a message to hide the spinner
type hideSpinnerMsg struct{}

// parseAcceptanceCriteria extracts acceptance criteria from the input string
// This method is kept for testing compatibility
func (f *UserStoryForm) parseAcceptanceCriteria(input string) []string {
	// Handle empty input
	if strings.TrimSpace(input) == "" {
		return []string{}
	}
	
	// If there are no newlines, we might have space-separated criteria
	if !strings.Contains(input, "\n") {
		// If this looks like a single sentence/phrase with multiple words,
		// treat it as a single criterion
		if !regexp.MustCompile(`\s{2,}`).MatchString(input) && 
		   !strings.Contains(input, ",") && !strings.Contains(input, ";") {
			trimmed := strings.TrimSpace(input)
			words := strings.Fields(trimmed)
			if len(words) <= 4 { // Likely individual criteria if 4 or fewer words
				return words
			}
			// Otherwise it's probably a sentence that should be kept intact
			return []string{trimmed}
		}
		// If we have double spaces or other separators, split by them
		return strings.Fields(input)
	}
	
	// For multi-line input
	var criteria []string
	lines := strings.Split(input, "\n")
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		
		// Check for bullet points or numbered lists
		if strings.HasPrefix(trimmedLine, "- ") {
			// Dash bullet point
			criteria = append(criteria, strings.TrimSpace(trimmedLine[2:]))
		} else if strings.HasPrefix(trimmedLine, "* ") {
			// Asterisk bullet point
			criteria = append(criteria, strings.TrimSpace(trimmedLine[2:]))
		} else if strings.HasPrefix(trimmedLine, "• ") {
			// Bullet point (•)
			criteria = append(criteria, strings.TrimSpace(trimmedLine[2:]))
		} else if matches := regexp.MustCompile(`^\d+\.\s+(.+)$`).FindStringSubmatch(trimmedLine); len(matches) > 1 {
			// Numbered list (1., 2., etc.)
			criteria = append(criteria, strings.TrimSpace(matches[1]))
		} else if matches := regexp.MustCompile(`^\(\d+\)\s+(.+)$`).FindStringSubmatch(trimmedLine); len(matches) > 1 {
			// Parenthesized numbers ((1), (2), etc.)
			criteria = append(criteria, strings.TrimSpace(matches[1]))
		} else {
			// Regular line - keep intact as a single criterion
			criteria = append(criteria, trimmedLine)
		}
	}
	
	return criteria
} 
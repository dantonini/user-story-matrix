// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package userstoryform

import (
	"context"
	"fmt"
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
	model           *formModels.UserStoryFormModel
	story           models.UserStory
	inputs          []textinput.Model
	cursor          cursor.Model
	focused         int
	submitted       bool
	width           int
	height          int
	spinner         spinner.Model
	fieldPrevValues map[int]string
	processor       llm.LLMProcessor
	configManager   *llm.ConfigManager
	processingCtx   context.Context
	processingCancel context.CancelFunc
	lastTimeoutCheck time.Time
}

// New creates a new UserStoryForm
func New(us models.UserStory, processor llm.LLMProcessor, configManager *llm.ConfigManager) *UserStoryForm {
	form := &UserStoryForm{
		story:           us,
		cursor:          cursor.New(),
		focused:         0,
		processor:       processor,
		configManager:   configManager,
		spinner:         spinner.New(),
		fieldPrevValues: make(map[int]string),
		lastTimeoutCheck: time.Now(),
	}

	// Initialize the model
	form.model = formModels.NewUserStoryFormModel(us, processor, configManager)

	// Initialize inputs
	form.inputs = make([]textinput.Model, 5)
	
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
	
	// Initialize acceptance criteria input
	form.inputs[4] = textinput.New()
	form.inputs[4].Placeholder = "Acceptance criteria (one per line)"
	form.inputs[4].Width = 60
	form.inputs[4].Prompt = ""
	form.inputs[4].CharLimit = 0
	
	// Set existing criteria if any
	if len(us.Criteria) > 0 {
		// Join criteria with spaces since textinput seems to replace newlines with spaces
		form.inputs[4].SetValue(strings.Join(us.Criteria, " "))
	}

	// Set values from user story
	if us.Title != "" {
		form.inputs[0].SetValue(us.Title)
	}
	
	// Store initial values for paste detection
	for i, input := range form.inputs {
		form.fieldPrevValues[i] = input.Value()
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
			// Handle tab navigation
			if msg.String() == "tab" {
				f.focused = (f.focused + 1) % len(f.inputs)
			} else {
				f.focused = (f.focused - 1 + len(f.inputs)) % len(f.inputs)
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
			
		case "up", "down":
			// Only navigate between fields with up/down when in the acceptance criteria field
			if f.focused == FieldIndex[AcceptanceCriteriaField] {
				// Handle navigation within multi-line text
				newInput, inputCmd := f.inputs[f.focused].Update(msg)
				f.inputs[f.focused] = newInput
				cmds = append(cmds, inputCmd)
				return f, tea.Batch(cmds...)
			}
			
			// Otherwise navigate between fields
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
			if f.focused == len(f.inputs)-1 {
				// Submit the form if the last input is focused
				f.submitted = true
				return f, tea.Quit
			}
			// Move to the next field
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
	if f.focused >= 0 && f.focused < len(f.inputs) {
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
	}
	
	return f, tea.Batch(cmds...)
}

// View renders the form
func (f *UserStoryForm) View() string {
	var b strings.Builder
	
	// Basic form header
	b.WriteString("Create a new user story\n\n")
	
	// Show spinner during processing
	if f.model.IsProcessingActive() {
		f.spinner.SetVisible(true)
		b.WriteString(f.spinner.View() + "\n\n")
	} else {
		f.spinner.SetVisible(false)
	}
	
	// Show processing error if there was one
	if f.model.GetProcessingState() == llm.ProcessingError {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		b.WriteString(errorStyle.Render("✗ Error: " + f.model.GetLastError()) + "\n\n")
	}
	
	// Show API key not configured message if needed
	if f.model.ShouldShowAPIKeyMessage() {
		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
		b.WriteString(infoStyle.Render("ℹ " + f.model.GetAPIKeyMessage()) + "\n\n")
	}
	
	// Render each input field
	for i, input := range f.inputs {
		// Field labels
		var fieldLabel string
		var fieldName string
		
		switch i {
		case FieldIndex[TitleField]:
			fieldLabel = "Title:"
			fieldName = TitleField
		case FieldIndex[AsAField]:
			fieldLabel = "As a:"
			fieldName = AsAField
		case FieldIndex[IWantField]:
			fieldLabel = "I want:"
			fieldName = IWantField
		case FieldIndex[SoThatField]:
			fieldLabel = "So that:"
			fieldName = SoThatField
		case FieldIndex[AcceptanceCriteriaField]:
			fieldLabel = "Acceptance criteria:"
			fieldName = AcceptanceCriteriaField
		}
		
		// Add highlighting for auto-populated fields
		if f.model.IsFieldAutoPopulated(fieldName) {
			// Get the confidence level for the field
			confidence := f.model.GetFieldConfidence(fieldName)
			
			// Apply color based on confidence
			var style lipgloss.Style
			if confidence >= 0.8 {
				// High confidence - green
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
				fieldLabel = style.Render(fieldLabel + " ✓")
			} else if confidence >= 0.5 {
				// Medium confidence - yellow
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
				fieldLabel = style.Render(fieldLabel + " ●")
			} else {
				// Low confidence - orange
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
				fieldLabel = style.Render(fieldLabel + " ○")
			}
		}
		
		// Highlight focused field
		if i == f.focused {
			focusedStyle := lipgloss.NewStyle().Bold(true)
			fieldLabel = focusedStyle.Render(fieldLabel)
		}
		
		b.WriteString(fieldLabel + "\n")
		b.WriteString(input.View() + "\n\n")
	}
	
	// Form submission instructions
	if !f.model.IsProcessingActive() {
		b.WriteString("\n")
		b.WriteString("Press Enter to navigate between fields.\n")
		b.WriteString("Press Enter in the last field to submit the form.\n")
		
		// Add paste instruction if LLM is configured
		if f.processor.IsConfigured() {
			b.WriteString("Paste unstructured text to auto-populate fields.\n")
		}
	}
	
	return b.String()
}

// GetUserStory returns the user story data from the form
func (f *UserStoryForm) GetUserStory() models.UserStory {
	// Parse acceptance criteria 
	criteria := []string{}
	
	if acValue := f.inputs[FieldIndex[AcceptanceCriteriaField]].Value(); acValue != "" {
		// Bubble Tea's textinput doesn't handle newlines well,
		// so we need to consider various formatting approaches
		
		// First try splitting by newlines (if the user manually entered them)
		parts := strings.Split(acValue, "\n")
		
		// If we only have one part and it contains spaces, it might be 
		// separate criteria on one line
		if len(parts) == 1 && strings.Contains(acValue, " ") {
			// We'll try to identify if these are actually separate criteria or 
			// just multiple words in a single criterion
			
			// Here we'll implement a simple heuristic:
			// Split by spaces, but then try to group words that form a single criterion
			// For now, we'll use a simple rule: each word is a separate criterion
			words := strings.Fields(acValue)
			for _, word := range words {
				if trimmed := strings.TrimSpace(word); trimmed != "" {
					criteria = append(criteria, trimmed)
				}
			}
		} else {
			// We had multiple lines already, so process each one
			for _, part := range parts {
				if trimmed := strings.TrimSpace(part); trimmed != "" {
					criteria = append(criteria, trimmed)
				}
			}
		}
	}
	
	// Build the user story description with the as-a, i-want, so-that format
	asA := f.inputs[FieldIndex[AsAField]].Value()
	iWant := f.inputs[FieldIndex[IWantField]].Value()
	soThat := f.inputs[FieldIndex[SoThatField]].Value()
	
	description := fmt.Sprintf("As a %s,\nI want %s,\nso that %s.", asA, iWant, soThat)
	
	// Update the story
	f.story.Title = f.inputs[FieldIndex[TitleField]].Value()
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
		f.inputs[FieldIndex[TitleField]].SetValue(formData.Title)
		f.fieldPrevValues[FieldIndex[TitleField]] = formData.Title
	}
	
	// Update the as-a field
	if f.model.IsFieldAutoPopulated(AsAField) {
		f.inputs[FieldIndex[AsAField]].SetValue(formData.AsA)
		f.fieldPrevValues[FieldIndex[AsAField]] = formData.AsA
	}
	
	// Update the i-want field
	if f.model.IsFieldAutoPopulated(IWantField) {
		f.inputs[FieldIndex[IWantField]].SetValue(formData.IWant)
		f.fieldPrevValues[FieldIndex[IWantField]] = formData.IWant
	}
	
	// Update the so-that field
	if f.model.IsFieldAutoPopulated(SoThatField) {
		f.inputs[FieldIndex[SoThatField]].SetValue(formData.SoThat)
		f.fieldPrevValues[FieldIndex[SoThatField]] = formData.SoThat
	}
	
	// Update the acceptance criteria field
	if f.model.IsFieldAutoPopulated(AcceptanceCriteriaField) {
		// Join criteria with spaces since textinput seems to replace newlines with spaces
		f.inputs[FieldIndex[AcceptanceCriteriaField]].SetValue(strings.Join(formData.AcceptanceCriteria, " "))
		f.fieldPrevValues[FieldIndex[AcceptanceCriteriaField]] = strings.Join(formData.AcceptanceCriteria, " ")
	}
	
	// Hide spinner
	f.spinner.SetVisible(false)
}

// getFieldNameByIndex returns the field name for the given index
func (f *UserStoryForm) getFieldNameByIndex(index int) string {
	for name, idx := range FieldIndex {
		if idx == index {
			return name
		}
	}
	return ""
} 
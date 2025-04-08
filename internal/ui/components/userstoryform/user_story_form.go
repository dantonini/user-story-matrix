// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package userstoryform

import (
	"context"
	"fmt"
	"os/exec"
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
	"github.com/user-story-matrix/usm/internal/ui/contracts"
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

// Verify UserStoryForm implements the contracts.UserStorySubmitter interface
var _ contracts.UserStorySubmitter = (*UserStoryForm)(nil)

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
	// Return both text input blink and spinner tick commands
	return tea.Batch(
		textinput.Blink,
		f.spinner.Init(), // Initialize spinner animation
	)
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
		return f.handleHideSpinner()
		
	case tea.KeyMsg:
		// Check for paste event
		if cmd := f.handlePasteEvent(msg); cmd != nil {
			return f, cmd
		}
		
		// Handle key messages
		if cmd := f.handleKeyPress(msg); cmd != nil {
			return f, cmd
		}
	
	case tea.WindowSizeMsg:
		f.handleWindowResize(msg)
	}
	
	// Handle processing timeout checks
	f.checkProcessingTimeout()
	
	// Update active input field
	if cmd := f.updateActiveField(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	return f, tea.Batch(cmds...)
}

// handleHideSpinner handles the hideSpinnerMsg
func (f *UserStoryForm) handleHideSpinner() (tea.Model, tea.Cmd) {
	// Set visibility directly rather than using SetVisible
	f.spinner.Visible = false
	return f, nil
}

// handlePasteEvent checks for and processes paste events
func (f *UserStoryForm) handlePasteEvent(msg tea.KeyMsg) tea.Cmd {
	// Only do paste detection for specific events, ignore normal typing
	// This prevents debug messages from interfering with normal input
	if !clipboard.IsPasteEvent(msg) {
		return nil
	}
	
	// For regular paste events, get the pasted text
	pastedText := getClipboardContent()
	
	// Process the pasted content if it's long enough
	if pastedText != "" && clipboard.IsLongEnoughForProcessing(pastedText) {
		// Use a more animated spinner for clipboard processing
		f.spinner.SetStyle(spinner.StyleBars)  // Changed from Points to Bars for smoother animation
		f.spinner.SetMessage("Processing pasted content...")
		f.spinner.SetVisible(true)
		f.processClipboardContent(pastedText)
	}
	
	return nil
}


// getClipboardContent retrieves content from the system clipboard
func getClipboardContent() string {
	// Try to run the appropriate clipboard command based on OS
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "windows":
		cmd = exec.Command("powershell.exe", "-command", "Get-Clipboard")
	default: // Linux and others
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
	}
	
	output, err := cmd.Output()
	if err != nil {
		// If command fails, try fallback for Linux
		if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
			fallbackCmd := exec.Command("xsel", "--clipboard", "--output")
			output, err = fallbackCmd.Output()
			if err != nil {
				return ""
			}
		} else {
			return ""
		}
	}
	
	return string(output)
}


// handleKeyPress processes individual key presses
func (f *UserStoryForm) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+a":
		return f.handleConfirmAllFields()
			
	case "esc":
		return f.handleEscapeKey()
			
	case "tab", "shift+tab", "up", "down":
		return f.handleTabNavigation(msg)
		
	case "enter":
		return f.handleEnterKey()
	}
	
	return nil
}

// handleConfirmAllFields handles the Ctrl+A key combination for confirming auto-populated fields
func (f *UserStoryForm) handleConfirmAllFields() tea.Cmd {
	if f.model == nil || !f.model.HasAutoPopulatedFields() {
		return nil
	}
	
	f.model.ConfirmAllFields()
	f.updateUIFromModel()
	
	// Show a confirmation message
	confirmMsg := fmt.Sprintf("Confirmed %d auto-populated fields", f.model.GetAutoPopulatedFieldCount())
	confirmStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Italic(true)
	
	// Set a message in the spinner and show it briefly with a growing bar spinner
	f.spinner.SetStyle(spinner.StyleGrowingBar)  // Changed from Pulse to GrowingBar
	f.spinner.SetMessage(confirmStyle.Render(confirmMsg))
	f.spinner.SetVisible(true)
	
	// Schedule the spinner to be hidden after a short delay
	return tea.Tick(time.Second*2, func(time.Time) tea.Msg {
		return hideSpinnerMsg{}
	})
}

// handleEscapeKey handles the Escape key press
func (f *UserStoryForm) handleEscapeKey() tea.Cmd {
	// Cancel processing if active
	if f.model.IsProcessingActive() {
		f.model.CancelProcessing()
		if f.processingCancel != nil {
			f.processingCancel()
		}
		// Set visibility directly rather than using SetVisible
		f.spinner.Visible = false
		return nil
	}
	
	// If not processing, ESC is used to cancel the form
	return tea.Quit
}

// handleTabNavigation handles Tab and Shift+Tab navigation
func (f *UserStoryForm) handleTabNavigation(msg tea.KeyMsg) tea.Cmd {
	// Get the key string
	keyStr := msg.String()
	
	// Handle forward/backward navigation based on key
	if keyStr == "tab" || keyStr == "down" {
		// Forward navigation
		if f.inCriteriaSection {
			return f.handleForwardCriteriaNavigation()
		}
		return f.handleForwardMainNavigation()
	} else if keyStr == "shift+tab" || keyStr == "up" {
		// Backward navigation
		if f.inCriteriaSection {
			return f.handleBackwardCriteriaNavigation()
		}
		return f.handleBackwardMainNavigation()
	}
	
	return nil
}

// handleForwardMainNavigation handles Tab/Down key when in main fields section
func (f *UserStoryForm) handleForwardMainNavigation() tea.Cmd {
	f.focused = (f.focused + 1) % len(f.inputs)
	
	// If we wrapped around back to first field, move to criteria section
	if f.focused == 0 {
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
		
		return nil
	}
	
	// Update field focus for main fields
	f.updateMainFieldsFocus()
	
	return nil
}

// handleBackwardMainNavigation handles Shift+Tab/Up key when in main fields section
func (f *UserStoryForm) handleBackwardMainNavigation() tea.Cmd {
	// If we're already at the first field, go to criteria section
	if f.focused == 0 {
		f.inCriteriaSection = true
		f.focusedCriteria = len(f.criteriasInputs) - 1
		
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
		
		return nil
	}
	
	// Otherwise, move to previous field
	f.focused = (f.focused - 1 + len(f.inputs)) % len(f.inputs)
	
	// Update field focus for main fields
	f.updateMainFieldsFocus()
	
	return nil
}

// handleForwardCriteriaNavigation handles Tab/Down key when in criteria section
func (f *UserStoryForm) handleForwardCriteriaNavigation() tea.Cmd {
	f.focusedCriteria = (f.focusedCriteria + 1) % len(f.criteriasInputs)
	
	// If we wrapped around back to the first criteria, go to main section
	if f.focusedCriteria == 0 {
		f.inCriteriaSection = false
		f.focused = 0 // Go to first main field
		f.updateAllFocus()
	} else {
		f.updateCriteriaFocus()
	}
	
	return nil
}

// handleBackwardCriteriaNavigation handles Shift+Tab/Up key when in criteria section
func (f *UserStoryForm) handleBackwardCriteriaNavigation() tea.Cmd {
	// If already at the first criteria, move to main fields
	if f.focusedCriteria == 0 {
		f.inCriteriaSection = false
		f.focused = len(f.inputs) - 1 // Go to last main field
		f.updateAllFocus()
		return nil
	}
	
	// Otherwise, move to previous criteria
	f.focusedCriteria = (f.focusedCriteria - 1 + len(f.criteriasInputs)) % len(f.criteriasInputs)
	f.updateCriteriaFocus()
	
	return nil
}

// handleEnterKey handles the Enter key press
func (f *UserStoryForm) handleEnterKey() tea.Cmd {
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
			return tea.Quit
		}
		return nil
	}
	
	// When on last regular field, move to criteria section
	if f.focused == len(f.inputs) - 1 {
		f.inputs[f.focused].Blur()
		f.inCriteriaSection = true
		f.focusedCriteria = 0
		f.criteriasInputs[f.focusedCriteria].Focus()
		return nil
	}
	
	// Otherwise move to next field
	f.focused = (f.focused + 1) % len(f.inputs)
	
	// Update field focus
	f.updateMainFieldsFocus()
	
	return nil
}

// handleWindowResize handles window resize events
func (f *UserStoryForm) handleWindowResize(msg tea.WindowSizeMsg) {
	// Update form dimensions
	f.width = msg.Width
	f.height = msg.Height
	
	// Update spinner width
	f.spinner.SetWidth(msg.Width)
}

// checkProcessingTimeout checks for timeout during processing
func (f *UserStoryForm) checkProcessingTimeout() {
	if !f.model.IsProcessingActive() {
		return
	}
	
	// Only check every 500ms to avoid unnecessary overhead
	if time.Since(f.lastTimeoutCheck) > 500*time.Millisecond {
		timeoutMsg := f.model.GetTimeoutMessage()
		if timeoutMsg != "" {
			// When we're experiencing a delay, change the spinner style to show progress
			f.spinner.SetStyle(spinner.StyleGrowingBar)  // Changed from Meter to GrowingBar
			f.spinner.SetAdditionalMessage(timeoutMsg)
		}
		f.lastTimeoutCheck = time.Now()
	}
}

// updateActiveField updates the active input field based on current focus
func (f *UserStoryForm) updateActiveField(msg tea.Msg) tea.Cmd {
	if !f.inCriteriaSection && f.focused >= 0 && f.focused < len(f.inputs) {
		return f.updateMainField(msg)
	} else if f.inCriteriaSection && f.focusedCriteria >= 0 && f.focusedCriteria < len(f.criteriasInputs) {
		return f.updateCriteriaField(msg)
	}
	return nil
}

// updateMainField updates the currently focused main field
func (f *UserStoryForm) updateMainField(msg tea.Msg) tea.Cmd {
	// Get the current value before the update
	prevValue := f.fieldPrevValues[f.focused]
	
	// Update the input
	newInput, inputCmd := f.inputs[f.focused].Update(msg)
	f.inputs[f.focused] = newInput
	
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
	
	return inputCmd
}

// updateCriteriaField updates the currently focused criteria field
func (f *UserStoryForm) updateCriteriaField(msg tea.Msg) tea.Cmd {
	// Update the criteria input
	newInput, inputCmd := f.criteriasInputs[f.focusedCriteria].Update(msg)
	f.criteriasInputs[f.focusedCriteria] = newInput
	
	// Get the current value after the update
	currentValue := f.criteriasInputs[f.focusedCriteria].Value()
	
	// Store the new value for paste detection
	f.criteriaPrevValues[f.focusedCriteria] = currentValue
	
	// Since criteria have changed, mark the field as manually edited
	f.model.MarkFieldEdited(AcceptanceCriteriaField)
	
	return inputCmd
}

// updateMainFieldsFocus updates the focus state for all main fields
func (f *UserStoryForm) updateMainFieldsFocus() {
	for i := range f.inputs {
		if i == f.focused {
			f.inputs[i].Focus()
		} else {
			f.inputs[i].Blur()
		}
	}
}

// updateCriteriaFocus updates the focus state for all criteria fields
func (f *UserStoryForm) updateCriteriaFocus() {
	for i := range f.criteriasInputs {
		if i == f.focusedCriteria {
			f.criteriasInputs[i].Focus()
		} else {
			f.criteriasInputs[i].Blur()
		}
	}
}

// updateAllFocus updates focus state for both main fields and criteria fields
func (f *UserStoryForm) updateAllFocus() {
	// Update main fields focus
	for i := range f.inputs {
		if i == f.focused && !f.inCriteriaSection {
			f.inputs[i].Focus()
		} else {
			f.inputs[i].Blur()
		}
	}
	
	// Update criteria fields focus
	for i := range f.criteriasInputs {
		if i == f.focusedCriteria && f.inCriteriaSection {
			f.criteriasInputs[i].Focus()
		} else {
			f.criteriasInputs[i].Blur()
		}
	}
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
		
		// Just set visibility flag - animation command comes from the Update method
		f.spinner.Visible = true
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
		
		// Show error with a special spinner style
		f.spinner.SetStyle(spinner.StyleBars)  // Changed from Pulse to Bars
		f.spinner.SetForegroundColor(lipgloss.Color("196")) // Red color for error
		f.spinner.SetMessage("Error occurred")
		f.spinner.SetAdditionalMessage(errMsg)
		f.spinner.SetVisible(true)
		
		b.WriteString("\n" + f.spinner.View() + "\n")
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
			// F2 key works the same on all platforms
			b.WriteString("\n" + pasteHelpStyle.Render("Tip: Press F2 to process clipboard text with LLM"))
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
	
	// Update the story fields
	f.story.Title = title
	f.story.Description = description
	f.story.Criteria = criteria
	f.story.LastUpdated = time.Now()
	
	// Build the content without metadata
	var contentWithoutMetadata strings.Builder
	
	// Add title
	contentWithoutMetadata.WriteString(fmt.Sprintf("# %s\n\n", title))
	
	// Add user story description
	contentWithoutMetadata.WriteString(description + "\n\n")
	
	// Add acceptance criteria
	contentWithoutMetadata.WriteString("## Acceptance criteria\n\n")
	for _, criterion := range criteria {
		contentWithoutMetadata.WriteString(fmt.Sprintf("- %s\n", criterion))
	}
	
	// Calculate the content hash
	contentHash := models.GenerateContentHash(contentWithoutMetadata.String())
	f.story.ContentHash = contentHash
	
	// Build the final content with metadata
	var finalContent strings.Builder
	finalContent.WriteString("---\n")
	finalContent.WriteString(fmt.Sprintf("file_path: %s\n", f.story.FilePath))
	finalContent.WriteString(fmt.Sprintf("created_at: %s\n", f.story.CreatedAt.Format(time.RFC3339)))
	finalContent.WriteString(fmt.Sprintf("last_updated: %s\n", f.story.LastUpdated.Format(time.RFC3339)))
	finalContent.WriteString(fmt.Sprintf("_content_hash: %s\n", contentHash))
	finalContent.WriteString("---\n\n")
	finalContent.WriteString(contentWithoutMetadata.String())
	
	// Set the final content
	f.story.Content = finalContent.String()
	
	return f.story
}

// processClipboardContent processes clipboard content with the LLM
func (f *UserStoryForm) processClipboardContent(content string) {
	// Show spinner with a more animated style for large content processing
	spinnerCmd := f.spinner.SetVisible(true)
	f.spinner.SetMessage("Processing pasted text...")
	
	// We need to send the spinner tick command to the main program loop
	// This is done via a goroutine to avoid blocking
	go func() {
		time.Sleep(10 * time.Millisecond) // Brief delay
		if spinnerCmd != nil {
			spinnerCmd()
		}
	}()
	
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
	
	// Reset spinner appearance to default
	f.resetSpinnerAppearance()
	
	// Hide spinner by setting visibility directly
	f.spinner.Visible = false
}

// resetSpinnerAppearance resets the spinner to its default appearance after special states
func (f *UserStoryForm) resetSpinnerAppearance() {
	// Reset to default style and color
	f.spinner.SetStyle(spinner.StyleCircle)  // Changed from MinimalDot to Circle
	f.spinner.SetForegroundColor(lipgloss.Color("205"))
	f.spinner.SetMessage("Processing...")
	f.spinner.SetAdditionalMessage("")
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

// GetConfirmSubmission returns whether the form was submitted
func (f *UserStoryForm) GetConfirmSubmission() bool {
	return f.submitted
}
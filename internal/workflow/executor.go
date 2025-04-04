// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"strings"
)

// StepExecutor handles the execution of workflow steps
type StepExecutor struct {
	fs FileSystem
	io UserOutput
}

// NewStepExecutor creates a new step executor instance
func NewStepExecutor(fs FileSystem, io UserOutput) *StepExecutor {
	return &StepExecutor{
		fs: fs,
		io: io,
	}
}

// ExecuteStep executes a workflow step and outputs the processed prompt to stdout.
// The outputFile parameter is only used for backward compatibility with the existing API,
// but no file is actually written.
func (e *StepExecutor) ExecuteStep(changeRequestPath string, step WorkflowStep, outputFile string) (bool, error) {
	// Print progress message only in debug mode
	if e.io.IsDebugEnabled() {
		e.io.PrintProgress(fmt.Sprintf(ProgressExecutingStep, step.ID, step.Description))
	}

	// Validate the prompt for syntax errors
	if step.Prompt != "" {
		if err := ValidatePrompt(step.Prompt); err != nil {
			e.io.PrintWarning(fmt.Sprintf("Prompt validation warning for step %s: %v", step.ID, err))
		}
	}

	// Check if the change request file exists
	if !e.fs.Exists(changeRequestPath) {
		e.io.PrintError(fmt.Sprintf(ErrFileNotFound, changeRequestPath))
		return false, fmt.Errorf(ErrFileNotFound, changeRequestPath)
	}

	// Process the prompt with variable interpolation
	processedPrompt, missingVars := InterpolatePromptWithMissingVars(step.Prompt, PromptVariables{
		ChangeRequestFilePath: changeRequestPath,
	})

	// Warn about missing variables
	if len(missingVars) > 0 {
		e.io.PrintWarning(fmt.Sprintf("Step %s contains undefined variables: %v", step.ID, missingVars))
	}

	// Print the processed prompt directly to stdout instead of writing to a file
	e.io.Print(processedPrompt)

	return true, nil
}

// formatPromptAsInstructions formats the prompt text as numbered instructions
func formatPromptAsInstructions(prompt string) string {
	if prompt == "" {
		return "No specific instructions provided."
	}

	// Handle special case for invalid sentences
	trimmedPrompt := strings.TrimSpace(prompt)
	if isInvalidSentence(trimmedPrompt) {
		return "No specific instructions provided."
	}

	// Extract key points from the prompt
	sentences := extractSentences(prompt)

	// Handle the case where no valid sentences were found
	if len(sentences) == 0 {
		return "No specific instructions provided."
	}

	// Format sentences as numbered instructions
	var result strings.Builder
	instructionCount := 0

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" || isInvalidSentence(sentence) {
			continue
		}

		instructionCount++
		// Add numbered point
		result.WriteString(fmt.Sprintf("%d. %s\n", instructionCount, sentence))
	}

	// If no valid instructions were created, provide a fallback
	if instructionCount == 0 {
		return "No specific instructions provided."
	}

	return result.String()
}

// isInvalidSentence checks if a string is not a meaningful sentence
func isInvalidSentence(s string) bool {
	// Remove all punctuation
	noPunct := strings.ReplaceAll(s, ".", "")
	noPunct = strings.ReplaceAll(noPunct, ",", "")
	noPunct = strings.ReplaceAll(noPunct, "!", "")
	noPunct = strings.ReplaceAll(noPunct, "?", "")
	noPunct = strings.ReplaceAll(noPunct, ":", "")
	noPunct = strings.ReplaceAll(noPunct, ";", "")

	// If the remaining string is just whitespace, it's invalid
	return strings.TrimSpace(noPunct) == ""
}

// extractSentences splits a text into individual sentences
func extractSentences(text string) []string {
	// If text is empty or just whitespace, return empty slice
	if strings.TrimSpace(text) == "" {
		return []string{}
	}

	// Clean up text by removing double punctuation
	text = cleanPunctuation(text)

	// Simple implementation - split on period followed by space or newline
	var sentences []string

	// Handle special cases where text might not end with punctuation
	ensureEndingPunctuation := func(t string) string {
		t = strings.TrimSpace(t)
		if t == "" {
			return t
		}

		lastChar := t[len(t)-1]
		if lastChar != '.' && lastChar != '?' && lastChar != '!' {
			return t + "."
		}
		return t
	}

	text = ensureEndingPunctuation(text)

	// Replace common ending punctuation with a special marker
	text = strings.ReplaceAll(text, ". ", ".|.")
	text = strings.ReplaceAll(text, ".\n", ".|.")
	text = strings.ReplaceAll(text, "? ", "?|.")
	text = strings.ReplaceAll(text, "?\n", "?|.")
	text = strings.ReplaceAll(text, "! ", "!|.")
	text = strings.ReplaceAll(text, "!\n", "!|.")

	// Split on the marker
	parts := strings.Split(text, "|.")

	// Process each part
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			sentences = append(sentences, part)
		}
	}

	// If we couldn't split properly, just use the whole text
	if len(sentences) == 0 && text != "" {
		sentences = append(sentences, strings.TrimSpace(text))
	}

	return sentences
}

// cleanPunctuation cleans up excessive punctuation in text
func cleanPunctuation(text string) string {
	// Replace double periods with single periods
	for strings.Contains(text, "..") {
		text = strings.ReplaceAll(text, "..", ".")
	}

	// Replace other excessive punctuation
	for strings.Contains(text, ",,") {
		text = strings.ReplaceAll(text, ",,", ",")
	}

	for strings.Contains(text, "!!") {
		text = strings.ReplaceAll(text, "!!", "!")
	}

	for strings.Contains(text, "??") {
		text = strings.ReplaceAll(text, "??", "?")
	}

	return text
}

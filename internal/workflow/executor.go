// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"path/filepath"
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

// ExecuteStep executes a workflow step and produces an output file
func (e *StepExecutor) ExecuteStep(changeRequestPath string, step WorkflowStep, outputFile string) (bool, error) {
	// Read the change request file
	content, err := e.fs.ReadFile(changeRequestPath)
	if err != nil {
		return false, fmt.Errorf("failed to read change request file: %w", err)
	}

	// Generate step-specific content
	outputContent, err := e.generateStepContent(string(content), step)
	if err != nil {
		return false, fmt.Errorf("failed to generate step content: %w", err)
	}

	// Create directory if it doesn't exist
	dirPath := filepath.Dir(outputFile)
	if dirPath != "" && !e.fs.Exists(dirPath) {
		if err := e.fs.MkdirAll(dirPath, 0755); err != nil {
			return false, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Write the output file
	if err := e.fs.WriteFile(outputFile, []byte(outputContent), 0644); err != nil {
		return false, fmt.Errorf("failed to write output file: %w", err)
	}

	return true, nil
}

// generateStepContent generates the content for a specific step
func (e *StepExecutor) generateStepContent(changeRequestContent string, step WorkflowStep) (string, error) {
	// Common header for all steps
	header := fmt.Sprintf("# %s\n\n", step.Description)

	// Step-specific content
	var content string
	switch step.ID {
	case "01-laying-the-foundation":
		content = "## Architecture & Design\n\n" +
			"This step focuses on setting up the architecture and structure for the implementation.\n\n" +
			"### Key Activities\n" +
			"1. Create necessary packages and interfaces\n" +
			"2. Define core data structures\n" +
			"3. Establish file organization\n" +
			"4. Set up testing infrastructure\n\n"

	case "01-laying-the-foundation-test":
		content = "## Foundation Testing\n\n" +
			"This step verifies the foundational changes made in the previous step.\n\n" +
			"### Test Coverage\n" +
			"1. Package structure validation\n" +
			"2. Interface completeness\n" +
			"3. Data structure integrity\n" +
			"4. Test infrastructure functionality\n\n"

	case "02-mvi":
		content = "## Minimum Viable Implementation\n\n" +
			"This step implements the core functionality with minimal features.\n\n" +
			"### Implementation Focus\n" +
			"1. Core business logic\n" +
			"2. Essential functionality\n" +
			"3. Basic error handling\n" +
			"4. Minimal user interface\n\n"

	case "02-mvi-test":
		content = "## MVI Testing\n\n" +
			"This step verifies the minimum viable implementation.\n\n" +
			"### Test Coverage\n" +
			"1. Core functionality tests\n" +
			"2. Basic error handling tests\n" +
			"3. Integration tests\n" +
			"4. User interface tests\n\n"

	case "03-extend-functionalities":
		content = "## Extended Functionality\n\n" +
			"This step adds additional features and improvements.\n\n" +
			"### Implementation Focus\n" +
			"1. Additional features\n" +
			"2. Enhanced error handling\n" +
			"3. Performance optimizations\n" +
			"4. User experience improvements\n\n"

	case "03-extend-functionalities-test":
		content = "## Extended Functionality Testing\n\n" +
			"This step verifies the extended functionality.\n\n" +
			"### Test Coverage\n" +
			"1. Feature tests\n" +
			"2. Error handling tests\n" +
			"3. Performance tests\n" +
			"4. User experience tests\n\n"

	case "04-final-iteration":
		content = "## Final Iteration\n\n" +
			"This step focuses on polishing and final adjustments.\n\n" +
			"### Implementation Focus\n" +
			"1. Code cleanup\n" +
			"2. Documentation updates\n" +
			"3. Final optimizations\n" +
			"4. User feedback incorporation\n\n"

	case "04-final-iteration-test":
		content = "## Final Testing\n\n" +
			"This step performs final verification and validation.\n\n" +
			"### Test Coverage\n" +
			"1. End-to-end tests\n" +
			"2. Documentation verification\n" +
			"3. Performance benchmarks\n" +
			"4. User acceptance tests\n\n"

	default:
		return "", fmt.Errorf("unknown step ID: %s", step.ID)
	}

	// Add change request context
	context := fmt.Sprintf("## Change Request Context\n\n"+
		"This step was executed for change request: %s\n\n"+
		"Step ID: %s\n"+
		"Step Description: %s\n"+
		"Is Test Step: %t\n\n",
		changeRequestContent,
		step.ID,
		step.Description,
		step.IsTest,
	)

	return header + content + context, nil
} 
// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"errors"
)

// Static error variables for the workflow package
var (
	// General errors
	ErrFile                  = errors.New("file error")
	ErrState                 = errors.New("state error")
	ErrExecution             = errors.New("execution error")
	ErrValidation            = errors.New("validation error")
	
	// Step validation specific errors
	ErrStepMissingID         = errors.New("step missing ID")
	ErrStepMissingDescription = errors.New("step missing description")
	ErrStepMissingOutputFile = errors.New("step missing output file template")
	ErrStepInvalidPrompt     = errors.New("invalid prompt in step")
)

// Message templates for user-friendly output
// These are separate from the error variables to maintain user-friendly formatting
const (
	MsgFileNotFound            = "❌ Error: File %s not found."
	MsgInvalidStateFile        = "⚠️ Warning: Invalid state file detected for %s. Starting from the beginning."
	MsgStateUpdateFailed       = "❌ Error: Failed to update workflow state: %s"
	MsgStepExecutionFailed     = "❌ Error: Failed to execute step: %s"
	MsgUnrecognizedStep        = "⚠️ Warning: Unrecognized step in %s. Consider resetting the workflow with --reset."
	MsgStateFileCorrupted      = "⚠️ Warning: State file for %s appears to be corrupted. Starting from step 1."
	MsgOutputFileCreateFailed  = "❌ Error: Failed to create output file: %s"
	MsgInvalidPrompt           = "❌ Error: Invalid prompt in step %s: %s"
	MsgStepValidationFailed    = "❌ Error: Step validation failed: %s"
) 
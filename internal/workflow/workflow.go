// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// WorkflowStep represents a single step in the implementation workflow
type WorkflowStep struct {
	ID          string // Unique identifier (e.g., "01-laying-the-foundation")
	Description string // Human-readable description
	Prompt      string // AI agent instructions with variable interpolation
	OutputFile  string // Template for output filename
}

// WorkflowState tracks the current state of a workflow for a specific change request
type WorkflowState struct {
	ChangeRequestPath string    // Path to the change request file
	CurrentStepIndex  int       // Index of the current step (0-based)
	LastModified      time.Time // When the state was last updated
	CompletedSteps    []string  // List of completed step IDs
}

// WorkflowManager handles workflow-related operations
type WorkflowManager struct {
	fs FileSystem
	io UserOutput
}

// FileSystem defines the file system operations needed by the workflow manager
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Exists(path string) bool
}

// UserOutput defines the interface for displaying output to the user
type UserOutput interface {
	Print(message string)
	PrintSuccess(message string)
	PrintError(message string)
	PrintWarning(message string)
	PrintProgress(message string)
	PrintStep(stepNumber int, totalSteps int, description string)
	IsDebugEnabled() bool
}

// Error message templates
const (
	ErrFileNotFound            = "❌ Error: File %s not found."
	ErrInvalidStateFile        = "⚠️ Warning: Invalid state file detected for %s. Starting from the beginning."
	ErrStateUpdateFailed       = "❌ Error: Failed to update workflow state: %s"
	ErrStepExecutionFailed     = "❌ Error: Failed to execute step: %s"
	ErrUnrecognizedStep        = "⚠️ Warning: Unrecognized step in %s. Consider resetting the workflow with --reset."
	ErrStateFileCorrupted      = "⚠️ Warning: State file for %s appears to be corrupted. Starting from step 1."
	ErrOutputFileCreateFailed  = "❌ Error: Failed to create output file: %s"
	ErrNegativeStepIndex       = "invalid step index: negative value"
	ErrExceedingStepIndex      = "invalid step index: exceeds number of steps"
	ErrFailedToLoadState       = "failed to load state: %w"
	ErrInvalidPrompt         = "❌ Error: Invalid prompt in step %s: %s"
	ErrStepValidationFailed  = "❌ Error: Step validation failed: %s"
	ErrInvalidOutputTemplate = "❌ Error: Invalid output file template: %s"
)

// Success message templates
const (
	SuccessStepCompleted     = "✅ Completed step %d of %d: %s"
	SuccessWorkflowCompleted = "🎉 All steps completed successfully for change request: %s"
	SuccessStateReset        = "🔄 Workflow for %s has been reset to the beginning."
)

// Progress message templates
const (
	ProgressExecutingStep = "⏳ Executing step %s: %s"
	ProgressSavingState   = "💾 Saving workflow state..."
	ProgressValidating    = "🔍 Validating workflow state..."
)

// Output file path interpolation variables
const (
	// PathVarChangeRequestBasename is the basename of the change request file (without extension)
	PathVarChangeRequestBasename = "${change_request_basename}"
	// PathVarBlueprintBasename is the basename of the blueprint file (without extension)
	PathVarBlueprintBasename = "${blueprint_basename}"
	// PathVarDirname is the directory of the change request file
	PathVarDirname = "${dirname}"
	// PathVarStepID is the ID of the current step
	PathVarStepID = "${stepid}"
	// PathVarStepName is the name of the current step (derived from ID)
	PathVarStepName = "${stepname}"
	// PathVarFullpath is the full path of the change request file (without extension)
	PathVarFullpath = "${fullpath}"
	// PathVarTimestamp is the current timestamp (YYYYMMDD-HHMMSS)
	PathVarTimestamp = "${timestamp}"
	// Deprecated: Use PathVarChangeRequestBasename instead
	PathVarBasename = "${basename}"
)

// StandardWorkflowSteps defines the predefined sequence of steps in the implementation workflow
var StandardWorkflowSteps = []WorkflowStep{
	{
		ID:          "01-laying-the-foundation",
		Description: "Laying the foundation - Setting up the architecture and structure",
		Prompt:      `You are a senior software engineer about to begin a new iteration of software development based on a set of user stories described in a blueprint document. 

The whole iteration is divided into 4 phases:
- Laid the foundation (scaffoling the solution, placeholders, key abstractions)
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to lay the foundation—that is, to prepare the codebase to safely and effectively accommodate the upcoming changes.


This phase includes two core responsibilities:
🧱 1. Architecture & Design Setup
🛠️ 2. Refactoring / Re-architecting
🧪 Mandatory Testing Requirements

# Architecture & Design Setup
Analyze the blueprint and user stories to identify the new modules, services, or components that will be needed.
Design the structural layout of these new elements, even if they are not yet fully implemented.
Define key interfaces, APIs, class responsibilities, and high-level data flows.
Introduce placeholders (e.g., method stubs, empty classes, files) as needed to scaffold the system and ensure developers can start working on each part independently.

Deliverables:

- Skeletons of new modules, files, and class/functions definitions
- Use a lot of comments / TODO and dummy implementations 
- Basic integration points with existing code
- Updated diagrams or descriptions (if applicable)

Goal:
- To create a solid, extensible structure that reflects the target design and allows iterative implementation with minimal rework.

# Refactoring / Re-architecting
Assess the existing codebase for parts that block or conflict with the new blueprint.
Identify code smells, tight couplings, or brittle logic that could hinder future development.
Carefully refactor or re-architect areas that require cleanup or realignment to match the upcoming design.
Maintain behavior during this process—this is not about adding new features yet.

Deliverables:
- Refactored modules or components with improved structure
- Simplified interfaces or decoupled logic
- Regression tests to ensure no existing functionality breaks
- An accomplishment report in ${change_request_file_path}.01-foundation.accomplished.md

The accomplishment report is not a summary, it is a "compass" to the changes you made, hence avoid general statements/claim, be precise:
Use always short code references (no code at all, 
 just a compact/understable reference to lookup for, do not use line numbers) 
 as foundation of your statements
 For example:
 - Instead of "Added tests for ..." / "Updated tests for ... " show me which test case has been added (using code references)
 - Instead of "Message templates are now centralized with clear naming conventions" show me where to find them (using code references)
 - Include a section of "blind spot" if any: leverage test coverage report to reinforce your statements
 - Include a dedicated section for potentially still not yet well implemented acceptance criteria.
 - Include any changes to original design decisions

Goal:
- To reduce technical debt and align the current codebase with the structural needs of the upcoming iteration.

# Mandatory Testing Requirements
After refactoring or structural changes, run the full test suite to confirm no regressions have been introduced.
If major components are touched, consider adding or updating smoke/regression tests to validate the foundation work.

📘 Instructions
- Clearly document each architectural or structural decision, especially where existing components were modified.
- Leave TODOs or comments where further implementation will happen in later phases.
- Do not implement the full logic yet—this phase is strictly for setting up structure and enabling smooth feature delivery.

Now:
Read the user stories using ./cat-user-stories-in-change-request.sh ${change_request_file_path}
Read the blueprint using cat ${change_request_file_path}
		`,
		OutputFile:  "${dirname}/${change_request_basename}.01-laying-the-foundation.md",
	},
	{
		ID:          "01-laying-the-foundation-test",
		Description: "Laying the foundation testing - Verifying the foundational changes",
		Prompt:      "Ensure all the tests are passing for the foundational changes implemented based on the blueprint at ${change_request_file_path}. Verify that the structure is appropriate and tests are in place.",
		OutputFile:  "${dirname}/${change_request_basename}.01-laying-the-foundation-test.md",
	},
	{
		ID:          "02-mvi",
		Description: "Minimum Viable Implementation - Building the core functionality",
		Prompt:      `You are about to continue a development iteration of software based on a set of user stories described in a blueprint document. 

The whole iteration is divided into 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions)
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to build the **simplest working implementation** for each user story that satisfies its requirements and passes all associated tests.

---

## 🔁 Process: One User Story at a Time

### 1. Review the User Story

- Read the user story and its acceptance criteria from the blueprint.

### 2. Write Verification Code

- For each acceptance criterion, write a corresponding automated test (unit or integration).
- Ensure the test clearly reflects the criterion, is easy to run, and produces a reliable outcome.
- The absence of implementation should cause these tests to fail initially.

### 3. Implement the Minimum Logic

- Write the **simplest code** needed to satisfy the user story and make the tests pass.
- Avoid unnecessary generalizations, optimizations, or edge case handling at this stage.
- Stick closely to the logic suggested by the blueprint.

### 4. Run the Full Test Suite

- After implementing each user story, run the **entire test suite**.
- All tests—existing and newly added—must pass.
- Fix any issues immediately before proceeding to the next user story.

---

## 📌 Principles

- Keep the implementation minimal but correct.
- Build confidence through verification.
- Avoid building more than what's needed to pass the tests and meet the blueprint requirements.
- Defer enhancements and broader handling to future iterations.

---

## ✅ Final Check

After completing MVI for all user stories:

- Run the test suite again to confirm full functionality.
- Ensure each feature is backed by at least one clear, reliable test.
- Write a summary of what you've accomplished.

---

Read a set of user stories using the command: ./cat-user-stories-in-change-request.sh ${change_request_file_path}
Read the implementation plan using the command: cat ${change_request_file_path}
Read the "laying the foundation" accomplished summary using the command: ${change_request_file_path}.01-foundation.accomplished.md
 
Now build the MVI for each user story.

At the end of your task write the summary of what you accomplished in ${change_request_file_path}.02-mvi.accomplished.md
Ensure to include a user story implementation section:
- in this section I'd like to have an easy way to check each acceptance criterion. I rely only on "facts". Please add explicit reference (no code at all, just a compact/understable reference to lookup for) to which test ensure that criterion is met. If no test was written about that specific criterion, mention it.
`,
		OutputFile:  "${dirname}/${change_request_basename}.02-mvi.md",
	},
	{
		ID:          "02-mvi-test",
		Description: "Minimum Viable Implementation testing - Verifying the core functionality",
		Prompt:      "Ensure all the tests are passing for the minimum viable implementation based on the blueprint at ${change_request_file_path}. Ensure all basic functionality works as expected.",
		OutputFile:  "${dirname}/${change_request_basename}.02-mvi-test.md",
	},
	{
		ID:          "03-extend-functionalities",
		Description: "Extending functionalities - Adding additional features and improvements",
		Prompt:      `You are about to continue a development iteration of software based on a set of user stories described in a blueprint document. 

The whole iteration is divided into 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions)
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

The initial blueprint is: ${change_request_file_path}
Retrieve the user stories mentioned in the blueprint are:
- ./cat-user-stories-in-change-request.sh ${change_request_file_path}

You have already:
- Laid the groundwork by scaffolding the solution and defining high-level architecture: read it using: cat 2025-03-31-081819-introduce-step-prompt.blueprint.md.01-foundation.accomplished.md
- Implemented a Minimal Viable Implementation (MVI) that satisfies the basic functionality required to pass the initial test suite: read it using: cat docs/changes-request/2025-03-31-081819-introduce-step-prompt.blueprint.md.02-mvi.accomplished.md

### 🎯 Goals of this Phase

- Extend the core logic to handle **all scenarios** described in the user stories and their acceptance criteria.
- Add meaningful logic to improve completeness while maintaining modularity.
- Update or create **new tests** to ensure coverage of extended functionality.

### 🧪 Formal Test Execution – Mandatory

At the end of this phase, you **must run the complete formal test suite**:
- All existing and new tests **must pass**.
- Add tests for any uncovered edge cases.
- Document any remaining limitations or areas needing future refinement.

### ✅ Guidelines

- Maintain alignment with the blueprint's structure.
- Keep the code production-quality: clear, modular, and documented.
- Refactor as needed to support new logic, but avoid premature optimization.
- Do not yet focus on performance tuning or polishing—prioritize completeness and correctness.

### 🛠️ Your task

1. For each user story:
   - Validate the acceptance criteria already implemented.
   - Review the acceptance criteria not yet implemented.
   - Identify edge cases and secondary scenarios.
   - Extend the implementation accordingly.

2. For each new case:
   - Add or update tests.
   - Ensure the verification logic (e.g., test assertions) remains aligned and meaningful for the user story.

3. Run the full test suite and validate all tests pass.
4. At the end of your task write the summary of what you accomplished in ${change_request_file_path}.03-extend-functionalities.accomplished.md.

The accomplishment report is not a summary, is a compass to the changes you made, hence avoid general statements/claim, be precise:
Use always short code references (no code at all, 
 just a compact/understable reference to lookup for, do not use line numbers) 
 as foundation of your statements
 For example:
 - Instead of "Added tests for ..." / "Updated tests for ... " show me which test case has been added (using code references)
 - Instead of "Message templates are now centralized with clear naming conventions" show me where to find them (using code references)
 - Include a section of "blind spot" if any: leverage test coverage report to reinforce your statements
 - Include a dedicated section for potentially still not yet well covered acceptance criteria.

--- 


Your task now is to proceed to **expand the implementation** to cover additional use cases, edge cases, and deferred features, as described in the blueprint.`,
		OutputFile:  "${dirname}/${change_request_basename}.03-extend-functionalities.md",
	},
	{
		ID:          "03-extend-functionalities-test",
		Description: "Extending functionalities testing - Verifying the additional features",
		Prompt:      "Ensure all the tests are passing for the extended functionality implemented based on the blueprint at ${change_request_file_path}. Verify all features work correctly.",
		OutputFile:  "${dirname}/${change_request_basename}.03-extend-functionalities-test.md",
	},
	{
		ID:          "04-final-iteration",
		Description: "Final iteration - Polishing and final adjustments",
		Prompt:      `Read a set of user stories using the command: ./cat-user-stories-in-change-request.sh ${change_request_file_path}

You have already:
- Laid the foundation (project structure, placeholders, key abstractions): cat ${change_request_file_path}.01-foundation.accomplished.md
- Completed the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria: cat ${change_request_file_path}.02-mvi.accomplished.md
- Extended the implementation to support more scenarios and edge cases: cat ${change_request_file_path}.03-extend-functionalities.accomplished.md

Now, execute the last phase of the iteration: **Refinement & Stabilization**.

### 🎯 Objectives:
- Refine the codebase for clarity, maintainability, and performance
- Enhance test coverage to simulate real-world usage
- Ensure robustness through thorough validation
- Finalize the iteration by producing production-quality code and test suites

---

### 🧩 Tasks to Perform:

1. **Refactor for Maintainability:**
   - Improve naming, structure, and modularity
   - Remove duplication, unused code, and reduce complexity
   - Ensure clear separation of concerns and adherence to clean code principles

2. **Optimize for Performance (if applicable):**
   - Profile critical paths
   - Optimize data structures and algorithms for efficiency
   - Avoid premature optimization — focus on known bottlenecks or risky parts

3. **Enhance Test Coverage:**
   - Add tests that simulate real-world edge cases and usage patterns
   - Ensure each user story and its acceptance criteria are covered
   - Include tests for:
     - Error handling and invalid inputs
     - Performance-sensitive areas
     - Integration between components

4. **Stabilize the Codebase:**
   - Resolve any known issues or inconsistencies
   - Finalize API boundaries and expected behaviors
   - Ensure the implementation is resilient and ready for review or release

---

### ✅ Mandatory Before Completion:

- **Ensure 100% coverage** of acceptance criteria from the blueprint
- **Perform a final code review** with an eye on polish and stability
- **Document any deviations** from the blueprint and rationale for changes

---

### 📝 Output Requirements:

- Final version of the code with inline comments
- Updated and complete test suite
- An accomplishment report in ${change_request_file_path}.04-refinement.accomplished.md

The accomplishment report is not a summary, it is a "compass" to the changes you made, hence avoid general statements/claim, be precise:
Use always short code references (no code at all, 
 just a compact/understable reference to lookup for, do not use line numbers) 
 as foundation of your statements
 For example:
 - Instead of "Added tests for ..." / "Updated tests for ... " show me which test case has been added (using code references)
 - Instead of "Message templates are now centralized with clear naming conventions" show me where to find them (using code references)
 - Include a section of "blind spot" if any: leverage test coverage report to reinforce your statements
 - Include a dedicated section for potentially still not yet well implemented acceptance criteria.
 - Include any changes to original design decisions
---

### ⚠️ Reminder:
Do not introduce new features at this stage. Focus only on refining and stabilizing the existing work to make it reliable and production-ready.

Proceed with the **Refinement & Stabilization** phase now.
`,
		OutputFile:  "${dirname}/${change_request_basename}.04-final-iteration.md",
	},
	{
		ID:          "04-final-iteration-test",
		Description: "Final iteration testing - Final verification and validation",
		Prompt:      "Ensure all the tests are passing for the final iteration based on the blueprint at ${change_request_file_path}. Ensure all requirements are met.",
		OutputFile:  "${dirname}/${change_request_basename}.04-final-iteration-test.md",
	},{
		ID:          "implementation",
		Description: "The implementation report of the change request",
		Prompt:      `You are a technical writer that needs to document the implementation of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration was composed by 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions)
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to document what has been accomplished so far in the file ${change_request_file_path}.implementation.md

Keep into consideration that this technical document will be archived as Architecture Decision Record. Use reference to the implemented code to look up for (do not use line numbers, they change too frequently)
Describe data structures and their purposes (if any)
Describe relevant algorithms and their purporses (if any)
 
The original blueprint: ./usm ${change_request_file_path}
the accomplished reports:
- foundations: cat ${change_request_file_path}.01-foundation.accomplished.md
- mvi: cat ${change_request_file_path}.02-mvi.accomplished.md
- extend: cat ${change_request_file_path}.03-extend-functionalities.accomplished.md 
- refine: cat ${change_request_file_path}.04-refinement.accomplished.md   
`,
		OutputFile:  "${dirname}/${change_request_basename}.implementation.md",
	},
}

// NewWorkflowManager creates a new workflow manager instance
func NewWorkflowManager(fs FileSystem, io UserOutput) *WorkflowManager {
	return &WorkflowManager{
		fs: fs,
		io: io,
	}
}

// GenerateStateFilePath generates the path for the state file based on the change request path
func GenerateStateFilePath(changeRequestPath string) string {
	dir := filepath.Dir(changeRequestPath)
	base := filepath.Base(changeRequestPath)
	return filepath.Join(dir, "."+base+".step")
}

// LoadState loads the workflow state from the state file
func (wm *WorkflowManager) LoadState(changeRequestPath string) (WorkflowState, error) {
	state := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
		CompletedSteps:    []string{},
	}

	stateFilePath := GenerateStateFilePath(changeRequestPath)
	if !wm.fs.Exists(stateFilePath) {
		return state, nil
	}

	// Only print progress message in debug mode
	if wm.io.IsDebugEnabled() {
		wm.io.PrintProgress(ProgressValidating)
	}

	data, err := wm.fs.ReadFile(stateFilePath)
	if err != nil {
		// Only print warning in debug mode
		if wm.io.IsDebugEnabled() {
			wm.io.PrintWarning(fmt.Sprintf(ErrStateFileCorrupted, changeRequestPath))
		}
		return state, err
	}

	if err := json.Unmarshal(data, &state); err != nil {
		// Only print warning in debug mode
		if wm.io.IsDebugEnabled() {
			wm.io.PrintWarning(fmt.Sprintf(ErrInvalidStateFile, changeRequestPath))
		}
		return state, err
	}

	// Validate the state
	if state.CurrentStepIndex < 0 || state.CurrentStepIndex > len(StandardWorkflowSteps) {
		// Only print warning in debug mode
		if wm.io.IsDebugEnabled() {
			wm.io.PrintWarning(fmt.Sprintf(ErrUnrecognizedStep, stateFilePath))
		}
		state.CurrentStepIndex = 0
		state.CompletedSteps = []string{}
	}

	return state, nil
}

// SaveState saves the workflow state to the state file
func (wm *WorkflowManager) SaveState(state WorkflowState) error {
	// Only print progress message in debug mode
	if wm.io.IsDebugEnabled() {
		wm.io.PrintProgress(ProgressSavingState)
	}
	
	state.LastModified = time.Now()
	
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf(ErrStateUpdateFailed, err)
	}
	
	stateFilePath := GenerateStateFilePath(state.ChangeRequestPath)
	if err := wm.fs.WriteFile(stateFilePath, data, 0644); err != nil {
		return fmt.Errorf(ErrStateUpdateFailed, err)
	}
	
	return nil
}

// DetermineNextStep determines the next step to execute based on the state
func (wm *WorkflowManager) DetermineNextStep(changeRequestPath string) (int, error) {
	// Only print progress message in debug mode
	if wm.io.IsDebugEnabled() {
		wm.io.PrintProgress(ProgressValidating)
	}
	
	state, err := wm.LoadState(changeRequestPath)
	if err != nil {
		// Only print warning in debug mode
		if wm.io.IsDebugEnabled() {
			wm.io.PrintWarning(fmt.Sprintf(ErrInvalidStateFile, changeRequestPath))
		}
		return 0, nil // Still start from beginning despite the error
	}

	// If we've completed all steps, return a special indicator
	if state.CurrentStepIndex >= len(StandardWorkflowSteps) {
		// Only print success in debug mode
		if wm.io.IsDebugEnabled() {
			wm.io.PrintSuccess(fmt.Sprintf(SuccessWorkflowCompleted, changeRequestPath))
		}
		return -1, nil
	}

	// Print current step information only in debug mode
	if wm.io.IsDebugEnabled() {
		wm.io.PrintStep(state.CurrentStepIndex+1, len(StandardWorkflowSteps), StandardWorkflowSteps[state.CurrentStepIndex].Description)
	}
	
	return state.CurrentStepIndex, nil
}

// UpdateState updates the workflow state after completing a step
func (wm *WorkflowManager) UpdateState(changeRequestPath string, newStepIndex int) error {
	// Only print progress message in debug mode
	if wm.io.IsDebugEnabled() {
		wm.io.PrintProgress(ProgressSavingState)
	}
	
	state, err := wm.LoadState(changeRequestPath)
	if err != nil {
		return fmt.Errorf(ErrStateUpdateFailed, err)
	}

	// Validate new step index
	if newStepIndex < 0 {
		return fmt.Errorf(ErrStateUpdateFailed, ErrNegativeStepIndex)
	}

	if newStepIndex > len(StandardWorkflowSteps) {
		return fmt.Errorf(ErrStateUpdateFailed, ErrExceedingStepIndex)
	}

	// Update the state
	state.CurrentStepIndex = newStepIndex
	
	// Update completed steps
	state.CompletedSteps = make([]string, 0, newStepIndex)
	for i := 0; i < newStepIndex; i++ {
		if i < len(StandardWorkflowSteps) {
			state.CompletedSteps = append(state.CompletedSteps, StandardWorkflowSteps[i].ID)
		}
	}
		
	// Print success message for the completed step only in debug mode
	if wm.io.IsDebugEnabled() {
		if newStepIndex > 0 && newStepIndex <= len(StandardWorkflowSteps) {
			completedStep := StandardWorkflowSteps[newStepIndex-1]
			wm.io.PrintSuccess(fmt.Sprintf(SuccessStepCompleted, newStepIndex, len(StandardWorkflowSteps), completedStep.Description))
		}
	}

	// Save the updated state
	return wm.SaveState(state)
}

// GenerateOutputFilename generates the output filename for a step
func (wm *WorkflowManager) GenerateOutputFilename(changeRequestPath string, step WorkflowStep) string {
	dir := filepath.Dir(changeRequestPath)
	base := filepath.Base(changeRequestPath)
	
	// Remove the .blueprint.md extension if present
	base = strings.TrimSuffix(base, ".blueprint.md")
	fullpath := filepath.Join(dir, base)
	
	// Extract step name from ID (e.g., "01-laying-the-foundation" -> "laying-the-foundation")
	stepName := step.ID
	if parts := strings.SplitN(step.ID, "-", 2); len(parts) > 1 {
		stepName = parts[1]
	}
	
	// Generate timestamp
	timestamp := time.Now().Format("20060102-150405")
	
	// Check for legacy format with %s first
	if strings.Contains(step.OutputFile, "%s") {
		filename := fmt.Sprintf(step.OutputFile, base)
		return filepath.Join(dir, filename)
	}
	
	// Replace variables in the output file template
	outputFile := step.OutputFile
	// Support both new and deprecated variables for backward compatibility
	outputFile = strings.ReplaceAll(outputFile, PathVarChangeRequestBasename, base)
	outputFile = strings.ReplaceAll(outputFile, PathVarBlueprintBasename, base)
	outputFile = strings.ReplaceAll(outputFile, PathVarBasename, base) // For backward compatibility
	outputFile = strings.ReplaceAll(outputFile, PathVarDirname, dir) 
	outputFile = strings.ReplaceAll(outputFile, PathVarStepID, step.ID)
	outputFile = strings.ReplaceAll(outputFile, PathVarStepName, stepName)
	outputFile = strings.ReplaceAll(outputFile, PathVarFullpath, fullpath)
	outputFile = strings.ReplaceAll(outputFile, PathVarTimestamp, timestamp)
	
	// If outputFile already contains a path separator, assume it's a relative or absolute path
	if strings.Contains(outputFile, string(filepath.Separator)) {
		return outputFile
	}
	
	// Otherwise, join with the directory of the change request
	return filepath.Join(dir, outputFile)
}

// IsWorkflowComplete checks if all workflow steps have been completed
func (wm *WorkflowManager) IsWorkflowComplete(changeRequestPath string) (bool, error) {
	state, err := wm.LoadState(changeRequestPath)
	if err != nil {
		return false, fmt.Errorf("failed to load state: %w", err)
	}

	return state.CurrentStepIndex >= len(StandardWorkflowSteps), nil
}

// ResetWorkflow resets the workflow to the beginning
func (wm *WorkflowManager) ResetWorkflow(changeRequestPath string) error {
	state := WorkflowState{
		ChangeRequestPath: changeRequestPath,
		CurrentStepIndex:  0,
		LastModified:      time.Now(),
		CompletedSteps:    []string{},
	}
	
	if err := wm.SaveState(state); err != nil {
		return err
	}
	
	// Only show success message in debug mode
	if wm.io.IsDebugEnabled() {
		wm.io.PrintSuccess(fmt.Sprintf(SuccessStateReset, changeRequestPath))
	}
	return nil
}

// ValidateWorkflowSteps validates all steps in a workflow
func (wm *WorkflowManager) ValidateWorkflowSteps(steps []WorkflowStep) []error {
	var errors []error
	
	for _, step := range steps {
		// Validate that required fields are present
		if step.ID == "" {
			errors = append(errors, fmt.Errorf("step missing ID"))
			continue
		}
		
		if step.Description == "" {
			errors = append(errors, fmt.Errorf("step %s missing description", step.ID))
		}
		
		if step.OutputFile == "" {
			errors = append(errors, fmt.Errorf("step %s missing output file template", step.ID))
		} else if err := ValidateOutputTemplate(step.OutputFile); err != nil {
			errors = append(errors, fmt.Errorf("step %s has invalid output file template: %w", step.ID, err))
		}
		
		// Validate prompt if present
		if step.Prompt != "" {
			if err := ValidatePrompt(step.Prompt); err != nil {
				errors = append(errors, fmt.Errorf("step %s has invalid prompt: %w", step.ID, err))
			}
		}
	}
	
	return errors
}

// ValidateOutputTemplate validates an output file template
func ValidateOutputTemplate(template string) error {
	// Check for format string for backward compatibility
	if strings.Contains(template, "%s") {
		// Allow the old format for now
		return nil
	}
	
	// Check for any other problematic patterns
	if strings.Contains(template, "${") && !strings.Contains(template, "}") {
		return fmt.Errorf("mismatched variable braces in template: %s", template)
	}
	
	// Define the list of valid variable names
	validVars := getValidVariableNames()
	
	// Extract all variable patterns
	varPattern := regexp.MustCompile(`\${([^}]+)}`)
	matches := varPattern.FindAllStringSubmatch(template, -1)
	
	// Check each variable against the valid list
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			valid := false
			for _, validVar := range validVars {
				if varName == validVar {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("unknown variable in template: ${%s}", varName)
			}
		}
	}
	
	return nil
}

// getValidVariableNames returns a list of valid variable names for interpolation
func getValidVariableNames() []string {
	return []string{
		// Path variables
		strings.TrimPrefix(strings.TrimSuffix(PathVarChangeRequestBasename, "}"), "${"),
		strings.TrimPrefix(strings.TrimSuffix(PathVarBlueprintBasename, "}"), "${"),
		strings.TrimPrefix(strings.TrimSuffix(PathVarDirname, "}"), "${"),
		strings.TrimPrefix(strings.TrimSuffix(PathVarStepID, "}"), "${"),
		strings.TrimPrefix(strings.TrimSuffix(PathVarStepName, "}"), "${"),
		strings.TrimPrefix(strings.TrimSuffix(PathVarFullpath, "}"), "${"),
		strings.TrimPrefix(strings.TrimSuffix(PathVarTimestamp, "}"), "${"),
		strings.TrimPrefix(strings.TrimSuffix(PathVarBasename, "}"), "${"), // Deprecated
		
		// Prompt variables
		"change_request_file_path",
	}
} 
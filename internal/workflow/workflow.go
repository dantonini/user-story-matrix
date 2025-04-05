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
	"time"
)

// WorkflowStep represents a single step in the implementation workflow
type WorkflowStep struct {
	ID          string // Unique identifier (e.g., "01-laying-the-foundation")
	Description string // Human-readable description
	Prompt      string // AI agent instructions with variable interpolation
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
	ErrNegativeStepIndex       = "invalid step index: negative value"
	ErrExceedingStepIndex      = "invalid step index: exceeds number of steps"
	ErrFailedToLoadState       = "failed to load state: %w"
	ErrInvalidPrompt           = "❌ Error: Invalid prompt in step %s: %s"
	ErrStepValidationFailed    = "❌ Error: Step validation failed: %s"
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

// StandardWorkflowSteps defines the predefined sequence of steps in the implementation workflow
var StandardWorkflowSteps = []WorkflowStep{
	{
		ID:          "01-laying-the-foundation",
		Description: "Laying the foundation - Setting up the architecture and structure",
		Prompt:      `You are a senior software engineer about to begin a new iteration of software development based on a set of user stories described in a blueprint document. 

The whole iteration is divided into 4 phases:
- Laid the foundation (sketch the solution, placeholders, key abstractions)
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
Design the structural layout of these new elements, they need to be sketched with minimal implementation details.
Define key interfaces, APIs, class responsibilities, and high-level data flows.
Introduce placeholders (e.g., method stubs, empty classes, files) as needed to sketch the solution and ensure developers can start working on each part independently.

Deliverables:
- Skeletons of new modules, files, and class/functions definitions
- Use a lot of comments / TODO and dummy implementations to guide the future implementation
- Basic integration points with existing code
- Updated diagrams or descriptions (if applicable)
- Do not implement any real logic, your functions should be stubs that will be implemented in the next phase.

Goal:
- To create a solid, extensible structure that reflects the target design and allows iterative implementation with minimal rework.

# Refactoring / Re-architecting
This phase is about to prepare the codebase to make it easier to implement the upcoming changes.
Hence we need to:
- Assess the existing codebase for parts that block or conflict with the new blueprint.
- Identify code smells, tight couplings, or brittle logic that could hinder future development.
- Carefully refactor or re-architect areas that require cleanup or realignment to match the upcoming design.
- Maintain existing behavior during this process — this is not about adding new features yet.

Deliverables:
- Refactored modules or components with improved structure
- Simplified interfaces or decoupled logic
- Regression tests to ensure no existing functionality breaks

Now, let's start the work:
- Read the user stories using ./usm cat ${change_request_file_path}
- Read the blueprint using cat ${change_request_file_path}
		`,
	},{
		ID:          "01-laying-the-foundation-accomplished",
		Description: "Laying the foundation accomplished - Summary of the changes",
		Prompt:      `You are a technical writer that needs to document the implementation of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration was composed by 4 phases:
- Laid the foundation (sketch the solution, placeholders, key abstractions) [Current step]
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to document what has been accomplished so far in the file ${change_request_file_path}.01-foundation.accomplished.md

The accomplishment report is not a summary, it is a "compass" to the changes you made, hence avoid general statements/claim, be precise:
Use always short code references (no code at all, just a compact/understable reference to lookup for, do not use line numbers) as foundation of your statements
For example:
- Instead of "Added tests for ..." / "Updated tests for ... " show me which test case has been added (using code references)
- Instead of "Message templates are now centralized with clear naming conventions" show me where to find them (using code references)
- Include a section of "blind spot" if any: leverage test coverage report to reinforce your statements
- Include a dedicated section for potentially still not yet well implemented acceptance criteria.
- Include any changes to original design decisions

Goal:
- Describe the changes you made to the codebase to prepare it for the upcoming changes.
- Describe how the subsequent phases should be executed relying on the current state of the codebase.

Now, let's start the work:
- Read the user stories using ./usm cat ${change_request_file_path}
- Read the blueprint using cat ${change_request_file_path}`,
	},{
		ID:          "01-laying-the-foundation-test",
		Description: "Laying the foundation test - Verifying the structural changes",
		Prompt:      `You are a senior software engineer that is working on a software iteration based on a set of user stories described in a blueprint document. 

Some structural changes has been made to the codebase, now your task is:
- Run the full test suite to confirm no regressions have been introduced.
- If major components are touched, consider adding or updating smoke/regression tests to validate the foundation work.

📘 Instructions
- Clearly document each architectural or structural decision, especially where existing components were modified.
- Leave TODOs or comments where further implementation will happen in later phases.
- Do not implement any real logic, your functions should be stubs that will be implemented in the next phase.

Now, let's start the work:
- Read the user stories using ./usm cat ${change_request_file_path}
- Read the blueprint using cat ${change_request_file_path}`,
	},
	{
		ID:          "01-make-lint",
		Description: "Make lint - Ensure the codebase is linted and formatted",
		Prompt:      "Execute the command: make lint. Ensure to fix all the linter issues.",
	},{
		ID:          "01-make-coverage",
		Description: "Make coverage - Ensure the codebase is covered by tests",
		Prompt:      "Execute the command: 'make coverage && ./coverage' and ensure to determine the coverage percentage",
	},{
		ID:          "01-make-coverage-report",
		Description: "Make coverage report - Update the accomplishment report",
		Prompt:      "Update the accomplishment report ${change_request_file_path}.01-foundation.accomplished.md with the new coverage percentage.",
	},
	{
		ID:          "02-mvi",
		Description: "Minimum Viable Implementation - Building the core functionality",
		Prompt:      `You are a software engineer about to continue a development iteration of software based on a set of user stories described in a blueprint document. 

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

Now build the MVI for each user story:
- Read a set of user stories using the command: ./usm cat ${change_request_file_path}
- Read the implementation plan using the command: cat ${change_request_file_path}
- Read the "laying the foundation" accomplished summary using the command: cat ${change_request_file_path}.01-foundation.accomplished.md
`,
	},
	{
		ID:          "02-mvi-accomplished",
		Description: "Minimum Viable Implementation accomplished - Summary of the changes",
		Prompt:      `You are a technical writer that needs to document the implementation of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration was composed by 4 phases:
- Laid the foundation (sketch the solution, placeholders, key abstractions)
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria [Current step]
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to document what has been accomplished so far in the file ${change_request_file_path}.02-mvi.accomplished.md

The accomplishment report is not a summary, it is a "compass" to the changes you made, hence avoid general statements/claim, be precise:
Use always short code references (no code at all, just a compact/understable reference to lookup for, do not use line numbers) as foundation of your statements
For example:
- Instead of "Added tests for ..." / "Updated tests for ... " show me which test case has been added (using code references)
- Instead of "Message templates are now centralized with clear naming conventions" show me where to find them (using code references)
- Include a section of "blind spot" if any: leverage test coverage report to reinforce your statements
- Include a dedicated section for potentially still not yet well implemented acceptance criteria.
- Include any changes to original design decisions

Goal:
- Describe the changes you made to the codebase to prepare it for the upcoming changes.
- Describe how the subsequent phases should be executed relying on the current state of the codebase.

Now, let's start the work:
- Read the user stories using ./usm cat ${change_request_file_path}
- Read the blueprint using cat ${change_request_file_path}
- Read the "laying the foundation" accomplished summary using the command: cat ${change_request_file_path}.01-foundation.accomplished.md
`,
	},{
		ID:          "02-mvi-test",
		Description: "Minimum Viable Implementation test - Verifying the core functionality",
		Prompt:      `You are a senior software engineer that is working on a software iteration based on a set of user stories described in a blueprint document. 

The MVI has been built, now your task is:
- Run the full test suite to confirm no regressions have been introduced.
- If major components are touched, consider adding or updating smoke/regression tests to validate the foundation work.

📘 Instructions
- Clearly document each architectural or structural decision, especially where existing components were modified.
- Leave TODOs or comments where further implementation will happen in later phases.
- Do not implement any real logic, your functions should be stubs that will be implemented in the next phase.

Now, let's start the work:
- Read the user stories using ./usm cat ${change_request_file_path}
- Read the blueprint using cat ${change_request_file_path}
- Read the "laying the foundation" accomplished summary using the command: cat ${change_request_file_path}.01-foundation.accomplished.md
`,
	},
	{
		ID:          "02-mvi-lint",
		Description: "MVI lint - Ensure the codebase is linted and formatted",
		Prompt:      "Execute the command: make lint. Ensure to fix all the linter issues.",
	},{
		ID:          "02-mvi-coverage",
		Description: "MVI coverage - Ensure the codebase is covered by tests",
		Prompt:      "Execute the command: 'make coverage && ./coverage' and ensure to increse the coverage percentage for the MVI new code.",
	},{
		ID:          "02-mvi-coverage-report",
		Description: "MVI coverage report - Update the accomplishment report",
		Prompt:      "Update the accomplishment report ${change_request_file_path}.02-mvi.accomplished.md with the new coverage percentage.",
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
- ./usm cat ${change_request_file_path}

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
	},
	{
		ID:          "03-extend-functionalities-test",
		Description: "Extending functionalities testing - Verifying the additional features",
		Prompt:      "Ensure all the tests are passing for the extended functionality implemented based on the blueprint at ${change_request_file_path}. Verify all features work correctly.",
	},
	{
		ID:          "04-final-iteration",
		Description: "Final iteration - Polishing and final adjustments",
		Prompt:      `Read a set of user stories using the command: ./usm cat ${change_request_file_path}

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
	},
	{
		ID:          "04-final-iteration-test",
		Description: "Final iteration testing - Final verification and validation",
		Prompt:      "Ensure all the tests are passing for the final iteration based on the blueprint at ${change_request_file_path}. Ensure all requirements are met.",
	},{
		ID:          "implementation",
		Description: "The implementation report of the change request",
		Prompt:      `You are a technical writer that needs to document the implementation of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration was composed by 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions)
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to document what has been accomplished so far in the file ${change_request_dirname}/${change_request_basename}.implementation.md

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
		
		// Validate prompt if present
		if step.Prompt != "" {
			if err := ValidatePrompt(step.Prompt); err != nil {
				errors = append(errors, fmt.Errorf("step %s has invalid prompt: %w", step.ID, err))
			}
		}
	}
	
	return errors
} 
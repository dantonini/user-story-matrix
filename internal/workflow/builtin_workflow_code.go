// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

// UsmCodeWorkflowName is the name of the code workflow
const UsmCodeStandardWorkflowName = "usm-code"

// UsmCodeStandardWorkflowSteps defines the predefined sequence of steps in the implementation workflow
// Deprecated: Use WorkflowRegistry.GetWorkflow("standard") instead.
// This global variable is maintained for backward compatibility with existing code.
// It will be synchronized with the registry's standard workflow to ensure consistency.
// Direct usage will be removed in a future version.
var UsmCodeStandardWorkflowSteps = []WorkflowStep{
	{
		ID:          "00-read-prompt-instruction",
		Description: "Read the detailed prompt instruction for the change request",
		Prompt:      `Your task is to produce a blueprint file for the change request.

A blueprint is a technical design document that outlines proposed codebase changes without actual implementation. It helps:
- Understand the proposed changes before coding
- Create a clear roadmap for upcoming development tasks

General Guidelines:
- The blueprint has a metadata section referencing a set of user stories. Each user story has a title and a filename. Read all the user stories at once using the command:
	./usm cat {{.ChangeRequestFilePath}}
- The document is not for writing code but for transmitting ideas, concepts, and plans.
- Follow a top-down (or break-down) approach: start with a high-level overview and progressively drill down into specifics.

# Overview
**Purpose:**  
Provide a brief summary that captures the essence of all user stories.  
- Highlight common themes and relationships among the user stories.
- Summarize overall objectives without detailing individual acceptance criteria.

## Foudamentals
**Purpose:**  
Outline the key technical concepts necessary to address the user stories:
- **Data Structures:** Define any high-level data structures, including their purposes.
- **Algorithms:** Describe key algorithms using pseudo-code, outlining their intended functionality.
- **Refactoring Strategy:** Summarize any broad refactoring plans for the existing codebase.

# How to verify – Detailed User Story Breakdown
**Purpose:**  
For each user story, detail how the changes will be verified:
- **Acceptance Criteria:** Break down each user story into its individual acceptance criteria.
- **Testing Scenarios:** For each criterion, provide clear, concise testing scenarios that are tangible and automatable.
- **Bottom-Up Detailing:** Start with basic criteria and work toward more complex conditions.

# What is the Plan – Detailed Action Items
**Purpose:**  
For each user story, outline a detailed plan for what needs to be done. Take into account the user story verification process described earlier so to make the verification process easy to implement.
- **Task Breakdown:** Describe each implementation step without writing actual code.
- **Specific Data Structures:** List any data structures that need to be defined or modified, along with their purposes.
- **Specific Algorithms:** Provide pseudo-code for any specific algorithms, explaining their function.
- **Targeted Refactoring:** Detail any precise refactoring steps required for the existing codebase.
- **Validation:** Ensure the plan is validated against the current codebase, ensuring feasibility and completeness.

**Note:**  
Remember, the blueprint should be a planning and communication tool. Do not include any actual code – only high-level pseudo-code and detailed action items that make the subsequent verification and development process straightforward.

validate them against the codebase, and define a detailed plan for the change. Don't do any implementation, just describe what needs to be done. You can describe data structures, algorithm in pseudo code, refactoring steps, etc. 
Store the plan in the change request file {{.ChangeRequestFilePath}} in markdown format in a section called \"Blueprint\". Ensure to include the steps required to satisfy the acceptance criteria of all mentioned user stories.
		`,
	},
	{
		ID:          "01-laying-the-foundation",
		Description: "Laying the foundation - Setting up the architecture and structure",
		Prompt: `You are a senior software engineer about to begin a new iteration of software development based on a set of user stories described in a blueprint document. 

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
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
		`,
	}, {
		ID:          "01-laying-the-foundation-accomplished",
		Description: "Laying the foundation accomplished - Summary of the changes",
		Prompt: `You are a technical writer that needs to document the implementation of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration was composed by 4 phases:
- Laid the foundation (sketch the solution, placeholders, key abstractions) [Current step]
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to document what has been accomplished so far in the file {{.ChangeRequestFilePath}}.01-foundation.accomplished.md

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
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
`,
	}, {
		ID:          "01-laying-the-foundation-test",
		Description: "Laying the foundation test - Verifying the structural changes",
		Prompt: `You are a senior software engineer that is working on a software iteration based on a set of user stories described in a blueprint document. 

Some structural changes has been made to the codebase, now your task is:
- Run the full test suite to confirm no regressions have been introduced.
- If major components are touched, consider adding or updating smoke/regression tests to validate the foundation work.

📘 Instructions
- Clearly document each architectural or structural decision, especially where existing components were modified.
- Leave TODOs or comments where further implementation will happen in later phases.
- Do not implement any real logic, your functions should be stubs that will be implemented in the next phase.

Now, let's start the work:
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
`,
	},
	{
		ID:          "01-make-lint",
		Description: "Make lint - Ensure the codebase is linted and formatted",
		Prompt:      "Execute the command: make lint. Ensure to fix all the linter issues.",
	}, {
		ID:          "01-make-coverage",
		Description: "Make coverage - Ensure the codebase is covered by tests",
		Prompt:      "Execute the command: 'make coverage && ./coverage' and ensure to determine the coverage percentage",
	}, {
		ID:          "01-make-coverage-report",
		Description: "Make coverage report - Update the accomplishment report",
		Prompt:      "Update the accomplishment report {{.ChangeRequestFilePath}}.01-foundation.accomplished.md with the new coverage percentage.",
	},
	{
		ID:          "02-mvi",
		Description: "Minimum Viable Implementation - Building the core functionality",
		Prompt: `You are a software engineer about to continue a development iteration of software based on a set of user stories described in a blueprint document. 

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
- Read a set of user stories using the command: ./usm cat {{.ChangeRequestFilePath}}
- Read the implementation plan using the command: cat {{.ChangeRequestFilePath}}
- Read the "laying the foundation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.01-foundation.accomplished.md
`,
	},
	{
		ID:          "02-mvi-accomplished",
		Description: "Minimum Viable Implementation accomplished - Summary of the changes",
		Prompt: `You are a technical writer that needs to document the implementation of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration was composed by 4 phases:
- Laid the foundation (sketch the solution, placeholders, key abstractions)
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria [Current step]
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to document what has been accomplished so far in the file {{.ChangeRequestFilePath}}.02-mvi.accomplished.md

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
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
- Read the "laying the foundation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.01-foundation.accomplished.md
`,
	}, {
		ID:          "02-mvi-test",
		Description: "Minimum Viable Implementation test - Verifying the core functionality",
		Prompt: `You are a senior software engineer that is working on a software iteration based on a set of user stories described in a blueprint document. 

The MVI has been built, now your task is:
- Run the full test suite to confirm no regressions have been introduced.
- If major components are touched, consider adding or updating smoke/regression tests to validate the foundation work.

📘 Instructions
- Clearly document each architectural or structural decision, especially where existing components were modified.
- Leave TODOs or comments where further implementation will happen in later phases.
- Do not implement any real logic, your functions should be stubs that will be implemented in the next phase.

Now, let's start the work:
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
- Read the "laying the foundation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.01-foundation.accomplished.md
`,
	},
	{
		ID:          "02-mvi-lint",
		Description: "MVI lint - Ensure the codebase is linted and formatted",
		Prompt:      "Execute the command: make lint. Ensure to fix all the linter issues.",
	}, {
		ID:          "02-mvi-coverage",
		Description: "MVI coverage - Ensure the codebase is covered by tests",
		Prompt:      "Execute the command: 'make coverage && ./coverage' and ensure to increse the coverage percentage for the MVI new code.",
	}, {
		ID:          "02-mvi-coverage-report",
		Description: "MVI coverage report - Update the accomplishment report",
		Prompt:      "Update the accomplishment report {{.ChangeRequestFilePath}}.02-mvi.accomplished.md with the new coverage percentage.",
	},
	{
		ID:          "03-extend-functionalities",
		Description: "Extending functionalities - Adding additional features and improvements",
		Prompt: `You are a senior software engineer that is working on a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration is divided into 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions) [done]
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria [done]
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task now is to proceed to **expand the implementation** to cover additional use cases, edge cases, and deferred features, as described in the blueprint.

### 🎯 Goals of this Phase

- Extend the core logic to handle **all scenarios** described in the user stories and their acceptance criteria.
- Add meaningful logic to improve completeness while maintaining modularity.
- Update or create **new tests** to ensure coverage of extended functionality.

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

Let's start the work:
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
- Read the "laying the foundation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.01-foundation.accomplished.md
- Read the "Minimum Viable Implementation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.02-mvi.accomplished.md
`,
	},
	{
		ID:          "03-extend-functionalities-accomplished",
		Description: "Extending functionalities accomplished - Summary of the changes",
		Prompt: `You are a technical writer that needs to document the implementation of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration is divided into 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions)
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to document what has been accomplished so far in the file {{.ChangeRequestFilePath}}.03-extend-functionalities.accomplished.md

The accomplishment report is not a summary, is a "compass" to the changes you made, hence avoid general statements/claim, be precise:
Use always short code references (no code at all, 
 just a compact/understable reference to lookup for, do not use line numbers) 
 as foundation of your statements
 For example:
 - Instead of "Added tests for ..." / "Updated tests for ... " show me which test case has been added (using code references)
 - Instead of "Message templates are now centralized with clear naming conventions" show me where to find them (using code references)
 - Include a section of "blind spot" if any: leverage test coverage report to reinforce your statements
 - Include a dedicated section for potentially still not yet well covered acceptance criteria.

Now, let's start the work:
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
- Read the "laying the foundation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.01-foundation.accomplished.md
- Read the "Minimum Viable Implementation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.02-mvi.accomplished.md
`,
	},
	{
		ID:          "03-extend-functionalities-test",
		Description: "Extending functionalities testing - Verifying the additional features",
		Prompt: `You are a senior software engineer that is working on a software iteration based on a set of user stories described in a blueprint document. 

The extended functionalities have been built, now your task is:
- Run the full test suite to confirm no regressions have been introduced.
- If major components are touched, consider adding or updating smoke/regression tests to validate the foundation work.

Now, let's start the work:
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
- Read the "laying the foundation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.01-foundation.accomplished.md
- Read the "Minimum Viable Implementation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.02-mvi.accomplished.md
`,
	},
	{
		ID:          "03-extend-functionalities-lint",
		Description: "Extending functionalities linting - Verifying the additional features",
		Prompt:      "Execute the command: make lint. Ensure to fix all the linter issues.",
	},
	{
		ID:          "03-extend-functionalities-coverage",
		Description: "Extending functionalities coverage - Verifying the additional features",
		Prompt:      "Execute the command: 'make coverage && ./coverage' and ensure to increse the coverage percentage for the extended functionalities.",
	},
	{
		ID:          "03-extend-functionalities-coverage-report",
		Description: "Extending functionalities coverage report - Update the accomplishment report",
		Prompt:      "Update the accomplishment report {{.ChangeRequestFilePath}}.03-extend-functionalities.accomplished.md with the new coverage percentage.",
	},
	{
		ID:          "04-final-iteration",
		Description: "Final iteration - Polishing and final adjustments",
		Prompt: `You are a senior software engineer that is working on a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration is divided into 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions) [done]
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria [done]
- Extend the implementation to support more scenarios and edge cases [done]
- Refine and stabilize the codebase for clarity, maintainability, and performance

Now, it's time to execute the last phase of the iteration: **Refinement & Stabilization**.

### 🎯 Objectives:
- Refine the codebase for clarity, maintainability, and performance
- Enhance test coverage to simulate real-world usage
- Ensure robustness through thorough validation
- Finalize the iteration by producing production-quality code and test suites

---

### 🧩 Tasks to Perform:

1. **Optimize for Performance (if applicable):**
   - Profile critical paths
   - Optimize data structures and algorithms for efficiency
   - Avoid premature optimization — focus on known bottlenecks or risky parts

2. **Enhance Test Coverage:**
   - Use make test to run the test suite with coverage
   - then use ./coverage -file <path-to-file> to identify the most uncovered parts
   - Add tests that simulate real-world edge cases and usage patterns
   - Ensure each user story and its acceptance criteria are covered
   - Include tests for:
     - Error handling and invalid inputs
     - Performance-sensitive areas
     - Integration between components
   - Use integration tests to implement at least one smoke test

3. **Stabilize the Codebase:**
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

- Final version of the code with inline meaningful comments (don't comment the obvious, comment the "why" and the "how")
- Updated and complete test suite

### ⚠️ Reminder:
Do not introduce new features at this stage. Focus only on refining and stabilizing the existing work to make it reliable and production-ready.

Let's start the work:
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
- Read the "laying the foundation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.01-foundation.accomplished.md
- Read the "Minimum Viable Implementation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.02-mvi.accomplished.md
- Read the "Extending functionalities" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.03-extend-functionalities.accomplished.md
`,
	},
	{
		ID:          "04-final-iteration-accomplished",
		Description: "Final iteration accomplished - Summary of the changes",
		Prompt: `You are a technical writer that needs to document the implementation of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration is divided into 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions) [done]
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria [done]
- Extend the implementation to support more scenarios and edge cases [done]
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to document what has been accomplished so far in the file {{.ChangeRequestFilePath}}.04-refinement.accomplished.md

The accomplishment report is not a summary, it is a "compass" to the changes you made, hence avoid general statements/claim, be precise:
Use always short code references (no code at all, just a compact/understable reference to lookup for, do not use line numbers) as foundation of your statements
For example:
- Instead of "Added tests for ..." / "Updated tests for ... " show me which test case has been added (using code references)
- Instead of "Message templates are now centralized with clear naming conventions" show me where to find them (using code references)
- Include a section of "blind spot" if any: leverage test coverage report to reinforce your statements
- Include a dedicated section for potentially still not yet well implemented acceptance criteria.
- Include any changes to original design decisions
`,
	},
	{
		ID:          "04-final-iteration-test",
		Description: "Final iteration testing - Final verification and validation",
		Prompt: `You are a senior software engineer that is working on the last phase of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration is divided into 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions) [done]
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria [done]
- Extend the implementation to support more scenarios and edge cases [done]
- Refine and stabilize the codebase for clarity, maintainability, and performance [done]

The final iteration has been completed, now your task is:
- Run the full test suite to confirm no regressions have been introduced.
- Increase the coverage percentage. Averall coverage should be > 65%.

Now, let's start the work:
- Read the user stories using ./usm cat {{.ChangeRequestFilePath}}
- Read the blueprint using cat {{.ChangeRequestFilePath}}
- Read the "laying the foundation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.01-foundation.accomplished.md
- Read the "Minimum Viable Implementation" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.02-mvi.accomplished.md
- Read the "Extending functionalities" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.03-extend-functionalities.accomplished.md
- Read the "Final iteration" accomplished summary using the command: cat {{.ChangeRequestFilePath}}.04-refinement.accomplished.md
`,
	},
	{
		ID:          "04-final-iteration-linter",
		Description: "Final iteration linter - Verifying the final iteration",
		Prompt:      "Execute the command: make lint. Ensure to fix all the linter issues.",
	},
	{
		ID:          "04-final-iteration-coverage",
		Description: "Final iteration coverage - Final verification and validation",
		Prompt: `Read the final iteration accomplished summary using the command: cat {{.ChangeRequestFilePath}}.04-refinement.accomplished.md
Execute the command 'make test' to run the test suite with coverage
then use ./coverage -file <path-to-file> to identify the most uncovered parts.
Don't give up until the overall coverage percentage is enough to satisfy the .github/workflows/coverage.yml file.`,
	}, {
		ID:          "04-final-iteration-coverage-report",
		Description: "Final iteration coverage report - Update the accomplishment report",
		Prompt:      "Update the accomplishment report {{.ChangeRequestFilePath}}.04-refinement.accomplished.md with the new coverage percentage.",
	}, {
		ID:          "implementation",
		Description: "The implementation report of the change request",
		Prompt: `You are a technical writer that needs to document the implementation of a software iteration based on a set of user stories described in a blueprint document. 

The whole iteration was composed by 4 phases:
- Laid the foundation (project structure, placeholders, key abstractions)
- Complete the Minimum Viable Implementation (MVI) to satisfy core acceptance criteria
- Extend the implementation to support more scenarios and edge cases
- Refine and stabilize the codebase for clarity, maintainability, and performance

Your task is to document what has been accomplished so far in the file {{.ChangeRequestDirname}}/{{.ChangeRequestBasename}}.implementation.md

Keep into consideration that this technical document will be archived as Architecture Decision Record. Use reference to the implemented code to look up for (do not use line numbers, they change too frequently)
Describe data structures and their purposes (if any)
Describe relevant algorithms and their purporses (if any)
 
The original blueprint: ./usm {{.ChangeRequestFilePath}}
the accomplished reports:
- foundations: cat {{.ChangeRequestFilePath}}.01-foundation.accomplished.md
- mvi: cat {{.ChangeRequestFilePath}}.02-mvi.accomplished.md
- extend: cat {{.ChangeRequestFilePath}}.03-extend-functionalities.accomplished.md 
- refine: cat {{.ChangeRequestFilePath}}.04-refinement.accomplished.md   
`,
	},
} 
# Foundation Accomplishment Report - Introduce Prompt in Step Definition

This report documents the foundational changes made to introduce a prompt field in the step definition.

## 1. Updated Data Structures

### 1.1 WorkflowStep Structure

Extended the `WorkflowStep` struct in `internal/workflow/workflow.go` to include the new `Prompt` field:

- Added `Prompt string // AI agent instructions with variable interpolation` field to `WorkflowStep`
- Updated all standard workflow steps to include default prompts with variable interpolation

### 1.2 PromptVariables Structure

Created the `PromptVariables` struct in `internal/workflow/prompt.go` to support variable management:

- Implemented as `PromptVariables` with a `ChangeRequestFilePath` field
- Added documentation for future extensibility

## 2. Implemented Variable Interpolation

Created new interpolation functions in `internal/workflow/prompt.go`:

- `interpolatePrompt`: Replaces variables in the format `${variable_name}` with their values
- `interpolatePromptWithMissingVars`: Interpolates variables and returns missing variables
- `interpolatePromptWithMap`: Supports custom variable mappings for extensibility
- `generateStepPrompt`: Main function to generate prompts for workflow steps
- `generateDefaultPrompt`: Creates a default prompt based on step description

## 3. Integrated Prompt Processing

Updated the step executor to use the prompt for generating step content:

- Modified `ExecuteStep` in `internal/workflow/executor.go` to process the prompt with variable interpolation
- Updated `generateStepContent` to include the interpolated prompt in the output

## 4. Added Tests

Added comprehensive tests for the new functionality:

- Created `internal/workflow/prompt_test.go` for testing prompt-related functionality:
  - `TestWorkflowStepStructure`: Verifies the WorkflowStep structure includes the Prompt field
  - `TestInterpolatePrompt`: Tests basic variable interpolation
  - `TestInterpolatePromptWithMissingVars`: Tests handling of missing variables
  - `TestInterpolatePromptWithMap`: Tests custom variable mappings
  - `TestGenerateStepPrompt`: Tests step prompt generation
  - `TestGenerateDefaultPrompt`: Tests default prompt generation

- Updated existing tests to accommodate the new functionality:
  - Modified `TestStepExecutor_ExecuteStep` to include validation for the prompt field
  - Updated `TestGenerateStepContent` to include the prompt parameter

## 5. Design Decisions

### 5.1 Prompt Interpolation

- Used a simple string replacement approach for the initial implementation
- Implemented a more flexible regex-based approach for future extensibility
- Added different interpolation functions to support various use cases

### 5.2 Default Prompts

- Provided meaningful default prompts for all standard workflow steps
- Ensured backward compatibility by generating defaults when prompt is empty

### 5.3 Future Content Generation Improvements

The current implementation introduces the prompt field while maintaining the existing content generation mechanism in `generateStepContent`. For future phases, we recommend:

- Shift away from hardcoded templates in `generateStepContent`
- Make `generateStepContent` primarily use the prompt field for content generation
- Keep step-specific formatting but derive core instructional content from the prompt

This will streamline the content generation process and reduce duplication between the prompt field and the hardcoded templates.

## 6. Blind Spots

Based on the implementation of the prompt interpolation functionality, there are a few potential blind spots:

- There's no validation of prompt templates during workflow creation/initialization
- Error handling for invalid variable references needs to be improved
- No extensive logging during interpolation to help with debugging
- No performance testing for large prompts or many variables

## 7. Acceptance Criteria Status

All acceptance criteria have been addressed in the foundation phase:

1. ✅ The step definition now includes a new prompt attribute
   - Implemented in `WorkflowStep` struct in `internal/workflow/workflow.go`
   - All standard workflow steps updated with default prompts

2. ✅ The prompt supports variable interpolation with `change_request_file_path`
   - Implemented in `interpolatePrompt` in `internal/workflow/prompt.go`
   - Added tests for variable replacement

3. ✅ The implementation is designed to be extendable
   - Created `PromptVariables` struct that can be extended
   - Implemented `interpolatePromptWithMap` for custom variable mappings
   - Added tests for extensibility 
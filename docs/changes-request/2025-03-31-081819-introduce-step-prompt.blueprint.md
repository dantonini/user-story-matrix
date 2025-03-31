---
name: introduce step prompt
created-at: 2025-03-31T08:18:19+02:00
user-stories:
  - title: Introduce Prompt in Existing Step Definition
    file: docs/user-stories/code-command/08-introduce-prompt-in-existing-step-definition.md
    content-hash: 55419763d3973fbbe3d750e40669f163

---

# Blueprint: Introduce Prompt in Step Definition

## Overview
This change request aims to extend the current step definition by introducing a prompt field that will provide actionable instructions to the AI agent. The prompt field will support variable interpolation, initially supporting the `change_request_file_path` variable, with a design that allows for future expansion of supported variables.

## Fundamentals

### Data Structures

1. **Extended WorkflowStep**
   ```go
   type WorkflowStep struct {
       ID          string // Unique identifier (e.g., "01-laying-the-foundation")
       Description string // Human-readable description
       Prompt      string // AI agent instructions with variable interpolation
       OutputFile  string // Template for output filename
   }
   ```

2. **PromptVariables**
   ```go
   type PromptVariables struct {
       ChangeRequestFilePath string
       // Future variables can be added here
   }
   ```

### Algorithms

1. **Variable Interpolation**
   ```pseudo
   function interpolatePrompt(prompt string, variables PromptVariables) string:
       // Replace variables in the format ${variable_name}
       result = prompt
       for each variable in variables:
           placeholder = "${" + variable.name + "}"
           result = replace(result, placeholder, variable.value)
       return result
   ```

2. **Prompt Generation**
   ```pseudo
   function generateStepPrompt(step WorkflowStep, changeRequestPath string) string:
       if step.Prompt == "":
           return generateDefaultPrompt(step)
           
       vars = PromptVariables{
           ChangeRequestFilePath: changeRequestPath
       }
       return interpolatePrompt(step.Prompt, vars)
   ```

## How to verify – Detailed User Story Breakdown

### Acceptance Criteria 1: Step Definition Includes Prompt Attribute
- **Test 1.1:** Verify WorkflowStep struct includes the Prompt field
  ```go
  func TestWorkflowStepStructure(t *testing.T) {
      step := WorkflowStep{}
      assertFieldExists(t, step, "Prompt")
  }
  ```
- **Test 1.2:** Verify prompt field can be set and retrieved
  ```go
  func TestPromptFieldAccess(t *testing.T) {
      step := WorkflowStep{Prompt: "Test prompt"}
      assertEqual(t, "Test prompt", step.Prompt)
  }
  ```

### Acceptance Criteria 2: Variable Interpolation Support
- **Test 2.1:** Verify basic variable interpolation
  ```go
  func TestBasicInterpolation(t *testing.T) {
      prompt := "Process the file at ${change_request_file_path}"
      vars := PromptVariables{ChangeRequestFilePath: "/path/to/file"}
      expected := "Process the file at /path/to/file"
      result := interpolatePrompt(prompt, vars)
      assertEqual(t, expected, result)
  }
  ```
- **Test 2.2:** Verify handling of missing variables
  ```go
  func TestMissingVariableHandling(t *testing.T) {
      prompt := "Process ${nonexistent_var} and ${another_missing_var}"
      vars := PromptVariables{}
      result, missingVars := interpolatePromptWithMissingVars(prompt, vars)
      assertEqual(t, prompt, result) // Should leave unknown variables unchanged
      assertContains(t, missingVars, "nonexistent_var")
      assertContains(t, missingVars, "another_missing_var")
      assertEqual(t, 2, len(missingVars))
  }
  ```

### Acceptance Criteria 3: Extensible Design
- **Test 3.1:** Verify PromptVariables struct can be extended
  ```go
  func TestPromptVariablesExtensibility(t *testing.T) {
      // Test with extended variables structure
      prompt := "Process ${change_request_file_path} with ${new_variable}"
      
      // Define extended variables type
      type ExtendedVariables struct {
          PromptVariables
          NewVariable string
      }
      
      // Create instance with values
      vars := ExtendedVariables{
          PromptVariables: PromptVariables{ChangeRequestFilePath: "/path"},
          NewVariable: "test",
      }
      
      // Create variable map adapter that works with our extended type
      varMap := map[string]string{
          "change_request_file_path": vars.ChangeRequestFilePath,
          "new_variable": vars.NewVariable,
      }
      
      expected := "Process /path with test"
      result := interpolatePromptWithMap(prompt, varMap)
      assertEqual(t, expected, result)
  }
  ```

## What is the Plan – Detailed Action Items

### 1. Update Data Structures
1. Modify `WorkflowStep` struct in `internal/workflow/workflow.go`:
   - Add the `Prompt` field
   - Update struct documentation
   - Update any direct struct initializations

2. Create new `PromptVariables` type in `internal/workflow/prompt.go`:
   - Define the structure with initial `ChangeRequestFilePath` field
   - Add documentation for future extensibility
   - Include methods for variable management

### 2. Implement Variable Interpolation
1. Create new `interpolation.go` file in `internal/workflow/`:
   - Implement `interpolatePrompt` function
   - Implement `interpolatePromptWithMissingVars` that returns the interpolated string and a list of missing variables
   - Implement `interpolatePromptWithMap` that allows for custom variable mappings
   - Add helper functions for variable extraction and replacement
   - Include comprehensive error handling

2. Update `StepExecutor` in `internal/workflow/executor.go`:
   - Add prompt processing to `ExecuteStep`
   - Integrate variable interpolation
   - Handle empty/missing prompts gracefully

### 3. Update Standard Workflow Steps
1. Modify `StandardWorkflowSteps` in `workflow.go`:
   - Add default prompts for existing steps
   - Ensure backward compatibility
   - Document prompt format and variables

### 4. Add Tests
1. Create new test file `internal/workflow/prompt_test.go`:
   - Add all test cases outlined in verification section
   - Include edge cases and error conditions
   - Test integration with existing functionality

### 5. Update Documentation
1. Update code documentation:
   - Add godoc comments for new types and functions
   - Include examples of prompt usage
   - Document variable interpolation syntax

2. Update user documentation:
   - Add section about prompt field in README
   - Include examples of prompt usage
   - Document supported variables

### 6. Validation and Error Handling
1. Add validation in `WorkflowManager`:
   - Validate prompt syntax during step creation
   - Add error handling for invalid variable references
   - Include warning for unknown variables

### Migration Notes
- Existing steps without prompts will continue to work
- Default prompts will be generated based on step description
- No database schema changes required
- No breaking changes to existing APIs

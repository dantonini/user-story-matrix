---
name: custom-workflow-phase2
created-at: 2025-04-09T06:30:03+02:00
user-stories:
  - title: Extract prompt files from standard workflow
    file: docs/user-stories/custom-workflow/dev-02-extract-prompt-files-from-standard-workflow.md
    content-hash: 8b7e57f181e7ccc95988df82a6fa37bbebd4ed05dfdcefa3f46c4c9a558521fe
  - title: Add workflow loading from filesystem
    file: docs/user-stories/custom-workflow/dev-03-add-workflow-loading-from-filesystem.md
    content-hash: d7e9dbeb981e92faa5201bee836e4e18e98f832dc22950fb802e01555ce28d39
  - title: Update workflow state format
    file: docs/user-stories/custom-workflow/dev-04-update-workflow-state-format.md
    content-hash: 720fca05da86821139e0e123c5624b4a5ab74f2e7a95f10e247f8a457b57d3ac

---
# Blueprint

## Overview

This phase of custom workflow implementation focuses on three critical aspects of the USM system:
1. Extracting prompt content from embedded code into separate files
2. Implementing filesystem-based workflow loading
3. Updating the workflow state format to track workflow identification

Together, these changes create a robust foundation for customizable workflows. By decoupling prompt content from code and adding file-based workflow definition capabilities, we enable users to create, modify, and share workflows without code changes. The workflow state format update ensures that change requests maintain a consistent association with their original workflow, improving tracking and compatibility.

## Fundamentals

### Data Structures

#### 1. WorkflowFileDefinition
A serializable representation of workflow definitions for file-based storage:

```go
// WorkflowFileDefinition represents the structure of a workflow.yaml file
type WorkflowFileDefinition struct {
    Name        string              `yaml:"name"`
    Description string              `yaml:"description"`
    Steps       []WorkflowFileStep  `yaml:"steps"`
}

// WorkflowFileStep represents a step in a workflow.yaml file
type WorkflowFileStep struct {
    ID          string              `yaml:"id"`
    Description string              `yaml:"description"`
    Prompt      string              `yaml:"prompt"` // Path to prompt file, relative to workflow dir
}
```

#### 2. Updated WorkflowState
The existing WorkflowState structure will be extended to include workflow identification:

```go
type WorkflowState struct {
    ChangeRequestPath string    // Path to the change request file
    CurrentStepIndex  int       // Index of the current step (0-based)
    LastModified      time.Time // When the state was last updated
    CompletedSteps    []string  // List of completed step IDs
    WorkflowName      string    // Name of the workflow being used
    WorkflowPath      string    // Optional path to the workflow definition
}
```

#### 3. Workflow Registry Cache
A cache for loaded workflows to improve performance:

```go
type workflowCache struct {
    workflows map[string]*WorkflowDefinition
    sources   map[string]string           // Maps workflow name to source path
    modified  map[string]time.Time        // Last modified timestamps
}
```

### Algorithms

#### 1. Extract Standard Workflow
This utility will extract the StandardWorkflowSteps to filesystem:

```
Function extractStandardWorkflow(outputDir string):
    1. Create directory structure:
       - outputDir/workflow.yaml
       - outputDir/prompts/
    2. For each step in StandardWorkflowSteps:
       a. Create a prompt file with step ID as filename
       b. Write step's prompt content to file
    3. Create workflow.yaml containing step metadata and references
    4. Return path to created workflow
```

#### 2. Workflow Loading Process
Algorithm for loading workflow definitions from disk:

```
Function loadWorkflowFromDirectory(directory string):
    1. Validate directory exists
    2. Read workflow.yaml file
    3. Parse YAML into WorkflowFileDefinition
    4. Validate required fields (name, steps with IDs)
    5. For each step:
       a. Resolve prompt file path
       b. Read prompt content
       c. Create WorkflowStep with loaded content
    6. Create WorkflowDefinition from parsed data
    7. Return complete definition
```

#### 3. State File Backward Compatibility
Algorithm to handle older state file formats:

```
Function loadStateWithBackwardCompatibility(path string):
    1. Read state file
    2. Try to parse as new format
    3. If new fields are missing:
       a. Set WorkflowName to StandardWorkflowName
       b. Set WorkflowPath to ""
    4. Return populated WorkflowState
```

### Refactoring Strategy

1. **Prompt Content Extraction**: 
   - Extract content without changing core functionality
   - Create a two-phase loading system (file first, then in-memory fallback)

2. **Workflow Registry Extension**:
   - Add new methods while preserving existing behavior
   - Implement caching with filesystem monitoring for performance

3. **State Format Update**:
   - Use backward compatibility layer for seamless migration
   - Update related functions to handle and preserve new fields

## How to verify – Detailed User Story Breakdown

### 1. Extract prompt files from standard workflow

#### Acceptance Criteria Verification

1. **All prompts are extracted to separate Markdown files**
   - Testing Scenario: After running the extraction utility, count the number of files in the prompts directory and compare with the number of steps in StandardWorkflowSteps
   - Automated Test: Create a test that iterates through StandardWorkflowSteps and verifies a corresponding file exists for each step

2. **Files are organized in the correct directory structure**
   - Testing Scenario: Verify the directory structure matches the specification (workflow.yaml at root, prompts subdirectory with step files)
   - Automated Test: Check that all expected paths exist after extraction

3. **Prompt file content matches original content**
   - Testing Scenario: For each prompt file, compare content with the corresponding StandardWorkflowSteps prompt
   - Automated Test: Read each extracted file and compare with the original in-memory prompt

4. **workflow.yaml is correctly generated**
   - Testing Scenario: Parse the generated workflow.yaml and verify it contains all steps with correct metadata
   - Automated Test: Deserialize workflow.yaml and compare with StandardWorkflowSteps

5. **workflow.yaml references prompt files correctly**
   - Testing Scenario: Check that each step's prompt field in workflow.yaml points to the correct file
   - Automated Test: Verify each prompt path resolves to an existing file

6. **StandardWorkflowSteps continues to work during transition**
   - Testing Scenario: Execute a workflow using the standard steps and verify behavior is unchanged
   - Automated Test: Run the same workflow tests before and after extraction to confirm identical results

7. **Fallback mechanism works for prompt loading**
   - Testing Scenario: Deliberately remove a prompt file and verify the system falls back to the embedded prompt
   - Automated Test: Test workflow execution with missing prompt files

### 2. Add workflow loading from filesystem

#### Acceptance Criteria Verification

1. **WorkflowRegistry can load definitions from disk**
   - Testing Scenario: Call LoadFromDirectory with a valid workflow directory and verify a WorkflowDefinition is returned
   - Automated Test: Create a test workflow on disk and load it with the new method

2. **Registry can discover workflows in standard locations**
   - Testing Scenario: Place test workflows in standard locations and call DiscoverWorkflows
   - Automated Test: Verify all test workflows are discovered and loaded correctly

3. **Workflow loading process validates and resolves properly**
   - Testing Scenario: Create test workflows with various configurations (valid, missing prompts, etc.)
   - Automated Test: Verify valid workflows load successfully and invalid ones fail with appropriate errors

4. **System handles errors gracefully**
   - Testing Scenario: Create test cases for each error condition (missing files, invalid YAML, etc.)
   - Automated Test: Verify each error case returns an appropriate error message

5. **Registry caches loaded workflows**
   - Testing Scenario: Load the same workflow multiple times and measure performance
   - Automated Test: Verify subsequent loads use cached data by mocking file system and counting reads

6. **Registry can reload workflows when they change**
   - Testing Scenario: Modify a workflow file and call the reload method
   - Automated Test: Verify the updated content is reflected in the loaded workflow

7. **System logs workflow loading information**
   - Testing Scenario: Load workflows with various conditions and check logs
   - Automated Test: Use a mock logger to capture and verify log messages

### 3. Update workflow state format

#### Acceptance Criteria Verification

1. **WorkflowState struct includes workflow identification**
   - Testing Scenario: Create a new workflow state and verify the new fields are present
   - Automated Test: Test struct field access and serialization

2. **State file format includes new fields**
   - Testing Scenario: Save a state file and verify the JSON contains the new fields
   - Automated Test: Deserialize saved state files and check for new fields

3. **Backward compatibility is maintained**
   - Testing Scenario: Create a state file in the old format and load it
   - Automated Test: Verify old-format files load with expected default values

4. **WorkflowManager methods handle new format**
   - Testing Scenario: Test LoadState, SaveState, and UpdateState with the new format
   - Automated Test: Verify state changes are preserved across operations

5. **Workflow switching handles compatibility**
   - Testing Scenario: Switch between workflows with different step structures
   - Automated Test: Test compatibility validation and warning messages

## What is the Plan – Detailed Action Items

### 1. Extract prompt files from standard workflow

1. **Create directory structure for standard workflow**:
   - Create `internal/workflow/templates/standard/` directory
   - Create `internal/workflow/templates/standard/prompts/` directory

2. **Implement extraction utility**:
   - Create a new file `internal/workflow/extract.go` with functions to export StandardWorkflowSteps to files:
     ```go
     // ExtractStandardWorkflow exports the standard workflow to the specified directory
     func ExtractStandardWorkflow(fs FileSystem, outputDir string) error
     
     // generateWorkflowYAML creates a workflow.yaml file from StandardWorkflowSteps
     func generateWorkflowYAML(fs FileSystem, steps []WorkflowStep, outputPath string) error
     
     // extractPromptToFile writes a single prompt to a markdown file
     func extractPromptToFile(fs FileSystem, promptsDir string, step WorkflowStep) error
     ```

3. **Create prompt loading mechanism**:
   - Modify the WorkflowStep structure to support file-based prompts:
     ```go
     // promptSourceType identifies where a prompt comes from
     type promptSourceType int
     
     const (
         promptSourceEmbedded promptSourceType = iota
         promptSourceFile
     )
     
     // promptSource tracks the origin of a prompt
     type promptSource struct {
         sourceType promptSourceType
         filePath   string // Only used when sourceType is promptSourceFile
     }
     
     // WorkflowStep with internal prompt tracking
     type WorkflowStep struct {
         ID          string
         Description string
         Prompt      string
         source      promptSource // internal field for tracking
     }
     ```
   
4. **Update prompt loading logic**:
   - Add methods to load prompt content from files when available:
     ```go
     // loadPromptContent loads prompt content, with priority to file sources
     func (step *WorkflowStep) loadPromptContent(fs FileSystem) (string, error)
     
     // setPromptFromFile sets the prompt source to a file path
     func (step *WorkflowStep) setPromptFromFile(path string)
     ```

5. **Implement CLI command to execute extraction**:
   - Add a command to extract standard workflow:
     ```go
     // extractWorkflowCmd represents the extract command
     var extractWorkflowCmd = &cobra.Command{
         Use:   "extract-workflow",
         Short: "Extract standard workflow to filesystem",
         Run: func(cmd *cobra.Command, args []string) {
             // Implementation
         },
     }
     ```

6. **Add tests for extraction functionality**:
   - Create test file `internal/workflow/extract_test.go`
   - Implement tests for each extraction-related function
   - Test backward compatibility with existing code

### 2. Add workflow loading from filesystem

1. **Add file-based workflow definition types**:
   - Create new types in `internal/workflow/loader.go`:
     ```go
     // WorkflowFileDefinition represents a workflow definition in YAML format
     type WorkflowFileDefinition struct {
         Name        string            `yaml:"name"`
         Description string            `yaml:"description"`
         Steps       []WorkflowFileStep `yaml:"steps"`
     }
     
     // WorkflowFileStep represents a step in a workflow YAML file
     type WorkflowFileStep struct {
         ID          string `yaml:"id"`
         Description string `yaml:"description"`
         Prompt      string `yaml:"prompt"` // Path to prompt file
     }
     ```

2. **Implement workflow loading functions**:
   - Add methods to WorkflowRegistry for loading from filesystem:
     ```go
     // LoadFromDirectory loads a workflow from a directory
     func (r *WorkflowRegistry) LoadFromDirectory(fs FileSystem, path string) (*WorkflowDefinition, error)
     
     // DiscoverWorkflows finds and loads workflows from standard locations
     func (r *WorkflowRegistry) DiscoverWorkflows(fs FileSystem) map[string]*WorkflowDefinition
     
     // loadWorkflowYAML parses a workflow.yaml file
     func loadWorkflowYAML(fs FileSystem, path string) (*WorkflowFileDefinition, error)
     
     // loadPromptFiles loads prompt content from referenced files
     func loadPromptFiles(fs FileSystem, baseDir string, fileDefinition *WorkflowFileDefinition) ([]WorkflowStep, error)
     ```

3. **Add caching mechanism**:
   - Add cache structure to WorkflowRegistry:
     ```go
     // workflowCache stores loaded workflows
     type workflowCache struct {
         workflows map[string]*WorkflowDefinition
         sources   map[string]string       // Maps workflow name to source path
         modified  map[string]time.Time    // Last modified timestamps
     }
     
     // WorkflowRegistry with cache
     type WorkflowRegistry struct {
         builtInWorkflows map[string]*WorkflowDefinition
         cache            workflowCache
         mutex            sync.RWMutex
     }
     ```
   
4. **Implement workflow reloading**:
   - Add method to check for and reload changed workflows:
     ```go
     // ReloadChangedWorkflows checks for modified workflow files and reloads them
     func (r *WorkflowRegistry) ReloadChangedWorkflows(fs FileSystem) []string
     
     // isWorkflowModified checks if a workflow file has been modified
     func (r *WorkflowRegistry) isWorkflowModified(fs FileSystem, name string) (bool, error)
     ```

5. **Add validation for loaded workflows**:
   - Implement validation functions for file-based workflows:
     ```go
     // validateWorkflowDefinition checks a workflow for required fields
     func validateWorkflowDefinition(def *WorkflowFileDefinition) []error
     
     // validatePromptReferences checks that all prompt files exist
     func validatePromptReferences(fs FileSystem, baseDir string, def *WorkflowFileDefinition) []error
     ```

6. **Add standard workflow directory discovery**:
   - Implement functions to find standard workflow locations:
     ```go
     // GetStandardWorkflowDirectories returns potential workflow locations
     func GetStandardWorkflowDirectories() []string
     ```

7. **Add tests for loading functionality**:
   - Create test file `internal/workflow/loader_test.go`
   - Implement tests for each loading-related function
   - Test error handling and caching behavior

### 3. Update workflow state format

1. **Update WorkflowState struct**:
   - Modify struct in `internal/workflow/workflow.go`:
     ```go
     type WorkflowState struct {
         ChangeRequestPath string    // Path to the change request file
         CurrentStepIndex  int       // Index of the current step (0-based)
         LastModified      time.Time // When the state was last updated
         CompletedSteps    []string  // List of completed step IDs
         WorkflowName      string    // Name of the workflow being used
         WorkflowPath      string    // Optional path to the workflow definition
     }
     ```

2. **Implement backward compatibility**:
   - Update LoadState method in WorkflowManager:
     ```go
     // LoadState with backward compatibility
     func (wm *WorkflowManager) LoadState(changeRequestPath string) (WorkflowState, error) {
         // Implementation with backward compatibility logic
     }
     ```

3. **Update SaveState method**:
   - Modify to always save in the new format:
     ```go
     // SaveState always saves in new format
     func (wm *WorkflowManager) SaveState(state WorkflowState) error {
         // Implementation using new format
     }
     ```

4. **Update UpdateState method**:
   - Preserve workflow identification when updating:
     ```go
     // UpdateState preserves workflow identification
     func (wm *WorkflowManager) UpdateState(changeRequestPath string, newStepIndex int) error {
         // Implementation that preserves workflow info
     }
     ```

5. **Add workflow switching validation**:
   - Implement function to validate workflow compatibility:
     ```go
     // ValidateWorkflowSwitch checks compatibility between workflows
     func (wm *WorkflowManager) ValidateWorkflowSwitch(oldWorkflowName, newWorkflowName string) []string
     
     // MapProgressBetweenWorkflows attempts to map progress between workflows
     func (wm *WorkflowManager) MapProgressBetweenWorkflows(oldState WorkflowState, newWorkflowName string) (WorkflowState, []string)
     ```

6. **Update affected commands**:
   - Update any CLI commands that interact with workflow state:
     ```go
     // stepCmd with workflow selection option
     var stepCmd = &cobra.Command{
         Use:   "step",
         Short: "Manage workflow steps",
         Run: func(cmd *cobra.Command, args []string) {
             // Updated implementation with workflow awareness
         },
     }
     ```

7. **Add tests for state format changes**:
   - Create test file `internal/workflow/state_test.go`
   - Test loading old format files
   - Test saving and loading new format
   - Test workflow switching validation

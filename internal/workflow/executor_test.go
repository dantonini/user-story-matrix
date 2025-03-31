// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// testFileSystem is a mock implementation of FileSystem for testing
type testFileSystem struct {
	files        map[string][]byte
	exists       map[string]bool
	mkdirErr     error
	writeFileErr error
}

func newTestFileSystem() *testFileSystem {
	return &testFileSystem{
		files:  make(map[string][]byte),
		exists: make(map[string]bool),
	}
}

func (m *testFileSystem) ReadFile(path string) ([]byte, error) {
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func (m *testFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	if m.writeFileErr != nil {
		return m.writeFileErr
	}
	m.files[path] = data
	m.exists[path] = true
	return nil
}

func (m *testFileSystem) Exists(path string) bool {
	return m.exists[path]
}

func (m *testFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirErr != nil {
		return m.mkdirErr
	}
	m.exists[path] = true
	return nil
}

// testUserOutput is a mock implementation of UserOutput for testing
type testUserOutput struct {
	messages         []string
	progressMessages []string
	errorMessages    []string
	warningMessages  []string
	stepMessages     []string
	successMessages  []string
}

func newTestUserOutput() *testUserOutput {
	return &testUserOutput{
		messages:         make([]string, 0),
		progressMessages: make([]string, 0),
		errorMessages:    make([]string, 0),
		warningMessages:  make([]string, 0),
		stepMessages:     make([]string, 0),
		successMessages:  make([]string, 0),
	}
}

func (t *testUserOutput) Print(msg string) {
	t.messages = append(t.messages, msg)
}

func (t *testUserOutput) PrintSuccess(msg string) {
	t.successMessages = append(t.successMessages, msg)
}

func (t *testUserOutput) PrintError(msg string) {
	t.errorMessages = append(t.errorMessages, msg)
}

func (t *testUserOutput) PrintWarning(msg string) {
	t.warningMessages = append(t.warningMessages, msg)
}

func (t *testUserOutput) PrintProgress(msg string) {
	t.progressMessages = append(t.progressMessages, msg)
}

func (t *testUserOutput) PrintStep(current int, total int, msg string) {
	t.stepMessages = append(t.stepMessages, fmt.Sprintf("Step %d/%d: %s", current, total, msg))
}

func TestStepExecutor_ExecuteStep(t *testing.T) {
	tests := []struct {
		name           string
		changeRequest  string
		step          WorkflowStep
		wantSuccess    bool
		wantErrorText  string
		expectedOutput []string
	}{
		{
			name: "Successful execution",
			changeRequest: `# Test Change Request
This is a test change request.`,
			step: WorkflowStep{
				ID:          "01-laying-the-foundation",
				Description: "Laying the foundation",
				Prompt:      "This is a test prompt with ${change_request_file_path} variable",
				OutputFile:  "%s.01-laying-the-foundation.md",
			},
			wantSuccess: true,
			expectedOutput: []string{
				"# Laying the foundation",
				"## Architecture & Design",
				"This step focuses on setting up the architecture and structure for the implementation.",
				"### Key Activities",
				"1. This is a test prompt with change-request.md variable",
				"Step ID: 01-laying-the-foundation",
				"Step Description: Laying the foundation",
				"Step Prompt: This is a test prompt with change-request.md variable",
			},
		},
		{
			name:          "File not found",
			changeRequest: "",
			step: WorkflowStep{
				ID:          "01-laying-the-foundation",
				Description: "Laying the foundation",
				Prompt:      "Test prompt",
				OutputFile:  "%s.01-laying-the-foundation.md",
			},
			wantSuccess:   false,
			wantErrorText: fmt.Sprintf(ErrFileNotFound, "change-request.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			fs := newTestFileSystem()
			io := newTestUserOutput()

			// Create executor
			executor := NewStepExecutor(fs, io)

			// Set up mock file system
			if tt.changeRequest != "" {
				fs.files["change-request.md"] = []byte(tt.changeRequest)
			}

			// Execute step
			success, err := executor.ExecuteStep("change-request.md", tt.step, "output.md")

			// Check success/failure
			if tt.wantSuccess {
				if !success || err != nil {
					t.Errorf("ExecuteStep() success = %v, error = %v, want success = true, error = nil", success, err)
				}

				// Check output file content
				output, exists := fs.files["output.md"]
				if !exists {
					t.Error("ExecuteStep() did not create output file")
				} else {
					outputStr := string(output)
					for _, expectedLine := range tt.expectedOutput {
						if !strings.Contains(outputStr, expectedLine) {
							t.Errorf("Expected line not found in output: %s", expectedLine)
						}
					}
				}
			} else {
				if success || err == nil {
					t.Error("ExecuteStep() expected error, got nil")
				} else if err.Error() != tt.wantErrorText {
					t.Errorf("ExecuteStep() error = %v, want %v", err, tt.wantErrorText)
				}
			}
		})
	}
}

func TestStepExecutor_ExecuteStep_FileSystemErrors(t *testing.T) {
	// Test cases for file system errors
	tests := []struct {
		name          string
		setupFS       func(*testFileSystem)
		wantSuccess   bool
		wantErrorText string
	}{
		{
			name: "change request file not found",
			setupFS: func(fs *testFileSystem) {
				// Don't add the change request file
			},
			wantSuccess:   false,
			wantErrorText: fmt.Sprintf(ErrFileNotFound, "change-request.md"),
		},
		{
			name: "mkdir error",
			setupFS: func(fs *testFileSystem) {
				fs.files["change-request.md"] = []byte("Test content")
				fs.exists["change-request.md"] = true
				fs.mkdirErr = fmt.Errorf("mkdir error")
			},
			wantSuccess:   false,
			wantErrorText: fmt.Sprintf(ErrOutputFileCreateFailed, "mkdir error"),
		},
		{
			name: "write file error",
			setupFS: func(fs *testFileSystem) {
				fs.files["change-request.md"] = []byte("Test content")
				fs.exists["change-request.md"] = true
				
				// Set the writeFileErr to simulate an error during WriteFile
				fs.writeFileErr = fmt.Errorf("write file error")
			},
			wantSuccess:   false,
			wantErrorText: fmt.Sprintf(ErrOutputFileCreateFailed, "write file error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fs := newTestFileSystem()
			io := newTestUserOutput()
			executor := NewStepExecutor(fs, io)

			// Setup file system state
			tt.setupFS(fs)

			// Execute
			step := WorkflowStep{
				ID:          "01-laying-the-foundation",
				Description: "Laying the foundation",
			}
			success, err := executor.ExecuteStep("change-request.md", step, "output/test.md")

			// Check success/error
			if success != tt.wantSuccess {
				t.Errorf("ExecuteStep() success = %v, want %v", success, tt.wantSuccess)
			}
			if err == nil || err.Error() != tt.wantErrorText {
				t.Errorf("ExecuteStep() error = %v, want %v", err, tt.wantErrorText)
			}
		})
	}
}

// TestGenerateStepContent tests the generateStepContent method with different step types
func TestGenerateStepContent(t *testing.T) {
	tests := []struct {
		name           string
		stepID         string
		description    string
		prompt         string
		inputContent   string
		expectedStrings []string
		expectError    bool
	}{
		{
			name:        "01-laying-the-foundation",
			stepID:      "01-laying-the-foundation",
			description: "Laying the foundation",
			prompt:      "Test prompt for foundation",
			inputContent: `# Test Change Request
This is a test change request.`,
			expectedStrings: []string{
				"# Laying the foundation",
				"## Architecture & Design",
				"This step focuses on setting up the architecture and structure for the implementation.",
				"Step ID: 01-laying-the-foundation",
				"Step Description: Laying the foundation",
				"Step Prompt: Test prompt for foundation",
			},
			expectError: false,
		},
		{
			name:        "01-laying-the-foundation-test",
			stepID:      "01-laying-the-foundation-test",
			description: "Laying the foundation test",
			prompt:      "Test prompt for foundation test",
			inputContent: `# Test Change Request
This is a test change request.`,
			expectedStrings: []string{
				"## Foundation Testing",
				"This step verifies the foundational changes made in the previous step.",
				"Step ID: 01-laying-the-foundation-test",
				"Step Prompt: Test prompt for foundation test",
			},
			expectError: false,
		},
		{
			name:        "Unknown step",
			stepID:      "unknown-step",
			description: "Unknown step",
			prompt:      "Test prompt for unknown step",
			inputContent: `# Test Change Request
This is a test change request.`,
			expectedStrings: []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create executor
			executor := NewStepExecutor(newTestFileSystem(), newTestUserOutput())

			// Create step
			step := WorkflowStep{
				ID:          tt.stepID,
				Description: tt.description,
				Prompt:      tt.prompt,
				OutputFile:  "%s.output.md",
			}

			// Call generateStepContent
			output, err := executor.generateStepContent(tt.inputContent, step, tt.prompt)
			
			// Check for expected error
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error for step ID %s, but got none", tt.stepID)
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Error calling generateStepContent: %v", err)
			}

			// Check output contains expected content
			for _, expected := range tt.expectedStrings {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it did not", expected)
				}
			}
		})
	}
}

func TestStepExecutor_ExecuteStep_WriteFileError(t *testing.T) {
	// Create mocks
	fs := newTestFileSystem()
	io := newTestUserOutput()
	
	// Set up the file system
	fs.files["change-request.md"] = []byte("# Test Change Request")
	fs.exists["change-request.md"] = true
	
	// Set a write file error
	fs.writeFileErr = fmt.Errorf("write file error")
	
	// Create executor
	executor := NewStepExecutor(fs, io)
	
	// Create a test step
	step := WorkflowStep{
		ID:          "01-laying-the-foundation",
		Description: "Laying the foundation",
	}
	
	// Execute step (expect failure)
	success, err := executor.ExecuteStep("change-request.md", step, "output.md")
	
	// Check results
	if success {
		t.Error("ExecuteStep() should return false when WriteFile fails")
	}
	
	if err == nil {
		t.Error("ExecuteStep() should return error when WriteFile fails")
	}
	
	// Verify the error message
	expectedError := fmt.Sprintf(ErrOutputFileCreateFailed, "write file error")
	if err.Error() != expectedError {
		t.Errorf("ExecuteStep() error = %v, want %v", err, expectedError)
	}
}

// This update adds new test cases for validating prompts with errors
func TestStepExecutor_ExecuteStep_PromptValidation(t *testing.T) {
	tests := []struct {
		name              string
		prompt            string
		expectWarnings    bool
		expectMissingVars bool
	}{
		{
			name:              "Valid prompt with existing variable",
			prompt:            "Process the file at ${change_request_file_path}",
			expectWarnings:    false,
			expectMissingVars: false,
		},
		{
			name:              "Valid prompt with missing variable",
			prompt:            "Process the file at ${unknown_variable}",
			expectWarnings:    false,
			expectMissingVars: true,
		},
		{
			name:              "Malformed prompt",
			prompt:            "Process the file at ${variable with spaces}",
			expectWarnings:    true,
			expectMissingVars: false,
		},
		{
			name:              "Unclosed variable syntax",
			prompt:            "Process the file at ${incomplete",
			expectWarnings:    true,
			expectMissingVars: false,
		},
		{
			name:              "Empty prompt",
			prompt:            "",
			expectWarnings:    false,
			expectMissingVars: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			fs := newTestFileSystem()
			io := newTestUserOutput()

			// Create executor
			executor := NewStepExecutor(fs, io)

			// Create test step
			step := WorkflowStep{
				ID:          "test-step",
				Description: "Test Step",
				Prompt:      tt.prompt,
				OutputFile:  "output.md",
			}

			// Set up mock file system
			fs.files["change-request.md"] = []byte("# Test Change Request\nThis is a test.")

			// Execute step
			executor.ExecuteStep("change-request.md", step, "output.md")

			// Check if warnings were generated as expected
			if tt.expectWarnings && len(io.warningMessages) == 0 {
				t.Error("Expected warnings for malformed prompt, but none were generated")
			}

			// Check if missing variables were reported as expected
			if tt.expectMissingVars {
				foundMissingVarWarning := false
				for _, msg := range io.warningMessages {
					if contains(msg, "undefined variables") {
						foundMissingVarWarning = true
						break
					}
				}
				if !foundMissingVarWarning {
					t.Error("Expected warning about undefined variables, but none was generated")
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestFormatPromptAsInstructions(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{
			name:     "Empty prompt",
			prompt:   "",
			expected: "No specific instructions provided.",
		},
		{
			name:     "Whitespace only prompt",
			prompt:   "   \n  \t  ",
			expected: "No specific instructions provided.",
		},
		{
			name:     "Single sentence",
			prompt:   "This is a test prompt.",
			expected: "1. This is a test prompt.\n",
		},
		{
			name:     "Multiple sentences",
			prompt:   "First sentence. Second sentence. Third sentence.",
			expected: "1. First sentence.\n2. Second sentence.\n3. Third sentence.\n",
		},
		{
			name:     "Different punctuation",
			prompt:   "Question? Exclamation! Statement.",
			expected: "1. Question?\n2. Exclamation!\n3. Statement.\n",
		},
		{
			name:     "Newlines",
			prompt:   "First line.\nSecond line.\nThird line.",
			expected: "1. First line.\n2. Second line.\n3. Third line.\n",
		},
		{
			name:     "Prompt with empty sentences",
			prompt:   "First sentence.. Second sentence.",
			expected: "1. First sentence.\n2. Second sentence.\n",
		},
		{
			name:     "No valid sentences",
			prompt:   "....,,,,",
			expected: "No specific instructions provided.",
		},
		{
			name:     "Sentences without punctuation",
			prompt:   "First sentence Second sentence Third sentence",
			expected: "1. First sentence Second sentence Third sentence.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPromptAsInstructions(tt.prompt)
			if result != tt.expected {
				t.Errorf("formatPromptAsInstructions() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractSentences(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "Empty text",
			text:     "",
			expected: []string{},
		},
		{
			name:     "Whitespace only",
			text:     "   \n  \t  ",
			expected: []string{},
		},
		{
			name:     "Single sentence",
			text:     "This is a test.",
			expected: []string{"This is a test"},
		},
		{
			name:     "Multiple sentences with periods",
			text:     "First sentence. Second sentence. Third sentence.",
			expected: []string{"First sentence", "Second sentence", "Third sentence"},
		},
		{
			name:     "Sentences with different punctuation",
			text:     "Question? Exclamation! Statement.",
			expected: []string{"Question", "Exclamation", "Statement"},
		},
		{
			name:     "Sentences with newlines",
			text:     "First line.\nSecond line.\nThird line.",
			expected: []string{"First line", "Second line", "Third line"},
		},
		{
			name:     "No ending punctuation",
			text:     "This sentence has no ending punctuation",
			expected: []string{"This sentence has no ending punctuation"},
		},
		{
			name:     "Mixed ending and no ending punctuation",
			text:     "First sentence. Second sentence without ending punctuation",
			expected: []string{"First sentence", "Second sentence without ending punctuation"},
		},
		{
			name:     "Empty sentences",
			text:     "First sentence... Second sentence.",
			expected: []string{"First sentence", "Second sentence"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSentences(tt.text)
			
			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("extractSentences() returned %d sentences, want %d", len(result), len(tt.expected))
				return
			}
			
			// Check content
			for i, s := range result {
				if !strings.Contains(s, tt.expected[i]) {
					t.Errorf("sentence %d = %q, want to contain %q", i, s, tt.expected[i])
				}
			}
		})
	}
}

func TestGenerateStepContent_PromptIntegration(t *testing.T) {
	tests := []struct {
		name               string
		stepID             string
		stepDescription    string
		prompt             string
		changeRequestContent string
		expectedContains   []string
	}{
		{
			name:               "Foundation step with prompt",
			stepID:             "01-laying-the-foundation",
			stepDescription:    "Laying the foundation",
			prompt:             "Create necessary packages. Define core data structures. Establish file organization.",
			changeRequestContent: "Test change request",
			expectedContains: []string{
				"# Laying the foundation",
				"## Architecture & Design",
				"This step focuses on setting up the architecture and structure for the implementation.",
				"### Key Activities",
				"1. Create necessary packages.",
				"2. Define core data structures.",
				"3. Establish file organization.",
				"Step Prompt: Create necessary packages. Define core data structures. Establish file organization.",
			},
		},
		{
			name:               "Test step with prompt",
			stepID:             "01-laying-the-foundation-test",
			stepDescription:    "Foundation testing",
			prompt:             "Verify package structure. Test interfaces. Validate data structures.",
			changeRequestContent: "Test change request",
			expectedContains: []string{
				"# Foundation testing",
				"## Foundation Testing",
				"This step verifies the foundational changes made in the previous step.",
				"### Test Coverage",
				"1. Verify package structure.",
				"2. Test interfaces.",
				"3. Validate data structures.",
				"Step Prompt: Verify package structure. Test interfaces. Validate data structures.",
			},
		},
		{
			name:               "MVI step with prompt",
			stepID:             "02-mvi",
			stepDescription:    "Minimum Viable Implementation",
			prompt:             "Implement core functionality. Add basic error handling. Create minimal UI.",
			changeRequestContent: "Test change request",
			expectedContains: []string{
				"# Minimum Viable Implementation",
				"## Minimum Viable Implementation",
				"This step implements the core functionality with minimal features.",
				"### Key Activities",
				"1. Implement core functionality.",
				"2. Add basic error handling.",
				"3. Create minimal UI.",
			},
		},
		{
			name:               "MVI test step with prompt",
			stepID:             "02-mvi-test",
			stepDescription:    "MVI Testing",
			prompt:             "Test core functionality. Verify error handling. Check integration.",
			changeRequestContent: "Test change request",
			expectedContains: []string{
				"# MVI Testing",
				"## MVI Testing",
				"This step verifies the minimum viable implementation.",
				"### Test Coverage",
				"1. Test core functionality.",
				"2. Verify error handling.",
				"3. Check integration.",
			},
		},
		{
			name:               "Extended functionality step",
			stepID:             "03-extend-functionalities",
			stepDescription:    "Extended Functionality",
			prompt:             "Add additional features. Enhance error handling. Optimize performance.",
			changeRequestContent: "Test change request",
			expectedContains: []string{
				"# Extended Functionality",
				"## Extended Functionality",
				"This step adds additional features and improvements.",
				"### Key Activities",
				"1. Add additional features.",
				"2. Enhance error handling.",
				"3. Optimize performance.",
			},
		},
		{
			name:               "Extended functionality test step",
			stepID:             "03-extend-functionalities-test",
			stepDescription:    "Extended Functionality Testing",
			prompt:             "Test all features. Verify error handling. Benchmark performance.",
			changeRequestContent: "Test change request",
			expectedContains: []string{
				"# Extended Functionality Testing",
				"## Extended Functionality Testing",
				"This step verifies the extended functionality.",
				"### Test Coverage",
				"1. Test all features.",
				"2. Verify error handling.",
				"3. Benchmark performance.",
			},
		},
		{
			name:               "Final iteration step",
			stepID:             "04-final-iteration",
			stepDescription:    "Final Iteration",
			prompt:             "Clean up code. Update documentation. Final optimizations.",
			changeRequestContent: "Test change request",
			expectedContains: []string{
				"# Final Iteration",
				"## Final Iteration",
				"This step focuses on polishing and final adjustments.",
				"### Key Activities",
				"1. Clean up code.",
				"2. Update documentation.",
				"3. Final optimizations.",
			},
		},
		{
			name:               "Final iteration test step",
			stepID:             "04-final-iteration-test",
			stepDescription:    "Final Testing",
			prompt:             "Run end-to-end tests. Verify documentation. Confirm performance targets.",
			changeRequestContent: "Test change request",
			expectedContains: []string{
				"# Final Testing",
				"## Final Testing",
				"This step performs final verification and validation.",
				"### Test Coverage",
				"1. Run end-to-end tests.",
				"2. Verify documentation.",
				"3. Confirm performance targets.",
			},
		},
		{
			name:               "Empty prompt",
			stepID:             "01-laying-the-foundation",
			stepDescription:    "Laying the foundation",
			prompt:             "",
			changeRequestContent: "Test change request",
			expectedContains: []string{
				"# Laying the foundation",
				"## Architecture & Design",
				"This step focuses on setting up the architecture and structure for the implementation.",
				"### Key Activities",
				"No specific instructions provided.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			fs := newTestFileSystem()
			io := newTestUserOutput()
			
			// Create executor
			executor := NewStepExecutor(fs, io)
			
			// Create step
			step := WorkflowStep{
				ID:          tt.stepID,
				Description: tt.stepDescription,
				Prompt:      tt.prompt,
				OutputFile:  "output.md",
			}
			
			// Generate content
			content, err := executor.generateStepContent(tt.changeRequestContent, step, tt.prompt)
			
			// Verify no error
			if err != nil {
				t.Errorf("generateStepContent() error = %v", err)
				return
			}
			
			// Verify content contains expected strings
			for _, expected := range tt.expectedContains {
				if !strings.Contains(content, expected) {
					t.Errorf("generateStepContent() = %q, want to contain %q", content, expected)
				}
			}
		})
	}
}

func TestGenerateStepContent_InvalidStepID(t *testing.T) {
	// Create mocks
	fs := newTestFileSystem()
	io := newTestUserOutput()
	
	// Create executor
	executor := NewStepExecutor(fs, io)
	
	// Create steps with invalid IDs
	testCases := []struct {
		name        string
		stepID      string
		wantErrText string
	}{
		{
			name:        "Completely invalid ID",
			stepID:      "invalid-id",
			wantErrText: "unknown step ID format: invalid-id",
		},
		{
			name:        "Invalid prefix",
			stepID:      "00-something",
			wantErrText: "unknown step ID format: 00-something",
		},
		{
			name:        "Invalid format",
			stepID:      "prefix-without-number",
			wantErrText: "unknown step ID format: prefix-without-number",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			step := WorkflowStep{
				ID:          tc.stepID,
				Description: "Test step",
				Prompt:      "Test prompt",
				OutputFile:  "output.md",
			}
			
			// Try to generate content with invalid step ID
			_, err := executor.generateStepContent("Test CR", step, "Test prompt")
			
			// Check error
			if err == nil {
				t.Error("generateStepContent() expected error, got nil")
			} else if err.Error() != tc.wantErrText {
				t.Errorf("generateStepContent() error = %v, want %v", err, tc.wantErrText)
			}
		})
	}
}

func TestIsInvalidSentence(t *testing.T) {
	tests := []struct {
		name     string
		sentence string
		expected bool
	}{
		{
			name:     "Normal sentence",
			sentence: "This is a valid sentence.",
			expected: false,
		},
		{
			name:     "Only punctuation",
			sentence: "....,,,!!!",
			expected: true,
		},
		{
			name:     "Only whitespace and punctuation",
			sentence: " . , ! ? ; : ",
			expected: true,
		},
		{
			name:     "Empty string",
			sentence: "",
			expected: true,
		},
		{
			name:     "Whitespace only",
			sentence: "   \t  \n  ",
			expected: true,
		},
		{
			name:     "Valid sentence with lots of punctuation",
			sentence: "This, is a valid sentence, with punctuation!!!",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInvalidSentence(tt.sentence)
			if result != tt.expected {
				t.Errorf("isInvalidSentence(%q) = %v, want %v", tt.sentence, result, tt.expected)
			}
		})
	}
}

func TestCleanPunctuation(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Double periods",
			text:     "Test sentence.. with double periods...",
			expected: "Test sentence. with double periods.",
		},
		{
			name:     "Multiple double punctuation",
			text:     "Test with multiple,, types!! of?? punctuation..",
			expected: "Test with multiple, types! of? punctuation.",
		},
		{
			name:     "No double punctuation",
			text:     "Normal sentence with no double punctuation.",
			expected: "Normal sentence with no double punctuation.",
		},
		{
			name:     "Many repeated punctuation",
			text:     "Test......... with many....... repeated punctuation!!!!!",
			expected: "Test. with many. repeated punctuation!",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPunctuation(tt.text)
			if result != tt.expected {
				t.Errorf("cleanPunctuation(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
} 
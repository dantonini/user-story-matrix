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
	debugEnabled     bool
}

func newTestUserOutput() *testUserOutput {
	return &testUserOutput{
		messages:         make([]string, 0),
		progressMessages: make([]string, 0),
		errorMessages:    make([]string, 0),
		warningMessages:  make([]string, 0),
		stepMessages:     make([]string, 0),
		successMessages:  make([]string, 0),
		debugEnabled:     false,
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

func (t *testUserOutput) PrintTable(headers []string, rows [][]string) {
	// Not needed for these tests
}

func (t *testUserOutput) IsDebugEnabled() bool {
	return t.debugEnabled
}

func TestStepExecutor_ExecuteStep(t *testing.T) {
	tests := []struct {
		name           string
		changeRequest  string
		step           WorkflowStep
		wantSuccess    bool
		wantErrorText  string
		expectedOutput string
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
			wantSuccess:    true,
			expectedOutput: "This is a test prompt with change-request.md variable",
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
				fs.exists["change-request.md"] = true
			}

			// Execute step
			success, err := executor.ExecuteStep("change-request.md", tt.step, "output.md")

			// Check success/failure
			if tt.wantSuccess {
				if !success || err != nil {
					t.Errorf("ExecuteStep() error = %v, success = %v", err, success)
				}

				// Check that the expected message was printed
				found := false
				for _, msg := range io.messages {
					if msg == tt.expectedOutput {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Expected message not found: %s\nActual messages: %v", tt.expectedOutput, io.messages)
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

			// Check result
			if tt.wantSuccess {
				if !success || err != nil {
					t.Errorf("ExecuteStep() success = %v, error = %v, want success = true, error = nil", success, err)
				}
			} else {
				if success || err == nil {
					t.Errorf("Expected error, got success = %v, error = nil", success)
				} else if err.Error() != tt.wantErrorText {
					t.Errorf("Error = %v, want %v", err, tt.wantErrorText)
				}
			}
		})
	}
}

// This update adds new test cases for validating prompts with errors
func TestStepExecutor_ExecuteStep_PromptValidation(t *testing.T) {
	tests := []struct {
		name            string
		prompt          string
		expectWarning   bool
		expectedWarning string
	}{
		{
			name:            "Valid prompt",
			prompt:          "This is a valid prompt with ${change_request_file_path}",
			expectWarning:   false,
			expectedWarning: "",
		},
		{
			name:            "Prompt with undefined variable",
			prompt:          "This prompt has an ${undefined_variable}",
			expectWarning:   true,
			expectedWarning: "Step 01-test contains undefined variables: [undefined_variable]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fs := newTestFileSystem()
			io := newTestUserOutput()
			executor := NewStepExecutor(fs, io)

			// Set up mock file system with a dummy change request
			fs.files["change-request.md"] = []byte("Test change request")
			fs.exists["change-request.md"] = true

			// Create step with test prompt
			step := WorkflowStep{
				ID:          "01-test",
				Description: "Test step",
				Prompt:      tt.prompt,
			}

			// Execute step
			success, err := executor.ExecuteStep("change-request.md", step, "output.md")
			if !success || err != nil {
				t.Errorf("ExecuteStep() failed: success=%v, error=%v", success, err)
			}

			// Check for warnings
			if tt.expectWarning {
				foundWarning := false
				for _, warning := range io.warningMessages {
					if warning == tt.expectedWarning {
						foundWarning = true
						break
					}
				}
				if !foundWarning {
					t.Errorf("Expected warning not found: %s\nActual warnings: %v", tt.expectedWarning, io.warningMessages)
				}
			} else {
				if len(io.warningMessages) > 0 {
					t.Errorf("Unexpected warnings found: %v", io.warningMessages)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Test formatPromptAsInstructions function
func TestFormatPromptAsInstructions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: "No specific instructions provided.",
		},
		{
			name:     "Single sentence",
			input:    "This is a simple instruction.",
			expected: "1. This is a simple instruction.\n",
		},
		{
			name:     "Multiple sentences",
			input:    "First instruction. Second instruction.",
			expected: "1. First instruction.\n2. Second instruction.\n",
		},
		{
			name:     "With punctuation",
			input:    "First instruction! Second instruction? Third instruction.",
			expected: "1. First instruction!\n2. Second instruction?\n3. Third instruction.\n",
		},
		{
			name:     "With newlines",
			input:    "First instruction.\nSecond instruction.",
			expected: "1. First instruction.\n2. Second instruction.\n",
		},
		{
			name:     "Invalid content",
			input:    "...",
			expected: "No specific instructions provided.",
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: "No specific instructions provided.",
		},
		{
			name:     "With excessive punctuation",
			input:    "First instruction... Second instruction!!! Third instruction???",
			expected: "1. First instruction.\n2. Second instruction!\n3. Third instruction?\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPromptAsInstructions(tt.input)
			if result != tt.expected {
				t.Errorf("formatPromptAsInstructions(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test extractSentences function
func TestExtractSentences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Single sentence with period",
			input:    "This is a single sentence.",
			expected: []string{"This is a single sentence."},
		},
		{
			name:     "Single sentence without ending punctuation",
			input:    "This is a single sentence without ending punctuation",
			expected: []string{"This is a single sentence without ending punctuation."},
		},
		{
			name:     "Multiple sentences with periods",
			input:    "First sentence. Second sentence. Third sentence.",
			expected: []string{"First sentence.", "Second sentence.", "Third sentence."},
		},
		{
			name:     "Sentences with different punctuation",
			input:    "First sentence. Second sentence! Third sentence?",
			expected: []string{"First sentence.", "Second sentence!", "Third sentence?"},
		},
		{
			name:     "Sentences with newlines",
			input:    "First sentence.\nSecond sentence.\nThird sentence.",
			expected: []string{"First sentence.", "Second sentence.", "Third sentence."},
		},
		{
			name:     "Sentences with excessive punctuation",
			input:    "First sentence... Second sentence!!! Third sentence???",
			expected: []string{"First sentence.", "Second sentence!", "Third sentence?"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSentences(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("extractSentences(%q) returned %d sentences, want %d", tt.input, len(result), len(tt.expected))
				t.Errorf("Got: %v, Want: %v", result, tt.expected)
				return
			}

			for i, sentence := range result {
				if sentence != tt.expected[i] {
					t.Errorf("extractSentences(%q)[%d] = %q, want %q", tt.input, i, sentence, tt.expected[i])
				}
			}
		})
	}
}

// Test isInvalidSentence function
func TestIsInvalidSentence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: true,
		},
		{
			name:     "Only punctuation",
			input:    "...",
			expected: true,
		},
		{
			name:     "Only punctuation with spaces",
			input:    ". . .",
			expected: true,
		},
		{
			name:     "Mixed punctuation",
			input:    ".,!?",
			expected: true,
		},
		{
			name:     "Valid sentence",
			input:    "This is a valid sentence.",
			expected: false,
		},
		{
			name:     "Single word",
			input:    "Hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInvalidSentence(tt.input)
			if result != tt.expected {
				t.Errorf("isInvalidSentence(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Test cleanPunctuation function
func TestCleanPunctuation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No excessive punctuation",
			input:    "This is a normal sentence.",
			expected: "This is a normal sentence.",
		},
		{
			name:     "Double periods",
			input:    "This has double periods..",
			expected: "This has double periods.",
		},
		{
			name:     "Triple periods",
			input:    "This has triple periods...",
			expected: "This has triple periods.",
		},
		{
			name:     "Multiple double punctuation",
			input:    "This has double periods.. and commas,, and exclamations!! and questions??",
			expected: "This has double periods. and commas, and exclamations! and questions?",
		},
		{
			name:     "Multiple mixed punctuation",
			input:    "This has many.... different,,, punctuation!!!! marks????",
			expected: "This has many. different, punctuation! marks?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPunctuation(tt.input)
			if result != tt.expected {
				t.Errorf("cleanPunctuation(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

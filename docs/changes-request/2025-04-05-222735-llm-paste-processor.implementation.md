# LLM Paste Processor Implementation

## Overview

The LLM Paste Processor feature enhances the User Story Matrix CLI tool by enabling intelligent processing of unstructured text to automatically populate user story form fields. This feature allows users to paste unstructured content (such as requirements documents, meeting notes, or emails) into the form, then processes that content using OpenAI's GPT models to extract structured information and populate the form fields with appropriate confidence indicators.

The implementation spans four development phases:
1. Foundation - Core architecture and interfaces
2. Minimum Viable Implementation (MVI) - Basic functionality
3. Extension - Enhanced parsing and user experience
4. Refinement - Code quality, testing, and stabilization

## Architecture

The implementation follows a modular architecture with clear separation of concerns:

### Core Components

1. **LLM Processing Interface Layer**
   - `LLMProcessor` interface defines the contract for LLM processing capabilities
   - `OpenAIProcessor` provides concrete implementation for OpenAI APIs
   - `ConfigManager` handles secure API key storage and configuration

2. **UI Components**
   - Page Object Model (POM) pattern for clean UI architecture
   - `UserStoryFormModel` for business logic and state management
   - `UserStoryForm` for presentation and user interaction
   - `Spinner` component for visual processing feedback

3. **Clipboard Integration**
   - `clipboard_helper.go` provides advanced paste detection algorithms
   - Detection for various paste scenarios: middle paste, large inserts

4. **Configuration Management**
   - Secure API key storage and validation
   - Graceful degradation when API key is not configured

## Data Structures

### Core Data Models

1. **`UserStoryData` Structure**
   ```go
   type UserStoryData struct {
       Title             string
       Description       string
       AsA               string
       IWant             string
       SoThat            string
       AcceptanceCriteria []string
       Confidence        map[string]float64
   }
   ```
   This structure holds the parsed results from the LLM processing, including confidence scores for each field.

2. **`ProcessingState` Enum**
   ```go
   type ProcessingState int

   const (
       ProcessingIdle ProcessingState = iota
       ProcessingActive
       ProcessingSuccess
       ProcessingError
       ProcessingTimeout
       ProcessingCancelled
       ProcessingNotConfigured
       ProcessingPartialSuccess
   )
   ```
   Tracks the state of LLM processing for proper UI feedback.

3. **`APIKeyConfig` Structure**
   ```go
   type APIKeyConfig struct {
       OpenAIKey string
       IsValid   bool
       LastValidated time.Time
   }
   ```
   Manages API key configuration and validation status.

4. **`UserStoryFormModel` Structure**
   ```go
   type UserStoryFormModel struct {
       UserStory models.UserStory
       LLMProcessor llm.LLMProcessor
       ConfigManager *llm.ConfigManager
       AutoPopulatedFields map[string]bool
       ConfidenceScores map[string]float64
       ProcessingState llm.ProcessingState
       ProcessingStartTime time.Time
       TimeoutThreshold time.Duration
       ProcessingCancelled bool
       ShowAPIKeyMessage bool
       LastError error
       FormData FormData
       UIState FormUIState
   }
   ```
   Holds the complete state of the user story form including LLM processing status.

### Interface Definitions

1. **`LLMProcessor` Interface**
   ```go
   type LLMProcessor interface {
       ProcessUnstructuredText(ctx context.Context, text string) (UserStoryData, error)
       GetConfidenceScores() map[string]float64
       IsConfigured() bool
       ValidateConfiguration(ctx context.Context) error
       Configure(config APIKeyConfig) error
       GetProcessingState() ProcessingState
   }
   ```
   Defines the contract for LLM processing capabilities, allowing for different implementations.

2. **`OpenAIClientInterface`**
   ```go
   type OpenAIClientInterface interface {
       CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
   }
   ```
   Abstract interface for OpenAI client to enable easier testing.

## Algorithms and Techniques

### 1. Paste Detection Algorithm

The clipboard paste detection implements multiple strategies to identify pasted content:

```go
func GetActiveFieldValue(currentValue, previousValue string) (string, bool) {
    // Try to detect a middle paste
    if inserted, ok := detectMiddlePaste(previousValue, currentValue); ok {
        return inserted, true
    }
    
    // Try to detect a large insertion
    if inserted := detectLargeInsert(previousValue, currentValue); inserted != "" {
        return inserted, true
    }
    
    // If the length difference is substantial, it might be a large paste
    if len(currentValue) > len(previousValue) + PasteThresholdLength {
        return currentValue, true
    }
    
    return "", false
}
```

The implementation uses multiple strategies:

1. **Middle Paste Detection**
   - Finds the longest common prefix and suffix between previous and current text
   - Identifies content inserted in the middle based on these boundaries
   - Handles edge cases for content pasted at beginning or end

2. **Large Insert Detection**
   - Uses a sliding window approach to find matching segments in both strings
   - Identifies gaps between matches as potentially pasted content
   - Implements specialized handling for very different string lengths

### 2. Acceptance Criteria Parsing

The criteria parsing algorithm supports various input formats:

```go
func parseAcceptanceCriteria(input string) []string {
    // Handle single-line inputs (space-separated or single criterion)
    // ...
    
    // For multi-line input with various formats
    var criteria []string
    lines := strings.Split(input, "\n")
    
    for _, line := range lines {
        // Check for bullet points (-, *, •)
        // Check for numbered lists (1., 2., etc.)
        // Check for parenthesized numbers ((1), (2), etc.)
        // Handle regular lines as single criteria
    }
    
    return criteria
}
```

This algorithm intelligently detects:
- Bullet points with various markers (`-`, `*`, `•`)
- Numbered lists (`1.`, `2.`, etc.)
- Parenthesized numbers (`(1)`, `(2)`, etc.)
- Multi-line criteria without formatting
- Space-separated single-line criteria

### 3. OpenAI API Integration

The `ProcessUnstructuredText` method implements a robust approach to API communication:

```go
func (p *OpenAIProcessor) ProcessUnstructuredText(ctx context.Context, text string) (UserStoryData, error) {
    // Set processing state to active
    p.processingState = ProcessingActive
    
    // Create a structured output prompt with JSON schema
    systemPrompt := `You are an AI trained to extract user story information from unstructured text.
    Extract the following components in JSON format:
    - title: A concise title for the user story (25 words or less)
    - description: A detailed description (if available)
    - as_a: The user type or role (from "As a...")
    - i_want: The capability or feature (from "I want...")
    - so_that: The benefit or reason (from "So that...")
    - acceptance_criteria: Array of acceptance criteria (bullet points or separate items)
    
    Also include a confidence score (0.0 to 1.0) for each field to indicate how confident you are in the extraction.
    `
    
    // Make API request with error handling and retries
    // Parse the JSON response
    // Extract and validate the structured data
    // Update processing state based on result
}
```

Key aspects:
- Uses OpenAI's structured output format (JSON mode)
- Provides detailed schema instructions for consistent output
- Implements confidence scoring for each extracted field
- Handles timeouts, cancellation, and error conditions

### 4. Batch Confirmation Algorithm

The system implements batch confirmation of auto-populated fields:

```go
func (m *UserStoryFormModel) ConfirmAllFields() {
    m.AutoPopulatedFields = make(map[string]bool)
}

func (m *UserStoryFormModel) GetAutoPopulatedFieldCount() int {
    return len(m.AutoPopulatedFields)
}

func (m *UserStoryFormModel) HasAutoPopulatedFields() bool {
    return len(m.AutoPopulatedFields) > 0
}
```

This functionality allows users to:
- See how many fields were auto-populated
- Confirm all auto-populated fields at once
- Maintain confidence scores even after confirmation

## Implementation Details

### Configuration Management

The `ConfigManager` handles secure storage of API keys:

```go
func (m *ConfigManager) SaveConfig() error {
    // Create the config directory if it doesn't exist
    // Marshal the config to JSON
    // Write the config file with secure permissions (0600)
}

func (m *ConfigManager) SetOpenAIKey(key string, processor LLMProcessor) error {
    // Validate the key is not empty
    // Configure the processor with the new key
    // Test the key with a simple API call
    // Update and save the config if successful
}
```

Key security features:
- Secure file permissions (0600)
- API key validation before storage
- Error handling for various failure scenarios

### UI Integration

The UI components implement:

1. **Processing State Management**
   - Visual indicator for active processing
   - Timeout warnings after threshold (default: 5 seconds)
   - Error message display for failed processing

2. **Field Highlighting**
   - Confidence-based visual feedback
   - Tracking of auto-populated vs. manually edited fields
   - Batch confirmation of all auto-populated fields

3. **Keyboard Shortcuts**
   - ESC key for cancelling processing
   - Ctrl+A for batch confirmation of auto-populated fields

### Graceful Degradation

The implementation handles various failure scenarios:

1. **Missing API Key**
   - Non-intrusive message about API key configuration
   - Fallback to manual form entry
   - Instructions for configuring the API key

2. **Processing Failures**
   - Timeout warnings for slow responses
   - Error messages for API failures
   - Retry logic for transient errors

3. **Partial Processing**
   - Handling of partially successful extractions
   - Only populating fields with sufficient confidence

## Testing Strategy

The implementation includes comprehensive testing:

1. **Unit Tests**
   - Isolated tests for each component and algorithm
   - Mock implementations for external dependencies
   - Comprehensive coverage of error paths

2. **Integration Tests**
   - End-to-end tests for the complete form flow
   - Tests for various acceptance criteria formats
   - Complex paste scenario testing

3. **Error Simulation**
   - Enhanced `MockFileSystem` with error simulation capabilities
   - Testing of API key validation failures
   - Context cancellation and timeout testing

## Performance Considerations

The implementation addresses performance in several ways:

1. **Asynchronous Processing**
   - Non-blocking UI during LLM processing
   - Progress indicators for long-running operations
   - Cancellation support for user control

2. **Optimized Algorithms**
   - Efficient paste detection with minimal false positives
   - Smart acceptance criteria parsing for various formats
   - Context management to prevent resource leaks

3. **Timeouts and Resilience**
   - Configurable timeout thresholds
   - Visual feedback for slow operations
   - Graceful handling of processing failures

## Future Considerations

Potential future enhancements include:

1. **API Rate Limiting Handling**
   - Improved backoff strategies for API rate limits
   - User feedback for rate limit situations
   - Queue management for multiple processing requests

2. **Enhanced Error Reporting**
   - More detailed error messages for specific failure cases
   - Integration with broader error reporting system
   - User-friendly suggestions for resolving errors

3. **Error Simulation Patterns**
   - Apply similar error simulation to other mock components
   - Enhanced test coverage for edge cases
   - More realistic simulation of network conditions

## Conclusion

The LLM Paste Processor feature enhances the usability of the User Story Matrix CLI tool by providing intelligent processing of unstructured text. The implementation follows a modular architecture with clean separation of concerns, robust error handling, and comprehensive testing. The feature successfully integrates OpenAI's GPT models to automate the extraction of structured information from unstructured text, while providing clear visual feedback and user control over the process.

The current implementation achieves 66.4% test coverage overall, with key components like the UI models package reaching 98.3% coverage. The code is well-structured, properly documented, and follows best practices for maintainability and extensibility. 
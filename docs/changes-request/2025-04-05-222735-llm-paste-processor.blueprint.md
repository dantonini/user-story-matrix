---
name: llm paste processor
created-at: 2025-04-05T22:27:35+02:00
user-stories:
  - title: Pasting Unstructured Text
    file: docs/user-stories/paste-text/01-pasting-unstructured-text.md
    content-hash: 72103a47ad4818b1ed09f79483da977cc3047a1e11d285fa1fdf9c5e7340cdcf
  - title: LLM Parsing Feedback & Validation
    file: docs/user-stories/paste-text/02-llm-parsing-feedback-validation.md
    content-hash: f9cec0c9d2f3bfcbee5e35a921e20afd40b5fcfda1d0553762e3713a4708db25
  - title: Progress Indicator During LLM Processing
    file: docs/user-stories/paste-text/03-progress-indicator-during-llm-processing.md
    content-hash: 363e5b9537009380343c63be566b843c42d1cdfe338d7e3b08ec0f3b1c383b2e

---

# Blueprint

## Overview

This change request enhances the `usm add user-story` command to support a more intelligent and user-friendly interface through LLM integration. The primary goal is to allow users to paste unstructured text (such as requirements, emails, meeting notes, etc.) into the TUI form, and have an LLM parse this text to automatically populate the relevant form fields. This integration includes visual feedback on the parsing process, confidence indicators for populated fields, and interactive validation controls.

The enhancement addresses three main user stories:
1. Pasting Unstructured Text - Detecting paste events and using OpenAI GPT-4o-mini to intelligently populate form fields
2. LLM Parsing Feedback & Validation - Providing clear feedback on auto-populated fields and allowing manual override
3. Progress Indicator During LLM Processing - Showing a visual indication of the LLM processing status with timeout warnings

## Fundamentals

### Data Structures

#### LLMProcessor Interface
This interface will define the contract for LLM processing capabilities:

```go
type LLMProcessor interface {
    // ProcessUnstructuredText takes unstructured text and returns structured user story data
    ProcessUnstructuredText(ctx context.Context, text string) (UserStoryData, error)
    
    // GetConfidenceScores returns confidence scores for each parsed field
    GetConfidenceScores() map[string]float64
    
    // IsConfigured returns whether the processor has been properly configured (API key, etc.)
    IsConfigured() bool
}
```

#### OpenAIProcessor Implementation
This concrete implementation will handle communication with OpenAI API:

```go
type OpenAIProcessor struct {
    client        *openai.Client
    apiKey        string
    model         string  // default: "gpt-4o-mini"
    maxTokens     int     // default: 1000
    temperature   float32 // default: 0.2
    isConfigured  bool
}
```

#### UserStoryData Structure
This structure will hold the parsed results from the LLM:

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

#### APIKeyConfig Structure
Configuration structure for storing and retrieving API keys:

```go
type APIKeyConfig struct {
    OpenAIKey string
    IsValid   bool
    LastValidated time.Time
}
```

#### ProcessingState Enum
Track the state of LLM processing:

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
)
```

### Algorithms

#### Paste Detection and Processing
```
Function processClipboard():
    1. Detect paste event from clipboard
    2. Extract clipboard content
    3. If content length > threshold:
        a. Check if API key is configured
        b. If not configured, display non-intrusive message and return
        c. Set processing state to ProcessingActive
        d. Start progress indicator
        e. Create cancellable context with timeout
        f. Send content to OpenAI processor with structured output format
        g. Handle results asynchronously
        h. Return to UI for immediate feedback
```

#### Form Field Population
```
Function populateFormFields(userData):
    1. For each field in the form:
        a. Check if corresponding data exists in userData
        b. If exists and confidence > threshold:
            i. Populate field with data
            ii. Mark field as auto-populated
            iii. Apply visual highlighting based on confidence
        c. Else:
            i. Leave field blank or unchanged
    2. Display summary feedback of fields populated
```

#### Visual Feedback System
```
Function updateFieldStatus(field, confidence):
    1. Apply color coding based on confidence level:
        - High (>0.8): Green
        - Medium (0.5-0.8): Yellow
        - Low (<0.5): Orange
    2. Add visual indicator (icon or border)
    3. Track field in auto-populated list
    4. Set field state to "needs confirmation"
```

#### API Key Configuration
```
Function configureAPIKey(key string):
    1. Validate API key with a simple test call to OpenAI
    2. If valid:
        a. Store key securely in configuration
        b. Update processor configuration
        c. Return success
    3. If invalid:
        a. Return error with specific reason
        b. Do not store invalid key
```

### Refactoring Strategy

1. **User Story Form Component**:
   - Extract paste event handling into a separate concern
   - Add states for processing feedback
   - Implement visual indicators for field status
   - Add API key configuration prompts when needed

2. **LLM Integration**:
   - Create a new package for LLM processing with OpenAI-specific implementation
   - Implement structured output format using OpenAI's features
   - Add secure API key storage and validation
   - Implement response parsing and confidence scoring

3. **UI Enhancement**:
   - Add progress spinner component
   - Implement field highlighting system
   - Add cancel button during processing
   - Add non-intrusive API key configuration messages

## How to Verify – Detailed User Story Breakdown

### User Story 1: Pasting Unstructured Text

#### Acceptance Criteria 1.1: Paste Detection
- **Testing Scenario**: Copy a sample user story text to clipboard, then paste into the TUI form.
- **Expected Result**: The application should detect the paste event and begin LLM processing.
- **Verification Method**: Visual confirmation of the spinner appearing after paste event.

#### Acceptance Criteria 1.2: API Key Configuration
- **Testing Scenario 1**: Run the application without configuring an OpenAI API key, then attempt to paste text.
- **Expected Result**: A non-intrusive message appears indicating API key is not configured with instructions to set it.
- **Verification Method**: Verify message appears and does not disrupt normal form operation.

- **Testing Scenario 2**: Configure a valid OpenAI API key through settings.
- **Expected Result**: Key is securely stored and validation confirmation appears.
- **Verification Method**: Verify key is accepted and stored correctly.

- **Testing Scenario 3**: Configure an invalid OpenAI API key through settings.
- **Expected Result**: Error message with specific validation failure reason appears.
- **Verification Method**: Verify error message is descriptive and key is not stored.

#### Acceptance Criteria 1.3: Auto-Population
- **Testing Scenario**: Paste a well-structured user story text with API key configured.
- **Expected Result**: The form fields should be automatically populated with relevant extracted information.
- **Verification Method**: Confirm that each field contains the expected content extracted from the text.

#### Acceptance Criteria 1.4: Fallback
- **Testing Scenario**: Paste ambiguous or unrelated text that the LLM cannot parse effectively.
- **Expected Result**: The application should show a notification that parsing was unsuccessful, and fields should remain empty or unchanged.
- **Verification Method**: Verify the error message and that fields remain in their original state.

#### Acceptance Criteria 1.5: Performance
- **Testing Scenario**: Measure the time from paste event to form update.
- **Expected Result**: Under normal network conditions, fields should update within 2 seconds.
- **Verification Method**: Use a timer to measure response time.

### User Story 2: LLM Parsing Feedback & Validation

#### Acceptance Criteria 2.1: Highlight Changed Fields
- **Testing Scenario**: After successful parsing, observe the UI state of populated fields.
- **Expected Result**: Auto-populated fields should be visually highlighted until confirmed by the user.
- **Verification Method**: Visual confirmation of highlighting on populated fields.

#### Acceptance Criteria 2.2: Manual Override
- **Testing Scenario**: After auto-population, attempt to modify a populated field.
- **Expected Result**: Users should be able to freely edit and override any field value.
- **Verification Method**: Verify that typing into an auto-populated field updates it as expected.

#### Acceptance Criteria 2.3: LLM Error Handling
- **Testing Scenario**: Simulate an LLM service error (network issue, invalid response).
- **Expected Result**: The TUI should display a clear, concise error message and log the event.
- **Verification Method**: Validate error message display and check log entries.

### User Story 3: Progress Indicator During LLM Processing

#### Acceptance Criteria 3.1: Loading Animation
- **Testing Scenario**: Initiate LLM processing through a paste event.
- **Expected Result**: A spinner or progress message should appear and remain visible while the LLM is processing.
- **Verification Method**: Visual confirmation of animation during processing.

#### Acceptance Criteria 3.2: Timeout Warning
- **Testing Scenario**: Simulate a slow LLM response that exceeds 5 seconds.
- **Expected Result**: After 5 seconds, a message indicating potential delays should appear.
- **Verification Method**: Verify appearance of timeout message after threshold.

#### Acceptance Criteria 3.3: Cancelable Action
- **Testing Scenario**: During LLM processing, press the Escape key.
- **Expected Result**: Processing should be canceled, spinner should stop, and a cancellation message should appear.
- **Verification Method**: Verify that processing stops and the appropriate message is displayed.

## What is the Plan – Detailed Action Items

### 1. LLM Integration Foundation

#### 1.1 Create OpenAI Service Implementation
- Define a new package `internal/llm` to contain LLM processing functionality.
- Implement the `LLMProcessor` interface with an `OpenAIProcessor` implementation specifically for GPT-4o-mini.
- Create a configuration system for OpenAI API keys with secure storage.
- Implement structured output format using OpenAI's JSON mode.
- Implement error handling, retries, and timeouts for API requests.

#### 1.2 Implement Text Processing Logic
- Create a service for parsing and structuring unstructured text.
- Implement prompt construction for sending to OpenAI with structured output requirements.
- Create parsing logic to extract structured data from API responses.
- Add confidence scoring for different field extractions.

#### 1.3 Create Testing Harnesses
- Implement mock LLM processors for testing without actual API calls.
- Add test utilities for simulating various API response scenarios.
- Create test fixtures for different text input types.
- Implement API key validation tests with mocked responses.

### 2. API Key Management

#### 2.1 Create API Key Configuration System
- Implement secure storage for API keys.
- Create validation logic to verify API key functionality.
- Add configuration UI for managing API keys.
- Implement graceful degradation when key is not configured.

#### 2.2 Add Non-intrusive Notifications
- Create a subtle but clear notification when API key is missing.
- Implement inline help text for API key configuration.
- Add validation feedback when keys are entered.

#### 2.3 Implement Configuration Persistence
- Create file-based secure storage for API keys.
- Implement encryption for sensitive configuration data.
- Add command-line options for configuration management.

### 3. TUI Enhancement for Processing Feedback

#### 3.1 Create Progress Indicator Component
- Implement a spinner component using Bubble Tea primitives.
- Add status text display for processing information.
- Create state management for showing/hiding the spinner based on processing state.
- Implement timeout detection and warning messages.

#### 3.2 Implement Visual Field Feedback System
- Update `UserStoryForm` to track auto-populated fields.
- Implement confidence-based highlighting styles.
- Create visual indicators for showing which fields were automatically filled.
- Add state for tracking user confirmation of auto-populated fields.

#### 3.3 Add Keyboard Controls for Processing
- Implement Escape key handling to cancel ongoing LLM requests.
- Add keyboard shortcut for triggering processing manually.
- Create field navigation shortcuts for reviewing auto-populated fields.

### 4. Clipboard Integration and Paste Detection

#### 4.1 Implement Clipboard Monitoring
- Create a clipboard service that can detect paste events in the TUI.
- Add threshold detection to determine if pasted content should trigger LLM processing.
- Implement content type detection to handle different clipboard formats.

#### 4.2 Add Paste Event Handlers
- Update `UserStoryForm` to listen for paste events.
- Add state management for tracking clipboard processing.
- Implement throttling to prevent multiple simultaneous processing attempts.
- Add API key availability check before initiating processing.

#### 4.3 Create Context Management for Cancellation
- Implement context-based processing that can be canceled.
- Add timeout management to automatically cancel long-running requests.
- Create cleanup functions to restore UI state after cancellation.

### 5. Form Field Population and Validation

#### 5.1 Implement Population Logic
- Create mapping between OpenAI structured output format and form fields.
- Implement conditional population based on confidence thresholds.
- Add detection for conflicting or inconsistent extraction results.

#### 5.2 Add User Confirmation and Editing Flow
- Update fields to visually indicate when they need user confirmation.
- Implement easy editing and override of auto-populated fields.
- Add batch confirmation capability to accept all suggested fields.

#### 5.3 Create Error Handling and Recovery
- Implement error display in the UI with actionable messages.
- Add recovery mechanisms for partial processing failures.
- Create graceful fallbacks when processing fails entirely.
- Add specific error handling for API key validation issues.

### 6. Interface and Experience Polishing

#### 6.1 Refine Visual Design
- Implement color schemes for different field states (auto-populated, confirmed, edited).
- Add subtle animations for state transitions.
- Improve layout to accommodate processing indicators without disrupting form flow.
- Add non-intrusive API key configuration reminders.

#### 6.2 Enhance Help and Documentation
- Update help text to explain clipboard functionality.
- Add tooltips or status messages explaining auto-population.
- Create keyboard shortcut reference for processing actions.
- Add documentation for API key configuration.

#### 6.3 Implement Configuration Options
- Add settings for controlling LLM behavior (timeout, confidence thresholds).
- Create toggles for enabling/disabling auto-processing.
- Add persistence for user preferences.
- Add API key management commands.

### 7. Testing and Integration

#### 7.1 Unit Tests
- Create comprehensive tests for LLM processing functions.
- Implement UI component tests with simulated inputs.
- Add error case testing for all processing paths.
- Test API key validation and configuration.

#### 7.2 Integration Tests
- Implement end-to-end tests for the complete clipboard-to-form flow.
- Create test cases for the various acceptance criteria.
- Add performance testing for response time requirements.
- Test for proper handling of API key absence or invalidity.

#### 7.3 Final Integration
- Connect all components with proper dependency injection.
- Implement graceful feature degradation when OpenAI services are unavailable.
- Add necessary documentation for the enhanced features.
- Create a streamlined API key setup experience.

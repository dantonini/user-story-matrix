// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package llm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user-story-matrix/usm/internal/io"
)

// MockLLMProcessor is a mock implementation of the LLMProcessor interface for testing
type MockLLMProcessor struct {
	mock.Mock
}

func (m *MockLLMProcessor) ProcessUnstructuredText(ctx context.Context, text string) (UserStoryData, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(UserStoryData), args.Error(1)
}

func (m *MockLLMProcessor) GetConfidenceScores() map[string]float64 {
	args := m.Called()
	return args.Get(0).(map[string]float64)
}

func (m *MockLLMProcessor) IsConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLLMProcessor) ValidateConfiguration(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockLLMProcessor) Configure(config APIKeyConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockLLMProcessor) GetProcessingState() ProcessingState {
	args := m.Called()
	return args.Get(0).(ProcessingState)
}

func TestNewConfigManager(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	
	// Act
	manager := NewConfigManager(mockFS)
	
	// Assert
	assert.NotNil(t, manager)
	assert.Equal(t, mockFS, manager.fileSystem)
	assert.Equal(t, ".usm/config/llm_config.json", manager.configPath)
	assert.False(t, manager.config.IsValid)
}

func TestLoadConfigWhenFileDoesNotExist(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	// Act
	err := manager.LoadConfig()
	
	// Assert
	assert.NoError(t, err)
	assert.False(t, manager.config.IsValid)
	assert.Empty(t, manager.config.OpenAIKey)
}

func TestLoadConfigWithValidFile(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	// Create a valid config file
	validConfig := `{
		"openai_key": "test-key",
		"is_valid": true,
		"last_validated": "2025-01-01T12:00:00Z"
	}`
	mockFS.AddFile(manager.configPath, []byte(validConfig))
	
	// Act
	err := manager.LoadConfig()
	
	// Assert
	assert.NoError(t, err)
	assert.True(t, manager.config.IsValid)
	assert.Equal(t, "test-key", manager.config.OpenAIKey)
	assert.Equal(t, time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC), manager.config.LastValidated)
}

func TestLoadConfigWithReadError(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	// Add file but set up read error
	mockFS.AddFile(manager.configPath, []byte(`{}`))
	mockFS.SetReadFileError(manager.configPath, errors.New("read error"))
	
	// Act
	err := manager.LoadConfig()
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfigWithInvalidJSON(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	// Create invalid JSON
	mockFS.AddFile(manager.configPath, []byte(`{invalid json`))
	
	// Act
	err := manager.LoadConfig()
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestSaveConfig(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	manager.config = APIKeyConfig{
		OpenAIKey:     "test-key",
		IsValid:       true,
		LastValidated: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	
	// Act
	err := manager.SaveConfig()
	
	// Assert
	assert.NoError(t, err)
	assert.True(t, mockFS.Exists(manager.configPath))
	
	// Verify the saved content
	data, _ := mockFS.ReadFile(manager.configPath)
	assert.Contains(t, string(data), "test-key")
	assert.Contains(t, string(data), "true")
	assert.Contains(t, string(data), "2025-01-01T12:00:00Z")
}

func TestSaveConfigWithMkdirError(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	// Setup mkdir to fail
	mockFS.SetMkdirAllError(".usm/config", errors.New("mkdir error"))
	
	// Act
	err := manager.SaveConfig()
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create config directory")
}

func TestSaveConfigWithWriteError(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	// Make sure directory exists but write fails
	mockFS.AddFile(".usm/config", []byte{}) // Creates the directory
	mockFS.SetWriteFileError(manager.configPath, errors.New("write error"))
	
	// Act
	err := manager.SaveConfig()
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write config file")
}

func TestSetOpenAIKey(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	mockProcessor := new(MockLLMProcessor)
	mockProcessor.On("Configure", mock.Anything).Return(nil)
	mockProcessor.On("ValidateConfiguration", mock.Anything).Return(nil)
	
	// Act
	err := manager.SetOpenAIKey("new-api-key", mockProcessor)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "new-api-key", manager.config.OpenAIKey)
	assert.True(t, manager.config.IsValid)
	assert.NotEmpty(t, manager.config.LastValidated)
	
	// Verify the config was saved
	assert.True(t, mockFS.Exists(manager.configPath))
	mockProcessor.AssertExpectations(t)
}

func TestSetOpenAIKeyWithEmptyKey(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	mockProcessor := new(MockLLMProcessor)
	
	// Act
	err := manager.SetOpenAIKey("", mockProcessor)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, "OpenAI API key cannot be empty", err.Error())
	assert.False(t, manager.config.IsValid)
	
	// Verify the config was not saved
	assert.False(t, mockFS.Exists(manager.configPath))
	mockProcessor.AssertNotCalled(t, "Configure")
}

func TestSetOpenAIKeyWithConfigureError(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	mockProcessor := new(MockLLMProcessor)
	mockProcessor.On("Configure", mock.Anything).Return(errors.New("configuration error"))
	
	// Act
	err := manager.SetOpenAIKey("test-key", mockProcessor)
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to configure processor")
	assert.False(t, manager.config.IsValid)
	
	// Verify the config was not saved
	assert.False(t, mockFS.Exists(manager.configPath))
	mockProcessor.AssertExpectations(t)
}

func TestSetOpenAIKeyWithValidationError(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	mockProcessor := new(MockLLMProcessor)
	mockProcessor.On("Configure", mock.Anything).Return(nil)
	mockProcessor.On("ValidateConfiguration", mock.Anything).Return(errors.New("validation error"))
	
	// Act
	err := manager.SetOpenAIKey("test-key", mockProcessor)
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate API key")
	assert.False(t, manager.config.IsValid)
	
	// Verify the config was not saved
	assert.False(t, mockFS.Exists(manager.configPath))
	mockProcessor.AssertExpectations(t)
}

func TestGetOpenAIKey(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	manager.config.OpenAIKey = "test-key"
	
	// Act
	key := manager.GetOpenAIKey()
	
	// Assert
	assert.Equal(t, "test-key", key)
}

func TestIsOpenAIKeyConfigured(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	
	// Act & Assert
	
	// Not configured
	assert.False(t, manager.IsOpenAIKeyConfigured())
	
	// Valid but empty
	manager.config.IsValid = true
	manager.config.OpenAIKey = ""
	assert.False(t, manager.IsOpenAIKeyConfigured())
	
	// Not valid but has key
	manager.config.IsValid = false
	manager.config.OpenAIKey = "test-key"
	assert.False(t, manager.IsOpenAIKeyConfigured())
	
	// Fully configured
	manager.config.IsValid = true
	manager.config.OpenAIKey = "test-key"
	assert.True(t, manager.IsOpenAIKeyConfigured())
}

func TestGetConfig(t *testing.T) {
	// Arrange
	mockFS := io.NewMockFileSystem()
	manager := NewConfigManager(mockFS)
	manager.config = APIKeyConfig{
		OpenAIKey:     "test-key",
		IsValid:       true,
		LastValidated: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	
	// Act
	config := manager.GetConfig()
	
	// Assert
	assert.Equal(t, manager.config, config)
} 
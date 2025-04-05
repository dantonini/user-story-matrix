// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/user-story-matrix/usm/internal/io"
)

// ConfigManager handles the configuration for LLM processors
type ConfigManager struct {
	// fileSystem is the file system implementation
	fileSystem io.FileSystem
	
	// configPath is the path to the configuration file
	configPath string
	
	// config is the current configuration
	config APIKeyConfig
}

// NewConfigManager creates a new config manager
func NewConfigManager(fileSystem io.FileSystem) *ConfigManager {
	configDir := filepath.Join(".usm", "config")
	configPath := filepath.Join(configDir, "llm_config.json")
	
	return &ConfigManager{
		fileSystem: fileSystem,
		configPath: configPath,
		config: APIKeyConfig{
			IsValid: false,
		},
	}
}

// LoadConfig loads the configuration from the config file
func (m *ConfigManager) LoadConfig() error {
	// If the config file doesn't exist, return an empty config
	if !m.fileSystem.Exists(m.configPath) {
		return nil
	}
	
	// Read the config file
	data, err := m.fileSystem.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse the config
	err = json.Unmarshal(data, &m.config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return nil
}

// SaveConfig saves the configuration to the config file
func (m *ConfigManager) SaveConfig() error {
	// Create the config directory if it doesn't exist
	configDir := filepath.Dir(m.configPath)
	if !m.fileSystem.Exists(configDir) {
		err := m.fileSystem.MkdirAll(configDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}
	
	// Marshal the config
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write the config file
	err = m.fileSystem.WriteFile(m.configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// SetOpenAIKey sets the OpenAI API key
func (m *ConfigManager) SetOpenAIKey(key string, processor LLMProcessor) error {
	if key == "" {
		return errors.New("OpenAI API key cannot be empty")
	}
	
	// Update the processor with the new key
	err := processor.Configure(APIKeyConfig{
		OpenAIKey: key,
		IsValid:   false, // Will be validated below
	})
	if err != nil {
		return fmt.Errorf("failed to configure processor: %w", err)
	}
	
	// Validate the key
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err = processor.ValidateConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}
	
	// Update the config
	m.config.OpenAIKey = key
	m.config.IsValid = true
	m.config.LastValidated = time.Now()
	
	// Save the config
	return m.SaveConfig()
}

// GetOpenAIKey returns the OpenAI API key
func (m *ConfigManager) GetOpenAIKey() string {
	return m.config.OpenAIKey
}

// IsOpenAIKeyConfigured returns whether the OpenAI API key is configured
func (m *ConfigManager) IsOpenAIKeyConfigured() bool {
	return m.config.IsValid && m.config.OpenAIKey != ""
}

// GetConfig returns the current API key configuration
func (m *ConfigManager) GetConfig() APIKeyConfig {
	return m.config
} 
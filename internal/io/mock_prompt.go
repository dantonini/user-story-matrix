// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package io

import (
	"fmt"

	"github.com/stretchr/testify/mock"
)

// MockIO implements both UserInput and UserOutput interfaces for testing
type MockIO struct {
	// For testing input responses
	PromptResponses     []string
	SelectResponses     []int
	MultiSelectResponses [][]int
	PromptIndex         int
	SelectIndex         int
	MultiSelectIndex    int

	// For capturing output
	Messages        []string
	SuccessMessages []string
	ErrorMessages   []string
	WarningMessages []string
	ProgressMessages []string
	StepMessages    []string
	Tables          []struct {
		Headers []string
		Rows    [][]string
	}
	
	// Debug mode flag
	DebugEnabled bool
}

// NewMockIO creates a new instance of MockIO
func NewMockIO() *MockIO {
	return &MockIO{
		PromptResponses:      []string{},
		SelectResponses:      []int{},
		MultiSelectResponses: [][]int{},
		Messages:             []string{},
		SuccessMessages:      []string{},
		ErrorMessages:        []string{},
		WarningMessages:      []string{},
		ProgressMessages:     []string{},
		StepMessages:         []string{},
		Tables:               []struct{Headers []string; Rows [][]string}{},
		DebugEnabled:         false,
	}
}

// Prompt returns the next predefined response
func (m *MockIO) Prompt(message string) (string, error) {
	if m.PromptIndex >= len(m.PromptResponses) {
		return "", nil
	}
	response := m.PromptResponses[m.PromptIndex]
	m.PromptIndex++
	return response, nil
}

// Select returns the next predefined selection
func (m *MockIO) Select(message string, options []string) (int, error) {
	if m.SelectIndex >= len(m.SelectResponses) {
		return 0, nil
	}
	selection := m.SelectResponses[m.SelectIndex]
	m.SelectIndex++
	return selection, nil
}

// MultiSelect returns the next predefined multi-selection
func (m *MockIO) MultiSelect(message string, options []string) ([]int, error) {
	if m.MultiSelectIndex >= len(m.MultiSelectResponses) {
		return []int{}, nil
	}
	selection := m.MultiSelectResponses[m.MultiSelectIndex]
	m.MultiSelectIndex++
	return selection, nil
}

// Print captures a regular message
func (m *MockIO) Print(message string) {
	m.Messages = append(m.Messages, message)
}

// PrintSuccess captures a success message
func (m *MockIO) PrintSuccess(message string) {
	m.SuccessMessages = append(m.SuccessMessages, message)
}

// PrintError captures an error message
func (m *MockIO) PrintError(message string) {
	m.ErrorMessages = append(m.ErrorMessages, message)
}

// PrintWarning captures a warning message
func (m *MockIO) PrintWarning(message string) {
	m.WarningMessages = append(m.WarningMessages, message)
}

// PrintProgress captures a progress message
func (m *MockIO) PrintProgress(message string) {
	m.ProgressMessages = append(m.ProgressMessages, message)
}

// PrintStep captures a step message
func (m *MockIO) PrintStep(stepNumber int, totalSteps int, description string) {
	message := fmt.Sprintf("Step %d/%d: %s", stepNumber, totalSteps, description)
	m.StepMessages = append(m.StepMessages, message)
}

// IsDebugEnabled returns the debug mode status
func (m *MockIO) IsDebugEnabled() bool {
	return m.DebugEnabled
}

// PrintTable captures table data
func (m *MockIO) PrintTable(headers []string, rows [][]string) {
	m.Tables = append(m.Tables, struct{Headers []string; Rows [][]string}{
		Headers: headers,
		Rows:    rows,
	})
}

// The following is for testify/mock style testing

// MockUserIO is a mock implementation of UserInput and UserOutput using testify/mock
type MockUserIO struct {
	mock.Mock
}

// Prompt mocks the Prompt method
func (m *MockUserIO) Prompt(message string) (string, error) {
	args := m.Called(message)
	return args.String(0), args.Error(1)
}

// Select mocks the Select method
func (m *MockUserIO) Select(message string, options []string) (int, error) {
	args := m.Called(message, options)
	return args.Int(0), args.Error(1)
}

// MultiSelect mocks the MultiSelect method
func (m *MockUserIO) MultiSelect(message string, options []string) ([]int, error) {
	args := m.Called(message, options)
	return args.Get(0).([]int), args.Error(1)
}

// Print mocks the Print method
func (m *MockUserIO) Print(message string) {
	m.Called(message)
}

// PrintSuccess mocks the PrintSuccess method
func (m *MockUserIO) PrintSuccess(message string) {
	m.Called(message)
}

// PrintError mocks the PrintError method
func (m *MockUserIO) PrintError(message string) {
	m.Called(message)
}

// PrintWarning mocks the PrintWarning method
func (m *MockUserIO) PrintWarning(message string) {
	m.Called(message)
}

// PrintProgress mocks the PrintProgress method
func (m *MockUserIO) PrintProgress(message string) {
	m.Called(message)
}

// PrintStep mocks the PrintStep method
func (m *MockUserIO) PrintStep(stepNumber int, totalSteps int, description string) {
	m.Called(stepNumber, totalSteps, description)
}

// IsDebugEnabled mocks the IsDebugEnabled method
func (m *MockUserIO) IsDebugEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

// PrintTable mocks the PrintTable method
func (m *MockUserIO) PrintTable(headers []string, rows [][]string) {
	m.Called(headers, rows)
} 
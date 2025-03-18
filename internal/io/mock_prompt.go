package io

import (
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
	Tables          []struct {
		Headers []string
		Rows    [][]string
	}
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
		Tables:               []struct{Headers []string; Rows [][]string}{},
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

// PrintTable mocks the PrintTable method
func (m *MockUserIO) PrintTable(headers []string, rows [][]string) {
	m.Called(headers, rows)
} 
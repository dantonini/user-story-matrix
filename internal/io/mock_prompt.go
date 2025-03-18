package io

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
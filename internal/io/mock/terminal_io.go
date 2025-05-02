// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package mock

import (
	"bytes"
	"sync"
)

// MockTerminalIO is a testable implementation of the TerminalIO interface
type MockTerminalIO struct {
	mutex       sync.Mutex
	Buffer      bytes.Buffer
	calls       []string
	lastMessage string
}

// Print records a call to Print and the message
func (m *MockTerminalIO) Print(message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls = append(m.calls, "Print")
	m.lastMessage = message
	m.Buffer.WriteString(message + "\n")
}

// PrintSuccess records a call to PrintSuccess and the message
func (m *MockTerminalIO) PrintSuccess(message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls = append(m.calls, "PrintSuccess")
	m.lastMessage = message
	m.Buffer.WriteString("✓ " + message + "\n")
}

// PrintError records a call to PrintError and the message
func (m *MockTerminalIO) PrintError(message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls = append(m.calls, "PrintError")
	m.lastMessage = message
	m.Buffer.WriteString("✗ " + message + "\n")
}

// PrintProgress records a call to PrintProgress and the message
func (m *MockTerminalIO) PrintProgress(message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls = append(m.calls, "PrintProgress")
	m.lastMessage = message
	m.Buffer.WriteString("⟳ " + message + "\n")
}

// PrintTable records a call to PrintTable and writes a simple representation
func (m *MockTerminalIO) PrintTable(headers []string, rows [][]string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls = append(m.calls, "PrintTable")
	
	// Write headers
	for i, header := range headers {
		if i > 0 {
			m.Buffer.WriteString("\t")
		}
		m.Buffer.WriteString(header)
	}
	m.Buffer.WriteString("\n")
	
	// Write rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				m.Buffer.WriteString("\t")
			}
			m.Buffer.WriteString(cell)
		}
		m.Buffer.WriteString("\n")
	}
}

// GetCalls returns the list of recorded function calls
func (m *MockTerminalIO) GetCalls() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]string{}, m.calls...)
}

// GetLastMessage returns the last recorded message
func (m *MockTerminalIO) GetLastMessage() string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.lastMessage
}

// GetOutput returns the entire buffer content
func (m *MockTerminalIO) GetOutput() string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Buffer.String()
}

// Reset clears all recorded calls and messages
func (m *MockTerminalIO) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls = nil
	m.lastMessage = ""
	m.Buffer.Reset()
} 
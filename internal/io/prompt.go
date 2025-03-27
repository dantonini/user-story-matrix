// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package io

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UserInput defines the interface for getting input from the user
type UserInput interface {
	Prompt(message string) (string, error)
	Select(message string, options []string) (int, error)
	MultiSelect(message string, options []string) ([]int, error)
}

// UserOutput defines the interface for displaying output to the user
type UserOutput interface {
	Print(message string)
	PrintSuccess(message string)
	PrintError(message string)
	PrintTable(headers []string, rows [][]string)
	PrintWarning(message string)
	PrintProgress(message string)
	PrintStep(stepNumber int, totalSteps int, description string)
}

// TerminalIO implements both UserInput and UserOutput interfaces for terminal interactions
type TerminalIO struct {
	styles struct {
		success lipgloss.Style
		error   lipgloss.Style
		info    lipgloss.Style
		header  lipgloss.Style
		cell    lipgloss.Style
		warning lipgloss.Style
		progress lipgloss.Style
		step    lipgloss.Style
	}
}

// NewTerminalIO creates a new instance of TerminalIO
func NewTerminalIO() *TerminalIO {
	t := &TerminalIO{}
	
	// Configure styles
	t.styles.success = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	t.styles.error = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	t.styles.info = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	t.styles.header = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	t.styles.cell = lipgloss.NewStyle().PaddingRight(2)
	t.styles.warning = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	t.styles.progress = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
	t.styles.step = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	
	return t
}

// Prompt displays a message and waits for user input
func (t *TerminalIO) Prompt(message string) (string, error) {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()

	m := promptModel{
		textInput: ti,
		prompt:    message,
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}

	return result.(promptModel).textInput.Value(), nil
}

// Select displays a list of options and returns the selected index
func (t *TerminalIO) Select(message string, options []string) (int, error) {
	items := make([]list.Item, len(options))
	for i, option := range options {
		items[i] = selectItem{
			title: option,
			desc:  "",
		}
	}

	m := selectModel{
		list:  list.New(items, list.NewDefaultDelegate(), 0, 0),
		title: message,
	}
	m.list.Title = message

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return -1, err
	}

	return result.(selectModel).selected, nil
}

// MultiSelect displays a list of options and returns the selected indices
func (t *TerminalIO) MultiSelect(message string, options []string) ([]int, error) {
	items := make([]multiSelectItem, len(options))
	for i, option := range options {
		items[i] = multiSelectItem{
			title:    option,
			selected: false,
		}
	}

	m := multiSelectModel{
		items:     items,
		title:     message,
		cursor:    0,
		confirmed: false,
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	resultModel := result.(multiSelectModel)
	if !resultModel.confirmed {
		return nil, fmt.Errorf("selection canceled")
	}

	var selected []int
	for i, item := range resultModel.items {
		if item.selected {
			selected = append(selected, i)
		}
	}

	return selected, nil
}

// Print displays a message
func (t *TerminalIO) Print(message string) {
	fmt.Println(message)
}

// PrintSuccess displays a success message
func (t *TerminalIO) PrintSuccess(message string) {
	fmt.Println(t.styles.success.Render("✓ " + message))
}

// PrintError displays an error message
func (t *TerminalIO) PrintError(message string) {
	fmt.Println(t.styles.error.Render("✗ " + message))
}

// PrintTable displays data in a table format
func (t *TerminalIO) PrintTable(headers []string, rows [][]string) {
	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print headers
	headerCells := make([]string, len(headers))
	for i, header := range headers {
		headerCells[i] = t.styles.header.Width(colWidths[i]).Render(header)
	}
	fmt.Println(strings.Join(headerCells, " "))

	// Print separator
	sep := make([]string, len(headers))
	for i, width := range colWidths {
		sep[i] = strings.Repeat("─", width)
	}
	fmt.Println(strings.Join(sep, " "))

	// Print rows
	for _, row := range rows {
		rowCells := make([]string, len(row))
		for i, cell := range row {
			if i < len(colWidths) {
				rowCells[i] = t.styles.cell.Width(colWidths[i]).Render(cell)
			}
		}
		fmt.Println(strings.Join(rowCells, " "))
	}
}

// PrintWarning displays a warning message
func (t *TerminalIO) PrintWarning(message string) {
	fmt.Println(t.styles.warning.Render(message))
}

// PrintProgress displays a progress message
func (t *TerminalIO) PrintProgress(message string) {
	fmt.Println(t.styles.progress.Render(message))
}

// PrintStep displays a step progress message
func (t *TerminalIO) PrintStep(stepNumber int, totalSteps int, description string) {
	message := fmt.Sprintf("Step %d/%d: %s", stepNumber, totalSteps, description)
	fmt.Println(t.styles.step.Render(message))
}

// Mock implementations of models for bubbletea

// promptModel is a model for text input
type promptModel struct {
	textInput textinput.Model
	prompt    string
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m promptModel) View() string {
	return fmt.Sprintf("%s\n%s\n", m.prompt, m.textInput.View())
}

// selectItem represents an item in the selection list
type selectItem struct {
	title, desc string
}

func (i selectItem) Title() string       { return i.title }
func (i selectItem) Description() string { return i.desc }
func (i selectItem) FilterValue() string { return i.title }

// selectModel is a model for single selection
type selectModel struct {
	list     list.Model
	title    string
	selected int
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			m.selected = m.list.Index()
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectModel) View() string {
	return m.list.View()
}

// multiSelectItem represents an item in the multi-selection list
type multiSelectItem struct {
	title    string
	selected bool
}

// multiSelectModel is a model for multiple selection
type multiSelectModel struct {
	items     []multiSelectItem
	title     string
	cursor    int
	confirmed bool
}

func (m multiSelectModel) Init() tea.Cmd {
	return nil
}

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			// Toggle selection
			m.items[m.cursor].selected = !m.items[m.cursor].selected
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiSelectModel) View() string {
	var s strings.Builder

	s.WriteString(m.title + "\n\n")

	for i, item := range m.items {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := "[ ]"
		if item.selected {
			checked = "[x]"
		}

		s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, checked, item.title))
	}

	s.WriteString("\nPress space to select, enter to confirm, q to quit\n")

	return s.String()
} 
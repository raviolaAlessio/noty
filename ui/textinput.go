package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

var inputTitleStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)

type textinputModel struct {
	textInput textinput.Model
	err       error
	output    *string
	header    string
	exit      *bool
}

// Intialize a Text Input
func NewTextInput(
	header string,
	output *string,
	placeholder string,
	exit *bool,
) textinputModel {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = placeholder
	ti.CharLimit = 156
	ti.Width = 80

	return textinputModel{
		textInput: ti,
		err:       nil,
		output:    output,
		header:    inputTitleStyle.Render(header),
		exit:      exit,
	}
}

// Intialize a Text Input with a validator
func NewValidated(
	header string,
	output *string,
	placeholder string,
	exit *bool,
	validator func(string) error,
) textinputModel {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = placeholder
	ti.CharLimit = 156
	ti.Width = 20
	ti.Validate = validator

	return textinputModel{
		textInput: ti,
		err:       nil,
		output:    output,
		header:    inputTitleStyle.Render(header),
		exit:      exit,
	}
}

func (m textinputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textinputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.textInput.Value()) > 1 {
				*m.output = m.textInput.Value()
				return m, tea.Quit
			} else if m.textInput.Placeholder != "" {
				*m.output = m.textInput.Placeholder
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			*m.exit = true
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		*m.exit = true
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m textinputModel) View() string {
	return fmt.Sprintf("\n%s\n%s\n\n",
		m.header,
		m.textInput.View(),
	)
}

func (m textinputModel) Err() string {
	return m.err.Error()
}

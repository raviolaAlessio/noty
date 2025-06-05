package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)


type textinputmodel struct {
	textInput textinput.Model
	err       error
	output    *string
	header    string
	exit      bool
}

// Intialize a Text Input
func NewTextInput(
	header string,
	output *string,
	placeholder string,
) textinputmodel {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = placeholder
	ti.CharLimit = 156
	ti.Width = 80

	return textinputmodel{
		textInput: ti,
		err:       nil,
		output:    output,
		header:    TitleStyle.Render(header),
		exit:      false,
	}
}

// Intialize a Text Input with a validator
func NewValidated(
	header string,
	output *string,
	placeholder string,
	validator func(string) error,
) textinputmodel {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = placeholder
	ti.CharLimit = 156
	ti.Width = 20
	ti.Validate = validator

	return textinputmodel{
		textInput: ti,
		err:       nil,
		output:    output,
		header:    TitleStyle.Render(header),
		exit:      false,
	}
}

func (m textinputmodel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textinputmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.exit = true
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case error:
		m.err = msg
		m.exit = true
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m textinputmodel) View() string {
	return fmt.Sprintf("\n%s\n%s\n\n",
		m.header,
		m.textInput.View(),
	)
}

func (m textinputmodel) Err() string {
	return m.err.Error()
}

func (m textinputmodel) Run() (bool, error) {
	tp := tea.NewProgram(m)
	if _, err := tp.Run(); err != nil {
		return m.exit, err
	}
	if m.exit {
		if err := tp.ReleaseTerminal(); err != nil {
			log.Fatal(err)
		}
	}
	return m.exit, nil
}

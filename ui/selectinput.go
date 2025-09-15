package ui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle       = lipgloss.NewStyle().Foreground(Accent)
	selectedItemStyle = lipgloss.NewStyle().Foreground(Accent)
)

type SelectItem[T any] struct {
	Value T
	Label string
}

func NewSelectItem[T any](label string, value T) SelectItem[T] {
	return SelectItem[T]{
		Label: label,
		Value: value,
	}
}

type modelselectinput[T any] struct {
	cursor    int
	header    string
	items     []SelectItem[T]
	selection *T
	exit      bool
}

func NewSelectInput[T any](
	header string,
	items []SelectItem[T],
	selection *T,
) modelselectinput[T] {
	return modelselectinput[T]{
		cursor:    0,
		header:    TitleStyle.Render(header),
		items:     items,
		selection: selection,
		exit:      false,
	}
}

func (m modelselectinput[T]) Init() tea.Cmd {
	return nil
}

func (m modelselectinput[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor -= 1
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor += 1
			}
		case "enter":
			*m.selection = m.items[m.cursor].Value
			return m, tea.Quit

		case "ctrl+c", "esc":
			m.exit = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m modelselectinput[T]) View() string {
	s := fmt.Sprintf("%s\n", m.header)

	for i, item := range m.items {
		cursor := " "
		label := item.Label

		if m.cursor == i {
			cursor = cursorStyle.Render(">")
			label = selectedItemStyle.Render(label)
		}

		s += fmt.Sprintf("%s %s\n", cursor, label)
	}

	return s + "\n"
}

func (m modelselectinput[T]) Run() (bool, error) {
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

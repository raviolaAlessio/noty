package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type StopSpinnerMsg struct {
	err error
}

func (s StopSpinnerMsg) Error() string {
	return s.err.Error()
}

type SpinnerTask struct {
	Header string
	Task   func() error
}

func (s SpinnerTask) Do() tea.Msg {
	err := s.Task()
	return StopSpinnerMsg{err: err}
}

type spinnerModel struct {
	spinner spinner.Model
	task    SpinnerTask
	err     error
	done    bool
	ok      *bool
}

func NewSpinner(task SpinnerTask, ok *bool) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Line

	return spinnerModel{
		spinner: s,
		task:    task,
		err:     nil,
		done:    false,
		ok:      ok,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.task.Do)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case StopSpinnerMsg:
		m.done = true
		*m.ok = true
		if msg.err != nil {
			*m.ok = false
			m.err = msg
		}
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m spinnerModel) View() string {
	s := ""
	if !m.done {
		s += InfoStyle.Render(fmt.Sprintf("%s %s", m.spinner.View(), m.task.Header))
	} else {
		if m.err != nil {
			s += ErrorStyle.Render(fmt.Sprintf("* %s ... Failed: %v", m.task.Header, m.err))
		} else {
			s += SuccessStyle.Render(fmt.Sprintf("* %s ... Done", m.task.Header))
		}
	}

	s += "\n"
	return s
}

func (m spinnerModel) Err() error {
	return m.err
}

func Spin(
	header string,
	task func() error,
) (bool, error) {
	var ok bool
	tp := tea.NewProgram(NewSpinner(
		SpinnerTask{
			Header: header,
			Task:   task,
		},
		&ok,
	))
	if _, err := tp.Run(); err != nil {
		return false, err
	}
	return ok, nil
}

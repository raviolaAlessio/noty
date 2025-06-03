package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var ErrorStyle = lipgloss.NewStyle().Foreground(ErrorFg).Bold(true)
var InfoStyle = lipgloss.NewStyle().Foreground(Fg).Faint(true)
var SuccessStyle = lipgloss.NewStyle().Foreground(SuccessFg)

func PrintlnfError(format string, a ...any) {
	fmt.Fprint(os.Stderr, ErrorStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

func PrintlnfInfo(format string, a ...any) {
	fmt.Fprint(os.Stderr, InfoStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

func PrintlnfSuccess(format string, a ...any) {
	fmt.Fprint(os.Stderr, SuccessStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

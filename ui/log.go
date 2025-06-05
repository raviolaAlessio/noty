package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var SuccessStyle = lipgloss.NewStyle().Foreground(Success)
var ErrorStyle = lipgloss.NewStyle().Foreground(Error).Bold(true)
var InfoStyle = lipgloss.NewStyle().Foreground(Fg).Faint(true)

func PrintlnfError(format string, a ...any) {
	fmt.Fprint(os.Stderr, ErrorStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

func PrintfInfo(format string, a ...any) {
	fmt.Fprint(os.Stderr, InfoStyle.Render(fmt.Sprintf(format, a...)))
}

func PrintlnfInfo(format string, a ...any) {
	fmt.Fprint(os.Stderr, InfoStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

func PrintlnfSuccess(format string, a ...any) {
	fmt.Fprint(os.Stderr, SuccessStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

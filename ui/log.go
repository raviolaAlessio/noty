package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var SuccessStyle = lipgloss.NewStyle().Foreground(Success)
var ErrorStyle = lipgloss.NewStyle().Foreground(Error).Bold(true)
var InfoStyle = lipgloss.NewStyle().Foreground(Fg).Faint(true)
var WarnStyle = lipgloss.NewStyle().Foreground(Accent).Faint(true)

func PrintlnfError(format string, a ...any) {
	fmt.Fprint(os.Stderr, ErrorStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

func PrintfInfo(format string, a ...any) {
	fmt.Fprint(os.Stderr, InfoStyle.Render(fmt.Sprintf(format, a...)))
}

func PrintlnInfo(string string) {
	fmt.Fprint(os.Stderr, InfoStyle.Render(string) + "\n")
}

func PrintlnfInfo(format string, a ...any) {
	fmt.Fprint(os.Stderr, InfoStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

func PrintfWarn(format string, a ...any) {
	fmt.Fprint(os.Stderr, WarnStyle.Render(fmt.Sprintf(format, a...)))
}

func PrintlnWarn(string string) {
	fmt.Fprint(os.Stderr, WarnStyle.Render(string) + "\n")
}

func PrintlnfWarn(format string, a ...any) {
	fmt.Fprint(os.Stderr, WarnStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

func PrintlnfSuccess(format string, a ...any) {
	fmt.Fprint(os.Stderr, SuccessStyle.Render(fmt.Sprintf(format, a...)) + "\n")
}

package ui

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	PrimaryFg = lipgloss.Color("0")
	Primary   = lipgloss.Color("4")

	SecondaryFg = lipgloss.Color("0")
	Secondary   = lipgloss.Color("3")

	Fg    = lipgloss.Color("15")
	DimFg = lipgloss.Color("8")

	Accent   = lipgloss.Color("3")
	Error    = lipgloss.Color("1")
	Success  = lipgloss.Color("2")
	Progress = lipgloss.Color("11")

	PriorityHigh   = lipgloss.Color("1")
	PriorityMedium = lipgloss.Color("3")
	PriorityLow    = lipgloss.Color("2")
)

var TitleStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)

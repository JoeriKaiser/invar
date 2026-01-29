package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginLeft(2)

	NormalStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	HighPriorityStyle = lipgloss.NewStyle().
				Foreground(ColorHigh)

	MediumPriorityStyle = lipgloss.NewStyle().
				Foreground(ColorMedium)

	LowPriorityStyle = lipgloss.NewStyle().
				Foreground(ColorLow)

	CompletedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Strikethrough(true)

	OverdueStyle = lipgloss.NewStyle().
			Foreground(ColorHigh).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Padding(0, 1)
)

func PriorityStyle(p string) lipgloss.Style {
	switch p {
	case "high":
		return HighPriorityStyle
	case "medium":
		return MediumPriorityStyle
	case "low":
		return LowPriorityStyle
	default:
		return NormalStyle
	}
}

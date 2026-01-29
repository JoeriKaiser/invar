package ui

import "github.com/charmbracelet/lipgloss"

// Tokyo Night color palette.
var (
	ColorBorder    = lipgloss.Color("#414868")
	ColorPrimary   = lipgloss.Color("#7AA2F7")
	ColorPrimaryDim = lipgloss.Color("#3D59A1")
	ColorFg        = lipgloss.Color("#C0CAF5")
	ColorMuted     = lipgloss.Color("#565F89")
	ColorHigh      = lipgloss.Color("#F7768E")
	ColorMedium    = lipgloss.Color("#E0AF68")
	ColorLow       = lipgloss.Color("#9ECE6A")
	ColorDark      = lipgloss.Color("#1A1B26")
)

// Component styles.
var (
	HeaderBar = lipgloss.NewStyle().
			Foreground(ColorFg).
			Bold(true).
			Padding(0, 2)

	TabActive = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorDark).
			Padding(0, 1).
			Bold(true)

	TabInactive = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	TaskRowNormal = lipgloss.NewStyle().
			Foreground(ColorFg).
			Padding(0, 2)

	TaskRowSelected = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 2)

	PriorityPillHigh = lipgloss.NewStyle().
				Background(ColorHigh).
				Foreground(ColorDark).
				Padding(0, 1).
				Bold(true)

	PriorityPillMed = lipgloss.NewStyle().
			Background(ColorMedium).
			Foreground(ColorDark).
			Padding(0, 1).
			Bold(true)

	PriorityPillLow = lipgloss.NewStyle().
			Background(ColorLow).
			Foreground(ColorDark).
			Padding(0, 1).
			Bold(true)

	DeadlineNormal = lipgloss.NewStyle().
			Foreground(ColorMuted)

	DeadlineOverdue = lipgloss.NewStyle().
			Foreground(ColorHigh).
			Bold(true)

	FooterStats = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 2)

	FooterHelp = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 2)

	OverlayCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Width(60)

	OverlayTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
)

// PriorityPill renders a styled priority badge.
func PriorityPill(p string) string {
	switch p {
	case "high":
		return PriorityPillHigh.Render(" HIGH ")
	case "medium":
		return PriorityPillMed.Render(" MED ")
	case "low":
		return PriorityPillLow.Render(" LOW ")
	default:
		return ""
	}
}

package l1

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Base       lipgloss.Style
	Title      lipgloss.Style
	Muted      lipgloss.Style
	Selected   lipgloss.Style
	Checked    lipgloss.Style
	Unchecked  lipgloss.Style
	Error      lipgloss.Style
	Help       lipgloss.Style
	Progress   lipgloss.Style
	Background lipgloss.Color
}

func DefaultTheme() Theme {
	muted := lipgloss.Color("244")
	strong := lipgloss.Color("252")
	accent := lipgloss.Color("115")
	warn := lipgloss.Color("203")

	return Theme{
		Base: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(strong).
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(muted),
		Selected: lipgloss.NewStyle().
			Foreground(strong).
			Bold(true),
		Checked: lipgloss.NewStyle().
			Foreground(accent),
		Unchecked: lipgloss.NewStyle().
			Foreground(muted),
		Error: lipgloss.NewStyle().
			Foreground(warn),
		Help: lipgloss.NewStyle().
			Foreground(muted),
		Progress: lipgloss.NewStyle().
			Foreground(accent),
		Background: lipgloss.Color("235"),
	}
}

package counter

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Counter   lipgloss.Style
	NameLabel lipgloss.Style
}

func defaultStyles() Styles {
	return Styles{
		Counter: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("241")),
		NameLabel: lipgloss.NewStyle().
			Align(lipgloss.Center),
	}
}

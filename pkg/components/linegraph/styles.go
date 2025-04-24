package linegraph

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Bars  lipgloss.Style
	Graph lipgloss.Style
}

func defaultStyles() Styles {
	return Styles{
		Bars: lipgloss.NewStyle().
			Foreground(lipgloss.Color("36")),
		Graph: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("241")),
	}
}

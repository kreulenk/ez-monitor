package linegraph

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Graph lipgloss.Style
}

func defaultStyles() Styles {
	return Styles{
		Graph: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("241")).
			Foreground(lipgloss.Color("36")),
	}
}

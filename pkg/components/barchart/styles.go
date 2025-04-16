package barchart

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Graph         lipgloss.Style
	BackGroundBar lipgloss.Style // What is left over in the background behind the current value
	ValueBar      lipgloss.Style // The actual value being filled in
	BarText       lipgloss.Style // The units displayed inside the bar
	NameLabel     lipgloss.Style
}

func defaultStyles() Styles {
	return Styles{
		Graph: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("241")),
		BackGroundBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("54")),
		ValueBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("57")),
		BarText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")),
		NameLabel: lipgloss.NewStyle().
			Align(lipgloss.Center),
	}
}

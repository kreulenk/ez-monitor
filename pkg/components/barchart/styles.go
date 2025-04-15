package barchart

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	BackGroundBar lipgloss.Style // What is left over in the background behind the current value
	ValueBar      lipgloss.Style // The actual value being filled in
	BarText       lipgloss.Style // The units displayed inside the bar
	NameLabel     lipgloss.Style
}

func defaultStyles() Styles {
	return Styles{
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

package barchart

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"math"
	"strings"
)

type Model struct {
	statName string
	unit     string

	minValue     float64
	maxValue     float64
	currentValue float64
	width        int
	height       int

	styles Styles
}

func New(statName, unit string, minValue, maxValue float64) Model {
	return Model{
		statName: statName,
		unit:     unit,

		minValue: minValue,
		maxValue: maxValue,

		styles: defaultStyles(),
	}
}

// Init initialises the baseModel on program load. It partly implements the tea.Model interface.
func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetCurrentValue(v float64) {
	m.currentValue = v
}

func (m *Model) SetMaxValue(v float64) {
	m.maxValue = v
}

func (m *Model) SetWidth(v int) {
	m.width = v
}

func (m *Model) SetHeight(v int) {
	m.height = v - 2
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	barPercent := m.currentValue / (m.maxValue - m.minValue)
	valueBarHeight := int(math.Floor(barPercent * float64(m.height)))

	// Create the full bar with background and overlay the value bar
	bars := make([]string, m.height)
	for i := 0; i < m.height; i++ {
		if i == 0 && m.maxValue != m.minValue {
			bars[i] = overlayTextOnBar(m.width, fmt.Sprintf("%.1f %s", m.maxValue, m.unit), m.styles.BackGroundBar)
		} else if i == m.height-valueBarHeight-1 {
			bars[i] = overlayTextOnBar(m.width, fmt.Sprintf("%.1f %s", m.currentValue, m.unit), m.styles.BackGroundBar)
		} else if i < m.height-valueBarHeight {
			bars[i] = m.styles.BackGroundBar.Render(strings.Repeat("█", m.width))
		} else {
			bars[i] = m.styles.ValueBar.Render(strings.Repeat("█", m.width))
		}
	}

	barView := lipgloss.JoinVertical(lipgloss.Top, bars...)
	statNameView := lipgloss.NewStyle().MarginLeft(m.width/2 - 3).Render(m.statName)

	return lipgloss.JoinVertical(lipgloss.Top, barView, statNameView)
}

// overlayTextOnBar will take in a bar of a specific length and will overlay given text onto the
// bar. The text will be centered. If the text width is too long, the … character will be used to
// truncate the value
func overlayTextOnBar(barWidth int, text string, barStyle lipgloss.Style) string {
	textWidth := lipgloss.Width(text)

	if textWidth > barWidth {
		// Truncate the text and add ellipsis
		if barWidth > 1 {
			text = text[:barWidth-1] + "…"
		} else {
			text = "…"
		}
		textWidth = lipgloss.Width(text)
	}

	leftBarWidth := (barWidth - textWidth) / 2
	rightBarWidth := barWidth - textWidth - leftBarWidth

	leftBar := barStyle.Render(strings.Repeat("█", leftBarWidth))
	rightBar := barStyle.Render(strings.Repeat("█", rightBarWidth))
	return leftBar + text + rightBar
}

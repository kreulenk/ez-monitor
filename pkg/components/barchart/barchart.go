package barchart

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/ez-monitor/pkg/renderutils"
	"github.com/kreulenk/ez-monitor/pkg/unit"
	"math"
	"strings"
)

type Model struct {
	statName string
	unit     unit.DataType

	minValue     float64
	maxValue     float64
	currentValue float64
	width        int
	height       int

	dataCollectionErr error

	styles Styles
}

func New(statName string, unit unit.DataType, minValue, maxValue float64) Model {
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
	m.dataCollectionErr = nil
}

func (m *Model) SetDataCollectionErr(err error) {
	m.dataCollectionErr = err
}

func (m *Model) SetMaxValue(v float64) {
	m.maxValue = v
}

func (m *Model) SetWidth(v int) {
	m.width = v
}

func (m *Model) SetHeight(v int) {
	m.height = renderutils.Max(0, v)
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	statNameView := lipgloss.NewStyle().Width(m.width - 2).AlignHorizontal(lipgloss.Center).Render(m.statName)
	totalHeightOfBar := renderutils.Max(0, m.height-3) // 2 for border and 1 for label

	if m.dataCollectionErr != nil {
		errText := lipgloss.NewStyle().Width(m.width - 2).AlignHorizontal(lipgloss.Center).Render(m.dataCollectionErr.Error())
		errText = lipgloss.NewStyle().PaddingBottom(totalHeightOfBar - lipgloss.Height(errText)).Render(errText) // Add padding so the label is at the bottom

		text := lipgloss.JoinVertical(lipgloss.Center, errText, statNameView)
		return m.styles.Graph.Width(m.width - 2).Height(totalHeightOfBar).Render(text)
	}

	barPercent := m.currentValue / (m.maxValue - m.minValue)
	valueBarHeight := int(math.Floor(barPercent * float64(totalHeightOfBar)))

	// Create the full bar with background and overlay the value bar
	bars := make([]string, totalHeightOfBar)
	for i := 0; i < totalHeightOfBar; i++ {
		if i == 0 && m.maxValue != m.minValue {
			bars[i] = overlayTextOnBar(m.width, unit.DisplayType(m.maxValue, m.unit), m.styles.BackGroundBar)
		} else if i == totalHeightOfBar-valueBarHeight-1 {
			bars[i] = overlayTextOnBar(m.width, unit.DisplayType(m.currentValue, m.unit), m.styles.BackGroundBar)
		} else if i < totalHeightOfBar-valueBarHeight {
			bars[i] = m.styles.BackGroundBar.Render(strings.Repeat("█", m.width))
		} else {
			bars[i] = m.styles.ValueBar.Render(strings.Repeat("█", m.width))
		}
	}

	barView := lipgloss.JoinVertical(lipgloss.Top, bars...)
	graph := lipgloss.JoinVertical(lipgloss.Top, barView, statNameView)
	return m.styles.Graph.Render(graph)
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

	leftBarWidth := renderutils.Max(0, (barWidth-textWidth)/2)
	rightBarWidth := renderutils.Max(0, barWidth-textWidth-leftBarWidth)

	leftBar := barStyle.Render(strings.Repeat("█", leftBarWidth))
	rightBar := barStyle.Render(strings.Repeat("█", rightBarWidth))
	return leftBar + text + rightBar
}

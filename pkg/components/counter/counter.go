package counter

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/ez-monitor/pkg/renderutils"
)

type Model struct {
	statName string
	unit     string

	currentValue float64
	width        int
	height       int

	dataCollectionErr error

	styles Styles
}

func New(statName, unit string) Model {
	return Model{
		statName: statName,
		unit:     unit,
		styles:   defaultStyles(),
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
	totalHeightOfBar := m.height - 3 // 2 for border and 1 for label

	if m.dataCollectionErr != nil {
		errText := lipgloss.NewStyle().Width(m.width - 2).AlignHorizontal(lipgloss.Center).Render(m.dataCollectionErr.Error())
		errText = lipgloss.NewStyle().PaddingBottom(totalHeightOfBar - lipgloss.Height(errText)).Render(errText) // Add padding so the label is at the bottom

		text := lipgloss.JoinVertical(lipgloss.Center, errText, statNameView)
		return m.styles.Counter.Width(m.width - 2).Height(totalHeightOfBar).Render(text)
	}

	currentValue := lipgloss.NewStyle().
		Width(m.width - 2).
		AlignHorizontal(lipgloss.Center).
		Background(lipgloss.Color("38")).
		Render(fmt.Sprintf("%.2f%s", m.currentValue, m.unit))
	currentValueWithPadding := lipgloss.NewStyle().
		PaddingBottom(totalHeightOfBar - lipgloss.Height(currentValue) + 1).
		Render(currentValue)

	counter := lipgloss.JoinVertical(lipgloss.Top, currentValueWithPadding, statNameView)
	return m.styles.Counter.Render(counter)
}

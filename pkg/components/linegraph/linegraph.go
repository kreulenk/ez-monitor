package linegraph

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/ez-monitor/pkg/renderutils"
)

type Model struct {
	statName string
	unit     string

	minValue float64
	maxValue float64
	allStats []float64
	width    int
	height   int

	dataCollectionErr error
}

func New(statName, unit string, minValue, maxValue float64) Model {
	return Model{
		statName: statName,
		unit:     unit,

		minValue: minValue,
		maxValue: maxValue,
	}
}

// Init initialises the baseModel on program load. It partly implements the tea.Model interface.
func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetAllStats(v []float64) {
	m.allStats = v
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
	return fmt.Sprintf("%v", len(m.allStats))
}

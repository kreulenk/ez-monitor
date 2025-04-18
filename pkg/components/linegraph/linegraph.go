package linegraph

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/ez-monitor/pkg/renderutils"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
)

type Model struct {
	statName string
	unit     string

	minValue float64
	maxValue float64
	allStats []statistics.HistoricalDataPoint
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

func (m *Model) SetAllStats(v []statistics.HistoricalDataPoint) {
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
	if m.dataCollectionErr != nil {
		return fmt.Sprintf("Error: %v", m.dataCollectionErr)
	}

	// Normalize data points to fit within the graph's height
	graph := make([][]rune, m.height)
	for i := range graph {
		graph[i] = make([]rune, len(m.allStats))
		for j := range graph[i] {
			graph[i][j] = ' ' // Initialize with empty space
		}
	}

	for i, point := range m.allStats {
		normalizedValue := int((point.Data - m.minValue) / (m.maxValue - m.minValue) * float64(m.height-1))
		normalizedValue = renderutils.Max(0, renderutils.Min(normalizedValue, m.height-1))
		graph[m.height-1-normalizedValue][i] = '█' // Plot the point
	}

	// Build the graph as a string
	var result string
	for _, row := range graph {
		result += string(row) + "\n"
	}

	return result
}

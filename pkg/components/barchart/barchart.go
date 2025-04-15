package barchart

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	minValue     float64
	maxValue     float64
	currentValue float64
}

func New(minValue, maxValue float64) Model {
	return Model{
		minValue: minValue,
		maxValue: maxValue,
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

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	return fmt.Sprintf("Max Val: %f\nMin Val: %f", m.maxValue, m.minValue)
}

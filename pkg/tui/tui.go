package tui

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/ez-monitor/pkg/components/barchart"
	"github.com/kreulenk/ez-monitor/pkg/inventory"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
	"github.com/muesli/termenv"
	"os"
)

// Model implements tea.Model, and manages the browser UI.
type Model struct {
	ctx  context.Context
	Help help.Model

	memBarChart  barchart.Model
	cpuBarChart  barchart.Model
	diskBarChart barchart.Model

	statsChan chan statistics.HostStats

	inventoryNameToIndexMap map[string]int // Mapping of the name of the host to the index in which it will be displayed
	inventoryIndexToNameMap map[int]string
	currentIndex            int
	statsCollector          map[string]statistics.HostStats // Mapping of hosts to all of their last collected stats
}

func Initialize(ctx context.Context, inventoryInfo []inventory.Host, statsChan chan statistics.HostStats) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	p := tea.NewProgram(initialModel(ctx, inventoryInfo, statsChan))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel(ctx context.Context, inventoryInfo []inventory.Host, statsChan chan statistics.HostStats) tea.Model {
	var inventoryNameToIndexMap = make(map[string]int)
	var inventoryIndexToNameMap = make(map[int]string)
	for i, host := range inventoryInfo {
		inventoryNameToIndexMap[host.Name] = i
		inventoryIndexToNameMap[i] = host.Name
	}

	return Model{
		ctx:  ctx,
		Help: help.New(),

		memBarChart:  barchart.New("memory", "MB", 0, 0), // 0 max value as we do not yet know the max
		cpuBarChart:  barchart.New("cpu", "%", 0, 100),
		diskBarChart: barchart.New("disk", "MB", 0, 0), // 0 max value as we do not yet know the max

		statsChan: statsChan,

		inventoryNameToIndexMap: inventoryNameToIndexMap,
		inventoryIndexToNameMap: inventoryIndexToNameMap,
		currentIndex:            0,
		statsCollector:          make(map[string]statistics.HostStats),
	}
}

func (m Model) Init() tea.Cmd {
	// Start listening to the statsChan
	return listenForStats(m.ctx, m.statsChan)
}

// listenForStats listens for messages from the statsChan and sends them as tea.Msg.
func listenForStats(ctx context.Context, statsChan chan statistics.HostStats) tea.Cmd {
	return func() tea.Msg {
		// Continuously listen for messages from the channel
		select {
		case stats := <-statsChan:
			return statsMsg(stats)
		case <-ctx.Done():
			return tea.Quit()
		}
	}
}

// statsMsg wraps the statistics.HostStats to implement tea.Msg.
type statsMsg statistics.HostStats

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Next):
			if m.currentIndex < len(m.inventoryNameToIndexMap)-1 {
				m.currentIndex++
				m.updateChildModelsWithLatestStats(m.statsCollector[m.inventoryIndexToNameMap[m.currentIndex]])
			}
		case key.Matches(msg, keys.Previous):
			if m.currentIndex > 0 {
				m.currentIndex--
				m.updateChildModelsWithLatestStats(m.statsCollector[m.inventoryIndexToNameMap[m.currentIndex]])
			}
		}
	case statsMsg:
		m.statsCollector[msg.NameOfHost] = statistics.HostStats(msg)

		// If the latest update came from the host we are on, update the charts with this data
		if m.currentIndex == m.inventoryNameToIndexMap[msg.NameOfHost] {
			m.updateChildModelsWithLatestStats(statistics.HostStats(msg))
		}

		return m, listenForStats(m.ctx, m.statsChan)

	case tea.WindowSizeMsg:
		m.memBarChart.SetWidth(msg.Width/3 - 2)
		m.memBarChart.SetHeight(msg.Height - 2)

		m.cpuBarChart.SetWidth(msg.Width/3 - 2)
		m.cpuBarChart.SetHeight(msg.Height - 2)

		m.diskBarChart.SetWidth(msg.Width/3 - 2)
		m.diskBarChart.SetHeight(msg.Height - 2)
	}

	return m, nil
}

func (m Model) View() string {
	currentHost := m.inventoryIndexToNameMap[m.currentIndex]
	if _, ok := m.statsCollector[currentHost]; ok {
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinVertical(lipgloss.Center, currentHost,
				lipgloss.JoinHorizontal(lipgloss.Left, m.memBarChart.View(), m.cpuBarChart.View(), m.diskBarChart.View())),
			m.HelpView(),
		)
	} else {
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinVertical(lipgloss.Center, currentHost, "Waiting for stats to be available...",
				m.HelpView(),
			),
		)
	}
}

func (m *Model) updateChildModelsWithLatestStats(stats statistics.HostStats) {
	if stats.MemoryError == nil {
		m.memBarChart.SetCurrentValue(stats.MemoryUsage)
		m.memBarChart.SetMaxValue(stats.MemoryTotal)
	} else {
		m.memBarChart.SetDataCollectionErr(stats.MemoryError)
	}

	if stats.DiskError == nil {
		m.diskBarChart.SetCurrentValue(stats.DiskUsage)
		m.diskBarChart.SetMaxValue(stats.DiskTotal)
	} else {
		m.diskBarChart.SetDataCollectionErr(stats.DiskError)
	}
	if stats.CPUError == nil {
		m.cpuBarChart.SetCurrentValue(stats.CPUUsage)
	} else {
		m.cpuBarChart.SetDataCollectionErr(stats.CPUError)
	}
}

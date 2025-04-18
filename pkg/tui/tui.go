package tui

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/ez-monitor/pkg/components/barchart"
	"github.com/kreulenk/ez-monitor/pkg/components/counter"
	"github.com/kreulenk/ez-monitor/pkg/components/linegraph"
	"github.com/kreulenk/ez-monitor/pkg/inventory"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
	"github.com/muesli/termenv"
	"os"
)

type ActiveView int

const (
	LiveData ActiveView = iota
	HistoricalData
)

// Model implements tea.Model, and manages the browser UI.
type Model struct {
	ctx  context.Context
	Help help.Model

	height int

	activeView ActiveView

	// Live data
	memBarChart             barchart.Model
	cpuBarChart             barchart.Model
	diskBarChart            barchart.Model
	networkingSentChart     counter.Model
	networkingReceivedChart counter.Model

	// Historical data
	cpuLineGraph linegraph.Model

	statsChan chan *statistics.HostStat

	inventoryNameToIndexMap map[string]int // Mapping of the name of the host to the index in which it will be displayed
	inventoryIndexToNameMap map[int]string
	currentIndex            int
	statsCollector          map[string][]*statistics.HostStat // Mapping of hosts to all of their last collected stats
}

func Initialize(ctx context.Context, inventoryInfo []inventory.Host, statsChan chan *statistics.HostStat) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	p := tea.NewProgram(initialModel(ctx, inventoryInfo, statsChan))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel(ctx context.Context, inventoryInfo []inventory.Host, statsChan chan *statistics.HostStat) tea.Model {
	var inventoryNameToIndexMap = make(map[string]int)
	var inventoryIndexToNameMap = make(map[int]string)
	for i, host := range inventoryInfo {
		inventoryNameToIndexMap[host.Name] = i
		inventoryIndexToNameMap[i] = host.Name
	}

	return Model{
		ctx:  ctx,
		Help: help.New(),

		activeView: LiveData,

		// Live data charts
		memBarChart:             barchart.New("memory", "MB", 0, 0), // 0 max value as we do not yet know the max
		cpuBarChart:             barchart.New("cpu", "%", 0, 100),
		diskBarChart:            barchart.New("disk", "MB", 0, 0), // 0 max value as we do not yet know the max
		networkingSentChart:     counter.New("Net Sent", "MB"),
		networkingReceivedChart: counter.New("Net Recv", "MB"),

		// Historical data charts
		cpuLineGraph: linegraph.New("cpu", "%", 0, 100),

		statsChan: statsChan,

		inventoryNameToIndexMap: inventoryNameToIndexMap,
		inventoryIndexToNameMap: inventoryIndexToNameMap,
		currentIndex:            0,
		statsCollector:          make(map[string][]*statistics.HostStat),
	}
}

func (m Model) Init() tea.Cmd {
	// Start listening to the statsChan
	return listenForStats(m.ctx, m.statsChan)
}

// listenForStats listens for messages from the statsChan and sends them as tea.Msg.
func listenForStats(ctx context.Context, statsChan chan *statistics.HostStat) tea.Cmd {
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

// statsMsg wraps the statistics.HostStat to implement tea.Msg.
type statsMsg *statistics.HostStat

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Next):
			if m.currentIndex < len(m.inventoryNameToIndexMap)-1 {
				m.currentIndex++
				m.updateChildModelsWithLatestStats(m.getLastDataPoint())
			}
		case key.Matches(msg, keys.Previous):
			if m.currentIndex > 0 {
				m.currentIndex--
				m.updateChildModelsWithLatestStats(m.getLastDataPoint())
			}
		case key.Matches(msg, keys.HistoricalView):
			m.activeView = HistoricalData
		case key.Matches(msg, keys.LiveView):
			m.activeView = LiveData
		}
	case statsMsg:
		// Append the statistic to the statsCollector for each host
		if _, ok := m.statsCollector[msg.NameOfHost]; ok {
			m.statsCollector[msg.NameOfHost] = append(m.statsCollector[msg.NameOfHost], msg)
		} else {
			m.statsCollector[msg.NameOfHost] = []*statistics.HostStat{msg}
		}

		// If the latest update came from the host we are on, update the charts with this data
		if m.currentIndex == m.inventoryNameToIndexMap[msg.NameOfHost] {
			m.updateChildModelsWithLatestStats(msg)
		}

		return m, listenForStats(m.ctx, m.statsChan)

	case tea.WindowSizeMsg:
		m.height = msg.Height

		m.memBarChart.SetWidth(msg.Width/4 - 2)
		m.memBarChart.SetHeight(msg.Height - 2)

		m.cpuBarChart.SetWidth(msg.Width/4 - 2)
		m.cpuBarChart.SetHeight(msg.Height - 2)

		m.diskBarChart.SetWidth(msg.Width/4 - 2)
		m.diskBarChart.SetHeight(msg.Height - 2)

		m.networkingSentChart.SetWidth(msg.Width/4 - 2)
		m.networkingSentChart.SetHeight(msg.Height/2 - 2)

		m.networkingReceivedChart.SetWidth(msg.Width/4 - 2)
		m.networkingReceivedChart.SetHeight(msg.Height/2 - 2)
	}

	return m, nil
}

func (m Model) getLastDataPoint() *statistics.HostStat {
	currentHostStats := m.statsCollector[m.inventoryIndexToNameMap[m.currentIndex]]
	if len(currentHostStats) > 0 {
		return currentHostStats[len(currentHostStats)-1]
	}
	return nil
}

func (m Model) View() string {
	currentHost := m.inventoryIndexToNameMap[m.currentIndex]
	if _, ok := m.statsCollector[currentHost]; ok {
		if m.activeView == LiveData {
			return m.renderLiveDataView(currentHost)
		} else {
			return m.renderHistoricalDataView(currentHost)
		}
	} else {
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinVertical(lipgloss.Center, currentHost, "Waiting for stats to be available...",
				m.HelpView(),
			),
		)
	}
}

func (m Model) renderLiveDataView(currentHost string) string {
	networkingCounters := joinVerticalStackedElementsWithBuffers(m.networkingSentChart.View(), m.networkingReceivedChart.View(), m.height)

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Center, currentHost,
			lipgloss.JoinHorizontal(lipgloss.Left, m.memBarChart.View(), m.cpuBarChart.View(), m.diskBarChart.View(), networkingCounters)),
		m.HelpView(),
	)
}

func (m Model) renderHistoricalDataView(currentHost string) string {
	return lipgloss.JoinVertical(lipgloss.Top, m.cpuLineGraph.View(), m.HelpView())
}

func (m *Model) updateChildModelsWithLatestStats(stats *statistics.HostStat) {
	if stats == nil {
		return
	}

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

	if stats.NetworkingError == nil {
		m.networkingSentChart.SetCurrentValue(stats.NetworkingMBSent)
		m.networkingReceivedChart.SetCurrentValue(stats.NetworkingMBReceived)
	} else {
		m.networkingSentChart.SetDataCollectionErr(stats.NetworkingError)
		m.networkingReceivedChart.SetDataCollectionErr(stats.NetworkingError)
	}
}

// joinVerticalStackedElementsWithBuffers will ensure that vertically stacked elements have the proper
// amount of buffer between them so that they are always the same height as other display elements
// TODO fix how height is calculated throughout the app as this algorithm is questionable at best...
func joinVerticalStackedElementsWithBuffers(element1, element2 string, height int) string {
	if height%2 != 0 {
		element2 = lipgloss.NewStyle().MarginTop(1).Render(element2)
	}
	return lipgloss.JoinVertical(lipgloss.Top, element1, element2)
}

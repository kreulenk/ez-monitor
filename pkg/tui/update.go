package tui

import (
	"context"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
)

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
				m.updateActiveCharts()
			}
		case key.Matches(msg, keys.Previous):
			if m.currentIndex > 0 {
				m.currentIndex--
				m.updateActiveCharts()
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
			m.updateActiveCharts()
		}

		return m, listenForStats(m.ctx, m.statsChan)
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		// TODO we should probably use an interface to set these values at this point..
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

		m.cpuLineGraph.SetWidth(msg.Width - 2)
		m.cpuLineGraph.SetHeight(msg.Height - 5)
	}

	return m, nil
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

func (m *Model) updateActiveCharts() {
	lastStat := m.getLastDataPoint()
	if lastStat == nil {
		return
	}
	m.updateLiveChildModelStats(lastStat)
	m.updateHistoricalChildModelStats(lastStat)
}

func (m *Model) updateLiveChildModelStats(stats *statistics.HostStat) {
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

func (m *Model) updateHistoricalChildModelStats(stats *statistics.HostStat) {
	if stats.CPUError == nil {
		m.cpuLineGraph.SetAllStats(m.getAllCPUDataPoints())
	} else {
		m.cpuLineGraph.SetDataCollectionErr(stats.CPUError)
	}
}

// TODO investigate caching this data or restructuring how we store the data
func (m Model) getAllCPUDataPoints() []statistics.HistoricalDataPoint {
	currentHostStats := m.statsCollector[m.inventoryIndexToNameMap[m.currentIndex]]
	cpuStats := make([]statistics.HistoricalDataPoint, 0, len(currentHostStats))
	if len(currentHostStats) > 0 {
		for _, hostStat := range currentHostStats {
			cpuStats = append(cpuStats, statistics.HistoricalDataPoint{Data: hostStat.CPUUsage, Timestamp: hostStat.Timestamp})
		}
		return cpuStats
	}
	return nil
}

func (m Model) getLastDataPoint() *statistics.HostStat {
	currentHostStats := m.statsCollector[m.inventoryIndexToNameMap[m.currentIndex]]
	if len(currentHostStats) > 0 {
		return currentHostStats[len(currentHostStats)-1]
	}
	return nil
}

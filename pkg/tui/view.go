package tui

import "github.com/charmbracelet/lipgloss"

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
	return lipgloss.JoinVertical(lipgloss.Center, currentHost, lipgloss.JoinVertical(lipgloss.Top, m.cpuLineGraph.View(), m.HelpView()))
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

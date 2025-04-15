package tui

import (
	"context"
	"ez-monitor/pkg/components/barchart"
	"ez-monitor/pkg/statistics"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"os"
)

// baseModel implements tea.Model, and manages the browser UI.
type baseModel struct {
	ctx         context.Context
	memBarChart barchart.Model

	statsChan chan statistics.HostStats
}

func Initialize(ctx context.Context, statsChan chan statistics.HostStats) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	p := tea.NewProgram(initialModel(ctx, statsChan))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel(ctx context.Context, statsChan chan statistics.HostStats) tea.Model {
	return baseModel{
		ctx: ctx,

		memBarChart: barchart.New(0, 0), // 0 max value as we do not yet know the max

		statsChan: statsChan,
	}
}

func (m baseModel) Init() tea.Cmd {
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

func (m baseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case statsMsg:
		// Handle the stats message
		m.memBarChart.SetCurrentValue(msg.MemoryUsage)
		m.memBarChart.SetMaxValue(msg.MemoryTotal)
		// Continue listening for more messages
		return m, listenForStats(m.ctx, m.statsChan)
	}

	return m, nil
}

func (m baseModel) View() string {
	return m.memBarChart.View()
}

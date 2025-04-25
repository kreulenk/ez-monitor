package tui

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/ez-monitor/pkg/components/barchart"
	"github.com/kreulenk/ez-monitor/pkg/components/counter"
	"github.com/kreulenk/ez-monitor/pkg/components/linegraph"
	"github.com/kreulenk/ez-monitor/pkg/inventory"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
	"github.com/kreulenk/ez-monitor/pkg/unit"
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
	width  int

	activeView ActiveView

	// Live data
	memBarChart             barchart.Model
	cpuBarChart             barchart.Model
	diskBarChart            barchart.Model
	networkingSentChart     counter.Model
	networkingReceivedChart counter.Model

	// Historical data
	memLineGraph  linegraph.Model
	cpuLineGraph  linegraph.Model
	diskLineGraph linegraph.Model

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
	var hostAliasToIndexMap = make(map[string]int)
	var hostIndexToAliasMap = make(map[int]string)
	for i, host := range inventoryInfo {
		hostAliasToIndexMap[host.Alias] = i
		hostIndexToAliasMap[i] = host.Alias
	}

	return Model{
		ctx:  ctx,
		Help: help.New(),

		activeView: LiveData,

		// Live data charts
		memBarChart:             barchart.New("memory", unit.Megabyte, 0, 0), // 0 max value as we do not yet know the max
		cpuBarChart:             barchart.New("cpu", unit.Percentage, 0, 100),
		diskBarChart:            barchart.New("disk", unit.Megabyte, 0, 0),
		networkingSentChart:     counter.New("Net Sent", unit.Megabyte),
		networkingReceivedChart: counter.New("Net Recv", unit.Megabyte),

		// Historical data charts
		memLineGraph:  linegraph.New("memory", unit.Megabyte, 0, 0),
		cpuLineGraph:  linegraph.New("cpu", unit.Percentage, 0, 100),
		diskLineGraph: linegraph.New("disk", unit.Megabyte, 0, 0),

		statsChan: statsChan,

		inventoryNameToIndexMap: hostAliasToIndexMap,
		inventoryIndexToNameMap: hostIndexToAliasMap,
		currentIndex:            0,
		statsCollector:          make(map[string][]*statistics.HostStat),
	}
}

func (m Model) Init() tea.Cmd {
	// Start listening to the statsChan
	return listenForStats(m.ctx, m.statsChan)
}

package linegraph

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/ez-monitor/pkg/renderutils"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
	"github.com/kreulenk/ez-monitor/pkg/unit"
	"log/slog"
	"math"
	"time"
)

type Model struct {
	statName string
	unit     unit.DataType

	minValue float64
	maxValue float64
	allStats []statistics.HistoricalDataPoint
	width    int
	height   int

	dataCollectionErr error

	styles Styles
}

func New(statName string, unit unit.DataType, minValue, maxValue float64) Model {
	return Model{
		statName: statName,
		unit:     unit,

		minValue: minValue,
		maxValue: maxValue,

		styles: defaultStyles(),
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
	if len(m.allStats) == 0 || m.height < 2 {
		return fmt.Sprintf("")
	}
	labelAdjustedHeight := renderutils.Max(1, m.height-2) // Take top and bottom spaces for statName and min/max values

	smallestTimestamp := m.allStats[0].Timestamp
	largestTimestamp := m.allStats[len(m.allStats)-1].Timestamp
	numBuckets := renderutils.Max(m.width, renderutils.Max(1, int(largestTimestamp.Unix()-smallestTimestamp.Unix())/m.width))
	var buckets []float64
	numBucketsWithActualData := renderutils.Min(numBuckets, len(m.allStats))
	durationPerBucket := largestTimestamp.Sub(smallestTimestamp) / time.Duration(numBucketsWithActualData)

	timesIterated := 0
	for allStatsIndex, bucketIndex := 0, 0; allStatsIndex < len(m.allStats); bucketIndex++ {
		timesIterated++
		if allStatsIndex >= len(m.allStats) {
			break
		}

		var maxNum float64
		maxTimestampInBucket := smallestTimestamp.Add(durationPerBucket * time.Duration(bucketIndex+1))

		for ; allStatsIndex < len(m.allStats); allStatsIndex++ {
			if m.statName == "memory" {
				slog.Info(fmt.Sprintf("placing stat %d with value %f in bucket %d", allStatsIndex, m.allStats[allStatsIndex].Data, bucketIndex))
			}
			maxNum = math.Max(maxNum, m.allStats[allStatsIndex].Data)
			if numBuckets > numBucketsWithActualData { // Only place data in a bucket once if we have more buckets than data
				allStatsIndex++
				break
			}
			if allStatsIndex+1 < len(m.allStats) && maxTimestampInBucket.Sub(m.allStats[allStatsIndex+1].Timestamp) < 0 {
				break
			}
		}
		normalizedValue := (maxNum - m.minValue) / (m.maxValue - m.minValue) * float64(labelAdjustedHeight-1)
		normalizedValue = math.Max(0, math.Min(normalizedValue, float64(labelAdjustedHeight-1)))
		buckets = append(buckets, normalizedValue)
	}

	// Create the canvas where other items will overwrite their data
	graph := make([][]rune, labelAdjustedHeight)
	for i := range graph {
		graph[i] = make([]rune, numBuckets)
		for j := range graph[i] {
			graph[i][j] = ' ' // Initialize with empty space
		}
	}

	// Add the actual bars onto the chart
	for bucketIndex, point := range buckets {
		if bucketIndex >= numBucketsWithActualData {
			break
		}
		for heightIndex := labelAdjustedHeight - int(point) - 1; heightIndex < labelAdjustedHeight; heightIndex++ {
			graph[heightIndex][bucketIndex] = '█' // Plot the point
		}
	}

	// Build the graph as a string
	var result string
	for i, row := range graph {
		if i == len(graph)-1 {
			result += string(row)
		} else {
			result += string(row) + "\n"
		}
	}

	maxValStr := unit.DisplayType(m.maxValue, m.unit)
	minValStr := unit.DisplayType(m.minValue, m.unit)
	return m.styles.Graph.Render(
		lipgloss.JoinVertical(lipgloss.Right,
			maxValStr,
			m.styles.Bars.Render(result),
			lipgloss.JoinHorizontal(lipgloss.Right,
				lipgloss.NewStyle().PaddingRight(len(graph[0])/2-len(minValStr)/2).Render(m.statName),
				minValStr,
			),
		),
	)
}

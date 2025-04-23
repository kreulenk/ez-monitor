package linegraph

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/ez-monitor/pkg/renderutils"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
	"math"
	"time"
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

	styles Styles
}

func New(statName, unit string, minValue, maxValue float64) Model {
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
	statsAdjustedHeight := renderutils.Max(1, m.height-2) // Used for normalization calculations to allow units to have their own row when displaying data

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
		maxTimestampInBucket := smallestTimestamp.Add(durationPerBucket * time.Duration(bucketIndex))
		dataPointsInBucket := 0
		for ; allStatsIndex < len(m.allStats) && maxTimestampInBucket.Sub(m.allStats[allStatsIndex].Timestamp) >= 0; allStatsIndex++ {
			maxNum = math.Max(maxNum, m.allStats[allStatsIndex].Data)
			dataPointsInBucket++
		}
		normalizedValue := (maxNum - m.minValue) / (m.maxValue - m.minValue) * float64(statsAdjustedHeight-1)
		normalizedValue = math.Max(0, math.Min(normalizedValue, float64(statsAdjustedHeight-1)))
		buckets = append(buckets, normalizedValue)
	}

	// Create the canvas where other items will overwrite their data
	graph := make([][]rune, m.height)
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
		for heightIndex := statsAdjustedHeight - int(point); heightIndex <= statsAdjustedHeight; heightIndex++ {
			graph[heightIndex][bucketIndex] = '█' // Plot the point
		}
	}

	// Place the max value along the top axis
	maxValStr := fmt.Sprintf("%.1f%s", m.maxValue, m.unit)
	if len(graph[0]) >= len(maxValStr) {
		for i, j := len(graph[0])-len(maxValStr), 0; i < len(graph[0]); i, j = i+1, j+1 {
			graph[0][i] = rune(maxValStr[j])
		}
	}

	// Place the min value along the top axis
	minValStr := fmt.Sprintf("%.1f%s", m.minValue, m.unit)
	if len(graph[len(graph)-1]) >= len(minValStr) {
		for i, j := len(graph[len(graph)-1])-len(minValStr), 0; i < len(graph[len(graph)-1]); i, j = i+1, j+1 {
			graph[len(graph)-1][i] = rune(minValStr[j])
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

	return m.styles.Graph.Render(result)
}

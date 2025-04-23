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
	if len(m.allStats) == 0 || m.height < 1 {
		return fmt.Sprintf("")
	}

	smallestTimestamp := m.allStats[0].Timestamp
	largestTimestamp := m.allStats[len(m.allStats)-1].Timestamp
	numBuckets := renderutils.Max(m.width, renderutils.Max(1, int(largestTimestamp.Unix()-smallestTimestamp.Unix())/m.width))
	var buckets []float64
	numBucketsWithActualData := renderutils.Min(numBuckets, len(m.allStats))
	durationPerBucket := largestTimestamp.Sub(smallestTimestamp) / time.Duration(numBucketsWithActualData)

	timesIterated := 0
	for allStatsIndex, bucketIndex := 0, 0; allStatsIndex < len(m.allStats); bucketIndex++ {
		timesIterated++
		var sum float64
		if allStatsIndex >= len(m.allStats) {
			break
		}

		maxTimestampInBucket := smallestTimestamp.Add(durationPerBucket * time.Duration(bucketIndex))
		dataPointsInBucket := 0
		for ; allStatsIndex < len(m.allStats) && maxTimestampInBucket.Sub(m.allStats[allStatsIndex].Timestamp) >= 0; allStatsIndex++ {
			sum += m.allStats[allStatsIndex].Data
			dataPointsInBucket++
		}
		avg := sum / float64(dataPointsInBucket)
		normalizedValue := (avg - m.minValue) / (m.maxValue - m.minValue) * float64(m.height-1)
		normalizedValue = math.Max(0, math.Min(normalizedValue, float64(m.height-1)))
		buckets = append(buckets, normalizedValue)
	}

	// Normalize data points to fit within the graph's height
	graph := make([][]rune, m.height)
	for i := range graph {
		graph[i] = make([]rune, numBuckets)
		for j := range graph[i] {
			graph[i][j] = ' ' // Initialize with empty space
		}
	}

	for bucketIndex, point := range buckets {
		if bucketIndex >= numBucketsWithActualData {
			break
		}

		for heightIndex := m.height - int(point) - 1; heightIndex < m.height; heightIndex++ {
			graph[heightIndex][bucketIndex] = '█' // Plot the point
		}

	}

	// Build the graph as a string
	var result string
	for _, row := range graph {
		result += string(row) + "\n"
	}

	return m.styles.Graph.Render(result)
}

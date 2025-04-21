package linegraph

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/ez-monitor/pkg/renderutils"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
	"log/slog"
	"math"
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
}

func New(statName, unit string, minValue, maxValue float64) Model {
	return Model{
		statName: statName,
		unit:     unit,

		minValue: minValue,
		maxValue: maxValue,
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
	if len(m.allStats) == 0 {
		return fmt.Sprintf("")
	}

	// 1. Find smallest and largest timestamp = m.allStats[0].Timestamp, m.allStats[len(allStats)-1].Timestamp
	// 2. Find width of table = m.width
	// 3. Find the number of time buckets = (largest_timestamp - smallest_timestamp) / m.width
	// 4. Find the number of data points in each time bucket = Math.Floor(len(m.allStats)/num_buckets)
	// 5. Iterate over all the data and place it into a slice of the length of dataPointsPerBucket where you take an average per number data points per bucket
	// Also normalize the value per the height = (avg_of_values_in_buckets - m.minValue) / (m.mavValue - m.minValue) * (m.height - 1)
	// Graph the new slice
	slog.Info(fmt.Sprintf("all the stats collected%v", m.allStats))

	smallestTimestamp := m.allStats[0].Timestamp
	largestTimestamp := m.allStats[len(m.allStats)-1].Timestamp
	numBuckets := renderutils.Max(m.width, renderutils.Max(1, int(largestTimestamp.Unix()-smallestTimestamp.Unix())/m.width))
	millisecondsPerBucket := int(largestTimestamp.Sub(smallestTimestamp).Milliseconds()) / numBuckets
	buckets := make([]float64, numBuckets)

	numBucketsWithActualData := renderutils.Min(len(buckets), len(m.allStats))
	for allStatsIndex, bucketIndex := 0, 0; allStatsIndex < numBucketsWithActualData; bucketIndex++ {
		var sum float64
		dataPointsInBucket := 0
		for j := 0; j <= millisecondsPerBucket*bucketIndex && allStatsIndex < len(m.allStats); j++ {
			sum += m.allStats[allStatsIndex].Data
			allStatsIndex++
			dataPointsInBucket++
		}
		avg := sum / float64(dataPointsInBucket)
		slog.Info(fmt.Sprintf("data before normalization %v", avg))
		normalizedValue := (avg - m.minValue) / (m.maxValue - m.minValue) * float64(m.height-1)
		normalizedValue = math.Max(0, math.Min(normalizedValue, float64(m.height-1)))
		slog.Info(fmt.Sprintf("data after normalization %v", avg))
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

	for i, point := range buckets {
		if i > numBucketsWithActualData {
			break
		}
		slog.Info(fmt.Sprintf("bucket %v %v", i, point))
		//slog.Info(fmt.Sprintf("Bucket %d: %v", allStatsIndex, point))
		graph[m.height-1-int(point)][i] = '█' // Plot the point
	}

	// Build the graph as a string
	var result string
	for _, row := range graph {
		result += string(row) + "\n"
	}

	return result
}

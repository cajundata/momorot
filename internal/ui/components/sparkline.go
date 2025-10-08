package components

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SparklineModel represents a sparkline chart.
type SparklineModel struct {
	data   []float64
	theme  SparklineTheme
	width  int
	height int
}

// SparklineTheme contains styling for the sparkline.
type SparklineTheme struct {
	PositiveStyle lipgloss.Style
	NegativeStyle lipgloss.Style
	NeutralStyle  lipgloss.Style
	BorderStyle   lipgloss.Style
}

// Block characters for different heights (8 levels)
var blocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// NewSparkline creates a new sparkline chart.
func NewSparkline(data []float64, width, height int) SparklineModel {
	return SparklineModel{
		data:   data,
		theme:  defaultSparklineTheme(),
		width:  width,
		height: height,
	}
}

// defaultSparklineTheme returns the default sparkline theme.
func defaultSparklineTheme() SparklineTheme {
	return SparklineTheme{
		PositiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")), // Green
		NegativeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")), // Red
		NeutralStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")), // Gray
		BorderStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1),
	}
}

// View renders the sparkline.
func (m SparklineModel) View() string {
	if len(m.data) == 0 {
		return m.theme.NeutralStyle.Render(strings.Repeat("▁", m.width))
	}

	// Determine color based on first vs last value
	var style lipgloss.Style
	if len(m.data) > 1 {
		change := m.data[len(m.data)-1] - m.data[0]
		if change > 0 {
			style = m.theme.PositiveStyle
		} else if change < 0 {
			style = m.theme.NegativeStyle
		} else {
			style = m.theme.NeutralStyle
		}
	} else {
		style = m.theme.NeutralStyle
	}

	// Build sparkline
	sparkline := m.buildSparkline()

	return style.Render(sparkline)
}

// ViewWithBorder renders the sparkline with a border.
func (m SparklineModel) ViewWithBorder() string {
	return m.theme.BorderStyle.Render(m.View())
}

// SetData updates the sparkline data.
func (m *SparklineModel) SetData(data []float64) {
	m.data = data
}

// SetWidth updates the sparkline width.
func (m *SparklineModel) SetWidth(width int) {
	m.width = width
}

// buildSparkline creates the ASCII sparkline representation.
func (m SparklineModel) buildSparkline() string {
	if len(m.data) == 0 {
		return strings.Repeat("▁", m.width)
	}

	// If we have fewer data points than width, use all data
	// If we have more, sample evenly
	sampled := m.sampleData()

	// Find min and max for scaling
	min, max := m.findMinMax(sampled)

	// Handle edge case where all values are the same
	if min == max {
		return strings.Repeat("▄", len(sampled))
	}

	// Build the sparkline
	var sb strings.Builder
	for _, value := range sampled {
		// Normalize to 0-1 range
		normalized := (value - min) / (max - min)

		// Map to block character index (0-7)
		index := int(normalized * float64(len(blocks)-1))
		if index < 0 {
			index = 0
		}
		if index >= len(blocks) {
			index = len(blocks) - 1
		}

		sb.WriteRune(blocks[index])
	}

	return sb.String()
}

// sampleData samples the data to fit the width.
func (m SparklineModel) sampleData() []float64 {
	if len(m.data) <= m.width {
		return m.data
	}

	// Sample evenly across the data
	sampled := make([]float64, m.width)
	step := float64(len(m.data)-1) / float64(m.width-1)

	for i := 0; i < m.width; i++ {
		index := int(math.Round(float64(i) * step))
		if index >= len(m.data) {
			index = len(m.data) - 1
		}
		sampled[i] = m.data[index]
	}

	return sampled
}

// findMinMax finds the minimum and maximum values in the data.
func (m SparklineModel) findMinMax(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}

	min := data[0]
	max := data[0]

	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	return min, max
}

// Stats returns summary statistics about the data.
type Stats struct {
	Min    float64
	Max    float64
	First  float64
	Last   float64
	Change float64
	PctChange float64
}

// GetStats returns statistics about the sparkline data.
func (m SparklineModel) GetStats() Stats {
	if len(m.data) == 0 {
		return Stats{}
	}

	min, max := m.findMinMax(m.data)
	first := m.data[0]
	last := m.data[len(m.data)-1]
	change := last - first
	pctChange := 0.0
	if first != 0 {
		pctChange = (change / first) * 100
	}

	return Stats{
		Min:       min,
		Max:       max,
		First:     first,
		Last:      last,
		Change:    change,
		PctChange: pctChange,
	}
}

package charts

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Chart styles
var (
	barStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))
	
	partialBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00"))
	
	lowBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))
	
	labelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
	
	valueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)
)

// BarChart creates a horizontal bar chart
func BarChart(data []ChartData, width int, title string) string {
	if len(data) == 0 {
		return "No data available"
	}

	var result strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#5d5d5d")).
		PaddingLeft(1).
		PaddingRight(1)
	
	result.WriteString(titleStyle.Render(title))
	result.WriteString("\n\n")

	// Find max value for scaling
	maxVal := float64(0)
	for _, item := range data {
		if item.Value > maxVal {
			maxVal = item.Value
		}
	}

	if maxVal == 0 {
		return result.String() + "No data to display"
	}

	barWidth := width - 30 // Leave space for labels and values

	for i, item := range data {
		// Calculate bar length
		barLength := int(math.Ceil((item.Value / maxVal) * float64(barWidth)))
		
		// Create the bar
		bar := createBar(barLength, item.Value, maxVal)
		
		// Format the line
		rank := fmt.Sprintf("%2d", i+1)
		label := truncateString(item.Label, 15)
		value := fmt.Sprintf("%.1f", item.Value)
		
		line := fmt.Sprintf("%s. %-15s %s %s",
			rank,
			labelStyle.Render(label),
			bar,
			valueStyle.Render(value))
		
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// SparklineChart creates a small inline chart
func SparklineChart(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}

	if len(values) > width {
		// Sample the values to fit the width
		step := float64(len(values)) / float64(width)
		sampled := make([]float64, width)
		for i := 0; i < width; i++ {
			idx := int(float64(i) * step)
			if idx >= len(values) {
				idx = len(values) - 1
			}
			sampled[i] = values[idx]
		}
		values = sampled
	}

	// Find min and max for scaling
	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == minVal {
		return strings.Repeat("▄", len(values))
	}

	// Create sparkline
	var result strings.Builder
	sparkChars := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	
	for _, v := range values {
		normalized := (v - minVal) / (maxVal - minVal)
		charIndex := int(normalized * float64(len(sparkChars)-1))
		if charIndex >= len(sparkChars) {
			charIndex = len(sparkChars) - 1
		}
		result.WriteString(sparkChars[charIndex])
	}

	return result.String()
}

// HistogramChart creates a vertical histogram
func HistogramChart(data []ChartData, width, height int, title string) string {
	if len(data) == 0 {
		return "No data available"
	}

	var result strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#5d5d5d")).
		PaddingLeft(1).
		PaddingRight(1)
	
	result.WriteString(titleStyle.Render(title))
	result.WriteString("\n\n")

	// Find max value for scaling
	maxVal := float64(0)
	for _, item := range data {
		if item.Value > maxVal {
			maxVal = item.Value
		}
	}

	if maxVal == 0 {
		return result.String() + "No data to display"
	}

	// Create histogram
	barWidth := width / len(data)
	if barWidth < 1 {
		barWidth = 1
	}

	// Draw from top to bottom
	for row := height; row > 0; row-- {
		threshold := (float64(row) / float64(height)) * maxVal
		
		for _, item := range data {
			if item.Value >= threshold {
				result.WriteString(strings.Repeat("█", barWidth))
			} else {
				result.WriteString(strings.Repeat(" ", barWidth))
			}
		}
		result.WriteString("\n")
	}

	// Draw labels at the bottom
	for _, item := range data {
		label := truncateString(item.Label, barWidth)
		result.WriteString(fmt.Sprintf("%-*s", barWidth, label))
	}
	result.WriteString("\n")

	return result.String()
}

// TrendChart creates a line chart showing trends over time
func TrendChart(data []TimeSeriesData, width, height int, title string) string {
	if len(data) == 0 {
		return "No data available"
	}

	var result strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#5d5d5d")).
		PaddingLeft(1).
		PaddingRight(1)
	
	result.WriteString(titleStyle.Render(title))
	result.WriteString("\n\n")

	// Find min/max values for scaling
	minVal, maxVal := data[0].Value, data[0].Value
	for _, item := range data {
		if item.Value < minVal {
			minVal = item.Value
		}
		if item.Value > maxVal {
			maxVal = item.Value
		}
	}

	if maxVal == minVal {
		maxVal = minVal + 1
	}

	// Create chart grid
	chart := make([][]rune, height)
	for i := range chart {
		chart[i] = make([]rune, width)
		for j := range chart[i] {
			chart[i][j] = ' '
		}
	}

	// Plot the line
	for i := 0; i < len(data)-1 && i < width-1; i++ {
		// Current and next points
		y1 := int((data[i].Value - minVal) / (maxVal - minVal) * float64(height-1))
		y2 := int((data[i+1].Value - minVal) / (maxVal - minVal) * float64(height-1))
		
		// Flip Y axis (0 at bottom)
		y1 = height - 1 - y1
		y2 = height - 1 - y2
		
		// Draw line segment
		if y1 == y2 {
			chart[y1][i] = '─'
		} else if y1 < y2 {
			chart[y1][i] = '╱'
		} else {
			chart[y1][i] = '╲'
		}
		
		// Mark data points
		chart[y1][i] = '●'
	}

	// Render the chart
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			if chart[row][col] == '●' {
				result.WriteString(valueStyle.Render("●"))
			} else if chart[row][col] == '─' {
				result.WriteString("─")
			} else if chart[row][col] == '╱' {
				result.WriteString("╱")
			} else if chart[row][col] == '╲' {
				result.WriteString("╲")
			} else {
				result.WriteString(" ")
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

// PieChart creates an ASCII pie chart
func PieChart(data []ChartData, title string) string {
	if len(data) == 0 {
		return "No data available"
	}

	var result strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#5d5d5d")).
		PaddingLeft(1).
		PaddingRight(1)
	
	result.WriteString(titleStyle.Render(title))
	result.WriteString("\n\n")

	// Calculate total
	total := float64(0)
	for _, item := range data {
		total += item.Value
	}

	if total == 0 {
		return result.String() + "No data to display"
	}

	// Create legend and segments
	colors := []lipgloss.Color{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7", "#DDA0DD", "#98D8C8"}
	
	for i, item := range data {
		percentage := (item.Value / total) * 100
		
		// Create visual bar
		barLength := int(percentage / 2) // Scale down for display
		if barLength < 1 && item.Value > 0 {
			barLength = 1
		}
		
		color := colors[i%len(colors)]
		segmentStyle := lipgloss.NewStyle().Foreground(color)
		
		bar := segmentStyle.Render(strings.Repeat("█", barLength))
		
		line := fmt.Sprintf("%-15s %s %.1f%% (%.1f)",
			truncateString(item.Label, 15),
			bar,
			percentage,
			item.Value)
		
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// ChartData represents a single data point
type ChartData struct {
	Label string
	Value float64
}

// TimeSeriesData represents a time-series data point
type TimeSeriesData struct {
	Time  string
	Value float64
}

// Helper functions
func createBar(length int, value, maxValue float64) string {
	if length <= 0 {
		return ""
	}

	// Determine bar color based on value percentage
	percentage := value / maxValue
	var style lipgloss.Style
	
	if percentage >= 0.7 {
		style = barStyle // Green for high values
	} else if percentage >= 0.4 {
		style = partialBarStyle // Yellow for medium values
	} else {
		style = lowBarStyle // Red for low values
	}

	return style.Render(strings.Repeat("█", length))
}

func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	if maxLength <= 3 {
		return s[:maxLength]
	}
	return s[:maxLength-3] + "..."
}
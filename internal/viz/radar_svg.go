package viz

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/xeon-zolt/codecompass/internal/metrics"
)

// RadarSVGGenerator generates SVG radar charts from metrics data
type RadarSVGGenerator struct {
	Width  int
	Height int
	CenterX int
	CenterY int
	Radius  int
}

// NewRadarSVGGenerator creates a new radar SVG generator with default dimensions
func NewRadarSVGGenerator() *RadarSVGGenerator {
	return &RadarSVGGenerator{
		Width:   400,
		Height:  400,
		CenterX: 200,
		CenterY: 200,
		Radius:  150,
	}
}

// GenerateRadarSVG creates an SVG radar chart from score data
func (g *RadarSVGGenerator) GenerateRadarSVG(scoreData *metrics.ScoreData) string {
	var buf bytes.Buffer

	// SVG header
	buf.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
`, g.Width, g.Height))
	buf.WriteString(`<style>
`)
	buf.WriteString(`.grid-line { stroke: #ddd; stroke-width: 1; fill: none; }
`)
	buf.WriteString(`.axis-line { stroke: #999; stroke-width: 2; }
`)
	buf.WriteString(`.strengths-area { fill: rgba(76, 175, 80, 0.3); stroke: #4CAF50; stroke-width: 2; }
`)
	buf.WriteString(`.weaknesses-area { fill: rgba(244, 67, 54, 0.3); stroke: #F44336; stroke-width: 2; }
`)
	buf.WriteString(`.label-text { font-family: Arial, sans-serif; font-size: 12px; text-anchor: middle; }
`)
	buf.WriteString(`.title-text { font-family: Arial, sans-serif; font-size: 16px; font-weight: bold; text-anchor: middle; }
`)
	buf.WriteString(`.score-text { font-family: Arial, sans-serif; font-size: 14px; text-anchor: middle; }
`)
	buf.WriteString(`</style>
`)

	// Background circle grids
	g.drawGrids(&buf)

	// Combine all metrics for comprehensive radar
	allMetrics := append(scoreData.Strengths, scoreData.Weaknesses...)
	if len(allMetrics) == 0 {
		// Default metrics if none provided
		allMetrics = []metrics.MetricScore{
			{Name: "Quality", Score: 75, Max: 100},
			{Name: "Coverage", Score: 60, Max: 100},
			{Name: "Performance", Score: 85, Max: 100},
			{Name: "Security", Score: 70, Max: 100},
			{Name: "Maintainability", Score: 80, Max: 100},
		}
	}

	// Draw axes and labels
	g.drawAxesAndLabels(&buf, allMetrics)

	// Draw radar areas
	if len(scoreData.Strengths) > 0 {
		g.drawRadarArea(&buf, scoreData.Strengths, "strengths-area")
	}
	if len(scoreData.Weaknesses) > 0 {
		g.drawRadarArea(&buf, scoreData.Weaknesses, "weaknesses-area")
	}

	// Draw combined area if we have both or default
	if len(scoreData.Strengths) == 0 && len(scoreData.Weaknesses) == 0 {
		g.drawRadarArea(&buf, allMetrics, "strengths-area")
	}

	// Title
	buf.WriteString(fmt.Sprintf(`<text x="%d" y="30" class="title-text">CodeCompass Radar Chart</text>
`, g.CenterX))

	// Overall score
	buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="score-text">Overall Score: %.1f</text>
`, g.CenterX, g.Height-20, scoreData.OverallScore))

	// Legend
	g.drawLegend(&buf)

	buf.WriteString(`</svg>
`)
	return buf.String()
}

// drawGrids draws concentric circles as grid lines
func (g *RadarSVGGenerator) drawGrids(buf *bytes.Buffer) {
	for i := 1; i <= 5; i++ {
		radius := g.Radius * i / 5
		buf.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" class="grid-line" />
`, g.CenterX, g.CenterY, radius))
	}
}

// drawAxesAndLabels draws radar axes and metric labels
func (g *RadarSVGGenerator) drawAxesAndLabels(buf *bytes.Buffer, metrics []metrics.MetricScore) {
	angleStep := 2 * math.Pi / float64(len(metrics))

	for i, metric := range metrics {
		angle := float64(i)*angleStep - math.Pi/2 // Start from top
		endX := g.CenterX + int(float64(g.Radius)*math.Cos(angle))
		endY := g.CenterY + int(float64(g.Radius)*math.Sin(angle))

		// Draw axis line
		buf.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="axis-line" />
`, g.CenterX, g.CenterY, endX, endY))

		// Draw label
		labelDistance := g.Radius + 25
		labelX := g.CenterX + int(float64(labelDistance)*math.Cos(angle))
		labelY := g.CenterY + int(float64(labelDistance)*math.Sin(angle))

		buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label-text">%s</text>
`, labelX, labelY+5, metric.Name))
	}
}

// drawRadarArea draws filled area for a set of metrics
func (g *RadarSVGGenerator) drawRadarArea(buf *bytes.Buffer, metricScores []metrics.MetricScore, cssClass string) {
	if len(metricScores) == 0 {
		return
	}

	angleStep := 2 * math.Pi / float64(len(metricScores))
	points := make([]string, 0, len(metricScores))

	for i, metric := range metricScores {
		angle := float64(i)*angleStep - math.Pi/2 // Start from top
		normalizedScore := metric.Score / metric.Max
		if normalizedScore > 1 {
			normalizedScore = 1
		}
		radialDistance := float64(g.Radius) * normalizedScore

		x := g.CenterX + int(radialDistance*math.Cos(angle))
		y := g.CenterY + int(radialDistance*math.Sin(angle))

		points = append(points, fmt.Sprintf("%d,%d", x, y))
	}

	// Create polygon
	buf.WriteString(fmt.Sprintf(`<polygon points="%s" class="%s" />
`, strings.Join(points, " "), cssClass))

	// Draw points
	for i, metric := range metricScores {
		angle := float64(i)*angleStep - math.Pi/2
		normalizedScore := metric.Score / metric.Max
		if normalizedScore > 1 {
			normalizedScore = 1
		}
		radialDistance := float64(g.Radius) * normalizedScore

		x := g.CenterX + int(radialDistance*math.Cos(angle))
		y := g.CenterY + int(radialDistance*math.Sin(angle))

		buf.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="4" fill="#333" />
`, x, y))
	}
}

// drawLegend draws a legend for the chart
func (g *RadarSVGGenerator) drawLegend(buf *bytes.Buffer) {
	legendX := 20
	legendY := g.Height - 60

	// Strengths legend
	buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="15" height="15" fill="rgba(76, 175, 80, 0.3)" stroke="#4CAF50" />
`, legendX, legendY))
	buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label-text" text-anchor="start">Strengths (≥75)</text>
`, legendX+20, legendY+12))

	// Weaknesses legend
	buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="15" height="15" fill="rgba(244, 67, 54, 0.3)" stroke="#F44336" />
`, legendX, legendY+20))
	buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label-text" text-anchor="start">Weaknesses (<75)</text>
`, legendX+20, legendY+32))
}

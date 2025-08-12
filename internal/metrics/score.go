package metrics

import (
	"math"
)

// ScoreData represents the scoring data for radar chart generation
type ScoreData struct {
	Strengths   []MetricScore `json:"strengths"`
	Weaknesses  []MetricScore `json:"weaknesses"`
	OverallScore float64      `json:"overall_score"`
}

// MetricScore represents a single metric with its score
type MetricScore struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
	Max   float64 `json:"max"`
}

// CalculateRadarScores calculates radar chart scores from leaderboard data
func CalculateRadarScores(data interface{}) *ScoreData {
	// Initialize score data
	scoreData := &ScoreData{
		Strengths:  make([]MetricScore, 0),
		Weaknesses: make([]MetricScore, 0),
	}

	// Example metrics calculation (to be customized based on actual data structure)
	metrics := []MetricScore{
		{Name: "Code Quality", Score: 85.0, Max: 100.0},
		{Name: "Test Coverage", Score: 72.0, Max: 100.0},
		{Name: "Documentation", Score: 68.0, Max: 100.0},
		{Name: "Performance", Score: 91.0, Max: 100.0},
		{Name: "Security", Score: 78.0, Max: 100.0},
		{Name: "Maintainability", Score: 82.0, Max: 100.0},
	}

	// Separate strengths and weaknesses based on threshold
	threshold := 75.0
	totalScore := 0.0

	for _, metric := range metrics {
		if metric.Score >= threshold {
			scoreData.Strengths = append(scoreData.Strengths, metric)
		} else {
			scoreData.Weaknesses = append(scoreData.Weaknesses, metric)
		}
		totalScore += metric.Score
	}

	// Calculate overall score
	scoreData.OverallScore = totalScore / float64(len(metrics))

	return scoreData
}

// NormalizeScore normalizes a score to a 0-1 range
func NormalizeScore(score, min, max float64) float64 {
	if max == min {
		return 0.0
	}
	return math.Max(0.0, math.Min(1.0, (score-min)/(max-min)))
}

// CalculateRadarPoints calculates radar chart points for visualization
func CalculateRadarPoints(scores []MetricScore, centerX, centerY, radius float64) [][]float64 {
	points := make([][]float64, len(scores))
	angleStep := 2 * math.Pi / float64(len(scores))

	for i, score := range scores {
		angle := float64(i) * angleStep
		normalizedScore := NormalizeScore(score.Score, 0, score.Max)
		radialDistance := radius * normalizedScore

		x := centerX + radialDistance*math.Cos(angle-math.Pi/2)
		y := centerY + radialDistance*math.Sin(angle-math.Pi/2)

		points[i] = []float64{x, y}
	}

	return points
}

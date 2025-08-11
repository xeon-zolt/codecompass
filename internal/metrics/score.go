package metrics

import (
	"math"
	"sort"
)

// RepoMetrics represents the raw metrics for a repository
type RepoMetrics struct {
	CodeCoverage     float64 // Percentage (0-100)
	LintIssuesDensity float64 // Issues per KLOC
	CodeChurn        float64 // Churn per file
	BugDensity       float64 // Bugs per KLOC
	TechnicalDebt    float64 // Lower is better
	SpellingIssues   float64 // Spell issues per KLOC
	Complexity       float64 // Cyclomatic complexity or lines per function
}

// NormalizedScores represents the normalized scores on a 0-100 scale
type NormalizedScores struct {
	CodeCoverage     float64 // Higher is better
	LintIssuesDensity float64 // Inverted: lower issues = higher score
	CodeChurn        float64 // Inverted: lower churn = higher score
	BugDensity       float64 // Inverted: lower bugs = higher score
	TechnicalDebt    float64 // Inverted: lower debt = higher score
	SpellingIssues   float64 // Inverted: lower issues = higher score
	Complexity       float64 // Inverted: lower complexity = higher score
}

// Normalizer provides methods to normalize metrics to 0-100 scale
type Normalizer struct {
	// Default clamps for normalization when no historical data is available
	MaxLintIssuesDensity float64
	MaxCodeChurn        float64
	MaxBugDensity       float64
	MaxTechnicalDebt    float64
	MaxSpellingIssues   float64
	MaxComplexity       float64
}

// NewNormalizer creates a new normalizer with reasonable defaults
func NewNormalizer() *Normalizer {
	return &Normalizer{
		MaxLintIssuesDensity: 50.0,  // 50 issues per KLOC
		MaxCodeChurn:        10.0,  // 10 churns per file
		MaxBugDensity:       20.0,  // 20 bugs per KLOC
		MaxTechnicalDebt:    100.0, // Arbitrary debt units
		MaxSpellingIssues:   10.0,  // 10 spelling issues per KLOC
		MaxComplexity:       20.0,  // 20 lines per function average
	}
}

// NormalizeMetrics converts raw metrics to normalized scores (0-100)
func (n *Normalizer) NormalizeMetrics(metrics RepoMetrics) NormalizedScores {
	return NormalizedScores{
		CodeCoverage:     n.normalizePositive(metrics.CodeCoverage, 100.0),
		LintIssuesDensity: n.normalizeInverted(metrics.LintIssuesDensity, n.MaxLintIssuesDensity),
		CodeChurn:        n.normalizeInverted(metrics.CodeChurn, n.MaxCodeChurn),
		BugDensity:       n.normalizeInverted(metrics.BugDensity, n.MaxBugDensity),
		TechnicalDebt:    n.normalizeInverted(metrics.TechnicalDebt, n.MaxTechnicalDebt),
		SpellingIssues:   n.normalizeInverted(metrics.SpellingIssues, n.MaxSpellingIssues),
		Complexity:       n.normalizeInverted(metrics.Complexity, n.MaxComplexity),
	}
}

// normalizePositive normalizes a metric where higher values are better
func (n *Normalizer) normalizePositive(value, max float64) float64 {
	if max == 0 {
		return 0
	}
	score := (value / max) * 100.0
	return math.Min(100.0, math.Max(0.0, score))
}

// normalizeInverted normalizes a metric where lower values are better
// Returns 100 - normalized(value) so lower is better becomes higher score
func (n *Normalizer) normalizeInverted(value, max float64) float64 {
	if max == 0 {
		return 100.0 // If no max, assume perfect score
	}
	normalizedValue := (value / max) * 100.0
	clampedValue := math.Min(100.0, math.Max(0.0, normalizedValue))
	return 100.0 - clampedValue
}

// NormalizeWithQuantiles normalizes metrics using quantile-based approach
// This is more robust when you have historical data
func (n *Normalizer) NormalizeWithQuantiles(metrics RepoMetrics, historicalData []RepoMetrics) NormalizedScores {
	if len(historicalData) == 0 {
		return n.NormalizeMetrics(metrics)
	}

	return NormalizedScores{
		CodeCoverage:     n.normalizeWithQuantile(metrics.CodeCoverage, extractCoverage(historicalData), false),
		LintIssuesDensity: n.normalizeWithQuantile(metrics.LintIssuesDensity, extractLintDensity(historicalData), true),
		CodeChurn:        n.normalizeWithQuantile(metrics.CodeChurn, extractChurn(historicalData), true),
		BugDensity:       n.normalizeWithQuantile(metrics.BugDensity, extractBugDensity(historicalData), true),
		TechnicalDebt:    n.normalizeWithQuantile(metrics.TechnicalDebt, extractDebt(historicalData), true),
		SpellingIssues:   n.normalizeWithQuantile(metrics.SpellingIssues, extractSpelling(historicalData), true),
		Complexity:       n.normalizeWithQuantile(metrics.Complexity, extractComplexity(historicalData), true),
	}
}

// normalizeWithQuantile normalizes using 90th percentile as reference
func (n *Normalizer) normalizeWithQuantile(value float64, data []float64, invert bool) float64 {
	if len(data) == 0 {
		return 50.0 // Default middle score
	}

	sort.Float64s(data)
	percentile90 := quantile(data, 0.9)
	percentile10 := quantile(data, 0.1)

	var score float64
	if percentile90 == percentile10 {
		score = 50.0 // All values are the same
	} else {
		// Normalize to 0-100 based on 10th-90th percentile range
		score = ((value - percentile10) / (percentile90 - percentile10)) * 100.0
		score = math.Min(100.0, math.Max(0.0, score))
	}

	if invert {
		score = 100.0 - score
	}

	return score
}

// quantile calculates the quantile of a sorted slice
func quantile(sortedData []float64, p float64) float64 {
	n := len(sortedData)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sortedData[0]
	}

	index := p * float64(n-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedData[lower]
	}

	fraction := index - float64(lower)
	return sortedData[lower]*(1-fraction) + sortedData[upper]*fraction
}

// Helper functions to extract specific metrics from historical data
func extractCoverage(data []RepoMetrics) []float64 {
	result := make([]float64, len(data))
	for i, d := range data {
		result[i] = d.CodeCoverage
	}
	return result
}

func extractLintDensity(data []RepoMetrics) []float64 {
	result := make([]float64, len(data))
	for i, d := range data {
		result[i] = d.LintIssuesDensity
	}
	return result
}

func extractChurn(data []RepoMetrics) []float64 {
	result := make([]float64, len(data))
	for i, d := range data {
		result[i] = d.CodeChurn
	}
	return result
}

func extractBugDensity(data []RepoMetrics) []float64 {
	result := make([]float64, len(data))
	for i, d := range data {
		result[i] = d.BugDensity
	}
	return result
}

func extractDebt(data []RepoMetrics) []float64 {
	result := make([]float64, len(data))
	for i, d := range data {
		result[i] = d.TechnicalDebt
	}
	return result
}

func extractSpelling(data []RepoMetrics) []float64 {
	result := make([]float64, len(data))
	for i, d := range data {
		result[i] = d.SpellingIssues
	}
	return result
}

func extractComplexity(data []RepoMetrics) []float64 {
	result := make([]float64, len(data))
	for i, d := range data {
		result[i] = d.Complexity
	}
	return result
}

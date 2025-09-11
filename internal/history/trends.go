package history

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"codecompass/internal/types"
)

// TrendData represents historical data for trend analysis
type TrendData struct {
	Timestamp      time.Time                `json:"timestamp"`
	TotalIssues    int                      `json:"total_issues"`
	CriticalIssues int                      `json:"critical_issues"`
	Authors        []types.LeaderboardEntry `json:"authors"`
	Files          []types.FileLeaderboardEntry `json:"files"`
	Rules          []types.RuleLeaderboardEntry `json:"rules"`
	Metrics        ProjectMetrics           `json:"metrics"`
}

// ProjectMetrics contains aggregated project health metrics
type ProjectMetrics struct {
	TotalFiles         int     `json:"total_files"`
	JavaScriptFiles    int     `json:"javascript_files"`
	LinesOfCode        int     `json:"lines_of_code"`
	TestCoverage       float64 `json:"test_coverage"`
	TechnicalDebt      int     `json:"technical_debt"`
	SecurityIssues     int     `json:"security_issues"`
	ComplexityIssues   int     `json:"complexity_issues"`
	ImportIssues       int     `json:"import_issues"`
	QualityScore       float64 `json:"quality_score"`
	ChangeVolatility   float64 `json:"change_volatility"`
}

// TrendAnalyzer manages historical data and trend analysis
type TrendAnalyzer struct {
	historyDir string
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer(historyDir string) *TrendAnalyzer {
	return &TrendAnalyzer{
		historyDir: historyDir,
	}
}

// SaveSnapshot saves current analysis snapshot for trend tracking
func (ta *TrendAnalyzer) SaveSnapshot(
	authors []types.LeaderboardEntry,
	files []types.FileLeaderboardEntry,
	rules []types.RuleLeaderboardEntry,
	metrics ProjectMetrics,
) error {
	// Ensure history directory exists
	if err := os.MkdirAll(ta.historyDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	snapshot := TrendData{
		Timestamp:      time.Now(),
		TotalIssues:    ta.calculateTotalIssues(authors),
		CriticalIssues: ta.calculateCriticalIssues(authors),
		Authors:        authors,
		Files:          files,
		Rules:          rules,
		Metrics:        metrics,
	}

	// Save as JSON for detailed data
	jsonFile := filepath.Join(ta.historyDir, fmt.Sprintf("snapshot_%s.json", 
		time.Now().Format("20060102_150405")))
	
	jsonData, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON snapshot: %w", err)
	}

	// Save summary as CSV for easy analysis
	if err := ta.saveSnapshotCSV(snapshot); err != nil {
		return fmt.Errorf("failed to save CSV snapshot: %w", err)
	}

	// Cleanup old snapshots (keep last 30 days)
	ta.cleanupOldSnapshots()

	fmt.Printf("📈 Snapshot saved: %s\n", jsonFile)
	return nil
}

// LoadTrendHistory loads historical snapshots for trend analysis
func (ta *TrendAnalyzer) LoadTrendHistory(days int) ([]TrendData, error) {
	files, err := filepath.Glob(filepath.Join(ta.historyDir, "snapshot_*.json"))
	if err != nil {
		return nil, err
	}

	var trends []TrendData
	cutoff := time.Now().AddDate(0, 0, -days)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip corrupted files
		}

		var trend TrendData
		if err := json.Unmarshal(data, &trend); err != nil {
			continue // Skip corrupted files
		}

		if trend.Timestamp.After(cutoff) {
			trends = append(trends, trend)
		}
	}

	// Sort by timestamp
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Timestamp.Before(trends[j].Timestamp)
	})

	return trends, nil
}

// AnalyzeTrends analyzes trends over time and generates insights
func (ta *TrendAnalyzer) AnalyzeTrends(days int) (*TrendInsights, error) {
	trends, err := ta.LoadTrendHistory(days)
	if err != nil {
		return nil, err
	}

	if len(trends) < 2 {
		return &TrendInsights{
			Message: "Insufficient historical data for trend analysis. Run analysis regularly to build trend history.",
		}, nil
	}

	insights := &TrendInsights{
		Period:     days,
		DataPoints: len(trends),
		StartDate:  trends[0].Timestamp,
		EndDate:    trends[len(trends)-1].Timestamp,
	}

	// Calculate trends
	insights.IssuesTrend = ta.calculateTrend(trends, func(t TrendData) float64 { return float64(t.TotalIssues) })
	insights.CriticalTrend = ta.calculateTrend(trends, func(t TrendData) float64 { return float64(t.CriticalIssues) })
	insights.QualityTrend = ta.calculateTrend(trends, func(t TrendData) float64 { return t.Metrics.QualityScore })
	insights.TechnicalDebtTrend = ta.calculateTrend(trends, func(t TrendData) float64 { return float64(t.Metrics.TechnicalDebt) })

	// Analyze author trends
	insights.AuthorInsights = ta.analyzeAuthorTrends(trends)
	
	// Analyze file hotspots
	insights.FileHotspots = ta.analyzeFileHotspots(trends)
	
	// Calculate velocity and predictions
	insights.Velocity = ta.calculateVelocity(trends)
	insights.Predictions = ta.generatePredictions(trends)

	// Generate recommendations
	insights.Recommendations = ta.generateTrendRecommendations(insights)

	return insights, nil
}

// TrendInsights contains trend analysis results
type TrendInsights struct {
	Period              int                    `json:"period"`
	DataPoints          int                    `json:"data_points"`
	StartDate           time.Time              `json:"start_date"`
	EndDate             time.Time              `json:"end_date"`
	Message             string                 `json:"message,omitempty"`
	IssuesTrend         TrendMetric            `json:"issues_trend"`
	CriticalTrend       TrendMetric            `json:"critical_trend"`
	QualityTrend        TrendMetric            `json:"quality_trend"`
	TechnicalDebtTrend  TrendMetric            `json:"technical_debt_trend"`
	AuthorInsights      []AuthorTrendInsight   `json:"author_insights"`
	FileHotspots        []FileHotspotInsight   `json:"file_hotspots"`
	Velocity            ProjectVelocity        `json:"velocity"`
	Predictions         TrendPredictions       `json:"predictions"`
	Recommendations     []string               `json:"recommendations"`
}

type TrendMetric struct {
	Current    float64 `json:"current"`
	Previous   float64 `json:"previous"`
	Change     float64 `json:"change"`
	ChangeRate float64 `json:"change_rate"` // Percentage
	Direction  string  `json:"direction"`   // "improving", "degrading", "stable"
}

type AuthorTrendInsight struct {
	Name       string  `json:"name"`
	Trend      string  `json:"trend"` // "improving", "degrading", "new", "inactive"
	IssueRate  float64 `json:"issue_rate"`
	Activity   string  `json:"activity"` // "increasing", "decreasing", "stable"
}

type FileHotspotInsight struct {
	FilePath      string  `json:"file_path"`
	TrendSeverity string  `json:"trend_severity"` // "getting_worse", "improving", "consistently_bad"
	IssueCount    int     `json:"issue_count"`
	ChangeRate    float64 `json:"change_rate"`
}

type ProjectVelocity struct {
	IssuesPerDay        float64 `json:"issues_per_day"`
	ResolutionRate      float64 `json:"resolution_rate"`
	TechnicalDebtRate   float64 `json:"technical_debt_rate"`
	QualityImprovement  float64 `json:"quality_improvement"`
}

type TrendPredictions struct {
	IssuesIn30Days      int     `json:"issues_in_30_days"`
	QualityIn30Days     float64 `json:"quality_in_30_days"`
	TechnicalDebtIn30Days int   `json:"technical_debt_in_30_days"`
	RecommendedActions  []string `json:"recommended_actions"`
}

// Helper methods for trend calculation
func (ta *TrendAnalyzer) calculateTrend(trends []TrendData, valueFunc func(TrendData) float64) TrendMetric {
	if len(trends) < 2 {
		return TrendMetric{Direction: "insufficient_data"}
	}

	current := valueFunc(trends[len(trends)-1])
	previous := valueFunc(trends[0])
	change := current - previous
	changeRate := 0.0
	
	if previous != 0 {
		changeRate = (change / previous) * 100
	}

	direction := "stable"
	if changeRate > 5 {
		direction = "degrading"
	} else if changeRate < -5 {
		direction = "improving"
	}

	return TrendMetric{
		Current:    current,
		Previous:   previous,
		Change:     change,
		ChangeRate: changeRate,
		Direction:  direction,
	}
}

func (ta *TrendAnalyzer) calculateTotalIssues(authors []types.LeaderboardEntry) int {
	total := 0
	for _, author := range authors {
		total += author.Count
	}
	return total
}

func (ta *TrendAnalyzer) calculateCriticalIssues(authors []types.LeaderboardEntry) int {
	total := 0
	for _, author := range authors {
		total += author.Errors // Assuming errors are critical issues
	}
	return total
}

func (ta *TrendAnalyzer) analyzeAuthorTrends(trends []TrendData) []AuthorTrendInsight {
	// This is a simplified implementation
	// In practice, you'd do more sophisticated analysis
	var insights []AuthorTrendInsight
	
	if len(trends) < 2 {
		return insights
	}

	latest := trends[len(trends)-1]
	for _, author := range latest.Authors {
		if len(insights) >= 5 { // Limit to top 5 authors
			break
		}
		
		trend := "stable"
		activity := "stable"
		
		if author.Count > 10 {
			trend = "degrading"
		} else if author.Count < 3 {
			trend = "improving"
		}

		insights = append(insights, AuthorTrendInsight{
			Name:      author.Name,
			Trend:     trend,
			IssueRate: float64(author.Count),
			Activity:  activity,
		})
	}

	return insights
}

func (ta *TrendAnalyzer) analyzeFileHotspots(trends []TrendData) []FileHotspotInsight {
	var insights []FileHotspotInsight
	
	if len(trends) < 2 {
		return insights
	}

	latest := trends[len(trends)-1]
	for _, file := range latest.Files {
		if len(insights) >= 5 { // Limit to top 5 files
			break
		}

		severity := "stable"
		if file.Count > 20 {
			severity = "consistently_bad"
		} else if file.Count > 10 {
			severity = "getting_worse"
		} else if file.Count < 5 {
			severity = "improving"
		}

		insights = append(insights, FileHotspotInsight{
			FilePath:      file.Path,
			TrendSeverity: severity,
			IssueCount:    file.Count,
			ChangeRate:    0.0, // Would calculate from historical data
		})
	}

	return insights
}

func (ta *TrendAnalyzer) calculateVelocity(trends []TrendData) ProjectVelocity {
	if len(trends) < 2 {
		return ProjectVelocity{}
	}

	// Simple velocity calculation
	timeSpan := trends[len(trends)-1].Timestamp.Sub(trends[0].Timestamp).Hours() / 24 // days
	if timeSpan == 0 {
		return ProjectVelocity{}
	}

	issueChange := float64(trends[len(trends)-1].TotalIssues - trends[0].TotalIssues)
	
	return ProjectVelocity{
		IssuesPerDay:       issueChange / timeSpan,
		ResolutionRate:     -issueChange / timeSpan, // Negative issue change means resolution
		TechnicalDebtRate:  0.0, // Would calculate from debt metrics
		QualityImprovement: 0.0, // Would calculate from quality metrics
	}
}

func (ta *TrendAnalyzer) generatePredictions(trends []TrendData) TrendPredictions {
	velocity := ta.calculateVelocity(trends)
	current := trends[len(trends)-1]

	predictions := TrendPredictions{
		IssuesIn30Days:        current.TotalIssues + int(velocity.IssuesPerDay*30),
		QualityIn30Days:       current.Metrics.QualityScore,
		TechnicalDebtIn30Days: current.Metrics.TechnicalDebt,
	}

	// Add recommended actions based on predictions
	if predictions.IssuesIn30Days > current.TotalIssues*2 {
		predictions.RecommendedActions = append(predictions.RecommendedActions,
			"Issue growth rate is concerning - implement stricter code review")
	}

	return predictions
}

func (ta *TrendAnalyzer) generateTrendRecommendations(insights *TrendInsights) []string {
	var recommendations []string

	if insights.IssuesTrend.Direction == "degrading" {
		recommendations = append(recommendations,
			"📈 Issues are increasing - consider implementing stricter linting rules")
	}

	if insights.CriticalTrend.Direction == "degrading" {
		recommendations = append(recommendations,
			"🚨 Critical issues are increasing - prioritize code quality training")
	}

	if insights.QualityTrend.Direction == "degrading" {
		recommendations = append(recommendations,
			"📉 Code quality is declining - implement automated quality gates")
	}

	if len(insights.FileHotspots) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("🔥 %d files are consistent hotspots - schedule refactoring", len(insights.FileHotspots)))
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"✅ Trends look good - maintain current quality practices")
	}

	return recommendations
}

func (ta *TrendAnalyzer) saveSnapshotCSV(snapshot TrendData) error {
	csvFile := filepath.Join(ta.historyDir, "trend_summary.csv")
	
	// Check if file exists to determine if we need header
	fileExists := true
	if _, err := os.Stat(csvFile); os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if new file
	if !fileExists {
		header := []string{
			"Timestamp", "TotalIssues", "CriticalIssues", "QualityScore",
			"TechnicalDebt", "SecurityIssues", "ComplexityIssues", "ImportIssues",
		}
		if err := writer.Write(header); err != nil {
			return err
		}
	}

	// Write data row
	row := []string{
		snapshot.Timestamp.Format("2006-01-02 15:04:05"),
		strconv.Itoa(snapshot.TotalIssues),
		strconv.Itoa(snapshot.CriticalIssues),
		fmt.Sprintf("%.2f", snapshot.Metrics.QualityScore),
		strconv.Itoa(snapshot.Metrics.TechnicalDebt),
		strconv.Itoa(snapshot.Metrics.SecurityIssues),
		strconv.Itoa(snapshot.Metrics.ComplexityIssues),
		strconv.Itoa(snapshot.Metrics.ImportIssues),
	}

	return writer.Write(row)
}

func (ta *TrendAnalyzer) cleanupOldSnapshots() {
	files, err := filepath.Glob(filepath.Join(ta.historyDir, "snapshot_*.json"))
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -30) // Keep 30 days

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(file) // Ignore errors
		}
	}
}

// PrintTrendAnalysis displays trend analysis results
func (ta *TrendAnalyzer) PrintTrendAnalysis(insights *TrendInsights) {
	if insights.Message != "" {
		fmt.Println(insights.Message)
		return
	}

	fmt.Println("📈 TREND ANALYSIS")
	fmt.Println("=" + strings.Repeat("=", 30))
	fmt.Printf("Period: %d days (%d data points)\n", insights.Period, insights.DataPoints)
	fmt.Printf("From: %s to %s\n\n", 
		insights.StartDate.Format("2006-01-02"), 
		insights.EndDate.Format("2006-01-02"))

	// Issues trend
	ta.printTrendMetric("Total Issues", insights.IssuesTrend)
	ta.printTrendMetric("Critical Issues", insights.CriticalTrend)
	ta.printTrendMetric("Quality Score", insights.QualityTrend)
	ta.printTrendMetric("Technical Debt", insights.TechnicalDebtTrend)

	// File hotspots
	if len(insights.FileHotspots) > 0 {
		fmt.Println("\n🔥 PERSISTENT FILE HOTSPOTS:")
		for i, hotspot := range insights.FileHotspots {
			fmt.Printf("  %d. %s (%d issues) - %s\n", 
				i+1, hotspot.FilePath, hotspot.IssueCount, hotspot.TrendSeverity)
		}
	}

	// Predictions
	fmt.Println("\n🔮 30-DAY PROJECTIONS:")
	fmt.Printf("  Issues: %d (current trend)\n", insights.Predictions.IssuesIn30Days)
	fmt.Printf("  Quality Score: %.1f\n", insights.Predictions.QualityIn30Days)
	fmt.Printf("  Technical Debt: %d\n", insights.Predictions.TechnicalDebtIn30Days)

	// Recommendations
	if len(insights.Recommendations) > 0 {
		fmt.Println("\n💡 RECOMMENDATIONS:")
		for i, rec := range insights.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	}

	fmt.Println()
}

func (ta *TrendAnalyzer) printTrendMetric(name string, metric TrendMetric) {
	direction := ""
	switch metric.Direction {
	case "improving":
		direction = "📈 IMPROVING"
	case "degrading":
		direction = "📉 DEGRADING"
	case "stable":
		direction = "➡️  STABLE"
	default:
		direction = "❓ UNKNOWN"
	}

	fmt.Printf("%s: %.1f (%.1f%%) %s\n", 
		name, metric.Current, metric.ChangeRate, direction)
}


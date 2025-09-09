package quality

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"codecompass/internal/charts"
	"codecompass/internal/types"

	"github.com/charmbracelet/lipgloss"
)

// QualityScorer calculates code quality scores
type QualityScorer struct {
	MaxScore float64
}

// NewQualityScorer creates a new quality scorer
func NewQualityScorer() *QualityScorer {
	return &QualityScorer{MaxScore: 100.0}
}

// QualityReport represents overall code quality metrics
type QualityReport struct {
	OverallScore    float64
	FileScores      []FileQualityScore
	AuthorScores    []AuthorQualityScore
	CategoryScores  map[string]float64
	Recommendations []string
	TrendDirection  string
}

// FileQualityScore represents quality metrics for a single file
type FileQualityScore struct {
	FilePath        string
	OverallScore    float64
	LintIssues      int
	Complexity      float64
	TechnicalDebt   int
	LinesOfCode     int
	Maintainability string
	TestCoverage    float64
}

// AuthorQualityScore represents quality metrics for an author
type AuthorQualityScore struct {
	Name            string
	Email           string
	OverallScore    float64
	IssuesPerCommit float64
	CodeSmell       int
	BestPractices   float64
	Contributions   int
}

// CalculateFileQuality calculates quality score for a file
func (qs *QualityScorer) CalculateFileQuality(
	filePath string,
	lintIssues int,
	loc int,
	technicalDebt int,
	coverage float64,
) FileQualityScore {
	
	// Base score starts at max
	score := qs.MaxScore
	
	// Deduct points for lint issues (max 40 points deduction)
	lintPenalty := math.Min(float64(lintIssues)*2, 40)
	score -= lintPenalty
	
	// Deduct points for technical debt (max 20 points deduction)
	debtPenalty := math.Min(float64(technicalDebt)*5, 20)
	score -= debtPenalty
	
	// Deduct points for poor coverage (max 25 points deduction)
	if coverage < 80 {
		coveragePenalty := (80 - coverage) / 80 * 25
		score -= coveragePenalty
	}
	
	// File size penalty (very large files get penalized)
	if loc > 1000 {
		sizePenalty := math.Min((float64(loc)-1000)/100, 15)
		score -= sizePenalty
	}
	
	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}
	
	maintainability := getMaintainabilityRating(score)
	complexity := calculateComplexity(loc, lintIssues, technicalDebt)
	
	return FileQualityScore{
		FilePath:        filePath,
		OverallScore:    score,
		LintIssues:      lintIssues,
		Complexity:      complexity,
		TechnicalDebt:   technicalDebt,
		LinesOfCode:     loc,
		Maintainability: maintainability,
		TestCoverage:    coverage,
	}
}

// CalculateAuthorQuality calculates quality score for an author
func (qs *QualityScorer) CalculateAuthorQuality(
	name, email string,
	totalIssues, totalCommits int,
	codeSmells int,
) AuthorQualityScore {
	
	score := qs.MaxScore
	
	// Calculate issues per commit
	issuesPerCommit := float64(0)
	if totalCommits > 0 {
		issuesPerCommit = float64(totalIssues) / float64(totalCommits)
	}
	
	// Deduct points for high issue rate
	issuePenalty := math.Min(issuesPerCommit*10, 40)
	score -= issuePenalty
	
	// Deduct points for code smells
	smellPenalty := math.Min(float64(codeSmells)*3, 30)
	score -= smellPenalty
	
	// Bonus for consistency (low issues per commit)
	if issuesPerCommit < 0.1 && totalCommits > 10 {
		score += 5
	}
	
	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}
	
	// Best practices score (inverse of issue rate)
	bestPractices := math.Max(0, 100-issuesPerCommit*50)
	
	return AuthorQualityScore{
		Name:            name,
		Email:           email,
		OverallScore:    score,
		IssuesPerCommit: issuesPerCommit,
		CodeSmell:       codeSmells,
		BestPractices:   bestPractices,
		Contributions:   totalCommits,
	}
}

// GenerateQualityReport creates a comprehensive quality report
func (qs *QualityScorer) GenerateQualityReport(
	fileStats map[string]*types.FileStats,
	authorStats map[string]*types.AuthorStats,
	trackedFiles map[string]bool,
) *QualityReport {
	
	var fileScores []FileQualityScore
	var authorScores []AuthorQualityScore
	categoryScores := make(map[string]float64)
	
	totalScore := float64(0)
	fileCount := 0
	
	// Calculate file quality scores
	for filePath := range trackedFiles {
		lintIssues := 0
		technicalDebt := 0
		loc := 0
		coverage := float64(75) // Default coverage assumption
		
		if stats, exists := fileStats[filePath]; exists {
			lintIssues = stats.Count
		}
		
		// Mock some additional data for demonstration
		if strings.HasSuffix(filePath, ".go") {
			// Go files typically have fewer issues
			technicalDebt = lintIssues / 3
		} else if strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".ts") {
			// JavaScript/TypeScript might have more debt
			technicalDebt = lintIssues / 2
		}
		
		// Estimate LOC (this would be better retrieved from actual file analysis)
		loc = estimateLinesOfCode(filePath)
		
		fileScore := qs.CalculateFileQuality(filePath, lintIssues, loc, technicalDebt, coverage)
		fileScores = append(fileScores, fileScore)
		
		totalScore += fileScore.OverallScore
		fileCount++
	}
	
	// Calculate author quality scores
	for email, stats := range authorStats {
		codeSmells := stats.Count / 5 // Rough estimation
		authorScore := qs.CalculateAuthorQuality(stats.Name, email, stats.Count, 50, codeSmells)
		authorScores = append(authorScores, authorScore)
	}
	
	// Sort by score
	sort.Slice(fileScores, func(i, j int) bool {
		return fileScores[i].OverallScore > fileScores[j].OverallScore
	})
	sort.Slice(authorScores, func(i, j int) bool {
		return authorScores[i].OverallScore > authorScores[j].OverallScore
	})
	
	// Calculate overall score
	overallScore := float64(0)
	if fileCount > 0 {
		overallScore = totalScore / float64(fileCount)
	}
	
	// Category scores
	categoryScores["Maintainability"] = calculateCategoryScore(fileScores, "maintainability")
	categoryScores["Code Quality"] = calculateCategoryScore(fileScores, "quality")
	categoryScores["Test Coverage"] = calculateCategoryScore(fileScores, "coverage")
	categoryScores["Technical Debt"] = calculateCategoryScore(fileScores, "debt")
	
	// Generate recommendations
	recommendations := generateRecommendations(fileScores, authorScores)
	
	return &QualityReport{
		OverallScore:    overallScore,
		FileScores:      fileScores,
		AuthorScores:    authorScores,
		CategoryScores:  categoryScores,
		Recommendations: recommendations,
		TrendDirection:  calculateTrendDirection(overallScore),
	}
}

// PrintQualityReport displays the quality report with charts
func PrintQualityReport(report *QualityReport, topN int) {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#5d5d5d")).
		PaddingLeft(1).
		PaddingRight(1)
	
	fmt.Println(style.Render("📊 Code Quality Report"))
	fmt.Println(strings.Repeat("═", 60))
	
	// Overall score with visual indicator
	scoreColor := getScoreColor(report.OverallScore)
	scoreStyle := lipgloss.NewStyle().Foreground(scoreColor).Bold(true)
	
	fmt.Printf("🎯 Overall Quality Score: %s (%.1f/100)\n", 
		scoreStyle.Render(fmt.Sprintf("%.1f", report.OverallScore)), report.OverallScore)
	fmt.Printf("📈 Trend: %s\n\n", report.TrendDirection)
	
	// Category breakdown with bar chart
	var categoryData []charts.ChartData
	for category, score := range report.CategoryScores {
		categoryData = append(categoryData, charts.ChartData{
			Label: category,
			Value: score,
		})
	}
	
	// Sort categories by score
	sort.Slice(categoryData, func(i, j int) bool {
		return categoryData[i].Value > categoryData[j].Value
	})
	
	categoryChart := charts.BarChart(categoryData, 80, "📋 Quality Categories Breakdown")
	fmt.Println(categoryChart)
	fmt.Println()
	
	// Top/Bottom files
	fmt.Println(style.Render("🏆 Top Quality Files"))
	printTopFiles(report.FileScores, topN, true)
	
	fmt.Println(style.Render("⚠️  Files Needing Attention"))
	printTopFiles(report.FileScores, topN, false)
	
	// Author quality distribution
	if len(report.AuthorScores) > 0 {
		fmt.Println(style.Render("👥 Author Quality Metrics"))
		var authorData []charts.ChartData
		for i, author := range report.AuthorScores {
			if i < topN {
				label := author.Name
				if len(label) > 15 {
					label = label[:12] + "..."
				}
				authorData = append(authorData, charts.ChartData{
					Label: label,
					Value: author.OverallScore,
				})
			}
		}
		
		authorChart := charts.BarChart(authorData, 80, "Author Quality Scores")
		fmt.Println(authorChart)
		fmt.Println()
	}
	
	// Recommendations
	if len(report.Recommendations) > 0 {
		fmt.Println(style.Render("💡 Recommendations"))
		for i, rec := range report.Recommendations {
			if i < 5 { // Show top 5 recommendations
				fmt.Printf("  %d. %s\n", i+1, rec)
			}
		}
		fmt.Println()
	}
	
	// Quality distribution pie chart
	qualityBuckets := categorizeQualityScores(report.FileScores)
	if len(qualityBuckets) > 0 {
		pieChart := charts.PieChart(qualityBuckets, "📊 Quality Distribution")
		fmt.Println(pieChart)
	}
}

// Helper functions
func getMaintainabilityRating(score float64) string {
	if score >= 90 {
		return "Excellent"
	} else if score >= 75 {
		return "Good"
	} else if score >= 60 {
		return "Fair"
	} else if score >= 40 {
		return "Poor"
	}
	return "Critical"
}

func calculateComplexity(loc, issues, debt int) float64 {
	if loc == 0 {
		return 0
	}
	return float64(issues+debt) / float64(loc) * 100
}

func estimateLinesOfCode(filePath string) int {
	// This is a rough estimation - in a real implementation,
	// you'd want to count actual lines
	if strings.Contains(filePath, "test") {
		return 50 + len(filePath)*2 // Test files are usually smaller
	}
	return 100 + len(filePath)*3 // Regular files
}

func calculateCategoryScore(fileScores []FileQualityScore, category string) float64 {
	if len(fileScores) == 0 {
		return 0
	}
	
	total := float64(0)
	count := 0
	
	for _, file := range fileScores {
		switch category {
		case "maintainability":
			total += file.OverallScore
		case "quality":
			total += math.Max(0, 100-float64(file.LintIssues)*2)
		case "coverage":
			total += file.TestCoverage
		case "debt":
			total += math.Max(0, 100-float64(file.TechnicalDebt)*10)
		}
		count++
	}
	
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func generateRecommendations(fileScores []FileQualityScore, authorScores []AuthorQualityScore) []string {
	var recommendations []string
	
	// File-based recommendations
	lowScoreFiles := 0
	highDebtFiles := 0
	
	for _, file := range fileScores {
		if file.OverallScore < 60 {
			lowScoreFiles++
		}
		if file.TechnicalDebt > 5 {
			highDebtFiles++
		}
	}
	
	if lowScoreFiles > len(fileScores)/4 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Focus on improving %d files with quality scores below 60", lowScoreFiles))
	}
	
	if highDebtFiles > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Address technical debt in %d files (TODO/FIXME/HACK comments)", highDebtFiles))
	}
	
	// Author-based recommendations
	if len(authorScores) > 0 {
		avgIssuesPerCommit := float64(0)
		for _, author := range authorScores {
			avgIssuesPerCommit += author.IssuesPerCommit
		}
		avgIssuesPerCommit /= float64(len(authorScores))
		
		if avgIssuesPerCommit > 0.5 {
			recommendations = append(recommendations, 
				"Consider code review processes to reduce lint issues per commit")
		}
	}
	
	// Generic recommendations
	recommendations = append(recommendations, "Set up automated linting in CI/CD pipeline")
	recommendations = append(recommendations, "Implement pre-commit hooks to catch issues early")
	recommendations = append(recommendations, "Regular code review sessions for knowledge sharing")
	
	return recommendations
}

func calculateTrendDirection(score float64) string {
	// This would ideally compare with historical data
	// For now, we'll provide a simple classification
	if score >= 80 {
		return "🟢 Excellent - Maintain current practices"
	} else if score >= 60 {
		return "🟡 Good - Room for improvement"
	} else {
		return "🔴 Needs Attention - Action required"
	}
}

func getScoreColor(score float64) lipgloss.Color {
	if score >= 80 {
		return lipgloss.Color("#00FF00") // Green
	} else if score >= 60 {
		return lipgloss.Color("#FFFF00") // Yellow
	} else {
		return lipgloss.Color("#FF0000") // Red
	}
}

func printTopFiles(fileScores []FileQualityScore, topN int, showBest bool) {
	scores := fileScores
	if !showBest {
		// Reverse sort for worst files
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].OverallScore < scores[j].OverallScore
		})
	}
	
	count := 0
	for _, file := range scores {
		if count >= topN {
			break
		}
		
		scoreColor := getScoreColor(file.OverallScore)
		scoreStyle := lipgloss.NewStyle().Foreground(scoreColor)
		
		fmt.Printf("  %d. %-30s %s (%.1f) - %s, %d issues, %d debt\n",
			count+1,
			truncateFilePath(file.FilePath, 30),
			scoreStyle.Render(fmt.Sprintf("%.1f", file.OverallScore)),
			file.OverallScore,
			file.Maintainability,
			file.LintIssues,
			file.TechnicalDebt)
		count++
	}
	fmt.Println()
}

func categorizeQualityScores(fileScores []FileQualityScore) []charts.ChartData {
	excellent := 0
	good := 0
	fair := 0
	poor := 0
	critical := 0
	
	for _, file := range fileScores {
		if file.OverallScore >= 90 {
			excellent++
		} else if file.OverallScore >= 75 {
			good++
		} else if file.OverallScore >= 60 {
			fair++
		} else if file.OverallScore >= 40 {
			poor++
		} else {
			critical++
		}
	}
	
	var buckets []charts.ChartData
	if excellent > 0 {
		buckets = append(buckets, charts.ChartData{Label: "Excellent (90+)", Value: float64(excellent)})
	}
	if good > 0 {
		buckets = append(buckets, charts.ChartData{Label: "Good (75-89)", Value: float64(good)})
	}
	if fair > 0 {
		buckets = append(buckets, charts.ChartData{Label: "Fair (60-74)", Value: float64(fair)})
	}
	if poor > 0 {
		buckets = append(buckets, charts.ChartData{Label: "Poor (40-59)", Value: float64(poor)})
	}
	if critical > 0 {
		buckets = append(buckets, charts.ChartData{Label: "Critical (<40)", Value: float64(critical)})
	}
	
	return buckets
}

func truncateFilePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= 3 {
		return path[:maxLen]
	}
	// Show end of path for better context
	return "..." + path[len(path)-(maxLen-3):]
}
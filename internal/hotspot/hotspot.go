package hotspot

import (
	"fmt"
	"math"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"codecompass/internal/charts"
	"codecompass/internal/types"

	"github.com/charmbracelet/lipgloss"
)

// HotspotDetector identifies problematic areas in the codebase
type HotspotDetector struct {
	WeightComplexity float64
	WeightChurn      float64
	WeightIssues     float64
	WeightSize       float64
}

// NewHotspotDetector creates a new hotspot detector
func NewHotspotDetector() *HotspotDetector {
	return &HotspotDetector{
		WeightComplexity: 0.3,
		WeightChurn:      0.3,
		WeightIssues:     0.3,
		WeightSize:       0.1,
	}
}

// Hotspot represents a problematic area in the code
type Hotspot struct {
	FilePath     string
	Score        float64
	Complexity   float64
	Churn        int
	Issues       int
	Size         int
	Risk         string
	LastModified time.Time
	Authors      []string
	Priority     int
}

// DetectHotspots analyzes the codebase to find hotspots
func (hd *HotspotDetector) DetectHotspots(
	fileStats map[string]*types.FileStats,
	trackedFiles map[string]bool,
	topN int,
) ([]Hotspot, error) {
	var hotspots []Hotspot
	
	// Get file churn data
	churnData, err := hd.getFileChurnData()
	if err != nil {
		return nil, err
	}
	
	// Get file complexity data (simplified)
	complexityData := hd.estimateComplexity(trackedFiles)
	
	// Calculate hotspot scores
	for filePath := range trackedFiles {
		if shouldIgnoreFile(filePath) {
			continue
		}
		
		issues := 0
		if stats, exists := fileStats[filePath]; exists {
			issues = stats.Count
		}
		
		churn := churnData[filePath]
		complexity := complexityData[filePath]
		size := hd.getFileSize(filePath)
		
		score := hd.calculateHotspotScore(complexity, float64(churn), float64(issues), float64(size))
		
		if score > 0 {
			risk := hd.getRiskLevel(score)
			lastModified := hd.getLastModified(filePath)
			authors := hd.getFileAuthors(filePath)
			priority := hd.calculatePriority(score, churn, issues)
			
			hotspot := Hotspot{
				FilePath:     filePath,
				Score:        score,
				Complexity:   complexity,
				Churn:        churn,
				Issues:       issues,
				Size:         size,
				Risk:         risk,
				LastModified: lastModified,
				Authors:      authors,
				Priority:     priority,
			}
			
			hotspots = append(hotspots, hotspot)
		}
	}
	
	// Sort by score descending
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].Score > hotspots[j].Score
	})
	
	if len(hotspots) > topN {
		hotspots = hotspots[:topN]
	}
	
	return hotspots, nil
}

// PrintHotspotAnalysis displays hotspot analysis with visual elements
func PrintHotspotAnalysis(hotspots []Hotspot) {
	if len(hotspots) == 0 {
		fmt.Println("🎉 No significant hotspots detected!")
		return
	}
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#FF6B6B")).
		PaddingLeft(1).
		PaddingRight(1)
	
	fmt.Println(titleStyle.Render("🔥 Code Hotspot Analysis"))
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println()
	
	// Create hotspot score chart
	var chartData []charts.ChartData
	for i, hotspot := range hotspots {
		if i < 10 { // Top 10 for chart
			label := truncateFilePath(hotspot.FilePath, 35)
			chartData = append(chartData, charts.ChartData{
				Label: label,
				Value: hotspot.Score,
			})
		}
	}
	
	if len(chartData) > 0 {
		// Use wider chart with better visibility
		chart := charts.BarChart(chartData, 120, "📊 Hotspot Severity Scores")
		fmt.Println(chart)
		fmt.Println()
		
		// Add numerical breakdown for clarity
		fmt.Println("📊 Severity Score Breakdown:")
		for i, data := range chartData {
			score := data.Value
			risk := "Low"
			if score >= 7 {
				risk = "Critical"
			} else if score >= 5 {
				risk = "High"
			} else if score >= 3 {
				risk = "Medium"
			}
			fmt.Printf("  %d. %-35s Score: %6.1f (%s)\n", i+1, data.Label, score, risk)
		}
		fmt.Println()
	}
	
	// Risk distribution
	riskCounts := make(map[string]int)
	for _, hotspot := range hotspots {
		riskCounts[hotspot.Risk]++
	}
	
	var riskData []charts.ChartData
	for risk, count := range riskCounts {
		riskData = append(riskData, charts.ChartData{
			Label: risk,
			Value: float64(count),
		})
	}
	
	if len(riskData) > 0 {
		riskChart := charts.PieChart(riskData, "⚠️  Risk Level Distribution")
		fmt.Println(riskChart)
		fmt.Println()
	}
	
	// Detailed hotspot list
	fmt.Println(titleStyle.Render("📋 Detailed Hotspot Report"))
	fmt.Println()
	
	for i, hotspot := range hotspots {
		if i >= 15 { // Show top 15 detailed
			break
		}
		
		riskColor := getRiskColor(hotspot.Risk)
		riskStyle := lipgloss.NewStyle().Foreground(riskColor).Bold(true)
		
		fmt.Printf("🔥 #%d %s (Score: %.1f)\n", i+1, hotspot.FilePath, hotspot.Score)
		fmt.Printf("   Risk: %s | Priority: %d\n", 
			riskStyle.Render(hotspot.Risk), hotspot.Priority)
		fmt.Printf("   📊 Complexity: %.1f | 🔄 Churn: %d | ❗ Issues: %d | 📏 Size: %d lines\n",
			hotspot.Complexity, hotspot.Churn, hotspot.Issues, hotspot.Size)
		
		if len(hotspot.Authors) > 0 {
			authors := strings.Join(hotspot.Authors, ", ")
			if len(authors) > 50 {
				authors = authors[:47] + "..."
			}
			fmt.Printf("   👥 Authors: %s\n", authors)
		}
		
		fmt.Printf("   📅 Last Modified: %s\n", hotspot.LastModified.Format("2006-01-02 15:04"))
		fmt.Println()
	}
	
	// Recommendations
	fmt.Println(titleStyle.Render("💡 Hotspot Recommendations"))
	fmt.Println()
	
	recommendations := generateHotspotRecommendations(hotspots)
	for i, rec := range recommendations {
		if i < 7 { // Top 7 recommendations
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	}
	fmt.Println()
}

// Helper methods
func (hd *HotspotDetector) calculateHotspotScore(complexity, churn, issues, size float64) float64 {
	// Normalize values (rough scaling)
	normalizedComplexity := math.Min(complexity/10, 10)
	normalizedChurn := math.Min(churn/20, 10)
	normalizedIssues := math.Min(issues/10, 10)
	normalizedSize := math.Min(size/1000, 10)
	
	score := (normalizedComplexity * hd.WeightComplexity) +
		(normalizedChurn * hd.WeightChurn) +
		(normalizedIssues * hd.WeightIssues) +
		(normalizedSize * hd.WeightSize)
	
	return score
}

func (hd *HotspotDetector) getFileChurnData() (map[string]int, error) {
	// Get file change frequency over last 90 days
	since := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	cmd := exec.Command("git", "log", "--since="+since, "--name-only", "--pretty=format:")
	output, err := cmd.Output()
	if err != nil {
		return make(map[string]int), nil // Return empty map if git fails
	}
	
	churnData := make(map[string]int)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			churnData[line]++
		}
	}
	
	return churnData, nil
}

func (hd *HotspotDetector) estimateComplexity(trackedFiles map[string]bool) map[string]float64 {
	complexity := make(map[string]float64)
	
	for filePath := range trackedFiles {
		// Simple complexity estimation based on file type and size
		baseComplexity := float64(1)
		
		if strings.HasSuffix(filePath, ".go") {
			baseComplexity = 2 // Go files tend to be more structured
		} else if strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".ts") {
			baseComplexity = 3 // JS/TS can be more complex
		} else if strings.HasSuffix(filePath, ".py") {
			baseComplexity = 2.5
		}
		
		// Adjust for file size (rough estimation)
		size := hd.getFileSize(filePath)
		sizeMultiplier := 1 + float64(size)/1000
		
		complexity[filePath] = baseComplexity * sizeMultiplier
	}
	
	return complexity
}

func (hd *HotspotDetector) getFileSize(filePath string) int {
	// Rough estimation - in practice, you'd read the actual file
	baseSize := 100
	if strings.Contains(filePath, "test") {
		baseSize = 50
	} else if strings.HasSuffix(filePath, ".md") {
		baseSize = 30
	}
	
	return baseSize + len(filePath)*2
}

func (hd *HotspotDetector) getRiskLevel(score float64) string {
	if score >= 7 {
		return "Critical"
	} else if score >= 5 {
		return "High"
	} else if score >= 3 {
		return "Medium"
	} else if score >= 1 {
		return "Low"
	}
	return "Minimal"
}

func (hd *HotspotDetector) getLastModified(filePath string) time.Time {
	cmd := exec.Command("git", "log", "-1", "--format=%ct", "--", filePath)
	output, err := cmd.Output()
	if err != nil {
		return time.Now().AddDate(0, 0, -30) // Default to 30 days ago
	}
	
	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return time.Now().AddDate(0, 0, -30)
	}
	
	return time.Unix(timestamp, 0)
}

func (hd *HotspotDetector) getFileAuthors(filePath string) []string {
	cmd := exec.Command("git", "log", "--format=%an", "--", filePath)
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}
	
	authorMap := make(map[string]bool)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			authorMap[line] = true
		}
	}
	
	var authors []string
	for author := range authorMap {
		authors = append(authors, author)
		if len(authors) >= 5 { // Limit to top 5 authors
			break
		}
	}
	
	return authors
}

func (hd *HotspotDetector) calculatePriority(score float64, churn, issues int) int {
	// Higher score, churn, and issues = higher priority
	priority := int(score * 2)
	
	if churn > 10 {
		priority += 2
	}
	if issues > 5 {
		priority += 3
	}
	
	// Cap at priority 10
	if priority > 10 {
		priority = 10
	}
	
	return priority
}

func shouldIgnoreFile(filePath string) bool {
	ignorePatterns := []string{
		".git/", "node_modules/", "vendor/", ".idea/",
		"build/", "dist/", "target/", ".DS_Store",
		".gitignore", "*.log", "*.tmp",
	}
	
	for _, pattern := range ignorePatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}
	
	return false
}

func getRiskColor(risk string) lipgloss.Color {
	switch risk {
	case "Critical":
		return lipgloss.Color("#FF0000")
	case "High":
		return lipgloss.Color("#FF6B6B")
	case "Medium":
		return lipgloss.Color("#FFFF00")
	case "Low":
		return lipgloss.Color("#90EE90")
	default:
		return lipgloss.Color("#CCCCCC")
	}
}

func generateHotspotRecommendations(hotspots []Hotspot) []string {
	var recommendations []string
	
	if len(hotspots) == 0 {
		return []string{"Great job! No significant hotspots detected."}
	}
	
	// Count risk levels
	criticalCount := 0
	highCount := 0
	
	for _, hotspot := range hotspots {
		if hotspot.Risk == "Critical" {
			criticalCount++
		} else if hotspot.Risk == "High" {
			highCount++
		}
	}
	
	if criticalCount > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("🚨 URGENT: Address %d critical risk files immediately", criticalCount))
		recommendations = append(recommendations, 
			"Consider refactoring or breaking down large, complex files")
	}
	
	if highCount > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("⚠️  Schedule refactoring for %d high-risk files", highCount))
	}
	
	// Check for high churn files
	highChurnFiles := 0
	for _, hotspot := range hotspots {
		if hotspot.Churn > 15 {
			highChurnFiles++
		}
	}
	
	if highChurnFiles > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("📈 %d files have high churn - consider stabilizing these areas", highChurnFiles))
	}
	
	// Generic recommendations
	recommendations = append(recommendations, 
		"📋 Implement code complexity metrics in CI/CD pipeline")
	recommendations = append(recommendations, 
		"👥 Assign code ownership for hotspot files to specific team members")
	recommendations = append(recommendations, 
		"🧪 Increase test coverage for high-risk files")
	recommendations = append(recommendations, 
		"📝 Add comprehensive documentation for complex files")
	
	return recommendations
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


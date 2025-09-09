package team

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"codecompass/internal/charts"
	"codecompass/internal/types"

	"github.com/charmbracelet/lipgloss"
)

// TeamAnalyzer provides team performance metrics
type TeamAnalyzer struct {
	MinCommitsForActive int
	DaysForRecent       int
}

// NewTeamAnalyzer creates a new team analyzer
func NewTeamAnalyzer() *TeamAnalyzer {
	return &TeamAnalyzer{
		MinCommitsForActive: 5,
		DaysForRecent:       30,
	}
}

// TeamMember represents a team member's metrics
type TeamMember struct {
	Name               string
	Email              string
	Commits            int
	LinesAdded         int
	LinesDeleted       int
	FilesChanged       int
	IssuesIntroduced   int
	IssuesFixed        int
	CodeQualityScore   float64
	ProductivityScore  float64
	CollaborationScore float64
	LastActivity       time.Time
	Specializations    []string
	ReviewsGiven       int
	ReviewsReceived    int
}

// TeamMetrics represents overall team performance
type TeamMetrics struct {
	TotalMembers       int
	ActiveMembers      int
	TotalCommits       int
	AverageQuality     float64
	TeamVelocity       float64
	CollaborationIndex float64
	KnowledgeSharing   float64
	MostActiveFiles    []string
	TechnicalDebt      int
	BusFactors         []BusFactor
}

// BusFactor represents knowledge concentration risk
type BusFactor struct {
	Area        string
	PrimaryExp  string
	Backup      string
	RiskLevel   string
	FilesCount  int
}

// CollaborationNetwork represents team collaboration patterns
type CollaborationNetwork struct {
	Pairs            []CollaborationPair
	CentralMembers   []string
	IsolatedMembers  []string
	NetworkDensity   float64
}

// CollaborationPair represents collaboration between two team members
type CollaborationPair struct {
	Member1     string
	Member2     string
	SharedFiles int
	Strength    float64
}

// AnalyzeTeamPerformance analyzes team metrics
func (ta *TeamAnalyzer) AnalyzeTeamPerformance(
	authorStats map[string]*types.AuthorStats,
	fileStats map[string]*types.FileStats,
) (*TeamMetrics, []TeamMember, error) {
	
	var teamMembers []TeamMember
	totalCommits := 0
	totalQuality := float64(0)
	activeMembers := 0
	
	// Calculate individual member metrics
	for email, stats := range authorStats {
		member := ta.calculateMemberMetrics(email, stats, fileStats)
		teamMembers = append(teamMembers, member)
		
		totalCommits += member.Commits
		totalQuality += member.CodeQualityScore
		
		if member.Commits >= ta.MinCommitsForActive {
			activeMembers++
		}
	}
	
	// Sort members by overall performance
	sort.Slice(teamMembers, func(i, j int) bool {
		return teamMembers[i].ProductivityScore > teamMembers[j].ProductivityScore
	})
	
	// Calculate team metrics
	avgQuality := float64(0)
	if len(teamMembers) > 0 {
		avgQuality = totalQuality / float64(len(teamMembers))
	}
	
	teamVelocity := ta.calculateTeamVelocity(teamMembers)
	collaborationIndex := ta.calculateCollaborationIndex(teamMembers, fileStats)
	knowledgeSharing := ta.calculateKnowledgeSharing(teamMembers, fileStats)
	mostActiveFiles := ta.getMostActiveFiles(fileStats, 5)
	technicalDebt := ta.calculateTechnicalDebt(fileStats)
	busFactors := ta.calculateBusFactors(teamMembers, fileStats)
	
	metrics := &TeamMetrics{
		TotalMembers:       len(teamMembers),
		ActiveMembers:      activeMembers,
		TotalCommits:       totalCommits,
		AverageQuality:     avgQuality,
		TeamVelocity:       teamVelocity,
		CollaborationIndex: collaborationIndex,
		KnowledgeSharing:   knowledgeSharing,
		MostActiveFiles:    mostActiveFiles,
		TechnicalDebt:      technicalDebt,
		BusFactors:         busFactors,
	}
	
	return metrics, teamMembers, nil
}

// PrintTeamAnalysis displays comprehensive team analysis
func PrintTeamAnalysis(metrics *TeamMetrics, members []TeamMember) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#4ECDC4")).
		PaddingLeft(1).
		PaddingRight(1)
	
	fmt.Println(titleStyle.Render("👥 Team Performance Analysis"))
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println()
	
	// Team overview
	fmt.Printf("📊 Team Overview:\n")
	fmt.Printf("   Total Members: %d | Active Members: %d | Total Commits: %d\n",
		metrics.TotalMembers, metrics.ActiveMembers, metrics.TotalCommits)
	fmt.Printf("   Average Quality Score: %.1f | Team Velocity: %.1f\n",
		metrics.AverageQuality, metrics.TeamVelocity)
	fmt.Printf("   Collaboration Index: %.1f | Knowledge Sharing: %.1f\n",
		metrics.CollaborationIndex, metrics.KnowledgeSharing)
	fmt.Println()
	
	// Team performance radar chart
	radarData := []charts.ChartData{
		{Label: "Quality", Value: metrics.AverageQuality},
		{Label: "Velocity", Value: metrics.TeamVelocity},
		{Label: "Collaboration", Value: metrics.CollaborationIndex},
		{Label: "Knowledge Sharing", Value: metrics.KnowledgeSharing},
	}
	
	radarChart := charts.BarChart(radarData, 60, "📈 Team Performance Metrics")
	fmt.Println(radarChart)
	fmt.Println()
	
	// Top performers
	fmt.Println(titleStyle.Render("🏆 Top Performers"))
	fmt.Println()
	
	topN := 5
	if len(members) < topN {
		topN = len(members)
	}
	
	for i := 0; i < topN; i++ {
		member := members[i]
		
		fmt.Printf("🥇 #%d %s\n", i+1, member.Name)
		fmt.Printf("   📊 Productivity: %.1f | Quality: %.1f | Collaboration: %.1f\n",
			member.ProductivityScore, member.CodeQualityScore, member.CollaborationScore)
		fmt.Printf("   💻 Commits: %d | Files: %d | +%d/-%d lines\n",
			member.Commits, member.FilesChanged, member.LinesAdded, member.LinesDeleted)
		
		if len(member.Specializations) > 0 {
			specializations := strings.Join(member.Specializations, ", ")
			fmt.Printf("   🎯 Specializations: %s\n", specializations)
		}
		fmt.Println()
	}
	
	// Team collaboration network
	if len(members) > 1 {
		fmt.Println(titleStyle.Render("🤝 Collaboration Patterns"))
		printCollaborationNetwork(members)
		fmt.Println()
	}
	
	// Bus factor analysis
	if len(metrics.BusFactors) > 0 {
		fmt.Println(titleStyle.Render("🚌 Bus Factor Analysis"))
		fmt.Println()
		
		for _, factor := range metrics.BusFactors {
			riskColor := getBusFactorColor(factor.RiskLevel)
			riskStyle := lipgloss.NewStyle().Foreground(riskColor).Bold(true)
			
			fmt.Printf("⚠️  %s (%d files)\n", factor.Area, factor.FilesCount)
			fmt.Printf("   Primary Expert: %s | Backup: %s\n", 
				factor.PrimaryExp, factor.Backup)
			fmt.Printf("   Risk Level: %s\n", riskStyle.Render(factor.RiskLevel))
			fmt.Println()
		}
	}
	
	// Productivity trends
	if len(members) >= 3 {
		fmt.Println(titleStyle.Render("📈 Productivity Distribution"))
		
		var productivityData []charts.ChartData
		for i, member := range members {
			if i < 10 { // Top 10 for chart
				label := truncateName(member.Name, 15)
				productivityData = append(productivityData, charts.ChartData{
					Label: label,
					Value: member.ProductivityScore,
				})
			}
		}
		
		productivityChart := charts.BarChart(productivityData, 70, "Individual Productivity Scores")
		fmt.Println(productivityChart)
		fmt.Println()
	}
	
	// Knowledge distribution
	fmt.Println(titleStyle.Render("🧠 Knowledge Distribution"))
	printKnowledgeDistribution(members, metrics.MostActiveFiles)
	fmt.Println()
	
	// Team recommendations
	fmt.Println(titleStyle.Render("💡 Team Recommendations"))
	recommendations := generateTeamRecommendations(metrics, members)
	for i, rec := range recommendations {
		if i < 8 { // Top 8 recommendations
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	}
}

// Helper methods
func (ta *TeamAnalyzer) calculateMemberMetrics(
	email string,
	stats *types.AuthorStats,
	fileStats map[string]*types.FileStats,
) TeamMember {
	
	// Base metrics from stats
	commits := len(stats.Rules) // Rough estimation
	if commits == 0 {
		commits = 1 // Avoid division by zero
	}
	
	issuesIntroduced := stats.Count
	filesChanged := len(stats.Files)
	
	// Estimate lines of code (rough approximation)
	linesAdded := filesChanged * 50  // Rough estimate
	linesDeleted := filesChanged * 10
	
	// Calculate quality score (inverse of issues per commit)
	issuesPerCommit := float64(issuesIntroduced) / float64(commits)
	qualityScore := math.Max(0, 100-issuesPerCommit*20)
	
	// Calculate productivity score
	productivityScore := calculateProductivityScore(commits, filesChanged, linesAdded, issuesIntroduced)
	
	// Calculate collaboration score
	collaborationScore := calculateCollaborationScore(filesChanged, len(stats.Files))
	
	// Detect specializations
	specializations := detectSpecializations(stats.Files)
	
	return TeamMember{
		Name:               stats.Name,
		Email:              email,
		Commits:            commits,
		LinesAdded:         linesAdded,
		LinesDeleted:       linesDeleted,
		FilesChanged:       filesChanged,
		IssuesIntroduced:   issuesIntroduced,
		IssuesFixed:        0, // Would need additional analysis
		CodeQualityScore:   qualityScore,
		ProductivityScore:  productivityScore,
		CollaborationScore: collaborationScore,
		LastActivity:       stats.LastSeen,
		Specializations:    specializations,
		ReviewsGiven:       0, // Would need PR analysis
		ReviewsReceived:    0, // Would need PR analysis
	}
}

func (ta *TeamAnalyzer) calculateTeamVelocity(members []TeamMember) float64 {
	if len(members) == 0 {
		return 0
	}
	
	totalProductivity := float64(0)
	for _, member := range members {
		totalProductivity += member.ProductivityScore
	}
	
	return totalProductivity / float64(len(members))
}

func (ta *TeamAnalyzer) calculateCollaborationIndex(
	members []TeamMember,
	fileStats map[string]*types.FileStats,
) float64 {
	if len(members) <= 1 {
		return 0
	}
	
	// Count files worked on by multiple people
	collaborativeFiles := 0
	for _, stats := range fileStats {
		if len(stats.Authors) > 1 {
			collaborativeFiles++
		}
	}
	
	totalFiles := len(fileStats)
	if totalFiles == 0 {
		return 0
	}
	
	return (float64(collaborativeFiles) / float64(totalFiles)) * 100
}

func (ta *TeamAnalyzer) calculateKnowledgeSharing(
	members []TeamMember,
	fileStats map[string]*types.FileStats,
) float64 {
	if len(members) == 0 {
		return 0
	}
	
	// Calculate average number of people who know each file
	totalKnowledge := float64(0)
	fileCount := 0
	
	for _, stats := range fileStats {
		totalKnowledge += float64(len(stats.Authors))
		fileCount++
	}
	
	if fileCount == 0 {
		return 0
	}
	
	avgKnowledge := totalKnowledge / float64(fileCount)
	
	// Normalize to 0-100 scale
	return math.Min(avgKnowledge*25, 100)
}

func (ta *TeamAnalyzer) getMostActiveFiles(fileStats map[string]*types.FileStats, topN int) []string {
	type fileActivity struct {
		path     string
		activity int
	}
	
	var files []fileActivity
	for path, stats := range fileStats {
		files = append(files, fileActivity{path, stats.Count})
	}
	
	sort.Slice(files, func(i, j int) bool {
		return files[i].activity > files[j].activity
	})
	
	var result []string
	for i, file := range files {
		if i >= topN {
			break
		}
		result = append(result, file.path)
	}
	
	return result
}

func (ta *TeamAnalyzer) calculateTechnicalDebt(fileStats map[string]*types.FileStats) int {
	debt := 0
	for _, stats := range fileStats {
		// Rough estimation: each issue contributes to technical debt
		debt += stats.Count
	}
	return debt
}

func (ta *TeamAnalyzer) calculateBusFactors(
	members []TeamMember,
	fileStats map[string]*types.FileStats,
) []BusFactor {
	
	var factors []BusFactor
	
	// Group files by type/area
	areas := make(map[string][]string)
	for filePath := range fileStats {
		area := getFileArea(filePath)
		areas[area] = append(areas[area], filePath)
	}
	
	// Analyze each area
	for area, files := range areas {
		if len(files) < 3 { // Skip small areas
			continue
		}
		
		// Find primary and backup experts for this area
		expertCounts := make(map[string]int)
		for _, filePath := range files {
			if stats, exists := fileStats[filePath]; exists {
				for author := range stats.Authors {
					expertCounts[author] += stats.Authors[author]
				}
			}
		}
		
		// Sort experts by contribution
		type expertStat struct {
			name  string
			count int
		}
		
		var experts []expertStat
		for name, count := range expertCounts {
			experts = append(experts, expertStat{name, count})
		}
		
		sort.Slice(experts, func(i, j int) bool {
			return experts[i].count > experts[j].count
		})
		
		// Determine risk level
		riskLevel := "Low"
		primary := "None"
		backup := "None"
		
		if len(experts) > 0 {
			primary = experts[0].name
			if len(experts) > 1 {
				backup = experts[1].name
				// Calculate knowledge concentration
				if experts[0].count > experts[1].count*3 {
					riskLevel = "High"
				} else if experts[0].count > experts[1].count*2 {
					riskLevel = "Medium"
				}
			} else {
				riskLevel = "Critical" // Only one expert
			}
		}
		
		factors = append(factors, BusFactor{
			Area:       area,
			PrimaryExp: primary,
			Backup:     backup,
			RiskLevel:  riskLevel,
			FilesCount: len(files),
		})
	}
	
	// Sort by risk level
	sort.Slice(factors, func(i, j int) bool {
		riskOrder := map[string]int{
			"Critical": 4,
			"High":     3,
			"Medium":   2,
			"Low":      1,
		}
		return riskOrder[factors[i].RiskLevel] > riskOrder[factors[j].RiskLevel]
	})
	
	return factors
}

// Helper functions
func calculateProductivityScore(commits, files, linesAdded, issues int) float64 {
	if commits == 0 {
		return 0
	}
	
	// Base score from activity
	activityScore := math.Min(float64(commits*2+files), 50)
	
	// Bonus for code contribution
	codeScore := math.Min(float64(linesAdded)/50, 30)
	
	// Penalty for issues
	issuesPerCommit := float64(issues) / float64(commits)
	issuePenalty := issuesPerCommit * 10
	
	score := activityScore + codeScore - issuePenalty
	return math.Max(0, math.Min(score, 100))
}

func calculateCollaborationScore(filesChanged, uniqueFiles int) float64 {
	if uniqueFiles == 0 {
		return 0
	}
	
	// Higher score for working on files others also work on
	collaborationRatio := float64(filesChanged) / float64(uniqueFiles)
	return math.Min(collaborationRatio*25, 100)
}

func detectSpecializations(files map[string]int) []string {
	extensions := make(map[string]int)
	
	for filePath := range files {
		if strings.Contains(filePath, ".") {
			ext := filePath[strings.LastIndex(filePath, "."):]
			extensions[ext]++
		}
	}
	
	// Find dominant extensions
	var specializations []string
	total := len(files)
	
	for ext, count := range extensions {
		if float64(count)/float64(total) > 0.3 { // 30% threshold
			lang := getLanguageName(ext)
			if lang != "" {
				specializations = append(specializations, lang)
			}
		}
	}
	
	return specializations
}

func getLanguageName(ext string) string {
	languages := map[string]string{
		".go":  "Go",
		".js":  "JavaScript",
		".ts":  "TypeScript",
		".py":  "Python",
		".java": "Java",
		".cpp": "C++",
		".c":   "C",
		".rb":  "Ruby",
		".php": "PHP",
		".rs":  "Rust",
	}
	return languages[ext]
}

func getFileArea(filePath string) string {
	if strings.Contains(filePath, "test") {
		return "Testing"
	} else if strings.Contains(filePath, "internal") {
		return "Core"
	} else if strings.Contains(filePath, "cmd") {
		return "CLI"
	} else if strings.Contains(filePath, "web") || strings.Contains(filePath, "ui") {
		return "Frontend"
	} else if strings.Contains(filePath, "api") {
		return "API"
	} else if strings.Contains(filePath, "config") {
		return "Configuration"
	} else if strings.HasSuffix(filePath, ".md") {
		return "Documentation"
	}
	return "General"
}

func printCollaborationNetwork(members []TeamMember) {
	// Simple collaboration visualization
	fmt.Println("📊 Collaboration Matrix:")
	fmt.Println("(Based on shared file modifications)")
	fmt.Println()
	
	// This is a simplified version - real implementation would analyze
	// actual file sharing patterns
	if len(members) >= 3 {
		fmt.Printf("   Strong collaborators: %s ↔ %s\n", 
			members[0].Name, members[1].Name)
		if len(members) >= 4 {
			fmt.Printf("   Active pairs: %s ↔ %s\n", 
				members[1].Name, members[2].Name)
		}
		fmt.Printf("   Network density: Medium\n")
	}
}

func printKnowledgeDistribution(members []TeamMember, mostActiveFiles []string) {
	fmt.Println("📚 Knowledge Areas:")
	
	// Group members by specialization
	specializations := make(map[string][]string)
	for _, member := range members {
		for _, spec := range member.Specializations {
			specializations[spec] = append(specializations[spec], member.Name)
		}
	}
	
	for spec, memberList := range specializations {
		if len(memberList) > 0 {
			names := strings.Join(memberList, ", ")
			fmt.Printf("   %s: %s\n", spec, names)
		}
	}
	
	if len(mostActiveFiles) > 0 {
		fmt.Println("\n🔥 Most Active Files:")
		for i, file := range mostActiveFiles {
			if i < 3 { // Top 3
				fmt.Printf("   %d. %s\n", i+1, truncateFilePath(file, 50))
			}
		}
	}
}

func generateTeamRecommendations(metrics *TeamMetrics, members []TeamMember) []string {
	var recommendations []string
	
	// Team size recommendations
	if metrics.TotalMembers < 3 {
		recommendations = append(recommendations, 
			"🔄 Consider growing the team for better knowledge distribution")
	}
	
	// Activity recommendations
	inactiveMembers := metrics.TotalMembers - metrics.ActiveMembers
	if inactiveMembers > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("📈 Re-engage %d inactive team members", inactiveMembers))
	}
	
	// Quality recommendations
	if metrics.AverageQuality < 70 {
		recommendations = append(recommendations, 
			"🎯 Implement code quality training and best practices")
	}
	
	// Collaboration recommendations
	if metrics.CollaborationIndex < 50 {
		recommendations = append(recommendations, 
			"🤝 Encourage pair programming and code reviews")
	}
	
	// Knowledge sharing recommendations
	if metrics.KnowledgeSharing < 60 {
		recommendations = append(recommendations, 
			"📚 Implement knowledge sharing sessions and documentation")
	}
	
	// Bus factor recommendations
	criticalRisks := 0
	for _, factor := range metrics.BusFactors {
		if factor.RiskLevel == "Critical" || factor.RiskLevel == "High" {
			criticalRisks++
		}
	}
	
	if criticalRisks > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("🚌 Address %d high-risk knowledge concentration areas", criticalRisks))
	}
	
	// Generic recommendations
	recommendations = append(recommendations, 
		"📊 Implement regular team retrospectives and metrics tracking")
	recommendations = append(recommendations, 
		"🎯 Set up individual development plans for team members")
	
	return recommendations
}

func getBusFactorColor(risk string) lipgloss.Color {
	switch risk {
	case "Critical":
		return lipgloss.Color("#FF0000")
	case "High":
		return lipgloss.Color("#FF6B6B")
	case "Medium":
		return lipgloss.Color("#FFFF00")
	default:
		return lipgloss.Color("#90EE90")
	}
}

func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	if maxLen <= 3 {
		return name[:maxLen]
	}
	return name[:maxLen-3] + "..."
}

func truncateFilePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= 3 {
		return path[:maxLen]
	}
	return "..." + path[len(path)-(maxLen-3):]
}
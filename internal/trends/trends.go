package trends

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"codecompass/internal/charts"
)

// TrendAnalyzer analyzes trends over time
type TrendAnalyzer struct {
	Days int
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer(days int) *TrendAnalyzer {
	return &TrendAnalyzer{Days: days}
}

// AnalyzeCommitTrends analyzes commit activity trends
func (ta *TrendAnalyzer) AnalyzeCommitTrends() ([]charts.TimeSeriesData, error) {
	since := time.Now().AddDate(0, 0, -ta.Days)
	cmd := exec.Command("git", "log", "--since="+since.Format("2006-01-02"), "--pretty=format:%at")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Group commits by day
	commitsByDay := make(map[string]int)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		timestamp, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			continue
		}
		
		date := time.Unix(timestamp, 0).Format("2006-01-02")
		commitsByDay[date]++
	}

	// Create time series data
	var trends []charts.TimeSeriesData
	for i := 0; i < ta.Days; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		count := commitsByDay[date]
		trends = append(trends, charts.TimeSeriesData{
			Time:  date,
			Value: float64(count),
		})
	}

	// Reverse to show oldest first
	for i, j := 0, len(trends)-1; i < j; i, j = i+1, j-1 {
		trends[i], trends[j] = trends[j], trends[i]
	}

	return trends, nil
}

// AnalyzeAuthorActivity analyzes author activity over time
func (ta *TrendAnalyzer) AnalyzeAuthorActivity(topN int) (map[string][]charts.TimeSeriesData, error) {
	since := time.Now().AddDate(0, 0, -ta.Days)
	cmd := exec.Command("git", "log", "--since="+since.Format("2006-01-02"), "--pretty=format:%at|%an")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Group commits by author and day
	authorCommits := make(map[string]map[string]int)
	authorTotals := make(map[string]int)
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}
		
		timestamp, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}
		
		author := parts[1]
		date := time.Unix(timestamp, 0).Format("2006-01-02")
		
		if authorCommits[author] == nil {
			authorCommits[author] = make(map[string]int)
		}
		
		authorCommits[author][date]++
		authorTotals[author]++
	}

	// Get top N authors by total commits
	type authorStat struct {
		name  string
		count int
	}
	
	var authors []authorStat
	for author, count := range authorTotals {
		authors = append(authors, authorStat{author, count})
	}
	
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].count > authors[j].count
	})
	
	if len(authors) > topN {
		authors = authors[:topN]
	}

	// Create time series for each top author
	result := make(map[string][]charts.TimeSeriesData)
	
	for _, author := range authors {
		var trends []charts.TimeSeriesData
		
		for i := 0; i < ta.Days; i++ {
			date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			count := authorCommits[author.name][date]
			trends = append(trends, charts.TimeSeriesData{
				Time:  date,
				Value: float64(count),
			})
		}
		
		// Reverse to show oldest first
		for i, j := 0, len(trends)-1; i < j; i, j = i+1, j-1 {
			trends[i], trends[j] = trends[j], trends[i]
		}
		
		result[author.name] = trends
	}

	return result, nil
}

// AnalyzeFileChurnTrends analyzes file change trends
func (ta *TrendAnalyzer) AnalyzeFileChurnTrends(topFiles int) (map[string][]charts.TimeSeriesData, error) {
	since := time.Now().AddDate(0, 0, -ta.Days)
	cmd := exec.Command("git", "log", "--since="+since.Format("2006-01-02"), "--name-only", "--pretty=format:%at")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse git log output
	fileChangesByDay := make(map[string]map[string]int)
	fileTotals := make(map[string]int)
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var currentDate string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Check if it's a timestamp
		if timestamp, err := strconv.ParseInt(line, 10, 64); err == nil {
			currentDate = time.Unix(timestamp, 0).Format("2006-01-02")
			continue
		}
		
		// It's a filename
		if currentDate != "" {
			if fileChangesByDay[line] == nil {
				fileChangesByDay[line] = make(map[string]int)
			}
			fileChangesByDay[line][currentDate]++
			fileTotals[line]++
		}
	}

	// Get top N files by total changes
	type fileStat struct {
		name  string
		count int
	}
	
	var files []fileStat
	for file, count := range fileTotals {
		files = append(files, fileStat{file, count})
	}
	
	sort.Slice(files, func(i, j int) bool {
		return files[i].count > files[j].count
	})
	
	if len(files) > topFiles {
		files = files[:topFiles]
	}

	// Create time series for each top file
	result := make(map[string][]charts.TimeSeriesData)
	
	for _, file := range files {
		var trends []charts.TimeSeriesData
		
		for i := 0; i < ta.Days; i++ {
			date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			count := fileChangesByDay[file.name][date]
			trends = append(trends, charts.TimeSeriesData{
				Time:  date,
				Value: float64(count),
			})
		}
		
		// Reverse to show oldest first
		for i, j := 0, len(trends)-1; i < j; i, j = i+1, j-1 {
			trends[i], trends[j] = trends[j], trends[i]
		}
		
		result[file.name] = trends
	}

	return result, nil
}

// GenerateTrendSummary creates a summary of all trends
func (ta *TrendAnalyzer) GenerateTrendSummary() (*TrendSummary, error) {
	commitTrends, err := ta.AnalyzeCommitTrends()
	if err != nil {
		return nil, err
	}

	authorTrends, err := ta.AnalyzeAuthorActivity(5)
	if err != nil {
		return nil, err
	}

	fileTrends, err := ta.AnalyzeFileChurnTrends(5)
	if err != nil {
		return nil, err
	}

	return &TrendSummary{
		CommitTrends:  commitTrends,
		AuthorTrends:  authorTrends,
		FileChurnTrends: fileTrends,
		Period:        ta.Days,
	}, nil
}

// TrendSummary contains all trend analysis results
type TrendSummary struct {
	CommitTrends    []charts.TimeSeriesData
	AuthorTrends    map[string][]charts.TimeSeriesData
	FileChurnTrends map[string][]charts.TimeSeriesData
	Period          int
}

// PrintTrendAnalysis prints formatted trend analysis
func PrintTrendAnalysis(summary *TrendSummary) {
	fmt.Printf("\n🧭 Trend Analysis (Last %d days)\n", summary.Period)
	fmt.Println(strings.Repeat("═", 50))

	// Commit activity trend
	if len(summary.CommitTrends) > 0 {
		// Show both daily and weekly aggregated views
		chart := charts.TrendChart(summary.CommitTrends, 80, 12, "📈 Daily Commit Activity")
		fmt.Println(chart)
		
		// Calculate statistics
		total := 0.0
		max := 0.0
		activeDays := 0
		
		for _, trend := range summary.CommitTrends {
			total += trend.Value
			if trend.Value > max {
				max = trend.Value
			}
			if trend.Value > 0 {
				activeDays++
			}
		}
		
		avg := total / float64(len(summary.CommitTrends))
		activityRate := float64(activeDays) / float64(len(summary.CommitTrends)) * 100
		
		fmt.Printf("📊 Stats: %.1f commits/day avg | %.0f max | %d/%d active days (%.1f%%)\n", 
			avg, max, activeDays, len(summary.CommitTrends), activityRate)
		
		// Calculate trend direction with more context
		if len(summary.CommitTrends) >= 7 {
			// Compare recent week vs previous week
			recentWeek := 0.0
			previousWeek := 0.0
			
			for i := len(summary.CommitTrends)-7; i < len(summary.CommitTrends); i++ {
				recentWeek += summary.CommitTrends[i].Value
			}
			
			if len(summary.CommitTrends) >= 14 {
				for i := len(summary.CommitTrends)-14; i < len(summary.CommitTrends)-7; i++ {
					previousWeek += summary.CommitTrends[i].Value
				}
				
				change := ((recentWeek - previousWeek) / previousWeek) * 100
				if change > 20 {
					fmt.Printf("📈 Trend: Strong increase (%.1f%% vs previous week)\n", change)
				} else if change > 5 {
					fmt.Printf("📈 Trend: Moderate increase (%.1f%% vs previous week)\n", change)
				} else if change < -20 {
					fmt.Printf("📉 Trend: Strong decrease (%.1f%% vs previous week)\n", change)
				} else if change < -5 {
					fmt.Printf("📉 Trend: Moderate decrease (%.1f%% vs previous week)\n", change)
				} else {
					fmt.Printf("➡️  Trend: Stable activity (%.1f%% change)\n", change)
				}
			}
		}
		
		// Show weekly breakdown
		if len(summary.CommitTrends) >= 7 {
			fmt.Println("\n📅 Weekly Breakdown:")
			weeklyData := aggregateWeekly(summary.CommitTrends)
			for i, week := range weeklyData {
				weekStart := time.Now().AddDate(0, 0, -(len(weeklyData)-i)*7).Format("Jan 2")
				fmt.Printf("  Week of %s: %.0f commits\n", weekStart, week)
			}
		}
		
		fmt.Println()
	}

	// Top author trends
	fmt.Println("👥 Top Author Activity Trends:")
	for author, trends := range summary.AuthorTrends {
		if len(trends) > 0 {
			sparkline := charts.SparklineChart(extractValues(trends), 30)
			fmt.Printf("  %-20s %s\n", truncate(author, 20), sparkline)
		}
	}
	fmt.Println()

	// File churn trends
	fmt.Println("📁 File Change Frequency Trends:")
	for file, trends := range summary.FileChurnTrends {
		if len(trends) > 0 {
			sparkline := charts.SparklineChart(extractValues(trends), 30)
			fmt.Printf("  %-30s %s\n", truncate(file, 30), sparkline)
		}
	}
}

// Helper functions
func extractValues(data []charts.TimeSeriesData) []float64 {
	values := make([]float64, len(data))
	for i, d := range data {
		values[i] = d.Value
	}
	return values
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// aggregateWeekly aggregates daily data into weekly totals
func aggregateWeekly(daily []charts.TimeSeriesData) []float64 {
	if len(daily) < 7 {
		return []float64{}
	}
	
	weeks := len(daily) / 7
	weekly := make([]float64, weeks)
	
	for i := 0; i < weeks; i++ {
		weekTotal := 0.0
		for j := 0; j < 7; j++ {
			idx := i*7 + j
			if idx < len(daily) {
				weekTotal += daily[idx].Value
			}
		}
		weekly[i] = weekTotal
	}
	
	return weekly
}
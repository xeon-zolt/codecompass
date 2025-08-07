package leaderboard

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"codecompass/internal/config"
	"codecompass/internal/coverage"
	"codecompass/internal/git"
	"codecompass/internal/spellcheck"
	"codecompass/internal/types"

	"github.com/charmbracelet/lipgloss"
)

func GenerateAuthorLeaderboard(authorStats map[string]*types.AuthorStats, topN int) []types.LeaderboardEntry {
	var entries []types.LeaderboardEntry
	for email, stats := range authorStats {
		var topRule string
		var topCount int
		for rule, count := range stats.Rules {
			if count > topCount {
				topRule = rule
				topCount = count
			}
		}
		if topRule == "" {
			topRule = "unknown"
		}

		entries = append(entries, types.LeaderboardEntry{
			Name:     stats.Name,
			Email:    email,
			Count:    stats.Count,
			TopRule:  topRule,
			TopCount: topCount,
			Files:    len(stats.Files),
			Errors:   stats.Errors,
			Warnings: stats.Warnings,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})

	return entries
}

func GenerateFileLeaderboard(fileStats map[string]*types.FileStats, topN int) []types.FileLeaderboardEntry {
	var entries []types.FileLeaderboardEntry
	for _, stats := range fileStats {
		var topRule string
		var topCount int
		for rule, count := range stats.Rules {
			if count > topCount {
				topRule = rule
				topCount = count
			}
		}
		if topRule == "" {
			topRule = "unknown"
		}

		entries = append(entries, types.FileLeaderboardEntry{
			Path:     stats.Path,
			Count:    stats.Count,
			TopRule:  topRule,
			TopCount: topCount,
			Authors:  len(stats.Authors),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})

	return entries
}

func GenerateRuleLeaderboard(ruleStats map[string]*types.RuleStats, topN int) []types.RuleLeaderboardEntry {
	var entries []types.RuleLeaderboardEntry
	for _, stats := range ruleStats {
		entries = append(entries, types.RuleLeaderboardEntry{
			Rule:    stats.Rule,
			Count:   stats.Count,
			Authors: len(stats.Authors),
			Files:   len(stats.Files),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})

	return entries
}

func GenerateLinesOfCodeLeaderboard(trackedFiles map[string]bool, topN int) []types.LinesOfCodeEntry {
	var entries []types.LinesOfCodeEntry

	for filePath := range trackedFiles {
		// Skip binary files and common non-code files
		if shouldSkipFile(filePath) {
			continue
		}

		lineCount, err := git.GetFileLineCount(filePath)
		if err != nil {
			continue
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		entries = append(entries, types.LinesOfCodeEntry{
			Path:  filePath,
			Lines: lineCount,
			Size:  fileInfo.Size(),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Lines > entries[j].Lines
	})

	return entries
}

func GenerateCommitCountLeaderboard(topN int) ([]types.CommitCountEntry, error) {
	authorCommits, err := git.GetAuthorCommitCounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit data: %w", err)
	}

	var entries []types.CommitCountEntry
	for _, stats := range authorCommits {
		entries = append(entries, stats)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Commits > entries[j].Commits
	})

	return entries, nil
}

func GenerateRecentContributorsLeaderboard(topN int) ([]types.RecentContributorEntry, error) {
	recentContributors, err := git.GetRecentContributors(30)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent contributor data: %w", err)
	}

	var entries []types.RecentContributorEntry
	for _, stats := range recentContributors {
		entries = append(entries, stats)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RecentCommits > entries[j].RecentCommits
	})

	return entries, nil
}

func GenerateSummaryStats(authorStats map[string]*types.AuthorStats, fileStats map[string]*types.FileStats, ruleStats map[string]*types.RuleStats) {
	fmt.Println(titleStyle.Render("Repository Summary"))

	totalIssues := 0
	totalErrors := 0
	totalWarnings := 0

	for _, stats := range authorStats {
		totalIssues += stats.Count
		totalErrors += stats.Errors
		totalWarnings += stats.Warnings
	}

	fmt.Printf("  • Total Issues: %s\n", cellStyle.Render(fmt.Sprintf("%d", totalIssues)))
	fmt.Printf("  • Errors: %s, Warnings: %s\n",
		errorStyle.Render(fmt.Sprintf("%d", totalErrors)),
		warningStyle.Render(fmt.Sprintf("%d", totalWarnings)))
	fmt.Printf("  • Authors with issues: %s\n", cellStyle.Render(fmt.Sprintf("%d", len(authorStats))))
	fmt.Printf("  • Files with issues: %s\n", cellStyle.Render(fmt.Sprintf("%d", len(fileStats))))
	fmt.Printf("  • Unique rule violations: %s\n", cellStyle.Render(fmt.Sprintf("%d", len(ruleStats))))

	if len(authorStats) > 0 {
		avgIssuesPerAuthor := float64(totalIssues) / float64(len(authorStats))
		fmt.Printf("  • Average issues per author: %s\n",
			cellStyle.Render(fmt.Sprintf("%.1f", avgIssuesPerAuthor)))
	}

	if len(fileStats) > 0 {
		avgIssuesPerFile := float64(totalIssues) / float64(len(fileStats))
		fmt.Printf("  • Average issues per file: %s\n",
			cellStyle.Render(fmt.Sprintf("%.1f", avgIssuesPerFile)))
	}
}

func GenerateCodeChurnLeaderboard(trackedFiles map[string]bool, topN int) ([]types.ChurnEntry, error) {
	cmd := exec.Command("git", "log", "--numstat", "--pretty=format:")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log data: %w", err)
	}

	churnData := make(map[string]*types.ChurnEntry)

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}

		// Skip binary files
		if parts[0] == "-" || parts[1] == "-" {
			continue
		}

		added, _ := strconv.Atoi(parts[0])
		deleted, _ := strconv.Atoi(parts[1])
		filePath := parts[2]

		if !trackedFiles[filePath] {
			continue
		}

		if churnData[filePath] == nil {
			churnData[filePath] = &types.ChurnEntry{Path: filePath}
		}

		entry := churnData[filePath]
		entry.Changes++
		entry.AddedLines += added
		entry.DeletedLines += deleted
		entry.NetLines += added - deleted
	}

	var entries []types.ChurnEntry
	for _, entry := range churnData {
		entries = append(entries, *entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Changes > entries[j].Changes
	})

	return entries, nil
}

func GenerateBugDensityLeaderboard(trackedFiles map[string]bool, topN int) ([]types.BugDensityEntry, error) {
	cmd := exec.Command("git", "log", "--name-only", "--pretty=format:%H|%s")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit data: %w", err)
	}

	bugRegex := regexp.MustCompile(`(?i)(fix|bug|issue|error|broken|crash|repair)`)

	fileStats := make(map[string]*types.BugDensityEntry)

	lines := strings.Split(string(output), "\n")
	var currentCommitIsBug bool
	var currentFiles []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "|") {
			// Process previous commit's files
			if len(currentFiles) > 0 {
				for _, filePath := range currentFiles {
					if fileStats[filePath] == nil {
						fileStats[filePath] = &types.BugDensityEntry{Path: filePath}
					}

					entry := fileStats[filePath]
					entry.TotalCommits++
					if currentCommitIsBug {
						entry.BugFixes++
					}
				}
			}

			// New commit
			parts := strings.Split(line, "|")
			if len(parts) == 2 {
				message := parts[1]
				currentCommitIsBug = bugRegex.MatchString(message)
				currentFiles = []string{}
			}
		} else {
			// File in current commit
			if trackedFiles[line] {
				currentFiles = append(currentFiles, line)
			}
		}
	}

	var entries []types.BugDensityEntry
	for _, entry := range fileStats {
		if entry.TotalCommits > 0 {
			entry.BugRatio = float64(entry.BugFixes) / float64(entry.TotalCommits) * 100
		}
		if entry.TotalCommits >= 5 { // Only show files with at least 5 commits
			entries = append(entries, *entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].BugRatio > entries[j].BugRatio
	})

	return entries, nil
}

func GenerateTechnicalDebtLeaderboard(trackedFiles map[string]bool, topN int) ([]types.TechnicalDebtEntry, error) {
	var entries []types.TechnicalDebtEntry

	todoRegex := regexp.MustCompile(`(?i)//\s*todo|#\s*todo|/\*\s*todo`)
	fixmeRegex := regexp.MustCompile(`(?i)//\s*fixme|#\s*fixme|/\*\s*fixme`)
	hackRegex := regexp.MustCompile(`(?i)//\s*hack|#\s*hack|/\*\s*hack`)

	for filePath := range trackedFiles {
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}

		var todoCount, fixmeCount, hackCount int
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			if todoRegex.MatchString(line) {
				todoCount++
			}
			if fixmeRegex.MatchString(line) {
				fixmeCount++
			}
			if hackRegex.MatchString(line) {
				hackCount++
			}
		}

		file.Close()

		totalDebt := todoCount + fixmeCount + hackCount
		if totalDebt > 0 {
			entries = append(entries, types.TechnicalDebtEntry{
				Path:       filePath,
				TodoCount:  todoCount,
				FixmeCount: fixmeCount,
				HackCount:  hackCount,
				TotalDebt:  totalDebt,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].TotalDebt > entries[j].TotalDebt
	})

	return entries, nil
}

func GenerateCodeCoverageLeaderboard(trackedFiles map[string]bool, coverageFile string, topN int) ([]types.CoverageEntry, float64) {
	coverageData, err := coverage.ParseCoverageFile(coverageFile)
	if err != nil {
		return nil, 0.0
	}

	entries := coverage.GetCoverageStats(coverageData, trackedFiles)

	// Sort by coverage percentage (lowest first - files that need attention)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CoveragePercent < entries[j].CoveragePercent
	})

	// Overall coverage summary
	totalLines := 0
	coveredLines := 0

	for _, entry := range entries {
		totalLines += entry.LinesTotal
		coveredLines += entry.LinesCovered
	}

	overallCoverage := 0.0
	if totalLines > 0 {
		overallCoverage = float64(coveredLines) / float64(totalLines) * 100
	}

	return entries, overallCoverage
}

func GenerateSpellCheckLeaderboard(trackedFiles map[string]bool, cfg *config.Config, topN int) ([]types.SpellCheckEntry, map[string]*types.SpellCheckAuthorStats, error) {
	entries, authorStats, err := spellcheck.AnalyzeSpelling(trackedFiles, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to analyze spelling: %w", err)
	}

	return entries, authorStats, nil
}

// Helper functions
func shouldSkipFile(filePath string) bool {
	skipExtensions := []string{
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".pdf", ".zip", ".tar", ".gz", ".exe", ".bin",
		".lock", ".log", ".tmp", ".cache",
	}

	skipPaths := []string{
		"node_modules/", ".git/", "dist/", "build/",
		"coverage/", ".nyc_output/", "vendor/",
	}

	for _, ext := range skipExtensions {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}

	for _, path := range skipPaths {
		if strings.Contains(filePath, path) {
			return true
		}
	}

	return false
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days > 365 {
		years := days / 365
		return fmt.Sprintf("%d years", years)
	} else if days > 30 {
		months := days / 30
		return fmt.Sprintf("%d months", months)
	} else if days > 0 {
		return fmt.Sprintf("%d days", days)
	} else if d.Hours() > 1 {
		return fmt.Sprintf("%.0f hours", d.Hours())
	} else {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	}
}

func formatNetLines(net int) string {
	if net > 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(fmt.Sprintf("+%d", net))
	} else if net < 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(fmt.Sprintf("%d", net))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#878787")).Render("0")
}

// Print functions (moved from main.go for better organization)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#5d5d5d")).
		PaddingLeft(1).
		PaddingRight(1)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA"))

	cellStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	rankStyle = cellStyle.Copy().
		Foreground(lipgloss.Color("#878787"))

	nameStyle = cellStyle.Copy().
		Foreground(lipgloss.Color("#d75f00"))

	emailStyle = cellStyle.Copy().
		Foreground(lipgloss.Color("#878787"))

	topRuleStyle = cellStyle.Copy().
		Foreground(lipgloss.Color("#ffd700"))

	errorStyle = cellStyle.Copy().
		Foreground(lipgloss.Color("#ff0000"))

	warningStyle = cellStyle.Copy().
		Foreground(lipgloss.Color("#ffff00"))
)

func PrintAuthorLeaderboard(entries []types.LeaderboardEntry, topN int) {
	fmt.Println(titleStyle.Render("Author Leaderboard - Most ESLint Issues"))

	if len(entries) == 0 {
		fmt.Println(cellStyle.Render("🎉 Everyone's clean. No one to shame."))
		return
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		name := nameStyle.Render(entry.Name)
		email := emailStyle.Render(fmt.Sprintf("(%s)", entry.Email))
		topRule := topRuleStyle.Render(entry.TopRule)
		errors := errorStyle.Render(fmt.Sprintf("%d", entry.Errors))
		warnings := warningStyle.Render(fmt.Sprintf("%d", entry.Warnings))

		fmt.Printf("%s. %s %s – %d issues (%s errors, %s warnings), %d files, top rule: %s (%d)\n",
			rank, name, email, entry.Count, errors, warnings, entry.Files, topRule, entry.TopCount)
	}
}

func PrintFileLeaderboard(entries []types.FileLeaderboardEntry, topN int) {
	fmt.Println(titleStyle.Render("File Leaderboard - Most Problematic Files"))

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		path := cellStyle.Render(entry.Path)
		topRule := topRuleStyle.Render(entry.TopRule)

		fmt.Printf("%s. %s – %d issues, %d authors, top rule: %s (%d)\n",
			rank, path, entry.Count, entry.Authors, topRule, entry.TopCount)
	}
}

func PrintRuleLeaderboard(entries []types.RuleLeaderboardEntry, topN int) {
	fmt.Println(titleStyle.Render("Rule Leaderboard - Most Violated Rules"))

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		rule := cellStyle.Render(entry.Rule)

		fmt.Printf("%s. %s – %d violations, %d authors, %d files\n",
			rank, rule, entry.Count, entry.Authors, entry.Files)
	}
}

func PrintLinesOfCodeLeaderboard(entries []types.LinesOfCodeEntry, topN int) {
	fmt.Println(titleStyle.Render("Lines of Code Leaderboard - Largest Files"))

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		path := cellStyle.Render(entry.Path)
		size := emailStyle.Render(formatFileSize(entry.Size))

		fmt.Printf("%s. %s – %s lines (%s)\n",
			rank, path, cellStyle.Render(fmt.Sprintf("%d", entry.Lines)), size)
	}
}

func PrintCommitCountLeaderboard(entries []types.CommitCountEntry, topN int) {
	fmt.Println(titleStyle.Render("Commit Count Leaderboard - Most Active Contributors"))

	if len(entries) == 0 {
		fmt.Println(cellStyle.Render("📭 No commit data found"))
		return
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		name := nameStyle.Render(entry.Name)
		email := emailStyle.Render(fmt.Sprintf("(%s)", entry.Email))
		timespan := emailStyle.Render(formatDuration(entry.LastCommit.Sub(entry.FirstCommit)))

		fmt.Printf("%s. %s %s – %s commits (active for %s)\n",
			rank, name, email, cellStyle.Render(fmt.Sprintf("%d", entry.Commits)), timespan)
	}
}

func PrintRecentContributorsLeaderboard(entries []types.RecentContributorEntry, topN int) {
	fmt.Println(titleStyle.Render("Recent Contributors Leaderboard - Most Active in Last 30 Days"))

	if len(entries) == 0 {
		fmt.Println(cellStyle.Render("📭 No commits in the last 30 days"))
		return
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		name := nameStyle.Render(entry.Name)
		email := emailStyle.Render(fmt.Sprintf("(%s)", entry.Email))
		lastCommitAgo := emailStyle.Render(formatDuration(time.Since(entry.LastCommit)))

		fmt.Printf("%s. %s %s – %s commits (last: %s ago)\n",
			rank, name, email, cellStyle.Render(fmt.Sprintf("%d", entry.RecentCommits)), lastCommitAgo)
	}
}

func PrintCodeChurnLeaderboard(entries []types.ChurnEntry, topN int) {
	fmt.Println(titleStyle.Render("Code Churn Leaderboard - Most Frequently Changed Files"))

	if len(entries) == 0 {
		fmt.Println(cellStyle.Render("📭 No churn data found"))
		return
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		path := cellStyle.Render(entry.Path)

		fmt.Printf("%s. %s – %s changes (%s lines added, %s deleted, net: %s)\n",
			rank, path, cellStyle.Render(fmt.Sprintf("%d", entry.Changes)),
			cellStyle.Render(fmt.Sprintf("%d", entry.AddedLines)),
			cellStyle.Render(fmt.Sprintf("%d", entry.DeletedLines)),
			formatNetLines(entry.NetLines))
	}
}

func PrintBugDensityLeaderboard(entries []types.BugDensityEntry, topN int) {
	fmt.Println(titleStyle.Render("Bug Density Leaderboard - Files with Highest Bug-Fix Ratio"))

	if len(entries) == 0 {
		fmt.Println(cellStyle.Render("📭 No bug density data found"))
		return
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		path := cellStyle.Render(entry.Path)

		var ratioStyle lipgloss.Style
		if entry.BugRatio > 30 {
			ratioStyle = errorStyle
		} else if entry.BugRatio > 15 {
			ratioStyle = warningStyle
		} else {
			ratioStyle = cellStyle
		}

		fmt.Printf("%s. %s – %s bug-fix ratio (%d fixes out of %d commits)\n",
			rank, path, ratioStyle.Render(fmt.Sprintf("%.1f%%", entry.BugRatio)),
			entry.BugFixes, entry.TotalCommits)
	}
}

func PrintTechnicalDebtLeaderboard(entries []types.TechnicalDebtEntry, topN int) {
	fmt.Println(titleStyle.Render("Technical Debt Leaderboard - Files with Most TODO/FIXME/HACK Comments"))

	if len(entries) == 0 {
		fmt.Println(cellStyle.Render("🎉 No technical debt found (or you have very clean code!)"))
		return
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		path := cellStyle.Render(entry.Path)

		var debtItems []string
		if entry.TodoCount > 0 {
			debtItems = append(debtItems, warningStyle.Render(fmt.Sprintf("%d TODOs", entry.TodoCount)))
		}
		if entry.FixmeCount > 0 {
			debtItems = append(debtItems, errorStyle.Render(fmt.Sprintf("%d FIXMEs", entry.FixmeCount)))
		}
		if entry.HackCount > 0 {
			debtItems = append(debtItems, topRuleStyle.Render(fmt.Sprintf("%d HACKs", entry.HackCount)))
		}

		fmt.Printf("%s. %s – %s total debt (%s)\n",
			rank, path, cellStyle.Render(fmt.Sprintf("%d", entry.TotalDebt)), strings.Join(debtItems, ", "))
	}
}

func PrintCodeCoverageLeaderboard(entries []types.CoverageEntry, overallCoverage float64, topN int) {
	fmt.Println(titleStyle.Render("Code Coverage Leaderboard - Coverage by File"))

	if len(entries) == 0 {
		fmt.Println(cellStyle.Render("📭 No coverage data found for tracked files"))
		return
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	fmt.Println(emailStyle.Render("  (Showing files with lowest coverage - need attention)"))

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		path := cellStyle.Render(entry.Path)

		var coverageStyle lipgloss.Style
		if entry.CoveragePercent >= 80 {
			coverageStyle = cellStyle
		} else if entry.CoveragePercent >= 60 {
			coverageStyle = warningStyle
		} else {
			coverageStyle = errorStyle
		}

		coverageStr := coverageStyle.Render(fmt.Sprintf("%.1f%%", entry.CoveragePercent))

		var additionalInfo []string
		if entry.LinesTotal > 0 {
			additionalInfo = append(additionalInfo, fmt.Sprintf("%d/%d lines", entry.LinesCovered, entry.LinesTotal))
		}
		if entry.FunctionsTotal > 0 {
			functionsPercent := float64(entry.FunctionsCovered) / float64(entry.FunctionsTotal) * 100
			additionalInfo = append(additionalInfo, fmt.Sprintf("%.0f%% functions", functionsPercent))
		}
		if entry.BranchesTotal > 0 {
			branchesPercent := float64(entry.BranchesCovered) / float64(entry.BranchesTotal) * 100
			additionalInfo = append(additionalInfo, fmt.Sprintf("%.0f%% branches", branchesPercent))
		}

		infoStr := ""
		if len(additionalInfo) > 0 {
			infoStr = fmt.Sprintf(" (%s)", emailStyle.Render(strings.Join(additionalInfo, ", ")))
		}

		fmt.Printf("%s. %s – %s%s\n", rank, path, coverageStr, infoStr)
	}

	if len(entries) > topN {
		fmt.Println(cellStyle.Render("\n🏆 Files with highest coverage:"))

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].CoveragePercent > entries[j].CoveragePercent
		})

		maxHighCoverage := 5
		if len(entries) < maxHighCoverage {
			maxHighCoverage = len(entries)
		}

		for i := 0; i < maxHighCoverage; i++ {
			entry := entries[i]
			if entry.CoveragePercent < 80 {
				continue
			}

			path := cellStyle.Render(entry.Path)
			coverageStr := cellStyle.Render(fmt.Sprintf("%.1f%%", entry.CoveragePercent))

			fmt.Printf("     %s – %s\n", path, coverageStr)
		}
	}

	if overallCoverage > 0 {
		fmt.Printf("\n  %s Overall Coverage: %s (%d/%d lines covered)\n",
			cellStyle.Render("📊"),
			cellStyle.Render(fmt.Sprintf("%.1f%%", overallCoverage)),
			(int)(overallCoverage/100*float64(entries[0].LinesTotal)), entries[0].LinesTotal)
	}
}

func PrintSpellCheckLeaderboard(entries []types.SpellCheckEntry, authorStats map[string]*types.SpellCheckAuthorStats, topN int) {
	fmt.Println(titleStyle.Render("Spell Check Leaderboard - Files with Most Spelling Errors"))

	if len(entries) == 0 {
		fmt.Println(cellStyle.Render("🎉 No spelling issues found or no text files to analyze"))
		return
	}

	printSpellCheckFileLeaderboard(entries, topN)
	printSpellCheckAuthorLeaderboard(authorStats, topN)
}

func printSpellCheckFileLeaderboard(entries []types.SpellCheckEntry, topN int) {
	fmt.Println(titleStyle.Render("Files with Most Spelling Errors"))

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		path := cellStyle.Render(entry.Path)

		var errorColor lipgloss.Style
		if entry.ErrorRate > 10 {
			errorColor = errorStyle
		} else if entry.ErrorRate > 5 {
			errorColor = warningStyle
		} else {
			errorColor = cellStyle
		}

		var topMisspellings []string
		type wordCount struct {
			word  string
			count int
		}
		var wordCounts []wordCount
		for word, count := range entry.TopMisspellings {
			wordCounts = append(wordCounts, wordCount{word, count})
		}
		sort.Slice(wordCounts, func(i, j int) bool {
			return wordCounts[i].count > wordCounts[j].count
		})

		for i, wc := range wordCounts {
			if i >= 3 {
				break
			}
			topMisspellings = append(topMisspellings, fmt.Sprintf("%s(%d)", wc.word, wc.count))
		}

		misspellingsStr := ""
		if len(topMisspellings) > 0 {
			misspellingsStr = fmt.Sprintf(" [%s]", strings.Join(topMisspellings, ", "))
		}

		fmt.Printf("    %s. %s – %s error rate (%d/%d words)%s\n",
			rank, path, errorColor.Render(fmt.Sprintf("%.1f%%", entry.ErrorRate)),
			entry.MisspelledWords, entry.TotalWords,
			emailStyle.Render(misspellingsStr))
	}

	if len(entries) > 0 && len(entries[0].Issues) > 0 {
		fmt.Printf("\n  %s Examples from %s:\n",
			warningStyle.Render("🔍"),
			cellStyle.Render(entries[0].Path))

		maxExamples := 5
		for i, issue := range entries[0].Issues {
			if i >= maxExamples {
				break
			}

			suggestionStr := ""
			if len(issue.Suggestions) > 0 {
				suggestionStr = fmt.Sprintf(" → %s", strings.Join(issue.Suggestions, ", "))
			}

			authorStr := ""
			if issue.Author != "unknown" {
				authorStr = fmt.Sprintf(" (by %s)", nameStyle.Render(issue.Author))
			}

			fmt.Printf("    Line %d: '%s' in %s%s%s\n",
				issue.Line,
				errorStyle.Render(issue.Word),
				issue.Type,
				authorStr,
				cellStyle.Render(suggestionStr))
		}
	}
}

func printSpellCheckAuthorLeaderboard(authorStats map[string]*types.SpellCheckAuthorStats, topN int) {
	fmt.Println(titleStyle.Render("Authors with Most Spelling Errors"))

	type authorEntry struct {
		Name            string
		Email           string
		TotalErrors     int
		Files           int
		TopMistake      string
		TopMistakeCount int
	}

	var entries []authorEntry
	for _, stats := range authorStats {
		var topMistake string
		var topCount int
		for mistake, count := range stats.CommonMistakes {
			if count > topCount {
				topMistake = mistake
				topCount = count
			}
		}

		entries = append(entries, authorEntry{
			Name:            stats.Name,
			Email:           stats.Email,
			TotalErrors:     stats.TotalErrors,
			Files:           len(stats.Files),
			TopMistake:      topMistake,
			TopMistakeCount: topCount,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].TotalErrors > entries[j].TotalErrors
	})

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := rankStyle.Render(fmt.Sprintf("%2d", i+1))
		name := nameStyle.Render(entry.Name)
		email := emailStyle.Render(fmt.Sprintf("(%s)", entry.Email))

		topMistakeStr := ""
		if entry.TopMistake != "" {
			topMistakeStr = fmt.Sprintf(", top mistake: %s(%d)",
				topRuleStyle.Render(entry.TopMistake), entry.TopMistakeCount)
		}

		fmt.Printf("    %s. %s %s – %d errors in %d files%s\n",
			rank, name, email, entry.TotalErrors, entry.Files, topMistakeStr)
	}
}
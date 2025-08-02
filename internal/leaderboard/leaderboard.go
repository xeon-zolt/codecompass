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

	"github.com/fatih/color"
)

func GenerateAuthorLeaderboard(authorStats map[string]*types.AuthorStats, topN int) {
	fmt.Printf("\n🏆 %s\n", color.New(color.Bold).Sprint("Author Leaderboard - Most ESLint Issues:"))

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

	if len(entries) == 0 {
		fmt.Println(color.GreenString("🎉 Everyone's clean. No one to shame."))
		return
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		name := color.RedString(entry.Name)
		email := color.New(color.FgHiBlack).Sprintf("(%s)", entry.Email)
		topRule := color.YellowString(entry.TopRule)
		errorColor := color.New(color.FgRed)
		warningColor := color.New(color.FgYellow)

		fmt.Printf("%s. %s %s – %d issues (%s errors, %s warnings), %d files, top rule: %s (%d)\n",
			rank, name, email, entry.Count,
			errorColor.Sprintf("%d", entry.Errors),
			warningColor.Sprintf("%d", entry.Warnings),
			entry.Files, topRule, entry.TopCount)
	}
}

func GenerateFileLeaderboard(fileStats map[string]*types.FileStats, topN int) {
	fmt.Printf("\n📁 %s\n", color.New(color.Bold).Sprint("File Leaderboard - Most Problematic Files:"))

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

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		path := color.CyanString(entry.Path)
		topRule := color.YellowString(entry.TopRule)

		fmt.Printf("%s. %s – %d issues, %d authors, top rule: %s (%d)\n",
			rank, path, entry.Count, entry.Authors, topRule, entry.TopCount)
	}
}

func GenerateRuleLeaderboard(ruleStats map[string]*types.RuleStats, topN int) {
	fmt.Printf("\n📏 %s\n", color.New(color.Bold).Sprint("Rule Leaderboard - Most Violated Rules:"))

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

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		rule := color.MagentaString(entry.Rule)

		fmt.Printf("%s. %s – %d violations, %d authors, %d files\n",
			rank, rule, entry.Count, entry.Authors, entry.Files)
	}
}

func GenerateLinesOfCodeLeaderboard(trackedFiles map[string]bool, topN int) {
	fmt.Printf("\n📝 %s\n", color.New(color.Bold).Sprint("Lines of Code Leaderboard - Largest Files:"))

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

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		path := color.GreenString(entry.Path)
		size := formatFileSize(entry.Size)

		fmt.Printf("%s. %s – %s lines (%s)\n",
			rank, path,
			color.New(color.Bold).Sprintf("%d", entry.Lines),
			color.New(color.FgHiBlack).Sprint(size))
	}
}

func GenerateCommitCountLeaderboard(topN int) {
	fmt.Printf("\n📊 %s\n", color.New(color.Bold).Sprint("Commit Count Leaderboard - Most Active Contributors:"))

	authorCommits, err := git.GetAuthorCommitCounts()
	if err != nil {
		fmt.Printf("  %s Failed to get commit data: %v\n", color.RedString("❌"), err)
		return
	}

	var entries []types.CommitCountEntry
	for _, stats := range authorCommits {
		entries = append(entries, stats)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Commits > entries[j].Commits
	})

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		name := color.BlueString(entry.Name)
		email := color.New(color.FgHiBlack).Sprintf("(%s)", entry.Email)
		timespan := entry.LastCommit.Sub(entry.FirstCommit)

		fmt.Printf("%s. %s %s – %s commits (active for %s)\n",
			rank, name, email,
			color.New(color.Bold).Sprintf("%d", entry.Commits),
			color.New(color.FgHiBlack).Sprint(formatDuration(timespan)))
	}
}

func GenerateRecentContributorsLeaderboard(topN int) {
	fmt.Printf("\n🕒 %s\n", color.New(color.Bold).Sprint("Recent Contributors Leaderboard - Most Active in Last 30 Days:"))

	recentContributors, err := git.GetRecentContributors(30)
	if err != nil {
		fmt.Printf("  %s Failed to get recent contributor data: %v\n", color.RedString("❌"), err)
		return
	}

	var entries []types.RecentContributorEntry
	for _, stats := range recentContributors {
		entries = append(entries, stats)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RecentCommits > entries[j].RecentCommits
	})

	if len(entries) == 0 {
		fmt.Printf("  %s No commits in the last 30 days\n", color.YellowString("📭"))
		return
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		name := color.MagentaString(entry.Name)
		email := color.New(color.FgHiBlack).Sprintf("(%s)", entry.Email)
		lastCommitAgo := time.Since(entry.LastCommit)

		fmt.Printf("%s. %s %s – %s commits (last: %s ago)\n",
			rank, name, email,
			color.New(color.Bold).Sprintf("%d", entry.RecentCommits),
			color.New(color.FgHiBlack).Sprint(formatDuration(lastCommitAgo)))
	}
}

func GenerateSummaryStats(authorStats map[string]*types.AuthorStats, fileStats map[string]*types.FileStats, ruleStats map[string]*types.RuleStats) {
	fmt.Printf("\n📋 %s\n", color.New(color.Bold).Sprint("Repository Summary:"))

	totalIssues := 0
	totalErrors := 0
	totalWarnings := 0

	for _, stats := range authorStats {
		totalIssues += stats.Count
		totalErrors += stats.Errors
		totalWarnings += stats.Warnings
	}

	fmt.Printf("  • Total Issues: %s\n", color.New(color.Bold).Sprintf("%d", totalIssues))
	fmt.Printf("  • Errors: %s, Warnings: %s\n",
		color.RedString("%d", totalErrors),
		color.YellowString("%d", totalWarnings))
	fmt.Printf("  • Authors with issues: %s\n", color.New(color.Bold).Sprintf("%d", len(authorStats)))
	fmt.Printf("  • Files with issues: %s\n", color.New(color.Bold).Sprintf("%d", len(fileStats)))
	fmt.Printf("  • Unique rule violations: %s\n", color.New(color.Bold).Sprintf("%d", len(ruleStats)))

	if len(authorStats) > 0 {
		avgIssuesPerAuthor := float64(totalIssues) / float64(len(authorStats))
		fmt.Printf("  • Average issues per author: %s\n",
			color.New(color.Bold).Sprintf("%.1f", avgIssuesPerAuthor))
	}

	if len(fileStats) > 0 {
		avgIssuesPerFile := float64(totalIssues) / float64(len(fileStats))
		fmt.Printf("  • Average issues per file: %s\n",
			color.New(color.Bold).Sprintf("%.1f", avgIssuesPerFile))
	}
}

func GenerateCodeChurnLeaderboard(trackedFiles map[string]bool, topN int) error {
	fmt.Printf("\n🔄 %s\n", color.New(color.Bold).Sprint("Code Churn Leaderboard - Most Frequently Changed Files:"))

	cmd := exec.Command("git", "log", "--numstat", "--pretty=format:")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  %s Failed to get git log data: %v\n", color.RedString("❌"), err)
		return err
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

	if len(entries) == 0 {
		fmt.Printf("  %s No churn data found\n", color.YellowString("📭"))
		return nil
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		path := color.CyanString(entry.Path)

		fmt.Printf("%s. %s – %s changes (%s lines added, %s deleted, net: %s)\n",
			rank, path,
			color.New(color.Bold).Sprintf("%d", entry.Changes),
			color.GreenString("%d", entry.AddedLines),
			color.RedString("%d", entry.DeletedLines),
			formatNetLines(entry.NetLines))
	}

	return nil
}

func GenerateBugDensityLeaderboard(trackedFiles map[string]bool, topN int) error {
	fmt.Printf("\n🐛 %s\n", color.New(color.Bold).Sprint("Bug Density Leaderboard - Files with Highest Bug-Fix Ratio:"))

	cmd := exec.Command("git", "log", "--name-only", "--pretty=format:%H|%s")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  %s Failed to get commit data: %v\n", color.RedString("❌"), err)
		return err
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

	if len(entries) == 0 {
		fmt.Printf("  %s No bug density data found\n", color.YellowString("📭"))
		return nil
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		path := color.CyanString(entry.Path)

		var ratioColor *color.Color
		if entry.BugRatio > 30 {
			ratioColor = color.New(color.FgRed)
		} else if entry.BugRatio > 15 {
			ratioColor = color.New(color.FgYellow)
		} else {
			ratioColor = color.New(color.FgGreen)
		}

		fmt.Printf("%s. %s – %s bug-fix ratio (%d fixes out of %d commits)\n",
			rank, path,
			ratioColor.Sprintf("%.1f%%", entry.BugRatio),
			entry.BugFixes, entry.TotalCommits)
	}

	return nil
}

func GenerateTechnicalDebtLeaderboard(trackedFiles map[string]bool, topN int) error {
	fmt.Printf("\n💸 %s\n", color.New(color.Bold).Sprint("Technical Debt Leaderboard - Files with Most TODO/FIXME/HACK Comments:"))

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

	if len(entries) == 0 {
		fmt.Printf("  %s No technical debt found (or you have very clean code!)\n", color.GreenString("🎉"))
		return nil
	}

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		path := color.CyanString(entry.Path)

		var debtItems []string
		if entry.TodoCount > 0 {
			debtItems = append(debtItems, color.YellowString("%d TODOs", entry.TodoCount))
		}
		if entry.FixmeCount > 0 {
			debtItems = append(debtItems, color.RedString("%d FIXMEs", entry.FixmeCount))
		}
		if entry.HackCount > 0 {
			debtItems = append(debtItems, color.MagentaString("%d HACKs", entry.HackCount))
		}

		fmt.Printf("%s. %s – %s total debt (%s)\n",
			rank, path,
			color.New(color.Bold).Sprintf("%d", entry.TotalDebt),
			strings.Join(debtItems, ", "))
	}

	return nil
}

func GenerateCodeCoverageLeaderboard(trackedFiles map[string]bool, coverageFile string, topN int) {
	fmt.Printf("\n📈 %s\n", color.New(color.Bold).Sprint("Code Coverage Leaderboard - Coverage by File:"))

	coverageData, err := coverage.ParseCoverageFile(coverageFile)
	if err != nil {
		fmt.Printf("  %s Failed to parse coverage data: %v\n", color.RedString("❌"), err)
		fmt.Printf("  %s Try specifying a coverage file with --coverage-file\n", color.YellowString("💡"))
		return
	}

	entries := coverage.GetCoverageStats(coverageData, trackedFiles)

	if len(entries) == 0 {
		fmt.Printf("  %s No coverage data found for tracked files\n", color.YellowString("📭"))
		return
	}

	// Sort by coverage percentage (lowest first - files that need attention)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CoveragePercent < entries[j].CoveragePercent
	})

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	fmt.Printf("  %s (Showing files with lowest coverage - need attention)\n",
		color.New(color.FgHiBlack).Sprint("Sorted by coverage percentage"))

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		path := color.CyanString(entry.Path)

		// Color-code coverage percentage
		var coverageColor *color.Color
		if entry.CoveragePercent >= 80 {
			coverageColor = color.New(color.FgGreen)
		} else if entry.CoveragePercent >= 60 {
			coverageColor = color.New(color.FgYellow)
		} else {
			coverageColor = color.New(color.FgRed)
		}

		coverageStr := coverageColor.Sprintf("%.1f%%", entry.CoveragePercent)

		// Build additional info
		var additionalInfo []string
		if entry.LinesTotal > 0 {
			additionalInfo = append(additionalInfo,
				fmt.Sprintf("%d/%d lines", entry.LinesCovered, entry.LinesTotal))
		}
		if entry.FunctionsTotal > 0 {
			functionsPercent := float64(entry.FunctionsCovered) / float64(entry.FunctionsTotal) * 100
			additionalInfo = append(additionalInfo,
				fmt.Sprintf("%.0f%% functions", functionsPercent))
		}
		if entry.BranchesTotal > 0 {
			branchesPercent := float64(entry.BranchesCovered) / float64(entry.BranchesTotal) * 100
			additionalInfo = append(additionalInfo,
				fmt.Sprintf("%.0f%% branches", branchesPercent))
		}

		infoStr := ""
		if len(additionalInfo) > 0 {
			infoStr = fmt.Sprintf(" (%s)",
				color.New(color.FgHiBlack).Sprint(strings.Join(additionalInfo, ", ")))
		}

		fmt.Printf("%s. %s – %s%s\n", rank, path, coverageStr, infoStr)
	}

	// Show a few high coverage files as well
	if len(entries) > topN {
		fmt.Printf("\n  %s\n", color.GreenString("🏆 Files with highest coverage:"))

		// Sort by coverage percentage (highest first)
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
				continue // Skip if not high coverage
			}

			path := color.CyanString(entry.Path)
			coverageStr := color.GreenString("%.1f%%", entry.CoveragePercent)

			fmt.Printf("     %s – %s\n", path, coverageStr)
		}
	}

	// Overall coverage summary
	totalLines := 0
	coveredLines := 0
	totalFunctions := 0
	coveredFunctions := 0
	totalBranches := 0
	coveredBranches := 0

	for _, entry := range entries {
		totalLines += entry.LinesTotal
		coveredLines += entry.LinesCovered
		totalFunctions += entry.FunctionsTotal
		coveredFunctions += entry.FunctionsCovered
		totalBranches += entry.BranchesTotal
		coveredBranches += entry.BranchesCovered
	}

	if totalLines > 0 {
		overallCoverage := float64(coveredLines) / float64(totalLines) * 100
		fmt.Printf("\n  %s Overall Coverage: %s (%d/%d lines covered)\n",
			color.BlueString("📊"),
			color.New(color.Bold).Sprintf("%.1f%%", overallCoverage),
			coveredLines, totalLines)
	}
}

func GenerateSpellCheckLeaderboard(trackedFiles map[string]bool, cfg *config.Config, topN int) error {
	fmt.Printf("\n📝 %s\n", color.New(color.Bold).Sprint("Spell Check Leaderboard - Files with Most Spelling Errors:"))

	entries, authorStats, err := spellcheck.AnalyzeSpelling(trackedFiles, cfg)
	if err != nil {
		fmt.Printf("  %s Failed to analyze spelling: %v\n", color.RedString("❌"), err)
		return err
	}

	if len(entries) == 0 {
		fmt.Printf("  %s No spelling issues found or no text files to analyze\n", color.GreenString("🎉"))
		return nil
	}

	// Show file leaderboard
	generateSpellCheckFileLeaderboard(entries, topN)

	// Show author leaderboard
	generateSpellCheckAuthorLeaderboard(authorStats, topN)

	return nil
}

func generateSpellCheckFileLeaderboard(entries []types.SpellCheckEntry, topN int) {
	fmt.Printf("\n  %s\n", color.New(color.Bold).Sprint("📁 Files with Most Spelling Errors:"))

	maxEntries := topN
	if len(entries) < maxEntries {
		maxEntries = len(entries)
	}

	for i := 0; i < maxEntries; i++ {
		entry := entries[i]
		rank := fmt.Sprintf("%2d", i+1)
		path := color.CyanString(entry.Path)

		var errorColor *color.Color
		if entry.ErrorRate > 10 {
			errorColor = color.New(color.FgRed)
		} else if entry.ErrorRate > 5 {
			errorColor = color.New(color.FgYellow)
		} else {
			errorColor = color.New(color.FgGreen)
		}

		// Show top misspellings
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
			if i >= 3 { // Show top 3 misspellings
				break
			}
			topMisspellings = append(topMisspellings, fmt.Sprintf("%s(%d)", wc.word, wc.count))
		}

		misspellingsStr := ""
		if len(topMisspellings) > 0 {
			misspellingsStr = fmt.Sprintf(" [%s]", strings.Join(topMisspellings, ", "))
		}

		fmt.Printf("    %s. %s – %s error rate (%d/%d words)%s\n",
			rank, path,
			errorColor.Sprintf("%.1f%%", entry.ErrorRate),
			entry.MisspelledWords, entry.TotalWords,
			color.New(color.FgHiBlack).Sprint(misspellingsStr))
	}

	// Show example issues with authors
	if len(entries) > 0 && len(entries[0].Issues) > 0 {
		fmt.Printf("\n  %s Examples from %s:\n",
			color.YellowString("🔍"),
			color.CyanString(entries[0].Path))

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
				authorStr = fmt.Sprintf(" (by %s)", color.BlueString(issue.Author))
			}

			fmt.Printf("    Line %d: '%s' in %s%s%s\n",
				issue.Line,
				color.RedString(issue.Word),
				issue.Type,
				authorStr,
				color.GreenString(suggestionStr))
		}
	}
}

func generateSpellCheckAuthorLeaderboard(authorStats map[string]*types.SpellCheckAuthorStats, topN int) {
	fmt.Printf("\n  %s\n", color.New(color.Bold).Sprint("👤 Authors with Most Spelling Errors:"))

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
		rank := fmt.Sprintf("%2d", i+1)
		name := color.RedString(entry.Name)
		email := color.New(color.FgHiBlack).Sprintf("(%s)", entry.Email)

		topMistakeStr := ""
		if entry.TopMistake != "" {
			topMistakeStr = fmt.Sprintf(", top mistake: %s(%d)",
				color.YellowString(entry.TopMistake), entry.TopMistakeCount)
		}

		fmt.Printf("    %s. %s %s – %d errors in %d files%s\n",
			rank, name, email, entry.TotalErrors, entry.Files, topMistakeStr)
	}
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
		return color.GreenString("+%d", net)
	} else if net < 0 {
		return color.RedString("%d", net)
	}
	return color.New(color.FgHiBlack).Sprint("0")
}

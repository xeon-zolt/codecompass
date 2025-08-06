package history

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"codecompass/internal/types"
)

// WriteLeaderboardToCSV writes a generic leaderboard to a CSV file.
func WriteLeaderboardToCSV(dir, filename string, header []string, data [][]string) error {
	if dir == "" {
		return fmt.Errorf("log directory not specified")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}

	filePath := filepath.Join(dir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file %s: %w", filePath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// WriteAuthorLeaderboardCSV writes the author leaderboard to a CSV file.
func WriteAuthorLeaderboardCSV(dir string, entries []types.LeaderboardEntry) error {
	filename := fmt.Sprintf("author_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Name", "Email", "Issues", "Errors", "Warnings", "Files", "TopRule", "TopRuleCount"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Name,
			entry.Email,
			fmt.Sprintf("%d", entry.Count),
			fmt.Sprintf("%d", entry.Errors),
			fmt.Sprintf("%d", entry.Warnings),
			fmt.Sprintf("%d", entry.Files),
			entry.TopRule,
			fmt.Sprintf("%d", entry.TopCount),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteFileLeaderboardCSV writes the file leaderboard to a CSV file.
func WriteFileLeaderboardCSV(dir string, entries []types.FileLeaderboardEntry) error {
	filename := fmt.Sprintf("file_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Path", "Issues", "Authors", "TopRule", "TopRuleCount"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Path,
			fmt.Sprintf("%d", entry.Count),
			fmt.Sprintf("%d", entry.Authors),
			entry.TopRule,
			fmt.Sprintf("%d", entry.TopCount),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteRuleLeaderboardCSV writes the rule leaderboard to a CSV file.
func WriteRuleLeaderboardCSV(dir string, entries []types.RuleLeaderboardEntry) error {
	filename := fmt.Sprintf("rule_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Rule", "Violations", "Authors", "Files"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Rule,
			fmt.Sprintf("%d", entry.Count),
			fmt.Sprintf("%d", entry.Authors),
			fmt.Sprintf("%d", entry.Files),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteLinesOfCodeLeaderboardCSV writes the lines of code leaderboard to a CSV file.
func WriteLinesOfCodeLeaderboardCSV(dir string, entries []types.LinesOfCodeEntry) error {
	filename := fmt.Sprintf("loc_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Path", "Lines", "Size"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Path,
			fmt.Sprintf("%d", entry.Lines),
			fmt.Sprintf("%d", entry.Size),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteCommitCountLeaderboardCSV writes the commit count leaderboard to a CSV file.
func WriteCommitCountLeaderboardCSV(dir string, entries []types.CommitCountEntry) error {
	filename := fmt.Sprintf("commit_count_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Name", "Email", "Commits", "FirstCommit", "LastCommit"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Name,
			entry.Email,
			fmt.Sprintf("%d", entry.Commits),
			entry.FirstCommit.Format("2006-01-02"),
			entry.LastCommit.Format("2006-01-02"),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteRecentContributorsLeaderboardCSV writes the recent contributors leaderboard to a CSV file.
func WriteRecentContributorsLeaderboardCSV(dir string, entries []types.RecentContributorEntry) error {
	filename := fmt.Sprintf("recent_contributors_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Name", "Email", "RecentCommits", "LastCommit"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Name,
			entry.Email,
			fmt.Sprintf("%d", entry.RecentCommits),
			entry.LastCommit.Format("2006-01-02"),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteCodeCoverageLeaderboardCSV writes the code coverage leaderboard to a CSV file.
func WriteCodeCoverageLeaderboardCSV(dir string, entries []types.CoverageEntry) error {
	filename := fmt.Sprintf("coverage_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Path", "LinesCovered", "LinesTotal", "CoveragePercent", "FunctionsCovered", "FunctionsTotal", "BranchesCovered", "BranchesTotal"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Path,
			fmt.Sprintf("%d", entry.LinesCovered),
			fmt.Sprintf("%d", entry.LinesTotal),
			fmt.Sprintf("%.2f", entry.CoveragePercent),
			fmt.Sprintf("%d", entry.FunctionsCovered),
			fmt.Sprintf("%d", entry.FunctionsTotal),
			fmt.Sprintf("%d", entry.BranchesCovered),
			fmt.Sprintf("%d", entry.BranchesTotal),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteCodeChurnLeaderboardCSV writes the code churn leaderboard to a CSV file.
func WriteCodeChurnLeaderboardCSV(dir string, entries []types.ChurnEntry) error {
	filename := fmt.Sprintf("churn_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Path", "Changes", "AddedLines", "DeletedLines", "NetLines"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Path,
			fmt.Sprintf("%d", entry.Changes),
			fmt.Sprintf("%d", entry.AddedLines),
			fmt.Sprintf("%d", entry.DeletedLines),
			fmt.Sprintf("%d", entry.NetLines),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteBugDensityLeaderboardCSV writes the bug density leaderboard to a CSV file.
func WriteBugDensityLeaderboardCSV(dir string, entries []types.BugDensityEntry) error {
	filename := fmt.Sprintf("bug_density_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Path", "BugFixes", "TotalCommits", "BugRatio"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Path,
			fmt.Sprintf("%d", entry.BugFixes),
			fmt.Sprintf("%d", entry.TotalCommits),
			fmt.Sprintf("%.4f", entry.BugRatio),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteTechnicalDebtLeaderboardCSV writes the technical debt leaderboard to a CSV file.
func WriteTechnicalDebtLeaderboardCSV(dir string, entries []types.TechnicalDebtEntry) error {
	filename := fmt.Sprintf("technical_debt_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Path", "TodoCount", "FixmeCount", "HackCount", "TotalDebt"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Path,
			fmt.Sprintf("%d", entry.TodoCount),
			fmt.Sprintf("%d", entry.FixmeCount),
			fmt.Sprintf("%d", entry.HackCount),
			fmt.Sprintf("%d", entry.TotalDebt),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

// WriteSpellCheckLeaderboardCSV writes the spell check leaderboard to a CSV file.
func WriteSpellCheckLeaderboardCSV(dir string, entries []types.SpellCheckEntry) error {
	filename := fmt.Sprintf("spell_check_leaderboard_%s.csv", time.Now().Format("20060102_150405"))
	header := []string{"Rank", "Path", "MisspelledWords", "TotalWords", "ErrorRate"}
	data := make([][]string, len(entries))
	for i, entry := range entries {
		data[i] = []string{
			fmt.Sprintf("%d", entry.Rank),
			entry.Path,
			fmt.Sprintf("%d", entry.MisspelledWords),
			fmt.Sprintf("%d", entry.TotalWords),
			fmt.Sprintf("%.4f", entry.ErrorRate),
		}
	}
	return WriteLeaderboardToCSV(dir, filename, header, data)
}

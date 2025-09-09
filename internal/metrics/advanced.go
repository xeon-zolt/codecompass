package metrics

import (
	"bufio"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"codecompass/internal/utils"
)

type ChurnEntry struct {
	Rank         int
	Path         string
	Changes      int
	AddedLines   int
	DeletedLines int
	NetLines     int
}

type BugDensityEntry struct {
	Rank         int
	Path         string
	BugFixes     int
	TotalCommits int
	BugRatio     float64
}

type TechnicalDebtEntry struct {
	Rank       int
	Path       string
	TodoCount  int
	FixmeCount int
	HackCount  int
	TotalDebt  int
}

func GetCodeChurnLeaderboard(trackedFiles map[string]bool) ([]ChurnEntry, error) {
	cmd := exec.Command("git", "log", "--numstat", "--pretty=format:")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	churnData := make(map[string]*ChurnEntry)

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

		added, err := strconv.Atoi(parts[0])
		if err != nil {
			continue // Skip line if we can't parse added count
		}
		deleted, err := strconv.Atoi(parts[1])
		if err != nil {
			continue // Skip line if we can't parse deleted count
		}
		filePath := parts[2]

		if !trackedFiles[filePath] {
			continue
		}

		if churnData[filePath] == nil {
			churnData[filePath] = &ChurnEntry{Path: filePath}
		}

		entry := churnData[filePath]
		entry.Changes++
		entry.AddedLines += added
		entry.DeletedLines += deleted
		entry.NetLines += added - deleted
	}

	var entries []ChurnEntry
	for _, entry := range churnData {
		entries = append(entries, *entry)
	}

	return entries, nil
}

func GetBugDensityLeaderboard(trackedFiles map[string]bool) ([]BugDensityEntry, error) {
	// Get all commits
	cmd := exec.Command("git", "log", "--name-only", "--pretty=format:%H|%s")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	bugRegex := regexp.MustCompile(`(?i)(fix|bug|issue|error|broken|crash)`)

	fileStats := make(map[string]*BugDensityEntry)

	lines := strings.Split(string(output), "\n")
	var currentCommitIsBug bool
	var currentFiles []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "|") {
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

		// Process files when we hit the next commit or end
		if strings.Contains(line, "|") && len(currentFiles) > 0 {
			for _, filePath := range currentFiles {
				if fileStats[filePath] == nil {
					fileStats[filePath] = &BugDensityEntry{Path: filePath}
				}

				entry := fileStats[filePath]
				entry.TotalCommits++
				if currentCommitIsBug {
					entry.BugFixes++
				}
			}
		}
	}

	var entries []BugDensityEntry
	for _, entry := range fileStats {
		if entry.TotalCommits > 0 {
			entry.BugRatio = float64(entry.BugFixes) / float64(entry.TotalCommits) * 100
		}
		entries = append(entries, *entry)
	}

	return entries, nil
}

func GetTechnicalDebtLeaderboard(trackedFiles map[string]bool) ([]TechnicalDebtEntry, error) {
	scanner := utils.NewTechnicalDebtScanner()
	entries, err := scanner.ScanFiles(trackedFiles)
	if err != nil {
		return nil, err
	}

	// Convert from utils types to metrics types
	var result []TechnicalDebtEntry
	for _, entry := range entries {
		result = append(result, TechnicalDebtEntry{
			Path:       entry.Path,
			TodoCount:  entry.TodoCount,
			FixmeCount: entry.FixmeCount,
			HackCount:  entry.HackCount,
			TotalDebt:  entry.TotalDebt,
		})
	}

	return result, nil
}

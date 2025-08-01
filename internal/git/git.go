package git

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"codecompass/internal/types"
	"codecompass/internal/utils"
)

var (
	blameCache    = make(map[string]map[int]types.BlameInfo)
	blameFailures = make(map[string]bool)
	cacheMutex    sync.Mutex
)

func ValidateRepository() error {
	_, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	return err
}

func GetTrackedFiles() (map[string]bool, error) {
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := make(map[string]bool)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line != "" {
			files[line] = true
		}
	}
	return files, nil
}

func GetFileLineCount(filePath string) (int, error) {
	cmd := exec.Command("wc", "-l", filePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return 0, fmt.Errorf("unable to parse line count")
	}

	return strconv.Atoi(parts[0])
}

func GetCommitHistory() ([]types.CommitInfo, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%H|%an|%ae|%at|%s", "--all")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []types.CommitInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 5 {
			continue
		}

		timestamp, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			continue
		}

		commits = append(commits, types.CommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    time.Unix(timestamp, 0),
			Message: parts[4],
		})
	}

	return commits, nil
}

func GetAuthorCommitCounts() (map[string]types.CommitCountEntry, error) {
	commits, err := GetCommitHistory()
	if err != nil {
		return nil, err
	}

	authorStats := make(map[string]types.CommitCountEntry)

	for _, commit := range commits {
		entry, exists := authorStats[commit.Email]
		if !exists {
			entry = types.CommitCountEntry{
				Name:        commit.Author,
				Email:       commit.Email,
				Commits:     0,
				FirstCommit: commit.Date,
				LastCommit:  commit.Date,
			}
		}

		entry.Commits++

		if commit.Date.Before(entry.FirstCommit) {
			entry.FirstCommit = commit.Date
		}
		if commit.Date.After(entry.LastCommit) {
			entry.LastCommit = commit.Date
		}

		entry.Name = commit.Author // Update to latest name
		authorStats[commit.Email] = entry
	}

	return authorStats, nil
}

func GetRecentContributors(days int) (map[string]types.RecentContributorEntry, error) {
	since := time.Now().AddDate(0, 0, -days)
	sinceStr := since.Format("2006-01-02")

	cmd := exec.Command("git", "log", "--since="+sinceStr, "--pretty=format:%an|%ae|%at")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	authorStats := make(map[string]types.RecentContributorEntry)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		timestamp, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}

		commitDate := time.Unix(timestamp, 0)
		email := parts[1]

		entry, exists := authorStats[email]
		if !exists {
			entry = types.RecentContributorEntry{
				Name:          parts[0],
				Email:         email,
				LastCommit:    commitDate,
				RecentCommits: 0,
			}
		}

		entry.RecentCommits++

		if commitDate.After(entry.LastCommit) {
			entry.LastCommit = commitDate
		}

		entry.Name = parts[0] // Update to latest name
		authorStats[email] = entry
	}

	return authorStats, nil
}

func BlameFile(filePath string, warningLogs *[]string, mu *sync.Mutex, semaphore *utils.Semaphore) (map[int]types.BlameInfo, error) {
	cacheMutex.Lock()
	if blameMap, exists := blameCache[filePath]; exists {
		cacheMutex.Unlock()
		return blameMap, nil
	}
	if blameFailures[filePath] {
		cacheMutex.Unlock()
		return make(map[int]types.BlameInfo), fmt.Errorf("file already failed")
	}
	cacheMutex.Unlock()

	semaphore.Acquire()
	defer semaphore.Release()

	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "blame", "--line-porcelain", "--", normalizedPath)
	output, err := cmd.Output()
	if err != nil {
		cacheMutex.Lock()
		blameFailures[filePath] = true
		cacheMutex.Unlock()

		mu.Lock()
		*warningLogs = append(*warningLogs, fmt.Sprintf("⚠️ Blame failed for %s: %v", filePath, err))
		mu.Unlock()

		return make(map[int]types.BlameInfo), err
	}

	blameMap := parseBlameOutput(string(output))

	cacheMutex.Lock()
	blameCache[filePath] = blameMap
	cacheMutex.Unlock()

	return blameMap, nil
}

func parseBlameOutput(output string) map[int]types.BlameInfo {
	blameMap := make(map[int]types.BlameInfo)
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentEmail, currentName string
	var currentLine int
	commitRegex := regexp.MustCompile(`^[0-9a-f]{40} `)

	for scanner.Scan() {
		line := scanner.Text()

		if commitRegex.MatchString(line) {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				if lineNum, err := strconv.Atoi(parts[2]); err == nil {
					currentLine = lineNum
				}
			}
		} else if strings.HasPrefix(line, "author ") {
			currentName = strings.TrimSpace(strings.TrimPrefix(line, "author "))
		} else if strings.HasPrefix(line, "author-mail ") {
			email := strings.TrimPrefix(line, "author-mail ")
			email = strings.Trim(email, "<>")
			currentEmail = strings.TrimSpace(email)
		} else if strings.HasPrefix(line, "\t") {
			if currentEmail != "" && currentLine > 0 {
				blameMap[currentLine] = types.BlameInfo{
					Email: currentEmail,
					Name:  currentName,
				}
			}
			currentEmail = ""
			currentName = ""
			currentLine = 0
		}
	}

	return blameMap
}

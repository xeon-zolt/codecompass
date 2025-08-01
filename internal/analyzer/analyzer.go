package analyzer

import (
	"sort"
	"sync"
	"time"

	"codecompass/internal/config"
	"codecompass/internal/git"
	"codecompass/internal/types"
	"codecompass/internal/utils"
)

type Analyzer struct {
	semaphore *utils.Semaphore
	mu        *sync.Mutex
}

func New(semaphore *utils.Semaphore, mu *sync.Mutex) *Analyzer {
	return &Analyzer{
		semaphore: semaphore,
		mu:        mu,
	}
}

func (a *Analyzer) ProcessIssue(
	issue types.Issue,
	authorStats map[string]*types.AuthorStats,
	fileStats map[string]*types.FileStats,
	ruleStats map[string]*types.RuleStats,
	warningLogs *[]string,
) error {
	return a.processIssueInternal(issue, nil, authorStats, fileStats, ruleStats, warningLogs)
}

func (a *Analyzer) ProcessIssueWithConfig(
	issue types.Issue,
	cfg *config.Config,
	authorStats map[string]*types.AuthorStats,
	fileStats map[string]*types.FileStats,
	ruleStats map[string]*types.RuleStats,
	warningLogs *[]string,
) error {
	// Check if file should be ignored
	if cfg != nil && cfg.ShouldIgnoreFile(issue.FilePath) {
		return nil
	}

	// Check if rule should be ignored
	if cfg != nil && cfg.ShouldIgnoreRule(issue.RuleID) {
		return nil
	}

	return a.processIssueInternal(issue, cfg, authorStats, fileStats, ruleStats, warningLogs)
}

func (a *Analyzer) processIssueInternal(
	issue types.Issue,
	cfg *config.Config,
	authorStats map[string]*types.AuthorStats,
	fileStats map[string]*types.FileStats,
	ruleStats map[string]*types.RuleStats,
	warningLogs *[]string,
) error {
	blameMap, err := git.BlameFile(issue.FilePath, warningLogs, a.mu, a.semaphore)
	if err != nil {
		return err
	}

	if len(blameMap) == 0 {
		return nil
	}

	// Find nearest line
	var lines []int
	for line := range blameMap {
		lines = append(lines, line)
	}
	sort.Ints(lines)

	var nearestLine int
	for _, line := range lines {
		if line >= issue.Line {
			nearestLine = line
			break
		}
	}
	if nearestLine == 0 && len(lines) > 0 {
		nearestLine = lines[len(lines)-1]
	}

	blameInfo, exists := blameMap[nearestLine]
	if !exists {
		return nil
	}

	// Check if author should be ignored
	if cfg != nil && cfg.ShouldIgnoreAuthor(blameInfo.Email, blameInfo.Name) {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()

	// Update author stats
	if authorStats[blameInfo.Email] == nil {
		authorStats[blameInfo.Email] = &types.AuthorStats{
			Name:      blameInfo.Name,
			Count:     0,
			Rules:     make(map[string]int),
			Files:     make(map[string]int),
			FirstSeen: now,
			LastSeen:  now,
		}
	}

	stats := authorStats[blameInfo.Email]
	stats.Name = blameInfo.Name
	stats.Count++
	stats.Rules[issue.RuleID]++
	stats.Files[issue.FilePath]++
	if issue.Severity == 2 {
		stats.Errors++
	} else {
		stats.Warnings++
	}
	if now.After(stats.LastSeen) {
		stats.LastSeen = now
	}

	// Update file stats
	if fileStats[issue.FilePath] == nil {
		fileStats[issue.FilePath] = &types.FileStats{
			Path:    issue.FilePath,
			Count:   0,
			Rules:   make(map[string]int),
			Authors: make(map[string]int),
		}
	}
	fileStats[issue.FilePath].Count++
	fileStats[issue.FilePath].Rules[issue.RuleID]++
	fileStats[issue.FilePath].Authors[blameInfo.Email]++

	// Update rule stats
	if ruleStats[issue.RuleID] == nil {
		ruleStats[issue.RuleID] = &types.RuleStats{
			Rule:    issue.RuleID,
			Count:   0,
			Authors: make(map[string]int),
			Files:   make(map[string]int),
		}
	}
	ruleStats[issue.RuleID].Count++
	ruleStats[issue.RuleID].Authors[blameInfo.Email]++
	ruleStats[issue.RuleID].Files[issue.FilePath]++

	return nil
}

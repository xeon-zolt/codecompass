package types

import "time"

type AuthorStats struct {
	Name      string
	Count     int
	Rules     map[string]int
	Files     map[string]int
	Errors    int
	Warnings  int
	FirstSeen time.Time
	LastSeen  time.Time
}

type FileStats struct {
	Path    string
	Count   int
	Rules   map[string]int
	Authors map[string]int
}

type RuleStats struct {
	Rule    string
	Count   int
	Authors map[string]int
	Files   map[string]int
}

// Existing leaderboard entries
type LeaderboardEntry struct {
	Rank     int
	Name     string
	Email    string
	Count    int
	TopRule  string
	TopCount int
	Files    int
	Errors   int
	Warnings int
}

type FileLeaderboardEntry struct {
	Rank     int
	Path     string
	Count    int
	TopRule  string
	TopCount int
	Authors  int
}

type RuleLeaderboardEntry struct {
	Rank    int
	Rule    string
	Count   int
	Authors int
	Files   int
}

// New leaderboard entries
type LinesOfCodeEntry struct {
	Rank  int
	Path  string
	Lines int
	Size  int64 // File size in bytes
}

type CommitCountEntry struct {
	Rank        int
	Name        string
	Email       string
	Commits     int
	FirstCommit time.Time
	LastCommit  time.Time
}

type RecentContributorEntry struct {
	Rank          int
	Name          string
	Email         string
	LastCommit    time.Time
	RecentCommits int // Commits in last 30 days
}

// Git commit info
type CommitInfo struct {
	Hash    string
	Author  string
	Email   string
	Date    time.Time
	Message string
}

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

// coverage types
type CoverageEntry struct {
	Rank             int
	Path             string
	LinesCovered     int
	LinesTotal       int
	CoveragePercent  float64
	FunctionsCovered int
	FunctionsTotal   int
	BranchesCovered  int
	BranchesTotal    int
}

type CoverageData struct {
	Files map[string]FileCoverage
}

type FileCoverage struct {
	Path             string
	LinesCovered     int
	LinesTotal       int
	FunctionsCovered int
	FunctionsTotal   int
	BranchesCovered  int
	BranchesTotal    int
}

// Spell check types
type SpellCheckAuthorStats struct {
	Name           string
	Email          string
	TotalErrors    int
	Files          map[string]int // filename -> error count
	CommonMistakes map[string]int // word -> count
}

type SpellIssue struct {
	Word        string
	Line        int
	Column      int
	Context     string
	Type        string // "comment", "string", "identifier"
	Suggestions []string
	Author      string
	AuthorEmail string
}

// Update existing SpellCheckEntry if needed
type SpellCheckEntry struct {
	Rank            int
	Path            string
	MisspelledWords int
	TotalWords      int
	ErrorRate       float64
	TopMisspellings map[string]int
	Issues          []SpellIssue
}

package leaderboard

import (
	"testing"

	"codecompass/internal/types"
)

func TestGenerateAuthorLeaderboard(t *testing.T) {
	authorStats := map[string]*types.AuthorStats{
		"test@example.com": {
			Name:  "Test User",
			Count: 10,
			Rules: map[string]int{"no-console": 10},
		},
	}

	entries := GenerateAuthorLeaderboard(authorStats, 10)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, but got %d", len(entries))
	}

	entry := entries[0]
	if entry.Name != "Test User" {
		t.Errorf("Expected name to be 'Test User', but got '%s'", entry.Name)
	}

	if entry.Count != 10 {
		t.Errorf("Expected count to be 10, but got %d", entry.Count)
	}

	if entry.TopRule != "no-console" {
		t.Errorf("Expected top rule to be 'no-console', but got '%s'", entry.TopRule)
	}
}

func TestGenerateFileLeaderboard(t *testing.T) {
	fileStats := map[string]*types.FileStats{
		"test.go": {
			Path:  "test.go",
			Count: 10,
			Rules: map[string]int{"no-console": 10},
		},
	}

	entries := GenerateFileLeaderboard(fileStats, 10)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, but got %d", len(entries))
	}

	entry := entries[0]
	if entry.Path != "test.go" {
		t.Errorf("Expected path to be 'test.go', but got '%s'", entry.Path)
	}

	if entry.Count != 10 {
		t.Errorf("Expected count to be 10, but got %d", entry.Count)
	}

	if entry.TopRule != "no-console" {
		t.Errorf("Expected top rule to be 'no-console', but got '%s'", entry.TopRule)
	}
}

func TestGenerateRuleLeaderboard(t *testing.T) {
	ruleStats := map[string]*types.RuleStats{
		"no-console": {
			Rule:  "no-console",
			Count: 10,
		},
	}

	entries := GenerateRuleLeaderboard(ruleStats, 10)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, but got %d", len(entries))
	}

	entry := entries[0]
	if entry.Rule != "no-console" {
		t.Errorf("Expected rule to be 'no-console', but got '%s'", entry.Rule)
	}

	if entry.Count != 10 {
		t.Errorf("Expected count to be 10, but got %d", entry.Count)
	}
}

package analyzer

import (
	"os"
	"os/exec"
	"sync"
	"testing"

	"codecompass/internal/config"
	"codecompass/internal/types"
	"codecompass/internal/utils"
)

func TestProcessIssueWithConfig(t *testing.T) {
	// Create a temporary directory
	tmpdir, err := os.MkdirTemp("", "analyzer_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Change to the temporary directory
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldwd)
	os.Chdir(tmpdir)

	// Initialize a git repository
	if err := exec.Command("git", "init").Run(); err != nil {
		t.Fatal(err)
	}

	// Create a file and commit it
	content := []byte("test file")
	if err := os.WriteFile("test.go", content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := exec.Command("git", "add", "test.go").Run(); err != nil {
		t.Fatal(err)
	}

	if err := exec.Command("git", "commit", "-m", "initial commit").Run(); err != nil {
		t.Fatal(err)
	}

	// Create a new analyzer
	semaphore := utils.NewSemaphore(1)
	mu := &sync.Mutex{}
	analyzer := New(semaphore, mu)

	// Create a new config
	cfg := config.NewConfig()

	// Create a new issue
	issue := types.Issue{
		FilePath: "test.go",
		Line:     1,
		RuleID:   "test-rule",
		Severity: 2,
	}

	// Create mock stats
	authorStats := make(map[string]*types.AuthorStats)
	fileStats := make(map[string]*types.FileStats)
	ruleStats := make(map[string]*types.RuleStats)
	warningLogs := make([]string, 0)

	// Process the issue
	err = analyzer.ProcessIssueWithConfig(issue, cfg, authorStats, fileStats, ruleStats, &warningLogs)
	if err != nil {
		t.Errorf("Error processing issue: %v", err)
	}

	// Check if the stats were updated
	if len(authorStats) != 1 {
		t.Errorf("Expected 1 author stat, but got %d", len(authorStats))
	}

	if len(fileStats) != 1 {
		t.Errorf("Expected 1 file stat, but got %d", len(fileStats))
	}

	if len(ruleStats) != 1 {
		t.Errorf("Expected 1 rule stat, but got %d", len(ruleStats))
	}
}

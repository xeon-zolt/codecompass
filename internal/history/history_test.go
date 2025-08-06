package history

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codecompass/internal/types"
)

func TestWriteAuthorLeaderboardCSV(t *testing.T) {
	// Create a temporary directory for test output
	tmpDir, err := ioutil.TempDir("", "test_history")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Sample data
	entries := []types.LeaderboardEntry{
		{
			Rank:     1,
			Name:     "John Doe",
			Email:    "john@example.com",
			Count:    100,
			Errors:   50,
			Warnings: 50,
			Files:    10,
			TopRule:  "no-unused-vars",
			TopCount: 20,
		},
		{
			Rank:     2,
			Name:     "Jane Smith",
			Email:    "jane@example.com",
			Count:    80,
			Errors:   30,
			Warnings: 50,
			Files:    8,
			TopRule:  "indent",
			TopCount: 15,
		},
	}

	// Write to CSV
	err = WriteAuthorLeaderboardCSV(tmpDir, entries)
	if err != nil {
		t.Fatalf("WriteAuthorLeaderboardCSV failed: %v", err)
	}

	// Construct expected file name pattern
	// The filename includes a timestamp, so we'll check for a prefix and suffix
	prefix := "author_leaderboard_"
	suffix := ".csv"

	files, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file in temp dir, got %d", len(files))
	}

	generatedFileName := files[0].Name()
	if !strings.HasPrefix(generatedFileName, prefix) || !strings.HasSuffix(generatedFileName, suffix) {
		t.Fatalf("Generated filename %s does not match expected pattern %s*%s", generatedFileName, prefix, suffix)
	}

	// Read the content of the generated CSV file
	filePath := filepath.Join(tmpDir, generatedFileName)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read generated CSV file: %v", err)
	}

	// Expected CSV content (excluding the dynamic timestamp in filename)
	expectedContent := "Rank,Name,Email,Issues,Errors,Warnings,Files,TopRule,TopRuleCount\n" +
		"1,John Doe,john@example.com,100,50,50,10,no-unused-vars,20\n" +
		"2,Jane Smith,jane@example.com,80,30,50,8,indent,15\n"

	if string(content) != expectedContent {
		t.Errorf("CSV content mismatch:\nExpected:\n%s\nGot:\n%s", expectedContent, string(content))
	}
}

func TestWriteLeaderboardToCSV_ErrorHandling(t *testing.T) {
	// Test case: empty directory path
	err := WriteLeaderboardToCSV("", "test.csv", []string{"Header"}, [][]string{{"Data"}})
	if err == nil || !strings.Contains(err.Error(), "log directory not specified") {
		t.Errorf("Expected 'log directory not specified' error, got: %v", err)
	}

	// Test case: invalid directory path (e.g., permissions issue, though hard to simulate reliably)
	// For now, we'll just test a path that's clearly not writable or creatable in a typical setup
	err = WriteLeaderboardToCSV("/root/nonexistent/path", "test.csv", []string{"Header"}, [][]string{{"Data"}})
	if err == nil || !strings.Contains(err.Error(), "failed to create log directory") {
		t.Errorf("Expected 'failed to create log directory' error, got: %v", err)
	}
}

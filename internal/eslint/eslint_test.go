package eslint

import (
	"testing"
)

func TestRunESLint(t *testing.T) {
	// Mock ESLint JSON output
	mockESLintOutput := `[
    {
        "filePath": "test.js",
        "messages": [
            {
                "ruleId": "no-console",
                "severity": 2,
                "message": "Unexpected console statement.",
                "line": 1,
                "column": 1
            }
        ]
    }
]`

	// Override runCommand for testing
	oldRunCommand := runCommand
	defer func() {
		runCommand = oldRunCommand
	}()
	runCommand = func(name string, arg ...string) ([]byte, error) {
		return []byte(mockESLintOutput), nil
	}

	// Create a dummy file path for testing purposes
	trackedFiles := map[string]bool{"test.js": true}
	issues, err := RunESLint(trackedFiles, []string{})
	if err != nil {
		t.Fatalf("RunESLint failed: %v", err)
	}

	// Check the results
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, but got %d", len(issues))
	}

	issue := issues[0]
	if issue.RuleID != "no-console" {
		t.Errorf("Expected rule ID to be 'no-console', but got '%s'", issue.RuleID)
	}

	if issue.Line != 1 {
		t.Errorf("Expected line number to be 1, but got %d", issue.Line)
	}

	if issue.FilePath != "test.js" {
		t.Errorf("Expected file path to be 'test.js', but got '%s'", issue.FilePath)
	}
}
package ruff

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"codecompass/internal/types"
)

// RuffIssue represents a single issue reported by Ruff.
type RuffIssue struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Location struct {
		Row    int `json:"row"`
		Column int `json:"column"`
	} `json:"location"`
	Filename string `json:"filename"`
	Fix      *struct{} `json:"fix"` // We don't care about the fix for now
}

// RunRuff executes the ruff linter and parses its JSON output.
func RunRuff(files []string, ruffRules []string, ruffIgnorePaths []string) ([]types.Issue, error) {
	args := []string{"check", "--output-format=json"}

	// Add rules if specified
	if len(ruffRules) > 0 {
		args = append(args, fmt.Sprintf("--select=%s", strings.Join(ruffRules, ",")))
	}

	// Add ignore paths if specified
	if len(ruffIgnorePaths) > 0 {
		args = append(args, fmt.Sprintf("--ignore=%s", strings.Join(ruffIgnorePaths, ",")))
	}

	// Add files to check
	args = append(args, files...)

	cmd := exec.Command("ruff", args...)
	output, err := cmd.Output()

	if err != nil {
		// Ruff returns non-zero exit code if issues are found, which is not an error for us.
		if exitError, ok := err.(*exec.ExitError); ok {
			output = exitError.Stderr // Ruff prints JSON to stdout even on non-zero exit
		} else {
			return nil, fmt.Errorf("failed to run ruff: %w", err)
		}
	}

	var ruffIssues []RuffIssue
	if err := json.Unmarshal(output, &ruffIssues); err != nil {
		return nil, fmt.Errorf("failed to parse ruff output: %w", err)
	}

	var issues []types.Issue
	for _, ruffIssue := range ruffIssues {
		// Convert RuffIssue to CodeCompass's generic Issue format
		issues = append(issues, types.Issue{
			FilePath: filepath.ToSlash(ruffIssue.Filename),
			Line:     ruffIssue.Location.Row,
			RuleID:   ruffIssue.Code,
			Message:  ruffIssue.Message,
			Severity: 1, // Ruff issues are typically errors/warnings, map to a default severity
		})
	}

	return issues, nil
}

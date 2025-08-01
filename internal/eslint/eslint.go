package eslint

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"codecompass/internal/types"
)

func RunESLint(trackedFiles map[string]bool, ignoredRules []string) ([]types.Issue, error) {
	cmd := exec.Command("npx", "eslint", ".", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			output = exitError.Stderr
			if len(output) == 0 {
				cmd = exec.Command("npx", "eslint", ".", "--format", "json")
				output, _ = cmd.Output()
			}
		}
	}

	var results []types.ESLintResult
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse ESLint output: %v", err)
	}

	var issues []types.Issue
	ignoredRulesMap := make(map[string]bool)
	for _, rule := range ignoredRules {
		ignoredRulesMap[rule] = true
	}

	cwd, _ := os.Getwd()

	for _, result := range results {
		relPath, err := filepath.Rel(cwd, result.FilePath)
		if err != nil {
			relPath = result.FilePath
		}

		if !trackedFiles[relPath] {
			continue
		}

		for _, message := range result.Messages {
			if ignoredRulesMap[message.RuleID] {
				continue
			}

			issues = append(issues, types.Issue{
				FilePath: relPath,
				Line:     message.Line,
				RuleID:   message.RuleID,
				Severity: message.Severity,
			})
		}
	}

	return issues, nil
}

package eslint

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"codecompass/internal/types"
)

// isESLintAvailable checks if ESLint is available in the system
func isESLintAvailable() bool {
	// Check if npx eslint is available
	_, err := exec.LookPath("npx")
	if err != nil {
		// If npx is not available, check for eslint directly
		_, err := exec.LookPath("eslint")
		return err == nil
	}
	
	// Check if eslint can be run via npx and has proper configuration
	cmd := exec.Command("npx", "eslint", "--version")
	err = cmd.Run()
	if err != nil {
		return false
	}
	
	// Check if ESLint can run on current directory (test for configuration)
	cmd = exec.Command("npx", "eslint", ".", "--format", "json")
	err = cmd.Run()
	// If ESLint exits with code 2, it usually means configuration error
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode() != 2
	}
	
	return err == nil
}

// runCommand is a variable that can be overridden for testing purposes.
var runCommand = func(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	return cmd.Output()
}

func RunESLint(trackedFiles map[string]bool, ignoredRules []string) ([]types.Issue, error) {
	// Check if there are any JavaScript/TypeScript files to lint
	hasJSFiles := false
	for file := range trackedFiles {
		if filepath.Ext(file) == ".js" || filepath.Ext(file) == ".jsx" || 
		   filepath.Ext(file) == ".ts" || filepath.Ext(file) == ".tsx" ||
		   filepath.Ext(file) == ".mjs" || filepath.Ext(file) == ".cjs" {
			hasJSFiles = true
			break
		}
	}
	
	if !hasJSFiles {
		return []types.Issue{}, nil // Return empty results, no JS/TS files to lint
	}

	// First check if ESLint is available
	if !isESLintAvailable() {
		return nil, fmt.Errorf("ESLint not found. Please install ESLint:\n  • Locally: npm install eslint\n  • Globally: npm install -g eslint\n  • Or ensure npx is available")
	}

	output, err := runCommand("npx", "eslint", ".", "--format", "json")
	
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// ESLint returns non-zero exit code if issues are found, which is not an error for us.
			// The JSON output is still valid and available in the output variable from cmd.Output()
			// We already have the output, so we can continue processing
		} else {
			return nil, fmt.Errorf("failed to run ESLint: %w", err)
		}
	}

	var results []types.ESLintResult
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse ESLint output: %w", err)
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

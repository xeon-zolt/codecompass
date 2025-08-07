package eslint

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	
)

func TestRunESLint(t *testing.T) {
	// Create a temporary directory
	tmpdir, err := os.MkdirTemp("", "eslint_test")
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

	// Create a package.json file
	packageJSON := `{ "devDependencies": { "eslint": "^8.0.0" } }`
	if err := os.WriteFile("package.json", []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an .eslintrc.js file
	eslintrc := `module.exports = { "rules": { "no-console": "error" } };`
	if err := os.WriteFile(".eslintrc.js", []byte(eslintrc), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a file with a linting error
	jsFile := `console.log("hello");`
	if err := os.WriteFile("test.js", []byte(jsFile), 0644); err != nil {
		t.Fatal(err)
	}

	// Install ESLint
	if err := exec.Command("npm", "install").Run(); err != nil {
		t.Fatal(err)
	}

	// Run ESLint
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

	if filepath.Base(issue.FilePath) != "test.js" {
		t.Errorf("Expected file path to be 'test.js', but got '%s'", issue.FilePath)
	}
}

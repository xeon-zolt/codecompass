package git

import (
	"os"
	"os/exec"
	"testing"
	
)

func TestGetTrackedFiles(t *testing.T) {
	// Create a temporary directory
	tmpdir, err := os.MkdirTemp("", "git_test")
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

	// Get the tracked files
	files, err := GetTrackedFiles()
	if err != nil {
		t.Fatal(err)
	}

	// Check the results
	if len(files) != 1 {
		t.Fatalf("Expected 1 tracked file, but got %d", len(files))
	}

	if !files["test.go"] {
		t.Errorf("Expected test.go to be a tracked file")
	}
}

func TestGetCommitHistory(t *testing.T) {
	// Create a temporary directory
	tmpdir, err := os.MkdirTemp("", "git_test")
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

	// Get the commit history
	commits, err := GetCommitHistory()
	if err != nil {
		t.Fatal(err)
	}

	// Check the results
	if len(commits) != 1 {
		t.Fatalf("Expected 1 commit, but got %d", len(commits))
	}

	commit := commits[0]
	if commit.Message != "initial commit" {
		t.Errorf("Expected commit message to be 'initial commit', but got '%s'", commit.Message)
	}
}

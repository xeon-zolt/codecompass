package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	c := NewConfig()

	if c.MaxFileSize != 5000 {
		t.Errorf("Expected MaxFileSize to be 5000, but got %d", c.MaxFileSize)
	}

	if c.MinCoverageThreshold != 80.0 {
		t.Errorf("Expected MinCoverageThreshold to be 80.0, but got %f", c.MinCoverageThreshold)
	}

	if c.MaxConcurrentBlame != 4 {
		t.Errorf("Expected MaxConcurrentBlame to be 4, but got %d", c.MaxConcurrentBlame)
	}

	if !c.CacheResults {
		t.Errorf("Expected CacheResults to be true, but got false")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	content := []byte("max-file-size = 10000\nignore-files = *.log,*.tmp")
	tmpfile, err := os.CreateTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}

defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Load the config from the temporary file
	c, err := LoadConfigFromFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if c.MaxFileSize != 10000 {
		t.Errorf("Expected MaxFileSize to be 10000, but got %d", c.MaxFileSize)
	}

	if len(c.IgnoredFiles) != 2 {
		t.Errorf("Expected 2 ignored files, but got %d", len(c.IgnoredFiles))
	}

	if c.IgnoredFiles[0] != "*.log" {
		t.Errorf("Expected ignored file to be *.log, but got %s", c.IgnoredFiles[0])
	}

	if c.IgnoredFiles[1] != "*.tmp" {
		t.Errorf("Expected ignored file to be *.tmp, but got %s", c.IgnoredFiles[1])
	}
}

func TestShouldIgnoreFile(t *testing.T) {
	c := NewConfig()
	c.IgnoredFiles = []string{"*.log", "/tmp/*"}
	c.IgnoredPaths = []string{"/tmp/"}

	if !c.ShouldIgnoreFile("test.log") {
		t.Errorf("Expected to ignore test.log")
	}

	if !c.ShouldIgnoreFile("/tmp/test") {
		t.Errorf("Expected to ignore /tmp/test")
	}

	if c.ShouldIgnoreFile("test.txt") {
		t.Errorf("Expected not to ignore test.txt")
	}
}

func TestShouldIgnoreAuthor(t *testing.T) {
	c := NewConfig()
	c.IgnoredAuthors = []string{"test@example.com", "Test User"}

	if !c.ShouldIgnoreAuthor("test@example.com", "") {
		t.Errorf("Expected to ignore test@example.com")
	}

	if !c.ShouldIgnoreAuthor("", "Test User") {
		t.Errorf("Expected to ignore Test User")
	}

	if c.ShouldIgnoreAuthor("another@example.com", "Another User") {
		t.Errorf("Expected not to ignore another@example.com")
	}
}

func TestShouldIgnoreRule(t *testing.T) {
	c := NewConfig()
	c.IgnoredRules = []string{"no-console", "no-debugger"}

	if !c.ShouldIgnoreRule("no-console") {
		t.Errorf("Expected to ignore no-console")
	}

	if c.ShouldIgnoreRule("no-alert") {
		t.Errorf("Expected not to ignore no-alert")
	}
}

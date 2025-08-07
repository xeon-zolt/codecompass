package main

import (
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Run tests
	os.Exit(m.Run())
}

func TestShowVersion(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showVersion()

	w.Close()
	os.Stdout = old

	// Read the output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)

	// Check if the version is printed
	if !contains(string(buf[:n]), VERSION) {
		t.Errorf("Expected version %s to be printed", VERSION)
	}
}

func TestShowCompassArt(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showCompassArt()

	w.Close()
	os.Stdout = old

	// Read the output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)

	// Check if the compass art is printed
	if !contains(string(buf[:n]), "CodeCompass") {
		t.Errorf("Expected compass art to be printed")
	}
}

// Helper function to check if a string contains a substring

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

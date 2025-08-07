package spellcheck

import (
	"testing"

	"codecompass/internal/config"
)

func TestSpellChecker(t *testing.T) {
	cfg := config.NewConfig()
	sc, err := NewSpellChecker(cfg)
	if err != nil {
		t.Fatalf("Failed to create spell checker: %v", err)
	}

	// Test correct words
	correctWords := []string{"hello", "world", "test", "code", "the", "and", "for"}
	for _, word := range correctWords {
		if !sc.IsCorrect(word) {
			t.Errorf("Expected '%s' to be correct", word)
		}
	}

	// Test incorrect words
	incorrectWords := []string{"helo", "wrold", "tset", "coed"}
	for _, word := range incorrectWords {
		if sc.IsCorrect(word) {
			t.Errorf("Expected '%s' to be incorrect", word)
		}
	}

	// Test suggestions
	suggestions := sc.GetSuggestions("helo")
	if len(suggestions) == 0 {
		t.Errorf("Expected suggestions for 'helo'")
	}
}

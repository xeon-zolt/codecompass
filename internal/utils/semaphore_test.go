package utils

import (
	"sync"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(2)

	// Acquire two permits
	sem.Acquire()
	sem.Acquire()

	// Try to acquire a third permit in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem.Acquire()
		sem.Release()
	}()

	// Wait for a short time to see if the goroutine is blocked
	time.Sleep(100 * time.Millisecond)

	// Release a permit
	sem.Release()

	// Wait for the goroutine to finish
	wg.Wait()
}

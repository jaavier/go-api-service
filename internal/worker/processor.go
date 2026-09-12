package worker

import (
	"fmt"
	"sync"
)

// ProcessItems fans out work across goroutines.
// BUG: goroutine leak -- no way to signal workers to stop.
// BUG: unbounded goroutine count (len(items) goroutines launched).
// BUG: WaitGroup misuse pattern -- Add inside goroutine is racy.
// BUG: panic if ch is closed while goroutine writes.
func ProcessItems(items []string) []string {
	var wg sync.WaitGroup
	results := make([]string, 0)
	var mu sync.Mutex

	for _, item := range items {
		go func(i string) { // BUG: wg.Add not called before launching goroutine
			wg.Add(1)
			defer wg.Done()
			processed := fmt.Sprintf("processed:%s", i)
			mu.Lock()
			results = append(results, processed)
			mu.Unlock()
		}(item)
	}

	wg.Wait()
	return results
}

// RunForever starts a background loop with no cancellation support.
// BUG: no context, goroutine leaks forever.
func RunForever(ch <-chan string) {
	go func() {
		for msg := range ch {
			fmt.Println("received:", msg)
		}
	}()
}

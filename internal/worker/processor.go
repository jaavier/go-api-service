package worker

import (
	"context"
	"fmt"
	"sync"
)

// ProcessItems fans out processing of items using a bounded worker pool.
//
// concurrency controls the maximum number of goroutines running in parallel.
// If concurrency <= 0 it defaults to 1.
// The function blocks until all items are processed or ctx is cancelled.
// It returns the processed results and any context error.
func ProcessItems(ctx context.Context, items []string, concurrency int) ([]string, error) {
	if concurrency <= 0 {
		concurrency = 1
	}

	results := make([]string, 0, len(items))
	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		sem = make(chan struct{}, concurrency) // semaphore / bounded pool
	)

	for _, item := range items {
		// Check for cancellation before launching each unit of work.
		select {
		case <-ctx.Done():
			wg.Wait()
			return results, ctx.Err()
		default:
		}

		sem <- struct{}{} // acquire slot — blocks when pool is full
		wg.Add(1)         // MUST be called before go, not inside

		go func(i string) {
			defer func() {
				<-sem // release slot
				wg.Done()
			}()

			processed := fmt.Sprintf("processed:%s", i)

			mu.Lock()
			results = append(results, processed)
			mu.Unlock()
		}(item)
	}

	wg.Wait()
	return results, nil
}

// Run reads messages from ch until the channel is closed or ctx is cancelled.
// The goroutine exits cleanly in both cases, preventing a goroutine leak.
func Run(ctx context.Context, ch <-chan string, handle func(string)) {
	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return // channel closed
				}
				handle(msg)
			case <-ctx.Done():
				return // context cancelled / timed out
			}
		}
	}()
}

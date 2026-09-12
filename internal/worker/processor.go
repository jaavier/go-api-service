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
//
// Return contract: on success it returns the processed results and a nil error.
// If ctx is cancelled before all items are dispatched, it waits for the
// in-flight goroutines to finish and returns the PARTIAL results gathered so
// far together with a non-nil ctx.Err(). Callers that only want complete
// results must check the error and discard the slice on a non-nil error.
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
		// Early exit if the context is already cancelled, regardless of
		// whether a slot is free in the pool. Without this check, when
		// concurrency >= len(items) (or a slot is always available) the
		// "sem <- struct{}{}" case below would win immediately and
		// ctx.Done() would never be observed, processing every item and
		// returning nil even for an already-cancelled context.
		select {
		case <-ctx.Done():
			wg.Wait()
			return results, ctx.Err()
		default:
		}

		// Acquire a slot, but also honour cancellation while blocked here.
		// A plain "sem <- struct{}{}" would block indefinitely when the pool
		// is full even if ctx was already cancelled; the select guarantees an
		// early exit under load.
		select {
		case sem <- struct{}{}: // acquire slot
		case <-ctx.Done():
			wg.Wait()
			return results, ctx.Err()
		}

		wg.Add(1) // MUST be called before go, not inside

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

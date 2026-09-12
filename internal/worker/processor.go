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
// If ctx is cancelled at any point — before all items are dispatched or after,
// while goroutines are still in flight — it waits for the in-flight goroutines
// to finish (no data race on the internal slice) and then returns a nil result
// slice together with a non-nil ctx.Err(). Partial results are intentionally
// discarded so that a caller can never accidentally use an incomplete slice by
// ignoring the error.
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
			return nil, ctx.Err()
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
			return nil, ctx.Err()
		}

		wg.Add(1) // MUST be called before go, not inside

		go func(i string) {
			defer func() {
				<-sem // release slot
				wg.Done()
			}()

			// Honour cancellation inside the goroutine too. Once items are
			// dispatched, a later cancellation must still be observed so the
			// worker skips its (potentially expensive) work instead of
			// blindly producing a result the caller will discard.
			select {
			case <-ctx.Done():
				return
			default:
			}

			processed := fmt.Sprintf("processed:%s", i)

			mu.Lock()
			results = append(results, processed)
			mu.Unlock()
		}(item)
	}

	wg.Wait()

	// A cancellation may have arrived after every item was dispatched but
	// before (or while) the goroutines ran. Check it here so the documented
	// contract holds end-to-end: on cancellation we discard partial results
	// and surface ctx.Err() rather than returning a possibly-incomplete slice
	// with a nil error.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Run reads messages from ch until the channel is closed or ctx is cancelled.
// The goroutine exits cleanly in both cases, preventing a goroutine leak.
//
// Run returns a *sync.WaitGroup already incremented for the spawned goroutine.
// Callers that need to guarantee the worker has fully drained before releasing
// shared resources (e.g. closing the DB during graceful shutdown) can call
// wg.Wait() after cancelling the context or closing ch. Multiple Run calls may
// share a single WaitGroup by passing the returned one back in via a wrapper,
// but the common pattern is to Wait on each returned group.
func Run(ctx context.Context, ch <-chan string, handle func(string)) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
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
	return &wg
}

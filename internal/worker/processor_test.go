package worker

import (
	"context"
	"sort"
	"testing"
	"time"
)

func TestProcessItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		items       []string
		concurrency int
		wantLen     int
		wantErr     bool
	}{
		{
			name:        "empty input",
			items:       nil,
			concurrency: 2,
			wantLen:     0,
		},
		{
			name:        "single item",
			items:       []string{"a"},
			concurrency: 1,
			wantLen:     1,
		},
		{
			name:        "multiple items concurrency 2",
			items:       []string{"a", "b", "c", "d"},
			concurrency: 2,
			wantLen:     4,
		},
		{
			name:        "concurrency defaults to 1 when zero",
			items:       []string{"x", "y"},
			concurrency: 0,
			wantLen:     2,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ProcessItems(context.Background(), tc.items, tc.concurrency)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tc.wantErr)
			}
			if len(got) != tc.wantLen {
				t.Errorf("len(results) = %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}

func TestProcessItems_ContentCorrect(t *testing.T) {
	t.Parallel()
	input := []string{"alpha", "beta", "gamma"}
	got, err := ProcessItems(context.Background(), input, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Order is non-deterministic with concurrent workers; sort before comparing.
	sort.Strings(got)
	want := []string{"processed:alpha", "processed:beta", "processed:gamma"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("got[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestProcessItems_ContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// With a cancelled context and many items, we may process 0 or more
	// but must never block and must return a context error.
	_, err := ProcessItems(ctx, []string{"a", "b", "c", "d", "e"}, 2)
	if err == nil {
		t.Error("expected context error, got nil")
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	ch := make(chan string, 3)
	ch <- "one"
	ch <- "two"
	ch <- "three"
	close(ch)

	var received []string
	done := make(chan struct{})

	ctx := context.Background()
	Run(ctx, ch, func(msg string) {
		received = append(received, msg)
		if len(received) == 3 {
			close(done)
		}
	})

	select {
	case <-done:
		// all messages received
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for all messages")
	}
}

func TestRun_ContextCancel(t *testing.T) {
	t.Parallel()

	ch := make(chan string) // unbuffered: a send only succeeds if someone reads
	ctx, cancel := context.WithCancel(context.Background())

	// Instrument handle so we can observe whether Run's goroutine is still
	// consuming after cancellation. A real consumer would push to this channel.
	calls := make(chan string, 1)
	Run(ctx, ch, func(msg string) { calls <- msg })

	// Cancel the context; Run's goroutine must observe ctx.Done() and exit.
	cancel()

	// After cancellation, attempt to hand a message to the worker. Because the
	// channel is unbuffered, the send can only proceed if the goroutine is
	// still reading — i.e. if it leaked. We therefore expect the send to time
	// out (nobody reading) and, crucially, handle to never be invoked.
	select {
	case ch <- "late":
		// Someone read from the channel after cancellation. Confirm whether
		// handle actually ran; if so, the goroutine leaked.
		select {
		case <-calls:
			t.Fatal("handle was invoked after context cancellation: Run goroutine leaked")
		case <-time.After(50 * time.Millisecond):
			t.Fatal("channel was read after cancellation but handle did not run: unexpected consumer")
		}
	case <-time.After(50 * time.Millisecond):
		// Expected: the goroutine already exited, so nobody consumes "late".
	}

	// Sanity: no spurious handle invocations happened.
	select {
	case msg := <-calls:
		t.Fatalf("unexpected handle invocation after cancel with msg %q", msg)
	default:
	}
}

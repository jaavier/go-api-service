package worker

import (
	"testing"
)

// BUG: no race-condition test (-race flag not mentioned),
// empty input not tested, result order not deterministic but not handled.
func TestProcessItems(t *testing.T) {
	got := ProcessItems([]string{"a", "b", "c"})
	if len(got) != 3 {
		t.Errorf("expected 3 results, got %d", len(got))
	}
}

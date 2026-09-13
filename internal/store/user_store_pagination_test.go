package store

import (
	"context"
	"testing"
)

// TestListPaginated_InvalidArgs verifies the guard clauses without needing a
// live database: invalid limit/offset must fail fast before touching the DB.
func TestListPaginated_InvalidArgs(t *testing.T) {
	s := &UserStore{Db: nil} // Db is never used because validation fails first

	tests := []struct {
		name   string
		limit  int
		offset int
	}{
		{"zero limit", 0, 0},
		{"negative limit", -1, 0},
		{"negative offset", 10, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := s.ListPaginated(context.Background(), tt.limit, tt.offset)
			if err == nil {
				t.Fatalf("ListPaginated(%d, %d) expected error, got nil", tt.limit, tt.offset)
			}
		})
	}
}

package store

import (
	"context"
	"errors"
	"testing"

	"github.com/jaavier/go-api-service/internal/model"
)

// TestListPaginated_InvalidArgs verifies the guard clauses without needing a
// live database: an invalid page must fail fast (ErrInvalidPage) before
// touching the DB, so a nil *sql.DB is safe here.
func TestListPaginated_InvalidArgs(t *testing.T) {
	s := &UserStore{Db: nil} // Db is never used because validation fails first

	tests := []struct {
		name string
		page model.Page
	}{
		{"zero size", model.Page{Number: 1, Size: 0}},
		{"negative size", model.Page{Number: 1, Size: -1}},
		{"zero page number", model.Page{Number: 0, Size: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := s.ListPaginated(context.Background(), tt.page)
			if err == nil {
				t.Fatalf("ListPaginated(%+v) expected error, got nil", tt.page)
			}
			if !errors.Is(err, ErrInvalidPage) {
				t.Fatalf("ListPaginated(%+v) error = %v, want ErrInvalidPage", tt.page, err)
			}
		})
	}
}

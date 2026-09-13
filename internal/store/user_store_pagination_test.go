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
			_, _, err := s.ListPaginated(context.Background(), tt.page, model.UserFilter{})
			if err == nil {
				t.Fatalf("ListPaginated(%+v) expected error, got nil", tt.page)
			}
			if !errors.Is(err, ErrInvalidPage) {
				t.Fatalf("ListPaginated(%+v) error = %v, want ErrInvalidPage", tt.page, err)
			}
		})
	}
}

// TestUserFilterSafeLiterals documents the injection-defense contract the store
// relies on: whatever a client sends for sort/order, the values the store
// concatenates into ORDER BY are always drawn from a fixed whitelist.
func TestUserFilterSafeLiterals(t *testing.T) {
	f := model.UserFilter{SortBy: "name); DROP TABLE users;--", Order: "desc); --"}
	if col := f.SortColumn(); col != "id" {
		t.Errorf("SortColumn() = %q, want fixed fallback %q", col, "id")
	}
	if dir := f.Direction(); dir != "ASC" {
		t.Errorf("Direction() = %q, want fixed fallback %q", dir, "ASC")
	}

	valid := model.UserFilter{SortBy: "email", Order: "desc"}
	if col := valid.SortColumn(); col != "email" {
		t.Errorf("SortColumn() = %q, want %q", col, "email")
	}
	if dir := valid.Direction(); dir != "DESC" {
		t.Errorf("Direction() = %q, want %q", dir, "DESC")
	}
}

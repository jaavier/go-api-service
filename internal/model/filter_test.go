package model

import "testing"

func TestUserFilterSortColumn(t *testing.T) {
	tests := []struct {
		name   string
		sortBy string
		want   string
	}{
		{"default id", "id", "id"},
		{"name", "name", "name"},
		{"email", "email", "email"},
		{"uppercase normalized", "EMAIL", "email"},
		{"empty falls back to id", "", "id"},
		{"injection attempt falls back to id", "name; DROP TABLE users", "id"},
		{"unknown column falls back to id", "created_at", "id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := UserFilter{SortBy: tt.sortBy}
			if got := f.SortColumn(); got != tt.want {
				t.Errorf("SortColumn() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUserFilterDirection(t *testing.T) {
	tests := []struct {
		name  string
		order string
		want  string
	}{
		{"default asc", "", "ASC"},
		{"asc", "asc", "ASC"},
		{"desc", "desc", "DESC"},
		{"desc uppercase normalized", "DESC", "DESC"},
		{"invalid falls back to asc", "weird", "ASC"},
		{"injection attempt falls back to asc", "asc; DROP TABLE users", "ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := UserFilter{Order: tt.order}
			if got := f.Direction(); got != tt.want {
				t.Errorf("Direction() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUserFilterHasQuery(t *testing.T) {
	if (UserFilter{Query: ""}).HasQuery() {
		t.Error("HasQuery() = true for empty query, want false")
	}
	if !(UserFilter{Query: "ada"}).HasQuery() {
		t.Error("HasQuery() = false for non-empty query, want true")
	}
}

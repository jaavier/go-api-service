package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaavier/go-api-service/internal/model"
)

// mockLister is a lightweight in-memory implementation of UserLister used to
// exercise the handler without a real database.
type mockLister struct {
	users []*model.User
	total int
	err   error

	gotPage   model.Page
	gotFilter model.UserFilter
}

func (m *mockLister) ListPaginated(_ context.Context, page model.Page, filter model.UserFilter) ([]*model.User, int, error) {
	m.gotPage = page
	m.gotFilter = filter
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.users, m.total, nil
}

func newHandler(l UserLister) *UserHandler {
	return &UserHandler{lister: l}
}

func TestListUsers(t *testing.T) {
	sample := []*model.User{
		{ID: 1, Name: "Ada", Email: "ada@example.com"},
		{ID: 2, Name: "Linus", Email: "linus@example.com"},
	}

	tests := []struct {
		name           string
		url            string
		lister         *mockLister
		wantStatus     int
		wantPage       int
		wantSize       int
		wantTotal      int
		wantTotalPages int
		wantLen        int
	}{
		{
			name:           "defaults when no query params",
			url:            "/users",
			lister:         &mockLister{users: sample, total: 42},
			wantStatus:     http.StatusOK,
			wantPage:       1,
			wantSize:       20,
			wantTotal:      42,
			wantTotalPages: 3,
			wantLen:        2,
		},
		{
			name:           "explicit page and page_size",
			url:            "/users?page=3&page_size=5",
			lister:         &mockLister{users: sample, total: 42},
			wantStatus:     http.StatusOK,
			wantPage:       3,
			wantSize:       5,
			wantTotal:      42,
			wantTotalPages: 9,
			wantLen:        2,
		},
		{
			name:           "page_size capped at max",
			url:            "/users?page_size=1000",
			lister:         &mockLister{users: sample, total: 42},
			wantStatus:     http.StatusOK,
			wantPage:       1,
			wantSize:       maxPageSize,
			wantTotal:      42,
			wantTotalPages: 1,
			wantLen:        2,
		},
		{
			name:           "invalid page falls back to default",
			url:            "/users?page=abc",
			lister:         &mockLister{users: sample, total: 42},
			wantStatus:     http.StatusOK,
			wantPage:       1,
			wantSize:       20,
			wantTotal:      42,
			wantTotalPages: 3,
			wantLen:        2,
		},
		{
			name:           "zero page_size falls back to default",
			url:            "/users?page_size=0",
			lister:         &mockLister{users: sample, total: 42},
			wantStatus:     http.StatusOK,
			wantPage:       1,
			wantSize:       20,
			wantTotal:      42,
			wantTotalPages: 3,
			wantLen:        2,
		},
		{
			name:       "store error returns 500",
			url:        "/users",
			lister:     &mockLister{err: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(tt.lister)
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			h.List(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var got model.PagedUsers
			if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Page != tt.wantPage {
				t.Errorf("page = %d, want %d", got.Page, tt.wantPage)
			}
			if got.PageSize != tt.wantSize {
				t.Errorf("page_size = %d, want %d", got.PageSize, tt.wantSize)
			}
			if got.Total != tt.wantTotal {
				t.Errorf("total = %d, want %d", got.Total, tt.wantTotal)
			}
			if got.TotalPages != tt.wantTotalPages {
				t.Errorf("total_pages = %d, want %d", got.TotalPages, tt.wantTotalPages)
			}
			if len(got.Data) != tt.wantLen {
				t.Errorf("len(data) = %d, want %d", len(got.Data), tt.wantLen)
			}
			if tt.lister.gotPage.Number != tt.wantPage || tt.lister.gotPage.Size != tt.wantSize {
				t.Errorf("store got page %+v, want {Number:%d Size:%d}", tt.lister.gotPage, tt.wantPage, tt.wantSize)
			}
		})
	}
}

// TestListUsersFilter verifies that the handler parses the q/sort/order params
// into a normalized model.UserFilter and forwards it to the store, including
// the injection-defense fallbacks.
func TestListUsersFilter(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		wantQuery     string
		wantSortCol   string
		wantDirection string
	}{
		{
			name:          "no filter keeps defaults",
			url:           "/users",
			wantQuery:     "",
			wantSortCol:   "id",
			wantDirection: "ASC",
		},
		{
			name:          "query passed through trimmed",
			url:           "/users?q=ada",
			wantQuery:     "ada",
			wantSortCol:   "id",
			wantDirection: "ASC",
		},
		{
			name:          "whitespace-only query ignored",
			url:           "/users?q=%20%20",
			wantQuery:     "",
			wantSortCol:   "id",
			wantDirection: "ASC",
		},
		{
			name:          "sort and order normalized",
			url:           "/users?sort=email&order=desc",
			wantQuery:     "",
			wantSortCol:   "email",
			wantDirection: "DESC",
		},
		{
			name:          "injection attempt falls back to id asc",
			url:           "/users?sort=drop%20table&order=weird",
			wantQuery:     "",
			wantSortCol:   "id",
			wantDirection: "ASC",
		},
		{
			name:          "filter combines with pagination",
			url:           "/users?q=lin&sort=name&order=asc&page=2&page_size=5",
			wantQuery:     "lin",
			wantSortCol:   "name",
			wantDirection: "ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lister := &mockLister{users: nil, total: 0}
			h := newHandler(lister)
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			h.List(w, req)

			if got := lister.gotFilter.Query; got != tt.wantQuery {
				t.Errorf("filter.Query = %q, want %q", got, tt.wantQuery)
			}
			if got := lister.gotFilter.SortColumn(); got != tt.wantSortCol {
				t.Errorf("filter.SortColumn() = %q, want %q", got, tt.wantSortCol)
			}
			if got := lister.gotFilter.Direction(); got != tt.wantDirection {
				t.Errorf("filter.Direction() = %q, want %q", got, tt.wantDirection)
			}
		})
	}
}

// TestParseUserFilterQueryCap ensures overly long search terms are truncated.
func TestParseUserFilterQueryCap(t *testing.T) {
	long := make([]byte, maxQueryLen+50)
	for i := range long {
		long[i] = 'a'
	}
	req := httptest.NewRequest(http.MethodGet, "/users?q="+string(long), nil)
	f := parseUserFilter(req)
	if len(f.Query) != maxQueryLen {
		t.Errorf("len(Query) = %d, want %d", len(f.Query), maxQueryLen)
	}
}

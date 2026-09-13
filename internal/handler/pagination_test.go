package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePageParams(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		wantPage     int
		wantPageSize int
		wantLimit    int
		wantOffset   int
	}{
		{
			name:         "defaults when empty",
			query:        "",
			wantPage:     1,
			wantPageSize: 20,
			wantLimit:    20,
			wantOffset:   0,
		},
		{
			name:         "explicit values",
			query:        "page=3&page_size=10",
			wantPage:     3,
			wantPageSize: 10,
			wantLimit:    10,
			wantOffset:   20,
		},
		{
			name:         "negative page falls back to default",
			query:        "page=-5&page_size=5",
			wantPage:     1,
			wantPageSize: 5,
			wantLimit:    5,
			wantOffset:   0,
		},
		{
			name:         "non-numeric falls back to defaults",
			query:        "page=abc&page_size=xyz",
			wantPage:     1,
			wantPageSize: 20,
			wantLimit:    20,
			wantOffset:   0,
		},
		{
			name:         "page_size capped at max",
			query:        "page=2&page_size=1000",
			wantPage:     2,
			wantPageSize: 100,
			wantLimit:    100,
			wantOffset:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users?"+tt.query, nil)
			p := parsePageParams(req)

			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %d, want %d", p.PageSize, tt.wantPageSize)
			}
			if p.limit() != tt.wantLimit {
				t.Errorf("limit() = %d, want %d", p.limit(), tt.wantLimit)
			}
			if p.offset() != tt.wantOffset {
				t.Errorf("offset() = %d, want %d", p.offset(), tt.wantOffset)
			}
		})
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	tests := []struct {
		name           string
		p              pageParams
		total          int
		wantTotalPages int
	}{
		{"exact multiple", pageParams{Page: 1, PageSize: 10}, 30, 3},
		{"with remainder", pageParams{Page: 1, PageSize: 10}, 31, 4},
		{"empty set", pageParams{Page: 1, PageSize: 10}, 0, 0},
		{"single partial page", pageParams{Page: 1, PageSize: 20}, 5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := newPaginatedResponse(nil, tt.p, tt.total)
			if resp.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", resp.TotalPages, tt.wantTotalPages)
			}
			if resp.Total != tt.total {
				t.Errorf("Total = %d, want %d", resp.Total, tt.total)
			}
			if resp.Page != tt.p.Page {
				t.Errorf("Page = %d, want %d", resp.Page, tt.p.Page)
			}
			if resp.PageSize != tt.p.PageSize {
				t.Errorf("PageSize = %d, want %d", resp.PageSize, tt.p.PageSize)
			}
		})
	}
}

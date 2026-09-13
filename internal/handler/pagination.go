package handler

import (
	"net/http"
	"strconv"
)

const (
	// defaultPage is used when the page query param is missing or invalid.
	defaultPage = 1
	// defaultPageSize is used when page_size is missing or invalid.
	defaultPageSize = 20
	// maxPageSize caps page_size to protect the backend from large scans.
	maxPageSize = 100
)

// pageParams holds a validated, clamped pagination request.
type pageParams struct {
	Page     int
	PageSize int
}

// limit returns the SQL LIMIT for the page.
func (p pageParams) limit() int { return p.PageSize }

// offset returns the SQL OFFSET for the page.
func (p pageParams) offset() int { return (p.Page - 1) * p.PageSize }

// parsePageParams reads page and page_size from the query string and applies
// sane defaults and bounds. Invalid or missing values fall back to defaults
// rather than erroring, which keeps the endpoint forgiving for clients.
func parsePageParams(r *http.Request) pageParams {
	q := r.URL.Query()

	page := defaultPage
	if v, err := strconv.Atoi(q.Get("page")); err == nil && v > 0 {
		page = v
	}

	pageSize := defaultPageSize
	if v, err := strconv.Atoi(q.Get("page_size")); err == nil && v > 0 {
		pageSize = v
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return pageParams{Page: page, PageSize: pageSize}
}

// paginatedResponse is the JSON envelope returned by the paginated list.
type paginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// newPaginatedResponse builds an envelope, computing total_pages from total.
func newPaginatedResponse(data interface{}, p pageParams, total int) paginatedResponse {
	totalPages := 0
	if p.PageSize > 0 {
		totalPages = (total + p.PageSize - 1) / p.PageSize
	}
	return paginatedResponse{
		Data:       data,
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

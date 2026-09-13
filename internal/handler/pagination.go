package handler

import (
	"net/http"
	"strconv"

	"github.com/jaavier/go-api-service/internal/model"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// pageParams holds the normalized pagination parameters parsed from the
// request query string.
type pageParams struct {
	Page     int
	PageSize int
}

// limit is the SQL LIMIT for the page (== PageSize).
func (p pageParams) limit() int {
	return p.PageSize
}

// offset is the zero-based SQL OFFSET for the page. It never goes negative.
func (p pageParams) offset() int {
	if p.Page < 1 || p.PageSize < 1 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// toModelPage converts the parsed params into the model.Page value expected
// by the store layer.
func (p pageParams) toModelPage() model.Page {
	return model.Page{Number: p.Page, Size: p.PageSize}
}

// parsePageParams extracts and normalizes the `page` and `page_size` query
// params. Missing, non-numeric or out-of-range values fall back to sane
// defaults; page_size is capped at maxPageSize to protect the backend.
func parsePageParams(r *http.Request) pageParams {
	q := r.URL.Query()

	page := parsePositiveInt(q.Get("page"), defaultPage)
	size := parsePositiveInt(q.Get("page_size"), defaultPageSize)
	if size > maxPageSize {
		size = maxPageSize
	}

	return pageParams{Page: page, PageSize: size}
}

// parsePositiveInt returns the parsed value when raw is a valid positive
// integer, otherwise it falls back to def (covers empty, non-numeric and
// non-positive input).
func parsePositiveInt(raw string, def int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return def
	}
	return v
}

// newPaginatedResponse builds the response envelope, computing total_pages as
// ceil(total/PageSize). total_pages is 0 when there are no results.
func newPaginatedResponse(data []*model.User, p pageParams, total int) model.PagedUsers {
	totalPages := 0
	if total > 0 && p.PageSize > 0 {
		totalPages = (total + p.PageSize - 1) / p.PageSize
	}
	return model.PagedUsers{
		Data:       data,
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

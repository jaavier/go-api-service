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

// parsePage extracts and normalizes the `page` and `page_size` query params.
//
// Missing values fall back to sane defaults; page_size is capped at maxPageSize
// to protect the backend. It returns an error only when a provided value is
// syntactically invalid or out of range, so the handler can answer 400.
func parsePage(r *http.Request) (model.Page, error) {
	q := r.URL.Query()

	page, err := parsePositiveInt(q.Get("page"), defaultPage)
	if err != nil {
		return model.Page{}, err
	}

	size, err := parsePositiveInt(q.Get("page_size"), defaultPageSize)
	if err != nil {
		return model.Page{}, err
	}
	if size > maxPageSize {
		size = maxPageSize
	}

	return model.Page{Number: page, Size: size}, nil
}

func parsePositiveInt(raw string, def int) (int, error) {
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return 0, errInvalidPagination
	}
	return v, nil
}

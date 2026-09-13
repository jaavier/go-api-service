package model

import "strings"

// Sort column names accepted from clients. These are the ONLY values that may
// influence the SQL ORDER BY clause; anything else falls back to the default.
const (
	SortByID    = "id"
	SortByName  = "name"
	SortByEmail = "email"
)

// Order directions accepted from clients.
const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

// allowedSortColumns is the strict whitelist mapping a requested sort field to
// the physical column that is safe to interpolate into an ORDER BY clause.
// Keeping this as a fixed map (never the raw client value) is what prevents SQL
// injection through the `sort` parameter.
var allowedSortColumns = map[string]string{
	SortByID:    "id",
	SortByName:  "name",
	SortByEmail: "email",
}

// UserFilter holds the normalized search and ordering criteria for a users
// listing. It is built from the HTTP query string and consumed by the store.
//
// Query is an already-trimmed substring; an empty Query means "no filter".
// SortBy and Order are validated through the SortColumn/Direction helpers,
// which always return safe literals regardless of the raw input.
type UserFilter struct {
	Query  string
	SortBy string
	Order  string
}

// SortColumn returns the physical column to order by. It maps SortBy through
// the strict whitelist and falls back to "id" for any unknown value, so the
// returned string is always safe to concatenate into an ORDER BY clause.
func (f UserFilter) SortColumn() string {
	if col, ok := allowedSortColumns[strings.ToLower(f.SortBy)]; ok {
		return col
	}
	return allowedSortColumns[SortByID]
}

// Direction returns the SQL sort direction as a fixed literal ("ASC" or
// "DESC"). Any value other than a case-insensitive "desc" normalizes to "ASC",
// so the returned string is always safe to concatenate into an ORDER BY clause.
func (f UserFilter) Direction() string {
	if strings.ToLower(f.Order) == OrderDesc {
		return "DESC"
	}
	return "ASC"
}

// HasQuery reports whether a substring search filter is set.
func (f UserFilter) HasQuery() bool {
	return f.Query != ""
}

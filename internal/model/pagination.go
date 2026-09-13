package model

// Page holds pagination parameters after validation and normalization.
type Page struct {
	Number int `json:"page"`
	Size   int `json:"page_size"`
}

// Offset returns the zero-based SQL OFFSET for the page.
//
// It never returns a negative value: a Number below 1 is treated as the
// first page so callers (and the store) can rely on Offset() >= 0 even if
// the Page was built without going through the HTTP validation path.
func (p Page) Offset() int {
	if p.Number < 1 || p.Size < 1 {
		return 0
	}
	return (p.Number - 1) * p.Size
}

// PagedUsers is the paginated response envelope for a list of users.
type PagedUsers struct {
	Data       []*User `json:"data"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	Total      int     `json:"total"`
	TotalPages int     `json:"total_pages"`
}

package model

// Page holds pagination parameters after validation and normalization.
type Page struct {
	Number int `json:"page"`
	Size   int `json:"page_size"`
}

// Offset returns the zero-based SQL OFFSET for the page.
func (p Page) Offset() int {
	return (p.Number - 1) * p.Size
}

// PagedUsers is the paginated response envelope for a list of users.
type PagedUsers struct {
	Data     []*User `json:"data"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
	Total    int     `json:"total"`
}

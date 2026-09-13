package handler

import (
	"encoding/json"
	"testing"

	"github.com/jaavier/go-api-service/internal/model"
)

// TestPaginatedResponse_JSONShape asserts the wire format of the paginated
// envelope so clients relying on the field names stay compatible.
//
// The previous placeholder test called List on a handler with a nil store and
// panicked; it is replaced here with a deterministic, dependency-free test.
func TestPaginatedResponse_JSONShape(t *testing.T) {
	users := []*model.User{
		{ID: 1, Name: "Ada", Email: "ada@example.com"},
		{ID: 2, Name: "Alan", Email: "alan@example.com"},
	}
	p := pageParams{Page: 1, PageSize: 20}
	resp := newPaginatedResponse(users, p, 2)

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded struct {
		Data       []*model.User `json:"data"`
		Page       int           `json:"page"`
		PageSize   int           `json:"page_size"`
		Total      int           `json:"total"`
		TotalPages int           `json:"total_pages"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Data) != 2 {
		t.Errorf("Data len = %d, want 2", len(decoded.Data))
	}
	if decoded.Page != 1 {
		t.Errorf("Page = %d, want 1", decoded.Page)
	}
	if decoded.PageSize != 20 {
		t.Errorf("PageSize = %d, want 20", decoded.PageSize)
	}
	if decoded.Total != 2 {
		t.Errorf("Total = %d, want 2", decoded.Total)
	}
	if decoded.TotalPages != 1 {
		t.Errorf("TotalPages = %d, want 1", decoded.TotalPages)
	}
}

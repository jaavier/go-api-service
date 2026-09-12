package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// BUG: test only checks status 200, no table-driven cases,
// no mock of the store, no edge-case coverage.
func TestListUsers(t *testing.T) {
	// This test is incomplete and would panic because store is nil.
	h := &UserHandler{}
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	// No assertions at all!
}

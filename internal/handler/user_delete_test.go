package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jaavier/go-api-service/internal/store"
)

// mockDeleter is a lightweight implementation of UserDeleter used to exercise
// the Delete handler without a real database.
type mockDeleter struct {
	err error

	called bool
	gotID  int64
	gotCtx context.Context
}

func (m *mockDeleter) DeleteByID(ctx context.Context, id int64) error {
	m.called = true
	m.gotCtx = ctx
	m.gotID = id
	return m.err
}

func newDeleteHandler(d UserDeleter) *UserHandler {
	return &UserHandler{deleter: d}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		pathID     string
		deleter    *mockDeleter
		wantStatus int
		wantError  string // expected "error" field; empty means no JSON error body
		wantCalled bool
	}{
		{
			name:       "existing id returns 204 with no body",
			pathID:     "7",
			deleter:    &mockDeleter{},
			wantStatus: http.StatusNoContent,
			wantCalled: true,
		},
		{
			name:       "missing id returns 404",
			pathID:     "7",
			deleter:    &mockDeleter{err: store.ErrNotFound},
			wantStatus: http.StatusNotFound,
			wantError:  "user not found",
			wantCalled: true,
		},
		{
			name:       "generic store error returns 500",
			pathID:     "7",
			deleter:    &mockDeleter{err: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal error",
			wantCalled: true,
		},
		{
			name:       "non-numeric id returns 400 and skips store",
			pathID:     "abc",
			deleter:    &mockDeleter{},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid id",
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newDeleteHandler(tt.deleter)
			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.pathID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.pathID})
			w := httptest.NewRecorder()

			h.Delete(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			if tt.deleter.called != tt.wantCalled {
				t.Errorf("store called = %v, want %v", tt.deleter.called, tt.wantCalled)
			}

			if tt.wantError == "" {
				// 204 must have no body and no JSON content type.
				body, _ := io.ReadAll(res.Body)
				if len(body) != 0 {
					t.Errorf("body = %q, want empty", string(body))
				}
				if ct := res.Header.Get("Content-Type"); ct == "application/json" {
					t.Errorf("Content-Type = %q, want no JSON content type on 204", ct)
				}
				if tt.deleter.gotID != 7 {
					t.Errorf("store got id = %d, want 7", tt.deleter.gotID)
				}
				if tt.deleter.gotCtx == nil {
					t.Error("expected request context to be propagated to the store")
				}
				return
			}

			if ct := res.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			var errBody map[string]string
			if err := json.NewDecoder(res.Body).Decode(&errBody); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if errBody["error"] != tt.wantError {
				t.Errorf("error = %q, want %q", errBody["error"], tt.wantError)
			}
		})
	}
}

// TestDeleteUserPropagatesContext asserts the exact request context reaches the
// store.
func TestDeleteUserPropagatesContext(t *testing.T) {
	d := &mockDeleter{}
	h := newDeleteHandler(d)

	type ctxKey string
	const key ctxKey = "trace"
	req := httptest.NewRequest(http.MethodDelete, "/users/7", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "7"})
	req = req.WithContext(context.WithValue(req.Context(), key, "abc123"))
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if d.gotCtx == nil {
		t.Fatal("context not propagated to the store")
	}
	if got := d.gotCtx.Value(key); got != "abc123" {
		t.Errorf("ctx value = %v, want abc123", got)
	}
}

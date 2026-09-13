package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jaavier/go-api-service/internal/model"
	"github.com/jaavier/go-api-service/internal/store"
)

// mockGetter is a lightweight in-memory implementation of UserGetter used to
// exercise the Get handler without a real database.
type mockGetter struct {
	user *model.User
	err  error

	gotID  int64
	gotCtx context.Context
}

func (m *mockGetter) GetByID(ctx context.Context, id int64) (*model.User, error) {
	m.gotCtx = ctx
	m.gotID = id
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func newGetHandler(g UserGetter) *UserHandler {
	return &UserHandler{getter: g}
}

func TestGetUser(t *testing.T) {
	sample := &model.User{ID: 7, Name: "Ada", Email: "ada@example.com"}

	tests := []struct {
		name       string
		pathID     string
		getter     *mockGetter
		wantStatus int
		wantError  string // expected "error" field; empty means success body
	}{
		{
			name:       "existing id returns 200",
			pathID:     "7",
			getter:     &mockGetter{user: sample},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing id returns 404",
			pathID:     "7",
			getter:     &mockGetter{err: store.ErrNotFound},
			wantStatus: http.StatusNotFound,
			wantError:  "user not found",
		},
		{
			name:       "generic store error returns 500",
			pathID:     "7",
			getter:     &mockGetter{err: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal error",
		},
		{
			name:       "non-numeric id returns 400",
			pathID:     "abc",
			getter:     &mockGetter{},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newGetHandler(tt.getter)
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.pathID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.pathID})
			w := httptest.NewRecorder()

			h.Get(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			if ct := res.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			if tt.wantError == "" {
				var got model.User
				if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if got != *sample {
					t.Errorf("body = %+v, want %+v", got, *sample)
				}
				if tt.getter.gotID != 7 {
					t.Errorf("store got id = %d, want 7", tt.getter.gotID)
				}
				if tt.getter.gotCtx == nil {
					t.Error("expected request context to be propagated to the store")
				}
				return
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

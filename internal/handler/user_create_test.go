package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jaavier/go-api-service/internal/model"
)

// mockCreator is a lightweight in-memory implementation of UserCreator used to
// exercise the Create handler without a real database. When err is nil it
// simulates the store assigning a generated primary key to u.ID.
type mockCreator struct {
	assignID int64
	err      error

	called bool
	gotCtx context.Context
	gotUser *model.User
}

func (m *mockCreator) Create(ctx context.Context, u *model.User) error {
	m.called = true
	m.gotCtx = ctx
	m.gotUser = u
	if m.err != nil {
		return m.err
	}
	u.ID = m.assignID
	return nil
}

func newCreateHandler(c UserCreator) *UserHandler {
	return &UserHandler{creator: c}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		creator     *mockCreator
		wantStatus  int
		wantError   string // expected "error" field; empty means success body
		wantCalled  bool   // whether the store should have been invoked
	}{
		{
			name:       "valid body returns 201 with assigned id",
			body:       `{"name":"Ada","email":"ada@example.com"}`,
			creator:    &mockCreator{assignID: 42},
			wantStatus: http.StatusCreated,
			wantCalled: true,
		},
		{
			name:       "invalid json returns 400",
			body:       `{not json`,
			creator:    &mockCreator{},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid body",
			wantCalled: false,
		},
		{
			name:       "blank name returns 400 and skips store",
			body:       `{"name":"   ","email":"ada@example.com"}`,
			creator:    &mockCreator{},
			wantStatus: http.StatusBadRequest,
			wantError:  "name and email are required",
			wantCalled: false,
		},
		{
			name:       "blank email returns 400 and skips store",
			body:       `{"name":"Ada","email":""}`,
			creator:    &mockCreator{},
			wantStatus: http.StatusBadRequest,
			wantError:  "name and email are required",
			wantCalled: false,
		},
		{
			name:       "generic store error returns 500",
			body:       `{"name":"Ada","email":"ada@example.com"}`,
			creator:    &mockCreator{err: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal error",
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newCreateHandler(tt.creator)
			req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.Create(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			if ct := res.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if tt.creator.called != tt.wantCalled {
				t.Errorf("store called = %v, want %v", tt.creator.called, tt.wantCalled)
			}

			if tt.wantError == "" {
				var got model.User
				if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				want := model.User{ID: 42, Name: "Ada", Email: "ada@example.com"}
				if got != want {
					t.Errorf("body = %+v, want %+v", got, want)
				}
				if tt.creator.gotCtx == nil {
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

// TestCreateUserPropagatesContext asserts the exact request context reaches the
// store, not just a non-nil one.
func TestCreateUserPropagatesContext(t *testing.T) {
	c := &mockCreator{assignID: 1}
	h := newCreateHandler(c)

	type ctxKey string
	const key ctxKey = "trace"
	body := bytes.NewReader([]byte(`{"name":"Ada","email":"ada@example.com"}`))
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	req = req.WithContext(context.WithValue(req.Context(), key, "abc123"))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if c.gotCtx == nil {
		t.Fatal("context not propagated to the store")
	}
	if got := c.gotCtx.Value(key); got != "abc123" {
		t.Errorf("ctx value = %v, want abc123", got)
	}
}

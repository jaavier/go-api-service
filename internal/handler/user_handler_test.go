package handler

import (
	"bytes"
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

// fakeRepo is an in-memory implementation of store.UserRepository for tests.
type fakeRepo struct {
	users []*model.User
	err   error // if set, every method returns this error
}

func (f *fakeRepo) List(_ context.Context) ([]*model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.users, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id int64) (*model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, store.ErrNotFound
}

func (f *fakeRepo) Create(_ context.Context, u *model.User) error {
	if f.err != nil {
		return f.err
	}
	u.ID = int64(len(f.users) + 1)
	f.users = append(f.users, u)
	return nil
}

func (f *fakeRepo) DeleteByID(_ context.Context, id int64) error {
	if f.err != nil {
		return f.err
	}
	for i, u := range f.users {
		if u.ID == id {
			f.users = append(f.users[:i], f.users[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

// routeRequest wires a request through a mux router so that mux.Vars works.
func routeRequest(method, path, pattern string, body []byte, h http.HandlerFunc) *httptest.ResponseRecorder {
	r := mux.NewRouter()
	r.HandleFunc(pattern, h).Methods(method)

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------- List -------------------------------------------------------------

func TestList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		repo       *fakeRepo
		wantStatus int
	}{
		{
			name:       "returns users",
			repo:       &fakeRepo{users: []*model.User{{ID: 1, Name: "Alice", Email: "a@x.com"}}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty list",
			repo:       &fakeRepo{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "store error",
			repo:       &fakeRepo{err: errors.New("db down")},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := NewUserHandler(tc.repo)
			w := routeRequest(http.MethodGet, "/users", "/users", nil, h.List)
			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

// ---------- Get --------------------------------------------------------------

func TestGet(t *testing.T) {
	t.Parallel()
	alice := &model.User{ID: 1, Name: "Alice", Email: "a@x.com"}
	tests := []struct {
		name       string
		repo       *fakeRepo
		path       string
		wantStatus int
	}{
		{
			name:       "found",
			repo:       &fakeRepo{users: []*model.User{alice}},
			path:       "/users/1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found",
			repo:       &fakeRepo{},
			path:       "/users/99",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "bad id",
			repo:       &fakeRepo{},
			path:       "/users/abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "store error",
			repo:       &fakeRepo{err: errors.New("db down")},
			path:       "/users/1",
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := NewUserHandler(tc.repo)
			w := routeRequest(http.MethodGet, tc.path, "/users/{id}", nil, h.Get)
			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

// ---------- Create -----------------------------------------------------------

func TestCreate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		body       any
		repo       *fakeRepo
		wantStatus int
	}{
		{
			name:       "created",
			body:       model.User{Name: "Bob", Email: "b@x.com"},
			repo:       &fakeRepo{},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing name",
			body:       model.User{Email: "b@x.com"},
			repo:       &fakeRepo{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing email",
			body:       model.User{Name: "Bob"},
			repo:       &fakeRepo{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       nil, // signals raw bytes
			repo:       &fakeRepo{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "store error",
			body:       model.User{Name: "Bob", Email: "b@x.com"},
			repo:       &fakeRepo{err: errors.New("db down")},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := NewUserHandler(tc.repo)
			var rawBody []byte
			if tc.body != nil {
				var err error
				rawBody, err = json.Marshal(tc.body)
				if err != nil {
					t.Fatalf("marshal body: %v", err)
				}
			} else {
				rawBody = []byte("{invalid")
			}
			w := routeRequest(http.MethodPost, "/users", "/users", rawBody, h.Create)
			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d (body: %s)", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}
}

// ---------- Delete -----------------------------------------------------------

func TestDelete(t *testing.T) {
	t.Parallel()
	alice := &model.User{ID: 1, Name: "Alice", Email: "a@x.com"}
	tests := []struct {
		name       string
		repo       *fakeRepo
		path       string
		wantStatus int
	}{
		{
			name:       "deleted",
			repo:       &fakeRepo{users: []*model.User{alice}},
			path:       "/users/1",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "not found",
			repo:       &fakeRepo{},
			path:       "/users/99",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "bad id",
			repo:       &fakeRepo{},
			path:       "/users/xyz",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "store error",
			repo:       &fakeRepo{err: errors.New("db down")},
			path:       "/users/1",
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := NewUserHandler(tc.repo)
			w := routeRequest(http.MethodDelete, tc.path, "/users/{id}", nil, h.Delete)
			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

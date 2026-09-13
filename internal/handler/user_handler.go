package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jaavier/go-api-service/internal/model"
	"github.com/jaavier/go-api-service/internal/store"
)

// UserLister is the read dependency the handler needs to serve paginated lists.
// Depending on an interface (rather than *store.UserStore) keeps the handler
// testable with a lightweight mock.
type UserLister interface {
	ListPaginated(ctx context.Context, page model.Page, filter model.UserFilter) ([]*model.User, int, error)
}

// UserGetter is the read-by-id dependency for the Get endpoint. Depending on an
// interface (rather than *store.UserStore) keeps the handler testable with a
// lightweight mock. *store.UserStore satisfies it via its GetByID method.
type UserGetter interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

// UserHandler handles HTTP requests for users.
//
// lister serves the paginated List endpoint and can be a mock in tests.
// getter serves the Get endpoint and can be a mock in tests.
// crud is the concrete store used by Create/Delete; it is nil in the
// read-only handler tests, which never exercise those routes.
type UserHandler struct {
	lister UserLister
	getter UserGetter
	crud   *store.UserStore
}

func NewUserHandler(s *store.UserStore) *UserHandler {
	return &UserHandler{lister: s, getter: s, crud: s}
}

// List returns a paginated page of users.
//
// Query params:
//   - page:      1-based page number (default 1)
//   - page_size: items per page (default 20, capped at 100)
//   - q:         case-insensitive substring filter on name OR email (optional)
//   - sort:      order field, one of id|name|email (default id)
//   - order:     order direction, asc|desc (default asc)
//
// Invalid or missing values fall back to the defaults; without any filter the
// endpoint behaves exactly as before (ordered by id asc, no filter).
//
// Response: {"data":[...],"page":N,"page_size":M,"total":T,"total_pages":P}
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	p := parsePageParams(r)
	filter := parseUserFilter(r)

	users, total, err := h.lister.ListPaginated(r.Context(), p.toModelPage(), filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, newPaginatedResponse(users, p, total))
}

// Get returns a single user by id.
//
// Contract:
//   - 200: the user as JSON.
//   - 400: {"error":"invalid id"} when the path id is not a valid integer.
//   - 404: {"error":"user not found"} when no user has that id.
//   - 500: {"error":"internal error"} on any other store failure.
//
// The request context is propagated end-to-end to the store so cancellations
// and deadlines reach the database query.
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	u, err := h.getter.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "user not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// Create inserts a new user.
// BUG: request body never closed (http.Request.Body leak).
// BUG: no input validation, no 201 status.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.crud.Create(&u); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u) // BUG: 200 instead of 201
}

// Delete removes a user.
// BUG: no distinction between not-found and server error.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := h.crud.DeleteByID(id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

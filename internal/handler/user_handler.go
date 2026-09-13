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

// errInvalidPagination is returned when page/page_size query params are invalid.
var errInvalidPagination = errors.New("invalid pagination parameters")

// UserLister is the read dependency the handler needs to serve paginated lists.
// Depending on an interface (rather than *store.UserStore) keeps the handler
// testable with a lightweight mock.
type UserLister interface {
	ListPaginated(ctx context.Context, page model.Page) ([]*model.User, int, error)
}

// UserHandler handles HTTP requests for users.
type UserHandler struct {
	store  *store.UserStore
	lister UserLister
}

func NewUserHandler(s *store.UserStore) *UserHandler {
	return &UserHandler{store: s, lister: s}
}

// List returns a paginated page of users.
//
// Query params:
//   - page:      1-based page number (default 1)
//   - page_size: items per page (default 20, capped at 100)
//
// Response: {"data":[...],"page":N,"page_size":M,"total":T}
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page, err := parsePage(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	users, total, err := h.lister.ListPaginated(r.Context(), page)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, model.PagedUsers{
		Data:     users,
		Page:     page.Number,
		PageSize: page.Size,
		Total:    total,
	})
}

// Get returns a single user.
// BUG: sql.ErrNoRows mapped to 500 instead of 404.
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	u, err := h.store.GetByID(id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError) // 404 missing
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
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
	if err := h.store.Create(&u); err != nil {
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
	if err := h.store.DeleteByID(id); err != nil {
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

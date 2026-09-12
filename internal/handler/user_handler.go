package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jaavier/go-api-service/internal/model"
	"github.com/jaavier/go-api-service/internal/store"
)

// UserHandler handles HTTP requests for users.
// BUG: depends on concrete *store.UserStore instead of an interface.
type UserHandler struct {
	store *store.UserStore
}

func NewUserHandler(s *store.UserStore) *UserHandler {
	return &UserHandler{store: s}
}

// List returns all users.
// BUG: no request context passed down, no error response body.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.List()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users) // BUG: encode error ignored
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

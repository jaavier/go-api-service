package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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

// UserCreator is the write dependency for the Create endpoint. Depending on an
// interface (rather than *store.UserStore) keeps the handler testable with a
// lightweight mock. *store.UserStore satisfies it via its Create method, which
// populates u.ID with the generated primary key.
type UserCreator interface {
	Create(ctx context.Context, u *model.User) error
}

// UserDeleter is the write dependency for the Delete endpoint. Depending on an
// interface (rather than *store.UserStore) keeps the handler testable with a
// lightweight mock. *store.UserStore satisfies it via its DeleteByID method.
type UserDeleter interface {
	DeleteByID(ctx context.Context, id int64) error
}

// UserHandler handles HTTP requests for users.
//
// lister serves the paginated List endpoint and can be a mock in tests.
// getter serves the Get endpoint and can be a mock in tests.
// creator serves the Create endpoint and can be a mock in tests.
// deleter serves the Delete endpoint and can be a mock in tests.
type UserHandler struct {
	lister  UserLister
	getter  UserGetter
	creator UserCreator
	deleter UserDeleter
}

// NewUserHandler wires the concrete *store.UserStore into every dependency the
// handler needs. The public signature is unchanged; the store satisfies each
// of the read/write interfaces.
func NewUserHandler(s *store.UserStore) *UserHandler {
	return &UserHandler{lister: s, getter: s, creator: s, deleter: s}
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
//
// Contract:
//   - 201: the created user as JSON, including the id assigned by the store.
//   - 400: {"error":"invalid body"} when the body is not valid JSON.
//   - 400: {"error":"name and email are required"} when name or email is blank.
//   - 500: {"error":"internal error"} on any store failure.
//
// The request body is always closed, and the request context is propagated
// end-to-end to the store so cancellations and deadlines reach the database.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(u.Name) == "" || strings.TrimSpace(u.Email) == "" {
		writeJSONError(w, http.StatusBadRequest, "name and email are required")
		return
	}
	if err := h.creator.Create(r.Context(), &u); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

// Delete removes a user by id.
//
// Contract:
//   - 204: empty body on success.
//   - 400: {"error":"invalid id"} when the path id is not a valid integer.
//   - 404: {"error":"user not found"} when no user has that id.
//   - 500: {"error":"internal error"} on any other store failure.
//
// The request context is propagated end-to-end to the store so cancellations
// and deadlines reach the database query.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.deleter.DeleteByID(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "user not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "internal error")
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

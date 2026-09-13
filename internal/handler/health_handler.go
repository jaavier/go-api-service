package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger is the minimal contract the health handler needs from a data store.
// Depending on an interface (instead of *sql.DB) keeps the handler testable
// and decoupled from the concrete database implementation.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// HealthHandler reports the liveness/readiness of the service and its
// dependencies.
type HealthHandler struct {
	db      Pinger
	timeout time.Duration
}

// NewHealthHandler builds a HealthHandler. A non-positive timeout falls back to
// a sensible default so the DB ping cannot hang a probe indefinitely.
func NewHealthHandler(db Pinger, timeout time.Duration) *HealthHandler {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &HealthHandler{db: db, timeout: timeout}
}

// healthResponse is the JSON payload returned by the /healthz endpoint.
type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
	Error  string `json:"error,omitempty"`
}

// Check handles GET /healthz. It pings the database within a bounded context
// and responds 200 when healthy or 503 when the dependency is unreachable.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	resp := healthResponse{Status: "ok", DB: "up"}
	status := http.StatusOK

	if err := h.db.PingContext(ctx); err != nil {
		resp.Status = "unavailable"
		resp.DB = "down"
		resp.Error = err.Error()
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Response is already committed; nothing actionable remains but we
		// avoid silently swallowing the error entirely.
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

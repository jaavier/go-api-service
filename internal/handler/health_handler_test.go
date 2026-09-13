package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// stubPinger is a test double for the Pinger interface.
type stubPinger struct {
	err error
}

func (s stubPinger) PingContext(ctx context.Context) error { return s.err }

func TestHealthHandler_Check_Healthy(t *testing.T) {
	h := NewHealthHandler(stubPinger{err: nil}, time.Second)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != "ok" || body.DB != "up" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestHealthHandler_Check_DBDown(t *testing.T) {
	h := NewHealthHandler(stubPinger{err: errors.New("connection refused")}, time.Second)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Check(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var body healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != "unavailable" || body.DB != "down" {
		t.Fatalf("unexpected body: %+v", body)
	}
	if body.Error == "" {
		t.Fatalf("expected error field to be populated")
	}
}

func TestNewHealthHandler_DefaultTimeout(t *testing.T) {
	h := NewHealthHandler(stubPinger{}, 0)
	if h.timeout <= 0 {
		t.Fatalf("expected default timeout to be applied, got %v", h.timeout)
	}
}

package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goth/internal/config"
	"goth/internal/handler"
)

// TestContactRouteRegistered guards the Phase 10a wiring: POST /api/contact
// must reach the handler (503 here because no mailer is configured) rather
// than falling through to a 404/405.
func TestContactRouteRegistered(t *testing.T) {
	h := handler.New(&config.Config{}, nil, nil, nil, nil)
	r := New(h)

	req := httptest.NewRequest(http.MethodPost, "/api/contact",
		strings.NewReader(`{"name":"Test Person","email":"visitor@example.com","message":"Hello, I would like to talk."}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/contact not registered: status %d", rec.Code)
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503 (unconfigured mailer)", rec.Code)
	}

	// GET must not be accepted on the endpoint.
	recGet := httptest.NewRecorder()
	r.ServeHTTP(recGet, httptest.NewRequest(http.MethodGet, "/api/contact", nil))
	if recGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /api/contact = %d, want 405", recGet.Code)
	}
}

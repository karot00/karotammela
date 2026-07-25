package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"goth/internal/config"
	"goth/internal/db"
)

func getStats(t *testing.T, h *Handlers, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/stats"+query, nil)
	rec := httptest.NewRecorder()
	h.Stats(rec, req)
	return rec
}

// TestStatsNoDatabase mirrors the Next.js 503 when TURSO_DATABASE_URL is
// missing: no configured database returns a service-unavailable error.
func TestStatsNoDatabase(t *testing.T) {
	h := &Handlers{cfg: &config.Config{}}
	rec := getStats(t, h, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body["error"] != "Stats database is not configured." {
		t.Errorf("error = %q", body["error"])
	}
}

// TestStatsInvalidLocale mirrors the zod querySchema: only en/fi (with fi as
// the default when absent) are accepted.
func TestStatsInvalidLocale(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "stats.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	h := &Handlers{cfg: &config.Config{}, conn: conn}

	rec := getStats(t, h, "?locale=sv")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body["error"] != "Invalid locale." {
		t.Errorf("error = %q", body["error"])
	}

	for _, q := range []string{"", "?locale=en", "?locale=fi"} {
		if rec := getStats(t, h, q); rec.Code != http.StatusOK {
			t.Errorf("query %q: status = %d, want 200", q, rec.Code)
		}
	}
}

// TestStatsPayload checks the response contract against the Next.js
// responseSchema: field names, semantics, and Date-style ISO serialization.
func TestStatsPayload(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "stats.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	h := &Handlers{cfg: &config.Config{}, conn: conn}

	// Empty database: zeros and a JSON null latestUnlockAt.
	rec := getStats(t, h, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var empty map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &empty); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	for _, key := range []string{"totalAttempts", "unlockedCount", "directUnlockCount", "highestLevel", "avgMessagesToUnlock", "latestUnlockAt"} {
		if _, ok := empty[key]; !ok {
			t.Errorf("missing response field %q", key)
		}
	}
	if empty["latestUnlockAt"] != nil {
		t.Errorf("latestUnlockAt = %v, want null", empty["latestUnlockAt"])
	}

	// Seed: one three-message session ending in a direct unlock.
	for _, turn := range []struct {
		input   string
		level   int
		success bool
	}{
		{"first attempt", 20, false},
		{"second attempt with reasoning", 45, false},
		{"protocol_k_2026", 100, true},
	} {
		if err := db.PersistTurn(conn, "sess-1", "fi", turn.input, "out", turn.level, turn.success); err != nil {
			t.Fatalf("PersistTurn: %v", err)
		}
	}
	if _, err := conn.Exec(`UPDATE logs SET timestamp = 1753350000000 WHERE success = 1`); err != nil {
		t.Fatalf("set timestamp: %v", err)
	}

	rec = getStats(t, h, "?locale=en")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got struct {
		TotalAttempts       int     `json:"totalAttempts"`
		UnlockedCount       int     `json:"unlockedCount"`
		DirectUnlockCount   int     `json:"directUnlockCount"`
		HighestLevel        int     `json:"highestLevel"`
		AvgMessagesToUnlock float64 `json:"avgMessagesToUnlock"`
		LatestUnlockAt      *string `json:"latestUnlockAt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if got.TotalAttempts != 3 || got.UnlockedCount != 1 || got.DirectUnlockCount != 1 || got.HighestLevel != 100 {
		t.Errorf("counters = %+v", got)
	}
	if got.AvgMessagesToUnlock != 3 {
		t.Errorf("avgMessagesToUnlock = %v, want 3", got.AvgMessagesToUnlock)
	}
	if got.LatestUnlockAt == nil {
		t.Fatalf("latestUnlockAt = nil, want ISO string")
	}
	if *got.LatestUnlockAt != "2025-07-24T09:40:00.000Z" {
		t.Errorf("latestUnlockAt = %q, want 2025-07-24T09:40:00.000Z", *got.LatestUnlockAt)
	}
}

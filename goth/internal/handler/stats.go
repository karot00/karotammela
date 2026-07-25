package handler

import (
	"encoding/json"
	"net/http"

	"goth/internal/ai"
	"goth/internal/db"
)

// statsResponse mirrors the zod responseSchema in src/app/api/stats/route.ts.
// LatestUnlockAt serializes like NextResponse.json serializes a Date: an
// ISO-8601 millisecond UTC string, or null when no unlock exists.
type statsResponse struct {
	TotalAttempts       int     `json:"totalAttempts"`
	UnlockedCount       int     `json:"unlockedCount"`
	DirectUnlockCount   int     `json:"directUnlockCount"`
	HighestLevel        int     `json:"highestLevel"`
	AvgMessagesToUnlock float64 `json:"avgMessagesToUnlock"`
	LatestUnlockAt      *string `json:"latestUnlockAt"`
}

// Stats handles GET /api/stats. Port of src/app/api/stats/route.ts:
// 503 when no database is configured, 400 for an invalid locale query
// value ("fi" is the default), 500 when reading fails, otherwise the
// aggregate payload defined by the Next.js response schema.
func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.conn == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Stats database is not configured."})
		return
	}

	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "fi"
	}
	if locale != "en" && locale != "fi" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid locale."})
		return
	}

	stats, err := db.GetStats(h.conn, ai.GetAccessCode())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read stats."})
		return
	}

	resp := statsResponse{
		TotalAttempts:       stats.TotalAttempts,
		UnlockedCount:       stats.UnlockedCount,
		DirectUnlockCount:   stats.DirectUnlockCount,
		HighestLevel:        stats.HighestLevel,
		AvgMessagesToUnlock: stats.AvgMessagesToUnlock,
	}
	if stats.LatestUnlock != nil {
		iso := stats.LatestUnlock.UTC().Format("2006-01-02T15:04:05.000Z07:00")
		resp.LatestUnlockAt = &iso
	}
	json.NewEncoder(w).Encode(resp)
}

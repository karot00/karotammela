package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"goth/internal/aipulse"
)

// refreshSourceJSON mirrors one source entry in the Next.js refresh response.
// Error is nil on success; with ?debug=1 it carries the raw error string,
// otherwise the redacted reference strings ("Fetch failed"/"DB insert failed").
type refreshSourceJSON struct {
	OK          bool    `json:"ok"`
	Inserted    int     `json:"inserted"`
	SkippedDup  *int    `json:"skippedDup,omitempty"`
	WindowHours *int    `json:"windowHours,omitempty"`
	Error       *string `json:"error"`
}

// refreshResponseJSON mirrors the Next.js route payload:
// { ok, ranAt, sources: { trends, repos, stocks } }.
type refreshResponseJSON struct {
	OK      bool   `json:"ok"`
	RanAt   string `json:"ranAt"`
	Sources struct {
		Trends refreshSourceJSON `json:"trends"`
		Repos  refreshSourceJSON `json:"repos"`
		Stocks refreshSourceJSON `json:"stocks"`
	} `json:"sources"`
}

func sourceJSON(res aipulse.RefreshSourceResult, debug bool, withTrendStats bool) refreshSourceJSON {
	out := refreshSourceJSON{OK: res.OK, Inserted: res.Inserted}
	if withTrendStats {
		skips, window := res.SkippedDup, res.WindowHours
		out.SkippedDup = &skips
		out.WindowHours = &window
	}
	if res.OK {
		return out
	}
	var msg string
	if debug {
		msg = res.ErrDetail
	} else if res.ErrKind == "db" {
		msg = "DB insert failed"
	} else {
		msg = "Fetch failed"
	}
	out.Error = &msg
	return out
}

// AiPulseRefresh handles POST /api/ai-pulse/refresh (Phase 12.5d), the Go port
// of src/app/api/ai-pulse/refresh/route.ts. Guarded by CRON_SECRET (Bearer),
// overlap-safe (409 while a run is active in this process), bounded by a
// 4-minute context, and always per-source isolated in its response.
func (h *Handlers) AiPulseRefresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Mirror the reference guard exactly: no secret configured, or a Bearer
	// mismatch, is the same 401.
	expected := h.cfg.CronSecret
	if expected == "" || r.Header.Get("Authorization") != "Bearer "+expected {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	if h.conn == nil || h.refresher == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "AI Pulse refresh is not configured."})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Minute)
	defer cancel()
	report, started := h.refresher.TryRun(ctx, h.conn)
	if !started {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Refresh already in progress."})
		return
	}

	debug := r.URL.Query().Get("debug") == "1"
	resp := refreshResponseJSON{
		OK:    true,
		RanAt: report.RanAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
	resp.Sources.Trends = sourceJSON(report.Trends, debug, true)
	resp.Sources.Repos = sourceJSON(report.Repos, debug, false)
	resp.Sources.Stocks = sourceJSON(report.Stocks, debug, false)
	json.NewEncoder(w).Encode(resp)
}

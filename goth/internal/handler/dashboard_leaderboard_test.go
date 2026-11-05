package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goth/internal/config"
	"goth/internal/security"
	"goth/internal/view"
)

// TestDashboardModelLeaderboard asserts the "who" section renders the ranked
// model list in order, with its notes, and that the removed H1/H2 copy is gone.
func TestDashboardModelLeaderboard(t *testing.T) {
	vr, err := view.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cfg := &config.Config{UnlockCookieSecret: "testsecret", BaseURL: "https://karotammela.fi"}
	h := New(cfg, vr, newTestDB(t), nil, nil)
	cookieVal := security.CreateUnlockCookieValue(security.UnlockPayload{SessionID: "x", Locale: "fi", UnlockedAt: 1753352400}, "testsecret")

	req := httptest.NewRequest(http.MethodGet, "/fi/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: "karot_unlock", Value: cookieVal})
	w := httptest.NewRecorder()
	h.Dashboard(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	out := w.Body.String()

	last := -1
	for _, want := range []string{
		"Kielimallien leaderboardini",
		"<li>Opus 5.5<span class=\"text-muted-foreground\"> – selkeä ykkönen ja pääkielimallini</span></li>",
		"<li>Sonnet 5.5</li>",
		"<li>Gemini 4 Argon</li>",
		"<li>GPT 6 Sol</li>",
		"<li>Qwen 3.8 27B<span class=\"text-muted-foreground\"> – paras paikallinen malli</span></li>",
		"<li>Sonnet 5</li>",
		"<li>GPT 5.6</li>",
		"<li>Qwen 3.7 Plus/Max</li>",
		"<li>Hy3</li>",
		"<li>Gemini 3.1 Flash Lite</li>",
	} {
		i := strings.Index(out, want)
		if i < 0 {
			t.Fatalf("output missing %q", want)
		}
		if i < last {
			t.Errorf("%q rendered out of order", want)
		}
		last = i
	}
	for _, gone := range []string{"H1 / 2026", "H2 / 2026", "Opus 4.7", "Koodaukseen:"} {
		if strings.Contains(out, gone) {
			t.Errorf("output still contains %q", gone)
		}
	}
}

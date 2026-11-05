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

// TestDashboardProjects asserts the projects view lists each new project with
// its description and link. Prosessistudio runs locally, so it has no link.
func TestDashboardProjects(t *testing.T) {
	vr, err := view.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cfg := &config.Config{UnlockCookieSecret: "testsecret", BaseURL: "https://karotammela.fi"}
	h := New(cfg, vr, newTestDB(t), nil, nil)
	cookieVal := security.CreateUnlockCookieValue(security.UnlockPayload{SessionID: "x", Locale: "fi", UnlockedAt: 1753352400}, "testsecret")

	req := httptest.NewRequest(http.MethodGet, "/fi/dashboard?view=projects", nil)
	req.AddCookie(&http.Cookie{Name: "karot_unlock", Value: cookieVal})
	w := httptest.NewRecorder()
	h.Dashboard(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	out := w.Body.String()

	for _, want := range []string{
		"Chalet Nulkan suoravaraussivusto",
		"Levin Hiihtogreenin vuokramökkisivusto",
		"Levi Now (map.levifinland.fi)",
		"UTM-linkkityökalu markkinointitiimeille",
		"BPMN-prosessityökalu työpajoihin",
		"Skibus Rally",
		`href="https://experiencelevi.fi"`,
		`href="https://hiihtogreeni.fi"`,
		`href="https://map.levifinland.fi"`,
		`href="https://kanavoi.com"`,
		`href="https://game.levifinland.fi"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("projects view missing %q", want)
		}
	}
	if strings.Contains(out, "prosessistudio.") || strings.Contains(out, "github.com/karot00/prosessistudio") {
		t.Error("Prosessistudio must not link anywhere")
	}
	if strings.Contains(out, "postikortti") {
		t.Error("projects view must not list the postcard")
	}
}

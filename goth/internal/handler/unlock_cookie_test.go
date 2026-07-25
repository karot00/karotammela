package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goth/internal/config"
)

// TestUnlockCookieAttributes pins the karot_unlock Set-Cookie flags to the
// shared 12.5f contract (shared/security/unlock-cookie-vectors.json spec):
// Path=/, HttpOnly, SameSite=Lax, Max-Age=1209600, Secure in production.
func TestUnlockCookieAttributes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cfg      *config.Config
		wantSafe string
	}{
		{"production", &config.Config{UnlockCookieSecret: "s", Env: "production"}, "; Secure"},
		{"development", &config.Config{UnlockCookieSecret: "s", Env: "development"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handlers{cfg: tc.cfg}
			req := httptest.NewRequest(http.MethodPost, "/api/sentinel", strings.NewReader(
				`{"sessionId":"s1","locale":"fi","currentLevel":0,"messages":[{"role":"user","content":"PROTOCOL_K_2026"}]}`))
			w := httptest.NewRecorder()
			h.SentinelCommit(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			setCookie := w.Header().Get("Set-Cookie")
			if setCookie == "" {
				t.Fatal("no Set-Cookie header")
			}
			for _, want := range []string{
				"karot_unlock=", "; Path=/", "; HttpOnly", "; SameSite=Lax", "; Max-Age=1209600",
			} {
				if !strings.Contains(setCookie, want) {
					t.Errorf("Set-Cookie missing %q: %s", want, setCookie)
				}
			}
			hasSecure := strings.Contains(setCookie, "; Secure")
			if (tc.wantSafe != "") != hasSecure {
				t.Errorf("Secure flag = %v, want %v in %s", hasSecure, tc.wantSafe != "", tc.name)
			}
		})
	}
}

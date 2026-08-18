package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goth/internal/config"
	"goth/internal/handler"
	"goth/internal/security"
	"goth/internal/view"
)

const vipTestUnlockSecret = "vip-router-test-unlock-secret"

// vipRouter builds the full production router with the VIP flag toggled.
func vipRouter(t *testing.T, enabled bool) http.Handler {
	t.Helper()
	vr, err := view.NewRenderer()
	if err != nil {
		t.Fatalf("view.NewRenderer() = %v", err)
	}
	cfg := &config.Config{
		VIPEnabled:         enabled,
		BaseURL:            "https://karotammela.fi",
		NextURL:            "https://next.karotammela.fi",
		UnlockCookieSecret: vipTestUnlockSecret,
	}
	return New(handler.New(cfg, vr, nil, nil, nil))
}

func vipGet(t *testing.T, rt http.Handler, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	rt.ServeHTTP(rec, req)
	return rec
}

func TestVIPRoutingDisabled(t *testing.T) {
	rt := vipRouter(t, false)

	for _, path := range []string{"/vip", "/fi/vip", "/en/vip"} {
		rec := vipGet(t, rt, path)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s disabled = %d, want 404", path, rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "" {
			t.Errorf("%s disabled Location = %q, want none", path, loc)
		}
	}

	rec := vipGet(t, rt, "/api/vip/status")
	if rec.Code != http.StatusOK {
		t.Fatalf("status disabled = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"enabled":false`) {
		t.Errorf("status disabled body = %s", body)
	}
}

func TestVIPRoutingEnabled(t *testing.T) {
	rt := vipRouter(t, true)

	for _, path := range []string{"/vip", "/fi/vip"} {
		rec := vipGet(t, rt, path)
		if rec.Code != http.StatusSeeOther {
			t.Errorf("%s enabled = %d, want 303", path, rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "/en/vip" {
			t.Errorf("%s enabled Location = %q, want /en/vip", path, loc)
		}
	}

	rec := vipGet(t, rt, "/en/vip")
	if rec.Code != http.StatusOK {
		t.Fatalf("/en/vip enabled = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `lang="en"`) {
		t.Error("/en/vip enabled did not render the English VIP document")
	}

	rec = vipGet(t, rt, "/api/vip/status")
	body := rec.Body.String()
	if !strings.Contains(body, `"enabled":true`) || !strings.Contains(body, "https://karotammela.fi/en/vip") {
		t.Errorf("status enabled body = %s", body)
	}
}

// TestVIPHeaderLinkVisibility covers the Phase 1 gate that no navigation shows
// VIP while disabled, and both locales link to the single English route while
// enabled.
func TestVIPHeaderLinkVisibility(t *testing.T) {
	disabled := vipRouter(t, false)
	for _, path := range []string{"/en", "/fi"} {
		rec := vipGet(t, disabled, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s disabled = %d, want 200", path, rec.Code)
		}
		if strings.Contains(rec.Body.String(), `href="/en/vip"`) {
			t.Errorf("%s disabled links to /en/vip", path)
		}
	}

	enabled := vipRouter(t, true)
	for _, path := range []string{"/en", "/fi"} {
		rec := vipGet(t, enabled, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s enabled = %d, want 200", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `href="/en/vip"`) {
			t.Errorf("%s enabled does not link to /en/vip", path)
		}
	}
}

// TestVIPDashboardLinkVisibility checks the conditional VIP anchor inside the
// unlocked Go dashboard navigation (desktop sidebar and mobile menu share the
// dashnav partial).
func TestVIPDashboardLinkVisibility(t *testing.T) {
	unlock := &http.Cookie{
		Name: "karot_unlock",
		Value: security.CreateUnlockCookieValue(security.UnlockPayload{
			SessionID:  "vip-router-test",
			Locale:     "en",
			UnlockedAt: 1,
		}, vipTestUnlockSecret),
	}

	disabled := vipRouter(t, false)
	rec := vipGet(t, disabled, "/en/dashboard", unlock)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard disabled = %d, want 200", rec.Code)
	}
	if strings.Contains(rec.Body.String(), `href="/en/vip"`) {
		t.Error("dashboard disabled links to /en/vip")
	}

	enabled := vipRouter(t, true)
	rec = vipGet(t, enabled, "/en/dashboard", unlock)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard enabled = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `href="/en/vip"`) {
		t.Error("dashboard enabled does not link to /en/vip")
	}
}

func TestRobotsDisallowsVIPPaths(t *testing.T) {
	rt := vipRouter(t, false)
	rec := vipGet(t, rt, "/robots.txt")
	if rec.Code != http.StatusOK {
		t.Fatalf("robots.txt = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, rule := range []string{
		"Disallow: /vip",
		"Disallow: /en/vip",
		"Disallow: /fi/vip",
		"Disallow: /api/vip/",
	} {
		if !strings.Contains(body, rule) {
			t.Errorf("robots.txt missing %q", rule)
		}
	}
}

// TestSitemapNeverContainsVIP asserts the sitemap is unchanged by the feature
// flag in both states (plan §4.5).
func TestSitemapNeverContainsVIP(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		rt := vipRouter(t, enabled)
		rec := vipGet(t, rt, "/sitemap.xml")
		if rec.Code != http.StatusOK {
			t.Fatalf("sitemap (enabled=%v) = %d, want 200", enabled, rec.Code)
		}
		if strings.Contains(strings.ToLower(rec.Body.String()), "vip") {
			t.Errorf("sitemap (enabled=%v) contains VIP urls", enabled)
		}
	}
}

// TestExistingLocaleRoutesSurviveVIPNodes guards against the static /en and
// /fi trie nodes (added for /en/vip and /fi/vip) shadowing the /{locale}
// wildcard routes.
func TestExistingLocaleRoutesSurviveVIPNodes(t *testing.T) {
	rt := vipRouter(t, true)

	rec := vipGet(t, rt, "/fi/blog")
	if rec.Code != http.StatusOK {
		t.Errorf("/fi/blog = %d, want 200", rec.Code)
	}
	rec = vipGet(t, rt, "/en/privacy")
	if rec.Code != http.StatusOK {
		t.Errorf("/en/privacy = %d, want 200", rec.Code)
	}
	rec = vipGet(t, rt, "/")
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/fi" {
		t.Errorf("/ = %d Location=%q, want 303 /fi", rec.Code, rec.Header().Get("Location"))
	}
}

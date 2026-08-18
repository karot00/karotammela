package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goth/internal/config"
	"goth/internal/view"
)

// vipTestHandlers builds a handler set with only the fields the VIP routes
// read. enabled toggles the single runtime flag under test.
func vipTestHandlers(t *testing.T, enabled bool) *Handlers {
	t.Helper()
	vr, err := view.NewRenderer()
	if err != nil {
		t.Fatalf("view.NewRenderer() = %v", err)
	}
	return &Handlers{
		cfg: &config.Config{
			VIPEnabled: enabled,
			BaseURL:    "https://karotammela.fi",
		},
		view: vr,
	}
}

func vipGet(t *testing.T, h *Handlers, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	switch path {
	case "/api/vip/status":
		h.VIPStatus(rec, req)
	case "/en/vip":
		h.VIPPage(rec, req)
	default:
		h.VIPEntry(rec, req)
	}
	return rec
}

func TestVIPStatusDisabled(t *testing.T) {
	h := vipTestHandlers(t, false)
	rec := vipGet(t, h, "/api/vip/status")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"enabled":false}` {
		t.Errorf("body = %q, want {\"enabled\":false}", got)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
	if xr := rec.Header().Get("X-Robots-Tag"); xr != vipNoIndex {
		t.Errorf("X-Robots-Tag = %q, want %q", xr, vipNoIndex)
	}
}

func TestVIPStatusEnabled(t *testing.T) {
	h := vipTestHandlers(t, true)
	rec := vipGet(t, h, "/api/vip/status")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"enabled":true`) {
		t.Errorf("body = %q, want enabled:true", body)
	}
	if !strings.Contains(body, `"url":"https://karotammela.fi/en/vip"`) {
		t.Errorf("body = %q, want canonical url", body)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
}

// TestVIPDisabledRoutesAreIndistinguishable404s is the Phase 1 gate: with the
// flag off, /vip, /fi/vip and /en/vip all return plain 404s that do not
// redirect to a login or reveal that a VIP feature exists.
func TestVIPDisabledRoutesAreIndistinguishable404s(t *testing.T) {
	h := vipTestHandlers(t, false)
	for _, path := range []string{"/vip", "/fi/vip", "/en/vip"} {
		rec := vipGet(t, h, path)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s disabled = %d, want 404", path, rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "" {
			t.Errorf("%s disabled set Location %q, want none", path, loc)
		}
		if strings.Contains(rec.Body.String(), "vip") && strings.Contains(rec.Body.String(), "login") {
			t.Errorf("%s disabled body leaks VIP/login hints", path)
		}
	}
}

func TestVIPEntryRedirectsWhenEnabled(t *testing.T) {
	h := vipTestHandlers(t, true)
	for _, path := range []string{"/vip", "/fi/vip"} {
		rec := vipGet(t, h, path)
		if rec.Code != http.StatusSeeOther {
			t.Errorf("%s enabled = %d, want 303", path, rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "/en/vip" {
			t.Errorf("%s Location = %q, want /en/vip", path, loc)
		}
	}
}

func TestVIPPageEnabledRendersEnglishDocument(t *testing.T) {
	h := vipTestHandlers(t, true)
	rec := vipGet(t, h, "/en/vip")

	if rec.Code != http.StatusOK {
		t.Fatalf("/en/vip enabled = %d, want 200", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
	if xr := rec.Header().Get("X-Robots-Tag"); xr != vipNoIndex {
		t.Errorf("X-Robots-Tag = %q, want %q", xr, vipNoIndex)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `lang="en"`) {
		t.Error("VIP page missing lang=\"en\"")
	}
	if strings.Contains(body, "hreflang") {
		t.Error("VIP page must not emit hreflang alternates")
	}
	if strings.Contains(body, "application/ld+json") {
		t.Error("VIP page must not emit JSON-LD")
	}
	if !strings.Contains(body, "Invitation required") || !strings.Contains(body, "Back to main page") {
		t.Error("restricted page missing neutral invitation copy or main-page navigation")
	}
	for _, leakedPurpose := range []string{"MeetingPackage", "job application", "private application"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(leakedPurpose)) {
			t.Errorf("unauthenticated page leaks purpose %q", leakedPurpose)
		}
	}
}

func TestVIPPageDisabledIs404(t *testing.T) {
	h := vipTestHandlers(t, false)
	rec := vipGet(t, h, "/en/vip")
	if rec.Code != http.StatusNotFound {
		t.Errorf("/en/vip disabled = %d, want 404", rec.Code)
	}
}

package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/argon2"

	"goth/internal/config"
	"goth/internal/content"
	"goth/internal/security"
	"goth/internal/view"
)

// vipAccessHandlers builds a fully wired handler set for the Phase 2 access
// flow: enabled VIP, a real argon2id hash of vipTestCode, a cookie secret, a
// mock mailer, a fresh throttle, and a zero timing floor so tests run fast.
func vipAccessHandlers(t *testing.T, mailer MailSender) *Handlers {
	t.Helper()
	vr, err := view.NewRenderer()
	if err != nil {
		t.Fatalf("view.NewRenderer() = %v", err)
	}
	vip, err := content.LoadVIP("../../content/vip")
	if err != nil {
		t.Fatalf("content.LoadVIP() = %v", err)
	}
	return &Handlers{
		cfg: &config.Config{
			VIPEnabled:       true,
			VIPPasswordHash:  vipTestArgon2Hash(t, vipTestCode),
			VIPCookieSecret:  "vip-access-test-cookie-secret-0123456789",
			ContactFromEmail: "info@karotammela.fi",
			ContactToEmail:   "karo@karotammela.fi",
			BaseURL:          "https://karotammela.fi",
			Env:              "development",
		},
		view:          vr,
		mailer:        mailer,
		vipThrottle:   security.NewVIPLoginThrottle(),
		vipLoginFloor: 0,
		vipContent:    vip,
	}
}

const vipTestCode = "invitation-code-xyz"

func vipTestArgon2Hash(t *testing.T, password string) string {
	t.Helper()
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}
	key := argon2.Key([]byte(password), salt, 1, 1024, 2, 32)
	return fmt.Sprintf("$argon2id$v=19$m=1024,t=1,p=2$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key))
}

func vipPost(t *testing.T, h *Handlers, path string, form url.Values, ip string, hdr http.Header) *httptest.ResponseRecorder {
	t.Helper()
	if ip == "" {
		ip = "203.0.113.50"
	}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", ip)
	req.RemoteAddr = "127.0.0.1:12345"
	for k, vs := range hdr {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	rec := httptest.NewRecorder()
	switch path {
	case "/api/vip/notify":
		h.VIPNotify(rec, req)
	case "/api/vip/login":
		h.VIPLogin(rec, req)
	case "/api/vip/logout":
		h.VIPLogout(rec, req)
	}
	return rec
}

// --- Notify ---

func TestVIPNotifyDisabled(t *testing.T) {
	h := vipAccessHandlers(t, &mockMailer{})
	h.cfg.VIPEnabled = false
	rec := vipPost(t, h, "/api/vip/notify", url.Values{"email": {"a@b.co"}}, "203.0.113.60", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("disabled notify = %d, want 404", rec.Code)
	}
}

func TestVIPNotifySuccessSendsAndReturnsPasswordPartial(t *testing.T) {
	mailer := &mockMailer{}
	h := vipAccessHandlers(t, mailer)
	rec := vipPost(t, h, "/api/vip/notify", url.Values{"email": {"  Recruiter@Example.com "}}, "203.0.113.61", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("notify = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "vip-code") {
		t.Error("notify did not return the password partial")
	}
	if strings.Contains(rec.Body.String(), "role=\"alert\"") {
		t.Error("successful notify should not render an error alert")
	}
	if len(mailer.sent) != 1 {
		t.Fatalf("deliveries = %d, want 1", len(mailer.sent))
	}
	msg := mailer.sent[0]
	if msg.To != "karo@karotammela.fi" {
		t.Errorf("To = %q, want private recipient", msg.To)
	}
	if msg.ReplyTo != "recruiter@example.com" {
		t.Errorf("ReplyTo = %q, want lowercased trimmed visitor email", msg.ReplyTo)
	}
	if !strings.Contains(msg.Text, "recruiter@example.com") {
		t.Error("notification body missing the visitor email")
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
}

// TestVIPNotifyOutcomesIndistinguishable proves honeypot, invalid-email,
// mailer-failure and mailer-missing all return the exact same generic partial
// as a successful delivery, so delivery state is never revealed (plan §6.1).
func TestVIPNotifyOutcomesIndistinguishable(t *testing.T) {
	cases := []struct {
		name   string
		mailer MailSender
		form   url.Values
	}{
		{"success", &mockMailer{}, url.Values{"email": {"a@b.co"}}},
		{"honeypot", &mockMailer{}, url.Values{"email": {"a@b.co"}, "website": {"spam"}}},
		{"invalid email", &mockMailer{}, url.Values{"email": {"not-an-email"}}},
		{"mailer failure", &mockMailer{err: fmt.Errorf("resend: unexpected status 500")}, url.Values{"email": {"a@b.co"}}},
		{"mailer missing", nil, url.Values{"email": {"a@b.co"}}},
	}

	bodies := map[string]string{}
	for i, tc := range cases {
		h := vipAccessHandlers(t, tc.mailer)
		rec := vipPost(t, h, "/api/vip/notify", tc.form, "198.51.100."+fmt.Sprint(100+i), nil)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", tc.name, rec.Code)
		}
		bodies[tc.name] = rec.Body.String()
	}
	ref := bodies["success"]
	for name, body := range bodies {
		if body != ref {
			t.Errorf("%s response differs from success (leaks delivery state)", name)
		}
	}
}

func TestVIPNotifyHoneypotDoesNotSend(t *testing.T) {
	mailer := &mockMailer{}
	h := vipAccessHandlers(t, mailer)
	rec := vipPost(t, h, "/api/vip/notify", url.Values{"email": {"a@b.co"}, "website": {"bot"}}, "203.0.113.62", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if len(mailer.sent) != 0 {
		t.Errorf("honeypot triggered %d deliveries, want 0", len(mailer.sent))
	}
}

func TestVIPNotifyIPRateLimit(t *testing.T) {
	h := vipAccessHandlers(t, &mockMailer{})
	ip := "203.0.113.63"
	var rec *httptest.ResponseRecorder
	for i := 0; i < vipNotifyIPLimit; i++ {
		rec = vipPost(t, h, "/api/vip/notify", url.Values{"email": {fmt.Sprintf("u%d@b.co", i)}}, ip, nil)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("notify rate-limited early at attempt %d", i+1)
		}
	}
	rec = vipPost(t, h, "/api/vip/notify", url.Values{"email": {"over@b.co"}}, ip, nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("6th notify = %d, want 429", rec.Code)
	}
	if ra := rec.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Errorf("Retry-After = %q, want positive seconds", ra)
	}
}

func TestVIPNotifyEmailHashRateLimit(t *testing.T) {
	h := vipAccessHandlers(t, &mockMailer{})
	// Same email from distinct IPs: the email-hash bucket (3/hour) fires, not
	// the per-IP bucket.
	var rec *httptest.ResponseRecorder
	for i := 0; i < vipNotifyEmailLimit; i++ {
		rec = vipPost(t, h, "/api/vip/notify", url.Values{"email": {"same@b.co"}}, fmt.Sprintf("192.0.2.%d", 200+i), nil)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("email rate-limited early at attempt %d", i+1)
		}
	}
	rec = vipPost(t, h, "/api/vip/notify", url.Values{"email": {"same@b.co"}}, "192.0.2.250", nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("4th notify for same email = %d, want 429", rec.Code)
	}
}

func TestVIPNotifyGlobalRateLimit(t *testing.T) {
	security.SetRateLimitStore(nil)
	t.Cleanup(func() { security.SetRateLimitStore(nil) })
	h := vipAccessHandlers(t, &mockMailer{})
	var rec *httptest.ResponseRecorder
	for i := 0; i < vipNotifyGlobalLimit; i++ {
		rec = vipPost(t, h, "/api/vip/notify", url.Values{
			"email": {fmt.Sprintf("global-%d@example.com", i)},
		}, fmt.Sprintf("198.51.100.%d", i%250), nil)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("global notify rate-limited early at attempt %d", i+1)
		}
	}
	rec = vipPost(t, h, "/api/vip/notify", url.Values{"email": {"global-over@example.com"}}, "203.0.113.240", nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("global notify attempt %d = %d, want 429", vipNotifyGlobalLimit+1, rec.Code)
	}
	if ra := rec.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Errorf("Retry-After = %q, want positive seconds", ra)
	}
}

// --- Login ---

func TestVIPLoginDisabled(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	h.cfg.VIPEnabled = false
	rec := vipPost(t, h, "/api/vip/login", url.Values{"code": {vipTestCode}}, "203.0.113.70", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("disabled login = %d, want 404", rec.Code)
	}
}

func TestVIPLoginSuccessSetsCookieAndRendersPortal(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	rec := vipPost(t, h, "/api/vip/login", url.Values{"code": {vipTestCode}}, "203.0.113.71", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("login = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "vipIntro") && !strings.Contains(rec.Body.String(), "x-data") {
		t.Error("login success did not render the portal shell")
	}

	cookies := rec.Result().Cookies()
	var vip *http.Cookie
	for _, c := range cookies {
		if c.Name == security.VIPCookieName {
			vip = c
		}
	}
	if vip == nil {
		t.Fatal("login success did not set the karot_vip cookie")
	}
	if !vip.HttpOnly {
		t.Error("VIP cookie missing HttpOnly")
	}
	if vip.SameSite != http.SameSiteLaxMode {
		t.Errorf("VIP cookie SameSite = %v, want Lax", vip.SameSite)
	}
	if vip.Path != "/" {
		t.Errorf("VIP cookie Path = %q, want /", vip.Path)
	}
	if vip.MaxAge != int(security.VIPCookieLifetime.Seconds()) {
		t.Errorf("VIP cookie MaxAge = %d, want %d", vip.MaxAge, int(security.VIPCookieLifetime.Seconds()))
	}
	if vip.Secure {
		t.Error("VIP cookie Secure in development, want false")
	}
	if security.VerifyVIPCookieValue(vip.Value, h.cfg.VIPCookieSecret, nowRef()) == nil {
		// Sanity: the minted cookie verifies.
		t.Error("minted VIP cookie failed verification")
	}
}

func TestVIPLoginSecureFlagInProduction(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	h.cfg.Env = "production"
	rec := vipPost(t, h, "/api/vip/login", url.Values{"code": {vipTestCode}}, "203.0.113.72", nil)
	var vip *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == security.VIPCookieName {
			vip = c
		}
	}
	if vip == nil {
		t.Fatal("no VIP cookie")
	}
	if !vip.Secure {
		t.Error("VIP cookie missing Secure in production")
	}
}

// TestVIPLoginFailuresGeneric proves wrong and malformed codes produce the same
// generic error partial and no cookie (plan §6.2).
func TestVIPLoginFailuresGeneric(t *testing.T) {
	bodies := map[string]string{}
	for name, code := range map[string]string{
		"wrong code": "definitely-wrong",
		"empty code": "",
		"oversized":  strings.Repeat("x", vipCodeMaxLen+1),
		"whitespace": "   ",
	} {
		h := vipAccessHandlers(t, nil)
		rec := vipPost(t, h, "/api/vip/login", url.Values{"code": {code}}, "198.51.100."+fmt.Sprint(150+len(bodies)), nil)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want 200 generic partial", name, rec.Code)
		}
		for _, c := range rec.Result().Cookies() {
			if c.Name == security.VIPCookieName && c.MaxAge > 0 {
				t.Errorf("%s: set a VIP cookie on failure", name)
			}
		}
		if !strings.Contains(rec.Body.String(), "could not be verified") {
			t.Errorf("%s: missing generic error copy", name)
		}
		bodies[name] = rec.Body.String()
	}
	ref := bodies["wrong code"]
	for name, body := range bodies {
		if body != ref {
			t.Errorf("%s failure body differs from wrong-code body (leaks validation detail)", name)
		}
	}
}

func TestVIPLoginRateLimitEscalatesTo429(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	ip := "203.0.113.73"
	var rec *httptest.ResponseRecorder
	for i := 0; i < 5; i++ {
		rec = vipPost(t, h, "/api/vip/login", url.Values{"code": {"wrong"}}, ip, nil)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("login rate-limited early at attempt %d", i+1)
		}
	}
	rec = vipPost(t, h, "/api/vip/login", url.Values{"code": {"wrong"}}, ip, nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("6th login = %d, want 429", rec.Code)
	}
	if ra := rec.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Errorf("Retry-After = %q, want positive seconds", ra)
	}
	// Even the correct code is blocked while cooling down.
	rec = vipPost(t, h, "/api/vip/login", url.Values{"code": {vipTestCode}}, ip, nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("correct code during cooldown = %d, want 429", rec.Code)
	}
}

func TestVIPLoginCrossSiteOriginRejected(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	hdr := http.Header{"Sec-Fetch-Site": {"cross-site"}}
	rec := vipPost(t, h, "/api/vip/login", url.Values{"code": {vipTestCode}}, "203.0.113.74", hdr)
	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-site login = %d, want 403", rec.Code)
	}
}

func TestVIPLoginOriginHeaderMismatchRejected(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	hdr := http.Header{"Origin": {"https://evil.example.com"}}
	rec := vipPost(t, h, "/api/vip/login", url.Values{"code": {vipTestCode}}, "203.0.113.75", hdr)
	if rec.Code != http.StatusForbidden {
		t.Errorf("mismatched Origin login = %d, want 403", rec.Code)
	}
}

func TestVIPLoginSameOriginAccepted(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	hdr := http.Header{
		"Sec-Fetch-Site": {"same-origin"},
		"Origin":         {"https://karotammela.fi"},
	}
	rec := vipPost(t, h, "/api/vip/login", url.Values{"code": {vipTestCode}}, "203.0.113.76", hdr)
	if rec.Code != http.StatusOK {
		t.Errorf("same-origin login = %d, want 200", rec.Code)
	}
}

// --- Logout ---

func TestVIPLogoutExpiresCookieAndRedirectsHTMXToHome(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	cookie := security.CreateVIPCookieValue(h.cfg.VIPCookieSecret, nowRef())

	req := httptest.NewRequest(http.MethodPost, "/api/vip/logout", nil)
	req.AddCookie(&http.Cookie{Name: security.VIPCookieName, Value: cookie})
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	h.VIPLogout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("HX-Redirect"); got != "/" {
		t.Errorf("HX-Redirect = %q, want /", got)
	}
	var cleared *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == security.VIPCookieName {
			cleared = c
		}
	}
	if cleared == nil || cleared.MaxAge > 0 {
		t.Errorf("logout did not expire the cookie: %+v", cleared)
	}
}

func TestVIPLogoutWithoutHTMXRedirectsToHome(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/vip/logout", nil)
	req.AddCookie(&http.Cookie{Name: security.VIPCookieName, Value: security.CreateVIPCookieValue(h.cfg.VIPCookieSecret, nowRef())})
	rec := httptest.NewRecorder()
	h.VIPLogout(rec, req)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("logout = %d Location=%q, want 303 /", rec.Code, rec.Header().Get("Location"))
	}
}

func TestVIPLogoutWithoutCookieIs404(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/vip/logout", nil)
	rec := httptest.NewRecorder()
	h.VIPLogout(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("unauthenticated logout = %d, want 404", rec.Code)
	}
}

// --- Page authorization boundary ---

func TestVIPPageRendersPortalWithValidCookie(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	h.cfg.VIPContactEmail = "karo@example.com"
	h.cfg.VIPContactPhone = "+358 400 234 711"
	cookie := security.CreateVIPCookieValue(h.cfg.VIPCookieSecret, nowRef())
	req := httptest.NewRequest(http.MethodGet, "/en/vip", nil)
	req.AddCookie(&http.Cookie{Name: security.VIPCookieName, Value: cookie})
	rec := httptest.NewRecorder()
	h.VIPPage(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("page = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Log out") {
		t.Error("authed page missing the portal (no logout control)")
	}
	if strings.Contains(rec.Body.String(), "render error:") {
		t.Fatalf("portal template failed during rendering: %s", rec.Body.String())
	}
	for _, marker := range []string{"Ask about the work.", "Why me", "Track record", "Build log", "What I have shipped is the evidence.", "MeetingPackage needs", "A pattern of ownership across domains.", "Levi Golf Green Fee Sales Platform"} {
		if !strings.Contains(rec.Body.String(), marker) {
			t.Errorf("portal missing Phase 3 marker %q", marker)
		}
	}
	for _, marker := range []string{"karo@example.com", "&#43;358 400 234 711", "mailto:karo@example.com", "tel:&#43;358400234711"} {
		if !strings.Contains(rec.Body.String(), marker) {
			t.Errorf("portal missing contact marker %q", marker)
		}
	}
}

func TestVIPCVRequiresCookieAndServesDeployTimeFile(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	dir := t.TempDir()
	path := filepath.Join(dir, "cv.pdf")
	want := []byte("%PDF-1.7 test cv")
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatal(err)
	}
	h.cfg.VIPCVPath = path

	unauthenticated := httptest.NewRecorder()
	h.VIPCV(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/vip/cv", nil))
	if unauthenticated.Code != http.StatusNotFound {
		t.Fatalf("unauthenticated CV = %d, want 404", unauthenticated.Code)
	}

	cookie := security.CreateVIPCookieValue(h.cfg.VIPCookieSecret, nowRef())
	req := httptest.NewRequest(http.MethodGet, "/api/vip/cv", nil)
	req.AddCookie(&http.Cookie{Name: security.VIPCookieName, Value: cookie})
	authenticated := httptest.NewRecorder()
	h.VIPCV(authenticated, req)
	if authenticated.Code != http.StatusOK {
		t.Fatalf("authenticated CV = %d, want 200", authenticated.Code)
	}
	if got := authenticated.Body.Bytes(); string(got) != string(want) {
		t.Fatalf("CV body = %q, want %q", got, want)
	}
	if got := authenticated.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	if got := authenticated.Header().Get("X-Robots-Tag"); got != vipNoIndex {
		t.Errorf("X-Robots-Tag = %q, want %q", got, vipNoIndex)
	}
	if got := authenticated.Header().Get("Content-Disposition"); got != `attachment; filename="Karo-Tammela-CV-2026.pdf"` {
		t.Errorf("Content-Disposition = %q", got)
	}
}

func TestVIPCVHiddenWhenUnconfigured(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	cookie := security.CreateVIPCookieValue(h.cfg.VIPCookieSecret, nowRef())
	req := httptest.NewRequest(http.MethodGet, "/api/vip/cv", nil)
	req.AddCookie(&http.Cookie{Name: security.VIPCookieName, Value: cookie})
	rec := httptest.NewRecorder()
	h.VIPCV(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unconfigured CV = %d, want 404", rec.Code)
	}
}

func TestVIPPageRendersLoginWithTamperedCookie(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	cookie := security.CreateVIPCookieValue("a-different-secret-0123456789abcdef", nowRef())
	req := httptest.NewRequest(http.MethodGet, "/en/vip", nil)
	req.AddCookie(&http.Cookie{Name: security.VIPCookieName, Value: cookie})
	rec := httptest.NewRecorder()
	h.VIPPage(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("page = %d, want 200", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "Log out") {
		t.Error("tampered cookie rendered the portal; want the login card")
	}
	if !strings.Contains(rec.Body.String(), "vip-email") {
		t.Error("tampered cookie did not fall back to the login card")
	}
}

func TestVIPPageExpiredCookieRendersLogin(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	// Minted 48 h ago: beyond the 24 h lifetime.
	cookie := security.CreateVIPCookieValue(h.cfg.VIPCookieSecret, nowRef().Add(-48*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/en/vip", nil)
	req.AddCookie(&http.Cookie{Name: security.VIPCookieName, Value: cookie})
	rec := httptest.NewRecorder()
	h.VIPPage(rec, req)
	if strings.Contains(rec.Body.String(), "Log out") {
		t.Error("expired cookie rendered the portal; want the login card")
	}
}

// nowRef returns a stable reference time for cookie minting in tests.
func nowRef() time.Time { return time.Now() }

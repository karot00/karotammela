package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"goth/internal/content"
	"goth/internal/email"
	"goth/internal/security"
)

// vipNoIndex is emitted on every VIP HTML/API response so the private portal
// can never appear in search results or shared caches (plan §4.5, threat T6).
const vipNoIndex = "noindex, nofollow, noarchive"

// vipCanonicalPath is the single English portal route owned by the Go app.
// /vip and /fi/vip are convenience redirects; the Next.js stack only links.
const vipCanonicalPath = "/en/vip"

// Access-flow limits (plan §6, §15). Rate/token-style tuning values stay code
// constants until operations show a real need.
const (
	vipMaxBodyBytes      = 64 << 10
	vipEmailMaxLen       = 254
	vipCodeMaxLen        = 128
	vipNotifyIPLimit     = 5  // per IP per hour
	vipNotifyEmailLimit  = 3  // per normalized email per hour
	vipNotifyGlobalLimit = 30 // across all valid addresses per hour
	vipAgentMaxLen       = 200

	// vipDefaultLoginFloor is the minimum login response duration that flattens
	// trivial timing signal around hash verification (plan §6.2). A random
	// jitter up to vipLoginJitterMax is added so identical requests do not all
	// return at exactly the same instant.
	vipDefaultLoginFloor = 300 * time.Millisecond
	vipLoginJitterMax    = 150 * time.Millisecond
)

// vipStatus is the minimal cross-stack status contract consumed by the Go
// header/dashboard and by the fail-closed Next.js dashboard link.
type vipStatus struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url,omitempty"`
}

// VIPStatus answers GET /api/vip/status. It is the only VIP route that
// responds when the feature is disabled; it reports {"enabled":false} so the
// Next.js dashboard can hide its link without any shared secret or flag.
func (h *Handlers) VIPStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", vipNoIndex)
	w.Header().Set("Content-Type", "application/json")

	s := vipStatus{Enabled: h.cfg.VIPEnabled}
	if h.cfg.VIPEnabled {
		s.URL = strings.TrimRight(h.cfg.BaseURL, "/") + vipCanonicalPath
	}
	b, err := json.Marshal(s)
	if err != nil {
		b = []byte(`{"enabled":false}`)
	}
	w.Write(b)
}

// VIPEntry serves the convenience paths /vip and /fi/vip: a redirect to the
// canonical English portal while enabled, a plain 404 while disabled.
func (h *Handlers) VIPEntry(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.VIPEnabled {
		h.vipNotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", vipNoIndex)
	http.Redirect(w, r, vipCanonicalPath, http.StatusSeeOther)
}

// VIPPage renders the English VIP portal at /en/vip: the complete portal when
// a valid VIP cookie is present, the login card otherwise (plan §6.1).
func (h *Handlers) VIPPage(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.VIPEnabled {
		h.vipNotFound(w, r)
		return
	}
	h.vipNoStoreHeaders(w)

	data := map[string]any{
		"Title":       "Restricted access — karotammela.fi",
		"Description": "A restricted area available by invitation only.",
		"Authed":      h.vipAuthorized(r),
	}
	if data["Authed"].(bool) {
		vip, err := content.VIP()
		if err != nil {
			http.Error(w, "portal content unavailable", http.StatusInternalServerError)
			return
		}
		data["VIP"] = vip
		data["CVAvailable"] = h.cfg.VIPCVPath != ""
		data["VIPContactEmail"] = h.cfg.VIPContactEmail
		data["VIPContactPhone"] = h.cfg.VIPContactPhone
		data["VIPContactPhoneHref"] = vipPhoneHref(h.cfg.VIPContactPhone)
	}
	h.render(w, "vip", data)
}

// VIPCV serves the deploy-time CV only to an authenticated VIP session.
func (h *Handlers) VIPCV(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.VIPEnabled || !h.vipAuthorized(r) || h.cfg.VIPCVPath == "" {
		h.vipNotFound(w, r)
		return
	}
	f, err := os.Open(h.cfg.VIPCVPath)
	if err != nil {
		h.vipNotFound(w, r)
		return
	}
	defer f.Close()
	h.vipNoStoreHeaders(w)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="Karo-Tammela-CV-2026.pdf"`)
	http.ServeContent(w, r, "Karo-Tammela-CV-2026.pdf", time.Time{}, f)
}

// VIPNotify handles POST /api/vip/notify: email-attempt notification through
// the existing Resend integration (plan §6.1, threat T4). Email submission is
// notification, not identity verification. Every non-protocol outcome returns
// the same password-field partial so delivery state, honeypot hits and
// validation failures are indistinguishable to the visitor.
func (h *Handlers) VIPNotify(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.VIPEnabled {
		h.vipNotFound(w, r)
		return
	}
	h.vipNoStoreHeaders(w)
	if !h.vipBrowserOriginOK(r) {
		log.Printf("vip.notify origin_rejected")
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	ip := security.GetClientIP(headerMap(r))
	if allowed, _, retryMs := security.EnforceRateLimit("vip-notify-ip", ip, vipNotifyIPLimit, time.Hour); !allowed {
		log.Printf("vip.notify rate_limited")
		h.vipRateLimited(w, retryMs, "Too many notifications from this connection. Please try again later.")
		return
	}

	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, vipMaxBodyBytes)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	visitorEmail := strings.ToLower(strings.TrimSpace(r.PostFormValue("email")))
	honeypot := strings.TrimSpace(r.PostFormValue("website"))

	// Honeypot, invalid and over-long addresses all short-circuit into the
	// same generic partial without sending mail (indistinguishable by design).
	if honeypot != "" {
		log.Printf("vip.notify honeypot")
		h.renderVIPPassword(w, "")
		return
	}
	if len(visitorEmail) > vipEmailMaxLen || !validContactEmail(visitorEmail) {
		log.Printf("vip.notify invalid_email")
		h.renderVIPPassword(w, "")
		return
	}
	if allowed, _, retryMs := security.EnforceRateLimit("vip-notify-email", vipHash(visitorEmail), vipNotifyEmailLimit, time.Hour); !allowed {
		log.Printf("vip.notify rate_limited")
		h.vipRateLimited(w, retryMs, "This address has notified recently. Please try again later.")
		return
	}
	if allowed, _, retryMs := security.EnforceRateLimit("vip-notify-global", "all", vipNotifyGlobalLimit, time.Hour); !allowed {
		log.Printf("vip.notify rate_limited")
		h.vipRateLimited(w, retryMs, "Too many notifications have been requested. Please try again later.")
		return
	}

	start := time.Now()
	sent := h.sendVIPNotifyEmail(r, visitorEmail, ip)
	// Redacted structured log: hashes only, never the raw address (threat T4).
	if sent {
		log.Printf("vip.notify ok=true emailHash=%s ipHash=%s duration_ms=%d", vipHash(visitorEmail), vipHash(ip), time.Since(start).Milliseconds())
	} else {
		log.Printf("vip.notify ok=false emailHash=%s ipHash=%s duration_ms=%d", vipHash(visitorEmail), vipHash(ip), time.Since(start).Milliseconds())
	}
	h.renderVIPPassword(w, "")
}

// VIPLogin handles POST /api/vip/login: access-code verification against the
// configured Argon2id/scrypt hash, then a signed VIP session cookie
// (plan §6.2, threats T1/T2). Failures are generic: malformed and incorrect
// codes are indistinguishable in the response.
func (h *Handlers) VIPLogin(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.VIPEnabled {
		h.vipNotFound(w, r)
		return
	}
	h.vipNoStoreHeaders(w)
	if !h.vipBrowserOriginOK(r) {
		log.Printf("vip.login origin_rejected")
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	ip := security.GetClientIP(headerMap(r))
	start := time.Now()
	allowed, retry := h.vipThrottle.Allow(ip, time.Now())
	if !allowed {
		log.Printf("vip.login rate_limited")
		h.vipRateLimited(w, retry.Milliseconds(), "Too many attempts from this connection. Please wait before trying again.")
		return
	}

	code := ""
	if ct := r.Header.Get("Content-Type"); strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		r.Body = http.MaxBytesReader(w, r.Body, vipMaxBodyBytes)
		if err := r.ParseForm(); err == nil {
			code = strings.TrimSpace(r.PostFormValue("code"))
		}
	}

	ok := false
	if code != "" && len(code) <= vipCodeMaxLen && h.cfg.VIPPasswordHash != "" {
		ok = security.VerifyVIPPassword(h.cfg.VIPPasswordHash, code)
	}
	// Uniform timing regardless of outcome, applied before the response is
	// written so verification cost does not leak through response time.
	h.vipSleepFloor(start)

	if !ok {
		log.Printf("vip.login ok=false ipHash=%s duration_ms=%d", vipHash(ip), time.Since(start).Milliseconds())
		h.renderVIPPassword(w, "The password could not be verified. Check it and try again.")
		return
	}

	log.Printf("vip.login ok=true ipHash=%s duration_ms=%d", vipHash(ip), time.Since(start).Milliseconds())
	h.vipSetCookie(w, time.Now())
	vip, err := content.VIP()
	if err != nil {
		http.Error(w, "portal content unavailable", http.StatusInternalServerError)
		return
	}
	h.render(w, "vip-portal", map[string]any{
		"VIP":                 vip,
		"CVAvailable":         h.cfg.VIPCVPath != "",
		"VIPContactEmail":     h.cfg.VIPContactEmail,
		"VIPContactPhone":     h.cfg.VIPContactPhone,
		"VIPContactPhoneHref": vipPhoneHref(h.cfg.VIPContactPhone),
	})
}

func vipPhoneHref(phone string) template.URL {
	var b strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' || r == '+' {
			b.WriteRune(r)
		}
	}
	return template.URL("tel:" + b.String())
}

// VIPLogout handles POST /api/vip/logout: expire the VIP cookie and return to
// the public home page. Every portal API verifies the cookie independently; an
// unauthenticated logout is an indistinguishable 404 (plan §6.2).
func (h *Handlers) VIPLogout(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.VIPEnabled {
		h.vipNotFound(w, r)
		return
	}
	h.vipNoStoreHeaders(w)
	if !h.vipAuthorized(r) {
		h.vipNotFound(w, r)
		return
	}
	log.Printf("vip.logout")
	h.vipExpireCookie(w)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// sendVIPNotifyEmail delivers the private notification to the configured
// recipient. The visitor address appears only in this private email — never in
// logs — and is used as Reply-To so Karo can answer directly.
func (h *Handlers) sendVIPNotifyEmail(r *http.Request, visitorEmail, ip string) bool {
	if h.mailer == nil || h.cfg.ContactFromEmail == "" || h.cfg.ContactToEmail == "" {
		return false
	}
	agent := r.Header.Get("User-Agent")
	if len(agent) > vipAgentMaxLen {
		agent = agent[:vipAgentMaxLen]
	}
	var b strings.Builder
	b.WriteString("The private MeetingPackage application was opened.\n\n")
	b.WriteString("Time:  " + time.Now().UTC().Format("2006-01-02 15:04:05 UTC") + "\n")
	b.WriteString("Email: " + visitorEmail + "\n")
	b.WriteString("IP:    " + ip + "\n")
	if agent != "" {
		b.WriteString("Agent: " + agent + "\n")
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	err := h.mailer.Send(ctx, email.ContactMessage{
		From:    h.cfg.ContactFromEmail,
		To:      h.cfg.ContactToEmail,
		ReplyTo: visitorEmail,
		Subject: "karotammela.fi VIP: private application opened",
		Text:    b.String(),
	})
	return err == nil
}

// vipAuthorized verifies the signed VIP cookie independently of any rendered
// UI state. Hiding controls in HTML is never authorization (plan §6.2).
func (h *Handlers) vipAuthorized(r *http.Request) bool {
	if !h.cfg.VIPEnabled || h.cfg.VIPCookieSecret == "" {
		return false
	}
	c, err := r.Cookie(security.VIPCookieName)
	if err != nil || c.Value == "" {
		return false
	}
	return security.VerifyVIPCookieValue(c.Value, h.cfg.VIPCookieSecret, time.Now()) != nil
}

// vipSetCookie issues the separate karot_vip session cookie: signed, versioned,
// HttpOnly, Secure in production, SameSite=Lax, 24 h server-verified lifetime.
func (h *Handlers) vipSetCookie(w http.ResponseWriter, now time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     security.VIPCookieName,
		Value:    security.CreateVIPCookieValue(h.cfg.VIPCookieSecret, now),
		Path:     "/",
		MaxAge:   int(security.VIPCookieLifetime.Seconds()),
		HttpOnly: true,
		Secure:   h.cfg.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})
}

// vipExpireCookie clears the VIP cookie on logout.
func (h *Handlers) vipExpireCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     security.VIPCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})
}

// vipBrowserOriginOK rejects cross-site browser submissions on state-changing
// VIP routes (plan §6.2). Sec-Fetch-Site is authoritative when present; the
// Origin header is checked against the request host and the configured base
// URL otherwise. Non-browser clients send neither and proceed to the normal
// rate-limited path.
func (h *Handlers) vipBrowserOriginOK(r *http.Request) bool {
	if sf := r.Header.Get("Sec-Fetch-Site"); sf != "" {
		return sf == "same-origin"
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	if u.Host == r.Host {
		return true
	}
	if base, perr := url.Parse(h.cfg.BaseURL); perr == nil && u.Host == base.Host {
		return true
	}
	return false
}

// vipSleepFloor pads the login response to the configured minimum duration
// plus random jitter, reducing trivial timing signal around hash verification.
// A zero floor (used by tests) returns immediately.
func (h *Handlers) vipSleepFloor(start time.Time) {
	floor := h.vipLoginFloor
	if floor <= 0 {
		return
	}
	target := floor + time.Duration(rand.Int64N(int64(vipLoginJitterMax)))
	if elapsed := time.Since(start); elapsed < target {
		time.Sleep(target - elapsed)
	}
}

// vipNoStoreHeaders sets the cache/search headers shared by every VIP HTML and
// API response (plan §4.5).
func (h *Handlers) vipNoStoreHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", vipNoIndex)
}

// vipRateLimited writes a 429 with Retry-After and a calm HTML partial. The
// partial is swappable by HTMX because vip.html extends htmx responseHandling
// to 429.
func (h *Handlers) vipRateLimited(w http.ResponseWriter, retryAfterMs int64, message string) {
	w.Header().Set("Retry-After", strconv.FormatInt((retryAfterMs+999)/1000, 10))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = h.view.Render(w, "vip-wait", map[string]any{"Message": message})
}

// renderVIPPassword renders the access-code step partial with an optional
// generic error message.
func (h *Handlers) renderVIPPassword(w http.ResponseWriter, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.view.Render(w, "vip-password", map[string]any{"Error": errMsg})
}

// vipHash redacts an identifier (email, IP) to a short hash prefix for
// structured logs. Never log the raw value (plan §14, threat T4).
func vipHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:12]
}

// vipNotFound returns a plain 404 indistinguishable from any other missing
// route. It must not redirect to a login or reveal that a VIP feature exists
// (plan §5.1).
func (h *Handlers) vipNotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

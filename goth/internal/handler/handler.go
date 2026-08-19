package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"goth/internal/ai"
	"goth/internal/aipulse"
	"goth/internal/config"
	"goth/internal/content"
	"goth/internal/email"
	"goth/internal/i18n"
	"goth/internal/security"
	"goth/internal/view"
)

// MailSender delivers a contact email. Implemented by email.ResendSender;
// nil means contact delivery is not configured (503).
type MailSender interface {
	Send(ctx context.Context, msg email.ContactMessage) error
}

// Handlers bundles shared dependencies.
type Handlers struct {
	cfg       *config.Config
	view      *view.Renderer
	conn      *sql.DB
	gemini    *ai.GeminiStreamer
	vipGemini *ai.GeminiStreamer
	mailer    MailSender
	refresher *aipulse.Refresher

	// VIP access flow (plan §6): per-IP login throttle with escalating
	// cooldowns, and the minimum login response duration that flattens trivial
	// timing signal. Tests construct Handlers literals and may zero the floor.
	vipThrottle   *security.VIPLoginThrottle
	vipLoginFloor time.Duration
	vipChat       *security.VIPChatLimiter
	vipContent    content.VIPContent
}

// New builds the handler set.
func New(cfg *config.Config, vr *view.Renderer, conn *sql.DB, g *ai.GeminiStreamer, mailer MailSender) *Handlers {
	return &Handlers{
		cfg:           cfg,
		view:          vr,
		conn:          conn,
		gemini:        g,
		mailer:        mailer,
		vipThrottle:   security.NewVIPLoginThrottle(),
		vipLoginFloor: vipDefaultLoginFloor,
		vipChat:       security.NewVIPChatLimiter(),
	}
}

// SetRefresher installs the AI Pulse refresh orchestrator used by
// POST /api/ai-pulse/refresh. nil (the default) makes the endpoint 503.
func (h *Handlers) SetRefresher(r *aipulse.Refresher) {
	h.refresher = r
}

// SetVIPGemini installs the VIP-only model client without changing Sentinel's
// configured model or transport behavior.
func (h *Handlers) SetVIPGemini(g *ai.GeminiStreamer)  { h.vipGemini = g }
func (h *Handlers) SetVIPContent(v content.VIPContent) { h.vipContent = v }

func themeFromCookie(r *http.Request) string {
	c, err := r.Cookie("theme")
	if err != nil || c.Value == "" {
		return "dark"
	}
	if c.Value == "light" {
		return "light"
	}
	return "dark"
}

func switchURLs(r *http.Request, locale string) (string, string) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "/en", "/fi"
	}
	other := make([]string, len(parts))
	copy(other, parts)
	other[0] = "en"
	if len(other) >= 3 && other[1] == "blog" {
		other[2] = translatedBlogSlug(locale, "en", other[2])
	}
	en := "/" + strings.Join(other, "/")
	other[0] = "fi"
	if len(other) >= 3 && other[1] == "blog" {
		other[2] = translatedBlogSlug(locale, "fi", parts[2])
	}
	fi := "/" + strings.Join(other, "/")
	return en, fi
}

func translatedBlogSlug(currentLocale, targetLocale, slug string) string {
	if currentLocale == targetLocale {
		return slug
	}

	pairs := map[string]string{
		"how-ai-unlocked-my-coding":                     "tekoaly-avasi-koodaukseni-lukot",
		"vibe-coding-vs-production-ready":               "vibe-coding-vs-tuotantovalmis",
		"mcp-bridging-the-knowledge-gap":                "mcp-bridgin-the-knowledge-gap",
		"whatsapp-chaos-to-automated-process-levi-golf": "whatsapp-kaoottisuudesta-automatisoituun-prosessiin-levi-golf",
		"agentic-ai-era-blogs-in-mdx":                   "agenttisen-tekoalyn-aikakausi-blogit-markdownilla",
		"modern-ai-era-blog-in-md-or-mdx":               "moderni-seo-blogi-markdown-vai-mdx",
		"go-htmx-vs-nextjs-hetzner":                     "go-htmx-nextjs-vertailu-hetzner",
		"how-i-moved-wordpress-site-to-next-js":         "miten-siirsin-wordpress-sivuston-next-js-teknologiaan",
		"home-ai-lab-qwen-3-8-27b":                      "kotilabra-ja-qwen-3-8-27b",
	}
	if currentLocale == "fi" {
		for english, finnish := range pairs {
			if slug == finnish {
				return english
			}
		}
	}
	if targetLocale == "fi" {
		if translated, ok := pairs[slug]; ok {
			return translated
		}
	}
	return slug
}

func (h *Handlers) common(r *http.Request, locale string) map[string]any {
	tr := func(key string) string { return i18n.T(locale, key) }
	en, fi := switchURLs(r, locale)
	consent := security.DefaultConsentState()
	if c, err := r.Cookie(security.ConsentCookieName); err == nil {
		consent = security.ParseConsentState(c.Value, time.Now())
	}
	canonical, ogURL, links := localizedAlternates(h.cfg, r.URL.Path)
	base := strings.TrimRight(h.cfg.BaseURL, "/")
	// VIP navigation is controlled by the single Go runtime flag (plan §5):
	// the label is the locale-invariant "VIP" and the destination is always
	// the English Go route. Empty string hides the link in both the global
	// header and the dashboard navigation.
	vipURL := ""
	if h.cfg.VIPEnabled {
		vipURL = vipCanonicalPath
	}
	return map[string]any{
		"Locale":      locale,
		"Theme":       themeFromCookie(r),
		"Tr":          tr,
		"Title":       tr("metadata.title"),
		"Description": tr("metadata.description"),
		// Canonical mirrors getLocalizedAlternates: the default-locale URL of
		// this path (avoids duplicate-content during the stack experiment).
		"Canonical": canonical,
		// og:url is the absolute URL of the current page (reference openGraph.url).
		"OGURL":      ogURL,
		"Hreflangs":  links,
		"OGType":     "website",
		"SiteName":   siteName,
		"OGImage":    base + siteOGImagePath,
		"OGImageAlt": siteOGImageAlt,
		"SwitchToEn": en,
		"SwitchToFi": fi,
		// Comparison origins for the Tech Switcher (cross-origin redirect) and
		// the performance widget (direct client-side pings). Go is the apex,
		// Next.js is hosted on its own Vercel subdomain (2026-07-25 decision).
		"GoURL":   strings.TrimRight(h.cfg.BaseURL, "/"),
		"NextURL": strings.TrimRight(h.cfg.NextURL, "/"),
		// Conditional VIP link target ("" hides the link everywhere).
		"VIPURL": vipURL,
		"Year":   time.Now().Year(),
		// Global WebSite + publisher Person JSON-LD (src/app/layout.tsx).
		"JSONLDSite": websiteJSONLD(h.cfg),
		// SSR gate for the consent banner: render only until the first
		// stored decision (mirrors isBannerRequired in the Next.js layout).
		"ConsentBannerRequired": security.IsConsentBannerRequired(consent),
	}
}

func (h *Handlers) render(w http.ResponseWriter, name string, data map[string]any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.view.Render(w, name, data); err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handlers) resolveLocale(r *http.Request) string {
	loc := chi.URLParam(r, "locale")
	if !i18n.Exists(loc) {
		return i18n.DefaultLocale
	}
	return loc
}

func (h *Handlers) hasUnlockedBefore(r *http.Request) bool {
	if h.cfg.UnlockCookieSecret == "" {
		return false
	}
	c, err := r.Cookie("karot_unlock")
	if err != nil {
		return false
	}
	return security.VerifyUnlockCookieValue(c.Value, h.cfg.UnlockCookieSecret) != nil
}

func (h *Handlers) sentinelConfig(r *http.Request, locale string) string {
	tr := func(key string) string { return i18n.T(locale, key) }
	cfg := map[string]any{
		"locale":              locale,
		"accessCode":          ai.GetAccessCode(),
		"hasUnlockedBefore":   h.hasUnlockedBefore(r),
		"dashboardPath":       "/" + locale + "/dashboard",
		"homePath":            "/" + locale,
		"title":               tr("home.sentinel.title"),
		"subtitle":            tr("home.sentinel.subtitle"),
		"meterLabel":          tr("home.sentinel.meterLabel"),
		"inputPlaceholder":    tr("home.sentinel.inputPlaceholder"),
		"sendLabel":           tr("home.sentinel.sendLabel"),
		"resetLabel":          tr("home.sentinel.resetLabel"),
		"initialMessage":      tr("home.sentinel.initialMessage"),
		"unlockedLabel":       tr("home.sentinel.unlockedLabel"),
		"unlockedCta":         tr("home.sentinel.unlockedCta"),
		"pendingLabel":        tr("home.sentinel.pendingLabel"),
		"errorLabel":          tr("home.sentinel.errorLabel"),
		"returnOverlayTitle":  tr("home.sentinel.returnOverlayTitle"),
		"returnOverlayBody":   tr("home.sentinel.returnOverlayBody"),
		"playAgainLabel":      tr("home.sentinel.playAgainLabel"),
		"goDashboardLabel":    tr("home.sentinel.goDashboardLabel"),
		"bypassInstructions":  tr("home.sentinel.bypassInstructions"),
		"revealPasscodeLabel": tr("home.sentinel.revealPasscodeLabel"),
		"passcodeLabel":       tr("home.sentinel.passcodeLabel"),
		"copyPasscodeLabel":   tr("home.sentinel.copyPasscodeLabel"),
		"copiedPasscodeLabel": tr("home.sentinel.copiedPasscodeLabel"),
		"directUnlockMessage": tr("home.sentinel.directUnlockMessage"),
	}
	b, _ := json.Marshal(cfg)
	// Guard against a literal "</" breaking out of the <script> JSON block.
	return strings.ReplaceAll(string(b), "</", "<\\/")
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	locale := h.resolveLocale(r)
	data := h.common(r, locale)
	tr := func(key string) string { return i18n.T(locale, key) }
	data["Badge"] = tr("home.phaseLabel")
	data["Intro"] = tr("home.intro")
	data["Body1"] = tr("home.body1")
	data["Body2"] = tr("home.body2")
	data["Body3"] = tr("home.body3")
	data["Body4"] = tr("home.body4")
	data["SentinelConfig"] = h.sentinelConfig(r, locale)
	h.render(w, "home", data)
}

func (h *Handlers) Privacy(w http.ResponseWriter, r *http.Request) {
	locale := h.resolveLocale(r)
	data := h.common(r, locale)
	h.render(w, "privacy", data)
}

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.setPingCORS(w, r)
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	target := r.URL.Query().Get("target")
	if target == "next" {
		ms := h.pingNext(r.Context())
		if ms == nil {
			w.Write([]byte(`{"stack":"next","ms":null}`))
			return
		}
		w.Write([]byte(`{"stack":"next","ms":` + strconv.Itoa(*ms) + `}`))
		return
	}
	ms := int(time.Since(start).Milliseconds())
	w.Write([]byte(`{"stack":"go","ms":` + strconv.Itoa(ms) + `}`))
}

// setPingCORS allows the cross-origin performance widget to fetch /api/ping.
// The Next.js comparison build (served from NextURL) pings Go client-side, so
// its origin must be allow-listed. The Origin is reflected only when it matches
// a known comparison origin; unknown origins receive no ACAO header.
func (h *Handlers) setPingCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Vary", "Origin")
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}
	allowed := map[string]bool{
		strings.TrimRight(h.cfg.NextURL, "/"): true,
		strings.TrimRight(h.cfg.BaseURL, "/"): true,
	}
	if !allowed[origin] {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func (h *Handlers) pingNext(ctx context.Context) *int {
	raw := h.cfg.NextPingURL
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil
	}
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	ms := int(time.Since(start).Milliseconds())
	return &ms
}

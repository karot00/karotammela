package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"goth/internal/config"
	"goth/internal/content"
	"goth/internal/view"
)

// newTestHandlers builds a Handlers with a real renderer for SEO output tests.
// DB/Gemini/mailer are nil because the public pages under test don't need them.
func newTestHandlers(t *testing.T) *Handlers {
	t.Helper()
	vr, err := view.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	return New(&config.Config{BaseURL: "https://karotammela.fi"}, vr, nil, nil, nil)
}

// serve calls an http.HandlerFunc directly.
func serve(t *testing.T, h http.HandlerFunc, path string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", path, rec.Code)
	}
	return rec.Body.String()
}

// serveWithParams calls an http.HandlerFunc after injecting chi route params so
// handlers using chi.URLParam can resolve {locale}/{slug}.
func serveWithParams(t *testing.T, h http.HandlerFunc, path string, params map[string]string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", path, rec.Code)
	}
	return rec.Body.String()
}

func TestHomeSEOAlternatesAndJSONLD(t *testing.T) {
	h := newTestHandlers(t)
	body := serve(t, h.Home, "/en")

	// Canonical must point at the default-locale (fi) version of this path.
	if !strings.Contains(body, `<link rel="canonical" href="https://karotammela.fi/fi" />`) {
		t.Errorf("missing/default-locale canonical in:\n%s", body)
	}
	// Both hreflang alternates present.
	if !strings.Contains(body, `hreflang="fi-FI" href="https://karotammela.fi/fi"`) {
		t.Errorf("missing fi-FI alternate")
	}
	if !strings.Contains(body, `hreflang="en-US" href="https://karotammela.fi/en"`) {
		t.Errorf("missing en-US alternate")
	}
	// og:url is the current page; og:type is website; site_name present.
	if !strings.Contains(body, `<meta property="og:url" content="https://karotammela.fi/en" />`) {
		t.Errorf("og:url should be current page")
	}
	if !strings.Contains(body, `<meta property="og:type" content="website" />`) {
		t.Errorf("og:type should be website on home")
	}
	if !strings.Contains(body, `<meta property="og:site_name" content="karotammela.fi" />`) {
		t.Errorf("missing og:site_name")
	}
	// Global WebSite JSON-LD present and names the owner as publisher Person.
	if !strings.Contains(body, `"@type":"WebSite"`) {
		t.Errorf("missing WebSite JSON-LD")
	}
	if !strings.Contains(body, `"@type":"Person"`) || !strings.Contains(body, `"name":"Karo Tammela"`) {
		t.Errorf("WebSite JSON-LD missing publisher Person")
	}
	// Parity OG image (reference uses karo-tammela-agentic-ai.png, not hero-poster.svg).
	if !strings.Contains(body, `<meta property="og:image" content="https://karotammela.fi/media/karo-tammela-agentic-ai.png" />`) {
		t.Errorf("OG image should be the reference PNG")
	}
}

func TestBlogPostSEOArticleMetadata(t *testing.T) {
	h := newTestHandlers(t)
	posts, err := content.GetAllBlogPosts("en")
	if err != nil || len(posts) == 0 {
		t.Skip("no en posts available")
	}
	slug := posts[0].Slug
	body := serveWithParams(t, h.BlogPost, "/en/blog/"+slug, map[string]string{
		"locale": "en",
		"slug":   slug,
	})

	if !strings.Contains(body, `<meta property="og:type" content="article" />`) {
		t.Errorf("blog post og:type should be article")
	}
	if !strings.Contains(body, `<meta property="article:published_time"`) {
		t.Errorf("missing article:published_time")
	}
	if !strings.Contains(body, `<meta property="article:modified_time"`) {
		t.Errorf("missing article:modified_time")
	}
	// Title must carry the site-name suffix, matching the reference.
	if !strings.Contains(body, `| karotammela.fi`) {
		t.Errorf("blog post title should include site name suffix")
	}
	// BlogPosting JSON-LD present.
	if !strings.Contains(body, `"@type":"BlogPosting"`) {
		t.Errorf("missing BlogPosting JSON-LD")
	}
	if !strings.Contains(body, `"@type":"Organization"`) {
		t.Errorf("BlogPosting JSON-LD missing publisher Organization")
	}
	// Canonical points at the fi (default) version even when viewed in en.
	if !strings.Contains(body, `<link rel="canonical" href="https://karotammela.fi/fi/blog/`+slug+`" />`) {
		t.Errorf("blog canonical should be default-locale path")
	}
}

func TestSitemapIncludesRootAndPosts(t *testing.T) {
	h := newTestHandlers(t)
	body := serve(t, h.Sitemap, "/sitemap.xml")

	if !strings.Contains(body, "<loc>https://karotammela.fi/</loc>") {
		t.Errorf("sitemap missing root /")
	}
	if !strings.Contains(body, `<priority>1.0</priority>`) {
		t.Errorf("root should have priority 1.0")
	}
	if !strings.Contains(body, `<changefreq>daily</changefreq>`) {
		t.Errorf("blog index should be daily")
	}
	// At least one blog post URL (fail-closed: only visible posts).
	posts, err := content.GetAllBlogPosts("en")
	if err == nil && len(posts) > 0 {
		if !strings.Contains(body, "/en/blog/"+posts[0].Slug) {
			t.Errorf("sitemap missing en blog post %s", posts[0].Slug)
		}
	}
}

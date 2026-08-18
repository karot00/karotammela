package handler

import (
	"fmt"
	"net/http"
	"strings"

	"goth/internal/content"
	"goth/internal/i18n"
)

func (h *Handlers) Robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	base := strings.TrimRight(h.cfg.BaseURL, "/")
	// VIP paths are disallowed regardless of the feature flag (plan §4.5):
	// the rules are harmless while disabled and must not be forgotten when
	// the portal is enabled. No VIP URL is ever added to the sitemap.
	fmt.Fprintf(w, "User-agent: *\nAllow: /\nDisallow: /vip\nDisallow: /en/vip\nDisallow: /fi/vip\nDisallow: /api/vip/\n\nSitemap: %s/sitemap.xml\n", base)
}

func (h *Handlers) Sitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	base := strings.TrimRight(h.cfg.BaseURL, "/")
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	// Root redirect target (mirrors src/app/sitemap.ts staticRoutes[0]).
	b.WriteString(sitemapURL(base+"/", "", "weekly", "1.0"))

	for _, loc := range i18n.Locales {
		b.WriteString(sitemapURL(base+"/"+loc, "", "weekly", "0.9"))
		b.WriteString(sitemapURL(base+"/"+loc+"/blog", "", "daily", "0.8"))
		b.WriteString(sitemapURL(base+"/"+loc+"/privacy", "", "monthly", "0.5"))
	}

	// Blog posts (fail-closed: only visible posts, matching the reference which
	// reads getAllBlogPosts for each locale).
	for _, loc := range i18n.Locales {
		posts, err := content.GetAllBlogPosts(loc)
		if err != nil {
			continue
		}
		for _, p := range posts {
			lastmod := articleISO(p.PublishedAt)
			b.WriteString(sitemapURL(base+"/"+loc+"/blog/"+p.Slug, lastmod, "monthly", "0.7"))
		}
	}

	b.WriteString("</urlset>\n")
	w.Write([]byte(b.String()))
}

// sitemapURL emits a <url> entry with optional lastmod (RFC3339), changefreq,
// and priority, mirroring the elements Next.js emits from MetadataRoute.Sitemap.
func sitemapURL(loc, lastmod, changefreq, priority string) string {
	var b strings.Builder
	b.WriteString("  <url><loc>" + loc + "</loc>")
	if lastmod != "" {
		b.WriteString("<lastmod>" + lastmod + "</lastmod>")
	}
	if changefreq != "" {
		b.WriteString("<changefreq>" + changefreq + "</changefreq>")
	}
	if priority != "" {
		b.WriteString("<priority>" + priority + "</priority>")
	}
	b.WriteString("</url>\n")
	return b.String()
}

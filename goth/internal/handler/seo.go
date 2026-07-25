package handler

import (
	"encoding/json"
	"html/template"
	"strings"
	"time"

	"goth/internal/config"
	"goth/internal/content"
	"goth/internal/i18n"
)

// jsonLD is a JSON-LD document safe to embed inside a <script> tag. Using
// template.JS (not template.HTML) prevents html/template from JS-escaping the
// quotes when the value is rendered in script context.
type jsonLD = template.JS

// SEO parity constants mirror src/lib/seo.ts.
const (
	siteName        = "karotammela.fi"
	siteOwnerName   = "Karo Tammela"
	siteOGImagePath = "/media/karo-tammela-agentic-ai.png"
	siteOGImageAlt  = "Karo Tammela - Agentic AI blog and experiment lab"
)

// hreflangLink is one <link rel="alternate" hreflang="..."> entry.
type hreflangLink struct {
	Hreflang string
	Href     string
}

// languageTag maps an app locale to its BCP-47 language tag, matching
// LOCALE_TO_LANGUAGE_TAG in src/lib/seo.ts.
func languageTag(locale string) string {
	if locale == "fi" {
		return "fi-FI"
	}
	return "en-US"
}

// pathWithoutLocale strips a leading /{locale} segment and returns the rest
// with a leading slash ("" for the locale root). e.g. "/fi/blog/x" -> "/blog/x",
// "/fi" -> "", "/" -> "".
func pathWithoutLocale(path string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 1 {
		return ""
	}
	return "/" + parts[1]
}

// localizedAlternates builds the hreflang alternate links and the canonical
// URL (always the default-locale path), mirroring getLocalizedAlternates in
// src/lib/seo.ts. ogURL is the absolute URL of the current request path.
func localizedAlternates(cfg *config.Config, path string) (canonical, ogURL string, links []hreflangLink) {
	base := strings.TrimRight(cfg.BaseURL, "/")
	rest := pathWithoutLocale(path)
	ogURL = base + path
	canonical = base + "/" + i18n.DefaultLocale + rest
	for _, loc := range i18n.Locales {
		links = append(links, hreflangLink{
			Hreflang: languageTag(loc),
			Href:     base + "/" + loc + rest,
		})
	}
	return canonical, ogURL, links
}

// websiteJSONLD emits the global WebSite + publisher Person JSON-LD, identical
// to the script in src/app/layout.tsx.
func websiteJSONLD(cfg *config.Config) template.JS {
	base := strings.TrimRight(cfg.BaseURL, "/")
	ld := map[string]any{
		"@context":   "https://schema.org",
		"@type":      "WebSite",
		"name":       siteName,
		"url":        base + "/",
		"inLanguage": []string{"fi-FI", "en-US"},
		"publisher": map[string]any{
			"@type": "Person",
			"name":  siteOwnerName,
			"url":   base + "/",
		},
	}
	b, _ := json.Marshal(ld)
	return template.JS(b)
}

// articleISO parses the post's publishedAt ("2006-01-02") into the same
// millisecond ISO-8601 form Next.js produces via new Date(...).toISOString().
func articleISO(publishedAt string) string {
	t, err := time.Parse("2006-01-02", publishedAt)
	if err != nil {
		return publishedAt
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

// articleJSONLD emits the BlogPosting JSON-LD for a single post, mirroring the
// script in src/app/[locale]/blog/[slug]/page.tsx.
func articleJSONLD(cfg *config.Config, post *content.BlogPost) template.JS {
	base := strings.TrimRight(cfg.BaseURL, "/")
	locale := post.Locale
	localeCode := languageTag(locale)
	iso := articleISO(post.PublishedAt)
	ld := map[string]any{
		"@context":      "https://schema.org",
		"@type":         "BlogPosting",
		"headline":      post.Title,
		"description":   post.Description,
		"datePublished": iso,
		"dateModified":  iso,
		"inLanguage":    localeCode,
		"author": map[string]any{
			"@type": "Person",
			"name":  siteOwnerName,
			"url":   base + "/" + locale,
		},
		"publisher": map[string]any{
			"@type": "Organization",
			"name":  siteName,
			"url":   base + "/",
		},
		"mainEntityOfPage": base + "/" + locale + "/blog/" + post.Slug,
		"keywords":         strings.Join(post.Tags, ", "),
	}
	b, _ := json.Marshal(ld)
	return template.JS(b)
}

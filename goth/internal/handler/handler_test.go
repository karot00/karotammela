package handler

import (
	"net/http/httptest"
	"testing"
)

func TestSwitchURLsTranslatesBlogSlugs(t *testing.T) {
	r := httptest.NewRequest("GET", "/en/blog/go-htmx-vs-nextjs-hetzner", nil)
	en, fi := switchURLs(r, "en")

	if en != "/en/blog/go-htmx-vs-nextjs-hetzner" {
		t.Errorf("English switch URL = %q", en)
	}
	if fi != "/fi/blog/go-htmx-nextjs-vertailu-hetzner" {
		t.Errorf("Finnish switch URL = %q", fi)
	}

	r = httptest.NewRequest("GET", "/fi/blog/go-htmx-nextjs-vertailu-hetzner", nil)
	en, fi = switchURLs(r, "fi")
	if en != "/en/blog/go-htmx-vs-nextjs-hetzner" {
		t.Errorf("English reverse switch URL = %q", en)
	}
	if fi != "/fi/blog/go-htmx-nextjs-vertailu-hetzner" {
		t.Errorf("Finnish reverse switch URL = %q", fi)
	}
}

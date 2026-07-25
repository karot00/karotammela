package view

import (
	"bytes"
	"strings"
	"testing"
)

// TestRenderAllTemplates executes every page template with a populated data map.
// This is the regression guard for the html/template `{{.Tr "x"}}` vs
// `{{call .Tr "x"}}` mistake: a function-valued field invoked as a method aborts
// the whole render, so any such error fails this test loudly.
func TestRenderAllTemplates(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	tr := func(key string) string { return key }

	base := map[string]any{
		"Locale":      "fi",
		"Theme":       "dark",
		"Title":       "Title",
		"Description": "Description",
		"Canonical":   "https://example.com/fi",
		"OGImage":     "https://example.com/i.svg",
		"SwitchToEn":  "/en",
		"SwitchToFi":  "/fi",
		"Year":        2026,
		"Tr":          tr,
	}

	sentinelCfg := `{"locale":"fi","accessCode":"PROTOCOL_K_2026"}`

	cases := []struct {
		name string
		data map[string]any
	}{
		{
			name: "home",
			data: merge(base, map[string]any{
				"Badge":          "badge",
				"Intro":          "intro",
				"Body1":          "b1",
				"Body2":          "b2",
				"Body3":          "b3",
				"SentinelConfig": sentinelCfg,
			}),
		},
		{
			name: "privacy",
			data: merge(base, map[string]any{}),
		},
		{
			name: "dashboard",
			data: merge(base, map[string]any{
				"BackLabel": "Back",
				"Stats": map[string]any{
					"TotalAttempts":       1,
					"UnlockedCount":       1,
					"DirectUnlockCount":   1,
					"AvgMessagesToUnlock": 3.0,
					"LatestUnlock":        nil,
				},
			}),
		},
		{
			name: "blog_list",
			data: merge(base, map[string]any{
				"Title":    "Blog",
				"Subtitle": "sub",
				"HasPosts": true,
				"Posts": []map[string]any{
					{"Slug": "s", "Title": "t", "Description": "d", "PublishedAt": "2026-01-01", "Draft": false, "Tags": []string{}},
				},
				"HasPrev":   false,
				"HasNext":   false,
				"PrevPage":  1,
				"NextPage":  1,
				"Page":      1,
				"Pages":     1,
				"PrevLabel": "Prev",
				"NextLabel": "Next",
				"PageLabel": "Page",
			}),
		},
		{
			name: "blog_post",
			data: merge(base, map[string]any{
				"BackLabel":   "Back",
				"PublishedAt": "2026-01-01",
				"Draft":       false,
				"Title":       "Post",
				"Description": "desc",
				"Tags":        []string{"a", "b"},
				"HTML":        "<p>hello</p>",
			}),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := r.Render(discard{}, c.name, c.data); err != nil {
				t.Fatalf("render %q failed: %v", c.name, err)
			}
		})
	}
}

// TestRenderDashboardViews renders the dashboard for every ?view= value with a
// fully populated data map, guarding the multi-view template against missing
// fields, the date/active/changeClass funcs, and the {{.Tr}} method mistake.
func TestRenderDashboardViews(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	tr := func(key string) string { return key }

	project := func(title, desc, href, gh string, internal bool) map[string]any {
		return map[string]any{"Title": title, "Description": desc, "Href": href, "GithubHref": gh, "IsInternal": internal, "GithubLabel": "GH"}
	}
	projectBadge := func(title, desc, href, badge string) map[string]any {
		m := project(title, desc, href, "", false)
		m["Badge"] = badge
		return m
	}
	techItem := func(id, name, desc string) map[string]any {
		return map[string]any{"ID": id, "Name": name, "Description": desc}
	}
	techCat := func(label string, items ...map[string]any) map[string]any {
		return map[string]any{"Label": label, "Items": items}
	}
	blogItem := func(slug, title, desc, pub string, draft bool) map[string]any {
		return map[string]any{"Slug": slug, "Title": title, "Description": desc, "PublishedAt": pub, "Draft": draft, "Tags": []string{}, "HTML": "<p>body</p>"}
	}

	views := []string{"overview", "projects", "tech", "blog", "changelog", "ai-pulse", "settings"}

	for _, view := range views {
		t.Run(view, func(t *testing.T) {
			data := map[string]any{
				"Locale":                     "fi",
				"Theme":                      "dark",
				"Title":                      "Title",
				"Description":                "Description",
				"Canonical":                  "https://example.com/fi",
				"OGImage":                    "https://example.com/i.svg",
				"SwitchToEn":                 "/en",
				"SwitchToFi":                 "/fi",
				"Year":                       2026,
				"Tr":                         tr,
				"ActiveView":                 view,
				"DashBase":                   "/fi/dashboard",
				"ProfileImage":               "/media/Karo-Tammela.jpg",
				"ProfileImageAlt":            "Karo Tammela",
				"BackLabel":                  "Back",
				"MenuLabel":                  "Menu",
				"NavigationLabel":            "Nav",
				"NavOverview":                "Overview",
				"NavProjects":                "Projects",
				"NavTech":                    "Tech",
				"NavBlog":                    "Blog",
				"NavChangelog":               "Changelog",
				"NavAiPulse":                 "AI Pulse",
				"NavSettings":                "Settings",
				"Badge":                      "badge",
				"Stats":                      map[string]any{"TotalAttempts": 1, "UnlockedCount": 1, "DirectUnlockCount": 1, "AvgMessagesToUnlock": 3.0, "LatestUnlock": nil},
				"StatsOffline":               false,
				"AnalyticsTitle":             "analytics",
				"TotalAttemptsLabel":         "attempts",
				"UnlockedCountLabel":         "unlocks",
				"DirectUnlockCountLabel":     "direct",
				"AvgMessagesToUnlockLabel":   "avg",
				"LatestUnlockLabel":          "latest",
				"SourceOffline":              "offline",
				"LatestUnlockNever":          "never",
				"LatestUnlock":               "",
				"ContactTitle":               "Contact",
				"ContactDescription":         "desc",
				"ContactNameLabel":           "Name",
				"ContactEmailLabel":          "Email",
				"ContactCompanyLabel":        "Company",
				"ContactMessageLabel":        "Message",
				"ContactSubmitLabel":         "Send",
				"ContactPendingLabel":        "Sending",
				"ContactSuccessLabel":        "OK",
				"ContactErrorLabel":          "Err",
				"ContactAvailabilityEyebrow": "avail",
				"ContactAvailabilityStatus":  "open",
				"ContactConnectLabel":        "connect",
				"ContactGithubLabel":         "GH",
				"ContactLinkedinLabel":       "LI",
				"GithubURL":                  "https://github.com/x",
				"LinkedinURL":                "https://linkedin.com/x",
				"AboutTitle":                 "About",
				"AboutBody":                  "body",
				"ToolkitTitle":               "tk",
				"CodingModelsLabel":          "cm",
				"ChatbotModelLabel":          "cb",
				"ToolkitTitleH2":             "tk2",
				"CodingModelsLabelH2":        "cm2",
				"ChatbotModelLabelH2":        "cb2",
				"ProjectsTitle":              "Projects",
				"ProjectsProductionSubtitle": "prod",
				"ProjectsShowcaseSubtitle":   "show",
				"ProjectsProduction": []map[string]any{
					project("P1", "d", "https://e.fi", "", false),
					project("P4", "d", "", "https://github.com/x", true),
				},
				"ProjectsShowcase": []map[string]any{
					projectBadge("P5", "d", "https://karotammela.fi/fi/postikortti", "Built with Next.js"),
				},
				"TechTitle": "Tech",
				"TechCategories": []map[string]any{
					techCat("Frontend", techItem("nextjs", "Next.js", "n")),
					techCat("Tools", techItem("resend", "Resend", "r")),
				},
				"BlogItems": []map[string]any{
					blogItem("slug-one", "T1", "d1", "2026-01-02", false),
					blogItem("slug-two", "T2", "d2", "2026-03-04", true),
				},
				"BlogPage":             1,
				"BlogPages":            1,
				"BlogHasPrev":          false,
				"BlogHasNext":          false,
				"BlogPrevPage":         1,
				"BlogNextPage":         1,
				"BlogSelected":         blogItem("slug-one", "T1", "d1", "2026-01-02", false),
				"BlogMissingPost":      false,
				"BlogTotal":            2,
				"BlogPrevLabel":        "Prev",
				"BlogNextLabel":        "Next",
				"BlogPageLabel":        "Page",
				"BlogNoPostsLabel":     "none",
				"BlogMissingPostLabel": "missing",
				"BlogPlaceholder":      "placeholder",
				"BlogShareLabel":       "share",
				"BlogShareCopiedLabel": "copied",
				"BlogShareErrorLabel":  "err",
				"BlogDraftBadgeLabel":  "draft",
				"Changelog": []map[string]any{
					{"Version": "v1.3", "Date": "2026-06-02", "Title": "t", "Changes": []map[string]any{{"Type": "added", "Text": "x"}}},
				},
				"ChangelogTitle":              "CL",
				"ChangelogLead":               "lead",
				"ChangelogEmptyLabel":         "empty",
				"ChangelogTypeAdded":          "added",
				"ChangelogTypeChanged":        "changed",
				"ChangelogTypeFixed":          "fixed",
				"ChangelogTypeRemoved":        "removed",
				"SettingsLanguageTitle":       "lang",
				"SettingsLanguageDescription": "d",
				"SettingsCookieTitle":         "cookie",
				"SettingsCookieDescription":   "d",
				"SettingsThemeTitle":          "theme",
				"SettingsThemeDescription":    "d",
				"LightModeLabel":              "light",
				"DarkModeLabel":               "dark",
				"PrivacyPolicyLink":           "privacy",
				"PrivacyReturn":               "/fi/privacy",
				"ConsentSettingsTrigger":      "consent",
				"ConsentSettingsTriggerAria":  "consent aria",
				"ConsentBannerRequired":       true,
				"AiPulseTitle":                "AI",
				"AiPulseDescription":          "d",
				"AiPulseTrendsTitle":          "trends",
				"AiPulseReposTitle":           "repos",
				"AiPulseStocksTitle":          "stocks",
				"AiPulseTickerLabel":          "ticker",
				"AiPulseNoTrendsLabel":        "no trends",
				"AiPulseNoReposLabel":         "no repos",
				"AiPulseLoadingLabel":         "loading",
				"AiPulseLastUpdatedLabel":     "updated",
				"AiPulseSourceLabel":          "source",
				"AiPulseStarsTodayLabel":      "stars",
				"AiPulseAvailable":            true,
				"AiPulseStockChartAria":       "chart",
				"AiPulseStockTableCaption":    "table",
				"AiPulseStockDateLabel":       "Date",
				"AiPulseStockCloseLabel":      "Close",
				"AiPulseNoStocksLabel":        "no stocks",
				"StockTickers": []map[string]any{
					{"Ticker": "NVDA", "Name": "NVIDIA"},
					{"Ticker": "MSFT", "Name": "Microsoft"},
				},
				"StockInitialTicker":   "NVDA",
				"StockCompanyName":     "NVIDIA",
				"StockInitialData":     []map[string]any{{"Date": "2026-07-01", "Close": 1.5}, {"Date": "2026-07-02", "Close": 2.0}},
				"StockInitialDataJSON": `[{"date":"2026-07-01","open":1,"high":2,"low":0.5,"close":1.5,"volume":null},{"date":"2026-07-02","open":1.5,"high":2.5,"low":1,"close":2,"volume":null}]`,
				"Trends": []map[string]any{
					{"Title": "Trend one", "Summary": "sum1", "URL": "https://hn.one", "SourceLabel": "HN", "Date": "Jul 1, 2026"},
					{"Title": "Trend two", "Summary": "sum2", "URL": "https://hn.two", "SourceLabel": "", "Date": "Jul 2, 2026"},
				},
				"Repos": []map[string]any{
					{"Title": "owner/repo", "Description": "desc", "URL": "https://gh.repo", "Language": "Go", "StarsToday": 12, "SourceLabel": "GitHub", "Date": "Jul 1, 2026"},
				},
			}
			var buf bytes.Buffer
			if err := r.Render(&buf, "dashboard", data); err != nil {
				t.Fatalf("render dashboard view %q failed: %v", view, err)
			}
			if view == "ai-pulse" {
				out := buf.String()
				updated := data["AiPulseLastUpdatedLabel"].(string)
				for _, want := range []string{
					">Trend one<", ">Trend two<", "sum1", "https://hn.one", "HN",
					">owner/repo<", "desc", "Go", "12 " + data["AiPulseStarsTodayLabel"].(string), "https://gh.repo",
					updated + ": " + "Jul 1, 2026",
				} {
					if !strings.Contains(out, want) {
						t.Errorf("ai-pulse view missing %q", want)
					}
				}
			}
		})
	}
}

// TestConsentBannerGating covers the SSR gate for the cookie-consent partial:
// the banner renders only while ConsentBannerRequired is true (no stored
// karot_consent decision), while the preferences modal and the
// window.__gothConsent script are always present so the dashboard settings
// trigger works on every page.
func TestConsentBannerGating(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	homeData := func(required bool) map[string]any {
		return map[string]any{
			"Locale":                "fi",
			"Theme":                 "dark",
			"Title":                 "Title",
			"Description":           "Description",
			"Canonical":             "https://example.com/fi",
			"OGImage":               "https://example.com/i.svg",
			"SwitchToEn":            "/en",
			"SwitchToFi":            "/fi",
			"Year":                  2026,
			"Tr":                    func(key string) string { return key },
			"Badge":                 "badge",
			"Intro":                 "intro",
			"Body1":                 "b1",
			"Body2":                 "b2",
			"Body3":                 "b3",
			"SentinelConfig":        `{"locale":"fi"}`,
			"ConsentBannerRequired": required,
		}
	}

	bannerMarker := `aria-live="polite"`
	modalMarker := `role="dialog"`
	scriptMarker := `window.__gothConsent`

	t.Run("banner shown while consent unset", func(t *testing.T) {
		var out bytes.Buffer
		if err := r.Render(&out, "home", homeData(true)); err != nil {
			t.Fatalf("render home: %v", err)
		}
		for _, marker := range []string{bannerMarker, modalMarker, scriptMarker, "consent-preferences-title", "consent-toggle-essential", "consent-toggle-marketing"} {
			if !strings.Contains(out.String(), marker) {
				t.Fatalf("expected %q in output", marker)
			}
		}
	})

	t.Run("banner hidden after stored decision", func(t *testing.T) {
		var out bytes.Buffer
		if err := r.Render(&out, "home", homeData(false)); err != nil {
			t.Fatalf("render home: %v", err)
		}
		if strings.Contains(out.String(), bannerMarker) {
			t.Fatal("banner must not render once a consent decision is stored")
		}
		for _, marker := range []string{modalMarker, scriptMarker} {
			if !strings.Contains(out.String(), marker) {
				t.Fatalf("expected %q in output even when banner hidden", marker)
			}
		}
	})
}

func TestSentinelConfigIsEmbeddedAsJSON(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	var out bytes.Buffer
	config := `{"title":"Puhu AI-portsari ympäri"}`
	data := map[string]any{
		"Locale":         "fi",
		"Theme":          "dark",
		"Title":          "Title",
		"Description":    "Description",
		"Canonical":      "https://example.com/fi",
		"SwitchToEn":     "/en",
		"SwitchToFi":     "/fi",
		"Year":           2026,
		"Tr":             func(key string) string { return key },
		"Badge":          "badge",
		"Intro":          "intro",
		"Body1":          "b1",
		"Body2":          "b2",
		"Body3":          "b3",
		"SentinelConfig": config,
	}
	if err := r.Render(&out, "home", data); err != nil {
		t.Fatalf("render home: %v", err)
	}

	want := `<script id="sentinel-config" type="application/json">` + config + `</script>`
	if !strings.Contains(out.String(), want) {
		t.Fatalf("sentinel config was not embedded as raw JSON: %s", out.String())
	}
}

func merge(base, extra map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(extra))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

// discard is a minimal io.Writer for template execution.
type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }

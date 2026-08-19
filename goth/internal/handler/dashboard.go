package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"goth/internal/ai"
	"goth/internal/aipulse"
	"goth/internal/content"
	"goth/internal/db"
	"goth/internal/i18n"
)

// dashboardViews are the valid ?view= values. Only blog/changelog/ai-pulse are
// URL-persisted in the Next.js reference; for the SSR Go build every view is
// reachable via ?view= so deep links and refreshes work without client state.
var dashboardViews = map[string]bool{
	"overview":  true,
	"projects":  true,
	"tech":      true,
	"blog":      true,
	"changelog": true,
	"ai-pulse":  true,
	"settings":  true,
}

func isValidView(v string) bool {
	return dashboardViews[v]
}

// Dashboard renders the full unlocked dashboard with the active view selected
// via ?view= (plus ?page= and ?post= for the blog reader).
func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	locale := h.resolveLocale(r)
	if !h.hasUnlockedBefore(r) {
		http.Redirect(w, r, "/"+locale, http.StatusSeeOther)
		return
	}

	viewParam := r.URL.Query().Get("view")
	if !isValidView(viewParam) {
		viewParam = "overview"
	}

	githubURL := "https://github.com/karot00/karotammela"
	linkedinURL := "https://www.linkedin.com/in/karo-tammela/"

	data := h.common(r, locale)
	tr := func(key string) string { return i18n.T(locale, key) }

	data["ActiveView"] = viewParam
	data["DashBase"] = "/" + locale + "/dashboard"
	data["ProfileImage"] = "/media/Karo-Tammela.jpg"
	data["ProfileImageAlt"] = tr("dashboard.profileImageAlt")
	data["BackLabel"] = tr("dashboard.homeLinkLabel")
	data["MenuLabel"] = tr("dashboard.menuLabel")
	data["NavOverview"] = tr("dashboard.navOverview")
	data["NavProjects"] = tr("dashboard.navProjects")
	data["NavTech"] = tr("dashboard.navTech")
	data["NavBlog"] = tr("dashboard.navBlog")
	data["NavChangelog"] = tr("dashboard.navChangelog")
	data["NavAiPulse"] = tr("dashboard.navAiPulse")
	data["NavSettings"] = tr("dashboard.navSettings")
	data["NavigationLabel"] = tr("dashboard.navigationLabel")
	data["Badge"] = tr("dashboard.badge")
	data["Title"] = tr("dashboard.title")
	data["Description"] = tr("dashboard.description")

	// Stats (overview analytics). Offline-safe.
	stats := db.Stats{}
	offline := true
	if h.conn != nil {
		if s, err := db.GetStats(h.conn, ai.GetAccessCode()); err == nil {
			stats = s
			offline = false
		}
	}
	data["Stats"] = stats
	data["StatsOffline"] = offline
	data["AnalyticsTitle"] = tr("dashboard.analyticsTitle")
	data["TotalAttemptsLabel"] = tr("dashboard.totalAttemptsLabel")
	data["UnlockedCountLabel"] = tr("dashboard.unlockedCountLabel")
	data["DirectUnlockCountLabel"] = tr("dashboard.directUnlockCountLabel")
	data["AvgMessagesToUnlockLabel"] = tr("dashboard.avgMessagesToUnlockLabel")
	data["LatestUnlockLabel"] = tr("dashboard.latestUnlockLabel")
	data["SourceOffline"] = tr("dashboard.sourceOffline")
	data["LatestUnlockNever"] = tr("dashboard.latestUnlockNever")
	data["LatestUnlock"] = ""
	if stats.LatestUnlock != nil {
		data["LatestUnlock"] = content.FormatDateTime(locale, *stats.LatestUnlock)
	}

	// Contact form (overview). Submission endpoint lands in a later phase.
	data["ContactNameLabel"] = tr("dashboard.contactNameLabel")
	data["ContactEmailLabel"] = tr("dashboard.contactEmailLabel")
	data["ContactCompanyLabel"] = tr("dashboard.contactCompanyLabel")
	data["ContactMessageLabel"] = tr("dashboard.contactMessageLabel")
	data["ContactSubmitLabel"] = tr("dashboard.contactSubmitLabel")
	data["ContactPendingLabel"] = tr("dashboard.contactPendingLabel")
	data["ContactSuccessLabel"] = tr("dashboard.contactSuccessLabel")
	data["ContactErrorLabel"] = tr("dashboard.contactErrorLabel")
	data["ContactTitle"] = tr("dashboard.contactTitle")
	data["ContactDescription"] = tr("dashboard.contactDescription")
	data["ContactAvailabilityEyebrow"] = tr("dashboard.contactAvailabilityEyebrow")
	data["ContactAvailabilityStatus"] = tr("dashboard.contactAvailabilityStatus")
	data["ContactConnectLabel"] = tr("dashboard.contactConnectLabel")
	data["ContactGithubLabel"] = tr("dashboard.contactGithubLabel")
	data["ContactLinkedinLabel"] = tr("dashboard.contactLinkedinLabel")
	data["GithubURL"] = githubURL
	data["LinkedinURL"] = linkedinURL

	// About / toolkit.
	data["AboutTitle"] = tr("dashboard.aboutTitle")
	data["AboutBody"] = tr("dashboard.aboutBody")
	data["ToolkitTitle"] = tr("dashboard.toolkitTitle")
	data["CodingModelsLabel"] = tr("dashboard.codingModelsLabel")
	data["ChatbotModelLabel"] = tr("dashboard.chatbotModelLabel")
	data["ToolkitTitleH2"] = tr("dashboard.toolkitTitleH2")
	data["CodingModelsLabelH2"] = tr("dashboard.codingModelsLabelH2")
	data["ChatbotModelLabelH2"] = tr("dashboard.chatbotModelLabelH2")

	// Projects, grouped like the Next.js ProjectsView: production first, then
	// the showcase group. The postcard is intentionally NOT ported to the Go
	// build (owner decision 2026-07-24): the card links to the Next.js
	// deployment and carries a "built with Next.js" badge.
	data["ProjectsTitle"] = tr("dashboard.projectsTitle")
	data["ProjectsProductionSubtitle"] = tr("dashboard.projectsProductionSubtitle")
	data["ProjectsShowcaseSubtitle"] = tr("dashboard.projectsShowcaseSubtitle")
	ghLabel := tr("dashboard.projectGithubLabel")
	data["ProjectsProduction"] = []map[string]any{
		projectCard(tr("dashboard.projectOneTitle"), tr("dashboard.projectOneDescription"), "https://levifinland.fi", "", false, ghLabel, ""),
		projectCard(tr("dashboard.projectTwoTitle"), tr("dashboard.projectTwoDescription"), "https://greenfee.levifinland.fi", "", false, ghLabel, ""),
		projectCard(tr("dashboard.projectThreeTitle"), tr("dashboard.projectThreeDescription"), "", "", false, ghLabel, ""),
		projectCard(tr("dashboard.projectFourTitle"), tr("dashboard.projectFourDescription"), "", githubURL, true, ghLabel, ""),
	}
	postcardURL := strings.TrimRight(h.cfg.BaseURL, "/") + "/" + locale + "/postikortti"
	data["ProjectsShowcase"] = []map[string]any{
		projectCard(tr("dashboard.projectFiveTitle"), tr("dashboard.projectFiveDescription"), postcardURL, "", false, ghLabel, tr("dashboard.projectPostcardStackBadge")),
	}

	// Technologies.
	data["TechTitle"] = tr("dashboard.techTitle")
	data["TechSecondaryTitle"] = tr("dashboard.techSecondaryTitle")
	data["TechCategories"] = []map[string]any{
		techCategory("frontend", tr("dashboard.techCategories.frontend"),
			techItem("nextjs", tr("dashboard.techItems.nextjs.name"), tr("dashboard.techItems.nextjs.description")),
			techItem("tailwindCss", tr("dashboard.techItems.tailwindCss.name"), tr("dashboard.techItems.tailwindCss.description")),
		),
		techCategory("backendDb", tr("dashboard.techCategories.backendDb"),
			techItem("postgresql", tr("dashboard.techItems.postgresql.name"), tr("dashboard.techItems.postgresql.description")),
			techItem("sqlite", tr("dashboard.techItems.sqlite.name"), tr("dashboard.techItems.sqlite.description")),
		),
		techCategory("infrastructure", tr("dashboard.techCategories.infrastructure"),
			techItem("vercel", tr("dashboard.techItems.vercel.name"), tr("dashboard.techItems.vercel.description")),
			techItem("cloudflare", tr("dashboard.techItems.cloudflare.name"), tr("dashboard.techItems.cloudflare.description")),
			techItem("github", tr("dashboard.techItems.github.name"), tr("dashboard.techItems.github.description")),
		),
		techCategory("tools", tr("dashboard.techCategories.tools"),
			techItem("resend", tr("dashboard.techItems.resend.name"), tr("dashboard.techItems.resend.description")),
			techItem("vsCode", tr("dashboard.techItems.vsCode.name"), tr("dashboard.techItems.vsCode.description")),
			techItem("kiloCode", tr("dashboard.techItems.kiloCode.name"), tr("dashboard.techItems.kiloCode.description")),
		),
	}
	data["TechSecondaryCategories"] = []map[string]any{
		techCategory("backendDb", tr("dashboard.techCategories.backendDb"),
			techItem("go", tr("dashboard.techItems.go.name"), tr("dashboard.techItems.go.description")),
			techItem("chi", tr("dashboard.techItems.chi.name"), tr("dashboard.techItems.chi.description")),
			techItem("htmx", tr("dashboard.techItems.htmx.name"), tr("dashboard.techItems.htmx.description")),
			techItem("sqliteWal", tr("dashboard.techItems.sqliteWal.name"), tr("dashboard.techItems.sqliteWal.description")),
		),
		techCategory("infrastructure", tr("dashboard.techCategories.infrastructure"),
			techItem("hetzner", tr("dashboard.techItems.hetzner.name"), tr("dashboard.techItems.hetzner.description")),
			techItem("caddy", tr("dashboard.techItems.caddy.name"), tr("dashboard.techItems.caddy.description")),
			techItem("systemd", tr("dashboard.techItems.systemd.name"), tr("dashboard.techItems.systemd.description")),
			techItem("letsEncrypt", tr("dashboard.techItems.letsEncrypt.name"), tr("dashboard.techItems.letsEncrypt.description")),
		),
		techCategory("tools", tr("dashboard.techCategories.tools"),
			techItem("githubActions", tr("dashboard.techItems.githubActions.name"), tr("dashboard.techItems.githubActions.description")),
			techItem("r2Rclone", tr("dashboard.techItems.r2Rclone.name"), tr("dashboard.techItems.r2Rclone.description")),
		),
	}

	// Blog reader.
	if viewParam == "blog" {
		h.dashboardBlogData(locale, r, data, tr)
	}

	// Changelog.
	if viewParam == "changelog" {
		cl, err := content.GetChangelog(locale)
		if err == nil {
			data["Changelog"] = cl.Releases
		} else {
			data["Changelog"] = []content.ChangelogRelease{}
		}
		data["ChangelogTitle"] = tr("dashboard.changelogTitle")
		data["ChangelogLead"] = tr("dashboard.changelogLead")
		data["ChangelogEmptyLabel"] = tr("dashboard.changelogEmptyLabel")
		data["ChangelogTypeAdded"] = tr("dashboard.changelogTypeAdded")
		data["ChangelogTypeChanged"] = tr("dashboard.changelogTypeChanged")
		data["ChangelogTypeFixed"] = tr("dashboard.changelogTypeFixed")
		data["ChangelogTypeRemoved"] = tr("dashboard.changelogTypeRemoved")
	}

	// Settings.
	if viewParam == "settings" {
		data["SettingsLanguageTitle"] = tr("dashboard.settingsLanguageTitle")
		data["SettingsLanguageDescription"] = tr("dashboard.settingsLanguageDescription")
		data["SettingsCookieTitle"] = tr("dashboard.settingsCookieTitle")
		data["SettingsCookieDescription"] = tr("dashboard.settingsCookieDescription")
		data["SettingsThemeTitle"] = tr("dashboard.settingsThemeTitle")
		data["SettingsThemeDescription"] = tr("dashboard.settingsThemeDescription")
		data["LightModeLabel"] = tr("dashboard.lightModeLabel")
		data["DarkModeLabel"] = tr("dashboard.darkModeLabel")
		data["PrivacyPolicyLink"] = tr("dashboard.privacyPolicyLink")
		data["PrivacyReturn"] = "/" + locale + "/privacy?returnTo=" + urlEncode("/"+locale+"/dashboard?view=settings")
		data["ConsentSettingsTrigger"] = tr("cookieConsent.settingsTrigger")
		data["ConsentSettingsTriggerAria"] = tr("cookieConsent.settingsTriggerAriaLabel")
	}

	// AI Pulse. The stocks chart is wired in 12f (reads Go's local cache);
	// trends/repos lists land in 12g.
	if viewParam == "ai-pulse" {
		data["AiPulseTitle"] = tr("dashboard.aiPulseTitle")
		data["AiPulseDescription"] = tr("dashboard.aiPulseDescription")
		data["AiPulseTrendsTitle"] = tr("dashboard.aiPulseTrendsTitle")
		data["AiPulseReposTitle"] = tr("dashboard.aiPulseReposTitle")
		data["AiPulseStocksTitle"] = tr("dashboard.aiPulseStocksTitle")
		data["AiPulseTickerLabel"] = tr("dashboard.aiPulseTickerLabel")
		data["AiPulseNoTrendsLabel"] = tr("dashboard.aiPulseNoTrendsLabel")
		data["AiPulseNoReposLabel"] = tr("dashboard.aiPulseNoReposLabel")
		data["AiPulseLoadingLabel"] = tr("dashboard.aiPulseLoadingLabel")
		data["AiPulseLastUpdatedLabel"] = tr("dashboard.aiPulseLastUpdatedLabel")
		data["AiPulseSourceLabel"] = tr("dashboard.aiPulseSourceLabel")
		data["AiPulseStarsTodayLabel"] = tr("dashboard.aiPulseStarsTodayLabel")
		data["AiPulseStockChartAria"] = tr("dashboard.aiPulseStockChartAria")
		data["AiPulseStockTableCaption"] = tr("dashboard.aiPulseStockTableCaption")
		data["AiPulseStockDateLabel"] = tr("dashboard.aiPulseStockDateLabel")
		data["AiPulseStockCloseLabel"] = tr("dashboard.aiPulseStockCloseLabel")
		data["AiPulseNoStocksLabel"] = tr("dashboard.aiPulseNoStocksLabel")

		// Trends + repos read from the 12b cache (fail-closed: empty
		// slices when no DB exists, so the lists render their empty states
		// instead of erroring — mirroring the Next.js offline behaviour).
		var trends, repos []map[string]any
		if h.conn != nil {
			if t, err := db.GetLatestTrends(h.conn, ""); err == nil {
				trends = buildTrendCards(locale, t)
			}
			if r, err := db.GetLatestRepos(h.conn, ""); err == nil {
				repos = buildRepoCards(locale, r)
			}
		} else {
			trends = []map[string]any{}
			repos = []map[string]any{}
		}
		data["Trends"] = trends
		data["Repos"] = repos

		// Supported ticker selector (mirrors the Next.js AI_TICKERS set).
		tickerMaps := make([]map[string]string, 0, len(aipulse.AITickers))
		for _, t := range aipulse.AITickers {
			tickerMaps = append(tickerMaps, map[string]string{"Ticker": t.Ticker, "Name": t.Name})
		}
		data["StockTickers"] = tickerMaps

		// Pick the initial ticker from the query (no-JS fallback path), else
		// the first cached ticker, else the first supported ticker.
		initial := strings.TrimSpace(r.URL.Query().Get("ticker"))
		var avail []string
		if h.conn != nil {
			if a, err := db.GetAvailableTickers(h.conn); err == nil {
				avail = a
			}
		}
		if !aipulse.IsValidTicker(initial) {
			if len(avail) > 0 {
				initial = avail[0]
			} else if len(aipulse.AITickers) > 0 {
				initial = aipulse.AITickers[0].Ticker
			}
		}

		var hist []db.Stock
		if h.conn != nil && initial != "" {
			if h2, err := db.GetStockHistory(h.conn, initial, 365); err == nil {
				hist = h2
			}
		}

		data["StockInitialTicker"] = initial
		data["StockCompanyName"] = aipulse.CompanyName(initial)
		data["StockInitialData"] = hist
		// template.JS (not a plain string) so html/template emits the JSON
		// array literally inside the <script> tag instead of re-escaping it
		// as a JS string literal (which would wrap it in quotes and leave
		// JSON.parse() on the client with a string, not an array — see the
		// AI Pulse stock chart bug fix memory entry).
		if b, err := json.Marshal(toStockPoints(hist)); err == nil {
			data["StockInitialDataJSON"] = template.JS(b)
		} else {
			data["StockInitialDataJSON"] = template.JS("[]")
		}
		// The stocks card is always available (the selector is a fixed set);
		// chart/table fall back to empty states when no cache exists yet.
		data["AiPulseAvailable"] = len(tickerMaps) > 0
	}

	h.render(w, "dashboard", data)
}

func (h *Handlers) dashboardBlogData(locale string, r *http.Request, data map[string]any, tr func(string) string) {
	posts, err := content.GetAllBlogPosts(locale)
	if err != nil {
		posts = []content.BlogPost{}
	}
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, e := strconv.Atoi(p); e == nil {
			page = v
		}
	}
	result := content.PaginateBlogPosts(posts, page, 10)

	postParam := strings.TrimSpace(r.URL.Query().Get("post"))
	var selected *content.BlogPost
	missing := false
	if postParam != "" {
		sp, e := content.GetBlogPostBySlug(locale, postParam)
		if e == nil && sp != nil {
			selected = sp
		} else {
			missing = true
		}
	} else if len(result.Items) > 0 {
		sp := result.Items[0]
		selected = &sp
	}

	data["BlogItems"] = result.Items
	data["BlogPage"] = result.Page
	data["BlogPages"] = result.Pages
	data["BlogHasPrev"] = result.HasPrev
	data["BlogHasNext"] = result.HasNext
	data["BlogPrevPage"] = result.Page - 1
	data["BlogNextPage"] = result.Page + 1
	data["BlogSelected"] = selected
	data["BlogMissingPost"] = missing
	data["BlogTotal"] = result.Total
	data["BlogPrevLabel"] = tr("dashboard.blogPreviousLabel")
	data["BlogNextLabel"] = tr("dashboard.blogNextLabel")
	data["BlogPageLabel"] = tr("dashboard.blogPageLabel")
	data["BlogNoPostsLabel"] = tr("dashboard.blogNoPostsLabel")
	data["BlogMissingPostLabel"] = tr("dashboard.blogMissingPostLabel")
	data["BlogPlaceholder"] = tr("dashboard.blogPlaceholder")
	data["BlogBackLabel"] = tr("dashboard.blogBackLabel")
	data["BlogShareLabel"] = tr("dashboard.blogShareLabel")
	data["BlogShareCopiedLabel"] = tr("dashboard.blogShareCopiedLabel")
	data["BlogShareErrorLabel"] = tr("dashboard.blogShareErrorLabel")
	data["BlogDraftBadgeLabel"] = tr("dashboard.blogDraftBadgeLabel")
}

func projectCard(title, description, href, githubHref string, isInternal bool, githubLabel, badge string) map[string]any {
	return map[string]any{
		"Title":       title,
		"Description": description,
		"Href":        href,
		"GithubHref":  githubHref,
		"IsInternal":  isInternal,
		"GithubLabel": githubLabel,
		"Badge":       badge,
	}
}

func techCategory(id, label string, items ...map[string]any) map[string]any {
	return map[string]any{"ID": id, "Label": label, "Items": items}
}

func techItem(id, name, description string) map[string]any {
	return map[string]any{"ID": id, "Name": name, "Description": description}
}

func urlEncode(s string) string {
	return url.QueryEscape(s)
}

// sourceLabel maps an upstream source id to its display badge (mirrors the
// Next.js trends list: "hackernews" -> "HN"). Unknown ids pass through.
func sourceLabel(src string) string {
	switch src {
	case "hackernews":
		return "HN"
	case "github":
		return "GitHub"
	default:
		return src
	}
}

// buildTrendCards localizes each trend for the active locale and flattens it
// into a template-friendly shape (already-resolved summary + date). A nil
// SummaryFi or a non-fi locale falls back to the default English summary.
func buildTrendCards(locale string, trends []db.Trend) []map[string]any {
	out := make([]map[string]any, 0, len(trends))
	for _, t := range trends {
		summary := t.Summary
		if locale == "fi" && t.SummaryFi != nil {
			summary = *t.SummaryFi
		}
		src := ""
		if t.Source != nil {
			src = sourceLabel(*t.Source)
		}
		out = append(out, map[string]any{
			"Title":       t.Title,
			"Summary":     summary,
			"URL":         t.URL,
			"SourceLabel": src,
			"Date":        content.FormatDate(locale, t.Date),
		})
	}
	return out
}

// buildRepoCards mirrors buildTrendCards for repositories: localized
// description, resolved language/stars, and a localized date.
func buildRepoCards(locale string, repos []db.Repo) []map[string]any {
	out := make([]map[string]any, 0, len(repos))
	for _, r := range repos {
		desc := ""
		if r.Description != nil {
			desc = *r.Description
		}
		if locale == "fi" && r.DescriptionFi != nil {
			desc = *r.DescriptionFi
		}
		lang := ""
		if r.Language != nil {
			lang = *r.Language
		}
		out = append(out, map[string]any{
			"Title":       r.RepoFullName,
			"Description": desc,
			"URL":         r.URL,
			"Language":    lang,
			"StarsToday":  r.StarsToday,
			"SourceLabel": sourceLabel(r.Source),
			"Date":        content.FormatDate(locale, r.Date),
		})
	}
	return out
}

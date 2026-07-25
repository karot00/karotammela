package aipulse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/google/uuid"

	"goth/internal/db"
)

// aiMLKeywords mirrors AI_ML_KEYWORDS in src/lib/ai/repos-fetcher.ts.
var aiMLKeywords = []string{
	"ai", "llm", "agent", "ml", "machine learning", "deep learning", "rag",
	"mcp", "transformer", "pytorch", "tensorflow", "inference", "embedding",
	"neural", "vision", "llama", "claude", "gpt", "gemini", "openai",
	"copilot", "nlp", "diffusers", "stable-diffusion", "midjourney",
}

var aiMLPatterns = func() []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(aiMLKeywords))
	for i, kw := range aiMLKeywords {
		out[i] = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(kw) + `\b`)
	}
	return out
}()

// IsAiMlRelated mirrors isAiMlRelated: any keyword matches as a whole word
// (case-insensitive) inside title + " " + description.
func IsAiMlRelated(title, description string) bool {
	text := title + " " + description
	for _, re := range aiMLPatterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

// GitHubRepo is one parsed trending row, mirroring the GitHubRepo interface in
// the reference fetcher.
type GitHubRepo struct {
	RepoFullName string
	URL          string
	Description  string
	Language     string // empty = none (null in the reference)
	Stars        int
	StarsToday   int
}

var starsTodayRe = regexp.MustCompile(`(?i)([\d,]+)\s*stars?`)

// ParseGitHubTrending parses the GitHub Trending page HTML with the same
// selectors as parseGitHubTrending (cheerio) in the reference:
// article.Box-row rows, h2 a href, p.col-9/col-12 description,
// [itemprop=programmingLanguage], a[href$=/stargazers], .float-sm-right
// "N stars today". Unknown shapes simply yield no rows.
func ParseGitHubTrending(doc string) []GitHubRepo {
	root, err := html.Parse(strings.NewReader(doc))
	if err != nil {
		return []GitHubRepo{}
	}
	var repos []GitHubRepo
	for _, article := range findAll(root, isElement("article", hasClass("Box-row"))) {
		link := findFirst(article, isElement("h2", nil))
		var a *html.Node
		if link != nil {
			a = findFirst(link, isElement("a", nil))
		}
		if a == nil {
			continue
		}
		repoLink := strings.TrimSpace(getAttr(a, "href"))
		if repoLink == "" {
			continue
		}
		parts := strings.FieldsFunc(repoLink, func(r rune) bool { return r == '/' })
		if len(parts) < 2 {
			continue
		}
		fullName := parts[0] + "/" + parts[1]

		desc := ""
		if p := findFirst(article, func(n *html.Node) bool {
			return n.Data == "p" && (hasClass("col-9")(n) || hasClass("col-12")(n))
		}); p != nil {
			desc = textContent(p)
		}

		lang := ""
		if l := findFirst(article, func(n *html.Node) bool {
			return getAttr(n, "itemprop") == "programmingLanguage"
		}); l != nil {
			lang = textContent(l)
		}

		stars := 0
		if s := findFirst(article, func(n *html.Node) bool {
			return n.Data == "a" && strings.HasSuffix(getAttr(n, "href"), "/stargazers")
		}); s != nil {
			stars = parseIntComma(textContent(s))
		}

		starsToday := 0
		if st := findFirst(article, hasClass("float-sm-right")); st != nil {
			if m := starsTodayRe.FindStringSubmatch(textContent(st)); m != nil {
				starsToday = parseIntComma(m[1])
			}
		}

		repos = append(repos, GitHubRepo{
			RepoFullName: fullName,
			URL:          "https://github.com" + repoLink,
			Description:  desc,
			Language:     lang,
			Stars:        stars,
			StarsToday:   starsToday,
		})
	}
	if repos == nil {
		return []GitHubRepo{}
	}
	return repos
}

func parseIntComma(s string) int {
	n, err := strconv.Atoi(strings.ReplaceAll(strings.TrimSpace(s), ",", ""))
	if err != nil {
		return 0
	}
	return n
}

// --- minimal x/net/html selector helpers (cheerio equivalents) ---

type nodePred func(*html.Node) bool

func isElement(name string, extra nodePred) nodePred {
	return func(n *html.Node) bool {
		if n.Type != html.ElementNode || n.Data != name {
			return false
		}
		return extra == nil || extra(n)
	}
}

func hasClass(class string) nodePred {
	return func(n *html.Node) bool {
		for _, c := range strings.Fields(getAttr(n, "class")) {
			if c == class {
				return true
			}
		}
		return false
	}
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// findFirst returns the first descendant (in document order) matching pred.
func findFirst(n *html.Node, pred nodePred) *html.Node {
	if n == nil {
		return nil
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if pred(c) {
			return c
		}
		if found := findFirst(c, pred); found != nil {
			return found
		}
	}
	return nil
}

func findAll(n *html.Node, pred nodePred) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			if pred(c) {
				out = append(out, c)
			}
			walk(c)
		}
	}
	walk(n)
	return out
}

// textContent mirrors jQuery .text().trim(): concatenated descendant text,
// whitespace-collapsed and trimmed.
func textContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		if cur.Type == html.TextNode {
			sb.WriteString(cur.Data)
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(sb.String()), " ")
}

// --- fetcher ---

const (
	githubDefaultBaseURL = "https://github.com"
	githubMaxRepos       = 7
	// githubTranslateMaxTokens bounds Gemini usage per description (reference: 60).
	githubTranslateMaxTokens = 60
	// githubUserAgent is the descriptive crawler identity required by the
	// Phase 12.5b task (robots-compliant, low-rate scraping) — deliberately
	// not a browser-spoofing string.
	githubUserAgent = "goth-ai-pulse/1.0 (+https://karotammela.fi; AI Pulse daily refresh)"
)

// githubTrendingPaths are the three trending pages from the reference fetcher.
var githubTrendingPaths = []string{
	"/trending?since=daily",
	"/trending/python?since=daily",
	"/trending/jupyter-notebook?since=daily",
}

// GitHubFetcher is the Go GitHub Trending writer (Phase 12.5b). It fetches
// the trending pages sequentially at a low rate with a descriptive UA, parses
// rows with markup-change detection, filters for AI/ML relevance, translates
// descriptions to Finnish via Gemini with fallback, and yields ai_repos rows.
type GitHubFetcher struct {
	// BaseURL overrides the github.com host for tests/drills.
	BaseURL string
	// Client is the HTTP client; nil uses a 15s-timeout default (bounded).
	Client *http.Client
	// Summarizer translates descriptions; nil selects the deterministic
	// English-description fallback.
	Summarizer TextSummarizer
	// Now supplies the current time; nil uses time.Now.
	Now func() time.Time
	// PageDelay is the pause between page fetches (low-rate scraping); zero
	// fetches back-to-back (tests).
	PageDelay time.Duration
}

func (f *GitHubFetcher) httpClient() *http.Client {
	if f.Client != nil {
		return f.Client
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func (f *GitHubFetcher) baseURL() string {
	if f.BaseURL != "" {
		return f.BaseURL
	}
	return githubDefaultBaseURL
}

func (f *GitHubFetcher) now() time.Time {
	if f.Now != nil {
		return f.Now()
	}
	return time.Now()
}

// FetchRepos runs the GitHub Trending pipeline and returns the rows to
// persist. A failure of one page does not abort the others; an error is
// returned only when every page failed or no page produced any parseable row
// (markup-change detection), so the refresh report can mark the source failed
// while last-known data stays intact.
func (f *GitHubFetcher) FetchRepos(ctx context.Context) ([]db.Repo, error) {
	byName := map[string]GitHubRepo{}
	pagesOK, pageFailures := 0, 0

	for i, path := range githubTrendingPaths {
		if i > 0 && f.PageDelay > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(f.PageDelay):
			}
		}
		repos, err := f.fetchPage(ctx, path)
		if err != nil {
			pageFailures++
			continue
		}
		pagesOK++
		for _, r := range repos {
			byName[r.RepoFullName] = r
		}
	}

	if pagesOK == 0 {
		return nil, fmt.Errorf("all %d trending pages failed", pageFailures)
	}
	if len(byName) == 0 {
		return nil, fmt.Errorf("no repos parsed from %d pages (markup change?)", pagesOK)
	}

	all := make([]GitHubRepo, 0, len(byName))
	for _, r := range byName {
		all = append(all, r)
	}
	filtered := all[:0]
	for _, r := range all {
		if IsAiMlRelated(r.RepoFullName, r.Description) {
			filtered = append(filtered, r)
		}
	}
	// Reference behavior (repos-fetcher.ts): sort by stars gained today,
	// descending, then take the top MAX_REPOS.
	sortByStarsTodayDesc(filtered)
	if len(filtered) > githubMaxRepos {
		filtered = filtered[:githubMaxRepos]
	}

	today := f.now().UTC().Format("2006-01-02")
	createdAt := f.now().UnixMilli()
	out := make([]db.Repo, 0, len(filtered))
	for _, r := range filtered {
		desc := r.Description
		descFi := f.translateDescription(ctx, r.Description)
		var lang *string
		if r.Language != "" {
			l := r.Language
			lang = &l
		}
		out = append(out, db.Repo{
			ID:            uuid.NewString(),
			Date:          today,
			RepoFullName:  r.RepoFullName,
			URL:           r.URL,
			Description:   &desc,
			DescriptionFi: &descFi,
			Language:      lang,
			Stars:         r.Stars,
			StarsToday:    r.StarsToday,
			Source:        "github-trending",
			CreatedAt:     createdAt,
		})
	}
	return out, nil
}

func sortByStarsTodayDesc(repos []GitHubRepo) {
	for i := 1; i < len(repos); i++ {
		for j := i; j > 0 && repos[j-1].StarsToday < repos[j].StarsToday; j-- {
			repos[j-1], repos[j] = repos[j], repos[j-1]
		}
	}
}

// fetchPage GETs one trending page and parses it. A 200 page that parses to
// zero rows is NOT an error by itself (the AI/ML filter may legitimately
// exclude everything on a quiet day); zero-total detection happens in
// FetchRepos.
func (f *GitHubFetcher) fetchPage(ctx context.Context, path string) ([]GitHubRepo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.baseURL()+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", githubUserAgent)
	req.Header.Set("Accept", "text/html")
	resp, err := f.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	return ParseGitHubTrending(string(body)), nil
}

// translateDescription mirrors translateDescription: empty stays empty,
// Gemini failure (or no summarizer) falls back to the English original.
func (f *GitHubFetcher) translateDescription(ctx context.Context, description string) string {
	if description == "" {
		return ""
	}
	if f.Summarizer == nil {
		return description
	}
	prompt := fmt.Sprintf(`Käännä seuraava GitHub-repositorion kuvaus suomeksi.

Ohjeet:
- Anna suoraan vain se yksi lopullinen käännös.
- ÄLÄ missään tapauksessa anna useita vaihtoehtoja tai selityksiä (kuten "Tässä muutama vaihtoehto:").
- Pidä käännös erittäin ytimekkäänä (maksimissaan 1 lause) ja kehittäjille luontevana.

Kuvaus: "%s"`, description)
	text, err := f.Summarizer.Generate(ctx, prompt, githubTranslateMaxTokens)
	if err != nil {
		return description
	}
	return text
}

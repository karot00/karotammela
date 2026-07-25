package aipulse

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func loadFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/github-trending.html")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(data)
}

// Mirrors the saved-fixture assertions in src/lib/ai/repos-fetcher.test.ts.
func TestParseGitHubTrendingFixture(t *testing.T) {
	repos := ParseGitHubTrending(loadFixture(t))
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	first := repos[0]
	if first.RepoFullName != "meta-llama/llama3" {
		t.Fatalf("fullName = %q", first.RepoFullName)
	}
	if first.URL != "https://github.com/meta-llama/llama3" {
		t.Fatalf("url = %q", first.URL)
	}
	wantDesc := "Llama 3 model code and model definitions. An advanced LLM architecture."
	if first.Description != wantDesc {
		t.Fatalf("description = %q", first.Description)
	}
	if first.Language != "Python" {
		t.Fatalf("language = %q", first.Language)
	}
	if first.Stars != 15234 {
		t.Fatalf("stars = %d", first.Stars)
	}
	if first.StarsToday != 1245 {
		t.Fatalf("starsToday = %d", first.StarsToday)
	}
	second := repos[1]
	if second.RepoFullName != "owner/webscraper" || second.Language != "TypeScript" ||
		second.Stars != 456 || second.StarsToday != 23 {
		t.Fatalf("second repo mismatch: %+v", second)
	}
}

func TestIsAiMlRelated(t *testing.T) {
	cases := []struct {
		title, desc string
		want        bool
	}{
		{"meta-llama/llama3", "Llama 3 model code and model definitions. An advanced LLM architecture.", true},
		{"owner/my-cool-agent", "An agent system built on top of Claude.", true},
		{"owner/simple-web-app", "Just a normal web dashboard with react and tailwind.", false},
		{"owner/stable-diffusion-webui", "One stop shop.", true},
		{"owner/email", "Send email easily.", false}, // "ai" must match a whole word
	}
	for _, c := range cases {
		if got := IsAiMlRelated(c.title, c.desc); got != c.want {
			t.Errorf("IsAiMlRelated(%q, %q) = %v, want %v", c.title, c.desc, got, c.want)
		}
	}
}

func TestParseGitHubTrendingMalformed(t *testing.T) {
	if got := ParseGitHubTrending("<html><body>Nothing here</body></html>"); len(got) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(got))
	}
	if got := ParseGitHubTrending(""); len(got) != 0 {
		t.Fatalf("expected 0 repos for empty input, got %d", len(got))
	}
}

// ghServer serves each trending path with the given HTML or a status code.
type ghServer struct {
	pages  map[string]string
	status map[string]int
}

func (s *ghServer) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path
		if r.URL.RawQuery != "" {
			key += "?" + r.URL.RawQuery
		}
		if code, ok := s.status[key]; ok {
			w.WriteHeader(code)
			return
		}
		if body, ok := s.pages[key]; ok {
			w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	})
}

func repoHTML(fullName, desc, lang string, stars, starsToday int) string {
	return fmt.Sprintf(`<main><div class="Box"><article class="Box-row">
<h2 class="h3 lh-condensed"><a href="/%s">%s</a></h2>
<p class="col-9 color-fg-muted my-1 pr-4">%s</p>
<div class="f6 color-fg-muted mt-2">
<span class="d-inline-block mr-3"><span itemprop="programmingLanguage">%s</span></span>
<a href="/%s/stargazers" class="Link--muted d-inline-block mr-3">%d</a>
<span class="d-inline-block float-sm-right">%d stars today</span>
</div></article></div></main>`, fullName, fullName, desc, lang, fullName, stars, starsToday)
}

func trendingPathsWith(bodies ...string) map[string]string {
	out := map[string]string{}
	for i, p := range githubTrendingPaths {
		if i < len(bodies) {
			out[p] = bodies[i]
		} else {
			out[p] = ""
		}
	}
	return out
}

func TestGitHubFetcherFiltersSortsAndTranslates(t *testing.T) {
	page1 := repoHTML("a/llm-kit", "An LLM toolkit.", "Python", 100, 50) +
		repoHTML("b/web-app", "A plain dashboard.", "Go", 10, 5) // filtered out
	page2 := repoHTML("c/agent-fw", "An agent framework.", "TypeScript", 200, 300)
	srv := &ghServer{pages: trendingPathsWith(page1, page2, "")}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	sum := &fakeSummarizer{text: "käännös"}
	f := &GitHubFetcher{BaseURL: ts.URL, Summarizer: sum, Now: fixedNow}
	repos, err := f.FetchRepos(context.Background())
	if err != nil {
		t.Fatalf("FetchRepos: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 AI/ML repos, got %d", len(repos))
	}
	// Sorted by stars today desc: c/agent-fw (300) before a/llm-kit (50).
	if repos[0].RepoFullName != "c/agent-fw" || repos[1].RepoFullName != "a/llm-kit" {
		t.Fatalf("bad order: %s, %s", repos[0].RepoFullName, repos[1].RepoFullName)
	}
	if repos[0].Source != "github-trending" {
		t.Fatalf("source = %q", repos[0].Source)
	}
	if repos[0].DescriptionFi == nil || *repos[0].DescriptionFi != "FI: käännös" {
		t.Fatalf("FI translation missing: %+v", repos[0].DescriptionFi)
	}
	if repos[0].Description == nil || *repos[0].Description != "An agent framework." {
		t.Fatalf("EN description missing: %+v", repos[0].Description)
	}
	if repos[0].Language == nil || *repos[0].Language != "TypeScript" {
		t.Fatalf("language missing: %+v", repos[0].Language)
	}
	// Two selected repos with non-empty descriptions → two translate calls.
	if got := sum.callCount(); got != 2 {
		t.Fatalf("translate calls = %d, want 2", got)
	}
	if !strings.Contains(sum.calls[0], "Käännä seuraava GitHub-repositorion kuvaus suomeksi.") {
		t.Fatalf("translation prompt mismatch: %q", sum.calls[0])
	}
	if repos[0].Date != fixedNow().UTC().Format("2006-01-02") || repos[0].ID == "" {
		t.Fatalf("row metadata missing: %+v", repos[0])
	}
}

func TestGitHubFetcherDedupsAcrossPages(t *testing.T) {
	row := repoHTML("a/llm-kit", "An LLM toolkit.", "Python", 100, 50)
	srv := &ghServer{pages: trendingPathsWith(row, row, row)}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &GitHubFetcher{BaseURL: ts.URL, Now: fixedNow}
	repos, err := f.FetchRepos(context.Background())
	if err != nil {
		t.Fatalf("FetchRepos: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 deduped repo, got %d", len(repos))
	}
}

func TestGitHubFetcherPartialFailureTolerated(t *testing.T) {
	srv := &ghServer{
		pages:  trendingPathsWith(repoHTML("a/llm-kit", "An LLM toolkit.", "Python", 1, 1)),
		status: map[string]int{"/trending/python?since=daily": 500, "/trending/jupyter-notebook?since=daily": 503},
	}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &GitHubFetcher{BaseURL: ts.URL, Now: fixedNow}
	repos, err := f.FetchRepos(context.Background())
	if err != nil {
		t.Fatalf("partial failure must not error: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo from surviving page, got %d", len(repos))
	}
}

func TestGitHubFetcherAllPagesFailErrors(t *testing.T) {
	srv := &ghServer{status: map[string]int{
		"/trending?since=daily":                  500,
		"/trending/python?since=daily":           500,
		"/trending/jupyter-notebook?since=daily": 500,
	}}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &GitHubFetcher{BaseURL: ts.URL, Now: fixedNow}
	if _, err := f.FetchRepos(context.Background()); err == nil {
		t.Fatal("expected error when all pages fail")
	}
}

func TestGitHubFetcherMarkupChangeDetected(t *testing.T) {
	// Every page 200-OK but unparsable → markup-change error, not silent empty.
	srv := &ghServer{pages: trendingPathsWith("<html>redesigned</html>", "<html>redesigned</html>", "<html>redesigned</html>")}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &GitHubFetcher{BaseURL: ts.URL, Now: fixedNow}
	_, err := f.FetchRepos(context.Background())
	if err == nil || !strings.Contains(err.Error(), "markup") {
		t.Fatalf("expected markup-change error, got %v", err)
	}
}

func TestGitHubFetcherNilSummarizerFallsBackToEnglish(t *testing.T) {
	srv := &ghServer{pages: trendingPathsWith(repoHTML("a/llm-kit", "An LLM toolkit.", "Python", 1, 1))}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &GitHubFetcher{BaseURL: ts.URL, Now: fixedNow}
	repos, err := f.FetchRepos(context.Background())
	if err != nil {
		t.Fatalf("FetchRepos: %v", err)
	}
	if len(repos) != 1 || *repos[0].DescriptionFi != "An LLM toolkit." {
		t.Fatalf("expected English fallback in description_fi: %+v", repos)
	}
}

func TestGitHubFetcherSendsDescriptiveUA(t *testing.T) {
	var ua string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		w.Write([]byte(repoHTML("a/llm-kit", "An LLM toolkit.", "Python", 1, 1)))
	}))
	defer ts.Close()

	f := &GitHubFetcher{BaseURL: ts.URL, Now: fixedNow}
	if _, err := f.FetchRepos(context.Background()); err != nil {
		t.Fatalf("FetchRepos: %v", err)
	}
	if !strings.HasPrefix(ua, "goth-ai-pulse/") || !strings.Contains(ua, "karotammela.fi") {
		t.Fatalf("UA not descriptive: %q", ua)
	}
}

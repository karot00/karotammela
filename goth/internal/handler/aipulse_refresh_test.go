package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"goth/internal/aipulse"
	"goth/internal/config"
)

// staticFetcher builds a Refresher whose sources all succeed instantly using
// local test servers (one trend, one repo, one stock row each).
func refreshTestSetup(t *testing.T) (*Handlers, string) {
	t.Helper()
	conn := newTestDB(t)

	hn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"hits":[{"title":"AI story","url":"https://example.com/a","points":50,"objectID":"1"},
			{"title":"LLM story","url":"https://example.com/b","points":40,"objectID":"2"},
			{"title":"GPT story","url":"https://example.com/c","points":30,"objectID":"3"},
			{"title":"Claude story","url":"https://example.com/d","points":20,"objectID":"4"},
			{"title":"Gemini story","url":"https://example.com/e","points":10,"objectID":"5"}]}`))
	}))
	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<main><div class="Box"><article class="Box-row">
<h2><a href="/a/llm-kit">a/llm-kit</a></h2><p class="col-9">An LLM toolkit.</p>
<div><a href="/a/llm-kit/stargazers">100</a><span class="float-sm-right">5 stars today</span></div>
</article></div></main>`))
	}))
	yh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"chart":{"result":[{"timestamp":[1800000000],"indicators":{"quote":[{"open":[1.0],"high":[2.0],"low":[0.5],"close":[1.5],"volume":[100]}]}}],"error":null}}`))
	}))
	t.Cleanup(func() { hn.Close(); gh.Close(); yh.Close() })

	refresher := &aipulse.Refresher{
		Trends: &aipulse.HNFetcher{BaseURL: hn.URL},
		Repos:  &aipulse.GitHubFetcher{BaseURL: gh.URL},
		Stocks: &aipulse.YahooFetcher{BaseURL: yh.URL},
	}
	h := &Handlers{cfg: &config.Config{CronSecret: "test-secret"}, conn: conn}
	h.SetRefresher(refresher)
	return h, "test-secret"
}

func refreshRequest(t *testing.T, h *Handlers, auth, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/ai-pulse/refresh"+query, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	w := httptest.NewRecorder()
	h.AiPulseRefresh(w, req)
	return w
}

func TestAiPulseRefreshUnauthorized(t *testing.T) {
	h, _ := refreshTestSetup(t)

	// No header.
	if w := refreshRequest(t, h, "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("no header: %d", w.Code)
	}
	// Wrong token.
	if w := refreshRequest(t, h, "Bearer wrong", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token: %d", w.Code)
	}
	// Body must be the reference error shape.
	w := refreshRequest(t, h, "", "")
	if strings.TrimSpace(w.Body.String()) != `{"error":"Unauthorized"}` {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestAiPulseRefreshNoSecretConfiguredIs401(t *testing.T) {
	conn := newTestDB(t)
	h := &Handlers{cfg: &config.Config{}, conn: conn}
	h.SetRefresher(&aipulse.Refresher{})
	if w := refreshRequest(t, h, "Bearer x", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (mirror reference)", w.Code)
	}
}

func TestAiPulseRefreshNotConfigured(t *testing.T) {
	h := &Handlers{cfg: &config.Config{CronSecret: "s"}}
	if w := refreshRequest(t, h, "Bearer s", ""); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
}

func TestAiPulseRefreshSuccessShape(t *testing.T) {
	h, secret := refreshTestSetup(t)
	w := refreshRequest(t, h, "Bearer "+secret, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp struct {
		OK      bool   `json:"ok"`
		RanAt   string `json:"ranAt"`
		Sources struct {
			Trends struct {
				OK          bool    `json:"ok"`
				Inserted    int     `json:"inserted"`
				SkippedDup  int     `json:"skippedDup"`
				WindowHours int     `json:"windowHours"`
				Error       *string `json:"error"`
			} `json:"trends"`
			Repos struct {
				OK       bool    `json:"ok"`
				Inserted int     `json:"inserted"`
				Error    *string `json:"error"`
			} `json:"repos"`
			Stocks struct {
				OK       bool    `json:"ok"`
				Inserted int     `json:"inserted"`
				Error    *string `json:"error"`
			} `json:"stocks"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v (%s)", err, w.Body.String())
	}
	if !resp.OK || resp.RanAt == "" {
		t.Fatalf("top-level shape wrong: %+v", resp)
	}
	// ranAt must look like a Node Date ISO string.
	if _, err := time.Parse("2006-01-02T15:04:05.000Z07:00", resp.RanAt); err != nil {
		t.Fatalf("ranAt not ISO: %q", resp.RanAt)
	}
	tr := resp.Sources.Trends
	if !tr.OK || tr.Inserted != 5 || tr.WindowHours != 24 || tr.Error != nil {
		t.Fatalf("trends entry wrong: %+v", tr)
	}
	if !resp.Sources.Repos.OK || resp.Sources.Repos.Inserted != 1 || resp.Sources.Repos.Error != nil {
		t.Fatalf("repos entry wrong: %+v", resp.Sources.Repos)
	}
	if !resp.Sources.Stocks.OK || resp.Sources.Stocks.Inserted != len(aipulse.AITickers) || resp.Sources.Stocks.Error != nil {
		t.Fatalf("stocks entry wrong: %+v", resp.Sources.Stocks)
	}
}

func TestAiPulseRefreshSourceFailureRedactedAndDebug(t *testing.T) {
	conn := newTestDB(t)
	// All HN queries 500 → trends fetch failure; other sources fine.
	hn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<article class="Box-row"><h2><a href="/a/llm-kit">a/llm-kit</a></h2><p class="col-9">An LLM toolkit.</p></article>`))
	}))
	yh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"chart":{"result":[{"timestamp":[1800000000],"indicators":{"quote":[{"open":[1.0],"high":[2.0],"low":[0.5],"close":[1.5],"volume":[100]}]}}],"error":null}}`))
	}))
	t.Cleanup(func() { hn.Close(); gh.Close(); yh.Close() })

	refresher := &aipulse.Refresher{
		Trends: &aipulse.HNFetcher{BaseURL: hn.URL},
		Repos:  &aipulse.GitHubFetcher{BaseURL: gh.URL},
		Stocks: &aipulse.YahooFetcher{BaseURL: yh.URL},
	}
	h := &Handlers{cfg: &config.Config{CronSecret: "s"}, conn: conn}
	h.SetRefresher(refresher)

	// Default: redacted.
	w := refreshRequest(t, h, "Bearer s", "")
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	trends := resp["sources"].(map[string]any)["trends"].(map[string]any)
	if trends["ok"].(bool) != false || trends["error"].(string) != "Fetch failed" {
		t.Fatalf("redacted trends entry: %+v", trends)
	}
	if resp["sources"].(map[string]any)["repos"].(map[string]any)["ok"].(bool) != true {
		t.Fatalf("repos must survive HN outage: %+v", resp)
	}

	// ?debug=1: raw detail.
	w = refreshRequest(t, h, "Bearer s", "?debug=1")
	json.Unmarshal(w.Body.Bytes(), &resp)
	trends = resp["sources"].(map[string]any)["trends"].(map[string]any)
	if !strings.Contains(trends["error"].(string), "all HN queries failed") {
		t.Fatalf("debug trends entry: %+v", trends)
	}
}

func TestAiPulseRefreshOverlapConflict(t *testing.T) {
	conn := newTestDB(t)
	release := make(chan struct{})
	started := make(chan struct{})
	var once sync.Once
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(started) })
		<-release
		w.Write([]byte(`{"chart":{"result":[],"error":null}}`))
	}))
	t.Cleanup(func() { slow.Close() })

	refresher := &aipulse.Refresher{
		Trends: &aipulse.HNFetcher{BaseURL: "http://127.0.0.1:1"},
		Repos:  &aipulse.GitHubFetcher{BaseURL: "http://127.0.0.1:1"},
		Stocks: &aipulse.YahooFetcher{BaseURL: slow.URL},
	}
	h := &Handlers{cfg: &config.Config{CronSecret: "s"}, conn: conn}
	h.SetRefresher(refresher)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		refreshRequest(t, h, "Bearer s", "")
	}()
	<-started

	w := refreshRequest(t, h, "Bearer s", "")
	if w.Code != http.StatusConflict {
		t.Fatalf("overlap status = %d, want 409; body=%s", w.Code, w.Body.String())
	}
	close(release)
	wg.Wait()
}

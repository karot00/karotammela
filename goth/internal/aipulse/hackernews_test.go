package aipulse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeSummarizer records Generate calls and returns scripted results.
type fakeSummarizer struct {
	mu    sync.Mutex
	calls []string
	text  string
	err   error
}

func (s *fakeSummarizer) Generate(_ context.Context, prompt string, _ int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, prompt)
	if s.err != nil {
		return "", s.err
	}
	if strings.Contains(prompt, "Vastaa suomeksi") || strings.Contains(prompt, "Käännä seuraava") {
		return "FI: " + s.text, nil
	}
	return "EN: " + s.text, nil
}

func (s *fakeSummarizer) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

// hnServer emulates the Algolia HN API. hits are treated as stories created at
// serveFromTs: they are served only when the request's created_at_i filter is
// <= serveFromTs. fail makes every query 500.
type hnServer struct {
	serveFromTs int64
	hits        []hnHit
	fail        bool
}

func (s *hnServer) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.fail {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		nf := r.URL.Query().Get("numericFilters")
		var minTs int64
		if _, err := fmt.Sscanf(nf, "created_at_i>%d", &minTs); err != nil {
			t.Errorf("bad numericFilters %q: %v", nf, err)
		}
		if r.URL.Query().Get("tags") != "story" || r.URL.Query().Get("hitsPerPage") != "30" {
			t.Errorf("missing tags/hitsPerPage in %s", r.URL.RawQuery)
		}
		if minTs > s.serveFromTs {
			json.NewEncoder(w).Encode(hnSearchResponse{Hits: []hnHit{}})
			return
		}
		json.NewEncoder(w).Encode(hnSearchResponse{Hits: s.hits})
	})
}

func makeHits(n int, prefix string) []hnHit {
	out := make([]hnHit, n)
	for i := 0; i < n; i++ {
		out[i] = hnHit{
			Title:  fmt.Sprintf("%s story %d", prefix, i),
			URL:    fmt.Sprintf("https://example.com/%s-%d", prefix, i),
			Points: 100 - i,
		}
	}
	return out
}

func fixedNow() time.Time { return time.Unix(1_800_000_000, 0) }

func TestHNFetcherSelectsTopByPointsAndSummarizes(t *testing.T) {
	srv := &hnServer{serveFromTs: fixedNow().Unix() - 24*3600, hits: makeHits(10, "ai")}
	ts := httptest.NewServer(srv.handler(t))
	defer ts.Close()

	sum := &fakeSummarizer{text: "summary"}
	f := &HNFetcher{BaseURL: ts.URL, Summarizer: sum, Now: fixedNow}

	trends, stats, err := f.FetchTrends(context.Background(), map[string]bool{})
	if err != nil {
		t.Fatalf("FetchTrends: %v", err)
	}
	if len(trends) != hnMaxSelected {
		t.Fatalf("expected %d trends, got %d", hnMaxSelected, len(trends))
	}
	// Points descending: story 0 (100 pts) first, story 6 (94) last; 7..9 cut.
	if trends[0].Title != "ai story 0" || trends[6].Title != "ai story 6" {
		t.Fatalf("unexpected order: %s ... %s", trends[0].Title, trends[6].Title)
	}
	if trends[0].Summary != "EN: summary" {
		t.Fatalf("EN summary not applied: %q", trends[0].Summary)
	}
	if trends[0].SummaryFi == nil || *trends[0].SummaryFi != "FI: summary" {
		t.Fatalf("FI summary not applied: %+v", trends[0].SummaryFi)
	}
	if trends[0].Source == nil || *trends[0].Source != "hackernews" {
		t.Fatalf("source not hackernews: %+v", trends[0].Source)
	}
	if trends[0].Date != fixedNow().UTC().Format("2006-01-02") {
		t.Fatalf("date not UTC today: %q", trends[0].Date)
	}
	if trends[0].ID == "" || trends[0].CreatedAt == 0 {
		t.Fatalf("id/createdAt not populated: %+v", trends[0])
	}
	// 7 stories × 2 languages.
	if got := sum.callCount(); got != 14 {
		t.Fatalf("expected 14 summarizer calls, got %d", got)
	}
	if stats.WindowHours != 24 || stats.CandidatePool != 10 || stats.DedupedOut != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestHNFetcherDedupsBeforeSummarizing(t *testing.T) {
	hits := makeHits(8, "dup")
	// Same URL twice across queries (deduped in-batch) + one excluded.
	hits[1].URL = hits[0].URL
	srv := &hnServer{serveFromTs: fixedNow().Unix() - 24*3600, hits: hits}
	ts := httptest.NewServer(srv.handler(t))
	defer ts.Close()

	sum := &fakeSummarizer{text: "s"}
	f := &HNFetcher{BaseURL: ts.URL, Summarizer: sum, Now: fixedNow}
	exclude := map[string]bool{"https://example.com/dup-2": true}

	trends, stats, err := f.FetchTrends(context.Background(), exclude)
	if err != nil {
		t.Fatalf("FetchTrends: %v", err)
	}
	// 8 raw hits − 1 in-batch dup − 1 excluded = 6 unique.
	if stats.CandidatePool != 7 { // URL-keyed map merges the two dup-0 rows
		t.Fatalf("candidatePool = %d, want 7", stats.CandidatePool)
	}
	if stats.DedupedOut != 1 {
		t.Fatalf("dedupedOut = %d, want 1 (the excluded row)", stats.DedupedOut)
	}
	if len(trends) != 6 {
		t.Fatalf("expected 6 trends, got %d", len(trends))
	}
	for _, tr := range trends {
		if exclude[tr.URL] {
			t.Fatalf("excluded URL was summarized: %s", tr.URL)
		}
	}
	if got := sum.callCount(); got != 12 {
		t.Fatalf("summarizer called %d times, want 12 (dedup before Gemini)", got)
	}
}

func TestHNFetcherExpandsWindowWhenUnderfilled(t *testing.T) {
	// 24h/48h windows empty; 72h serves only 3 hits (< 5) → expand to 168h,
	// which also serves 3 and is the last window → partial selection returned.
	srv := &hnServer{serveFromTs: fixedNow().Unix() - 72*3600, hits: makeHits(3, "old")}
	ts := httptest.NewServer(srv.handler(t))
	defer ts.Close()

	f := &HNFetcher{BaseURL: ts.URL, Now: fixedNow}
	trends, stats, err := f.FetchTrends(context.Background(), map[string]bool{})
	if err != nil {
		t.Fatalf("FetchTrends: %v", err)
	}
	if len(trends) != 3 {
		t.Fatalf("expected 3 partial trends, got %d", len(trends))
	}
	if stats.WindowHours != 168 {
		t.Fatalf("windowHours = %d, want 168", stats.WindowHours)
	}
	// Nil summarizer → deterministic title fallback.
	if trends[0].Summary != trends[0].Title || *trends[0].SummaryFi != trends[0].Title {
		t.Fatalf("fallback content not applied: %+v", trends[0])
	}
}

func TestHNFetcherAllQueriesFailReturnsError(t *testing.T) {
	srv := &hnServer{fail: true}
	ts := httptest.NewServer(srv.handler(t))
	defer ts.Close()

	f := &HNFetcher{BaseURL: ts.URL, Now: fixedNow}
	_, _, err := f.FetchTrends(context.Background(), map[string]bool{})
	if err == nil {
		t.Fatal("expected error when every query fails")
	}
}

func TestHNFetcherZeroHitsAllWindowsReturnsEmpty(t *testing.T) {
	srv := &hnServer{serveFromTs: fixedNow().Unix(), hits: []hnHit{}}
	ts := httptest.NewServer(srv.handler(t))
	defer ts.Close()

	f := &HNFetcher{BaseURL: ts.URL, Now: fixedNow}
	trends, _, err := f.FetchTrends(context.Background(), map[string]bool{})
	if err != nil {
		t.Fatalf("zero hits must not error (fallback shows last-known): %v", err)
	}
	if len(trends) != 0 {
		t.Fatalf("expected 0 trends, got %d", len(trends))
	}
}

func TestHNFetcherSummarizerErrorFallsBackToTitle(t *testing.T) {
	srv := &hnServer{serveFromTs: fixedNow().Unix() - 24*3600, hits: makeHits(5, "err")}
	ts := httptest.NewServer(srv.handler(t))
	defer ts.Close()

	f := &HNFetcher{BaseURL: ts.URL, Summarizer: &fakeSummarizer{err: fmt.Errorf("quota")}, Now: fixedNow}
	trends, _, err := f.FetchTrends(context.Background(), map[string]bool{})
	if err != nil {
		t.Fatalf("FetchTrends: %v", err)
	}
	if len(trends) != 5 {
		t.Fatalf("expected 5 trends, got %d", len(trends))
	}
	for _, tr := range trends {
		if tr.Summary != tr.Title || *tr.SummaryFi != tr.Title {
			t.Fatalf("quota failure must fall back to title: %+v", tr)
		}
	}
}

func TestHNFetcherPromptParity(t *testing.T) {
	sum := &fakeSummarizer{text: "x"}
	f := &HNFetcher{Summarizer: sum}
	f.summarizeStory(context.Background(), "Some Title", "https://example.com/x")

	if len(sum.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(sum.calls))
	}
	// EN and FI run concurrently; order is nondeterministic.
	var en, fi string
	for _, c := range sum.calls {
		if strings.Contains(c, "Vastaa suomeksi") {
			fi = c
		} else {
			en = c
		}
	}
	wantEn := `Summarize this tech news story in 1–2 sentences for a developer audience. Be concise and factual. Title: "Some Title". URL: https://example.com/x`
	wantFi := `Tiivistä tämä teknologiauutinen 1–2 lauseeseen kehittäjäyleisölle. Ole ytimekäs ja asiallinen. Vastaa suomeksi. Otsikko: "Some Title". URL: https://example.com/x`
	if en != wantEn {
		t.Fatalf("EN prompt mismatch:\n got: %q\nwant: %q", en, wantEn)
	}
	if fi != wantFi {
		t.Fatalf("FI prompt mismatch:\n got: %q\nwant: %q", fi, wantFi)
	}
}

func TestGeminiSummarizerRequestResponse(t *testing.T) {
	var gotAuth, gotCT string
	var gotBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("x-goog-api-key")
		gotCT = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"candidates": []any{
				map[string]any{"content": map[string]any{
					"parts": []any{
						map[string]any{"text": "  Hello "},
						map[string]any{"text": "world  "},
					},
				}},
			},
		})
	}))
	defer ts.Close()

	g := &GeminiSummarizer{APIKey: "k", Model: "m", BaseURL: ts.URL}
	out, err := g.Generate(context.Background(), "prompt here", 120)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out != "Hello world" {
		t.Fatalf("unexpected concat/trim: %q", out)
	}
	if gotAuth != "k" || gotCT != "application/json" {
		t.Fatalf("headers wrong: %q %q", gotAuth, gotCT)
	}
	cfg := gotBody["generationConfig"].(map[string]any)
	if cfg["maxOutputTokens"].(float64) != 120 {
		t.Fatalf("maxOutputTokens not bounded: %v", cfg)
	}
}

func TestGeminiSummarizerNon200RedactsBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"secret quota detail that must not leak"}`))
	}))
	defer ts.Close()

	g := &GeminiSummarizer{APIKey: "k", Model: "m", BaseURL: ts.URL}
	_, err := g.Generate(context.Background(), "p", 60)
	if err == nil {
		t.Fatal("expected error on 429")
	}
	if strings.Contains(err.Error(), "secret") || !strings.Contains(err.Error(), "429") {
		t.Fatalf("error must carry status only, got %q", err.Error())
	}
}

func TestGeminiSummarizerEmptyTextIsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"candidates": []any{}})
	}))
	defer ts.Close()

	g := &GeminiSummarizer{APIKey: "k", Model: "m", BaseURL: ts.URL}
	if _, err := g.Generate(context.Background(), "p", 60); err == nil {
		t.Fatal("expected error on empty candidates")
	}
}

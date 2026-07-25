package aipulse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"goth/internal/db"
)

// hnWindowsHours is the stepwise freshness window expansion (24h → 48h → 72h
// → 7d) from src/lib/ai/trends-fetcher.ts.
var hnWindowsHours = []int{24, 48, 72, 168}

// hnSearchQueries are the Algolia story queries from the reference fetcher.
var hnSearchQueries = []string{"AI", "LLM", "machine learning", "GPT", "Claude", "Gemini"}

const (
	hnDefaultBaseURL = "https://hn.algolia.com"
	hnMaxSelected    = 7
	hnMinFresh       = 5
	// hnSummaryMaxTokens bounds Gemini usage per summary (reference: 120).
	hnSummaryMaxTokens = 120
)

// TrendStats mirrors the stats object returned by fetchAndSummarizeTrends so
// the refresh report can log windowHours/candidatePool/dedupedOut exactly like
// the Next.js route.
type TrendStats struct {
	WindowHours   int
	CandidatePool int
	DedupedOut    int
}

type hnHit struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Points   int    `json:"points"`
	ObjectID string `json:"objectID"`
}

type hnSearchResponse struct {
	Hits []hnHit `json:"hits"`
}

// HNFetcher is the Go Hacker News writer (Phase 12.5a), a port of
// src/lib/ai/trends-fetcher.ts. It queries the Algolia HN API over stepwise
// time windows, dedups against recent URLs before any Gemini call, summarizes
// the top stories in EN+FI with bounded tokens, and falls back to the raw
// title when Gemini is unavailable or fails.
type HNFetcher struct {
	// BaseURL overrides the Algolia host for tests/drills; empty uses the real one.
	BaseURL string
	// Client is the HTTP client; nil uses a 15s-timeout default (bounded).
	Client *http.Client
	// Summarizer generates the EN/FI summaries; nil selects the deterministic
	// title fallback (offline behavior).
	Summarizer TextSummarizer
	// Now supplies the current time; nil uses time.Now (UTC dates).
	Now func() time.Time
}

func (f *HNFetcher) httpClient() *http.Client {
	if f.Client != nil {
		return f.Client
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func (f *HNFetcher) baseURL() string {
	if f.BaseURL != "" {
		return f.BaseURL
	}
	return hnDefaultBaseURL
}

func (f *HNFetcher) now() time.Time {
	if f.Now != nil {
		return f.Now()
	}
	return time.Now()
}

// FetchTrends runs the full HN pipeline and returns the rows to persist plus
// the run stats. excludeURLs carries the recent-URL dedup set (db.GetRecentTrendUrls).
//
// Error semantics: an error is returned only when every Algolia query in every
// window failed (total outage) so the refresh report can mark the source
// failed; zero-hit windows and under-filled final windows return an empty or
// partial slice without an error, exactly like the reference.
func (f *HNFetcher) FetchTrends(ctx context.Context, excludeURLs map[string]bool) ([]db.Trend, TrendStats, error) {
	var stats TrendStats
	anyQueryOK := false

	for _, windowHours := range hnWindowsHours {
		stats.WindowHours = windowHours
		minTimestamp := f.now().Unix() - int64(windowHours)*3600

		hits, windowOK := f.fetchWindow(ctx, minTimestamp)
		anyQueryOK = anyQueryOK || windowOK
		if len(hits) == 0 {
			continue
		}
		stats.CandidatePool = len(hits)

		sort.SliceStable(hits, func(i, j int) bool { return hits[i].Points > hits[j].Points })

		unique := make([]hnHit, 0, len(hits))
		seenInBatch := map[string]bool{}
		deduped := 0
		for _, hit := range hits {
			if excludeURLs[hit.URL] || seenInBatch[hit.URL] {
				deduped++
				continue
			}
			seenInBatch[hit.URL] = true
			unique = append(unique, hit)
		}
		stats.DedupedOut = deduped

		if len(unique) >= hnMinFresh || windowHours == hnWindowsHours[len(hnWindowsHours)-1] {
			if len(unique) > hnMaxSelected {
				unique = unique[:hnMaxSelected]
			}
			today := f.now().UTC().Format("2006-01-02")
			createdAt := f.now().UnixMilli()
			src := "hackernews"
			trends := make([]db.Trend, 0, len(unique))
			for _, hit := range unique {
				sumEn, sumFi := f.summarizeStory(ctx, hit.Title, hit.URL)
				trends = append(trends, db.Trend{
					ID:        uuid.NewString(),
					Date:      today,
					Title:     hit.Title,
					Summary:   sumEn,
					SummaryFi: &sumFi,
					URL:       hit.URL,
					Source:    &src,
					CreatedAt: createdAt,
				})
			}
			return trends, stats, nil
		}
	}

	if !anyQueryOK {
		return nil, stats, fmt.Errorf("all HN queries failed")
	}
	return []db.Trend{}, stats, nil
}

// fetchWindow fetches every search query concurrently (Promise.allSettled
// semantics: individual query failures are tolerated) and returns the hits
// deduped by URL, plus whether at least one query succeeded.
func (f *HNFetcher) fetchWindow(ctx context.Context, minTimestamp int64) ([]hnHit, bool) {
	type result struct {
		hits []hnHit
		ok   bool
	}
	results := make([]result, len(hnSearchQueries))
	var wg sync.WaitGroup
	for i, q := range hnSearchQueries {
		wg.Add(1)
		go func(i int, q string) {
			defer wg.Done()
			hits, err := f.fetchQuery(ctx, q, minTimestamp)
			if err == nil {
				results[i] = result{hits: hits, ok: true}
			}
		}(i, q)
	}
	wg.Wait()

	byURL := map[string]hnHit{}
	anyOK := false
	for _, res := range results {
		if !res.ok {
			continue
		}
		anyOK = true
		for _, hit := range res.hits {
			if hit.Title == "" || hit.URL == "" {
				continue
			}
			byURL[hit.URL] = hit
		}
	}
	out := make([]hnHit, 0, len(byURL))
	for _, hit := range byURL {
		out = append(out, hit)
	}
	return out, anyOK
}

func (f *HNFetcher) fetchQuery(ctx context.Context, query string, minTimestamp int64) ([]hnHit, error) {
	// NOTE: numericFilters must be percent-encoded (created_at_i%3E...); the
	// reference interpolates a raw ">", which Algolia now rejects with 400.
	q := url.Values{
		"query":          {query},
		"tags":           {"story"},
		"numericFilters": {fmt.Sprintf("created_at_i>%d", minTimestamp)},
		"hitsPerPage":    {"30"},
	}
	u := f.baseURL() + "/api/v1/search_by_date?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var body hnSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body.Hits, nil
}

// summarizeStory generates the EN and FI summaries concurrently (mirroring the
// reference's Promise.all over two generateText calls). On any failure — or
// when no summarizer is configured — both fall back to the story title, the
// deterministic offline content.
func (f *HNFetcher) summarizeStory(ctx context.Context, title, storyURL string) (string, string) {
	if f.Summarizer == nil {
		return title, title
	}
	promptEn := fmt.Sprintf(`Summarize this tech news story in 1–2 sentences for a developer audience. Be concise and factual. Title: "%s". URL: %s`, title, storyURL)
	promptFi := fmt.Sprintf(`Tiivistä tämä teknologiauutinen 1–2 lauseeseen kehittäjäyleisölle. Ole ytimekäs ja asiallinen. Vastaa suomeksi. Otsikko: "%s". URL: %s`, title, storyURL)

	var wg sync.WaitGroup
	en, fi := title, title
	wg.Add(2)
	go func() {
		defer wg.Done()
		if text, err := f.Summarizer.Generate(ctx, promptEn, hnSummaryMaxTokens); err == nil {
			en = text
		}
	}()
	go func() {
		defer wg.Done()
		if text, err := f.Summarizer.Generate(ctx, promptFi, hnSummaryMaxTokens); err == nil {
			fi = text
		}
	}()
	wg.Wait()
	return en, fi
}

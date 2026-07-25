package aipulse

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"goth/internal/db"
)

func newRefreshTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

// okServers builds HN/GitHub/Yahoo test servers that all succeed with one
// row each.
func okServers(t *testing.T) (hn, gh, yh *httptest.Server) {
	hn = httptest.NewServer((&hnServer{serveFromTs: fixedNow().Unix() - 24*3600, hits: makeHits(5, "r")}).handler(t))
	gh = httptest.NewServer((&ghServer{pages: trendingPathsWith(repoHTML("a/llm-kit", "An LLM toolkit.", "Python", 1, 1))}).handler())
	yh = httptest.NewServer((&yahooServer{bodies: map[string]map[string]any{"NVDA": oneDayChart()}}).handler())
	t.Cleanup(func() { hn.Close(); gh.Close(); yh.Close() })
	return hn, gh, yh
}

func testRefresher(hnURL, ghURL, yhURL string) *Refresher {
	return &Refresher{
		Trends: &HNFetcher{BaseURL: hnURL, Now: fixedNow},
		Repos:  &GitHubFetcher{BaseURL: ghURL, Now: fixedNow},
		Stocks: &YahooFetcher{BaseURL: yhURL, Now: fixedNow, sleep: func(d time.Duration) {}},
	}
}

func TestRefresherRunPersistsAllSources(t *testing.T) {
	conn := newRefreshTestDB(t)
	hn, gh, yh := okServers(t)

	rep := testRefresher(hn.URL, gh.URL, yh.URL).Run(context.Background(), conn)
	if rep.AnyFailed() {
		t.Fatalf("expected clean run: %+v", rep)
	}
	if rep.Trends.Inserted != 5 || rep.Repos.Inserted != 1 || rep.Stocks.Inserted != 1 {
		t.Fatalf("unexpected counts: %+v", rep)
	}
	if rep.Trends.SkippedDup != 0 || rep.Trends.WindowHours != 24 {
		t.Fatalf("unexpected trend stats: %+v", rep.Trends)
	}

	// Rows actually persisted.
	trends, err := db.GetLatestTrends(conn, fixedNow().UTC().Format("2006-01-02"))
	if err != nil || len(trends) != 5 {
		t.Fatalf("trends persisted = %d, err=%v", len(trends), err)
	}
	repos, err := db.GetLatestRepos(conn, fixedNow().UTC().Format("2006-01-02"))
	if err != nil || len(repos) != 1 {
		t.Fatalf("repos persisted = %d, err=%v", len(repos), err)
	}
	stocks, err := db.GetStockHistory(conn, "NVDA", 365)
	if err != nil || len(stocks) != 1 {
		t.Fatalf("stocks persisted = %d, err=%v", len(stocks), err)
	}

	// Cache meta written per source.
	for _, src := range []string{"trends", "repos", "stocks"} {
		meta, err := db.GetCacheMeta(conn, src)
		if err != nil || meta == nil || meta.Status != "ok" {
			t.Fatalf("cache meta %s = %+v, err=%v", src, meta, err)
		}
	}
}

func TestRefresherRunIsIdempotent(t *testing.T) {
	conn := newRefreshTestDB(t)
	hn, gh, yh := okServers(t)
	r := testRefresher(hn.URL, gh.URL, yh.URL)

	r.Run(context.Background(), conn)
	r.Run(context.Background(), conn) // second run: same fetched rows

	count := func(table string) int {
		var n int
		if err := conn.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		return n
	}
	if got := count("ai_trends"); got != 5 {
		t.Fatalf("ai_trends rows = %d after 2 runs, want 5 (dedup)", got)
	}
	if got := count("ai_repos"); got != 1 {
		t.Fatalf("ai_repos rows = %d after 2 runs, want 1 (dedup)", got)
	}
	if got := count("ai_stocks"); got != 1 {
		t.Fatalf("ai_stocks rows = %d after 2 runs, want 1 (dedup)", got)
	}
}

func TestRefresherSourceFailureIsolated(t *testing.T) {
	conn := newRefreshTestDB(t)
	_, gh, yh := okServers(t)
	hnDown := httptest.NewServer((&hnServer{fail: true}).handler(t))
	defer hnDown.Close()

	rep := testRefresher(hnDown.URL, gh.URL, yh.URL).Run(context.Background(), conn)
	if rep.Trends.OK || rep.Trends.ErrKind != "fetch" {
		t.Fatalf("trends should be a fetch failure: %+v", rep.Trends)
	}
	if !rep.Repos.OK || !rep.Stocks.OK {
		t.Fatalf("repos/stocks must survive an HN outage: %+v", rep)
	}
	meta, _ := db.GetCacheMeta(conn, "trends")
	if meta == nil || meta.Status != "error" || meta.Detail != "fetch failed" {
		t.Fatalf("trends cache meta = %+v", meta)
	}
	// Repos cache meta is ok — failures stay isolated.
	meta, _ = db.GetCacheMeta(conn, "repos")
	if meta == nil || meta.Status != "ok" {
		t.Fatalf("repos cache meta = %+v", meta)
	}
}

func TestRefresherDBFailureMarkedPerSource(t *testing.T) {
	// A DB without migrations: every upsert fails while fetches succeed.
	raw, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "raw.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer raw.Close()

	hn, gh, yh := okServers(t)
	rep := testRefresher(hn.URL, gh.URL, yh.URL).Run(context.Background(), raw)
	if rep.Trends.ErrKind != "db" || rep.Repos.ErrKind != "db" || rep.Stocks.ErrKind != "db" {
		t.Fatalf("expected db failures: %+v", rep)
	}
	if rep.Trends.Inserted != 0 || rep.Repos.Inserted != 0 || rep.Stocks.Inserted != 0 {
		t.Fatalf("inserted must be 0 on db failure: %+v", rep)
	}
}

func TestRefresherTryRunOverlap(t *testing.T) {
	conn := newRefreshTestDB(t)
	// Block the stocks fetch until released.
	release := make(chan struct{})
	started := make(chan struct{})
	var once sync.Once
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(started) })
		<-release
		json.NewEncoder(w).Encode(oneDayChart())
	}))
	defer slow.Close()
	hn, gh, _ := okServers(t)

	r := testRefresher(hn.URL, gh.URL, slow.URL)
	done := make(chan RefreshReport, 1)
	go func() {
		rep, _ := r.TryRun(context.Background(), conn)
		done <- rep
	}()

	// Wait until the first run is genuinely in-flight, then verify a second
	// run is rejected.
	<-started
	_, started2 := r.TryRun(context.Background(), conn)
	if started2 {
		t.Fatal("overlapping run must not start")
	}
	close(release)
	<-done
}

func TestRefresherRecentURLDedupFeedsFetcher(t *testing.T) {
	conn := newRefreshTestDB(t)
	// Pre-seed a trend whose URL the fetcher would otherwise re-summarize.
	seededURL := "https://example.com/r-0"
	if err := db.UpsertTrends(conn, []db.Trend{{
		ID: "seed", Date: fixedNow().UTC().Format("2006-01-02"), Title: "seed",
		Summary: "s", URL: seededURL, CreatedAt: 1,
	}}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	hn, gh, yh := okServers(t)
	rep := testRefresher(hn.URL, gh.URL, yh.URL).Run(context.Background(), conn)
	if !rep.Trends.OK {
		t.Fatalf("run: %+v", rep)
	}
	if rep.Trends.SkippedDup != 1 {
		t.Fatalf("skippedDup = %d, want 1 (seeded URL)", rep.Trends.SkippedDup)
	}
	if rep.Trends.Inserted != 4 { // 5 fetched − 1 excluded
		t.Fatalf("inserted = %d, want 4", rep.Trends.Inserted)
	}
	var n int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM ai_trends WHERE url = ?`, seededURL).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("seeded URL re-inserted %d times", n)
	}
}

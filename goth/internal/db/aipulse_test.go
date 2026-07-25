package db

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

func ptr(s string) *string { return &s }

func TestAIPulseEmpty(t *testing.T) {
	conn := openTestDB(t)

	if rows, err := GetLatestTrends(conn, ""); err != nil || len(rows) != 0 {
		t.Errorf("GetLatestTrends = (%d, %v), want (0, nil)", len(rows), err)
	}
	if rows, err := GetLatestRepos(conn, ""); err != nil || len(rows) != 0 {
		t.Errorf("GetLatestRepos = (%d, %v), want (0, nil)", len(rows), err)
	}
	if rows, err := GetStockHistory(conn, "NVDA", 365); err != nil || len(rows) != 0 {
		t.Errorf("GetStockHistory = (%d, %v), want (0, nil)", len(rows), err)
	}
	if tickers, err := GetAvailableTickers(conn); err != nil || len(tickers) != 0 {
		t.Errorf("GetAvailableTickers = (%d, %v), want (0, nil)", len(tickers), err)
	}
	if meta, err := GetCacheMeta(conn, "trends"); err != nil || meta != nil {
		t.Errorf("GetCacheMeta = (%v, %v), want (nil, nil)", meta, err)
	}
}

func TestAIPulseTrendUpsertAndGet(t *testing.T) {
	conn := openTestDB(t)

	trends := []Trend{
		{ID: "t1", Date: "2026-07-24", Title: "Alpha", Summary: "s1", URL: "https://a", Source: ptr("hackernews"), CreatedAt: 1},
		{ID: "t2", Date: "2026-07-24", Title: "Beta", Summary: "s2", SummaryFi: ptr("fi2"), URL: "https://b", CreatedAt: 2},
	}
	if err := UpsertTrends(conn, trends); err != nil {
		t.Fatalf("UpsertTrends: %v", err)
	}

	got, err := GetLatestTrends(conn, "2026-07-24")
	if err != nil {
		t.Fatalf("GetLatestTrends: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d trends, want 2", len(got))
	}
	// Ordered by created_at ascending.
	if got[0].ID != "t1" || got[1].ID != "t2" {
		t.Errorf("order = %s,%s want t1,t2", got[0].ID, got[1].ID)
	}
	if got[0].Source == nil || *got[0].Source != "hackernews" {
		t.Errorf("t1 source = %v, want hackernews", got[0].Source)
	}
	if got[1].SummaryFi == nil || *got[1].SummaryFi != "fi2" {
		t.Errorf("t2 summaryFi = %v, want fi2", got[1].SummaryFi)
	}
}

func TestAIPulseTrendFallbackToLatestDay(t *testing.T) {
	conn := openTestDB(t)

	// Only an old day has rows; today is empty.
	old := []Trend{{ID: "t1", Date: "2026-01-01", Title: "Old", Summary: "s", URL: "https://o", CreatedAt: 1}}
	if err := UpsertTrends(conn, old); err != nil {
		t.Fatalf("UpsertTrends: %v", err)
	}

	got, err := GetLatestTrends(conn, "")
	if err != nil {
		t.Fatalf("GetLatestTrends: %v", err)
	}
	if len(got) != 1 || got[0].Date != "2026-01-01" {
		t.Errorf("fallback returned %+v, want the 2026-01-01 row", got)
	}
}

func TestAIPulseTrendDedupIdempotent(t *testing.T) {
	conn := openTestDB(t)

	trend := Trend{ID: "t1", Date: "2026-07-24", Title: "Same", Summary: "s", URL: "https://u", CreatedAt: 1}
	if err := UpsertTrends(conn, []Trend{trend}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := UpsertTrends(conn, []Trend{trend}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	got, err := GetLatestTrends(conn, "2026-07-24")
	if err != nil {
		t.Fatalf("GetLatestTrends: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("rows = %d after re-ingest, want 1 (idempotent)", len(got))
	}
}

// Phase 12.5c: the (ticker, date) index is UNIQUE (migration 0003), so stock
// refreshes are idempotent like the Next.js onConflictDoNothing.
func TestAIPulseStockDedupIdempotent(t *testing.T) {
	conn := openTestDB(t)

	stock := Stock{ID: "s1", Date: "2026-07-24", Ticker: "NVDA", CompanyName: "NVIDIA", Open: 1, High: 2, Low: 1, Close: 2, CreatedAt: 1}
	if err := UpsertStocks(conn, []Stock{stock}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	// Same (ticker, date), new row id — must be ignored, not duplicated.
	dup := stock
	dup.ID = "s2"
	if err := UpsertStocks(conn, []Stock{dup}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	got, err := GetStockHistory(conn, "NVDA", 365)
	if err != nil {
		t.Fatalf("GetStockHistory: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("rows = %d after re-fetch, want 1 (idempotent)", len(got))
	}
}

// Phase 12.5c: migration 0003 dedupes pre-existing duplicate (ticker, date)
// rows before enforcing uniqueness (the 12b plain index allowed them).
func TestMigration0003DedupesExistingRows(t *testing.T) {
	// Build a DB with ONLY the first two migrations applied, inserting
	// duplicate stock rows while the plain index is in place.
	path := filepath.Join(t.TempDir(), "old.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	for _, m := range []string{"0001_init.sql", "0002_ai_pulse.sql"} {
		data, err := migrationFS.ReadFile("migrations/" + m)
		if err != nil {
			t.Fatalf("read %s: %v", m, err)
		}
		if _, err := raw.Exec(string(data)); err != nil {
			t.Fatalf("apply %s: %v", m, err)
		}
	}
	for _, id := range []string{"a", "b", "c"} {
		if _, err := raw.Exec(`INSERT INTO ai_stocks
			(id, date, ticker, company_name, open, high, low, close, volume, created_at)
			VALUES (?, '2026-07-24', 'NVDA', 'NVIDIA', 1, 2, 1, 2, NULL, 1)`, id); err != nil {
			t.Fatalf("seed dup: %v", err)
		}
	}
	raw.Close()

	// Open runs all migrations incl. 0003: dups must be collapsed to one row.
	conn, err := Open(path)
	if err != nil {
		t.Fatalf("Open with 0003: %v", err)
	}
	defer conn.Close()
	var n int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM ai_stocks WHERE ticker = 'NVDA' AND date = '2026-07-24'`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Errorf("rows after 0003 = %d, want 1 (deduped)", n)
	}
	// And the unique index now rejects new duplicates.
	if err := UpsertStocks(conn, []Stock{
		{ID: "x", Date: "2026-07-24", Ticker: "NVDA", CompanyName: "NVIDIA", Open: 1, High: 2, Low: 1, Close: 2, CreatedAt: 2},
	}); err != nil {
		t.Fatalf("idempotent re-upsert: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM ai_stocks`).Scan(&n); err != nil {
		t.Fatalf("recount: %v", err)
	}
	if n != 1 {
		t.Errorf("rows after re-upsert = %d, want 1", n)
	}
}

func TestAIPulseReposAndStocks(t *testing.T) {
	conn := openTestDB(t)

	repos := []Repo{
		{ID: "r1", Date: "2026-07-24", RepoFullName: "o/n", URL: "https://gh", Stars: 100, StarsToday: 12, Source: "github-trending", CreatedAt: 1},
	}
	if err := UpsertRepos(conn, repos); err != nil {
		t.Fatalf("UpsertRepos: %v", err)
	}
	gotRepos, err := GetLatestRepos(conn, "2026-07-24")
	if err != nil {
		t.Fatalf("GetLatestRepos: %v", err)
	}
	if len(gotRepos) != 1 || gotRepos[0].StarsToday != 12 {
		t.Errorf("repos = %+v, want 1 repo with starsToday 12", gotRepos)
	}

	vol := int64(5_000_000)
	stocks := []Stock{
		{ID: "s1", Date: "2026-07-20", Ticker: "NVDA", CompanyName: "NVIDIA", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: &vol, CreatedAt: 1},
		{ID: "s2", Date: "2026-07-21", Ticker: "NVDA", CompanyName: "NVIDIA", Open: 1.5, High: 2.5, Low: 1, Close: 2, CreatedAt: 2},
	}
	if err := UpsertStocks(conn, stocks); err != nil {
		t.Fatalf("UpsertStocks: %v", err)
	}
	gotStocks, err := GetStockHistory(conn, "NVDA", 365)
	if err != nil {
		t.Fatalf("GetStockHistory: %v", err)
	}
	if len(gotStocks) != 2 {
		t.Fatalf("stock history = %d, want 2", len(gotStocks))
	}
	if gotStocks[0].Date != "2026-07-20" || gotStocks[1].Date != "2026-07-21" {
		t.Errorf("stock order = %s,%s want ascending by date", gotStocks[0].Date, gotStocks[1].Date)
	}
	if gotStocks[0].Volume == nil || *gotStocks[0].Volume != vol {
		t.Errorf("s1 volume = %v, want %d", gotStocks[0].Volume, vol)
	}
	if gotStocks[1].Volume != nil {
		t.Errorf("s2 volume = %v, want nil", gotStocks[1].Volume)
	}

	tickers, err := GetAvailableTickers(conn)
	if err != nil {
		t.Fatalf("GetAvailableTickers: %v", err)
	}
	if len(tickers) != 1 || tickers[0] != "NVDA" {
		t.Errorf("tickers = %v, want [NVDA]", tickers)
	}
}

func TestAIPulseStockHistoryCutoff(t *testing.T) {
	conn := openTestDB(t)

	old := time.Now().UTC().AddDate(0, 0, -400).Format("2006-01-02")
	recent := time.Now().UTC().AddDate(0, 0, -10).Format("2006-01-02")
	if err := UpsertStocks(conn, []Stock{
		{ID: "old", Date: old, Ticker: "AAPL", CompanyName: "Apple", Open: 1, High: 1, Low: 1, Close: 1, CreatedAt: 1},
		{ID: "new", Date: recent, Ticker: "AAPL", CompanyName: "Apple", Open: 2, High: 2, Low: 2, Close: 2, CreatedAt: 2},
	}); err != nil {
		t.Fatalf("UpsertStocks: %v", err)
	}

	got, err := GetStockHistory(conn, "AAPL", 365)
	if err != nil {
		t.Fatalf("GetStockHistory: %v", err)
	}
	if len(got) != 1 || got[0].Date != recent {
		t.Errorf("history within 365d = %+v, want only the recent row", got)
	}
}

func TestAIPulseCacheMeta(t *testing.T) {
	conn := openTestDB(t)

	if err := UpsertCacheMeta(conn, "trends", 123, "ok", 7, ""); err != nil {
		t.Fatalf("UpsertCacheMeta: %v", err)
	}
	meta, err := GetCacheMeta(conn, "trends")
	if err != nil {
		t.Fatalf("GetCacheMeta: %v", err)
	}
	if meta == nil || meta.Status != "ok" || meta.Rows != 7 || meta.RanAt != 123 {
		t.Errorf("meta = %+v, want status ok rows 7 ranAt 123", meta)
	}

	// Re-ingest records over the same source row.
	if err := UpsertCacheMeta(conn, "trends", 456, "error", 0, "boom"); err != nil {
		t.Fatalf("UpsertCacheMeta (update): %v", err)
	}
	meta, _ = GetCacheMeta(conn, "trends")
	if meta.Status != "error" || meta.Detail != "boom" || meta.RanAt != 456 {
		t.Errorf("updated meta = %+v, want status error detail boom ranAt 456", meta)
	}
}

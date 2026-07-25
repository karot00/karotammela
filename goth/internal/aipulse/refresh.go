package aipulse

import (
	"context"
	"database/sql"
	"log"
	"sync/atomic"
	"time"

	"goth/internal/db"
)

// RefreshSourceResult is one source's outcome in a refresh run. ErrDetail
// carries the raw error for ?debug=1 responses; public responses and logs use
// only the redacted ErrKind ("fetch"/"db").
type RefreshSourceResult struct {
	OK          bool
	Inserted    int
	SkippedDup  int    // trends only
	WindowHours int    // trends only
	ErrKind     string // "", "fetch", or "db" — safe for logs/responses
	ErrDetail   string // raw error; debug responses only, never logged
}

// RefreshReport summarizes one full run across all sources.
type RefreshReport struct {
	RanAt  time.Time
	Trends RefreshSourceResult
	Repos  RefreshSourceResult
	Stocks RefreshSourceResult
}

// AnyFailed reports whether at least one source failed.
func (r RefreshReport) AnyFailed() bool {
	return !r.Trends.OK || !r.Repos.OK || !r.Stocks.OK
}

// Refresher orchestrates one independent AI Pulse refresh run (Phase 12.5d),
// the Go port of src/app/api/ai-pulse/refresh/route.ts: recent-URL dedup,
// all-settled source fan-out, independent persistence per source, per-source
// cache-meta, and a single structured redacted log line. Runs are idempotent
// (upserts are ON CONFLICT DO NOTHING) and source failures stay isolated.
type Refresher struct {
	Trends *HNFetcher
	Repos  *GitHubFetcher
	Stocks *YahooFetcher

	running atomic.Bool
}

// TryRun executes a run unless one is already in progress in this process.
// It returns (report, true) on execution or (zero, false) when overlapping.
// Inter-process overlap is prevented by flock in the systemd unit (12.5e).
func (r *Refresher) TryRun(ctx context.Context, conn *sql.DB) (RefreshReport, bool) {
	if !r.running.CompareAndSwap(false, true) {
		return RefreshReport{}, false
	}
	defer r.running.Store(false)
	return r.Run(ctx, conn), true
}

// Run executes the refresh. Sources run concurrently; each failure mode
// (fetch vs persist) is recorded per source and never aborts the others.
func (r *Refresher) Run(ctx context.Context, conn *sql.DB) RefreshReport {
	report := RefreshReport{RanAt: time.Now().UTC()}

	// Recent-URL dedup input; a read failure degrades to an empty set, exactly
	// like the reference route.
	excludeURLs := map[string]bool{}
	if urls, err := db.GetRecentTrendUrls(conn, 7); err == nil {
		excludeURLs = urls
	} else {
		log.Printf("ai-pulse.refresh.recent_urls_failed")
	}

	type trendsOut struct {
		rows  []db.Trend
		stats TrendStats
		err   error
	}
	type reposOut struct {
		rows []db.Repo
		err  error
	}
	type stocksOut struct {
		rows []db.Stock
		err  error
	}
	trendsCh := make(chan trendsOut, 1)
	reposCh := make(chan reposOut, 1)
	stocksCh := make(chan stocksOut, 1)

	go func() {
		rows, stats, err := r.Trends.FetchTrends(ctx, excludeURLs)
		trendsCh <- trendsOut{rows, stats, err}
	}()
	go func() {
		rows, err := r.Repos.FetchRepos(ctx)
		reposCh <- reposOut{rows, err}
	}()
	go func() {
		rows, err := r.Stocks.FetchStocks(ctx)
		stocksCh <- stocksOut{rows, err}
	}()

	tRes := <-trendsCh
	if tRes.err != nil {
		report.Trends = RefreshSourceResult{ErrKind: "fetch", ErrDetail: tRes.err.Error(), WindowHours: tRes.stats.WindowHours}
	} else if err := db.UpsertTrends(conn, tRes.rows); err != nil {
		// inserted stays 0 on failure (nothing was persisted); the fetch
		// stats remain as diagnostic detail.
		report.Trends = RefreshSourceResult{ErrKind: "db", ErrDetail: err.Error(), SkippedDup: tRes.stats.DedupedOut, WindowHours: tRes.stats.WindowHours}
	} else {
		report.Trends = RefreshSourceResult{OK: true, Inserted: len(tRes.rows), SkippedDup: tRes.stats.DedupedOut, WindowHours: tRes.stats.WindowHours}
	}
	rRes := <-reposCh
	if rRes.err != nil {
		report.Repos = RefreshSourceResult{ErrKind: "fetch", ErrDetail: rRes.err.Error()}
	} else if err := db.UpsertRepos(conn, rRes.rows); err != nil {
		report.Repos = RefreshSourceResult{ErrKind: "db", ErrDetail: err.Error()}
	} else {
		report.Repos = RefreshSourceResult{OK: true, Inserted: len(rRes.rows)}
	}
	sRes := <-stocksCh
	if sRes.err != nil {
		report.Stocks = RefreshSourceResult{ErrKind: "fetch", ErrDetail: sRes.err.Error()}
	} else if err := db.UpsertStocks(conn, sRes.rows); err != nil {
		report.Stocks = RefreshSourceResult{ErrKind: "db", ErrDetail: err.Error()}
	} else {
		report.Stocks = RefreshSourceResult{OK: true, Inserted: len(sRes.rows)}
	}

	recordCacheMeta(conn, "trends", report.Trends)
	recordCacheMeta(conn, "repos", report.Repos)
	recordCacheMeta(conn, "stocks", report.Stocks)

	// One structured redacted line per run (counts/kinds only — never error
	// detail, which may embed query or provider content).
	log.Printf("ai-pulse.refresh.complete trends_ok=%v trends_inserted=%d trends_skipped=%d trends_window=%dh repos_ok=%v repos_inserted=%d stocks_ok=%v stocks_inserted=%d failed_kinds=%s",
		report.Trends.OK, report.Trends.Inserted, report.Trends.SkippedDup, report.Trends.WindowHours,
		report.Repos.OK, report.Repos.Inserted,
		report.Stocks.OK, report.Stocks.Inserted,
		failedKinds(report),
	)
	return report
}

// recordCacheMeta writes the per-source ai_cache_meta row with a redacted
// detail string (the raw error stays in the report for debug responses only).
func recordCacheMeta(conn *sql.DB, source string, res RefreshSourceResult) {
	status, detail := "ok", ""
	if !res.OK {
		status = "error"
		switch res.ErrKind {
		case "fetch":
			detail = "fetch failed"
		case "db":
			detail = "db insert failed"
		default:
			detail = "failed"
		}
	}
	_ = db.UpsertCacheMeta(conn, source, time.Now().UnixMilli(), status, res.Inserted, detail)
}

func failedKinds(r RefreshReport) string {
	out := ""
	for _, s := range []struct {
		name string
		res  RefreshSourceResult
	}{{"trends", r.Trends}, {"repos", r.Repos}, {"stocks", r.Stocks}} {
		if !s.res.OK {
			if out != "" {
				out += ","
			}
			out += s.name + ":" + s.res.ErrKind
		}
	}
	if out == "" {
		return "none"
	}
	return out
}

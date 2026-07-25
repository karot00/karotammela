package aipulse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// chartBody builds a minimal v8 chart payload.
func chartBody(tss []int64, close []*float64, open []*float64, vol []*int64) map[string]any {
	quote := map[string]any{
		"open":   open,
		"high":   close, // reuse to keep the fixture small
		"low":    close,
		"close":  close,
		"volume": vol,
	}
	return map[string]any{
		"chart": map[string]any{
			"result": []any{
				map[string]any{"timestamp": tss, "indicators": map[string]any{"quote": []any{quote}}},
			},
			"error": nil,
		},
	}
}

func f64(v float64) *float64 { return &v }
func i64(v int64) *int64     { return &v }

func TestYahooParseNormalization(t *testing.T) {
	// Day 1: full row. Day 2: null close → skipped. Day 3: null open →
	// defaults to close; null volume stays null.
	tss := []int64{1_800_000_000, 1_800_086_400, 1_800_172_800}
	close := []*float64{f64(100.5), nil, f64(99.25)}
	open := []*float64{f64(100.0), f64(101.0), nil}
	vol := []*int64{i64(1234), i64(0), nil}
	body, _ := json.Marshal(chartBody(tss, close, open, vol))

	rows, err := parseYahooChart(body, TickerOption{Ticker: "NVDA", Name: "NVIDIA"}, fixedNow())
	if err != nil {
		t.Fatalf("parseYahooChart: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (null close skipped), got %d", len(rows))
	}
	r1 := rows[0]
	if r1.Ticker != "NVDA" || r1.CompanyName != "NVIDIA" || r1.Close != 100.5 || r1.Open != 100.0 {
		t.Fatalf("row1 mismatch: %+v", r1)
	}
	if r1.Date != time.Unix(1_800_000_000, 0).UTC().Format("2006-01-02") {
		t.Fatalf("row1 date = %q", r1.Date)
	}
	if r1.Volume == nil || *r1.Volume != 1234 {
		t.Fatalf("row1 volume = %+v", r1.Volume)
	}
	r3 := rows[1]
	if r3.Open != 99.25 || r3.High != 99.25 || r3.Low != 99.25 {
		t.Fatalf("row3 OHLC must default to close: %+v", r3)
	}
	if r3.Volume != nil {
		t.Fatalf("row3 volume must stay null: %+v", r3.Volume)
	}
	if r3.ID == "" || r3.CreatedAt == 0 {
		t.Fatalf("row3 metadata missing: %+v", r3)
	}
}

func TestYahooParseNoResult(t *testing.T) {
	if _, err := parseYahooChart([]byte(`{"chart":{"result":null,"error":null}}`), TickerOption{}, fixedNow()); err == nil {
		t.Fatal("expected error on empty result")
	}
	if rows, err := parseYahooChart([]byte(`not json`), TickerOption{}, fixedNow()); err == nil || rows != nil {
		t.Fatalf("expected error on malformed JSON, got %v %v", rows, err)
	}
}

// yahooServer routes per-ticker responses: body, or a script of statuses.
type yahooServer struct {
	mu      sync.Mutex
	bodies  map[string]map[string]any
	scripts map[string][]int // per-ticker status sequence before 200
	calls   map[string]int
	retry   string // Retry-After value to send on 429s
}

func (s *yahooServer) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.calls == nil {
			s.calls = map[string]int{}
		}
		parts := strings.Split(r.URL.Path, "/")
		ticker := parts[len(parts)-1]
		s.calls[ticker]++
		if script := s.scripts[ticker]; len(script) > 0 {
			status := script[0]
			s.scripts[ticker] = script[1:]
			if s.retry != "" && status == 429 {
				w.Header().Set("Retry-After", s.retry)
			}
			w.WriteHeader(status)
			return
		}
		body, ok := s.bodies[ticker]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(body)
	})
}

func oneDayChart() map[string]any {
	return chartBody([]int64{1_800_000_000}, []*float64{f64(42.0)}, []*float64{f64(41.0)}, []*int64{i64(7)})
}

func TestYahooFetcherAllTickers(t *testing.T) {
	srv := &yahooServer{bodies: map[string]map[string]any{}}
	for _, tk := range AITickers {
		srv.bodies[tk.Ticker] = oneDayChart()
	}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &YahooFetcher{BaseURL: ts.URL, Now: fixedNow, sleep: func(time.Duration) {}}
	rows, err := f.FetchStocks(context.Background())
	if err != nil {
		t.Fatalf("FetchStocks: %v", err)
	}
	if len(rows) != len(AITickers) {
		t.Fatalf("expected %d rows, got %d", len(AITickers), len(rows))
	}
	seen := map[string]bool{}
	for _, r := range rows {
		seen[r.Ticker] = true
		if r.Close != 42.0 {
			t.Fatalf("close mismatch for %s: %+v", r.Ticker, r)
		}
	}
	for _, tk := range AITickers {
		if !seen[tk.Ticker] {
			t.Fatalf("ticker %s missing", tk.Ticker)
		}
	}
}

func TestYahooFetcherRetriesOn429HonoringRetryAfter(t *testing.T) {
	srv := &yahooServer{
		bodies:  map[string]map[string]any{"NVDA": oneDayChart()},
		scripts: map[string][]int{"NVDA": {429, 429}},
		retry:   "30",
	}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	var waits []time.Duration
	f := &YahooFetcher{
		BaseURL: ts.URL,
		Now:     fixedNow,
		sleep:   func(d time.Duration) { waits = append(waits, d) },
	}
	// Restrict tickers via a single-ticker fetch.
	rows, err := f.fetchTickerHistory(context.Background(), AITickers[0])
	if err != nil {
		t.Fatalf("fetchTickerHistory: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if srv.calls["NVDA"] != 3 {
		t.Fatalf("calls = %d, want 3 (2 retries)", srv.calls["NVDA"])
	}
	// Retry-After: 30s must be honored but capped at 10s; plus the attempt backoff.
	var capped bool
	for _, w := range waits {
		if w == yahooRetryAfterCap {
			capped = true
		}
	}
	if !capped {
		t.Fatalf("Retry-After 30s not capped at %v: waits=%v", yahooRetryAfterCap, waits)
	}
}

func TestYahooFetcherRetriesOn500ThenSuccess(t *testing.T) {
	srv := &yahooServer{
		bodies:  map[string]map[string]any{"NVDA": oneDayChart()},
		scripts: map[string][]int{"NVDA": {500}},
	}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &YahooFetcher{BaseURL: ts.URL, Now: fixedNow, sleep: func(time.Duration) {}}
	rows, err := f.fetchTickerHistory(context.Background(), AITickers[0])
	if err != nil {
		t.Fatalf("fetchTickerHistory: %v", err)
	}
	if len(rows) != 1 || srv.calls["NVDA"] != 2 {
		t.Fatalf("rows=%d calls=%d", len(rows), srv.calls["NVDA"])
	}
}

func TestYahooFetcherNoRetryOn404(t *testing.T) {
	srv := &yahooServer{bodies: map[string]map[string]any{}}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &YahooFetcher{BaseURL: ts.URL, Now: fixedNow, sleep: func(time.Duration) {}}
	if _, err := f.fetchTickerHistory(context.Background(), AITickers[0]); err == nil {
		t.Fatal("expected error on persistent 404")
	}
	if srv.calls["NVDA"] != 1 {
		t.Fatalf("404 must not be retried: calls=%d", srv.calls["NVDA"])
	}
}

func TestYahooFetcherPartialFailureTolerated(t *testing.T) {
	srv := &yahooServer{bodies: map[string]map[string]any{}, scripts: map[string][]int{}}
	for _, tk := range AITickers {
		if tk.Ticker == "NVDA" || tk.Ticker == "MSFT" {
			srv.scripts[tk.Ticker] = []int{500, 500, 500} // exhaust retries
			continue
		}
		srv.bodies[tk.Ticker] = oneDayChart()
	}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &YahooFetcher{BaseURL: ts.URL, Now: fixedNow, sleep: func(time.Duration) {}}
	rows, err := f.FetchStocks(context.Background())
	if err != nil {
		t.Fatalf("partial failure must not error: %v", err)
	}
	if len(rows) != len(AITickers)-2 {
		t.Fatalf("expected %d rows, got %d", len(AITickers)-2, len(rows))
	}
}

func TestYahooFetcherAllFailErrors(t *testing.T) {
	srv := &yahooServer{bodies: map[string]map[string]any{}, scripts: map[string][]int{}}
	for _, tk := range AITickers {
		srv.scripts[tk.Ticker] = []int{500, 500, 500}
	}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	f := &YahooFetcher{BaseURL: ts.URL, Now: fixedNow, sleep: func(time.Duration) {}}
	if _, err := f.FetchStocks(context.Background()); err == nil {
		t.Fatal("expected error when every ticker fails")
	}
}

func TestParseRetryAfter(t *testing.T) {
	if got := parseRetryAfter("5"); got != 5*time.Second {
		t.Fatalf("parseRetryAfter(5) = %v", got)
	}
	if got := parseRetryAfter(""); got != 0 {
		t.Fatalf("empty = %v", got)
	}
	if got := parseRetryAfter("Mon, 02 Jan 2006 15:04:05 GMT"); got != 0 {
		t.Fatalf("http-date unsupported = %v", got)
	}
	if got := parseRetryAfter("-3"); got != 0 {
		t.Fatalf("negative = %v", got)
	}
}

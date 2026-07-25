package aipulse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"goth/internal/db"
)

const (
	yahooDefaultBaseURL = "https://query1.finance.yahoo.com"
	// yahooMaxAttempts bounds per-ticker retries (initial try + 2 retries).
	yahooMaxAttempts = 3
	// yahooRetryAfterCap bounds how long a Retry-After header can pause the
	// run, keeping total runtime bounded for the systemd unit.
	yahooRetryAfterCap = 10 * time.Second
)

// yahooChartResponse mirrors the v8 chart payload consumed by
// src/lib/ai/stocks-fetcher.ts. OHLC entries are nullable; volume is nullable.
type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []*float64 `json:"open"`
					High   []*float64 `json:"high"`
					Low    []*float64 `json:"low"`
					Close  []*float64 `json:"close"`
					Volume []*int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error any `json:"error"`
	} `json:"chart"`
}

// YahooFetcher is the Go Yahoo Finance stocks writer (Phase 12.5c), a port of
// src/lib/ai/stocks-fetcher.ts with bounded retries honoring Retry-After.
// Per-ticker failures degrade to an empty series for that ticker — cached
// last-known data is never erased because persistence is insert-only.
type YahooFetcher struct {
	// BaseURL overrides the Yahoo host for tests/drills.
	BaseURL string
	// Client is the HTTP client; nil uses a 20s-timeout default (bounded).
	Client *http.Client
	// Now supplies the current time; nil uses time.Now.
	Now func() time.Time
	// sleep is the backoff wait hook; nil uses time.Sleep (tests inject a
	// no-op and count calls).
	sleep func(time.Duration)
}

func (f *YahooFetcher) httpClient() *http.Client {
	if f.Client != nil {
		return f.Client
	}
	return &http.Client{Timeout: 20 * time.Second}
}

func (f *YahooFetcher) baseURL() string {
	if f.BaseURL != "" {
		return f.BaseURL
	}
	return yahooDefaultBaseURL
}

func (f *YahooFetcher) now() time.Time {
	if f.Now != nil {
		return f.Now()
	}
	return time.Now()
}

func (f *YahooFetcher) wait(d time.Duration) {
	if f.sleep != nil {
		f.sleep(d)
		return
	}
	time.Sleep(d)
}

// FetchStocks fetches one year of daily OHLCV per supported ticker and
// returns every row. An error is returned only when all tickers failed, so a
// provider outage marks the source failed in the refresh report while partial
// data still persists.
func (f *YahooFetcher) FetchStocks(ctx context.Context) ([]db.Stock, error) {
	var out []db.Stock
	failures := 0
	for _, entry := range AITickers {
		rows, err := f.fetchTickerHistory(ctx, entry)
		if err != nil {
			failures++
			continue
		}
		out = append(out, rows...)
	}
	if failures == len(AITickers) {
		return nil, fmt.Errorf("all %d ticker fetches failed", failures)
	}
	return out, nil
}

// fetchTickerHistory fetches and parses one ticker's 1y daily series with
// bounded retries. Retryable conditions: network errors, 429, and 5xx. A
// Retry-After header on 429 is honored up to yahooRetryAfterCap.
func (f *YahooFetcher) fetchTickerHistory(ctx context.Context, entry TickerOption) ([]db.Stock, error) {
	u := fmt.Sprintf("%s/v8/finance/chart/%s?range=1y&interval=1d&includePrePost=false", f.baseURL(), entry.Ticker)

	var lastErr error
	for attempt := 0; attempt < yahooMaxAttempts; attempt++ {
		if attempt > 0 {
			f.wait(time.Duration(attempt) * 500 * time.Millisecond)
		}
		body, status, retryAfter, err := f.doGet(ctx, u)
		if err != nil {
			lastErr = err
			continue // network error: retryable
		}
		if status == http.StatusOK {
			return parseYahooChart(body, entry, f.now())
		}
		lastErr = fmt.Errorf("status %d", status)
		if status == http.StatusTooManyRequests || status >= 500 {
			if retryAfter > 0 {
				f.wait(min(retryAfter, yahooRetryAfterCap))
			}
			continue
		}
		// Other 4xx: not retryable.
		return nil, lastErr
	}
	return nil, lastErr
}

// doGet performs one GET and returns the body (nil on non-200), the status,
// and any Retry-After wait. The response body is discarded on non-200.
func (f *YahooFetcher) doGet(ctx context.Context, u string) ([]byte, int, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := f.httpClient().Do(req)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()
	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, retryAfter, nil
	}
	var body []byte
	body, err = io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, resp.StatusCode, 0, err
	}
	return body, resp.StatusCode, 0, nil
}

// parseRetryAfter interprets a Retry-After header in seconds (HTTP-date form
// is intentionally unsupported; Yahoo sends seconds).
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	secs, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || secs < 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

// parseYahooChart validates and normalizes the chart payload into rows:
// points with a null/NaN close are skipped; open/high/low default to close
// (reference behavior); non-finite values are dropped.
func parseYahooChart(body []byte, entry TickerOption, now time.Time) ([]db.Stock, error) {
	var data yahooChartResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("no result in response")
	}
	result := data.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 || len(result.Timestamp) == 0 {
		return []db.Stock{}, nil
	}
	quote := result.Indicators.Quote[0]
	n := len(result.Timestamp)
	at := func(arr []*float64, i int) *float64 {
		if i < len(arr) {
			return arr[i]
		}
		return nil
	}
	volAt := func(arr []*int64, i int) *int64 {
		if i < len(arr) {
			return arr[i]
		}
		return nil
	}

	createdAt := now.UnixMilli()
	rows := make([]db.Stock, 0, n)
	for i := 0; i < n; i++ {
		closep := at(quote.Close, i)
		if closep == nil || math.IsNaN(*closep) || math.IsInf(*closep, 0) {
			continue
		}
		close := *closep
		num := func(p *float64) float64 {
			if p == nil || math.IsNaN(*p) || math.IsInf(*p, 0) {
				return close
			}
			return *p
		}
		date := time.Unix(result.Timestamp[i], 0).UTC().Format("2006-01-02")
		rows = append(rows, db.Stock{
			ID:          uuid.NewString(),
			Date:        date,
			Ticker:      entry.Ticker,
			CompanyName: entry.Name,
			Open:        num(at(quote.Open, i)),
			High:        num(at(quote.High, i)),
			Low:         num(at(quote.Low, i)),
			Close:       close,
			Volume:      volAt(quote.Volume, i),
			CreatedAt:   createdAt,
		})
	}
	return rows, nil
}

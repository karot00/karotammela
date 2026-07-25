package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"goth/internal/aipulse"
	"goth/internal/config"
	"goth/internal/db"
	"goth/internal/security"
	"goth/internal/view"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestAiPulseStocksValidTicker(t *testing.T) {
	conn := newTestDB(t)

	vol := int64(5_000_000)
	if err := db.UpsertStocks(conn, []db.Stock{
		{ID: "s1", Date: "2026-07-20", Ticker: "NVDA", CompanyName: "NVIDIA", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: &vol, CreatedAt: 1},
		{ID: "s2", Date: "2026-07-21", Ticker: "NVDA", CompanyName: "NVIDIA", Open: 1.5, High: 2.5, Low: 1, Close: 2, CreatedAt: 2},
	}); err != nil {
		t.Fatalf("UpsertStocks: %v", err)
	}

	h := &Handlers{conn: conn}
	req := httptest.NewRequest(http.MethodGet, "/api/ai-pulse/stocks?ticker=NVDA", nil)
	w := httptest.NewRecorder()
	h.AiPulseStocks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	var resp aiPulseStocksResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Ticker != "NVDA" {
		t.Errorf("ticker = %q, want NVDA", resp.Ticker)
	}
	if resp.CompanyName != "NVIDIA" {
		t.Errorf("companyName = %q, want NVIDIA", resp.CompanyName)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("data len = %d, want 2", len(resp.Data))
	}
	if resp.Data[0].Date != "2026-07-20" || resp.Data[1].Close != 2 {
		t.Errorf("data = %+v, want ascending by date with close 2 last", resp.Data)
	}
	if resp.Data[0].Volume == nil || *resp.Data[0].Volume != vol {
		t.Errorf("data[0].volume = %v, want %d", resp.Data[0].Volume, vol)
	}
}

func TestAiPulseStocksInvalidTicker(t *testing.T) {
	conn := newTestDB(t)
	h := &Handlers{conn: conn}

	req := httptest.NewRequest(http.MethodGet, "/api/ai-pulse/stocks?ticker=NOTREAL", nil)
	w := httptest.NewRecorder()
	h.AiPulseStocks(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] != "Invalid ticker" {
		t.Errorf("error = %q, want 'Invalid ticker'", body["error"])
	}
}

func TestAiPulseStocksMissingTicker(t *testing.T) {
	conn := newTestDB(t)
	h := &Handlers{conn: conn}

	req := httptest.NewRequest(http.MethodGet, "/api/ai-pulse/stocks", nil)
	w := httptest.NewRecorder()
	h.AiPulseStocks(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestAiPulseStocksNoDB(t *testing.T) {
	h := &Handlers{conn: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/ai-pulse/stocks?ticker=NVDA", nil)
	w := httptest.NewRecorder()
	h.AiPulseStocks(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
}

func TestAiPulseStocksEmptyHistory(t *testing.T) {
	conn := newTestDB(t)
	h := &Handlers{conn: conn}

	if !aipulse.IsValidTicker("MSFT") {
		t.Fatal("MSFT should be a valid ticker")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/ai-pulse/stocks?ticker=MSFT", nil)
	w := httptest.NewRecorder()
	h.AiPulseStocks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp aiPulseStocksResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.CompanyName != "Microsoft" || len(resp.Data) != 0 {
		t.Errorf("resp = %+v, want Microsoft with empty data", resp)
	}
}

// TestDashboardAiPulseRenders wires the stocks chart end-to-end through the
// Dashboard handler: a valid unlock cookie, a seeded stock row, and a fresh
// renderer. It asserts the ai-pulse view emits the chart script, the ticker
// selector, and the embedded initial-data JSON.
func TestDashboardAiPulseRenders(t *testing.T) {
	conn := newTestDB(t)
	vol := int64(4_000_000)
	if err := db.UpsertStocks(conn, []db.Stock{
		{ID: "s1", Date: "2026-07-20", Ticker: "NVDA", CompanyName: "NVIDIA", Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: &vol, CreatedAt: 1},
	}); err != nil {
		t.Fatalf("UpsertStocks: %v", err)
	}

	vr, err := view.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cfg := &config.Config{UnlockCookieSecret: "testsecret", BaseURL: "https://karotammela.fi"}
	h := New(cfg, vr, conn, nil, nil)

	cookieVal := security.CreateUnlockCookieValue(security.UnlockPayload{SessionID: "x", Locale: "fi", UnlockedAt: 1753352400}, "testsecret")

	req := httptest.NewRequest(http.MethodGet, "/fi/dashboard?view=ai-pulse", nil)
	req.AddCookie(&http.Cookie{Name: "karot_unlock", Value: cookieVal})
	w := httptest.NewRecorder()
	h.Dashboard(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	body, _ := io.ReadAll(w.Body)
	out := string(body)
	for _, want := range []string{
		"lightweight-charts.standalone.production.js",
		`id="ticker-select"`,
		`id="stock-chart-data"`,
		`id="stock-chart-tbody"`,
		">NVIDIA (NVDA)<",
		"1.50",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("dashboard ai-pulse output missing %q", want)
		}
	}
}

// TestDashboardAiPulseLists walks the trends + repos lists end-to-end through
// the Dashboard handler: seeded cache rows render localized (fi) summaries and
// dates, the per-source badges (HN / GitHub), and the "last updated" line.
func TestDashboardAiPulseLists(t *testing.T) {
	conn := newTestDB(t)

	sumFi := "Suomenkielinen tiivistelmä"
	descFi := "Suomenkielinen kuvaus"
	if err := db.UpsertTrends(conn, []db.Trend{
		{ID: "t1", Date: "2026-07-20", Title: "AI story", Summary: "English summary", SummaryFi: &sumFi, URL: "https://hn.test/s", Source: ptr("hackernews"), CreatedAt: 1},
	}); err != nil {
		t.Fatalf("UpsertTrends: %v", err)
	}
	if err := db.UpsertRepos(conn, []db.Repo{
		{ID: "r1", Date: "2026-07-20", RepoFullName: "owner/repo", URL: "https://gh.test/r", Description: ptr("English desc"), DescriptionFi: &descFi, Language: ptr("Go"), StarsToday: 7, Source: "github", CreatedAt: 2},
	}); err != nil {
		t.Fatalf("UpsertRepos: %v", err)
	}

	vr, err := view.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cfg := &config.Config{UnlockCookieSecret: "testsecret", BaseURL: "https://karotammela.fi"}
	h := New(cfg, vr, conn, nil, nil)

	cookieVal := security.CreateUnlockCookieValue(security.UnlockPayload{SessionID: "x", Locale: "fi", UnlockedAt: 1753352400}, "testsecret")

	req := httptest.NewRequest(http.MethodGet, "/fi/dashboard?view=ai-pulse", nil)
	req.AddCookie(&http.Cookie{Name: "karot_unlock", Value: cookieVal})
	w := httptest.NewRecorder()
	h.Dashboard(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	body, _ := io.ReadAll(w.Body)
	out := string(body)
	for _, want := range []string{
		">AI story<",
		"Suomenkielinen tiivistelmä", // fi summary chosen
		">HN<",                       // hackernews -> HN badge
		"20.07.2026",                 // fi localized date
		">owner/repo<",
		"Suomenkielinen kuvaus", // fi description
		">Go<",
		"7 " + "tähteä tänään",            // stars today (fi label)
		">GitHub<",                        // github -> GitHub badge
		"Päivitetty viimeksi: 20.07.2026", // fi last-updated line
	} {
		if !strings.Contains(out, want) {
			t.Errorf("dashboard ai-pulse lists missing %q", want)
		}
	}
}

func ptr(s string) *string { return &s }

// TestAiPulseStocksRateLimited verifies the Phase 12j per-IP rate limit on the
// stocks endpoint: the 61st request from one IP within the window is rejected
// with 429 + Retry-After and the shared error body, and the emitted log line is
// redacted (only the static event tag, never the client IP).
func TestAiPulseStocksRateLimited(t *testing.T) {
	security.SetRateLimitStore(nil)
	t.Cleanup(func() { security.SetRateLimitStore(nil) })

	h := &Handlers{conn: newTestDB(t)}
	const ip = "203.0.113.42"
	do := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/ai-pulse/stocks?ticker=NVDA", nil)
		req.Header.Set("X-Forwarded-For", ip)
		w := httptest.NewRecorder()
		h.AiPulseStocks(w, req)
		return w
	}

	for i := 1; i <= 60; i++ {
		if w := do(); w.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d rate-limited early", i)
		}
	}

	var logbuf bytes.Buffer
	oldOut, oldFlags := log.Writer(), log.Flags()
	log.SetOutput(&logbuf)
	log.SetFlags(0)
	defer func() { log.SetOutput(oldOut); log.SetFlags(oldFlags) }()

	w := do()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("missing Retry-After header on 429")
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] != "Rate limit exceeded. Retry shortly." {
		t.Errorf("error = %q", body["error"])
	}

	logged := logbuf.String()
	if !strings.Contains(logged, "ai-pulse.stocks.rate_limited") {
		t.Errorf("log missing event tag; got %q", logged)
	}
	if strings.Contains(logged, ip) {
		t.Errorf("log leaked client IP: %q", logged)
	}
}

// TestDashboardAiPulseNoDBFailClosed asserts the ai-pulse dashboard view renders
// (with empty states) when no database is configured, instead of erroring.
func TestDashboardAiPulseNoDBFailClosed(t *testing.T) {
	vr, err := view.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cfg := &config.Config{UnlockCookieSecret: "testsecret", BaseURL: "https://karotammela.fi"}
	h := New(cfg, vr, nil, nil, nil)

	cookieVal := security.CreateUnlockCookieValue(security.UnlockPayload{SessionID: "x", Locale: "fi", UnlockedAt: 1753352400}, "testsecret")

	req := httptest.NewRequest(http.MethodGet, "/fi/dashboard?view=ai-pulse", nil)
	req.AddCookie(&http.Cookie{Name: "karot_unlock", Value: cookieVal})
	w := httptest.NewRecorder()
	h.Dashboard(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	out := string(body)
	if !strings.Contains(out, "stock-chart-tbody") {
		t.Error("expected stock chart table even with no DB")
	}
	// Locale is fi: the empty-state label is the Finnish copy.
	if !strings.Contains(out, "Osakedataa ei ole vielä saatavilla") {
		t.Error("expected empty-state label when no cache exists")
	}
}

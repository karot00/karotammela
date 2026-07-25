package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"goth/internal/aipulse"
	"goth/internal/db"
)

// aiPulseStockPoint mirrors the JSON shape produced by the Next.js
// src/app/api/ai-pulse/stocks/route.ts for each cached daily point. Only
// `date` and `close` are consumed by the chart, but the full OHLCV record is
// returned for parity and an accessible table fallback.
type aiPulseStockPoint struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume *int64  `json:"volume"`
}

// aiPulseStocksResponse mirrors the Next.js route payload:
// { ticker, companyName, data }.
type aiPulseStocksResponse struct {
	Ticker      string              `json:"ticker"`
	CompanyName string              `json:"companyName"`
	Data        []aiPulseStockPoint `json:"data"`
}

func toStockPoints(rows []db.Stock) []aiPulseStockPoint {
	out := make([]aiPulseStockPoint, 0, len(rows))
	for _, r := range rows {
		out = append(out, aiPulseStockPoint{
			Date:   r.Date,
			Open:   r.Open,
			High:   r.High,
			Low:    r.Low,
			Close:  r.Close,
			Volume: r.Volume,
		})
	}
	return out
}

// AiPulseStocks handles GET /api/ai-pulse/stocks?ticker=XXX. Port of
// src/app/api/ai-pulse/stocks/route.ts: 400 for a missing/invalid ticker
// (validated against the supported AITickers set), 503 when no database is
// configured, 500 on read failure, otherwise the ticker payload read from
// Go's local cache via db.GetStockHistory.
//
// Phase 12j hardening beyond the reference: a per-IP rate limit (shared bounded
// limiter) guards this public JSON endpoint, and every log line is redacted —
// only static event tags are emitted, never the (whitelisted) ticker, the
// client IP, or DB error detail.
func (h *Handlers) AiPulseStocks(w http.ResponseWriter, r *http.Request) {
	// Rate limit before any work; even a missing ticker consumes quota.
	if !h.enforceIPRateLimit(w, r, "ai-pulse-stocks", 60, time.Minute, "ai-pulse.stocks.rate_limited", "Rate limit exceeded. Retry shortly.") {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if h.conn == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "AI Pulse database is not configured."})
		return
	}

	ticker := r.URL.Query().Get("ticker")
	if !aipulse.IsValidTicker(ticker) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid ticker"})
		return
	}

	hist, err := db.GetStockHistory(h.conn, ticker, 365)
	if err != nil {
		// Redacted: never log err (may carry query/path detail) or the ticker.
		log.Printf("ai-pulse.stocks.read_failed")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		return
	}

	resp := aiPulseStocksResponse{
		Ticker:      ticker,
		CompanyName: aipulse.CompanyName(ticker),
		Data:        toStockPoints(hist),
	}
	json.NewEncoder(w).Encode(resp)
}

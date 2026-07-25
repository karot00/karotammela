package aipulse

// TickerOption pairs a supported stock ticker with its company name. It
// mirrors the shape used by the Next.js AI Pulse shell (src/lib/ai/stocks-fetcher.ts
// AI_TICKERS) so the Go and Next.js selectors stay in lockstep.
type TickerOption struct {
	Ticker string
	Name   string
}

// AITickers is the canonical set of supported AI-company tickers. This is the
// source of truth for the stocks selector and the validity check on
// GET /api/ai-pulse/stocks. Keep it in sync with src/lib/ai/stocks-fetcher.ts.
var AITickers = []TickerOption{
	{Ticker: "NVDA", Name: "NVIDIA"},
	{Ticker: "MSFT", Name: "Microsoft"},
	{Ticker: "GOOGL", Name: "Alphabet"},
	{Ticker: "META", Name: "Meta"},
	{Ticker: "AMD", Name: "AMD"},
	{Ticker: "ARM", Name: "Arm Holdings"},
	{Ticker: "AVGO", Name: "Broadcom"},
	{Ticker: "CRWV", Name: "CoreWeave"},
	{Ticker: "PLTR", Name: "Palantir"},
	{Ticker: "AI", Name: "C3.ai"},
	{Ticker: "SNOW", Name: "Snowflake"},
	{Ticker: "SOUN", Name: "SoundHound AI"},
}

var aiTickerIndex = func() map[string]string {
	m := make(map[string]string, len(AITickers))
	for _, t := range AITickers {
		m[t.Ticker] = t.Name
	}
	return m
}()

// IsValidTicker reports whether ticker is in the supported AITickers set,
// mirroring the VALID_TICKERS check in src/app/api/ai-pulse/stocks/route.ts.
func IsValidTicker(ticker string) bool {
	_, ok := aiTickerIndex[ticker]
	return ok
}

// CompanyName returns the company name for ticker, or the ticker itself when
// unknown (mirrors AI_TICKERS.find(...).name ?? ticker in the Next.js route).
func CompanyName(ticker string) string {
	if name, ok := aiTickerIndex[ticker]; ok {
		return name
	}
	return ticker
}

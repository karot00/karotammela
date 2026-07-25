package db

import (
	"database/sql"
	"time"
)

// Trend mirrors a row of the Next.js ai_trends table copied into Go's local
// cache by the Phase 12.5 refresh pipeline (originally seeded for dev). Field
// names match the Next.js Drizzle columns (src/lib/db/schema.ts) so the copy is
// a straight remap.
type Trend struct {
	ID        string
	Date      string
	Title     string
	Summary   string
	SummaryFi *string
	URL       string
	Source    *string
	CreatedAt int64
}

// Repo mirrors a Next.js ai_repos row.
type Repo struct {
	ID            string
	Date          string
	RepoFullName  string
	URL           string
	Description   *string
	DescriptionFi *string
	Language      *string
	Stars         int
	StarsToday    int
	Source        string
	CreatedAt     int64
}

// Stock mirrors a Next.js ai_stocks row (one daily OHLCV point per ticker).
type Stock struct {
	ID          string
	Date        string
	Ticker      string
	CompanyName string
	Open        float64
	High        float64
	Low         float64
	Close       float64
	Volume      *int64
	CreatedAt   int64
}

// CacheMeta records the outcome of the last refresh run for one source.
type CacheMeta struct {
	Source string
	RanAt  int64
	Status string
	Rows   int
	Detail string
}

func sqlString(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

// UpsertTrends writes the Go-owned trends. ON CONFLICT DO NOTHING mirrors the
// Next.js onConflictDoNothing() and keeps re-insertion idempotent.
func UpsertTrends(conn *sql.DB, rows []Trend) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
		INSERT INTO ai_trends (id, date, title, summary, summary_fi, url, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, r := range rows {
		if _, err := stmt.Exec(r.ID, r.Date, r.Title, r.Summary, sqlString(r.SummaryFi), r.URL, sqlString(r.Source), r.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// UpsertRepos copies Next.js-produced repos.
func UpsertRepos(conn *sql.DB, rows []Repo) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
		INSERT INTO ai_repos (id, date, repo_full_name, url, description, description_fi, language, stars, stars_today, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, r := range rows {
		if _, err := stmt.Exec(r.ID, r.Date, r.RepoFullName, r.URL, sqlString(r.Description), sqlString(r.DescriptionFi), sqlString(r.Language), r.Stars, r.StarsToday, r.Source, r.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// UpsertStocks copies Next.js-produced daily stock points.
func UpsertStocks(conn *sql.DB, rows []Stock) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
		INSERT INTO ai_stocks (id, date, ticker, company_name, open, high, low, close, volume, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, r := range rows {
		if _, err := stmt.Exec(r.ID, r.Date, r.Ticker, r.CompanyName, r.Open, r.High, r.Low, r.Close, sqlInt64(r.Volume), r.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func sqlInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

// todayUTC returns the current date as YYYY-MM-DD in UTC, matching the
// Next.js getLatestTrends/getLatestRepos fallback base date.
func todayUTC() string {
	return time.Now().UTC().Format("2006-01-02")
}

// GetLatestTrends returns the trends for date, or — when none exist for that
// date — the most recent day that has rows (mirrors getLatestTrends).
func GetLatestTrends(conn *sql.DB, date string) ([]Trend, error) {
	if date == "" {
		date = todayUTC()
	}
	rows, err := queryTrends(conn, date)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		return rows, nil
	}
	var latest string
	err = conn.QueryRow(`SELECT date FROM ai_trends ORDER BY date DESC LIMIT 1`).Scan(&latest)
	if err == sql.ErrNoRows {
		return []Trend{}, nil
	}
	if err != nil {
		return nil, err
	}
	return queryTrends(conn, latest)
}

func queryTrends(conn *sql.DB, date string) ([]Trend, error) {
	rows, err := conn.Query(`
		SELECT id, date, title, summary, summary_fi, url, source, created_at
		FROM ai_trends WHERE date = ? ORDER BY created_at ASC
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Trend
	for rows.Next() {
		var t Trend
		var sumFi, src sql.NullString
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Summary, &sumFi, &t.URL, &src, &t.CreatedAt); err != nil {
			return nil, err
		}
		if sumFi.Valid {
			t.SummaryFi = &sumFi.String
		}
		if src.Valid {
			t.Source = &src.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetRecentTrendUrls mirrors getRecentTrendUrls(days): the set of trend URLs
// inserted in the last days days, used to dedup before Gemini summarization.
// The cutoff is a UTC date, matching `cutoff.toISOString().slice(0,10)`.
func GetRecentTrendUrls(conn *sql.DB, days int) (map[string]bool, error) {
	if days <= 0 {
		days = 7
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02")
	rows, err := conn.Query(`SELECT url FROM ai_trends WHERE date >= ?`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		out[u] = true
	}
	return out, rows.Err()
}

// GetLatestRepos mirrors getLatestRepos (date match, else most recent day).
func GetLatestRepos(conn *sql.DB, date string) ([]Repo, error) {
	if date == "" {
		date = todayUTC()
	}
	rows, err := queryRepos(conn, date)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		return rows, nil
	}
	var latest string
	err = conn.QueryRow(`SELECT date FROM ai_repos ORDER BY date DESC LIMIT 1`).Scan(&latest)
	if err == sql.ErrNoRows {
		return []Repo{}, nil
	}
	if err != nil {
		return nil, err
	}
	return queryRepos(conn, latest)
}

func queryRepos(conn *sql.DB, date string) ([]Repo, error) {
	rows, err := conn.Query(`
		SELECT id, date, repo_full_name, url, description, description_fi, language, stars, stars_today, source, created_at
		FROM ai_repos WHERE date = ? ORDER BY created_at ASC
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Repo
	for rows.Next() {
		var r Repo
		var desc, descFi, lang sql.NullString
		if err := rows.Scan(&r.ID, &r.Date, &r.RepoFullName, &r.URL, &desc, &descFi, &lang, &r.Stars, &r.StarsToday, &r.Source, &r.CreatedAt); err != nil {
			return nil, err
		}
		if desc.Valid {
			r.Description = &desc.String
		}
		if descFi.Valid {
			r.DescriptionFi = &descFi.String
		}
		if lang.Valid {
			r.Language = &lang.String
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetStockHistory mirrors getStockHistory: daily points for ticker within the
// last days days (default 365), oldest first.
func GetStockHistory(conn *sql.DB, ticker string, days int) ([]Stock, error) {
	if days <= 0 {
		days = 365
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02")
	rows, err := conn.Query(`
		SELECT id, date, ticker, company_name, open, high, low, close, volume, created_at
		FROM ai_stocks WHERE ticker = ? AND date >= ? ORDER BY date ASC
	`, ticker, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Stock
	for rows.Next() {
		var s Stock
		var vol sql.NullInt64
		if err := rows.Scan(&s.ID, &s.Date, &s.Ticker, &s.CompanyName, &s.Open, &s.High, &s.Low, &s.Close, &vol, &s.CreatedAt); err != nil {
			return nil, err
		}
		if vol.Valid {
			s.Volume = &vol.Int64
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// GetAvailableTickers mirrors getAvailableTickers: distinct tickers, ascending.
func GetAvailableTickers(conn *sql.DB) ([]string, error) {
	rows, err := conn.Query(`SELECT DISTINCT ticker FROM ai_stocks ORDER BY ticker ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// UpsertCacheMeta records the outcome of a refresh run for one source,
// replacing any prior row for that source.
func UpsertCacheMeta(conn *sql.DB, source string, ranAt int64, status string, rows int, detail string) error {
	_, err := conn.Exec(`
		INSERT INTO ai_cache_meta (source, ran_at, status, rows, detail)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(source) DO UPDATE SET
			ran_at = excluded.ran_at,
			status = excluded.status,
			rows   = excluded.rows,
			detail = excluded.detail
	`, source, ranAt, status, rows, detail)
	return err
}

// GetCacheMeta returns the last refresh record for source, or nil if none.
func GetCacheMeta(conn *sql.DB, source string) (*CacheMeta, error) {
	var m CacheMeta
	var detail sql.NullString
	err := conn.QueryRow(`
		SELECT source, ran_at, status, rows, detail FROM ai_cache_meta WHERE source = ?
	`, source).Scan(&m.Source, &m.RanAt, &m.Status, &m.Rows, &detail)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if detail.Valid {
		m.Detail = detail.String
	}
	return &m, nil
}

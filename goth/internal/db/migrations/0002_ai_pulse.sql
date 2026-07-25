-- AI Pulse local cache (Phase 12b).
--
-- Mirrors the Next.js ai_trends / ai_repos / ai_stocks tables so the Go
-- dashboard (12f/12g) reads locally and stays resilient if Next.js is briefly
-- down. As of Phase 12.5 Go is the writer: the refresh pipeline populates
-- these tables directly. (Phase 12b briefly used a temporary Next.js reader.)
--
-- Column names match the Next.js Drizzle schema (src/lib/db/schema.ts) so the
-- copy is a straight column remap. Unique indexes match the Next.js unique
-- indexes so re-insertion is idempotent (ON CONFLICT DO NOTHING).

CREATE TABLE IF NOT EXISTS ai_trends (
    id         TEXT    PRIMARY KEY,
    date       TEXT    NOT NULL,
    title      TEXT    NOT NULL,
    summary    TEXT    NOT NULL,
    summary_fi TEXT,
    url        TEXT    NOT NULL,
    source     TEXT,
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000)
);

CREATE INDEX  IF NOT EXISTS ai_trends_date_idx        ON ai_trends(date);
CREATE UNIQUE INDEX IF NOT EXISTS ai_trends_date_title_idx ON ai_trends(date, title);

CREATE TABLE IF NOT EXISTS ai_repos (
    id             TEXT    PRIMARY KEY,
    date           TEXT    NOT NULL,
    repo_full_name TEXT    NOT NULL,
    url            TEXT    NOT NULL,
    description    TEXT,
    description_fi TEXT,
    language       TEXT,
    stars          INTEGER NOT NULL DEFAULT 0,
    stars_today    INTEGER NOT NULL DEFAULT 0,
    source         TEXT    NOT NULL DEFAULT 'github-trending',
    created_at     INTEGER NOT NULL DEFAULT (unixepoch() * 1000)
);

CREATE INDEX  IF NOT EXISTS ai_repos_date_idx            ON ai_repos(date);
CREATE UNIQUE INDEX IF NOT EXISTS ai_repos_date_repo_idx ON ai_repos(date, repo_full_name);

CREATE TABLE IF NOT EXISTS ai_stocks (
    id           TEXT    PRIMARY KEY,
    date         TEXT    NOT NULL,
    ticker       TEXT    NOT NULL,
    company_name TEXT    NOT NULL,
    open         REAL    NOT NULL,
    high         REAL    NOT NULL,
    low          REAL    NOT NULL,
    close        REAL    NOT NULL,
    volume       INTEGER,
    created_at   INTEGER NOT NULL DEFAULT (unixepoch() * 1000)
);

CREATE INDEX IF NOT EXISTS ai_stocks_ticker_date_idx ON ai_stocks(ticker, date);

-- Cache-metadata: one row per source tracking the last successful/failed
-- refresh run. Keeps the dashboard honest about data freshness and lets
-- 12f/12g render an offline/empty state safely.
CREATE TABLE IF NOT EXISTS ai_cache_meta (
    source  TEXT    PRIMARY KEY,
    ran_at  INTEGER NOT NULL,
    status  TEXT    NOT NULL,
    rows    INTEGER NOT NULL DEFAULT 0,
    detail  TEXT
);

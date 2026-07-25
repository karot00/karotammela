-- Phase 12.5c: make ai_stocks_ticker_date_idx UNIQUE, matching the Next.js
-- schema (src/lib/db/schema.ts uniqueIndex on (ticker, date)). The Phase 12b
-- migration created it as a plain index, so repeated stock refreshes would
-- insert duplicate daily points instead of being idempotent.
--
-- Defensive dedupe first: if any duplicate (ticker, date) rows already exist
-- (possible while the plain index was in place), keep the oldest row so the
-- unique index build cannot fail.

DELETE FROM ai_stocks
WHERE rowid NOT IN (
    SELECT MIN(rowid) FROM ai_stocks GROUP BY ticker, date
);

DROP INDEX IF EXISTS ai_stocks_ticker_date_idx;
CREATE UNIQUE INDEX IF NOT EXISTS ai_stocks_ticker_date_idx ON ai_stocks(ticker, date);

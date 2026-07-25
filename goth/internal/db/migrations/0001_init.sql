-- SENTINEL-7 persistence: mirrors sessions + logs from the Next.js app.
CREATE TABLE IF NOT EXISTS sessions (
    session_id TEXT PRIMARY KEY,
    locale TEXT NOT NULL CHECK (locale IN ('en', 'fi')),
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    last_level INTEGER NOT NULL DEFAULT 0,
    unlocked INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS logs (
    id TEXT PRIMARY KEY,
    timestamp INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    session_id TEXT NOT NULL REFERENCES sessions(session_id) ON DELETE CASCADE,
    locale TEXT NOT NULL CHECK (locale IN ('en', 'fi')),
    user_input TEXT NOT NULL,
    assistant_output TEXT,
    level_reached INTEGER NOT NULL,
    success INTEGER NOT NULL DEFAULT 0,
    CHECK (level_reached >= 0 AND level_reached <= 100)
);

CREATE INDEX IF NOT EXISTS logs_session_idx ON logs(session_id);
CREATE INDEX IF NOT EXISTS logs_timestamp_idx ON logs(timestamp);

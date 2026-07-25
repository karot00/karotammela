package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"
)

// Stats mirrors the Next.js getDashboardStats() result
// (src/lib/db/queries.ts) field-for-field:
//   - TotalAttempts:       count(*) over logs.
//   - UnlockedCount:       count of successful log rows (NOT distinct sessions).
//   - DirectUnlockCount:   successful rows whose trimmed, uppercased input
//     equals the access code.
//   - HighestLevel:        max(level_reached), 0 when empty.
//   - AvgMessagesToUnlock: average of per-session total message counts over
//     sessions that contain at least one successful row, 0 when none.
//   - LatestUnlock:        timestamp of the most recent successful row, nil
//     when none.
type Stats struct {
	TotalAttempts       int
	UnlockedCount       int
	DirectUnlockCount   int
	HighestLevel        int
	AvgMessagesToUnlock float64
	LatestUnlock        *time.Time
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// UpsertSession ensures a session row exists and updates its state. The
// conflict clause mirrors persistSentinelTurn's onConflictDoUpdate in
// src/lib/db/queries.ts: locale, last_level, unlocked, and updated_at are
// all overwritten with the latest turn's values (unlocked is the latest
// turn's success flag, not a sticky maximum).
func UpsertSession(conn *sql.DB, sessionID, locale string, level int, unlocked bool) error {
	now := time.Now().UnixMilli()
	_, err := conn.Exec(`
		INSERT INTO sessions (session_id, locale, created_at, updated_at, last_level, unlocked)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET
			locale = excluded.locale,
			updated_at = excluded.updated_at,
			last_level = excluded.last_level,
			unlocked = excluded.unlocked
	`, sessionID, locale, now, now, level, boolToInt(unlocked))
	return err
}

// PersistTurn records a single sentinel exchange.
func PersistTurn(conn *sql.DB, sessionID, locale, userInput, assistantOutput string, level int, success bool) error {
	if err := UpsertSession(conn, sessionID, locale, level, success); err != nil {
		return err
	}
	_, err := conn.Exec(`
		INSERT INTO logs (id, session_id, locale, user_input, assistant_output, level_reached, success)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, newID(), sessionID, locale, userInput, nullStr(assistantOutput), level, boolToInt(success))
	return err
}

// GetStats returns the aggregate counters with the exact Next.js
// getDashboardStats semantics. accessCode is the sentinel access code used
// for the direct-unlock comparison (uppercased before matching).
func GetStats(conn *sql.DB, accessCode string) (Stats, error) {
	var s Stats

	var unlocked, direct, highest sql.NullInt64
	err := conn.QueryRow(`
		SELECT
			COUNT(*),
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END),
			SUM(CASE WHEN success = 1 AND UPPER(TRIM(user_input)) = ? THEN 1 ELSE 0 END),
			MAX(level_reached)
		FROM logs
	`, strings.ToUpper(accessCode)).Scan(&s.TotalAttempts, &unlocked, &direct, &highest)
	if err != nil {
		return Stats{}, err
	}
	s.UnlockedCount = int(unlocked.Int64)
	s.DirectUnlockCount = int(direct.Int64)
	s.HighestLevel = int(highest.Int64)

	var avg sql.NullFloat64
	err = conn.QueryRow(`
		SELECT COALESCE(AVG(cnt), 0) FROM (
			SELECT COUNT(*) AS cnt FROM logs
			GROUP BY session_id
			HAVING SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) > 0
		)
	`).Scan(&avg)
	if err != nil {
		return Stats{}, err
	}
	s.AvgMessagesToUnlock = avg.Float64

	var latest sql.NullInt64
	err = conn.QueryRow(`
		SELECT timestamp FROM logs WHERE success = 1 ORDER BY timestamp DESC LIMIT 1
	`).Scan(&latest)
	if err != nil && err != sql.ErrNoRows {
		return Stats{}, err
	}
	if latest.Valid {
		t := time.UnixMilli(latest.Int64)
		s.LatestUnlock = &t
	}
	return s, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

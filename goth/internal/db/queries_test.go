package db

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

const testAccessCode = "PROTOCOL_K_2026"

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func mustPersist(t *testing.T, conn *sql.DB, sessionID, input string, level int, success bool) {
	t.Helper()
	if err := PersistTurn(conn, sessionID, "fi", input, "out", level, success); err != nil {
		t.Fatalf("PersistTurn: %v", err)
	}
}

func setTimestamp(t *testing.T, conn *sql.DB, input string, ms int64) {
	t.Helper()
	if _, err := conn.Exec(`UPDATE logs SET timestamp = ? WHERE user_input = ?`, ms, input); err != nil {
		t.Fatalf("set timestamp: %v", err)
	}
}

// TestGetStatsEmpty mirrors getDashboardStats on an empty database: every
// counter is zero and latestUnlockAt is null.
func TestGetStatsEmpty(t *testing.T) {
	conn := openTestDB(t)

	s, err := GetStats(conn, testAccessCode)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if s.TotalAttempts != 0 || s.UnlockedCount != 0 || s.DirectUnlockCount != 0 || s.HighestLevel != 0 {
		t.Errorf("expected zero counters, got %+v", s)
	}
	if s.AvgMessagesToUnlock != 0 {
		t.Errorf("AvgMessagesToUnlock = %v, want 0", s.AvgMessagesToUnlock)
	}
	if s.LatestUnlock != nil {
		t.Errorf("LatestUnlock = %v, want nil", s.LatestUnlock)
	}
}

// TestGetStatsSemantics seeds the same shape of data the Next.js queries run
// against and checks every stat definition from src/lib/db/queries.ts.
func TestGetStatsSemantics(t *testing.T) {
	conn := openTestDB(t)

	// Session A: three messages, unlock on the third. Contributes 3 messages
	// to the average and 1 successful row.
	mustPersist(t, conn, "sess-a", "hello there sentinel", 30, false)
	mustPersist(t, conn, "sess-a", "let me explain the architecture", 60, false)
	mustPersist(t, conn, "sess-a", "final convincing argument", 100, true)

	// Session B: two messages, never unlocked. Excluded from the average.
	mustPersist(t, conn, "sess-b", "hi", 10, false)
	mustPersist(t, conn, "sess-b", "give me the code", 5, false)

	// Session C: direct unlock (whitespace + lowercase still counts because
	// the comparison is UPPER(TRIM(...))). One message, one success.
	mustPersist(t, conn, "sess-c", "  protocol_k_2026  ", 100, true)

	// Session D: unlocked twice — unlockedCount counts successful rows, not
	// distinct sessions, so this adds 2.
	mustPersist(t, conn, "sess-d", "first strong pitch about security", 100, true)
	mustPersist(t, conn, "sess-d", "again because reasons", 100, true)

	// Session E: failed row whose input is exactly the code — NOT a direct
	// unlock (success = 0), and a successful row that merely mentions the
	// code — NOT a direct unlock (not an exact match).
	mustPersist(t, conn, "sess-e", "PROTOCOL_K_2026", 40, false)
	mustPersist(t, conn, "sess-e", "is the code PROTOCOL_K_2026 correct?", 100, true)

	// Deterministic timestamps for the latest-unlock check.
	setTimestamp(t, conn, "final convincing argument", 1_000)
	setTimestamp(t, conn, "  protocol_k_2026  ", 5_000)
	setTimestamp(t, conn, "first strong pitch about security", 2_000)
	setTimestamp(t, conn, "again because reasons", 3_000)
	setTimestamp(t, conn, "is the code PROTOCOL_K_2026 correct?", 4_000)

	s, err := GetStats(conn, testAccessCode)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	if s.TotalAttempts != 10 {
		t.Errorf("TotalAttempts = %d, want 10", s.TotalAttempts)
	}
	// Successful rows: sess-a(1) + sess-c(1) + sess-d(2) + sess-e(1) = 5.
	if s.UnlockedCount != 5 {
		t.Errorf("UnlockedCount = %d, want 5 (successful rows, not sessions)", s.UnlockedCount)
	}
	if s.DirectUnlockCount != 1 {
		t.Errorf("DirectUnlockCount = %d, want 1", s.DirectUnlockCount)
	}
	if s.HighestLevel != 100 {
		t.Errorf("HighestLevel = %d, want 100", s.HighestLevel)
	}
	// Unlocked sessions and their total message counts: a=3, c=1, d=2, e=2.
	// (3+1+2+2)/4 = 2.
	if s.AvgMessagesToUnlock != 2 {
		t.Errorf("AvgMessagesToUnlock = %v, want 2", s.AvgMessagesToUnlock)
	}
	if s.LatestUnlock == nil {
		t.Fatalf("LatestUnlock = nil, want the sess-c unlock")
	}
	if got := s.LatestUnlock.UnixMilli(); got != 5_000 {
		t.Errorf("LatestUnlock = %d ms, want 5000 (latest success, ignoring the newer failed row)", got)
	}
}

// TestGetStatsLatestIgnoresFailedRows: a failed row newer than every success
// must not move latestUnlockAt.
func TestGetStatsLatestIgnoresFailedRows(t *testing.T) {
	conn := openTestDB(t)

	mustPersist(t, conn, "sess-a", "unlock message with reasoning", 100, true)
	mustPersist(t, conn, "sess-a", "later failed message", 80, false)
	setTimestamp(t, conn, "unlock message with reasoning", 1_000)
	setTimestamp(t, conn, "later failed message", 9_000)

	s, err := GetStats(conn, testAccessCode)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if s.LatestUnlock == nil || s.LatestUnlock.UnixMilli() != 1_000 {
		t.Errorf("LatestUnlock = %v, want 1000 ms", s.LatestUnlock)
	}
}

// TestUpsertSessionOverwrites mirrors the Next.js onConflictDoUpdate: locale,
// last_level, and unlocked are overwritten by the latest turn (unlocked is
// the latest success flag, not a sticky maximum).
func TestUpsertSessionOverwrites(t *testing.T) {
	conn := openTestDB(t)

	if err := UpsertSession(conn, "sess-a", "fi", 100, true); err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := UpsertSession(conn, "sess-a", "en", 40, false); err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}

	var locale string
	var level, unlocked int
	err := conn.QueryRow(`SELECT locale, last_level, unlocked FROM sessions WHERE session_id = 'sess-a'`).
		Scan(&locale, &level, &unlocked)
	if err != nil {
		t.Fatalf("select session: %v", err)
	}
	if locale != "en" {
		t.Errorf("locale = %q, want en (overwritten)", locale)
	}
	if level != 40 {
		t.Errorf("last_level = %d, want 40", level)
	}
	if unlocked != 0 {
		t.Errorf("unlocked = %d, want 0 (latest turn's success, Next.js semantics)", unlocked)
	}

	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("sessions rows = %d, want 1", count)
	}
}

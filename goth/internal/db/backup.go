package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// backupPrefix and backupExt define the file naming scheme for automatic
// backups: goth-YYYYMMDD-HHMMSS.db. Prune only ever touches files matching
// this exact shape so a manually named snapshot is never deleted.
const (
	backupPrefix = "goth-"
	backupExt    = ".db"
)

// Backup writes a consistent snapshot of the SQLite database at srcPath to
// destPath using VACUUM INTO (safe against a live WAL-mode writer), then
// verifies the copy with PRAGMA integrity_check before moving it into place.
//
// The snapshot is written to destPath+".tmp" first and renamed only after
// verification, so destPath either does not appear at all or is a verified,
// complete backup. destPath must not already exist.
func Backup(srcPath, destPath string) error {
	if _, err := os.Stat(srcPath); err != nil {
		return fmt.Errorf("backup: source database: %w", err)
	}
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("backup: destination %s already exists", destPath)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o750); err != nil {
		return fmt.Errorf("backup: create destination dir: %w", err)
	}

	tmpPath := destPath + ".tmp"
	// A leftover .tmp from a crashed run would make VACUUM INTO fail.
	if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("backup: clear stale temp file: %w", err)
	}

	// Read-side connection: no migrations, no schema changes. busy_timeout
	// lets the snapshot wait briefly for a concurrent writer instead of
	// failing immediately.
	src, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)", srcPath))
	if err != nil {
		return fmt.Errorf("backup: open source: %w", err)
	}
	defer src.Close()

	if _, err := src.Exec("VACUUM INTO ?", tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("backup: vacuum into: %w", err)
	}

	if err := verifySnapshot(tmpPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("backup: finalize: %w", err)
	}
	return nil
}

// verifySnapshot opens the freshly written snapshot read-only and requires
// PRAGMA integrity_check to report "ok". A backup that cannot be verified is
// worse than no backup because it silently poisons the restore drill.
func verifySnapshot(path string) error {
	conn, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", path))
	if err != nil {
		return fmt.Errorf("backup: open snapshot: %w", err)
	}
	defer conn.Close()

	var result string
	if err := conn.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil {
		return fmt.Errorf("backup: integrity check: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("backup: integrity check failed: %s", result)
	}
	return nil
}

// BackupFilename returns the timestamped basename used for automatic backups.
func BackupFilename(now time.Time) string {
	return backupPrefix + now.UTC().Format("20060102-150405") + backupExt
}

// PruneBackups removes the oldest automatic backups in dir, keeping the
// newest keep files. Only files matching the goth-YYYYMMDD-HHMMSS.db naming
// scheme are considered; anything else in the directory is left alone.
// It returns the paths that were removed.
func PruneBackups(dir string, keep int) ([]string, error) {
	if keep < 1 {
		return nil, fmt.Errorf("prune: keep must be >= 1, got %d", keep)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("prune: read dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if isAutoBackupName(e.Name()) {
			names = append(names, e.Name())
		}
	}
	// The timestamp format sorts lexicographically = chronologically.
	sort.Strings(names)
	if len(names) <= keep {
		return nil, nil
	}

	var removed []string
	for _, name := range names[:len(names)-keep] {
		path := filepath.Join(dir, name)
		if err := os.Remove(path); err != nil {
			return removed, fmt.Errorf("prune: remove %s: %w", path, err)
		}
		removed = append(removed, path)
	}
	return removed, nil
}

// isAutoBackupName reports whether name matches goth-YYYYMMDD-HHMMSS.db.
func isAutoBackupName(name string) bool {
	if !strings.HasPrefix(name, backupPrefix) || !strings.HasSuffix(name, backupExt) {
		return false
	}
	stamp := strings.TrimSuffix(strings.TrimPrefix(name, backupPrefix), backupExt)
	_, err := time.Parse("20060102-150405", stamp)
	return err == nil
}

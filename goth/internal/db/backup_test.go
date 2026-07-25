package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "goth.db")
	conn, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	if _, err := conn.Exec(
		`INSERT INTO sessions (session_id, locale, last_level, unlocked) VALUES ('backup-test', 'fi', 42, 0)`,
	); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return path
}

func TestBackupCreatesVerifiedSnapshot(t *testing.T) {
	src := newTestDB(t)
	dest := filepath.Join(t.TempDir(), "snap.db")

	if err := Backup(src, dest); err != nil {
		t.Fatalf("backup: %v", err)
	}

	conn, err := Open(dest)
	if err != nil {
		t.Fatalf("open snapshot: %v", err)
	}
	defer conn.Close()

	var level int
	if err := conn.QueryRow(`SELECT last_level FROM sessions WHERE session_id = 'backup-test'`).Scan(&level); err != nil {
		t.Fatalf("query snapshot: %v", err)
	}
	if level != 42 {
		t.Fatalf("last_level = %d, want 42", level)
	}

	if _, err := os.Stat(dest + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temp file left behind: %v", err)
	}
}

func TestBackupRefusesExistingDestination(t *testing.T) {
	src := newTestDB(t)
	dest := filepath.Join(t.TempDir(), "snap.db")
	if err := os.WriteFile(dest, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Backup(src, dest); err == nil {
		t.Fatal("expected error for existing destination")
	}
}

func TestBackupMissingSource(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "snap.db")
	if err := Backup(filepath.Join(t.TempDir(), "missing.db"), dest); err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestBackupFilename(t *testing.T) {
	ts := time.Date(2026, 7, 24, 8, 30, 15, 0, time.UTC)
	got := BackupFilename(ts)
	want := "goth-20260724-083015.db"
	if got != want {
		t.Fatalf("BackupFilename = %q, want %q", got, want)
	}
	if !isAutoBackupName(got) {
		t.Fatalf("generated name %q not recognized by isAutoBackupName", got)
	}
}

func TestPruneBackupsKeepsNewestAndForeignFiles(t *testing.T) {
	dir := t.TempDir()
	auto := []string{
		"goth-20260101-000000.db",
		"goth-20260102-000000.db",
		"goth-20260103-000000.db",
		"goth-20260104-000000.db",
	}
	foreign := []string{"manual-snapshot.db", "goth-notatimestamp.db", "notes.txt"}
	for _, name := range append(append([]string{}, auto...), foreign...) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	removed, err := PruneBackups(dir, 2)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if len(removed) != 2 {
		t.Fatalf("removed %d files, want 2: %v", len(removed), removed)
	}

	for _, name := range auto[2:] {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("newest backup %s should remain: %v", name, err)
		}
	}
	for _, name := range auto[:2] {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("oldest backup %s should be removed", name)
		}
	}
	for _, name := range foreign {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("foreign file %s must never be pruned: %v", name, err)
		}
	}
}

func TestPruneBackupsNoopUnderKeep(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "goth-20260101-000000.db"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	removed, err := PruneBackups(dir, 5)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if len(removed) != 0 {
		t.Fatalf("expected no removals, got %v", removed)
	}
}

func TestPruneBackupsRejectsZeroKeep(t *testing.T) {
	if _, err := PruneBackups(t.TempDir(), 0); err == nil {
		t.Fatal("expected error for keep=0")
	}
}

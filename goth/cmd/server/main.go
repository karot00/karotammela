package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goth/internal/ai"
	"goth/internal/aipulse"
	"goth/internal/config"
	"goth/internal/content"
	"goth/internal/db"
	"goth/internal/email"
	"goth/internal/handler"
	"goth/internal/router"
	"goth/internal/view"
)

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func main() {
	loadDotEnv(".env")

	cfg := config.Load()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			runMigrate(cfg)
			return
		case "refresh":
			runRefresh(cfg)
			return
		case "backup":
			runBackup(cfg, os.Args[2:])
			return
		default:
			log.Fatalf("unknown command %q (expected migrate|refresh|backup)", os.Args[1])
		}
	}

	runServer(cfg)
}

func runMigrate(cfg *config.Config) {
	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}
	conn.Close()
	fmt.Println("migrations applied")
}

// runBackup is the plan §11.3 online-backup path: VACUUM INTO a temp file,
// verify with PRAGMA integrity_check, rename into place, then prune old
// automatic snapshots. With an explicit destination argument it writes a
// one-off snapshot and skips pruning; without one it writes a timestamped
// file into GOTH_BACKUP_DIR and keeps the newest GOTH_BACKUP_KEEP files.
// Exit status is observable for the systemd timer: 0 = verified backup on
// disk, non-zero = no new backup (the tmp file is removed on failure).
func runBackup(cfg *config.Config, args []string) {
	if len(args) > 1 {
		log.Fatalf("backup: expected at most one destination argument, got %d", len(args))
	}

	if len(args) == 1 {
		dest := args[0]
		if err := db.Backup(cfg.DBPath, dest); err != nil {
			log.Fatalf("backup: %v", err)
		}
		fmt.Printf("backup written: %s\n", dest)
		return
	}

	dest := filepath.Join(cfg.BackupDir, db.BackupFilename(time.Now()))
	if err := db.Backup(cfg.DBPath, dest); err != nil {
		log.Fatalf("backup: %v", err)
	}
	fmt.Printf("backup written: %s\n", dest)

	removed, err := db.PruneBackups(cfg.BackupDir, cfg.BackupKeep)
	if err != nil {
		log.Fatalf("backup: prune (new backup is intact at %s): %v", dest, err)
	}
	for _, path := range removed {
		fmt.Printf("pruned: %s\n", path)
	}
	fmt.Printf("retention: keeping newest %d automatic backups in %s\n", cfg.BackupKeep, cfg.BackupDir)
}

// buildRefresher wires the Phase 12.5 AI Pulse writer pipeline. Without a
// Gemini key the fetchers use their deterministic fallback content (raw
// titles/descriptions), so the pipeline still refreshes offline.
func buildRefresher(cfg *config.Config) *aipulse.Refresher {
	var summarizer aipulse.TextSummarizer
	if cfg.GoogleAPIKey != "" {
		summarizer = &aipulse.GeminiSummarizer{APIKey: cfg.GoogleAPIKey, Model: cfg.AIModel, BaseURL: cfg.GeminiBaseURL}
	}
	return &aipulse.Refresher{
		Trends: &aipulse.HNFetcher{BaseURL: cfg.HNBaseURL, Summarizer: summarizer},
		Repos:  &aipulse.GitHubFetcher{BaseURL: cfg.GitHubBaseURL, Summarizer: summarizer, PageDelay: time.Second},
		Stocks: &aipulse.YahooFetcher{BaseURL: cfg.YahooBaseURL},
	}
}

// runRefresh is the internal/manual refresh command path (Phase 12.5d): it
// runs the same orchestrator as POST /api/ai-pulse/refresh directly against
// the local database. Exit status is observable for the systemd unit (12.5e):
// 0 when every source succeeded, 1 when any source failed or the DB is
// unavailable.
func runRefresh(cfg *config.Config) {
	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("refresh: db open: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	rep, ran := buildRefresher(cfg).TryRun(ctx, conn)
	if !ran {
		log.Fatal("refresh: another run is already in progress in this process")
	}
	fmt.Printf("refresh complete: trends=%d(ok=%v) repos=%d(ok=%v) stocks=%d(ok=%v)\n",
		rep.Trends.Inserted, rep.Trends.OK, rep.Repos.Inserted, rep.Repos.OK, rep.Stocks.Inserted, rep.Stocks.OK)
	if rep.AnyFailed() {
		os.Exit(1)
	}
}

func runServer(cfg *config.Config) {
	// Fail closed on a half-configured VIP portal before binding (plan §5.1).
	if err := cfg.VIPStartupError(); err != nil {
		log.Fatalf("vip config: %v", err)
	}
	var vipContent content.VIPContent
	if cfg.VIPEnabled {
		var err error
		vipContent, err = content.LoadVIP(cfg.VIPContentDir)
		if err != nil {
			log.Fatalf("vip content: %v", err)
		}
	}

	vr, err := view.NewRenderer()
	if err != nil {
		log.Fatalf("view init: %v", err)
	}

	conn, derr := db.Open(cfg.DBPath)
	if derr != nil {
		log.Printf("db open failed (stats disabled): %v", derr)
		conn = nil
	}
	var gemini *ai.GeminiStreamer
	if cfg.GoogleAPIKey != "" {
		gemini = &ai.GeminiStreamer{APIKey: cfg.GoogleAPIKey, Model: cfg.AIModel}
	}
	var vipGemini *ai.GeminiStreamer
	if cfg.GoogleAPIKey != "" {
		vipGemini = &ai.GeminiStreamer{APIKey: cfg.GoogleAPIKey, Model: cfg.VIPAIModel}
	}
	var mailer handler.MailSender
	if cfg.ContactConfigured() {
		mailer = &email.ResendSender{APIKey: cfg.ResendAPIKey}
	}

	h := handler.New(cfg, vr, conn, gemini, mailer)
	h.SetVIPGemini(vipGemini)
	h.SetVIPContent(vipContent)
	h.SetRefresher(buildRefresher(cfg))
	r := router.New(h)

	addr := cfg.Host + ":" + cfg.Port
	fmt.Printf("GOTH listening on %s (env=%s, ai=%v, contact=%v, vip=%v)\n", addr, cfg.Env, cfg.GoogleAPIKey != "", cfg.ContactConfigured(), cfg.VIPEnabled)
	server := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

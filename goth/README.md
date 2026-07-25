# GOTH — Go + HTMX copy of karotammela.fi

An interactive **Tech Switcher** case study: a low-overhead rebuild of the public
Agentic AI Lab in **Go + HTMX + Alpine.js + Tailwind v4**, toggleable against the
live Next.js build in real time behind a Caddy reverse proxy.

The Go app is visually identical to the Next.js version (shared Tailwind design
tokens) while shipping **near-zero JavaScript** (HTMX + ~3 kB Alpine) and a single
deployable binary.

## Stack

| Concern        | Choice                                            |
| -------------- | ------------------------------------------------- |
| Language       | Go 1.23+                                          |
| Router         | `go-chi/chi/v5`                                   |
| Templating     | `html/template` (embedded, no codegen)            |
| Interactivity  | `htmx.org` (SSE) + `alpinejs`                     |
| Styling        | Tailwind CSS v4 (`@tailwindcss/cli`)              |
| DB             | `modernc.org/sqlite` (CGO-free)                   |
| Markdown       | `goldmark` + `goldmark-meta` + `bluemonday`       |
| AI             | Google Gemini REST `streamGenerateContent` (SSE)  |

## Layout

```
goth/
├── cmd/server/        entrypoint
├── internal/          config, router, handler, ai, content, db, i18n, security, view
├── web/               input.css, embedded static/ (compiled app.css, htmx, alpine)
├── content/blog/      embedded markdown posts (en/fi)
├── migrations/        0001_init.sql
├── Makefile
├── deploy/            production artifacts: caddy/, systemd/, deploy.sh, goth.env.example, smoke-local.sh
├── docs/              contracts.md, runbook.md
├── Caddyfile.example  cookie-routing snippet (dev; production lives in deploy/caddy/)
└── .env.example
```

## Develop

```bash
# install JS tooling for CSS (Tailwind v4)
npm install

# build CSS + run the Go server on :8080
make run

# or everything from scratch
make build && ./bin/goth
```

Set `GOOGLE_GENERATIVE_AI_API_KEY` in a local `.env` for live SENTINEL-7. Without it
the terminal falls back to a deterministic offline stub stream.

## Tech Switcher + Perf Widget

The floating widget sets a `tech=go|next` cookie (path `/`) and reloads. To compare
both builds behind one hostname, run Caddy from `Caddyfile.example` so it routes by
cookie to Next.js (`:3000`) or Go (`:8080`).

## AI Pulse pipeline (Phase 12.5)

Go owns the AI Pulse writers end to end: Hacker News (Algolia), GitHub Trending,
and Yahoo Finance fetchers with Gemini EN/FI summarization, persisted to the local
`ai_trends` / `ai_repos` / `ai_stocks` tables.

```bash
./bin/goth migrate   # apply SQLite migrations (idempotent)
./bin/goth refresh   # run one refresh; exit 1 if any source failed
./bin/goth backup    # verified SQLite snapshot into GOTH_BACKUP_DIR + retention prune
```

`POST /api/ai-pulse/refresh` exposes the same orchestrator over HTTP, guarded by
`Authorization: Bearer $CRON_SECRET` (409 while a run is in progress, `?debug=1`
for raw per-source errors). Every run emits one structured redacted log line and
per-source `ai_cache_meta` rows; sources fail independently and last-known data is
never erased.

Production scheduling ships as systemd units in `deploy/systemd/`
(`goth-refresh.timer` daily 08:00 UTC, `goth-backup.timer` daily 03:30 UTC, both
with randomized delay + flock overlap protection); activation is a Phase 13 step.

## Backups (plan §11.3)

`goth backup` performs an online SQLite backup: `VACUUM INTO` a temp file,
`PRAGMA integrity_check`, atomic rename into `GOTH_BACKUP_DIR`
(default `backups/`), then prunes to the newest `GOTH_BACKUP_KEEP` (default 14)
timestamped snapshots. `goth backup <path>` writes a one-off snapshot without
pruning. Safe against the live WAL-mode writer — no downtime. Restore drill and
off-server copy guidance: `docs/runbook.md` §5.

## Production deploy

See `docs/runbook.md`. Short version: `make release`, copy the tarball + `.sha256`
to the host, run `deploy/deploy.sh` (checksum → extract → migrate → atomic
`/opt/goth/current` swap → restart → health check with auto-rollback). Production
Caddy routing lives in `deploy/caddy/Caddyfile` (Go default, `tech=next` →
Next.js, `/__compare/*` probe routes); validate + exercise it locally with
`deploy/smoke-local.sh`.

## Drills and interop vectors

```bash
./test/drill-standalone.sh   # Phase 12.5g: standalone + failure-drill gate
```

The unlock-cookie cross-stack contract lives in
`../shared/security/unlock-cookie-vectors.json` (regenerate with
`node --import tsx ../scripts/generate-unlock-cookie-vectors.ts` then
`go run ./test/genvectors`), consumed by both stacks' test suites.

## Test & lint

```bash
make test      # unit tests
make test-race # race detector
make vet
make fmt
```

## Release artifact

```bash
make release   # dist/goth-<date>-linux-amd64.tar.gz + .sha256 (self-contained binary)
```

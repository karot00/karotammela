# GOTH Production Runbook — karotammela.fi on Hetzner CX23

Operations reference for Phase 13 and beyond. Contains **no secrets**; every
secret lives only in `/etc/goth/goth.env` on the host (mode 0600, root-owned).

Legend: `[host]` = run on the CX23 over SSH, `[dev]` = run locally in `goth/`.

---

## 1. Production layout (plan §11.2)

| Path | Purpose | Owner |
| --- | --- | --- |
| `/opt/goth/releases/<name>/` | Extracted release artifacts | `goth` |
| `/opt/goth/current` | Symlink to the active release | root (symlink) |
| `/etc/goth/goth.env` | Environment for all units — see `deploy/goth.env.example` | root, 0600 |
| `/var/lib/goth/goth.db` | Production SQLite (WAL mode) | `goth` |
| `/var/lib/goth/backups/` | Local `goth backup` snapshots | `goth` |
| `/etc/caddy/Caddyfile` | From `deploy/caddy/Caddyfile` | root |
| `/etc/systemd/system/goth*.{service,timer}` | From `deploy/systemd/` | root |

Env-name contract: the implemented variables are `GOTH_PORT` and
`GOTH_DB_PATH` (plan §11.4's `GOTH_ADDR`/`GOTH_DATABASE_PATH` are legacy names
that were never implemented). `deploy/goth.env.example` is authoritative.

## 2. First install (once per host)

```bash
# [host] system user + directories
useradd --system --home /var/lib/goth --shell /usr/sbin/nologin goth
mkdir -p /opt/goth/releases /var/lib/goth/backups /etc/goth
chown -R goth:goth /var/lib/goth

# [host] environment (fill in secrets by hand, never over unencrypted channels)
cp deploy/goth.env.example /etc/goth/goth.env
chmod 0640 /etc/goth/goth.env && chown root:goth /etc/goth/goth.env

# [host] units + caddy config
cp deploy/systemd/goth.service deploy/systemd/goth-refresh.{service,timer} \
   deploy/systemd/goth-backup.{service,timer} /etc/systemd/system/
cp deploy/caddy/Caddyfile /etc/caddy/Caddyfile
caddy validate --config /etc/caddy/Caddyfile
systemctl daemon-reload
```

Timers ship **disabled**. Activation is an explicit Phase 13 step (§6).

## 3. Deploy (plan §11.6)

```bash
# [dev] build the release artifact
make release          # -> dist/goth-YYYYMMDD-linux-amd64.tar.gz + .sha256

# [dev] copy artifact + checksum to the host
scp dist/goth-*-linux-amd64.tar.gz{,.sha256} cx23:/tmp/

# [host] deploy (checksum -> extract -> migrate -> atomic symlink swap ->
#        restart -> health check; auto-rollback on failed health check)
sudo ./deploy.sh /tmp/goth-YYYYMMDD-linux-amd64.tar.gz
```

`deploy.sh` holds `/opt/goth/.deploy.lock` (flock) — a concurrent deploy is
rejected, never queued. Old releases stay in `/opt/goth/releases/` for
rollback; prune manually when confident.

Verify after deploy:

```bash
# [host]
systemctl status goth.service
curl -s http://127.0.0.1:8080/api/ping        # {"stack":"go","ms":0}
curl -s https://karotammela.fi/__compare/go/ping   # after DNS cutover
```

## 4. Migrations

- Migrations are embedded in the binary and idempotent (`CREATE ... IF NOT
  EXISTS`); they run automatically at boot (`db.Open`) and explicitly via
  `goth migrate`.
- `deploy.sh` runs `goth migrate` with the **new** binary *before* switching
  the symlink, so a failing migration never takes traffic.
- Manual run (the env file is root:goth 0640 so goth can read it):
  `[host] sudo -u goth bash -c 'set -a; source /etc/goth/goth.env; set +a; /opt/goth/current/goth migrate'`

## 5. Backups (plan §11.3)

- `goth backup` = SQLite online backup: `VACUUM INTO` a temp file → `PRAGMA
  integrity_check` → atomic rename into `GOTH_BACKUP_DIR` → prune to the
  newest `GOTH_BACKUP_KEEP` (default 14). Safe with the live WAL writer; no
  downtime.
- One-off snapshot to an explicit path (no pruning):
  `goth backup /var/lib/goth/backups/pre-deploy-manual.db`
- Automatic: `goth-backup.timer` daily at 03:30 UTC (offset from the 08:00
  UTC refresh).
- **Off-server copy (owner task):** sync `/var/lib/goth/backups/` to a
  Hetzner Storage Box or S3 with restic/rclone + encryption. Local snapshots
  alone do not survive host loss.

### Restore drill (do this at least once before cutover)

```bash
# [host]
systemctl stop goth.service
sudo -u goth cp /var/lib/goth/backups/goth-YYYYMMDD-HHMMSS.db /var/lib/goth/goth.db.restore
sudo -u goth sqlite3 /var/lib/goth/goth.db.restore 'PRAGMA integrity_check;'   # must print ok
sudo -u goth mv /var/lib/goth/goth.db /var/lib/goth/goth.db.broken
sudo -u goth mv /var/lib/goth/goth.db.restore /var/lib/goth/goth.db
rm -f /var/lib/goth/goth.db-wal /var/lib/goth/goth.db-shm   # stale WAL of the old file
systemctl start goth.service
curl -s http://127.0.0.1:8080/api/ping
```

## 6. Timers (refresh + backup)

```bash
# [host] activate (Phase 13 step, not before)
systemctl enable --now goth-refresh.timer goth-backup.timer

# [host] observe
systemctl list-timers 'goth-*'
journalctl -u goth-refresh.service -n 50
journalctl -u goth-backup.service -n 50
```

Exit-status contract: both oneshot services exit 0 only on full success; any
failure marks the unit failed (that *is* the alert signal). Overlap
protection: `flock -n` on `/var/lib/goth/.refresh.lock` / `.backup.lock`.

Manual runs: `systemctl start goth-refresh.service` (or `goth-backup.service`).

## 7. Rollback

Application rollback (bad release, healthy DB):

```bash
# [host] point current at the previous release and restart
ls -1 /opt/goth/releases/
ln -sfn /opt/goth/releases/<previous> /opt/goth/current.tmp
mv -Tf /opt/goth/current.tmp /opt/goth/current
systemctl restart goth.service
curl -s http://127.0.0.1:8080/api/ping
```

Notes:
- `deploy.sh` performs this automatically when the post-deploy health check
  fails.
- Migrations are additive (`IF NOT EXISTS`); an older binary runs fine
  against a newer schema. If a future migration ever becomes destructive, it
  must ship with its own down-path in the release notes before deploy.

Traffic rollback (Go build misbehaving, Next.js still running):
- Emergency: edit `/etc/caddy/Caddyfile` map default from `localhost:8080` to
  `localhost:3000`, then `systemctl reload caddy`. This flips the *default*
  stack without touching user cookies.

DNS rollback: restore the previous A/AAAA records (kept documented at
cutover time). Low TTL (§8) bounds the exposure window.

## 8. DNS cutover + TLS (owner tasks, sequence)

1. Lower `karotammela.fi` TTL to 300s ≥24h before cutover.
2. Confirm the host is healthy locally: `curl -s http://127.0.0.1:8080/api/ping`
   and `curl -sk -H 'Host: karotammela.fi' http://127.0.0.1/` through Caddy.
3. Point A (and AAAA if applicable) at the CX23 IP.
4. Caddy obtains the certificate automatically via ACME on first request —
   watch `journalctl -u caddy -f` for issuance.
5. Verify: `curl -sI https://karotammela.fi` (200, HSTS header),
   `/__compare/go/ping` and `/__compare/next/ping` both respond, FI/EN pages
   render, `tech=next` cookie routes to Next.js.
6. After a stable week: raise HSTS max-age in the Caddyfile, restore TTL.

## 9. Health checks & monitoring

| Check | Command / URL | Healthy |
| --- | --- | --- |
| Go app direct | `curl -s http://127.0.0.1:8080/api/ping` | `{"stack":"go",...}` |
| Through Caddy | `https://karotammela.fi/__compare/go/ping` | `{"stack":"go",...}` |
| Next.js upstream | `https://karotammela.fi/__compare/next/ping` | `{"stack":"next",...}` |
| Refresh ran | `systemctl status goth-refresh.service` | `status=0` recently |
| Backup ran | newest file in `/var/lib/goth/backups/` < 25h old | yes |
| Disk | `df -h /var/lib` | < 80 % |

Minimum viable alerting: an external uptime probe on
`https://karotammela.fi/__compare/go/ping` + `OnFailure=` hooks or a daily
`systemctl --failed` glance. Logs: `journalctl -u goth.service -f`.

## 10. Disaster recovery (host loss)

1. Provision a replacement CX23 (or restore the Hetzner snapshot backup).
2. §2 First install (secrets re-entered from the password manager — this is
   why they must exist outside the host too).
3. Deploy the latest release artifact (§3).
4. Stop goth.service; restore the newest **off-server** backup as
   `/var/lib/goth/goth.db` (§5 restore drill); start goth.service.
5. Point DNS at the new IP (§8). Caddy re-issues TLS automatically.
6. Re-enable timers (§6) and run one manual `goth refresh` to warm AI Pulse.

RPO = age of the newest off-server backup (≤24h with the daily timer).

## 11. Local staging smoke test

`deploy/smoke-local.sh` validates `deploy/caddy/Caddyfile` and runs the
tech-cookie routing matrix (default→Go, `tech=go`→Go, `tech=next`→Next.js,
probe routes) against a local Caddy on :8090 using `Caddyfile.local`.
Requires the `caddy` binary; Next.js assertions are skipped politely when
:3000 is not running.

## 12. VIP portal (MeetingPackage application)

One runtime flag, `VIP_ENABLED`, controls every VIP route in both stacks.
Disabled (the default), all VIP routes return indistinguishable 404s and no
navigation shows a link; the Next.js dashboard link disappears on its next
request without a redeploy (plan §5).

### 12.1 Generate credentials (once, before first enable)

Credentials are generated offline with `cmd/viphash` and never committed.
Prefer piping the access code so it never appears in the process list:

```bash
# [dev] prints VIP_PASSWORD_HASH (Argon2id) and VIP_COOKIE_SECRET (32 random bytes)
printf '%s' "$ACCESS_CODE" | go run ./cmd/viphash
```

Paste both values into `/etc/goth/goth.env`. The access code itself is stored
nowhere — only its hash.

### 12.2 Enable

```bash
# [host]
sudoedit /etc/goth/goth.env      # VIP_ENABLED=true (+ hash, secret, optional VIP_CV_PATH)
sudo systemctl restart goth.service
curl -s https://karotammela.fi/api/vip/status    # {"enabled":true,"url":...}
```

The notification form is protected by three independent limits: 5 accepted
notification attempts per source IP per hour, 3 per normalized email address per
hour, and 30 accepted notifications globally per hour across the instance. The
global cap is deliberately generous for a single recruiter while preventing a
distributed bot from turning the form into an email flood. Honeypot, malformed,
and invalid-address submissions do not send mail. A 429 response includes
`Retry-After`; the cap resets automatically after its one-hour window.

Provision the CV file separately from the release artifact when used, e.g.
`/var/lib/goth/private/Karo_Tammela_CV_2026_08_en.pdf` readable by the goth
user; leave `VIP_CV_PATH` unset to hide the download entirely.

### 12.3 Disable (kill switch)

```bash
# [host]
sudoedit /etc/goth/goth.env      # VIP_ENABLED=false
sudo systemctl restart goth.service
curl -s https://karotammela.fi/api/vip/status                    # {"enabled":false}
curl -s -o /dev/null -w '%{http_code}\n' https://karotammela.fi/en/vip   # 404
```

Keep the Caddy VIP path rules after disabling; removing them makes
re-enablement riskier (plan §5.3).

The production Caddyfile uses `admin off`, so its systemd service cannot use the
usual Caddy admin-API reload operation. After changing the Caddyfile, validate it
and restart the service instead:

```bash
caddy validate --config /etc/caddy/Caddyfile
systemctl restart caddy
```

### 12.4 Rotate credentials

Rotate `VIP_PASSWORD_HASH` and `VIP_COOKIE_SECRET` before any re-enable after
the application window or on suspected code sharing. Rotation invalidates all
outstanding VIP sessions immediately.

```bash
# [dev] generate a fresh hash + secret
printf '%s' "$NEW_ACCESS_CODE" | go run ./cmd/viphash
# [host] replace both values in /etc/goth/goth.env, then
sudo systemctl restart goth.service
```

### 12.5 Observe

Structured logs are redacted by construction (hash prefixes only — never the
access code, full email, cookie, or transcript):

```bash
# [host]
sudo journalctl -u goth.service | grep -E 'vip\.(notify|login|logout)'
```

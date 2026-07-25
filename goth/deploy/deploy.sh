#!/usr/bin/env bash
# GOTH production deploy (plan §11.6). Run ON the CX23 as root:
#
#   sudo ./deploy.sh /path/to/goth-YYYYMMDD-linux-amd64.tar.gz
#
# The tarball is produced by `make release` and MUST have its .sha256 file
# next to it (same directory, same basename + .sha256).
#
# What this does, in order:
#   1. Takes an exclusive lock — concurrent deploys are rejected, not queued.
#   2. Verifies the sha256 checksum of the artifact.
#   3. Extracts into a fresh /opt/goth/releases/<name> directory.
#   4. Runs `goth migrate` with the production env BEFORE switching traffic.
#   5. Atomically swaps the /opt/goth/current symlink (ln -sfn + mv -T).
#   6. Restarts goth.service and health-checks GET /api/ping.
#   7. On a failed health check: rolls the symlink back to the previous
#      release, restarts, and exits non-zero.
#
# No secrets live in this script; everything comes from /etc/goth/goth.env
# (see deploy/goth.env.example). Idempotent: re-deploying the same artifact
# re-extracts and re-points the symlink.
set -euo pipefail

GOTH_ROOT="${GOTH_ROOT:-/opt/goth}"
GOTH_ENV_FILE="${GOTH_ENV_FILE:-/etc/goth/goth.env}"
GOTH_USER="${GOTH_USER:-goth}"
SERVICE="${GOTH_SERVICE:-goth.service}"
HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:8080/api/ping}"
HEALTH_TRIES="${HEALTH_TRIES:-20}"
LOCK_FILE="$GOTH_ROOT/.deploy.lock"

log()  { printf '[deploy] %s\n' "$*"; }
fail() { printf '[deploy] ERROR: %s\n' "$*" >&2; exit 1; }

[ "$(id -u)" -eq 0 ] || fail "run as root (systemctl + $GOTH_ROOT writes)"
[ $# -eq 1 ] || fail "usage: deploy.sh <goth-YYYYMMDD-linux-amd64.tar.gz>"

TARBALL=$(readlink -f "$1") || fail "artifact not found: $1"
[ -f "$TARBALL" ] || fail "artifact not found: $TARBALL"
SHA_FILE="$TARBALL.sha256"
[ -f "$SHA_FILE" ] || fail "checksum file missing: $SHA_FILE (ship it with the artifact)"
[ -f "$GOTH_ENV_FILE" ] || fail "env file missing: $GOTH_ENV_FILE (see deploy/goth.env.example)"
id "$GOTH_USER" >/dev/null 2>&1 || fail "user $GOTH_USER does not exist"

mkdir -p "$GOTH_ROOT/releases"

# --- 1. Concurrency lock ----------------------------------------------------
exec 9>"$LOCK_FILE"
flock -n 9 || fail "another deploy is already running (lock: $LOCK_FILE)"

# --- 2. Checksum ------------------------------------------------------------
log "verifying checksum"
(cd "$(dirname "$TARBALL")" && sha256sum -c "$(basename "$SHA_FILE")") \
  || fail "sha256 verification FAILED — refusing to deploy"

# --- 3. Extract -------------------------------------------------------------
NAME=$(basename "$TARBALL" .tar.gz)
RELEASE_DIR="$GOTH_ROOT/releases/$NAME"
log "extracting to $RELEASE_DIR"
rm -rf "$RELEASE_DIR.partial"
mkdir -p "$RELEASE_DIR.partial"
tar -xzf "$TARBALL" -C "$RELEASE_DIR.partial" --strip-components=1
[ -x "$RELEASE_DIR.partial/goth" ] || fail "tarball does not contain an executable 'goth' binary"
rm -rf "$RELEASE_DIR"
mv "$RELEASE_DIR.partial" "$RELEASE_DIR"
chown -R "$GOTH_USER:$GOTH_USER" "$RELEASE_DIR"

# --- 4. Migrate before switching traffic ------------------------------------
# /etc/goth/goth.env is root:goth 0640 so the goth service account can read it
# (systemd reads the same file as root via EnvironmentFile). Source it as goth,
# then run the migration as goth so the SQLite files are owned by the account.
log "running migrations with $GOTH_ENV_FILE"
sudo -u "$GOTH_USER" bash -c "set -a; source '$GOTH_ENV_FILE'; set +a; exec '$RELEASE_DIR/goth' migrate" \
  || fail "migration failed — current release untouched, traffic not switched"

# --- 5. Atomic symlink swap --------------------------------------------------
PREVIOUS=""
if [ -L "$GOTH_ROOT/current" ]; then
  PREVIOUS=$(readlink -f "$GOTH_ROOT/current")
fi
log "switching current -> $RELEASE_DIR (previous: ${PREVIOUS:-none})"
ln -sfn "$RELEASE_DIR" "$GOTH_ROOT/current.tmp"
mv -Tf "$GOTH_ROOT/current.tmp" "$GOTH_ROOT/current"

# --- 6. Restart + health check ----------------------------------------------
log "restarting $SERVICE"
systemctl restart "$SERVICE"

healthy=0
for i in $(seq 1 "$HEALTH_TRIES"); do
  sleep 1
  if curl -fsS --max-time 3 "$HEALTH_URL" | grep -q '"stack":"go"'; then
    healthy=1
    break
  fi
done

if [ "$healthy" -eq 1 ]; then
  log "health check OK ($HEALTH_URL)"
  log "deployed $NAME"
  log "old releases in $GOTH_ROOT/releases are kept for rollback; prune manually when confident"
  exit 0
fi

# --- 7. Rollback -------------------------------------------------------------
log "health check FAILED after $HEALTH_TRIES tries"
if [ -n "$PREVIOUS" ] && [ -d "$PREVIOUS" ]; then
  log "rolling back to $PREVIOUS"
  ln -sfn "$PREVIOUS" "$GOTH_ROOT/current.tmp"
  mv -Tf "$GOTH_ROOT/current.tmp" "$GOTH_ROOT/current"
  systemctl restart "$SERVICE"
  fail "deploy failed — rolled back to previous release (check: journalctl -u $SERVICE)"
fi
fail "deploy failed and no previous release to roll back to (check: journalctl -u $SERVICE)"

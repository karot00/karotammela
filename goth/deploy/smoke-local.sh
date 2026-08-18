#!/usr/bin/env bash
# Local staging smoke test for the production Caddy routing (runbook §11).
#
# Proves BEFORE touching the real host:
#   1. deploy/caddy/Caddyfile (production) passes `caddy validate`.
#   2. The tech-cookie map routes correctly (default -> Go, tech=go -> Go,
#      tech=next -> Next.js) using Caddyfile.local on :8090.
#   3. The fixed probe routes /__compare/go/ping and /__compare/next/ping
#      bypass cookie routing.
#
# Requirements: caddy binary, Go toolchain. If nothing listens on :8080 a
# temporary goth server is started with a scratch DB in /tmp. Next.js (:3000)
# assertions are skipped politely when it is not running.
set -u
cd "$(dirname "$0")/.." # goth module root

WORK=/tmp/goth-smoke
CADDY_URL="http://localhost:8090"
PASS=0
FAIL=0
SKIP=0

say()  { printf '\n=== %s ===\n' "$*"; }
ok()   { PASS=$((PASS+1)); printf '  ok: %s\n' "$*"; }
bad()  { FAIL=$((FAIL+1)); printf '  FAIL: %s\n' "$*"; }
skip() { SKIP=$((SKIP+1)); printf '  skip: %s\n' "$*"; }

cleanup() {
  [ -n "${CADDY_PID:-}" ] && kill "$CADDY_PID" 2>/dev/null
  [ -n "${SRV_PID:-}" ] && kill "$SRV_PID" 2>/dev/null
  rm -rf "$WORK"
}
trap cleanup EXIT
rm -rf "$WORK"; mkdir -p "$WORK"

command -v caddy >/dev/null 2>&1 || {
  echo "FATAL: caddy binary not found — install it (https://caddyserver.com/docs/install) and re-run" >&2
  exit 1
}

say "1. validate production Caddyfile"
if caddy validate --config deploy/caddy/Caddyfile >"$WORK/validate.log" 2>&1; then
  ok "deploy/caddy/Caddyfile is valid"
else
  bad "production Caddyfile invalid:"
  cat "$WORK/validate.log"
fi

say "2. backing services"
if ss -tln 2>/dev/null | grep -q ':8090 '; then
  echo "FATAL: port 8090 already in use — stop the stale process first" >&2
  exit 1
fi

if ss -tln 2>/dev/null | grep -q ':8080 '; then
  ok "reusing existing server on :8080"
else
  go build -o "$WORK/goth" ./cmd/server || { echo "build failed" >&2; exit 1; }
  # The scratch server runs VIP-enabled so the VIP precedence assertions below
  # can observe live Go responses (development env: no startup validation).
  ( cd "$WORK" && GOTH_PORT=8080 GOTH_DB_PATH="$WORK/smoke.db" GOTH_ENV=development \
      UNLOCK_COOKIE_SECRET=smoke-secret \
      VIP_ENABLED=true VIP_PASSWORD_HASH=smoke-vip-hash VIP_COOKIE_SECRET=smoke-vip-cookie-secret \
      ./goth >"$WORK/goth.log" 2>&1 ) &
  SRV_PID=$!
  for _ in $(seq 1 50); do
    curl -fsS --max-time 1 http://localhost:8080/api/ping >/dev/null 2>&1 && break
    sleep 0.2
  done
  curl -fsS --max-time 2 http://localhost:8080/api/ping >/dev/null 2>&1 \
    && ok "scratch goth server started on :8080" \
    || { bad "goth server did not become ready"; cat "$WORK/goth.log"; exit 1; }
fi

NEXT_UP=0
if curl -fsS --max-time 3 -o /dev/null http://localhost:3000/ 2>/dev/null; then
  NEXT_UP=1
  ok "Next.js detected on :3000"
else
  skip "Next.js not running on :3000 — tech=next assertions limited"
fi

say "3. start local Caddy (:8090, production routing logic)"
caddy run --config deploy/caddy/Caddyfile.local >"$WORK/caddy.log" 2>&1 &
CADDY_PID=$!
for _ in $(seq 1 50); do
  curl -fsS --max-time 1 "$CADDY_URL/api/ping" >/dev/null 2>&1 && break
  sleep 0.2
done
curl -fsS --max-time 2 "$CADDY_URL/api/ping" >/dev/null 2>&1 \
  && ok "caddy ready on :8090" \
  || { bad "caddy did not become ready"; cat "$WORK/caddy.log"; exit 1; }

say "4. tech-cookie routing matrix"
body=$(curl -fsS --max-time 3 "$CADDY_URL/api/ping" 2>/dev/null)
case "$body" in
  *'"stack":"go"'*) ok "no cookie -> Go (production default)";;
  *) bad "no cookie -> expected stack go, got: $body";;
esac

body=$(curl -fsS --max-time 3 -H 'Cookie: tech=go' "$CADDY_URL/api/ping" 2>/dev/null)
case "$body" in
  *'"stack":"go"'*) ok "tech=go -> Go";;
  *) bad "tech=go -> expected stack go, got: $body";;
esac

body=$(curl -fsS --max-time 3 -H 'Cookie: tech=unknown-value' "$CADDY_URL/api/ping" 2>/dev/null)
case "$body" in
  *'"stack":"go"'*) ok "unknown cookie value -> Go (map default)";;
  *) bad "unknown cookie value -> expected stack go, got: $body";;
esac

if [ "$NEXT_UP" -eq 1 ]; then
  # Next.js's /api/ping mirror is a separate effort (contracts §7); assert
  # routing via the home page instead: Go never emits _next assets.
  body=$(curl -fsS --max-time 5 -H 'Cookie: tech=next' "$CADDY_URL/" 2>/dev/null)
  case "$body" in
    *_next*) ok "tech=next -> Next.js (home page served by :3000)";;
    *) bad "tech=next -> response does not look like Next.js output";;
  esac
else
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 3 -H 'Cookie: tech=next' "$CADDY_URL/api/ping")
  if [ "$code" = "502" ]; then
    ok "tech=next -> routed to :3000 (502 because Next.js is down, as expected)"
  else
    bad "tech=next with Next.js down -> expected 502 from :3000 upstream, got HTTP $code"
  fi
fi

say "4b. VIP path precedence over tech cookie (plan §4.2, threat T9)"
# /api/vip/status exists only in the Go build: receiving its JSON while
# carrying tech=next proves the VIP handle outranks the cookie map.
body=$(curl -fsS --max-time 3 -H 'Cookie: tech=next' "$CADDY_URL/api/vip/status" 2>/dev/null)
case "$body" in
  *'"enabled"'*) ok "/api/vip/status -> Go status JSON even with tech=next";;
  *) bad "/api/vip/status with tech=next -> expected Go status JSON, got: $body";;
esac

if [ -n "${SRV_PID:-}" ]; then
  case "$body" in
    *'"enabled":true'*) ok "/api/vip/status reports enabled:true (scratch server runs VIP-enabled)";;
    *) bad "/api/vip/status -> expected enabled:true, got: $body";;
  esac

  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 3 -H 'Cookie: tech=next' "$CADDY_URL/en/vip")
  if [ "$code" = "200" ]; then
    ok "/en/vip with tech=next -> served by Go (HTTP 200)"
  else
    bad "/en/vip with tech=next -> expected Go HTTP 200, got HTTP $code"
  fi

  loc=$(curl -s -o /dev/null -w '%{redirect_url}' --max-time 3 -H 'Cookie: tech=next' "$CADDY_URL/vip")
  case "$loc" in
    */en/vip*) ok "/vip with tech=next -> redirects to /en/vip via Go";;
    *) bad "/vip with tech=next -> expected redirect to /en/vip, got: $loc";;
  esac
else
  skip "VIP enabled-state assertions (reusing an existing :8080 server; VIP state unknown)"
fi

say "5. comparison probe routes (bypass cookie routing)"
body=$(curl -fsS --max-time 3 -H 'Cookie: tech=next' "$CADDY_URL/__compare/go/ping" 2>/dev/null)
case "$body" in
  *'"stack":"go"'*) ok "/__compare/go/ping -> Go even with tech=next";;
  *) bad "/__compare/go/ping -> expected stack go, got: $body";;
esac

if [ "$NEXT_UP" -eq 1 ]; then
  # Reaches the :3000 upstream regardless of cookie; Next.js's /api/ping
  # mirror may not exist yet, so accept any non-502 upstream response.
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 -H 'Cookie: tech=go' "$CADDY_URL/__compare/next/ping")
  if [ "$code" != "502" ] && [ "$code" != "000" ]; then
    ok "/__compare/next/ping -> reaches :3000 upstream even with tech=go (HTTP $code)"
  else
    bad "/__compare/next/ping -> upstream unreachable (HTTP $code)"
  fi
else
  skip "/__compare/next/ping assertion (Next.js down)"
fi

say "result"
printf 'passed=%d failed=%d skipped=%d\n' "$PASS" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]

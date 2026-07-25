#!/usr/bin/env bash
# Phase 12.5g — standalone and failure-drill gate.
#
# Proves the Go build runs with Next.js stopped and loopback to it denied:
#   A. standalone boot (fresh DB, NEXT_PING_URL dead, no NEXT_DB_PATH)
#   B. refresh happy path + idempotent re-run + HTTP overlap rejection
#   C. each provider failing independently + malformed GitHub markup
#   D. Gemini quota failure → deterministic fallback content, run still ok
#   E. server restart during an active refresh → cached data survives
#   F. dashboard empty/cached states + public routes healthy throughout
#
# Exits non-zero on the first failed assertion. Safe to run repeatedly; all
# state lives under /tmp/goth-drill and is cleaned up on exit.
set -u
cd "$(dirname "$0")/.." # goth module root

WORK=/tmp/goth-drill
PORT=18080
DRILL_ADDR=127.0.0.1:8099
BASE="http://127.0.0.1:$PORT"
SECRET=drill-cron-secret
PASS=0
FAIL=0

say()  { printf '\n=== %s ===\n' "$*"; }
ok()   { PASS=$((PASS+1)); printf '  ok: %s\n' "$*"; }
bad()  { FAIL=$((FAIL+1)); printf '  FAIL: %s\n' "$*"; }
check(){ # check <label> <expected> <actual>
  if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (want [$2], got [$3])"; fi
}

cleanup() {
  [ -n "${SRV_PID:-}" ] && kill "$SRV_PID" 2>/dev/null
  [ -n "${FIX_PID:-}" ] && kill "$FIX_PID" 2>/dev/null
  rm -rf "$WORK" bin/drillserver
}
trap cleanup EXIT
rm -rf "$WORK"; mkdir -p "$WORK"

# Pre-flight: the drill fails fast if the ports are held by anything else
# (stale dev servers have caused false negatives before).
for P in 8099 $PORT; do
  if ss -tln 2>/dev/null | grep -q ":$P "; then
    echo "FATAL: port $P is already in use — stop the stale process first (ss -tlnp | grep $P)" >&2
    exit 1
  fi
done

say "build"
go build -o bin/goth ./cmd/server || { echo "build failed"; exit 1; }
go build -o bin/drillserver ./test/drillserver || { echo "drillserver build failed"; exit 1; }
ok "binaries built"

start_fixtures() { # start_fixtures [extra drillserver args...]
  [ -n "${FIX_PID:-}" ] && kill "$FIX_PID" 2>/dev/null && wait "$FIX_PID" 2>/dev/null
  ./bin/drillserver -addr "$DRILL_ADDR" "$@" > "$WORK/fixtures.log" 2>&1 &
  FIX_PID=$!
  # Readiness: require the DRILL marker, not just any HTTP listener — a stale
  # foreign server on this port must fail fast instead of skewing results.
  for _ in $(seq 1 50); do
    [ "$(curl -s "http://$DRILL_ADDR/healthz" 2>/dev/null)" = "drillserver-ok" ] && return 0
    sleep 0.1
  done
  echo "FATAL: drillserver did not come up on $DRILL_ADDR" >&2
  exit 1
}

start_server() { # start_server <db> [extra env...]
  [ -n "${SRV_PID:-}" ] && kill "$SRV_PID" 2>/dev/null && wait "$SRV_PID" 2>/dev/null
  env GOTH_PORT=$PORT GOTH_DB_PATH="$1" GOTH_ENV=development \
      UNLOCK_COOKIE_SECRET=drill-unlock-secret CRON_SECRET=$SECRET \
      NEXT_PING_URL=http://127.0.0.1:1/api/ping \
      GOTH_HN_BASE_URL="http://$DRILL_ADDR" \
      GOTH_GITHUB_BASE_URL="http://$DRILL_ADDR" \
      GOTH_YAHOO_BASE_URL="http://$DRILL_ADDR" \
      GOOGLE_GENERATIVE_AI_API_KEY=drill-key \
      GOTH_GEMINI_BASE_URL="http://$DRILL_ADDR/gemini" \
      "${@:2}" ./bin/goth > "$WORK/server.log" 2>&1 &
  SRV_PID=$!
  for _ in $(seq 1 50); do curl -sf "$BASE/api/ping" > /dev/null 2>&1 && break; sleep 0.1; done
}

COOKIE=$(UNLOCK_COOKIE_SECRET=drill-unlock-secret go run ./test/mintcookie)

say "A. standalone boot (fresh DB, Next.js loopback denied)"
start_fixtures
start_server "$WORK/a.db"
check "/fi 200" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/fi")"
check "/api/ping stack=go" '{"stack":"go","ms":0}' "$(curl -s "$BASE/api/ping")"
check "next probe returns null (no crash)" '{"stack":"next","ms":null}' "$(curl -s "$BASE/api/ping?target=next")"
check "dashboard 303 without cookie" 303 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/fi/dashboard")"
check "sitemap healthy" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/sitemap.xml")"
check "blog healthy" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/en/blog")"

say "B. refresh happy path + idempotency + HTTP overlap"
BODY=$(curl -s -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh")
echo "  run1: $BODY"
check "run1 ok:true" '"ok":true' "$(echo "$BODY" | grep -o '"ok":true' | head -1)"
check "run1 trends ok" '"trends":{"ok":true' "$(echo "$BODY" | grep -o '"trends":{"ok":true' )"
check "run1 repos ok" '"repos":{"ok":true' "$(echo "$BODY" | grep -o '"repos":{"ok":true')"
check "run1 stocks ok" '"stocks":{"ok":true' "$(echo "$BODY" | grep -o '"stocks":{"ok":true')"
BODY2=$(curl -s -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh")
check "run2 still ok:true (idempotent)" '"ok":true' "$(echo "$BODY2" | grep -o '"ok":true' | head -1)"
STOCKS_N1=$(python3 -c "import sqlite3;print(sqlite3.connect('$WORK/a.db').execute('select count(*) from ai_stocks').fetchone()[0])")
sleep 1
BODY3=$(curl -s -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh")
STOCKS_N2=$(python3 -c "import sqlite3;print(sqlite3.connect('$WORK/a.db').execute('select count(*) from ai_stocks').fetchone()[0])")
check "stocks row count stable across runs" "$STOCKS_N1" "$STOCKS_N2"
check "no auth → 401" 401 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE/api/ai-pulse/refresh")"

# Overlap: slow Yahoo, fire two POSTs — second must 409.
start_fixtures -delay-yahoo 4s
curl -s -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh" > "$WORK/overlap1.json" &
OVER=$!
sleep 1
check "overlap second POST → 409" 409 "$(curl -s -o /dev/null -w '%{http_code}' -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh")"
wait $OVER
check "overlap first run completes ok" '"ok":true' "$(grep -o '"ok":true' "$WORK/overlap1.json" | head -1)"
start_fixtures # back to fast fixtures

say "C. per-provider failure isolation"
for SRC in hn github yahoo; do
  start_fixtures -fail "$SRC"
  BODY=$(curl -s -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh")
  case $SRC in
    hn)     KEY=trends ;;
    github) KEY=repos ;;
    yahoo)  KEY=stocks ;;
  esac
  check "$SRC down → $KEY not ok" '' "$(echo "$BODY" | grep -o "\"$KEY\":{\"ok\":true")"
  OTHERS_OK=$(echo "$BODY" | python3 -c "import json,sys; d=json.load(sys.stdin)['sources']; print(all(v['ok'] for k,v in d.items() if k != '$KEY'))")
  check "$SRC down → other sources still ok" True "$OTHERS_OK"
  check "$SRC down → route still 200" 200 "$(curl -s -o /dev/null -w '%{http_code}' -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh")"
done
start_fixtures -malformed-github
BODY=$(curl -s -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh?debug=1")
check "malformed GH markup → repos not ok" '' "$(echo "$BODY" | grep -o '"repos":{"ok":true')"
check "malformed GH markup → debug error mentions markup" 'markup' "$(echo "$BODY" | grep -o 'markup' | head -1)"
check "malformed GH markup → trends/stocks ok" '"trends":{"ok":true' "$(echo "$BODY" | grep -o '"trends":{"ok":true')"
start_fixtures

say "D. Gemini quota failure → fallback content, sources still ok"
start_fixtures -fail gemini
BODY=$(curl -s -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh")
check "gemini 429 → trends still ok (title fallback)" '"trends":{"ok":true' "$(echo "$BODY" | grep -o '"trends":{"ok":true')"
check "gemini 429 → repos still ok (EN fallback)" '"repos":{"ok":true' "$(echo "$BODY" | grep -o '"repos":{"ok":true')"
start_fixtures

say "E. restart during refresh → cached data survives"
start_fixtures -delay-yahoo 5s
curl -s -X POST -H "Authorization: Bearer $SECRET" "$BASE/api/ai-pulse/refresh" > /dev/null &
OVER=$!
sleep 1
kill -9 "$SRV_PID" 2>/dev/null; wait "$SRV_PID" 2>/dev/null; SRV_PID=
kill "$OVER" 2>/dev/null; wait "$OVER" 2>/dev/null
start_fixtures
start_server "$WORK/a.db" # same DB as before
check "server healthy after kill -9" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/api/ping")"
check "cached stocks survive restart" '"ticker":"NVDA"' "$(curl -s "$BASE/api/ai-pulse/stocks?ticker=NVDA" | grep -o '"ticker":"NVDA"')"
check "cached series non-empty after restart" '"date"' "$(curl -s "$BASE/api/ai-pulse/stocks?ticker=NVDA" | grep -o '"date"' | head -1)"

say "F. dashboard empty/cached states + routes healthy"
start_server "$WORK/empty.db" # fresh empty DB
EMPTY_HTML=$(curl -s --cookie "karot_unlock=$COOKIE" "$BASE/fi/dashboard?view=ai-pulse")
check "empty-DB ai-pulse view 200" 200 "$(curl -s -o /dev/null -w '%{http_code}' --cookie "karot_unlock=$COOKIE" "$BASE/fi/dashboard?view=ai-pulse")"
check "empty-DB shows fi empty-state label" 'Trendidataa ei ole' "$(echo "$EMPTY_HTML" | grep -o 'Trendidataa ei ole' | head -1)"
start_server "$WORK/a.db" # populated DB
CACHED_HTML=$(curl -s --cookie "karot_unlock=$COOKIE" "$BASE/fi/dashboard?view=ai-pulse")
check "cached ai-pulse shows drill trend" 'Drill story one' "$(echo "$CACHED_HTML" | grep -o 'Drill story one' | head -1)"
check "cached ai-pulse shows drill repo" 'meta-llama/llama3' "$(echo "$CACHED_HTML" | grep -o 'meta-llama/llama3' | head -1)"
check "home still healthy" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/fi")"
check "stocks API healthy" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE/api/ai-pulse/stocks?ticker=NVDA")"

say "log redaction spot check"
if grep -Eq 'sess-|Bearer|drill-key|quote' "$WORK/server.log"; then
  bad "server log contains secrets/user content"
  grep -E 'sess-|Bearer|drill-key' "$WORK/server.log" | head -3
else
  ok "server log carries no secrets/user content"
fi
grep -c 'ai-pulse.refresh.complete' "$WORK/server.log" > /dev/null && ok "structured refresh lines present in log"

say "RESULT"
echo "  passed=$PASS failed=$FAIL"
[ "$FAIL" -eq 0 ]

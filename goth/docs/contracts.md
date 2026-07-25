# GOTH Parity Contracts & Stat Semantics

> Phase 7 deliverable. Defines the shared runtime contracts that the Go app and
> the Next.js reference must honor so the Tech Switcher preserves state, and the
> semantics of aggregate statistics plus how the production SQLite is seeded.

## 1. Cookie contracts

| Name | Purpose | Options | Set by |
| --- | --- | --- | --- |
| `tech` | Selects the active stack (`go` or `next`). | `Path=/; Max-Age=31536000; SameSite=Lax`; `Secure` added in production TLS | Tech Switcher widget (both stacks) |
| `karot_unlock` | Signed proof of SENTINEL-7 unlock; grants dashboard access. | `Path=/; HttpOnly; SameSite=Lax; Max-Age=1209600`; `Secure` in production | Go `/api/sentinel` + commit; Next.js equivalent |
| `theme` | Persistent `dark`/`light` theme. | `Path=/; Max-Age=31536000` | Theme toggle (both stacks) |
| `locale` | Optional explicit locale preference. | `Path=/; Max-Age=31536000` | Locale switcher (both stacks) |
| `karot_consent` | Consent decision payload (see §9). | `Path=/; Max-Age=15552000; SameSite=Lax`; `Secure` on HTTPS | Consent banner/modal client JS (both stacks) |

Cross-stack unlock compatibility: both stacks verify the same HMAC-signed
`karot_unlock` payload (`sessionId`, `locale`, `unlockedAt` — camelCase,
matching `JSON.stringify`/`json.Marshal` output in both implementations) with
the shared `UNLOCK_COOKIE_SECRET`. Switching stacks never resets unlock, theme,
locale, or consent because the cookie names and options are identical.

**Value format (12.5f golden-vector contract):**
`base64url(JSON, no padding) + "." + base64url(HMAC-SHA256(secret, payload-part),
no padding)`. JSON keys are emitted in `sessionId, locale, unlockedAt` order
with no HTML escaping (`<`, `>`, `&` stay raw, like `JSON.stringify`).
Validation: HMAC first (constant time, length-checked), then each field must be
present with the right JSON type (string, string, number) — empty strings and
`unlockedAt: 0` are ACCEPTED (reference `typeof` semantics, explicitly aligned
in Phase 12.5f); missing keys, null, and wrong JSON types are rejected. Only
the first two `.`-separated segments are examined; trailing segments are
ignored (reference `split(".")` semantics). Interop is proven by the shared
golden vectors in `shared/security/unlock-cookie-vectors.json`, consumed by
`goth/internal/security/unlock_cookie_golden_test.go` and
`src/lib/security/unlock-cookie.golden.test.ts`.

## 2. Locale behavior

- Supported locales: `en`, `fi`. `fi` is the default.
- All content routes are locale-prefixed: `/{locale}`, `/{locale}/blog`,
  `/{locale}/blog/{slug}`, `/{locale}/privacy`, `/{locale}/dashboard`,
  `/{locale}/postikortti`.
- An unsupported or missing locale segment returns 404; it never silently falls
  back to Finnish. (`resolveLocale` clamps unknown values to the default only
  for the active page render, but route matching requires a valid prefix.)

## 3. Route / query contracts

- Dashboard query state: `?view=overview|projects|tech|blog|changelog|ai-pulse|settings`,
  plus blog `page` and `post` parameters. Switching stacks preserves the current
  path, query string, hash, and locale.
- Blog pagination uses `?page=N`; out-of-range values clamp instead of erroring.
- The Tech Switcher performs a **cross-origin redirect** to the other stack's
  origin, preserving the exact current path + query + hash (no `tech` cookie).
  Go apex is `https://karotammela.fi`; the Next.js build is
  `https://next.karotammela.fi` (Vercel). "Go" → `{GO_ORIGIN}{path}`, "Next.js"
  → `{NEXT_ORIGIN}{path}`.

## 4. Metadata contract

Every page emits identical `<title>`, description, `rel=canonical` (pointing at
the Next.js primary URL), and OG/Twitter tags in both stacks to avoid
duplicate-content penalties during the experiment. `robots.txt` and
`sitemap.xml` are parity-equivalent.

## 5. Comparison endpoints

- Go serves `GET /api/ping` (no-store, identifies stack `go`, no DB/external work).
- Next.js serves an equivalent `GET /api/ping` returning
  `{"status":"ok","stack":"next"}` (no-store, no DB/external work).
- Both endpoints emit CORS headers allow-listing the two comparison origins
  (`GO_ORIGIN`, `NEXT_ORIGIN`): the request `Origin` is reflected only when it
  matches, plus `Vary: Origin` and an `OPTIONS` preflight responder. Unknown
  origins receive no `Access-Control-Allow-Origin`.
- The performance widget runs in the browser and measures **both** origins with
  direct client-side `fetch()` + `performance.now()`, reporting the median of a
  few samples; it shows `n/a` rather than faking `0 ms`. Because each stack is
  hosted natively (no reverse proxy), this measures real end-user TTFB with no
  proxy-induced skew.
- Legacy: the Go `GET /api/ping?target=next` server-side probe and the Caddy
  `/__compare/{go,next}/ping` routes remain available but are no longer used by
  the widget after the 2026-07-25 subdomain architecture change.

## 6. Stat semantics (`db.Stats`)

Aggregates are derived live from the `logs` and `sessions` tables:

- `TotalMessages` — total rows in `logs` (every Sentinel exchange).
- `UnlockedCount` — distinct sessions in `sessions` with `unlocked = 1`.
- `DirectUnlockCount` — `logs` rows whose `user_input` (uppercased/trimmed) equals
  the access code `PROTOCOL_K_2026`.
- `AvgMessagesToUnlock` — average number of `logs` rows per session that reached
  `success = 1`.
- `LatestUnlock` — max `timestamp` among `logs` where `success = 1`, or null.

These counters are display-only and never gate access.

## 7. Production SQLite seeding

- The CX23-local SQLite starts empty. Statistics populate from real Sentinel
  usage; **no external import or synchronization is required** to seed it.
- A fresh `goth.db` is created on first `db.Open` via `migrations/0001_init.sql`.
- Do not backfill from Turso or any other source; production runtime reads are
  local only. If historical counters are ever needed, they are imported by an
  explicit, audited one-off command — never implicitly.

## 8. Draft / future-date policy (fail closed)

- `draft: true` posts are excluded from public lists, direct detail routes, the
  dashboard blog reader, sitemap, metadata, and feeds in **every** environment.
  There is no unauthenticated draft-preview mode.
- Posts with a `publishedAt` in the future are excluded identically.
- `GOTH_ENV` is irrelevant to this filtering: `config.IsProduction()` is true only
  for the literal value `production`; any other (including missing) value keeps the
  build non-production, and draft/future filtering is unconditional regardless.

## 9. Consent contract (`karot_consent`)

Ported from `src/modules/cookie-consent/*`; Go implementation lives in
`internal/security/consent.go` (server-side gate) and
`internal/view/templates/partials/consent.html` (banner + preferences modal,
Alpine). Both stacks honor the same cookie so switching never re-prompts.

- **Value**: `encodeURIComponent(JSON)` of
  `{ "version": 1, "updatedAt": "<ISO-8601 ms UTC>", "categories": { "essential": bool, "functional": bool, "analytics": bool, "marketing": bool } }`.
- **Schema version**: `1`. Older or partially malformed payloads are normalized
  up; values with no `categories` key (or undecodable JSON) are treated as
  *unset* (the default state).
- **Banner gate**: the banner is required exactly while `updatedAt` equals the
  unset sentinel `1970-01-01T00:00:00.000Z`. Any stored decision (accept,
  reject, or saved selection) moves `updatedAt` to the decision time and hides
  the banner. SSR renders the banner only when required, so it never flashes or
  reappears after a choice.
- **Categories**: `essential` is forced `true` on normalize and on every write;
  the four buckets are `essential`, `functional`, `analytics`, `marketing`.
- **Write path**: client-side JS only (neither stack sets the cookie from the
  server). The cookie is **not signed** — it carries display preferences, not
  authorization. Written with `Path=/; Max-Age=15552000` (180 days);
  `SameSite=Lax`; `Secure` appended on HTTPS.
- **Normalization rules** (identical in both stacks): non-boolean category
  values become `false`; a missing/unparseable `updatedAt` becomes the current
  time (banner hidden); `"categories": null` counts as a stored shape.
- **Storage inventory**: the preferences modal lists the same items as
  `config.ts` (`karot_unlock`, `NEXT_LOCALE`, `karot-theme`, Umami cookieless
  analytics). The reference does not localize item purposes/durations, so both
  locales show identical strings.
- **Reopen**: `window.__gothConsent.open()` / `.close()` are exposed on every
  page (including the dashboard) for settings triggers.

## 10. Contact endpoint (`POST /api/contact`)

Ported from `src/app/api/contact/route.ts`; both stacks accept the same JSON
payload and return the same statuses/bodies so the dashboard contact form works
identically after a stack switch.

- **Request**: JSON `{ name, email, message, company?, website? }` — `name`
  2–80 chars, `email` ≤200 + valid (zod `.email()` shape), `message` 10–4000,
  `company` ≤120 optional, `website` ≤1 (honeypot). All fields trimmed before
  validation; lengths measured in characters, not bytes.
- **Responses**: `200 {"ok":true}` · `400 {"error":"Invalid contact payload."}`
  · `429 {"error":"Rate limit exceeded. Retry shortly."}` + `Retry-After`
  seconds · `500 {"error":"Failed to send contact email."}` ·
  `503 {"error":"Contact delivery is not configured."}`.
- **Order of operations**: per-IP rate limit (`contact-ip`, 8/min) is enforced
  before parsing, so even malformed payloads consume quota. Validation precedes
  the honeypot check, so a `website` value longer than one character is a 400
  schema violation rather than a fake success.
- **Honeypot**: a non-empty `website` returns `200 {"ok":true}` without
  delivering anything.
- **Delivery**: Resend REST `POST /emails` (`from`/`to` from
  `CONTACT_FROM_EMAIL`/`CONTACT_TO_EMAIL`, `replyTo` = visitor email, subject
  `karotammela.fi contact: {name}`, text body `Name/Email[/Company]//message`).
  Any missing config value disables delivery (503). Provider response bodies
  are discarded so provider echoes of user input never reach logs.
- **Observability**: the reference's telemetry events are emitted as redacted
  log lines (`contact.submitted hasCompany=…`, `contact.honeypot_triggered`,
  `contact.rate_limited`, `contact.send_failed`) — never names, addresses, or
  message content.

## 11. AI Pulse stocks endpoint (`GET /api/ai-pulse/stocks`)

Ported from `src/app/api/ai-pulse/stocks/route.ts`, reading Go's local cache.

- **Request**: `?ticker=` must be one of the supported `AITickers`.
- **Responses**: `200 {ticker, companyName, data}` (OHLCV points, ascending by
  date) · `400 {"error":"Invalid ticker"}` for a missing/unknown ticker ·
  `429 {"error":"Rate limit exceeded. Retry shortly."}` + `Retry-After` seconds ·
  `500 {"error":"Internal server error"}` on read failure ·
  `503 {"error":"AI Pulse database is not configured."}` when no DB.
- **Order of operations** (Phase 12j hardening beyond the reference): the per-IP
  rate limit (`ai-pulse-stocks`, 60/min) is enforced **before** any other check,
  so even a missing/invalid ticker consumes quota.
- **Observability / redaction**: only static event tags are logged
  (`ai-pulse.stocks.rate_limited`, `ai-pulse.stocks.read_failed`) — never the
  client IP, the ticker, or DB error detail.

## 12. Rate limiting (shared limiter)

All abuse-prone endpoints share one limiter (`internal/security`), a bounded,
distributed-ready store fronted by `EnforceRateLimit(scope, key, limit, window)`.

- **Scopes / limits**: `contact-ip` 8/min, `sentinel-ip` 40/min,
  `sentinel-session` 16/5min, `ai-pulse-stocks` 60/min. Keys are the client IP
  (from `X-Forwarded-For` first hop, then `X-Real-IP`, else `unknown`) except the
  sentinel session bucket, which is keyed by session id.
- **Bounded**: the default in-memory store caps at 4096 buckets, evicting expired
  and then soonest-expiring entries, so a spoofed-IP key flood cannot exhaust
  memory.
- **Distributed-ready**: `security.RateLimitStore` is a pluggable backend;
  `security.SetRateLimitStore` can install a shared (e.g. Redis) store for
  multi-instance deployments. `nil` restores the default.
- **429 shape**: handlers set `Retry-After` (ceil seconds, min 1) and return a
  JSON `{"error": …}` body; the rejection is logged with a static event tag only.

## 13. AI Pulse refresh endpoint (Phase 12.5d)

`POST /api/ai-pulse/refresh` — port of `src/app/api/ai-pulse/refresh/route.ts`.

- **Auth**: `Authorization: Bearer $CRON_SECRET`; 401 `{"error":"Unauthorized"}`
  when unset or mismatched (identical to the reference).
- **Overlap**: an in-process run in progress → 409
  `{"error":"Refresh already in progress."}`.
- **Config**: no DB or no orchestrator → 503 `{"error":"AI Pulse refresh is not
  configured."}`.
- **Run**: 4-minute bounded context; sources (trends/repos/stocks) run
  concurrently with all-settled semantics; each source persists independently
  (upserts are `ON CONFLICT DO NOTHING` → idempotent re-runs); per-source
  `ai_cache_meta` is written with a redacted detail string.
- **Response 200**: `{ok:true, ranAt, sources:{trends:{ok,inserted,skippedDup,
  windowHours,error}, repos:{ok,inserted,error}, stocks:{ok,inserted,error}}}`.
  `ranAt` is a Node-Date-style ISO string. `error` is `null` on success,
  `"Fetch failed"`/`"DB insert failed"` on failure; `?debug=1` returns the raw
  error detail instead.
- **Logs**: exactly one structured redacted line per run
  (`ai-pulse.refresh.complete` with counts/ok flags/failed kinds) — never raw
  errors, query text, or provider payloads.
- **CLI**: `goth refresh` runs the same orchestrator directly (the systemd
  timer path); exit 0 when all sources succeeded, 1 otherwise.

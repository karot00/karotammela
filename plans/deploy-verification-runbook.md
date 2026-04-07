# Deploy Verification Runbook

Use this runbook for every production deployment to `karotammela.fi`.

## 1) Pre-Deploy Gate (local or CI)

Required checks:

1. `npm run lint`
2. `npm run test`
3. `npm run build`

Do not deploy if any gate fails.

## 2) Environment Validation (Vercel)

Verify these env vars exist and are non-empty:

- `GOOGLE_GENERATIVE_AI_API_KEY`
- `AI_MODEL`
- `TURSO_DATABASE_URL`
- `TURSO_AUTH_TOKEN`
- `UNLOCK_COOKIE_SECRET`
- `APP_URL` (`https://karotammela.fi`)
- `RESEND_API_KEY`
- `CONTACT_FROM_EMAIL`
- `CONTACT_TO_EMAIL`
- `NEXT_PUBLIC_DEFAULT_LOCALE` (`fi`)

Optional:

- `NEXT_PUBLIC_UMAMI_WEBSITE_ID`
- `NEXT_PUBLIC_UMAMI_SCRIPT_URL`
- `NEXT_PUBLIC_HERO_VIDEO_ENABLED`

## 3) Production Smoke Checklist

Run immediately after deploy:

1. Open `/` and confirm redirect to `/fi`.
2. Switch locale `fi <-> en` and confirm copy updates.
3. Submit 1 Sentinel turn and verify:
   - response arrives
   - level updates
   - no UI error
4. Complete unlock flow and verify:
   - dashboard opens
   - refresh keeps access (unlock cookie works)
5. Verify stats render on dashboard (Live Pulse values present).
6. Submit contact form and verify API success (and message received at destination inbox).
7. If Umami env is set, verify pageview events are visible.

## 4) API Health Quick Checks

Expected behavior:

- `POST /api/sentinel`
  - `200` valid request
  - `400` invalid payload
  - `429` when rate-limited
- `GET /api/stats?locale=fi`
  - `200` with numeric totals
  - `503` only if Turso is missing
- `POST /api/contact`
  - `200` valid request
  - `400` invalid payload
  - `429` when rate-limited
  - `503` if Resend env config missing

## 5) Incident Fallback Procedures

### A) AI Provider Outage (Gemini)

Symptoms:
- Sentinel returns `500` consistently
- elevated latency/timeouts on `/api/sentinel`

Actions:
1. Verify `GOOGLE_GENERATIVE_AI_API_KEY` and `AI_MODEL` in Vercel.
2. Switch `AI_MODEL` to a known stable fallback model.
3. Redeploy and re-run Sentinel smoke check.
4. If still failing, show temporary maintenance notice in UI and disable Sentinel interaction until provider recovers.

### B) Turso/DB Outage

Symptoms:
- dashboard stats missing or `/api/stats` fails
- logs/sessions persistence degraded

Actions:
1. Verify `TURSO_DATABASE_URL` and `TURSO_AUTH_TOKEN`.
2. Check Turso service status.
3. Keep app online: dashboard should continue with fallback/empty values.
4. Re-run migrations after recovery and verify new logs appear.

### C) Resend Outage

Symptoms:
- contact API returns `500`
- delivery failures in Resend dashboard

Actions:
1. Verify `RESEND_API_KEY`, sender domain verification, and `CONTACT_FROM_EMAIL`.
2. Check Resend service status.
3. Temporarily route contact CTA to direct `mailto:info@karotammela.fi` if prolonged outage.
4. After recovery, send a live test submission and confirm inbox delivery.

## 6) Rollback Decision Rule

Rollback if any of these occur after deploy:

- Unlock flow is broken for valid users.
- Sentinel route error rate remains high for more than 10 minutes.
- Contact flow fails without workaround.

After rollback:

1. Post incident note (what failed, impact, mitigation).
2. Open follow-up fix task with reproduction steps.
3. Re-release only after smoke checklist passes again.

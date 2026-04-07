# Post-Launch Enhancements

This phase starts after core Phase 7 delivery is complete.

## Objectives
- Harden production operations.
- Improve observability and anti-abuse controls.
- Add quality automation for continuous confidence.

## Recommended Sequence

1. **Rate limiting and abuse protection**
   - Add per-IP/session rate limits to `POST /api/sentinel` and `POST /api/contact`.
   - Cap Sentinel turns per session and enforce cooldown on repeated failures.

2. **Telemetry and event model**
   - Add structured events for:
     - sentinel turn submitted
     - unlock success
     - dashboard opened
     - contact submitted
   - Keep PII minimal in analytics payloads.

3. **Parser safety tightening**
   - Limit max level delta per turn to reduce model drift.
   - Add deterministic fallback behavior when `[LEVEL:XX]` is malformed.

4. **Dashboard performance pass**
   - Introduce short TTL server-side cache for aggregate stats.
   - Keep unlock guard and user-specific checks dynamic.

5. **Test suite expansion**
   - Unit tests for sentinel parser and cookie verification.
   - API integration tests for sentinel, stats, and contact routes.
   - Basic e2e smoke for locale routing and unlock flow.

6. **Deploy verification runbook**
   - Record smoke checklist for each production deployment.
   - Include fallback steps for AI provider, Turso, and Resend outages.

## Status

- [x] 1) Rate limiting and abuse protection
- [x] 2) Telemetry and event model
- [x] 3) Parser safety tightening
- [x] 4) Dashboard performance pass
- [x] 5) Test suite expansion (unit baseline)
- [x] 6) Deploy verification runbook

Runbook file:

- `plans/deploy-verification-runbook.md`

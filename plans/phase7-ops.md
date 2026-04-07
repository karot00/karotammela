# Phase 7 Ops Notes

## Media Integration
- Hero section supports video and image fallback:
  - Poster path: `/public/media/hero-poster.svg`
  - Optional video paths: `/public/media/hero-loop.webm`, `/public/media/hero-loop.mp4`
- Video playback is controlled with `NEXT_PUBLIC_HERO_VIDEO_ENABLED`:
  - `0` => image fallback only
  - `1` => video enabled for non-reduced-motion users

## Umami Analytics
- Script is injected only when both env vars are set:
  - `NEXT_PUBLIC_UMAMI_WEBSITE_ID`
  - `NEXT_PUBLIC_UMAMI_SCRIPT_URL`
- Integration point: `src/components/umami-script.tsx`

## Resend Contact Flow
- API route: `POST /api/contact`
- Requires:
  - `RESEND_API_KEY`
  - `CONTACT_FROM_EMAIL`
  - `CONTACT_TO_EMAIL`
- Includes basic anti-spam honeypot field (`website`).

## Recommended Production Setup
1. Configure DNS for `karotammela.fi` in Vercel.
2. Configure Resend domain and verify DKIM/SPF.
3. Point `CONTACT_FROM_EMAIL` to verified sender identity.
4. Set `APP_URL=https://karotammela.fi`.
5. Enable Umami script URL + website ID in production env.

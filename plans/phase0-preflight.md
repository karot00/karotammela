# Phase 0 Preflight - Completed (2026-04-07)

## Compatibility Checks

- Next.js 16 stable line selected: `next@16.2.2`
- Node runtime requirement validated: `>=20.9.0`
- Local Node version detected: `v25.8.0` (compatible)
- AI SDK stable production line selected: `ai@5.0.169` (`ai-v5` dist-tag)
- Drizzle + Turso recommended package set validated via docs:
  - `drizzle-orm`
  - `@libsql/client`
  - `drizzle-kit`
- i18n strategy validated for App Router with `next-intl@4.9.0`
- Styling stack validated:
  - `tailwindcss@4.2.2`
  - `@tailwindcss/postcss@4.2.2`
  - `framer-motion@12.38.0`

## Final Dependency Matrix (Exact Versions)

### Core App
- `next@16.2.2`
- `react@19.2.4`
- `react-dom@19.2.4`

### TypeScript + Lint
- `typescript@6.0.2`
- `@types/node@25.5.2`
- `eslint-config-next@16.2.2`

### i18n
- `next-intl@4.9.0`

### AI
- `ai@5.0.169`
- `@ai-sdk/google@2.0.67`
- Default model: `gemini-3.1-flash-lite-preview`

### Database + Validation
- `drizzle-orm@0.45.2`
- `drizzle-kit@0.31.10`
- `@libsql/client@0.17.2`
- `zod@4.3.6`
- `drizzle-zod@0.8.3`

### UI + Motion
- `tailwindcss@4.2.2`
- `@tailwindcss/postcss@4.2.2`
- `framer-motion@12.38.0`
- `lucide-react@1.7.0`
- `class-variance-authority@0.7.1`
- `clsx@2.1.1`
- `tailwind-merge@3.5.0`

### Tooling CLI
- `shadcn@4.1.2` (invoked via `npx`)

## Repository State and Bootstrap Strategy

- Local repository state: initialized git repo with no commits.
- GitHub repository state: `karot00/karotammela` exists and is empty.
- Local `git remote` currently not configured.
- Bootstrap strategy:
  1. Scaffold Next.js 16 app directly in repo root.
  2. Keep planning docs, but isolate runtime app code under the new Next.js structure.
  3. Add strict `.gitignore` before first commit (`.env*`, local tooling artifacts, build output).
  4. Keep `.env.example` tracked as env contract.
  5. Run `npm run lint` and `npm run build` before first push.
  6. Configure `origin` to `karot00/karotammela` and push initial scaffold commit.

## Security Fixes Applied in Phase 0

- Replaced hardcoded token in `.kilo/kilo.json` with template variable:
  - `GITHUB_PERSONAL_ACCESS_TOKEN: "{{GITHUB_PERSONAL_ACCESS_TOKEN}}"`
- Added env contract file `.env.example` with all required runtime variables.

## Environment Contract

Tracked in `.env.example`:

- `GOOGLE_GENERATIVE_AI_API_KEY`
- `AI_MODEL` (default `gemini-3.1-flash-lite-preview`)
- `RESEND_API_KEY`
- `CONTACT_FROM_EMAIL`
- `CONTACT_TO_EMAIL`
- `NEXT_PUBLIC_UMAMI_WEBSITE_ID`
- `NEXT_PUBLIC_UMAMI_SCRIPT_URL`
- `TURSO_DATABASE_URL`
- `TURSO_AUTH_TOKEN`
- `APP_URL`
- `UNLOCK_COOKIE_SECRET`
- `NEXT_PUBLIC_DEFAULT_LOCALE` (default `fi`)

# karotammela.fi

Professional portfolio platform for Karo Tammela, focused on agentic AI orchestration, practical product delivery, and modern web engineering.

## What This Project Is

This repository powers the live site and protected dashboard experience at `karotammela.fi`.

It combines a public-facing portfolio with an interactive AI gate and a localized backstage dashboard that presents projects, technology stack, and long-form blog content.

## Current Product Scope

- Public landing page with brand narrative and conversion-focused hero section
- AI Bouncer (Sentinel) challenge flow with trust/compromise progression
- Secure dashboard access flow with signed unlock cookie validation
- Privacy Policy page (localized)
- Native sharing capabilities on home and blog pages
- Dashboard views:
  - Overview
  - Projects
  - Technologies (categorized, localized card layout)
  - Blog (markdown-backed, paginated, localized)
  - Changelog (curated, bilingual summary of shipped improvements)
  - Settings (language, theme, and cookie preferences)
- Contact form integration for inbound business inquiries
- Bilingual UX (`fi`, `en`) using structured locale messages
- Custom SVG favicon

## Technology Stack

- Framework: Next.js 16 (App Router), React 19, TypeScript
- UI: Tailwind CSS v4, shadcn/ui primitives, Lucide icons, Framer Motion
- Internationalization: next-intl
- AI: Vercel AI SDK + Google provider
- Data: Turso (LibSQL) + Drizzle ORM
- Content: Markdown blog pipeline (`gray-matter`, `marked`, `sanitize-html`)
- Messaging: Resend (contact delivery)
- Deployment: Vercel

## Engineering Characteristics

- Token-based theme system with dark/light compatibility
- Server-side locale routing and translation contracts
- Typed data boundaries across dashboard and content flows
- Security-oriented unlock/session handling
- Structured, maintainable component architecture for iterative delivery

## Local Development

```bash
npm install
npm run dev
```

Open `http://localhost:3000`.

### Core Scripts

```bash
npm run lint
npm run test
npm run build
```

## Environment Variables

Configure required values in `.env.local` for local runtime:

- `GOOGLE_GENERATIVE_AI_API_KEY`
- `AI_MODEL`
- `TURSO_DATABASE_URL`
- `TURSO_AUTH_TOKEN`
- `UNLOCK_COOKIE_SECRET`
- `APP_URL`
- `RESEND_API_KEY`
- `CONTACT_FROM_EMAIL`
- `CONTACT_TO_EMAIL`
- `NEXT_PUBLIC_DEFAULT_LOCALE`

Optional analytics:

- `NEXT_PUBLIC_UMAMI_WEBSITE_ID`
- `NEXT_PUBLIC_UMAMI_SCRIPT_URL`

> Note: Sensitive files, local databases (`*.sqlite`, `*.db`), environment files (`.env*`), and development-specific instructions (`AGENTS.md`, `CLAUDE.md`, `.kilo/`, `plans/`) are ignored by Git to maintain repository security.

## Updating the Changelog

The dashboard changelog is curated by hand in two locale files:

- `src/content/changelog/en.json`
- `src/content/changelog/fi.json`

To add a release, prepend a new entry to the `releases` array in both files. Each entry takes the shape:

```jsonc
{
  "version": "v0.8",
  "date": "2026-05-01",
  "title": "Short release theme",
  "changes": [
    { "type": "added", "text": "User-facing description of what shipped." },
    { "type": "changed", "text": "…" },
    { "type": "fixed", "text": "…" },
    { "type": "removed", "text": "…" },
  ],
}
```

`type` must be one of `added | changed | fixed | removed`. Keep the English and Finnish files in lockstep so the locales stay aligned.

## Repository Purpose

This codebase is both a production portfolio and an execution environment for testing high-velocity, agent-assisted software workflows with business-grade output quality.

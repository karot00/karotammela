# karotammela.fi — Agentic AI Architect Portfolio

This is the central hub for my work, experiments, and thoughts on the transition from traditional software development to **Agentic AI Workflows**.

## 🚀 The Vision
I am a self-taught developer and "AI-native" architect. I believe that the future of coding isn't just about writing syntax—it's about orchestrating autonomous agents to build scalable solutions at 10x speed. This site serves as a live testbed for these theories.

## 🛠 The Stack
This project is built with a high-performance "Edge-first" stack:
- **Framework:** [Next.js 16+](https://nextjs.org/) (App Router)
- **Database:** [Turso](https://turso.tech/) (LibSQL at the edge)
- **AI Orchestration:** [Vercel AI SDK](https://sdk.vercel.ai/)
- **Styling:** Tailwind CSS & [shadcn/ui](https://ui.shadcn.com/)
- **Animations:** Framer Motion
- **Deployment:** [Vercel](https://vercel.com/)

## Stack Versions
- next `16.2.2`
- react `19.2.4`
- react-dom `19.2.4`
- tailwindcss `4.2.2`
- @tailwindcss/postcss `4.2.2`
- shadcn `4.1.2`
- framer-motion `12.38.0`
- next-intl `4.9.0`
- ai `5.0.169`
- @ai-sdk/google `2.0.67`
- drizzle-orm `0.45.2`
- @libsql/client `0.17.2`
- drizzle-kit `0.31.10`
- zod `4.3.6`
- drizzle-zod `0.8.3`
- resend `6.10.0`

## Environment Setup (`.env.local`)

Copy `.env.example` to `.env.local` and configure values:

- `GOOGLE_GENERATIVE_AI_API_KEY` - required for Sentinel AI route.
- `AI_MODEL` - defaults to `gemini-3.1-flash-lite-preview`.
- `TURSO_DATABASE_URL` and `TURSO_AUTH_TOKEN` - required for stats/log persistence.
- `UNLOCK_COOKIE_SECRET` - required for secure dashboard unlock cookies.
- `APP_URL` - local: `http://localhost:3000`, production: `https://karotammela.fi`.
- `RESEND_API_KEY`, `CONTACT_FROM_EMAIL`, `CONTACT_TO_EMAIL` - required for `POST /api/contact`.
- `NEXT_PUBLIC_DEFAULT_LOCALE` - recommended `fi`.
- `NEXT_PUBLIC_HERO_VIDEO_ENABLED` - `0` (image fallback only) or `1` (enable hero video).

Optional analytics:

- `NEXT_PUBLIC_UMAMI_WEBSITE_ID`
- `NEXT_PUBLIC_UMAMI_SCRIPT_URL`

If Umami variables are missing, analytics script is not injected and the app works normally.

## Testing

- Run unit tests: `npm run test`
- Run lint checks: `npm run lint`
- Build validation: `npm run build`
- Current unit tests:
  - `src/lib/ai/sentinel.test.ts`
  - `src/lib/security/unlock-cookie.test.ts`

## Operations Runbook

- Deployment verification and incident fallback procedures are documented in:
  - `plans/deploy-verification-runbook.md`

## Resend + Cloudflare Notes

- Resend handles **outbound sending** for `info@karotammela.fi`.
- Cloudflare Email Routing can forward **inbound mail** from `info@karotammela.fi` to your personal inbox.
- You do not need a separate mailbox host to use this contact flow.

## 🤖 My Agentic Toolkit
To achieve maximum velocity, I leverage the best of the 2026 AI ecosystem:
- **Primary IDE:** [Kilo Code](https://kilo.dev/) (Optimized for agentic workflows)
- **Favorite LLM:** **GPT Codex 5.3** (My go-to for complex architecture and logic)
- **Secondary Models:** **Opus 4.6** for deep reasoning and various lighter models for specialized sub-tasks to keep the workflow agile and cost-effective.

## 🔒 Key Feature: The Sentinel Challenge
To demonstrate agentic reasoning and prompt security, I've implemented **Sentinel-7**, a high-status security agent guarding the "inner vault" of this site.
- **Challenge:** Use social engineering or technical reasoning to bypass its defensive layers.
- **Technology:** Stateful AI interactions with a real-time "Compromise Meter" backed by Turso.

## 🧠 Built with Agents
True to the mission, this entire repository was architected and built in approximately **3 hours** using Agentic AI coding workflows. 

Traditional development would have taken days. I use this methodology to lower costs and increase speed for both my own SaaS products and my consultancy clients.

---

### Connect with me
- **LinkedIn:** [Your Link Here]
- **X (Twitter):** [Your Link Here]
- **Website:** [karotammela.fi](https://karotammela.fi)

*"The best way to predict the future is to build it with agents."*

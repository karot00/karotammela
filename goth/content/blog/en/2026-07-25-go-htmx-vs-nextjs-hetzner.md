---
title: "Go + HTMX vs. Next.js: I built a second production site on Hetzner"
description: "Why I built an identical Go+HTMX implementation alongside my Next.js site, and what I learned about setting up a VPS on Hetzner and running SQLite on my own server."
publishedAt: "2026-07-25"
slug: "go-htmx-vs-nextjs-hetzner"
draft: false
tags: ["coding", "go", "htmx", "devops", "sqlite"]
---

The site you're reading right now is built with the Go+HTMX stack and hosted on a Hetzner cloud server. But if you flip the letter in the bottom-right corner to the other view, you can take a look at the second implementation. It's made with Next.js and hosted on Vercel. The database is Turso. Both are the same site — same landing page, same AI bouncer, same dashboard — but two completely different technology choices in one package. There are small differences in the site's design in some implementations, but nothing significant or affecting the site's performance.

### Why I set out to rebuild the entire site from scratch

Three reasons clicked for me at the same time. First, I was curious to find out what Go+HTMX is actually capable of compared to Next.js, with which I've built most of my projects. I wanted to build the same web application twice and measure the difference live, from the user's browser. Second, I wanted to learn how to set up a VPS on Hetzner from zero: SSH, firewall, systemd, Caddy, and Let's Encrypt by hand, no managed platform doing the work for me. Third, I wanted to try what SQLite actually feels like in production on my own disk instead of leaning on Turso — WAL mode, migrations, backups, and restore drills, with no third-party control panel in the way.

And as you've probably already gathered from the content of my site, everything is built using agent-based coding. My tool of choice is Kilo Code. For this project I used the language models **GPT-5.6 Sol**, **Sonnet 5**, **Kimi K3**, and **Hy3**. I may have wandered off at some point and used **Gemini 3.6 Flash**. Alongside that, I discussed different implementation approaches with Gemini as they came up during the project's progress, and validated both the plans and the implementations with its help as well.

A side note on the language models. During this project I really grew fond of Tencent's Hy3 model. I also concluded that Kimi K3 gets the job done, but it rambles and its implementations take absurdly long compared to, say, Sonnet 5 or GPT-5.6.

### 13 phases, one binary

The scope suddenly grew, as often happens with agent-based coding, into the entire public site: the landing page, the SENTINEL-7 bouncer, the blog, privacy, and the full locked dashboard with all of its views (overview, projects, tech, blog reader, changelog, AI Pulse, settings). The one deliberate cut was the postcard generator — it stayed Next.js-exclusive, because porting it wouldn't have brought any educational added value to the comparison. I built the implementation phase by phase (13 phases + one extra standalone gate), and after each one I logged what was done and what broke. The end result is a single self-contained Go binary — templates, static assets, media, and migrations are all embedded — `go build` and that's it, no separate `node_modules` dependency at runtime.

### The hard part: keeping two stacks in sync

The easy part was copying the UI — html/template + Alpine.js + Tailwind mirror React components more directly than I expected. The hard part was making sure two completely different languages produce *bit-identical* output in the critical spots. The unlock cookie is HMAC-signed with a shared secret, and it has to validate identically no matter which stack created it — you can switch stacks mid-session and the user shouldn't notice a thing. I proved this with Node↔Go golden-vector tests, not by eyeballing it. The SENTINEL-7 bouncer's level and unlock logic is ported line by line, and the stats semantics (how `unlockedCount` or `avgMessagesToUnlock` are computed) went through its own separate correction pass once I noticed the Go version was counting sessions where Next.js counts rows.

The biggest piece of proof was built only at the very end: a standalone drill that runs 40 assertions and forces Go to operate completely without Next.js — the Next.js loopback is deliberately killed, and the drill checks that AI Pulse's HN/GitHub/Yahoo feeds, the Gemini summaries, a server restart mid-refresh, and every public route all survive it cleanly. This now runs on every single deploy in CI before anything is pushed to production.

### A VPS on Hetzner and SQLite in production

The Hetzner setup: a CX23 cloud server in Helsinki, Ubuntu, key-based SSH, and a Cloud Firewall. Caddy handles TLS automatically from Let's Encrypt — and this is where I got my first real lesson: the first certificate accidentally came from the staging CA, which browsers reject. The fix was simple (the right `acme_ca` directive + removing the stale cert + restart), but it taught me not to declare TLS "done" without checking the issuer with `openssl`. SQLite runs in WAL mode at `/var/lib/goth/goth.db`, and the database backs itself up both locally and to Cloudflare R2 via an `rclone` cron every night — a round-trip check (`sha256sum` locally vs. R2) and a complete restore drill were both done before I touched any DNS. `deploy.sh` verifies the checksum, runs migrations, flips the symlink atomically, and rolls itself back if the health check doesn't pass.

### Next.js got its own domain, and the comparison became fair

Originally Next.js was meant to be taken to the same CX23 behind Caddy's cookie routing. I ended up with a different solution: Next.js got its own subdomain (`next.karotammela.fi`) and stayed on Vercel, while Go took over the apex domain. That way the comparison actually measures two different deployment models — "self-hosted Go VPS in Helsinki" versus "Vercel Edge" — instead of the artificial delay a Caddy proxy would have added to one side. The Tech Switcher (that round button in the bottom-right corner) now does a cross-origin redirect that preserves the path and locale, and the Live Performance widget pings both `/api/ping` routes directly from your browser and shows the median TTFB — not a server-side estimate, but what your own browser actually measured. In the first measurements Go+HTMX responded in about 50 ms, Next.js from Vercel in about 190 ms — the difference shows up directly in the bar lengths. On this site, of course, that latency doesn't matter at all, but you do notice the speed, for example in how quickly the AI bouncer returns a response or how fast the different pages load in the browser.

### The final piece: deploy with a single push

Until now, releasing was manual: `make release`, `scp` to the server, `sudo ./deploy.sh`. I finished turning it into an automated GitHub Actions pipeline today: a dedicated, passphrase-less deploy key used only by CI, three repo secrets, and a workflow where `test` (vet, unit tests, the standalone drill) → `release` (an identical build to the manual one) → `deploy` (scp + `deploy.sh` + a production health check) run back to back every time changes in the `goth/` folder are pushed to the `main` branch. The first automated run succeeded right away — tests, build, and deploy in under four minutes, with production answering healthy right after.

### What I learned

Go+HTMX surprised me with how far you get with so little code once you don't need React hydration or bundler optimizations — a single 20–30 MB binary handles everything, and TTFB measured from my own machine in Helsinki comes out to roughly a quarter of the Vercel edge. The price for that, of course, is paid in manual labor: a lot of things in Next.js come straight from the framework. Neither stack is absolutely "better" — but now I know where the difference comes from, because I built both myself and measured it from my own browser. I'll probably move more lightweight sites over to Go+HTMX, since they load significantly faster and require fewer resources. A CX23 instance can also fit several projects at roughly 5 € per month. I can drop my web hosting costs to a minimum this way, since I also set up a Purelymail account for another one of my projects, which gets email working for a 10 € annual cost.

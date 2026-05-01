---
title: "From WhatsApp Chaos to Automation: How I Built the Sales Platform for Levi Golf Green Fees with Agentic AI"
description: "How manual messaging and share management evolved into a fully automated platform powered by a modern tech stack and an AI-driven dev team."
publishedAt: "2026-05-01"
slug: "from-whatsapp-chaos-to-automation-green-fee"
draft: false
tags: ["coding", "ai", "golf", "automation", "business"]
---

Through my company, I own several shares in Levi Golf and manage the sale of playing rights for other owners as well. With each share providing 35 usable tickets per season—but restricted to only one use per day—manual management had become an incredible burden. It was a constant cycle of WhatsApp threads, manual spreadsheet checks, and endless back-and-forth messaging just to close a single sale. The risk of double-booking was real, and the customer experience was far from modern.

I decided to build a fully automated platform where customers could handle everything themselves: select a date, pay securely, and receive an instant confirmation.

This was the birth of [greenfee.levifinland.fi](https://greenfee.levifinland.fi).

### The Challenge: The Complex Logic of Playing Rights

The technical heart of this system isn't just about processing payments; it’s about managing a complex set of rules. In the Levi Golf model, availability is tied to specific shares. Each share includes 35 tickets per season, but with the strict limitation that only one ticket per share can be sold for any given day. In total, this means managing a pool of several hundred tickets throughout the season.

To solve this, I implemented a relational database architecture that scans all managed shares in real-time. When a payment is confirmed via Stripe, the system automatically allocates the ticket to the first available share. It also evaluates which shares have the most remaining inventory and prioritizes them accordingly. My goal is to distribute sales evenly throughout the season; I don't want to "sell out" one share early, which would leave less daily inventory for the remainder of the year. This removes the possibility of human error, prevents overbooking, and ensures tickets are sold proportionately across the entire portfolio.

This is how [greenfee.levifinland.fi](https://greenfee.levifinland.fi) evolved from actual need to automate the sales process.

![Green Fee sales platform](/media/greenfee-levifinland-fi.png)

### Database and Security at the Core

I chose **Turso (LibSQL)** as the database. It is a distributed SQLite-based solution that allows data to be kept close to the user at the Edge. This guarantees lightning-fast load times, whether the customer is booking their ticket from Levi, Helsinki, or Amsterdam.

Security is baked into the architecture:
* **Isolated Data:** The system is a multi-tenant platform. Every shareholder sees only their own data, analytics, and share status on their personal dashboard.
* **Stripe Webhooks:** Ticket allocation only occurs once the Stripe server confirms a successful payment. This ensures the booking system remains perfectly synchronized with actual cash flow.

### Dashboard – Management Without the Effort

I built a comprehensive management panel where shareholders can track their 35-ticket quota and download reports for accounting. I also integrated the **Resend** email API, which sends customers immediate purchase confirmations along with instructions on how to activate and use their tickets.

Since Levi attracts international travelers, the application is fully localized into five languages: Finnish, English, French, German, and Dutch.

### Agentic AI – Orchestrating Models in Production

The most interesting part of this project was how it was built. I utilized **Agentic AI** coding, specifically leveraging the [kilo.ai](https://kilo.ai) platform—my primary Agentic AI tool within VS Code. The AI agent acts as a software engineer, implementing file changes and executing plans according to my instructions.

I used several different large language models (LLMs) throughout the process, playing to their individual strengths:

1.  **The Builder (GPT Codex 5.3):** Handled 80% of the "heavy lifting" by translating business requirements into Next.js components and API routes. Codex is currently my clear favorite for price-to-performance.
2.  **The Architects (Opus 4.6 & 4.7):** Senior-level auditors. These models handled the critical sections, such as payment logic security and security auditing. While I would love to use these constantly, their token costs are significantly higher, so I optimize their use for mission-critical tasks only.
3.  **The Localization Expert (Gemini 3.1 Flash Lite Preview):** Ensured that the localization across five languages didn't sound like a machine translation, but rather natural and native. The token pricing for this model is the most affordable of the group, making high-volume translations cost only a few cents.



### The Result: A Frictionless, Automated Purchase Process

The transformation has been immense. What used to take hours of messaging and manual verification during the golf season has now been reduced to nearly zero.

For me, this way of coding feels like the world-building games I played as a kid—like *Civilization* or *Command & Conquer*—but instead of a game world, I’m building real-world applications that simplify life. Thanks to Agentic AI, the barriers to coding have vanished, leaving only pure problem-solving.

As far as I'm concerned, the manual era is officially over.

---
*Are you a Levi Golf shareholder looking to automate your sales? Explore the app at [greenfee.levifinland.fi](https://greenfee.levifinland.fi) or learn more about the future of coding at [kilo.ai](https://kilo.ai).*
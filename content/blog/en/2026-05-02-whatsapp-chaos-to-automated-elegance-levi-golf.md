---
title: "From WhatsApp Chaos to Automated Process: Building the Levi Golf Green Fee Sales Platform"
description: "How a manual ticket sales workflow evolved into a secure, multilingual, multi-vendor platform built with Agentic AI."
publishedAt: "2026-05-02"
slug: "whatsapp-chaos-to-automated-process-levi-golf"
draft: false
tags: ["coding", "ai", "golf", "automation", "saas"]
---

For anyone who has managed a high-demand service, the manual era is a familiar nightmare. For a long time, selling green fee tickets for Levi Golf was exactly that: phone calls, fragmented WhatsApp messages, and spreadsheet Tetris to keep track of which shares were available on which dates.

The risk of double-booking was real, the administrative stress was massive. I wanted a system where customers could do everything themselves: choose a date, pay securely, and receive confirmation instantly, with almost zero manual work on my side.

This is how I built the Levi Golf Green Fee application using an Agentic AI workflow.

### The Core Challenge: Share-Based Allocation Logic

The hardest part was not payments, but business logic. In this model, availability is tied to shares. Each share includes 35 tickets per season, with one strict rule: only one ticket per share can be sold per day.

To make this reliable, I implemented a relational architecture with Drizzle ORM and SQLite. Instead of only counting remaining tickets, the system tracks the state of every share. When a customer picks a date, the app performs a real-time availability check and returns only shares that are valid for that exact day.

After Stripe confirms payment, the platform allocates the booking to the first eligible share automatically. This guarantees the one-per-day rule and removes human error and overselling risk.

![Green Fee sales platform](/media/green-fee-levi-golf-booking-en.png)

### Security, Stability, and Admin Control

Because the system handles payments and customer data, security is non-negotiable. The frontend is intentionally thin, while all critical operations happen server-side: payment verification, share allocation, and transactional messaging.

Stripe Webhooks are central to this architecture. A ticket is created only after Stripe confirms a successful payment event, which protects the system from fake bookings and race-condition edge cases.

I also built a full Admin Dashboard with scalability in mind. The platform supports onboarding additional shareholders as vendors, so each vendor can manage their own shares and monitor their own sales without touching core code. This turns a personal workflow tool into a true multi-vendor sales platform.

![Stripe payment flow](/media/green-fee-levi-golf.png)

### Communication and Localization

A booking flow is incomplete without instant and trustworthy communication. To remove manual follow-up messages, I integrated the Resend API. As soon as a payment succeeds, the system sends a confirmation email with the exact ticket numbers required at the caddiemaster office.

Levi Golf serves an international audience, so the application is fully localized in Finnish, English, French, German, and Dutch. From landing page to confirmation email, the experience remains native and consistent for each customer segment.

### The Build Method: Agentic AI and Model Orchestration

The most interesting part of this project is how it was built. I used Agentic AI coding with Kilo Code.

Unlike traditional chat-based coding, the agent operates like a software engineer: it reads the codebase, follows plans, updates files, runs checks, and iterates on fixes. That changed both implementation speed and reliability.

The real value comes from model orchestration. I used different models for specific responsibilities:

1. **The Builder (GPT Codex 5.3):** primary implementation engine for React UI, Next.js routes, and rapid feature delivery. This model offers excellent value for money. 
2. **The Auditor (Opus 4.6 & 4.7):** deep validation for security-sensitive logic, transaction integrity, and webhook idempotency edge cases. Opus is excellent and reliable work horse, but the cost is clearly higher than almost any other.
3. **The Copywriter (Gemini 3.1 Flash Lite Preview):** multilingual copy and tone adaptation across five languages. This model is fast, cheap, and handles the tasks that doesn't require deep technical understanding.

This role-based approach produced better outcomes than forcing one model to do everything.

### Result: The End of Manual Operations

The operational difference is dramatic. Tasks that previously required hours of manual coordination now run automatically: availability checks, payment confirmation, ticket allocation, customer email delivery, and vendor-level tracking.

I no longer spend time reconciling spreadsheets and chat threads. The platform is now secure, scalable, and self-sufficient.

The manual era is officially over. 

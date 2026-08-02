---
title: "How I moved a 10-year-old WordPress site to Next.js in one workday"
description: "A project diary about moving Hiihtogreeni.fi from WordPress to Next.js 16 with AI while preserving Google visibility, content, and legacy URLs."
publishedAt: "2026-08-02"
slug: "how-i-moved-wordpress-site-to-next-js"
draft: false
tags: ["Next.js", "WordPress", "AI", "web development", "SEO"]
---

## How I moved a 10-year-old WordPress site to Next.js in one workday with AI

Many company websites reach a point where their content and Google visibility are valuable, but maintenance has become unnecessarily heavy. In this project I moved the ten-year-old WordPress/Elementor site [Hiihtogreeni.fi](https://hiihtogreeni.fi/) to a modern Next.js 16 architecture in one workday.

The goal was not simply to build something newer and faster. The important part was preserving the business value in the existing content, images, URL structure, and search visibility.

## TL;DR

- **Starting point:** Hiihtogreeni.fi ran on a heavy WordPress/Elementor combination. The content and SEO were solid, but maintenance was cumbersome.
- **Goal:** Move to Next.js 16 without losing search traffic or legacy URLs.
- **Implementation:** AI-assisted development with Kilo Code, structural auditing, static generation, and exact redirects.
- **Result:** A fast, secure, low-maintenance site. Form spam dropped to zero after the rebuild.

## Phase 1: Understand the old site before building the new one

The biggest risk in a website redesign is starting to code before mapping the old content and URL structure. I began by collecting the visible text, page structures, links, image references, and contact details from the old site with Kiloclaw, a hosted OpenClaw implementation.

I asked GPT-5.6 Sol to compare the collected data with the live site and create a chronological migration plan. The audit found an old English gallery URL (`/en/gallery/`) that was still public and indexed by Google. Without this phase, that URL would easily have been broken in the redesign.

AI produces code quickly, but speed only matters when the inputs and goals are correct. A visually successful new site is still a failed project if important pages disappear from Google.

## Phase 2: Visual identity and brand DNA

The technology stack became Next.js 16, React 19, Tailwind CSS 4, and TypeScript. Most pages can be generated as static HTML, so JavaScript is loaded in the browser only where it is needed, such as the mobile menu, image gallery, and contact form.

![Hiihtogreeni.fi rebuilt with Next.js](/media/from_wordpress_to_next_js_hiihtogreeni.png)

*The new Hiihtogreeni.fi: a familiar brand with a cleaner and faster interface.*

Together with AI, I defined the visual direction as **Arctic calm, timber warmth**. It is built around strong winter photography, quality interiors, the familiar blue-green shade (`#486C7A`), and a warm orange accent. Navigation and language controls are rendered in static HTML, so the core site remains usable even if browser JavaScript fails to load.

## Phase 3: SEO protection and preserving legacy files

WordPress often stores images and files in paths such as `/wp-content/uploads/2022/10/...`. External sites and Google Image Search may still link to those URLs. I therefore recreated the original WordPress directory structure in the new Next.js project's `public/` directory for images and the important booking-terms PDF.

The old images now return a real `200 OK` response from the new server without unnecessary redirects.

## Redirects and 410 Gone

The SEO-safe migration depended on handling old URLs precisely:

- The old English homepage and gallery permanently redirect (`301`) to `/en/home/` and `/en/photo-gallery/`.
- The old WordPress sitemap `/wp-sitemap.xml` redirects to the new sitemap.
- System paths such as `/wp-admin/`, `/feed/`, and `/wp-json/` return `410 Gone` instead of redirecting to the homepage. This clearly tells search engines that the paths no longer exist and reduces wasted bot crawling.

## Phase 4: Performance, LCP, and the value of manual QA

AI is not infallible. Manual quality assurance uncovered two critical issues.

The first was the mobile menu. Automated tests passed, but on a real phone the menu became trapped inside the blurred header because of `backdrop-filter`. I fixed it by rendering the menu through a portal directly into `document.body`.

The second issue was Largest Contentful Paint, or LCP. The initial mobile Lighthouse performance score was 72 because the large hero image loaded as a heavy JPEG through the old WordPress structure. I created a separate LCP path for hero images using AVIF and WebP, a `<picture>` element, and preload. The old images stayed in place for external links.

![WordPress PageSpeed result before the migration](/media/pagespeed_insights_hiihtogreeni.fi_wordpress.png)

*The WordPress mobile result: performance 54 and SEO 92.*

![Next.js PageSpeed result after the migration](/media/excellent_page_speed_with_next_js.png)

*The Next.js mobile result: performance 92, accessibility 95, best practices 100, and SEO 100.*

The new formats raised the performance score to 95 and reduced LCP to 0.6 seconds.

## Phase 5: Privacy first, contact form, and analytics

A static site still needs a server-side solution to receive contact messages. I chose Resend and sent messages through a Next.js server route. No API keys are exposed in the browser, and the server filters spam and prevents duplicate submissions independently.

![A contact form protected from spam](/media/contact_form_with_no_spam.png)

*The form stays simple while server-side processing keeps spam under control.*

Before the rebuild, the form could generate several spam messages per day. Afterwards, the amount dropped to zero.

I also ported a cookie-consent implementation from earlier projects. By default, the site sets only one necessary cookie for storing consent. Google Analytics 4 runs in Consent Mode v2 and activates only after explicit user permission. We track conversions such as form submissions and email clicks without collecting personal data.

## Project figures

- Data collection, planning, and groundwork: about 1 h 50 min
- Page content, form, SEO, and rules: about 1 h 20 min
- QA, performance fixes, and production launch: about 3 h 30 min
- Cookies and GA4: about 1 h 20 min
- **Total active work: about 8 hours**

Total AI API cost was approximately **€9.55**. The tools and models included Kiloclaw, GPT-5.6 Sol, Qwen3.7 max, and Hy3.

## Conclusion: the value of AI in web development

The Hiihtogreeni.fi rebuild shows how AI should be used in software development. It does not replace human understanding of business goals and SEO, but it is a powerful tool when tasks are scoped and results are checked in stages.

Agent-based coding is fast, but I do not ship production code in YOLO mode. I keep control of the work one task at a time and manually verify the critical details. With less than ten euros in AI costs and roughly one workday of effort, a decade-old WordPress installation became a modern, fast Next.js application without anything disappearing from Google's perspective.

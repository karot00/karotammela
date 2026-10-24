---
title: "llms.txt, GEO and the Agentic Web: How to Prepare Your Website for the AI Era"
description: "What llms.txt, llms-full.txt and GEO actually do, what they do not do, and how agentic commerce is changing websites. With my own UTM tool Kanavoi.com as the worked example."
publishedAt: "2026-10-24"
slug: "agentic-shopping-and-ai-web-browsing"
draft: false
tags: ["AI", "llms.txt", "GEO", "SEO", "agents", "ecommerce"]
---

The information architecture of the internet has gone through several major shifts. In the first wave, websites were built for human browsers and visual browsing. In the second, they were optimized for search crawlers such as Google's and for pages of ten blue links (SEO).

We are now living through a third shift, possibly the fastest so far: **the move from a web for people to a web for language models and AI agents**.

Microsoft Advertising's 2026 playbook [*Prepare for the Age of AI Shopping*](https://about.ads.microsoft.com/content/dam/sites/msa-about/global/common/content-lib/pdf/prepare-for-the-age-of-ai-shop.pdf) brings together some recent numbers:

- According to HUMAN Security's 2026 report, **automated traffic is growing eight times faster than human traffic**, and AI-driven sessions nearly tripled in 2025.
- According to Adobe Analytics (Q1 2026), **AI-referred shoppers convert 42% better** than other traffic.
- In Microsoft's US consumer survey (n = 1,028), **more than 72% of respondents expected brands to offer AI-powered shopping experiences within a year**.

AI systems no longer just fetch links. They read, compare, make decisions and increasingly carry out tasks on people's behalf. That gives web developers and marketers a new question to ask: does AI understand my website correctly?

Throughout this post I use my own project, [Kanavoi.com](https://kanavoi.com), as the example. Kanavoi is a Finnish UTM tool for marketing teams and agencies. It keeps campaign link naming consistent and shows, before you publish, which Google Analytics 4 channel (such as Paid Social or Email) the traffic will be reported in. It makes a good example because it is a real, small and new website. When I started this post, some of the AI-era basics were still missing, so I implemented them while writing. Every example below is one of Kanavoi's real files, live in production.

Three concepts dominate the conversation: the open proposal **`llms.txt`**, the companion convention **`llms-full.txt`**, and the broader discipline of **GEO (Generative Engine Optimization)**. They are often mixed up, so let us separate them properly.

---

### 1. The big picture: three eras of the web

Microsoft's playbook divides how people use the web into three parallel modes:

```
+-----------------------------------------------------------------------------------+
| 1. THE HUMAN WEB           2. THE LLM WEB              3. THE AGENTIC WEB         |
| "Help me find it"    -->   "Help me choose"      -->   "Do it for me"             |
|                                                                                   |
| Traditional browser,       Language models summarize,  The human sets the goal    |
| tabs, search engine        compare options and         and an agent finds,        |
| links and manual           curate recommendations.     evaluates and completes    |
| research.                  The human decides.          the purchase or task.      |
+-----------------------------------------------------------------------------------+
```

All three are in use at the same time. The human web still brings in most traffic, but the biggest new opportunities are in the LLM and agentic webs.

#### Every interaction now has a third party

Marketers and web developers used to focus on convincing a **human**: striking imagery, visual storytelling, animations and carefully placed calls to action (CTAs).

Now there is often **a third party involved: an AI assistant or agent**.

- AI is not persuaded by CSS animations, gradients or brand videos.
- Language models and agents evaluate products and services primarily on **structured, machine-readable and up-to-date information**.
- Agents can read ordinary web pages, but then they have to infer details from content designed for people, which easily leads to mistakes. If information is incomplete or hard to interpret, **there is a real risk that your product never makes it into a recommendation**.

---

### 2. The "noise problem" of modern websites

A modern website is expensive, slow and risky for an AI model to read:

1. **HTML clutter and scripts:** Sites are full of JavaScript, tracking pixels, cookie banners, pop-ups and navigation bars.
2. **Wasted context window:** Language models measure input in tokens. On a script-heavy page, much of the downloaded data is presentation or tracking code with no value to the model.
3. **Risk of misinterpretation:** When the facts are buried among ads and recommendation widgets, the risk grows that the model picks up the wrong price, an outdated condition, or invents missing information (hallucinates).

Two different approaches have been proposed for this:

- **`llms.txt`:** In September 2024, Jeremy Howard of Answer.AI proposed a Markdown file in the website root: a concise guide for language models to the site's most important content. The proposal is published at [llmstxt.org](https://llmstxt.org).
- **GEO (Generative Engine Optimization):** The term comes from a 2023 [research paper](https://arxiv.org/abs/2311.09735) (Aggarwal et al., including Princeton and IIT Delhi). GEO is not a file format but a broader discipline: how to write and structure content so that generative search engines include it in their answers and cite it. In the study, adding citations, quotations and concrete statistics improved visibility.

So `llms.txt` is one GEO tool, not a synonym for it.

---

### 3. Do these files help Claude, ChatGPT and Gemini?

This deserves an honest answer, because there is a lot of hype around the topic. AI uses web content at three levels:

| Scenario | How the AI works | Impact of llms.txt |
| :--- | :--- | :--- |
| **Agents and coding tools** (e.g. Claude Code, Cursor, GitHub Copilot, OpenAI Codex) | Fetch documentation and source material from a URL when the user or agent points them to it. | **Clearest benefit.** When an agent is sent to a site, `llms.txt` tells it right away which pages are worth reading and saves tokens. This is why documentation sites publish it most often. |
| **Real-time search in chat apps** (ChatGPT search, Gemini, Microsoft Copilot, Perplexity) | Pull results from a search index and read the pages they find, usually as HTML. | **Indirect.** The major services have not said they use `llms.txt`, and Google has stated that Google Search ignores it. What matters is that the content of the pages themselves is clear and easy to extract. |
| **Model training** | Crawlers such as GPTBot and ClaudeBot collect the web at scale. | **No demonstrated effect.** Access is still controlled with `robots.txt`. |

As of October 2026, none of the major AI companies has committed to reading `llms.txt` files automatically, even though many of them, Anthropic included, publish one for their own documentation. Independent measurements have found no link between having the file and how often language models cite a site.

So why do it anyway? It is cheap to implement, it genuinely helps agents and developer tools, and above all **writing it forces you to condense the core facts of your service**: what you do, for whom, at what price and on what terms. That is exactly the information AI looks for on your regular pages too.

---

### 4. The roles of the files: llms.txt vs. llms-full.txt

Both files live in the **website root**, next to the familiar `robots.txt` and `sitemap.xml`. On Kanavoi they look like this:

```
https://kanavoi.com/robots.txt               <- Crawler access rules
https://kanavoi.com/sitemap.xml              <- Sitemap for search engines
https://kanavoi.com/llms.txt                 <- Table of contents for AI
https://kanavoi.com/llms-full.txt            <- Key content in a single file
https://kanavoi.com/ohjeet/utm-parametrit.md <- A single guide as Markdown
```

#### llms.txt: the site's table of contents

`llms.txt` is a lightweight file that gives a language model a quick overview:

- a short core description of the site or service
- a curated list of links to the most important sections
- a brief description of what each link contains.

**Who it is for:** Agents with a limited context window, or agents looking for a precise answer to one question, such as a single API endpoint or a return policy.

#### llms-full.txt: all the key information at once

`llms-full.txt` gathers the site's key content into one long Markdown document. It is not part of the original proposal, which instead describes `llms-ctx.txt` and `llms-ctx-full.txt` files generated with a tool. `llms-full.txt` is a convention popularized by documentation platforms such as Mintlify, and it has become a de facto standard.

The context windows of today's leading models, such as Anthropic's Claude Opus 5.5 and the latest models from OpenAI and Google, hold hundreds of thousands of tokens, and some hold up to a million. When a model is given `llms-full.txt`:

1. It does not need to make dozens of separate HTTP requests, one page at a time.
2. It can load an entire product catalog or documentation set into working memory at once.
3. It can cross-reference pages without the risk of missing a subpage.

A large context is not free, though: every token costs money, and accuracy drops in very long contexts. Keep `llms-full.txt` focused on what matters.

#### robots.txt still decides access

`llms.txt` neither grants nor blocks access. `robots.txt` does. AI companies now run several crawlers for different purposes: for example OpenAI's `GPTBot` (training), `OAI-SearchBot` (ChatGPT search) and `ChatGPT-User` (user-initiated fetches), and Anthropic's `ClaudeBot`, `Claude-SearchBot` and `Claude-User`. Google's `Google-Extended` token lets you opt out of your content being used for Gemini models without affecting Google Search.

Kanavoi's `robots.txt` is currently simple: all crawlers may read the public pages. Kanavoi wants to be found in AI search, so that is a deliberate choice. The key point is not to block search bots by accident if you want to appear in AI answers.

---

### 5. File anatomy: Kanavoi's llms.txt

The format is plain Markdown. The structure follows the order in the llmstxt.org proposal:

1. **H1 heading (`#`):** The name of the project or site. This is the only required part.
2. **Blockquote (`>`):** A short summary containing the information needed to understand the rest of the file.
3. **Free-form paragraphs or lists:** More detail about the site (no headings).
4. **H2 sections (`##`):** Link lists in the form `- [Title](url): Optional description`.
5. **`## Optional` section:** Secondary material an agent can skip when context is tight.

The proposal also recommends offering a Markdown version of each page at the same URL with a `.md` suffix. On Kanavoi every guide is available this way, for example [kanavoi.com/ohjeet/utm-parametrit.md](https://kanavoi.com/ohjeet/utm-parametrit.md), and `llms.txt` links straight to these versions.

#### Kanavoi's [llms.txt](https://kanavoi.com/llms.txt)

Kanavoi's site is in Finnish, so its files are in Finnish too. This is the real file as it is served:

```markdown
# Kanavoi

> Kanavoi on suomalainen UTM-työkalu markkinointitiimeille ja -toimistoille. Se pitää UTM-linkkien nimeämisen yhtenäisenä ja näyttää jo ennen julkaisua, missä Google Analytics 4:n kanavassa (esim. Paid Social tai Email) liikenne raportoidaan.

Ilmaiset työkalut toimivat ilman kirjautumista. Palvelin on Helsingissä, ja varmuuskopiot ovat EU:ssa. Ei seurantaevästeitä. Sivuston kävijämäärät lasketaan ilman evästeitä.

## Ilmaiset työkalut

- [UTM-generaattori](https://kanavoi.com/#generaattori): Rakentaa UTM-linkin, siistii arvot (esim. Kevät Ale → kevat_ale) ja näyttää ennen julkaisua, missä GA4:n kanavassa liikenne raportoidaan.
- [GA4-kanavatarkistin](https://kanavoi.com/ga4-kanavatarkistin): Tarkista ilmaiseksi, mihin Google Analytics 4:n kanavaan utm_source- ja utm_medium-arvot ohjaavat liikenteen, ja korjaa Unassigned-liikenne.
- [QR-koodi](https://kanavoi.com/qr-koodi): Tee ilmainen QR-koodi linkille. Lisää UTM-tunnisteet ja lataa koodi SVG-muodossa painoon tai PNG-kuvana. Ei rekisteröitymistä.

## Ohjeet

- [UTM-parametrit: kattava opas suomeksi](https://kanavoi.com/ohjeet/utm-parametrit.md): Mitä viisi UTM-parametria tarkoittavat, mitkä niistä ovat välttämättömiä ja miten nimeät ne niin, että GA4:n raportit pysyvät siisteinä.
- [Miksi GA4 näyttää liikenteen Unassigned-kanavassa ja miten sen korjaa](https://kanavoi.com/ohjeet/ga4-unassigned.md): Unassigned tarkoittaa, ettei mikään GA4:n kanavasääntö täsmännyt. Näin löydät syyn ja korjaat sekä linkit että raportit.
- [Google Ads ja UTM: tarvitaanko niitä gclidin rinnalla?](https://kanavoi.com/ohjeet/google-ads-utm.md): Automaattinen merkintä (gclid) riittää GA4:lle. Näin päätät, tarvitaanko UTM-tunnisteita lisäksi, ja lisäät ne sotkematta raportteja.
- [UTM-tagit Meta-mainoksiin oikein](https://kanavoi.com/ohjeet/utm-meta-mainokset.md): Facebook- ja Instagram-mainosten UTM-tunnisteet: oikea lähde ja media, dynaamiset parametrit ja miksi lähde ei saa olla meta.
- [UTM-tagit uutiskirjeisiin](https://kanavoi.com/ohjeet/utm-uutiskirjeet.md): Uutiskirjeen linkkien UTM-tunnisteet: media email, kampanja jokaiselle lähetykselle ja sisältö jokaiselle linkille. Toimii kaikissa postituspalveluissa.

## Hinnat

- [Hinnat](https://kanavoi.com/#hinnat): Ilmainen 0 €, Tiimi 14 €/kk, Toimisto 49 €/kk. Hinnat ilman arvonlisäveroa.
- [Odotuslista](https://kanavoi.com/#odotuslista): Tilit avataan odotuslistalla oleville. Paketin voi valita liittyessä.

## Optional

- [Koko sisältö yhtenä tiedostona](https://kanavoi.com/llms-full.txt): Palvelun kuvaus, hinnat, usein kysytyt kysymykset ja kaikki ohjeet.
- [Tietosuojaseloste](https://kanavoi.com/tietosuoja): Mitä henkilötietoja Kanavoi käsittelee, miksi ja kuinka kauan.
- [Käyttöehdot](https://kanavoi.com/kayttoehdot): Kanavoi-palvelun käyttöehdot yritysasiakkaille.
```

Note that prices and privacy commitments are written out, not just linked. An agent that reads only `llms.txt` still gets the basic facts right.

#### Kanavoi's [llms-full.txt](https://kanavoi.com/llms-full.txt)

`llms-full.txt` gathers the same information as actual text in a single file (about 30 KB), with sections separated by horizontal rules (`---`). It contains the service description, how it works, pricing, privacy commitments, the FAQ and every guide. Here is the beginning, translated for this post:

````markdown
# Kanavoi: service description, pricing and guides

> Kanavoi is a Finnish UTM tool for marketing teams and agencies. It keeps UTM link naming consistent and shows, before publishing, which Google Analytics 4 channel (e.g. Paid Social or Email) the traffic will be reported in.

This file gathers Kanavoi's key content into a single text for language models. Up-to-date information: https://kanavoi.com/

## How Kanavoi works

1. **You agree on the rules once.** Allowed sources and mediums, and the campaign name format. A ready-made template follows GA4's rules, so getting started is quick.
2. **Everyone creates links with the same form.** You cannot pick a wrong value, and spelling is cleaned up automatically: Kevät Ale → kevat_ale.
3. **You see the result before publishing.** Next to the link you see which GA4 channel the traffic will be reported in, with a warning and a suggested fix if it would land in the wrong place.

## Pricing

Accounts open soon. These prices apply from opening: a flat monthly price for the whole team, no per-user fees. Join the waitlist to get an invite among the first. Prices exclude VAT.

- **Free, 0 €**
  - Generator, GA4 channel check and QR codes
  - One user
  - The 50 most recent links saved
- **Team, 14 €/month**
  - Unlimited users and links
  - Naming rules and campaign name checks
  - Many channels at once, Meta and Google Ads formats
  - Link health check, API
...
````

Two details matter more than they seem. First, the pricing also states its condition: the prices apply from when accounts open. Without that, an agent might tell a user they can buy the service today. Second, the precise facts in the guides survive. On the website a GA4 channel is shown as a coloured box, but in the Markdown version it is an ordinary sentence:

```markdown
GA4-kanava, kun lähde on uutiskirje ja media newsletter: **Unassigned**. Mikään GA4:n oletuskanavaryhmän sääntö ei täsmää, joten liikenne näkyy raporteissa kanavattomana.
```

("GA4 channel when the source is uutiskirje and the medium newsletter: Unassigned. No rule in GA4's default channel group matches, so the traffic shows up in reports without a channel.") This is exactly the kind of fact a language model cannot infer from marketing copy.

---

### 6. How to implement the files technically

The approach depends on your site's architecture.

#### Option A: Static sites (Astro, Next.js, Hugo, VitePress)

If your content is already in Markdown files, run a build script:

```javascript
// scripts/generate-llms.mjs
import fs from "node:fs";
import path from "node:path";

const DOCS_DIR = "./content/docs";
const OUTPUT_FULL = "./public/llms-full.txt";

const files = fs
  .readdirSync(DOCS_DIR)
  .filter((file) => file.endsWith(".md"))
  .sort();

let fullContent = "# Site name: full documentation\n\n";

for (const file of files) {
  const content = fs.readFileSync(path.join(DOCS_DIR, file), "utf-8");
  // Strip front matter so only the actual content remains.
  const body = content.replace(/^---\n[\s\S]*?\n---\n/, "");
  fullContent += `\n---\n\n${body.trim()}\n`;
}

fs.writeFileSync(OUTPUT_FULL, fullContent);
console.log("llms-full.txt generated.");
```

#### Option B: A dynamic route (Next.js App Router, WordPress, Shopify)

If your content lives in a database, a server-side route can convert it to Markdown and return it as plain text. In Next.js, the folder can simply be named `llms.txt`:

```typescript
// app/llms.txt/route.ts (Next.js App Router)
export const dynamic = "force-static";

export function GET() {
  const markdown = [
    "# Company Ltd",
    "",
    "> A one-sentence description of the company.",
    "",
    "## Content",
    "- [About us](https://company.com/about): Profile and values.",
  ].join("\n");

  return new Response(markdown, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Cache-Control": "public, max-age=86400",
    },
  });
}
```

#### Option C: A Go server (the Kanavoi way)

Kanavoi is a Go monolith whose guide articles are Markdown files embedded in the binary. All three file types are built at startup from the same sources as the HTML pages: the translation file's texts, the prices and the guides. The content can never drift apart, and a new guide ends up in the files automatically.

This is how a single guide's Markdown version is served:

```go
// handleDocsMarkdown serves /ohjeet/<slug>.md. The canonical link header keeps
// search engines on the HTML page.
func (a *App) handleDocsMarkdown(w http.ResponseWriter, r *http.Request, slug string) {
	art, ok := a.docs.Get(slug)
	if !ok {
		a.handleNotFound(w, r)
		return
	}
	w.Header().Set("Link", `<`+a.SiteURL+art.Path()+`>; rel="canonical"`)
	writeMarkdown(w, []byte(a.articleMarkdown(art)))
}
```

Remember the `Link: rel="canonical"` header. Without it, a search engine may treat the Markdown version as a duplicate of the same content and show it in search results instead of the HTML page. The HTML page's `<head>` in turn announces the Markdown version: `<link rel="alternate" type="text/markdown" href="…/utm-parametrit.md">`.

Kanavoi's articles use a few custom components (example links, GA4 channels, tip boxes and "Try it yourself" cards). The build step turns them into plain Markdown and makes the site's internal links absolute, so they work outside the site too. Otherwise the file would contain markup a language model does not understand.

> **Technical tip:** Make sure your server returns the header `Content-Type: text/plain; charset=utf-8` or `Content-Type: text/markdown; charset=utf-8` for these files. That tells any reader right away that this is not an HTML page meant for a browser.

---

### 7. Agentic commerce standards: UCP and buying inside the conversation

Beyond documentation, the agentic shift hits ecommerce especially hard. The core message of Microsoft's playbook is that being discoverable is not enough: **the path to purchase is compressing into the conversation itself**.

#### Universal Commerce Protocol (UCP)

Universal Commerce Protocol is an open standard announced by Google and Shopify in January 2026, co-developed with companies such as Etsy, Wayfair, Target and Walmart. It lets AI assistants and commerce platforms talk to each other without custom integrations for every bot. Microsoft also supports UCP in Merchant Center. According to Microsoft's playbook, UCP builds on your existing product feed and enriches it with information that matters to agents:

- return policy and customer support contacts (in store-wide settings)
- product-level warning and regulatory attributes (`consumer_notice_type` and `consumer_notice_message`), which are required for categories such as toys and regulated goods
- a unique product identifier (`merchant_item_id`) for tracking an item across feeds and systems.

The foundation is still the basic product data: accurate titles, descriptions, GTIN/MPN, real-time availability, pricing and product category. According to the playbook, outdated prices and stock mismatches are the most common reasons products lose visibility, so refresh your feed at least daily.

#### Buying inside the conversation (agentic checkout)

When a user asks AI to handle a purchase for them ("Do it for me"), the order is completed without leaving the conversation. Microsoft's **Copilot Checkout** works on Copilot.com and in the mobile app, and is available through Shopify (eligible merchants are enrolled automatically), PayPal and Stripe. The business remains the *merchant of record*, and payments, fulfillment and the customer relationship stay in its own systems. OpenAI, for its part, has built the open Agentic Commerce Protocol (ACP) with Stripe for purchases in ChatGPT.

**An important note for merchants outside the US:** both Microsoft's UCP support and Copilot Checkout are currently US-only, and Copilot Checkout is still in pilot. Investing in product data quality pays off now regardless, because it helps in every channel.

#### What if you do not sell products?

Kanavoi is a SaaS service, not a webshop, so it has no product feed. For a service, the agentic-web equivalent is an interface agents can use directly. Kanavoi's roadmap includes a REST API and an MCP server (Model Context Protocol) with the tools `create_utm_link`, `list_allowed_values` and `classify_ga4_channel`. The idea is that an AI agent can create campaign links according to the team's rules, but cannot invent new values for locked fields. The agent does the work, and the team's rules still hold.

---

### 8. How to measure AI visibility

Traditional analytics tells you what happened on your site, but rarely **why AI recommended a competitor**. Two significant changes to measurement arrived during 2026.

#### GA4's own AI Assistant channel

In May 2026, Google added an **AI Assistant** channel to GA4's default channel group. When a session's referrer is recognized as an AI assistant (such as ChatGPT, Gemini or Claude), GA4 sets the medium to `ai-assistant` and reports the session as its own channel. Three things are worth knowing:

- The change does not apply to historical data, so there is no comparable history.
- If an AI app does not pass referrer information, the visit still shows up as Direct.
- Check your own filters and channel groups. AI traffic that used to appear under Referral now shows up as its own channel with the medium `ai-assistant`.

Kanavoi's [GA4 channel checker](https://kanavoi.com/ga4-kanavatarkistin) already knows this rule. It shows that `chatgpt.com / ai-assistant` lands in the AI Assistant channel, while `chatgpt.com / referral` lands in Referral. When a team creates links for, say, AI tool directories or partners, the checker shows in advance which report row the traffic will end up in.

#### Microsoft Clarity's AI Visibility insights

According to Microsoft's playbook, the AI visibility tools in Microsoft Clarity track three things:

1. **Discoverability (Bot activity, Citations):** Which pages AI bots fetch, how often your content is cited, and which grounding queries bring it into answers.
2. **Visibility (Topic Insights):** Topics where competitors are cited and you are not (*content gaps*), with prioritized recommendations.
3. **Behavior after arrival (AI-referred traffic):** Visitors coming from AI are often at the end of their comparison. If they leave without buying, the reason is often a missing detail, such as a size chart or delivery time. Heatmaps and session recordings show what the visitor was looking for.

Clarity is free and does not replace GA4; it complements it. From a privacy point of view, note that session recordings and heatmaps require a tracking script and, in the EU, usually cookie consent. Kanavoi made a different choice: its visitor statistics are counted on the server without cookies.

---

### 9. A combined strategy for the best AI visibility

`llms.txt` does not work in a vacuum. It is one layer in a four-level whole:

```
+--------------------------------------------------------------+
|                     1. TRADITIONAL SEO                       |
|  - Technical speed, Core Web Vitals and indexing             |
|  - Quality backlinks and brand mentions across the web       |
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|                    2. STRUCTURED DATA                        |
|  - Schema.org (JSON-LD): Organization, Product, Offer,       |
|    Article, FAQPage                                          |
|  - Helps search engines and AI identify entities             |
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|                   3. ANSWER-FIRST CONTENT                    |
|  - Clear definitions, comparison tables and summaries        |
|  - Answer the user's question right at the top of the page   |
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|        4. LLMS.TXT, LLMS-FULL.TXT AND PRODUCT FEED (UCP)     |
|  - Removes HTML and script noise                             |
|  - Gives agents cheap, direct access to the facts            |
+--------------------------------------------------------------+
```

#### Where Kanavoi stands

I am building Kanavoi in public, so here is a status check:

- ✅ **Technical foundation:** A lightweight Go and HTMX site without heavy JavaScript, plus `robots.txt` and a dynamic `sitemap.xml`.
- ✅ **Structured data on guides:** Every guide article has schema.org `Article` and `BreadcrumbList` markup.
- ✅ **Answer-first content:** Each guide answers one question directly, such as "Why does GA4 show traffic as Unassigned?".
- ✅ **Structured data on the landing page:** `Organization`, `WebSite`, `SoftwareApplication` (plans as `Offer` items) and `FAQPage`.
- ✅ **llms.txt, llms-full.txt and `.md` versions of the guides**, all generated from the same sources as the pages.

Google's [Rich Results Test](https://search.google.com/test/rich-results) accepted the organization and the software application on the landing page. The test taught me three things:

1. **What Google accepts is not the same as what schema.org allows.** I first described the prices with a more detailed `UnitPriceSpecification` (monthly price, excluding VAT). Schema.org allows it, but Google's test flagged it as invalid on a software offer. Price and currency are enough, so the billing period and VAT moved into the offer description.
2. **A missing rating is not fixed by inventing one.** The test still notes that `aggregateRating` is missing. Kanavoi has no genuine customer reviews yet, and made-up stars would break Google's guidelines. The rating will be added once there are reviews.
3. **Not everything is shown in search results, and it does not need to be.** `FAQPage` does not appear in the test at all, because Google only shows FAQ rich results for government and health sites. The markup is still useful to other readers, such as AI search.

#### What NOT to put in these files

1. **Cookie policies or general legal jargon:** Do not fill the model's context window with repeated contract terms. A link in the `Optional` section is enough.
2. **Personal data or internal keys:** The file is meant for machines, but it is still a public file on the web.
3. **Low-quality archive content:** Include only content that gives the reader genuinely new information (*information gain*).
4. **Information that contradicts the site:** If `llms.txt` promises a different price than the pricing page, AI cannot know which one to believe. Generate the files from the same source as the pages.

---

### 10. Summary

The move from the human web to the agentic web does not mean the traditional website is dying, but its role is changing. Your website is no longer just a shop window for people; it is also a data source and an interface for AI.

`llms.txt` is not a shortcut to the top of AI search, and none of the major players has promised to read it. It is, however, a cheap and useful part of a whole whose real core is this:

- **Core facts stated clearly in one place:** What you do, for whom, at what price and on what terms.
- **Structured data and an up-to-date product feed:** Agents can verify facts without guessing, which reduces misinterpretation.
- **Measurement:** GA4's AI Assistant channel and AI visibility tools tell you whether any of this is working.

Organizations that get their machine-readable foundation in order now will be ready when agentic commerce expands beyond the US.

---

**Do you run campaigns yourself?** If you want your campaigns to show up in GA4 as a single row in the right channel, whether the traffic comes from paid social, a newsletter or an AI assistant, try [Kanavoi's free UTM generator](https://kanavoi.com/#generaattori) and [GA4 channel checker](https://kanavoi.com/ga4-kanavatarkistin). They work without signing in, and the site is currently in Finnish. You can join the Team and Agency plans through the [waitlist](https://kanavoi.com/#odotuslista), and the first 50 paying customers get a permanent 50% discount. And if you send your own AI agent to check out the service, point it to [kanavoi.com/llms.txt](https://kanavoi.com/llms.txt).

#### Sources

- Microsoft Advertising: [Prepare for the Age of AI Shopping](https://about.ads.microsoft.com/content/dam/sites/msa-about/global/common/content-lib/pdf/prepare-for-the-age-of-ai-shop.pdf) (2026)
- Jeremy Howard / Answer.AI: [The /llms.txt file](https://llmstxt.org) (2024)
- Aggarwal et al.: [GEO: Generative Engine Optimization](https://arxiv.org/abs/2311.09735) (2023, KDD 2024)

---
title: "Building a Modern, SEO-Optimized Blog: Pragmatic Markdown vs. MDX"
description: "A deep technical dive into building high-performance, localized blogs with Next.js, detailing the architectural choice between interactive MDX and pure Markdown."
publishedAt: "2026-06-06"
slug: "modern-ai-era-blog-in-md-or-mdx"
draft: false
tags: ["nextjs", "markdown", "mdx", "seo", "webdev"]
---

In the age of agentic AI, building a blog doesn't require complex CMS setups, heavy databases, or endless security patching. This guide walks you through how to build a highly performant, SEO-optimized, and multilingual blog using Next.js and static files, using real-world architectures I've deployed for [levifinland.fi](https://levifinland.fi) and my own portfolio here at [karotammela.fi](https://karotammela.fi).

Best of all? This architecture treats content as code, making it perfectly optimized for human developers and AI agents alike.

### The Pragmatic Architect: MDX vs. Pure Markdown Simplicity

Before I dive into the code, I must address an important architectural choice: Do you actually need MDX, or is pure Markdown (`.md`) the better choice?

I maintain two different blogging platforms using this static architecture, and I chose different tools for each based on their specific requirements:

#### Case 1: MDX for levifinland.fi/en/blog

For a commercial, high-traffic regional hub like levifinland.fi, conversion and user interaction are key metrics. The blog needs to do more than display text; it must actively drive bookings for golf green fees, and cabin rental.

For this, MDX is indispensable. It allows me to inject complex, interactive React components (like custom dynamic booking widgets, interactive trail maps, or effective call-to-action buttons) directly into standard editorial content. At the moment of writing this blog post, I have utilized CTA buttons in the blog posts to drive traffic to the booking pages.

#### Case 2: Pure Markdown (.md) for karotammela.fi

For my personal developer portfolio, my primary goal was frictionless simplicity. I wanted to focus entirely on writing without worrying about compiling complex React component scopes or debugging broken JSX tags in my content.

By choosing pure Markdown (`.md`), I keep my writing completely decoupled from the frontend framework. If I decide to migrate this site from Next.js to Astro, Hugo, or a custom Rust generator in following years, my `.md` files will parse effortlessly anywhere out of the box. By the way - [Astro](https://astro.build/) is something I really want to dive into during the following months.

I still get beautiful typography using a small custom `.blog-prose` stylesheet (rather than pulling in the Tailwind Typography plugin), but with zero client-side JavaScript overhead and absolute ease of maintenance.

### Architectural Breakdown: Traditional CMS vs. Static MD/MDX

| Feature | Traditional Database / CMS | My Static Setup |
| :--- | :--- | :--- |
| **Hosting Cost** | Server + Database fees ($/mo) | $0 (Static hosting on Vercel/Netlify) |
| **Maintenance** | Security patches, plugin updates | Zero maintenance (It's just files and Git) |
| **AI Readiness** | Complex API schemas or UI clicking | Perfect. LLMs natively read/write Markdown |
| **Performance** | Database queries and server-side render | Blazing-fast pre-rendered static HTML |

### What is MDX (and MD)?

Markdown is a lightweight markup language with plain-text formatting syntax. MDX is a powerful superset of Markdown that lets you write JSX (React components) directly inside your content.

A typical MDX file with custom elements looks like this:

```mdx
---
title: "Levi golf season has started!"
description: "Book your green fees for the 2026 season."
date: "2026-05-30"
author: "Levi Finland"
image: "/images/blog/levi-golf-2026/cover.jpg"
tags: ["golf", "summer"]
draft: false
---

# The Golf Season is Here!

It's finally time to hit the greens. [Book your tee time now](https://greenfee.levifinland.fi).

<CTAButton href="https://greenfee.levifinland.fi">Book Green Fees</CTAButton>
```

The metadata at the top wrapped in `---` is called frontmatter. When I use pure Markdown (`.md`) on a site like karotammela.fi, I write the exact same frontmatter and Markdown, but omit custom React elements like `<CTAButton />` to keep things simple.

### The Tech Stack

To build a production-grade engine around these static files, I use a concise set of libraries:

- **Next.js (App Router):** The React framework handling file-system routing, image optimization, and static rendering.
- **gray-matter:** A parser to separate frontmatter metadata from the body text.
- **Zod:** A TypeScript-first schema validation library used to guarantee frontmatter integrity at build time.
- **next-mdx-remote (or marked for pure .md):** A lightweight utility to safely compile files into HTML elements.
- **next-intl:** A robust localization routing library for seamless multi-language support.

### How It Works

#### 1. The File Structure

I separate my raw content files from the Next.js routing logic. Content lives in a dedicated root directory, organized by slug and locale:

```text
content/blog/
  levi-golf-2026/
    fi.md (or fi.mdx)
    en.md (or en.mdx)
  sarkitunturi-hike/
    fi.md
```

The actual dynamic routing happens inside the Next.js App Router. I create a dynamic folder structure at `app/[locale]/blog/[slug]/page.tsx`. This file acts as the compile-time shell that fetches, validates, and renders the files based on the URL parameters.

#### 2. Reading and Validating Content with Zod

To ensure that human authors or AI writers don't break the production build by forgetting a required field (like a missing cover image or publication date), I enforce a strict schema runtime check using Zod.

Here is how my data-fetching utility (`src/lib/blog.ts`) handles reading and validation:

```typescript
import fs from 'fs';
import matter from 'gray-matter';
import { z } from 'zod';

// Define strict rules for what a blog post's metadata MUST contain
const BlogFrontmatterSchema = z.object({
  title: z.string().min(1),
  description: z.string().max(160), // Optimized for SEO meta descriptions
  date: z.string(),
  author: z.string(),
  image: z.string(),
  tags: z.array(z.string()),
  draft: z.boolean().default(false),
});

export function getPost(slug: string, locale: string) {
  // Check for both .md and .mdx extensions to support hybrid setups
  const baseDir = `content/blog/${slug}`;
  const mdxPath = `${baseDir}/${locale}.mdx`;
  const mdPath = `${baseDir}/${locale}.md`;

  const filePath = fs.existsSync(mdxPath) ? mdxPath : fs.existsSync(mdPath) ? mdPath : null;

  if (!filePath) {
    return null;
  }

  const fileContents = fs.readFileSync(filePath, 'utf8');
  const { data, content } = matter(fileContents);

  // Safely parse and validate frontmatter against the schema
  const validatedFrontmatter = BlogFrontmatterSchema.parse(data);

  return {
    frontmatter: validatedFrontmatter,
    content,
    isMdx: filePath.endsWith('.mdx'),
  };
}
```

#### 3. Static Site Generation (SSG)

To get perfect performance scores, I want Next.js to pre-render every single blog post combination into raw HTML at build time. I achieve this using `generateStaticParams`.

Instead of hardcoding supported languages, I import my active locales directly from my `next-intl` routing configuration to maintain a single source of truth:

```typescript
import { routing } from '@/i18n/routing';
import { getPostSlugs } from '@/lib/blog';

export async function generateStaticParams() {
  const slugs = getPostSlugs(); // Returns e.g., ['levi-golf-2026', 'sarkitunturi-hike']
  const locales = routing.locales; // Returns e.g., ['en', 'fi']

  // Maps into: [{ slug: 'levi-golf-2026', locale: 'en' }, { slug: 'levi-golf-2026', locale: 'fi' }, ...]
  return slugs.flatMap((slug) =>
    locales.map((locale) => ({ slug, locale }))
  );
}
```

When a visitor hits a blog route, the edge network serves a lightweight, fully formed HTML file instantly. No database lookups, no cold starts.

### SEO and GEO (Generative Engine Optimization)

In the evolving landscape of AI search engines (like Perplexity and Google's AI Overviews), clean semantics and explicit data mapping are vital.

#### Structured Data (JSON-LD)

To make it incredibly straightforward for web crawlers and LLM parsers to index my content accurately, I inject a semantic `BlogPosting` schema directly into the document head.

```tsx
export default async function BlogPostPage({ params }: Props) {
  // In Next.js 16, route `params` is a Promise and must be awaited
  const { slug, locale } = await params;
  const post = await getPost(slug, locale);
  if (!post) return notFound();

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    "headline": post.frontmatter.title,
    "description": post.frontmatter.description,
    "datePublished": post.frontmatter.date,
    "image": post.frontmatter.image,
    "author": {
      "@type": "Person",
      "name": post.frontmatter.author,
      "url": "https://karotammela.fi"
    }
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      {/* Article Content */}
    </>
  );
}
```

#### Hreflang and Internationalization

Search engine crawlers penalize poorly mapped localized content. By using Next.js Metadata generation, I declare explicit language alternates, ensuring the correct version is served based on regional intent.

```typescript
export async function generateMetadata({ params }: Props) {
  const baseUrl = 'https://karotammela.fi';
  // In Next.js 16, route `params` is a Promise and must be awaited
  const { slug } = await params;

  return {
    alternates: {
      languages: {
        fi: `${baseUrl}/fi/blog/${slug}`,
        en: `${baseUrl}/en/blog/${slug}`,
        'x-default': `${baseUrl}/en/blog/${slug}`
      }
    }
  };
}
```

### How I Render Content: MDX vs. Pure MD

Depending on whether the file is an `.md` or `.mdx` file, I compile the content differently:

#### Under the Hood of MDX (used for Levi Finland)

For MDX files, I pass a collection of custom React components into my remote renderer shell so authors can write `<CTAButton>` directly in the text editor:

```tsx
import Link from 'next/link';
import { MDXRemote } from 'next-mdx-remote/rsc';

const customComponents = {
  a: ({ href, ...props }: any) => (
    <Link className="text-blue-600 underline hover:text-blue-800" href={href} {...props}/>
  ),
  CTAButton: ({ href, children }: { href: string; children: React.ReactNode }) => (
    <Link className="inline-block bg-blue-600 hover:bg-blue-700 text-white font-semibold px-8 py-4 rounded-xl transition-all transform hover:-translate-y-0.5 shadow-lg" href={href}>
      {children}
    </Link>
  ),
};

export default function PostBody({ content }: { content: string }) {
  return <MDXRemote components={customComponents} source={content}/>;
}
```

#### Under the Hood of Pure MD (used for karotammela.fi)

For pure Markdown files, I don't need any complex dynamic hydration shells or React context. I parse the file with a simple markdown parser (`marked`), then run the output through `sanitize-html` with an explicit allowlist before injecting it inside my custom `.blog-prose` styles. Even though the content is author-trusted, sanitizing is the single source of truth that keeps the markup safe and predictable:

```tsx
import { marked } from 'marked';
import sanitizeHtml from 'sanitize-html';

export default function PostBody({ content }: { content: string }) {
  // marked.parse can return string | Promise<string>, so guard the type
  const rendered = marked.parse(content, { breaks: true, gfm: true });
  const html = typeof rendered === 'string' ? rendered : '';

  const safeHtml = sanitizeHtml(html, {
    allowedTags: sanitizeHtml.defaults.allowedTags.concat(['img', 'pre', 'code']),
  });

  return (
    <div
      className="blog-prose max-w-none"
      dangerouslySetInnerHTML={{ __html: safeHtml }}
    />
  );
}
```

This is elegant, fast, has zero JS hydration overhead, and uses standard web standards.

### The AI Co-Authoring Workflow

Whether you choose Markdown or MDX, shifting away from a graphical database dashboard to a file-based structure radically enhances how you write code and content simultaneously. Here is how an AI assistant accelerates this setup:

- **Automated Boilerplate:** Prompt an LLM to generate complex Zod schema structures or type definitions in seconds.
- **Seamless Localizations:** Instruct an AI agent to read `content/blog/post-slug/en.md` and output a perfectly idiomatic, frontmatter-compliant `fi.md` translation in the exact same directory structure.
- **Markdown is the AI's Native Language:** Large Language Models are trained natively on Markdown syntax. An AI writing agent will never make syntactic layout mistakes when generating standard `.md` files, whereas it can sometimes struggle with unclosed JSX tags in complex MDX setups.

### Summary: Choose What Fits Your Goal

If you are building a commercial product or website that requires highly custom interactive funnels, CTAs, or interactive charts within your content, MDX is the gold standard. It lets you bridge the gap between static content and dynamic apps.

But if you are building a personal site, technical journal, or thought-leadership blog, do not over-engineer it. Pure Markdown is a developer superpower. It gives you portability, absolute simplicity, and blistering performance without the baggage of complex compilation chains.

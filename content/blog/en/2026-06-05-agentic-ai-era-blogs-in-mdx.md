---
title: "The Agentic AI Era: Writing, Running, and Hosting Blogs in MDX"
description: "How the shift from generative chatbots to autonomous agentic systems is redefining the content management lifecycle, local file-based Markdown, and production workflows in 2026."
publishedAt: "2026-06-05"
slug: "agentic-ai-era-blogs-in-mdx"
draft: false
tags: ["AI", "Agentic-AI", "MDX", "Next.js", "Kilo-Code", "Web-Development"]
---

We have crossed a fundamental threshold in how software, websites, and content are built. In 2026, the transition from simple generative AI (chatbots that respond to queries) to fully **agentic AI** (autonomous systems that reason, plan, and take real actions) is complete.

This very website was scaffolded and deployed in a matter of hours, but as the project grew, so did the need for a robust and developer-centric way to manage content. Enter file-based Markdown and MDX. Instead of logging into a heavy, siloed database or a complex headless CMS, modern developers are managing their content right alongside their codebase.

![Kilo Code executing tasks in the workspace](/media/karo-tammela-agentic-ai.png)

### The Shift: Generative vs. Agentic

In the chatbot era, you would ask an AI to write a paragraph, copy it, open your editor, format it manually, and commit it. In the **agentic era**, you describe a goal: *"Write a technical blog post about the agentic stack, format it according to our project's frontmatter schema, verify its local code snippets work, and prepare it for commit."*

The agent doesn't just draft the text; it explores your filesystem, reads your parsing validation schemas (such as Zod files or gray-matter loaders), checks image paths, and modifies the workspace directly.

Let's compare the capabilities:

| Feature / Capability | Chatbot Era (Generative) | Agentic Era (Autonomous) |
| :--- | :--- | :--- |
| **Interaction Model** | Single-turn prompts & answers | Goal-driven iterative loop |
| **Task Complexity** | Basic drafting & suggestions | Multi-step reasoning & workflow |
| **Tool Access** | None or restricted chat plugins | Direct filesystem, local CLI, web searching |
| **Execution** | Human must copy-paste | Agent writes, tests, and verifies code |

### Why This Blog Runs on Plain Markdown, Not MDX

I'm using this site as the example because I built it myself, so I can show you the real plumbing rather than a generic tutorial. For lightweight portfolios, personal websites, or docs, a database-driven CMS is overkill. Storing each post as a raw `.md` file in a folder like `content/blog/en/` keeps the entire blog in git, versioned right alongside the application code.

I deliberately chose plain `.md` over `.mdx` here. MDX is powerful, because it lets you embed live React components directly inside your prose, but that power has a price: every post effectively becomes executable code that must be compiled and trusted. For a writing-first blog I don't need interactive widgets in the middle of a sentence; I need fast, predictable, portable text. Plain Markdown keeps the goal simple: write prose, add frontmatter, commit. In fact the loader in this project only reads `.md` files and ignores everything else.

You can see the opposite tradeoff in my own projects. The [Levi Golf 2026 season post](https://levifinland.fi/en/blog/levi-golf-season-2026-has-started) on the Levi Finland blog runs on MDX, not plain Markdown, and that is on purpose. Those posts need clear call-to-action buttons, so I built one reusable CTA component. To add a button to a new post, I just call the component and change its text. That is exactly the job MDX is built for. This portfolio blog has no such need, so plain `.md` stays the simpler, safer choice here.

### Anatomy of a Post File

A typical blog post file on this site looks like this, nothing more than a frontmatter block followed by a Markdown body:

```markdown
---
title: "The Agentic AI Era: Writing, Running, and Hosting Blogs"
description: "How agentic AI is reshaping the content lifecycle in 2026."
publishedAt: "2026-06-05"
slug: "agentic-ai-era-blogs-in-mdx"
draft: false
tags: ["AI", "Agentic-AI", "Markdown", "Next.js"]
---

We have crossed a fundamental threshold in how software,
websites, and content are built...
```

But how does the Next.js application turn that raw file into a safe, typed object? Here is the actual parsing logic from this project:

```typescript
// src/lib/blog.ts
import matter from "gray-matter";
import { z } from "zod";

const blogFrontmatterSchema = z.object({
  title: z.string().min(1),
  description: z.string().min(1),
  publishedAt: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
  slug: z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/),
  draft: z.boolean().optional().default(false),
  tags: z.array(z.string().min(1)).optional().default([]),
});

function parsePostFile(filePath: string, source: string) {
  const { data, content } = matter(source);
  const parsed = blogFrontmatterSchema.parse(data);
  return {
    ...parsed,
    body: content.trim(),
  };
}
```

`gray-matter` splits the frontmatter from the body, and Zod validates that metadata against a strict schema. If a date is malformed or a slug contains illegal characters, the build fails loudly instead of shipping a broken post. This gives total control over metadata (titles, publishing dates, tags) while keeping the content itself as readable, portable Markdown.

### Visual Styling for Code Blocks

A wall of monospace text is functional, but it isn't inviting. To make snippets actually pull readers in, I dressed every fenced code block on this site in a faux **terminal window**: a dark title bar topped with the three classic macOS traffic-light dots. No syntax-highlighting library and no extra dependency, just two CSS pseudo-elements targeting `.blog-prose pre` in the global stylesheet:

```css
/* Faux terminal-window chrome for blog code blocks */
.blog-prose pre {
  position: relative;
  background-color: #0d0f16;
  border: 1px solid var(--border);
  border-radius: 0.75rem;
  padding: 2.85rem 1.15rem 1.15rem; /* top room for the title bar */
  overflow-x: auto;
}

/* The dark title bar */
.blog-prose pre::before {
  content: "";
  position: absolute;
  inset: 0 0 auto 0;
  height: 2.05rem;
  background: linear-gradient(180deg, #1b1f2a, #14171f);
}

/* Three traffic-light dots, drawn with a single box-shadow */
.blog-prose pre::after {
  content: "";
  position: absolute;
  top: 0.72rem;
  left: 1.1rem;
  width: 0.66rem;
  height: 0.66rem;
  border-radius: 50%;
  background: #ff5f56;
  box-shadow: 1.15rem 0 0 #ffbd2e, 2.3rem 0 0 #27c93f;
}
```

The trick is doing the whole thing with `::before` and `::after`. The first pseudo-element paints the dark title bar, and the second draws the red dot, then conjures the amber and green ones out of thin air using two offset `box-shadow` copies. The result is a high-contrast, modern terminal look that matches the visual aesthetics of an engineering dashboard, and it's exactly what renders the code blocks you see throughout this post.

### The Role of Agents in Content Lifecycles

In 2026, content is no longer static. An agentic AI can regularly run automated checks across your entire library of blog posts. A typical maintenance loop looks like this:

1. **Lint the Markdown structure** to catch broken links and malformed frontmatter before they ship.
2. **Translate drafts** between languages automatically, while keeping slugs, tags, and formatting intact.
3. **Verify API freshness** by spotting an import from a library that was just updated, checking the changelog, patching the code block, and test-compiling it locally.

This integration of agentic workflows turns a simple blog folder into a living, self-healing knowledge hub. By leveraging tools like Kilo Code, developers stay focused on building features while the autonomous loop handles the minutiae of formatting, layout checking, and repository management.

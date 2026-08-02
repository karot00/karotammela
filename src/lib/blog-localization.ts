import type { BlogLocale } from "@/lib/blog";

const translatedSlugs: Record<BlogLocale, Record<string, string>> = {
  en: {
    "how-ai-unlocked-my-coding": "tekoaly-avasi-koodaukseni-lukot",
    "vibe-coding-vs-production-ready": "vibe-coding-vs-tuotantovalmis",
    "mcp-bridging-the-knowledge-gap":
      "mcp-bridgin-the-knowledge-gap",
    "whatsapp-chaos-to-automated-process-levi-golf":
      "whatsapp-kaoottisuudesta-automatisoituun-prosessiin-levi-golf",
    "agentic-ai-era-blogs-in-mdx": "agenttisen-tekoalyn-aikakausi-blogit-markdownilla",
    "modern-ai-era-blog-in-md-or-mdx": "moderni-seo-blogi-markdown-vai-mdx",
    "go-htmx-vs-nextjs-hetzner": "go-htmx-nextjs-vertailu-hetzner",
    "how-i-moved-wordpress-site-to-next-js":
      "miten-siirsin-wordpress-sivuston-next-js-teknologiaan",
  },
  fi: {
    "tekoaly-avasi-koodaukseni-lukot": "how-ai-unlocked-my-coding",
    "vibe-coding-vs-tuotantovalmis": "vibe-coding-vs-production-ready",
    "mcp-bridgin-the-knowledge-gap":
      "mcp-bridging-the-knowledge-gap",
    "whatsapp-kaoottisuudesta-automatisoituun-prosessiin-levi-golf":
      "whatsapp-chaos-to-automated-process-levi-golf",
    "agenttisen-tekoalyn-aikakausi-blogit-markdownilla":
      "agentic-ai-era-blogs-in-mdx",
    "moderni-seo-blogi-markdown-vai-mdx": "modern-ai-era-blog-in-md-or-mdx",
    "go-htmx-nextjs-vertailu-hetzner": "go-htmx-vs-nextjs-hetzner",
    "miten-siirsin-wordpress-sivuston-next-js-teknologiaan":
      "how-i-moved-wordpress-site-to-next-js",
  },
};

export function getTranslatedBlogSlug(locale: BlogLocale, slug: string): string {
  return translatedSlugs[locale][slug] ?? slug;
}

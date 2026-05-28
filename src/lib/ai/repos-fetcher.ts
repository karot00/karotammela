import { google } from "@ai-sdk/google";
import { generateText } from "ai";
import * as cheerio from "cheerio";

import type { NewAiRepo } from "@/lib/db/schema";

const AI_ML_KEYWORDS = [
  "ai",
  "llm",
  "agent",
  "ml",
  "machine learning",
  "deep learning",
  "rag",
  "mcp",
  "transformer",
  "pytorch",
  "tensorflow",
  "inference",
  "embedding",
  "neural",
  "vision",
  "llama",
  "claude",
  "gpt",
  "gemini",
  "openai",
  "copilot",
  "nlp",
  "diffusers",
  "stable-diffusion",
  "midjourney",
];

function getModelName(): string {
  return process.env.AI_MODEL ?? "gemini-3.1-flash-lite";
}

export function isAiMlRelated(title: string, description: string): boolean {
  const text = `${title} ${description}`.toLowerCase();
  return AI_ML_KEYWORDS.some((keyword) => {
    const regex = new RegExp(`\\b${keyword}\\b`, "i");
    return regex.test(text);
  });
}

export interface GitHubRepo {
  repoFullName: string;
  url: string;
  description: string;
  language: string | null;
  stars: number;
  starsToday: number;
}

export function parseGitHubTrending(html: string): GitHubRepo[] {
  const $ = cheerio.load(html);
  const repos: GitHubRepo[] = [];

  $("article.Box-row").each((_, element) => {
    const $el = $(element);

    const repoLink = $el.find("h2 a").attr("href")?.trim() || "";
    if (!repoLink) return;

    const parts = repoLink.split("/").filter(Boolean);
    if (parts.length < 2) return;
    const owner = parts[0];
    const name = parts[1];
    const repoFullName = `${owner}/${name}`;

    const description = $el.find("p.col-9, p.col-12").text().trim() || "";

    const language = $el.find('[itemprop="programmingLanguage"]').text().trim() || null;

    const starsText = $el.find('a[href$="/stargazers"]').text().trim().replace(/,/g, "");
    const stars = Number.parseInt(starsText, 10) || 0;

    const starsTodayText = $el.find(".float-sm-right, .d-inline-block.float-sm-right").text().trim();
    const starsTodayMatch = starsTodayText.match(/([\d,]+)\s*stars?/i);
    const starsToday = starsTodayMatch ? Number.parseInt(starsTodayMatch[1].replace(/,/g, ""), 10) : 0;

    repos.push({
      repoFullName,
      url: `https://github.com${repoLink}`,
      description,
      language,
      stars,
      starsToday,
    });
  });

  return repos;
}

async function translateDescription(description: string): Promise<string> {
  if (!description) return "";
  try {
    const result = await generateText({
      model: google(getModelName()),
      prompt: `Käännä seuraava GitHub-repositorion kuvaus suomeksi.

Ohjeet:
- Anna suoraan vain se yksi lopullinen käännös.
- ÄLÄ missään tapauksessa anna useita vaihtoehtoja tai selityksiä (kuten "Tässä muutama vaihtoehto:").
- Pidä käännös erittäin ytimekkäänä (maksimissaan 1 lause) ja kehittäjille luontevana.

Kuvaus: "${description}"`,
      maxOutputTokens: 60,
    });
    return result.text.trim();
  } catch (err) {
    console.warn("[repos-fetcher] Gemini translation error", err);
    return description;
  }
}

export async function fetchAndSummarizeRepos(): Promise<NewAiRepo[]> {
  const urls = [
    "https://github.com/trending?since=daily",
    "https://github.com/trending/python?since=daily",
    "https://github.com/trending/jupyter-notebook?since=daily",
  ];

  const fetchResults = await Promise.allSettled(
    urls.map((url) =>
      fetch(url, {
        headers: {
          "User-Agent":
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        },
        next: { revalidate: 0 },
      }).then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.text();
      }),
    ),
  );

  const allReposMap = new Map<string, GitHubRepo>();

  for (const res of fetchResults) {
    if (res.status === "fulfilled") {
      const parsed = parseGitHubTrending(res.value);
      for (const r of parsed) {
        allReposMap.set(r.repoFullName, r);
      }
    } else {
      console.warn("[repos-fetcher] Failed to fetch trending page:", res.reason);
    }
  }

  const allRepos = Array.from(allReposMap.values());

  const filtered = allRepos.filter((r) => isAiMlRelated(r.repoFullName, r.description));

  const sorted = filtered.sort((a, b) => b.starsToday - a.starsToday);

  const MAX_REPOS = 7;
  const selected = sorted.slice(0, MAX_REPOS);

  const today = new Date().toISOString().slice(0, 10);
  const resultRepos: NewAiRepo[] = [];

  for (const r of selected) {
    const descriptionFi = await translateDescription(r.description);
    resultRepos.push({
      id: crypto.randomUUID(),
      date: today,
      repoFullName: r.repoFullName,
      url: r.url,
      description: r.description,
      descriptionFi,
      language: r.language,
      stars: r.stars,
      starsToday: r.starsToday,
      source: "github-trending",
    });
  }

  return resultRepos;
}

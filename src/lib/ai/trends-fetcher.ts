import { google } from "@ai-sdk/google";
import { generateText } from "ai";

import type { NewAiTrend } from "@/lib/db/schema";

type HnHit = {
  title: string;
  url: string;
  points: number;
  objectID: string;
};

type HnSearchResponse = {
  hits: HnHit[];
};

function getModelName(): string {
  return process.env.AI_MODEL ?? "gemini-3.1-flash-lite-preview";
}

async function summarizeStory(title: string, url: string): Promise<string> {
  try {
    const { text } = await generateText({
      model: google(getModelName()),
      prompt: `Summarize this tech news story in 1–2 sentences for a developer audience. Be concise and factual. Title: "${title}". URL: ${url}`,
      maxOutputTokens: 120,
    });

    return text.trim();
  } catch (err) {
    console.warn("[trends-fetcher] Gemini summarize error", err);
    return title;
  }
}

export async function fetchAndSummarizeTrends(): Promise<NewAiTrend[]> {
  const yesterday = Math.floor(Date.now() / 1000) - 86400;
  const hnUrl = `https://hn.algolia.com/api/v1/search?query=AI+machine+learning&tags=story&numericFilters=created_at_i>${yesterday}&hitsPerPage=7`;

  let hits: HnHit[] = [];

  try {
    const res = await fetch(hnUrl, { next: { revalidate: 0 } });

    if (!res.ok) {
      console.warn(`[trends-fetcher] HN API HTTP ${res.status}`);
      return [];
    }

    const data = (await res.json()) as HnSearchResponse;
    hits = data.hits ?? [];
  } catch (err) {
    console.warn("[trends-fetcher] HN API fetch error", err);
    return [];
  }

  if (hits.length === 0) {
    return [];
  }

  const today = new Date().toISOString().slice(0, 10);

  // Sort by points descending and take top 7
  const sorted = [...hits]
    .sort((a, b) => (b.points ?? 0) - (a.points ?? 0))
    .slice(0, 7);

  const trends: NewAiTrend[] = [];

  for (const hit of sorted) {
    if (!hit.title || !hit.url) continue;

    const summary = await summarizeStory(hit.title, hit.url);

    trends.push({
      id: crypto.randomUUID(),
      date: today,
      title: hit.title,
      summary,
      url: hit.url,
      source: "hackernews",
    });
  }

  return trends;
}

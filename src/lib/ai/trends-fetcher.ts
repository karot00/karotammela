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

async function summarizeStory(
  title: string,
  url: string,
): Promise<{ en: string; fi: string }> {
  try {
    const [enResult, fiResult] = await Promise.all([
      generateText({
        model: google(getModelName()),
        prompt: `Summarize this tech news story in 1–2 sentences for a developer audience. Be concise and factual. Title: "${title}". URL: ${url}`,
        maxOutputTokens: 120,
      }),
      generateText({
        model: google(getModelName()),
        prompt: `Tiivistä tämä teknologiauutinen 1–2 lauseeseen kehittäjäyleisölle. Ole ytimekäs ja asiallinen. Vastaa suomeksi. Otsikko: "${title}". URL: ${url}`,
        maxOutputTokens: 120,
      }),
    ]);

    return { en: enResult.text.trim(), fi: fiResult.text.trim() };
  } catch (err) {
    console.warn("[trends-fetcher] Gemini summarize error", err);
    return { en: title, fi: title };
  }
}

export async function fetchAndSummarizeTrends(): Promise<NewAiTrend[]> {
  // Fetch top AI stories by points — no date filter (numericFilters is unreliable
  // on the Algolia HN API). We use search_by_date endpoint and sort client-side.
  const hnUrl =
    "https://hn.algolia.com/api/v1/search_by_date?query=AI+LLM+language+model+machine+learning&tags=story&hitsPerPage=50";

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

  // Sort by points descending, take top 7 with a valid URL
  const sorted = [...hits]
    .filter((h) => h.title && h.url)
    .sort((a, b) => (b.points ?? 0) - (a.points ?? 0))
    .slice(0, 7);

  const trends: NewAiTrend[] = [];

  for (const hit of sorted) {
    const summary = await summarizeStory(hit.title, hit.url);

    trends.push({
      id: crypto.randomUUID(),
      date: today,
      title: hit.title,
      summary: summary.en,
      summaryFi: summary.fi,
      url: hit.url,
      source: "hackernews",
    });
  }

  return trends;
}

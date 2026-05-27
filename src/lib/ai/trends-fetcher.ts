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
  return process.env.AI_MODEL ?? "gemini-3.1-flash-lite";
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

export async function fetchAndSummarizeTrends(
  excludeUrls: Set<string> = new Set(),
): Promise<{
  trends: NewAiTrend[];
  stats: { windowHours: number; candidatePool: number; dedupedOut: number };
}> {
  const windows = [24, 48, 72, 168]; // 24h, 48h, 72h, 7d
  let trends: NewAiTrend[] = [];
  let finalWindowHours = 24;
  let finalCandidatePool = 0;
  let finalDedupedOut = 0;

  for (const windowHours of windows) {
    finalWindowHours = windowHours;
    const nowSecs = Math.floor(Date.now() / 1000);
    const minTimestamp = nowSecs - windowHours * 3600;

    const hnUrl = `https://hn.algolia.com/api/v1/search_by_date?query=AI+LLM+language+model+machine+learning&tags=story&numericFilters=created_at_i>${minTimestamp}&hitsPerPage=50`;

    let hits: HnHit[] = [];

    try {
      const res = await fetch(hnUrl, { next: { revalidate: 0 } });

      if (!res.ok) {
        console.warn(`[trends-fetcher] HN API HTTP ${res.status} for ${windowHours}h window`);
        continue;
      }

      const data = (await res.json()) as HnSearchResponse;
      hits = data.hits ?? [];
    } catch (err) {
      console.warn(`[trends-fetcher] HN API fetch error for ${windowHours}h window`, err);
      continue;
    }

    if (hits.length === 0) {
      continue;
    }

    finalCandidatePool = hits.length;

    // Filter by points descending, unique URLs, and excludeUrls
    const uniqueHits: HnHit[] = [];
    const seenUrlsInBatch = new Set<string>();

    const sortedHits = [...hits]
      .filter((h) => h.title && h.url)
      .sort((a, b) => (b.points ?? 0) - (a.points ?? 0));

    let localDedupedCount = 0;

    for (const hit of sortedHits) {
      if (excludeUrls.has(hit.url) || seenUrlsInBatch.has(hit.url)) {
        localDedupedCount++;
        continue;
      }
      seenUrlsInBatch.add(hit.url);
      uniqueHits.push(hit);
    }

    finalDedupedOut = localDedupedCount;

    if (uniqueHits.length >= 5 || windowHours === 168) {
      const today = new Date().toISOString().slice(0, 10);
      const selectedHits = uniqueHits.slice(0, 7);

      for (const hit of selectedHits) {
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
      break;
    }
  }

  return {
    trends,
    stats: {
      windowHours: finalWindowHours,
      candidatePool: finalCandidatePool,
      dedupedOut: finalDedupedOut,
    },
  };
}

import { NextRequest, NextResponse } from "next/server";

import { fetchStockHistory } from "@/lib/ai/stocks-fetcher";
import { fetchAndSummarizeTrends } from "@/lib/ai/trends-fetcher";
import { fetchAndSummarizeRepos } from "@/lib/ai/repos-fetcher";
import { getRecentTrendUrls, upsertStocks, upsertTrends, upsertRepos } from "@/lib/db/queries";

export const runtime = "nodejs";

async function handleRefresh(req: NextRequest): Promise<NextResponse> {
  const authHeader = req.headers.get("authorization");
  const expectedToken = process.env.CRON_SECRET;

  if (!expectedToken || authHeader !== `Bearer ${expectedToken}`) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const url = new URL(req.url);
  const isDebug = url.searchParams.get("debug") === "1";

  const results = {
    trends: { ok: false, inserted: 0, skippedDup: 0, windowHours: 24, error: null as string | null },
    repos: { ok: false, inserted: 0, error: null as string | null },
    stocks: { ok: false, inserted: 0, error: null as string | null },
  };

  try {
    let recentTrendUrls = new Set<string>();
    try {
      recentTrendUrls = await getRecentTrendUrls(7);
    } catch (err) {
      console.error("[ai-pulse/refresh] Failed to fetch recent trend URLs:", err);
    }

    const [trendsRes, reposRes, stocksResult] = await Promise.allSettled([
      fetchAndSummarizeTrends(recentTrendUrls),
      fetchAndSummarizeRepos(),
      fetchStockHistory(),
    ]);

    if (trendsRes.status === "fulfilled") {
      const { trends, stats } = trendsRes.value;
      try {
        await upsertTrends(trends);
        results.trends = {
          ok: true,
          inserted: trends.length,
          skippedDup: stats.dedupedOut,
          windowHours: stats.windowHours,
          error: null,
        };
      } catch (err) {
        console.error("[ai-pulse/refresh] Failed to upsert trends:", err);
        results.trends.error = isDebug ? String(err) : "DB insert failed";
      }
    } else {
      console.error("[ai-pulse/refresh] Trends fetcher failed:", trendsRes.reason);
      results.trends.error = isDebug ? String(trendsRes.reason) : "Fetch failed";
    }

    if (reposRes.status === "fulfilled") {
      const repos = reposRes.value;
      try {
        await upsertRepos(repos);
        results.repos = {
          ok: true,
          inserted: repos.length,
          error: null,
        };
      } catch (err) {
        console.error("[ai-pulse/refresh] Failed to upsert repos:", err);
        results.repos.error = isDebug ? String(err) : "DB insert failed";
      }
    } else {
      console.error("[ai-pulse/refresh] Repos fetcher failed:", reposRes.reason);
      results.repos.error = isDebug ? String(reposRes.reason) : "Fetch failed";
    }

    if (stocksResult.status === "fulfilled") {
      const stocks = stocksResult.value;
      try {
        await upsertStocks(stocks);
        results.stocks = {
          ok: true,
          inserted: stocks.length,
          error: null,
        };
      } catch (err) {
        console.error("[ai-pulse/refresh] Failed to upsert stocks:", err);
        results.stocks.error = isDebug ? String(err) : "DB insert failed";
      }
    } else {
      console.error("[ai-pulse/refresh] Stocks fetcher failed:", stocksResult.reason);
      results.stocks.error = isDebug ? String(stocksResult.reason) : "Fetch failed";
    }

    console.log(
      `[ai-pulse/refresh] Run complete. trends_inserted=${results.trends.inserted} trends_skipped=${results.trends.skippedDup} trends_window=${results.trends.windowHours}h repos_inserted=${results.repos.inserted} stocks_inserted=${results.stocks.inserted}`,
    );

    return NextResponse.json({
      ok: true,
      ranAt: new Date().toISOString(),
      sources: results,
    });
  } catch (err) {
    console.error("[ai-pulse/refresh] Critical crash:", err);
    return NextResponse.json(
      {
        error: "Internal server error",
        message: isDebug ? String(err) : undefined,
      },
      { status: 500 },
    );
  }
}

export async function GET(req: NextRequest): Promise<NextResponse> {
  return handleRefresh(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return handleRefresh(req);
}

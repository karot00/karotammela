import { NextRequest, NextResponse } from "next/server";

import { fetchStockHistory } from "@/lib/ai/stocks-fetcher";
import { fetchAndSummarizeTrends } from "@/lib/ai/trends-fetcher";
import { upsertStocks, upsertTrends } from "@/lib/db/queries";

export const runtime = "nodejs";

export async function POST(req: NextRequest): Promise<NextResponse> {
  const authHeader = req.headers.get("authorization");
  const expectedToken = process.env.CRON_SECRET;

  if (!expectedToken || authHeader !== `Bearer ${expectedToken}`) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  try {
    const [trends, stocks] = await Promise.all([
      fetchAndSummarizeTrends(),
      fetchStockHistory(),
    ]);

    await Promise.all([upsertTrends(trends), upsertStocks(stocks)]);

    return NextResponse.json({
      ok: true,
      trends: trends.length,
      stocks: stocks.length,
    });
  } catch (err) {
    console.error("[ai-pulse/refresh] Error:", err);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 },
    );
  }
}

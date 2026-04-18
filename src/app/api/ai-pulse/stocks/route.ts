import { NextRequest, NextResponse } from "next/server";

import { AI_TICKERS } from "@/lib/ai/stocks-fetcher";
import { getStockHistory } from "@/lib/db/queries";

export const revalidate = 3600;

const VALID_TICKERS: ReadonlySet<string> = new Set<string>(
  AI_TICKERS.map((t) => t.ticker),
);

export async function GET(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const ticker = searchParams.get("ticker");

  if (!ticker || !VALID_TICKERS.has(ticker)) {
    return NextResponse.json({ error: "Invalid ticker" }, { status: 400 });
  }

  const entry = AI_TICKERS.find((t) => t.ticker === ticker);
  const companyName = entry?.name ?? ticker;

  try {
    const data = await getStockHistory(ticker, 365);

    return NextResponse.json({ ticker, companyName, data });
  } catch (err) {
    console.error("[ai-pulse/stocks] Error:", err);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 },
    );
  }
}

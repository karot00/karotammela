import type { NewAiStock } from "@/lib/db/schema";

export const AI_TICKERS = [
  { ticker: "NVDA", name: "NVIDIA" },
  { ticker: "MSFT", name: "Microsoft" },
  { ticker: "GOOGL", name: "Alphabet" },
  { ticker: "META", name: "Meta" },
  { ticker: "AMD", name: "AMD" },
  { ticker: "ARM", name: "Arm Holdings" },
  { ticker: "AVGO", name: "Broadcom" },
  { ticker: "CRWV", name: "CoreWeave" },
  { ticker: "PLTR", name: "Palantir" },
  { ticker: "AI", name: "C3.ai" },
  { ticker: "SNOW", name: "Snowflake" },
  { ticker: "SOUN", name: "SoundHound AI" },
] as const;

type TickerEntry = (typeof AI_TICKERS)[number];

type YahooQuoteResult = {
  timestamp: number[];
  indicators: {
    quote: Array<{
      open: (number | null)[];
      high: (number | null)[];
      low: (number | null)[];
      close: (number | null)[];
      volume: (number | null)[];
    }>;
  };
};

type YahooChartResponse = {
  chart: {
    result: YahooQuoteResult[] | null;
    error: unknown;
  };
};

async function fetchTickerHistory(entry: TickerEntry): Promise<NewAiStock[]> {
  const url = `https://query1.finance.yahoo.com/v8/finance/chart/${entry.ticker}?range=1y&interval=1d&includePrePost=false`;

  let data: YahooChartResponse;
  try {
    const res = await fetch(url, {
      headers: { "User-Agent": "Mozilla/5.0" },
      next: { revalidate: 0 },
    });

    if (!res.ok) {
      console.warn(`[stocks-fetcher] ${entry.ticker}: HTTP ${res.status}`);
      return [];
    }

    data = (await res.json()) as YahooChartResponse;
  } catch (err) {
    console.warn(`[stocks-fetcher] ${entry.ticker}: fetch error`, err);
    return [];
  }

  const result = data?.chart?.result?.[0];
  if (!result) {
    console.warn(`[stocks-fetcher] ${entry.ticker}: no result in response`);
    return [];
  }

  const { timestamp, indicators } = result;
  const quote = indicators?.quote?.[0];

  if (!quote || !timestamp?.length) {
    return [];
  }

  const rows: NewAiStock[] = [];

  for (let i = 0; i < timestamp.length; i++) {
    const close = quote.close[i];

    if (close === null || close === undefined) continue;

    const open = quote.open[i] ?? close;
    const high = quote.high[i] ?? close;
    const low = quote.low[i] ?? close;
    const volume = quote.volume[i] ?? null;

    const date = new Date(timestamp[i] * 1000).toISOString().slice(0, 10);

    rows.push({
      id: crypto.randomUUID(),
      date,
      ticker: entry.ticker,
      companyName: entry.name,
      open,
      high,
      low,
      close,
      volume,
    });
  }

  return rows;
}

export async function fetchStockHistory(): Promise<NewAiStock[]> {
  const results: NewAiStock[] = [];

  for (const entry of AI_TICKERS) {
    const rows = await fetchTickerHistory(entry);
    results.push(...rows);
  }

  return results;
}

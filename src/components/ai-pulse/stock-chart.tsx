"use client";

import { useState } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import type { AiStock } from "@/lib/db/schema";

type StockChartCopy = {
  tickerLabel: string;
  loadingLabel: string;
};

type StockChartProps = {
  initialTicker: string;
  initialData: AiStock[];
  tickers: string[];
  copy: StockChartCopy;
};

type ChartDataPoint = {
  date: string;
  close: number;
  changePercent: number | null;
};

function buildChartData(data: AiStock[]): ChartDataPoint[] {
  const firstClose = data[0]?.close ?? null;

  return data.map((row) => ({
    date: row.date,
    close: row.close,
    changePercent:
      firstClose && firstClose !== 0
        ? ((row.close - firstClose) / firstClose) * 100
        : null,
  }));
}

function formatMonthLabel(dateStr: string): string {
  const parsed = new Date(dateStr + "T00:00:00Z");

  if (Number.isNaN(parsed.getTime())) return dateStr;

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  }).format(parsed);
}

function formatCompactNumber(value: number): string {
  if (value >= 1000) {
    return `$${(value / 1000).toFixed(1)}k`;
  }

  return `$${value.toFixed(2)}`;
}

type TooltipPayloadItem = {
  value: number;
  payload: ChartDataPoint;
};

function ChartTooltip({
  active,
  payload,
  label,
}: {
  active?: boolean;
  payload?: TooltipPayloadItem[];
  label?: string;
}): React.ReactElement | null {
  if (!active || !payload?.length) return null;

  const item = payload[0];
  const changePercent = item?.payload.changePercent;

  return (
    <div className="rounded-lg border border-border bg-card px-3 py-2 text-sm shadow">
      <p className="font-medium text-foreground">{label}</p>
      <p className="text-foreground">
        Close: <span className="font-semibold">${item?.value.toFixed(2)}</span>
      </p>
      {changePercent !== null && changePercent !== undefined ? (
        <p
          className={changePercent >= 0 ? "text-emerald-400" : "text-rose-400"}
        >
          {changePercent >= 0 ? "+" : ""}
          {changePercent.toFixed(2)}% from start
        </p>
      ) : null}
    </div>
  );
}

export function StockChart({
  initialTicker,
  initialData,
  tickers,
  copy,
}: StockChartProps): React.ReactElement {
  const [ticker, setTicker] = useState(initialTicker);
  const [data, setData] = useState<AiStock[]>(initialData);
  const [loading, setLoading] = useState(false);

  const onTickerChange = async (
    e: React.ChangeEvent<HTMLSelectElement>,
  ): Promise<void> => {
    const next = e.target.value;

    if (next === ticker) return;

    setTicker(next);
    setLoading(true);

    try {
      const res = await fetch(`/api/ai-pulse/stocks?ticker=${next}`);

      if (res.ok) {
        const json = (await res.json()) as { data: AiStock[] };
        setData(json.data);
      }
    } catch {
      // Silently fall back to existing data
    } finally {
      setLoading(false);
    }
  };

  const chartData = buildChartData(data);

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <label
          htmlFor="ticker-select"
          className="text-xs font-semibold text-muted-foreground"
        >
          {copy.tickerLabel}
        </label>
        <select
          id="ticker-select"
          value={ticker}
          onChange={onTickerChange}
          className="rounded-md border border-border bg-background px-2 py-1 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
        >
          {tickers.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
      </div>

      {loading ? (
        <div className="h-64 animate-pulse rounded-xl bg-muted/40" />
      ) : (
        <ResponsiveContainer width="100%" height={256}>
          <AreaChart
            data={chartData}
            margin={{ top: 8, right: 8, bottom: 0, left: 0 }}
          >
            <defs>
              <linearGradient id="stockGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#00d2ff" stopOpacity={0.2} />
                <stop offset="100%" stopColor="#00d2ff" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" opacity={0.1} />
            <XAxis
              dataKey="date"
              tickFormatter={formatMonthLabel}
              tick={{ fontSize: 11, fill: "hsl(var(--muted-foreground))" }}
              tickLine={false}
              axisLine={false}
              interval="preserveStartEnd"
            />
            <YAxis
              tickFormatter={formatCompactNumber}
              tick={{ fontSize: 11, fill: "hsl(var(--muted-foreground))" }}
              tickLine={false}
              axisLine={false}
              width={56}
            />
            <Tooltip content={<ChartTooltip />} />
            <Area
              type="monotone"
              dataKey="close"
              stroke="#00d2ff"
              fill="url(#stockGrad)"
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4, fill: "#00d2ff" }}
            />
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}

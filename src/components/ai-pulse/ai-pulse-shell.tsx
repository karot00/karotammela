import type { AiStock, AiTrend } from "@/lib/db/schema";

import { StockChart } from "./stock-chart";
import { TrendsList } from "./trends-list";

export type AiPulseCopy = {
  aiPulseTitle: string;
  aiPulseDescription: string;
  aiPulseTrendsTitle: string;
  aiPulseStocksTitle: string;
  aiPulseTickerLabel: string;
  aiPulseNoTrendsLabel: string;
  aiPulseLastUpdatedLabel: string;
  aiPulseSourceLabel: string;
  aiPulseLoadingLabel: string;
};

export type AiPulseData = {
  trends: AiTrend[];
  initialTicker: string;
  initialStockData: AiStock[];
  availableTickers: string[];
};

type AiPulseShellProps = {
  data: AiPulseData;
  copy: AiPulseCopy;
};

export function AiPulseShell({
  data,
  copy,
}: AiPulseShellProps): React.ReactElement {
  const displayTickers =
    data.availableTickers.length > 0
      ? data.availableTickers
      : [data.initialTicker];

  return (
    <section className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold text-foreground">
          {copy.aiPulseTitle}
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          {copy.aiPulseDescription}
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-xl border border-border bg-card p-5">
          <h3 className="flex items-center gap-2 text-xs font-semibold tracking-[0.16em] text-muted-foreground uppercase after:h-px after:flex-1 after:bg-border">
            {copy.aiPulseTrendsTitle}
          </h3>
          <div className="mt-4">
            <TrendsList
              trends={data.trends}
              lastUpdatedLabel={copy.aiPulseLastUpdatedLabel}
              noTrendsLabel={copy.aiPulseNoTrendsLabel}
            />
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-5">
          <h3 className="flex items-center gap-2 text-xs font-semibold tracking-[0.16em] text-muted-foreground uppercase after:h-px after:flex-1 after:bg-border">
            {copy.aiPulseStocksTitle}
          </h3>
          <div className="mt-4">
            <StockChart
              initialTicker={data.initialTicker}
              initialData={data.initialStockData}
              tickers={displayTickers}
              copy={{
                tickerLabel: copy.aiPulseTickerLabel,
                loadingLabel: copy.aiPulseLoadingLabel,
              }}
            />
          </div>
        </div>
      </div>
    </section>
  );
}

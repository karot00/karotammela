import type { AiTrend } from "@/lib/db/schema";

type TrendsListProps = {
  trends: AiTrend[];
  lastUpdatedLabel: string;
  noTrendsLabel: string;
  locale?: string;
};

export function TrendsList({
  trends,
  lastUpdatedLabel,
  noTrendsLabel,
  locale,
}: TrendsListProps): React.ReactElement {
  if (trends.length === 0) {
    return (
      <div className="flex min-h-[200px] items-center justify-center rounded-xl border border-border bg-card p-6">
        <p className="text-sm text-muted-foreground">{noTrendsLabel}</p>
      </div>
    );
  }

  const lastDate = trends[0]?.date ?? null;

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        {trends.map((trend) => (
          <article
            key={trend.id}
            className="rounded-lg border border-border bg-background/60 p-4 transition-colors hover:bg-muted/40"
          >
            <a
              href={trend.url}
              target="_blank"
              rel="noreferrer noopener"
              className="group"
            >
              <p className="text-sm font-semibold text-foreground transition-colors group-hover:text-primary">
                {trend.title}
              </p>
            </a>
            <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
              {locale === "fi" && trend.summaryFi
                ? trend.summaryFi
                : trend.summary}
            </p>
            <div className="mt-2 flex items-center gap-2">
              {trend.source ? (
                <span className="inline-flex items-center rounded-full border border-border bg-muted px-2 py-0.5 text-[11px] font-semibold text-foreground uppercase">
                  {trend.source === "hackernews" ? "HN" : trend.source}
                </span>
              ) : null}
              <span className="text-[11px] text-muted-foreground">
                {trend.date}
              </span>
            </div>
          </article>
        ))}
      </div>

      {lastDate ? (
        <p className="text-xs text-muted-foreground">
          {lastUpdatedLabel}: {lastDate}
        </p>
      ) : null}
    </div>
  );
}

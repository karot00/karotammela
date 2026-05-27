import type { AiRepo } from "@/lib/db/schema";
import { Star } from "lucide-react";

type ReposListProps = {
  repos: AiRepo[];
  lastUpdatedLabel: string;
  noReposLabel: string;
  starsTodayLabel: string;
  locale?: string;
};

export function ReposList({
  repos,
  lastUpdatedLabel,
  noReposLabel,
  starsTodayLabel,
  locale,
}: ReposListProps): React.ReactElement {
  if (repos.length === 0) {
    return (
      <div className="flex min-h-[200px] items-center justify-center rounded-xl border border-border bg-card p-6">
        <p className="text-sm text-muted-foreground">{noReposLabel}</p>
      </div>
    );
  }

  const lastDate = repos[0]?.date ?? null;

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        {repos.map((repo) => (
          <article
            key={repo.id}
            className="rounded-lg border border-border bg-background/60 p-4 transition-colors hover:bg-muted/40"
          >
            <a
              href={repo.url}
              target="_blank"
              rel="noreferrer noopener"
              className="group"
            >
              <p className="text-sm font-semibold text-foreground transition-colors group-hover:text-primary">
                {repo.repoFullName}
              </p>
            </a>
            <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
              {locale === "fi" && repo.descriptionFi
                ? repo.descriptionFi
                : repo.description}
            </p>
            <div className="mt-3 flex flex-wrap items-center gap-2">
              {repo.language ? (
                <span className="inline-flex items-center rounded-full border border-border bg-muted px-2 py-0.5 text-[11px] font-semibold text-foreground">
                  {repo.language}
                </span>
              ) : null}
              <span className="inline-flex items-center gap-1 rounded-full border border-accent/20 bg-accent/5 px-2 py-0.5 text-[11px] font-semibold text-accent">
                <Star className="h-3 w-3 fill-current" />
                <span>
                  {repo.starsToday} {starsTodayLabel}
                </span>
              </span>
              <span className="text-[11px] text-muted-foreground ml-auto">
                {repo.date}
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

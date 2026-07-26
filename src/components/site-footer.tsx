"use client";

import { useEffect, useState } from "react";

import { useTranslations } from "next-intl";

import { GO_ORIGIN, NEXT_ORIGIN } from "@/lib/comparison";
import { Link } from "@/i18n/navigation";
import { ShareButton } from "@/components/share-button";

async function medianTtfb(
  origin: string,
  samples: number,
): Promise<number | null> {
  const url = `${origin.replace(/\/$/, "")}/api/ping`;
  const values: number[] = [];

  for (let i = 0; i < samples; i += 1) {
    const start = performance.now();
    try {
      const res = await fetch(url, { cache: "no-store", mode: "cors" });
      if (!res.ok) continue;
      await res.text();
      values.push(performance.now() - start);
    } catch {
      // Ignore this sample.
    }
  }

  if (values.length === 0) return null;
  values.sort((a, b) => a - b);
  const mid = Math.floor(values.length / 2);
  const median =
    values.length % 2 ? values[mid] : (values[mid - 1] + values[mid]) / 2;
  return Math.round(median);
}

function barWidth(value: number | null): number {
  if (value === null) return 2;
  // Directly proportional to the measured time: a longer TTFB draws a
  // longer bar (0 ms -> 2% floor, 400 ms -> 100% cap).
  return Math.max(2, Math.min(100, (value / 400) * 100));
}

export function SiteFooter() {
  const t = useTranslations("comparison");
  const th = useTranslations("home");
  const [goMs, setGoMs] = useState<number | null>(null);
  const [nextMs, setNextMs] = useState<number | null>(null);
  const [goProbed, setGoProbed] = useState(false);
  const [nextProbed, setNextProbed] = useState(false);

  useEffect(() => {
    let active = true;
    void medianTtfb(GO_ORIGIN, 4).then((v) => {
      if (!active) return;
      setGoMs(v);
      setGoProbed(true);
    });
    void medianTtfb(NEXT_ORIGIN, 4).then((v) => {
      if (!active) return;
      setNextMs(v);
      setNextProbed(true);
    });
    return () => {
      active = false;
    };
  }, []);

  function perfValue(ms: number | null, probed: boolean): string {
    if (ms !== null) return `${ms} ${t("perf.unit")}`;
    return probed ? t("perf.unavailable") : t("perf.loading");
  }

  const year = new Date().getFullYear();

  return (
    <footer className="mt-16 border-t border-border/60 bg-card/30">
      <div className="mx-auto max-w-5xl px-4 py-10 sm:px-6">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-start sm:justify-between">
          <div className="max-w-md">
            <p className="text-sm font-semibold tracking-[0.18em] text-foreground uppercase">
              Karo Tammela
            </p>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
              {th("intro")}
            </p>
            <div className="mt-4 flex flex-wrap items-center gap-3">
              <ShareButton
                url={
                  typeof window !== "undefined" ? window.location.pathname : "/"
                }
                title={th("phaseLabel")}
                text={th("intro")}
                label={th("shareLabel")}
                copiedLabel={th("shareCopiedLabel")}
                errorLabel={th("shareErrorLabel")}
              />
              <Link
                href="/privacy"
                className="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground hover:text-foreground"
              >
                {th("privacyPolicyLink")}
              </Link>
            </div>
          </div>

          <div className="w-full rounded-2xl border border-border/60 bg-background/60 p-4 sm:w-72">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">
              {t("perf.title")}
            </p>
            <p className="mt-1 text-[11px] text-muted-foreground">
              {t("perf.caption")}
            </p>
            <div className="mt-3 space-y-3">
              <div>
                <div className="flex items-center justify-between text-xs">
                  <span className="font-medium text-foreground">
                    {t("perf.go")}
                  </span>
                  <span className="font-mono text-primary">
                    {perfValue(goMs, goProbed)}
                  </span>
                </div>
                <div className="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-secondary">
                  <div
                    className="h-full bg-primary transition-[width] duration-500"
                    style={{ width: `${barWidth(goMs)}%` }}
                  />
                </div>
              </div>
              <div>
                <div className="flex items-center justify-between text-xs">
                  <span className="font-medium text-foreground">
                    {t("perf.next")}
                  </span>
                  <span className="font-mono text-accent">
                    {perfValue(nextMs, nextProbed)}
                  </span>
                </div>
                <div className="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-secondary">
                  <div
                    className="h-full bg-accent transition-[width] duration-500"
                    style={{ width: `${barWidth(nextMs)}%` }}
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
        <p className="mt-8 text-xs text-muted-foreground">
          © {year} Karo Tammela — Next.js build
        </p>
      </div>
    </footer>
  );
}

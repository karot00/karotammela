"use client";

import { useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";

import { GO_ORIGIN, NEXT_ORIGIN } from "@/lib/comparison";

type Stack = "go" | "next";

/**
 * Direct client-side TTFB probe: fetch `/api/ping` on the target origin a few
 * times and return the median so a single cold hit does not dominate.
 * Cross-origin failures (CORS/offline) resolve to `null`, never a fake `0`.
 */
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
  return Math.max(2, Math.min(100, 100 - (value / 400) * 100));
}

/**
 * Floating Tech Switcher + performance widget for the comparison experiment.
 *
 * This build is served from the Next.js subdomain, so the current stack is
 * always `next`. Switching redirects cross-origin to the Go apex (or back),
 * preserving the current path, query, and hash — no cookie routing.
 */
export function TechComparison() {
  const t = useTranslations("comparison");
  const [open, setOpen] = useState(false);
  const [goMs, setGoMs] = useState<number | null>(null);
  const [nextMs, setNextMs] = useState<number | null>(null);
  const [goProbed, setGoProbed] = useState(false);
  const [nextProbed, setNextProbed] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // This build is served from the Next.js subdomain, so the current stack is
  // always "next"; the Go button is the cross-origin destination.
  const currentStack: Stack = "next";

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

  useEffect(() => {
    if (!open) return undefined;
    function onPointerDown(event: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", onPointerDown);
    return () => document.removeEventListener("mousedown", onPointerDown);
  }, [open]);

  function switchTo(target: Stack) {
    const base = (target === "next" ? NEXT_ORIGIN : GO_ORIGIN).replace(
      /\/$/,
      "",
    );
    const rest =
      window.location.pathname + window.location.search + window.location.hash;
    const dest = base + rest;
    if (dest === window.location.href) {
      setOpen(false);
      return;
    }
    window.location.href = dest;
  }

  function perfValue(ms: number | null, probed: boolean): string {
    if (ms !== null) return `${ms} ${t("perf.unit")}`;
    return probed ? t("perf.unavailable") : t("perf.loading");
  }

  return (
    <div ref={containerRef} className="fixed right-4 bottom-4 z-50">
      {open ? (
        <div className="mb-2 w-64 rounded-2xl border border-border/70 bg-popover/95 p-4 shadow-2xl backdrop-blur-md">
          <p className="text-xs font-semibold tracking-[0.16em] text-muted-foreground uppercase">
            {t("switcher.title")}
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            {t("switcher.description")}
          </p>
          <div className="mt-3 grid grid-cols-2 gap-2">
            <button
              type="button"
              onClick={() => switchTo("go")}
              className="rounded-md border border-border/60 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-muted-foreground uppercase transition-colors"
            >
              {t("switcher.go")}
            </button>
            <button
              type="button"
              onClick={() => switchTo("next")}
              className="rounded-md border border-primary/70 bg-primary/20 px-3 py-2 text-xs font-semibold tracking-[0.12em] text-primary uppercase transition-colors"
            >
              {t("switcher.next")}
            </button>
          </div>

          <div className="mt-4 border-t border-border/60 pt-3">
            <p className="text-xs font-semibold tracking-[0.16em] text-muted-foreground uppercase">
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
      ) : null}

      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        className="ml-auto flex h-12 w-12 items-center justify-center rounded-full border border-primary/60 bg-primary/15 text-primary shadow-lg backdrop-blur-md transition-colors hover:bg-primary/25"
        aria-label={t("switcher.ariaLabel", { stack: currentStack })}
        aria-expanded={open}
        title={t("switcher.toggleLabel")}
      >
        <span className="text-lg font-bold">N</span>
      </button>
    </div>
  );
}

"use client";

import { useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";

import { GO_ORIGIN, NEXT_ORIGIN } from "@/lib/comparison";

type Stack = "go" | "next";

/**
 * Floating Tech Switcher for the comparison experiment.
 *
 * This build is served from the Next.js subdomain, so the current stack is
 * always `next`. Switching redirects cross-origin to the Go apex (or back),
 * preserving the current path, query, and hash — no cookie routing.
 */
export function TechComparison() {
  const t = useTranslations("comparison");
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // This build is served from the Next.js subdomain, so the current stack is
  // always "next"; the Go button is the cross-origin destination.
  const currentStack: Stack = "next";

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

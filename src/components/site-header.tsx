"use client";

import { useTranslations } from "next-intl";

import { Link } from "@/i18n/navigation";
import { LocaleSwitcher } from "@/components/locale-switcher";

export function SiteHeader() {
  const t = useTranslations("home");

  return (
    <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex max-w-5xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
        <Link
          href="/"
          className="text-sm font-semibold tracking-[0.18em] text-foreground uppercase"
        >
          Karo Tammela
        </Link>
        <nav className="flex items-center gap-2 sm:gap-3" aria-label="Primary">
          <Link
            href="/blog"
            className="text-xs font-medium uppercase tracking-[0.12em] text-muted-foreground hover:text-foreground"
          >
            {t("blogLink")}
          </Link>
          <Link
            href="/privacy"
            className="text-xs font-medium uppercase tracking-[0.12em] text-muted-foreground hover:text-foreground"
          >
            {t("privacyPolicyLink")}
          </Link>
          <LocaleSwitcher />
        </nav>
      </div>
    </header>
  );
}

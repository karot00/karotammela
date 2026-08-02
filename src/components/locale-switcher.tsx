"use client";

import { useLocale, useTranslations } from "next-intl";

import { Link, usePathname } from "@/i18n/navigation";
import { routing, type AppLocale } from "@/i18n/routing";
import { getTranslatedBlogSlug } from "@/lib/blog-localization";

export function LocaleSwitcher() {
  const t = useTranslations("localeSwitcher");
  const locale = useLocale() as AppLocale;
  const pathname = usePathname();

  return (
    <nav className="flex items-center gap-2" aria-label={t("label")}>
      {routing.locales.map((nextLocale) => {
        const isActive = nextLocale === locale;
        const targetPathname = pathname.startsWith("/blog/")
          ? `/blog/${getTranslatedBlogSlug(
              locale,
              pathname.slice("/blog/".length),
            )}`
          : pathname;

        return (
          <Link
            key={nextLocale}
            href={targetPathname}
            locale={nextLocale}
            className={`inline-flex min-w-16 items-center justify-center rounded-md border px-3 py-1.5 text-xs font-semibold uppercase tracking-[0.12em] transition-colors ${
              isActive
                ? "border-primary/70 bg-primary/20 text-primary"
                : "border-border/60 text-muted-foreground hover:border-primary/40 hover:text-foreground"
            }`}
          >
            {t(nextLocale)}
          </Link>
        );
      })}
    </nav>
  );
}

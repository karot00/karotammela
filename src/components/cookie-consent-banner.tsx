"use client";

import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { useCookieConsent } from "@/modules/cookie-consent";

export function CookieConsentBanner() {
  const t = useTranslations("cookieConsent.banner");
  const { isBannerVisible, acceptAll, rejectNonEssential, openSettings } =
    useCookieConsent();

  if (!isBannerVisible) {
    return null;
  }

  return (
    <aside
      className="fixed right-3 bottom-3 z-50 w-[min(24rem,calc(100vw-1.5rem))] rounded-xl border border-border/80 bg-card/95 p-4 text-card-foreground shadow-[0_18px_44px_rgba(0,0,0,0.35)] backdrop-blur-sm sm:right-5 sm:bottom-5"
      aria-live="polite"
      aria-label={t("ariaLabel")}
    >
      <p className="text-[0.72rem] font-semibold tracking-[0.15em] text-primary uppercase">
        {t("eyebrow")}
      </p>
      <h2 className="mt-1 text-base font-semibold">{t("title")}</h2>
      <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
        {t("description")}
      </p>

      <div className="mt-4 flex flex-wrap gap-2">
        <Button type="button" size="sm" onClick={acceptAll}>
          {t("acceptAll")}
        </Button>
        <Button
          type="button"
          size="sm"
          variant="secondary"
          onClick={rejectNonEssential}
        >
          {t("rejectNonEssential")}
        </Button>
        <Button type="button" size="sm" variant="ghost" onClick={openSettings}>
          {t("manageSettings")}
        </Button>
      </div>
    </aside>
  );
}

import { routing, type AppLocale } from "@/i18n/routing";
import { toAbsoluteUrl } from "@/lib/site-url";

export const LOCALE_TO_LANGUAGE_TAG: Record<AppLocale, string> = {
  fi: "fi-FI",
  en: "en-US",
};

export const SITE_NAME = "karotammela.fi";
export const SITE_DESCRIPTION =
  "Agentic AI architect portfolio and Sentinel challenge";
export const SITE_OWNER_NAME = "Karo Tammela";
export const SITE_OG_IMAGE_PATH = "/media/karo-tammela-agentic-ai.png";
export const SITE_OG_IMAGE_ALT =
  "Karo Tammela - Agentic AI blog and experiment lab";

export function getDefaultSocialImage() {
  return {
    url: toAbsoluteUrl(SITE_OG_IMAGE_PATH),
    width: 1200,
    height: 630,
    alt: SITE_OG_IMAGE_ALT,
  };
}

export function getLocaleFromSegment(value: string): AppLocale {
  return value === "en" ? "en" : "fi";
}

export function buildLocalizedPath(locale: AppLocale, pathname = "") {
  const normalizedPath = pathname ? `/${pathname.replace(/^\/+/, "")}` : "";
  return `/${locale}${normalizedPath}`;
}

export function getLocalizedAlternates(pathname = "") {
  const languages = Object.fromEntries(
    routing.locales.map((locale) => [
      LOCALE_TO_LANGUAGE_TAG[locale],
      toAbsoluteUrl(buildLocalizedPath(locale, pathname)),
    ]),
  );

  return {
    languages,
    canonical: toAbsoluteUrl(
      buildLocalizedPath(routing.defaultLocale, pathname),
    ),
  };
}

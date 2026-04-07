import { defineRouting } from "next-intl/routing";

const supportedLocales = ["fi", "en"] as const;
const preferredDefaultLocale = process.env.NEXT_PUBLIC_DEFAULT_LOCALE;
const defaultLocale =
  preferredDefaultLocale === "fi" || preferredDefaultLocale === "en"
    ? preferredDefaultLocale
    : "fi";

export const routing = defineRouting({
  locales: supportedLocales,
  defaultLocale,
  localePrefix: "always",
});

export type AppLocale = (typeof routing.locales)[number];

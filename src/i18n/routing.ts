import { defineRouting } from "next-intl/routing";

const supportedLocales = ["fi", "en"] as const;

export const routing = defineRouting({
  locales: supportedLocales,
  defaultLocale: "fi",
  localePrefix: "always",
});

export type AppLocale = (typeof routing.locales)[number];

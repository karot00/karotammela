import type { ReactNode } from "react";
import { hasLocale, NextIntlClientProvider } from "next-intl";
import { getMessages, setRequestLocale } from "next-intl/server";
import { notFound } from "next/navigation";

import { CookieConsentBanner } from "@/components/cookie-consent-banner";
import { CookieConsentPreferences } from "@/components/cookie-consent-preferences";
import { SiteFooter } from "@/components/site-footer";
import { SiteHeader } from "@/components/site-header";
import { TechComparison } from "@/components/tech-comparison";
import { routing } from "@/i18n/routing";

type LocaleLayoutProps = {
  children: ReactNode;
  params: Promise<{ locale: string }>;
};

export default async function LocaleLayout({
  children,
  params,
}: LocaleLayoutProps) {
  const { locale } = await params;

  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  setRequestLocale(locale);
  const messages = await getMessages();

  return (
    <NextIntlClientProvider messages={messages}>
      <div lang={locale} className="flex min-h-full flex-1 flex-col">
        <SiteHeader />
        {children}
        <SiteFooter />
        <CookieConsentPreferences />
        <CookieConsentBanner />
        <TechComparison />
      </div>
    </NextIntlClientProvider>
  );
}

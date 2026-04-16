import type { ReactNode } from "react";
import { hasLocale, NextIntlClientProvider } from "next-intl";
import { getMessages, setRequestLocale } from "next-intl/server";
import { notFound } from "next/navigation";

import { CookieConsentBanner } from "@/components/cookie-consent-banner";
import { CookieConsentPreferences } from "@/components/cookie-consent-preferences";
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
      <div className="flex min-h-full flex-1 flex-col">
        {children}
        <CookieConsentPreferences />
        <CookieConsentBanner />
      </div>
    </NextIntlClientProvider>
  );
}

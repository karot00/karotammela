import Link from "next/link";
import { getTranslations, setRequestLocale } from "next-intl/server";

import { Button } from "@/components/ui/button";
import { LocaleSwitcher } from "@/components/locale-switcher";

type Props = {
  params: Promise<{ locale: string }>;
  searchParams: Promise<{ returnTo?: string }>;
};

export default async function PrivacyPage({ params, searchParams }: Props) {
  const { locale } = await params;
  const { returnTo } = await searchParams;
  setRequestLocale(locale);

  const t = await getTranslations("privacyPolicy");
  const fallbackPath = `/${locale}`;
  const returnPath =
    returnTo && returnTo.startsWith(fallbackPath) ? returnTo : fallbackPath;

  return (
    <main className="flex flex-1 px-6 py-10 sm:py-16">
      <div className="mx-auto flex w-full max-w-4xl flex-col gap-8">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold">{t("title")}</h1>
          <LocaleSwitcher />
        </div>

        <div>
          <Button asChild variant="link" className="h-auto p-0 text-sm">
            <Link href={returnPath}>{t("backLinkLabel")}</Link>
          </Button>
        </div>

        <div className="prose dark:prose-invert max-w-none">
          <h2>{t("introTitle")}</h2>
          <p>{t("introBody")}</p>

          <h2>{t("dataTitle")}</h2>
          <h3>{t("contactForm")}</h3>
          <p>{t("contactFormBody")}</p>
          <h3>{t("analytics")}</h3>
          <p>{t("analyticsBody")}</p>
          <h3>{t("cookies")}</h3>
          <p>{t("cookiesBody")}</p>
          <ul>
            <li>
              <code>karot_unlock</code>: {t("cookieUnlock")}
            </li>
            <li>
              <code>consent</code>: {t("cookieConsent")}
            </li>
          </ul>

          <h2>{t("rightsTitle")}</h2>
          <p>{t("rightsBody")}</p>

          <h2>{t("contactTitle")}</h2>
          <p>{t("contactBody")}</p>
        </div>
      </div>
    </main>
  );
}

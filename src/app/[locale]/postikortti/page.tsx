import type { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { getTranslations } from "next-intl/server";

import { routing } from "@/i18n/routing";
import { getLocaleFromSegment, getLocalizedAlternates } from "@/lib/seo";
import { toAbsoluteUrl } from "@/lib/site-url";
import { PostcardClient } from "@/components/postikortti/postcard-client";

type Props = {
  params: Promise<{ locale: string }>;
};

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params;
  const currentLocale = getLocaleFromSegment(locale);
  const t = await getTranslations({
    locale: currentLocale,
    namespace: "postcard",
  });

  return {
    title: `${t("title")} | Karo Tammela`,
    description: t("description"),
    alternates: getLocalizedAlternates("postikortti"),
    openGraph: {
      type: "website",
      title: `${t("title")} | Karo Tammela`,
      description: t("description"),
      locale: currentLocale === "fi" ? "fi_FI" : "en_US",
      url: toAbsoluteUrl(`/${currentLocale}/postikortti`),
    },
    twitter: {
      card: "summary_large_image",
      title: `${t("title")} | Karo Tammela`,
      description: t("description"),
    },
  };
}

export default async function PostcardPage({ params }: Props) {
  const { locale } = await params;
  setRequestLocale(locale);

  return (
    <main className="flex flex-1 px-6 py-10 sm:py-16">
      <PostcardClient locale={locale} />
    </main>
  );
}

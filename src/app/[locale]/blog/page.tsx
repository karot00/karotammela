import Link from "next/link";
import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";

import { routing } from "@/i18n/routing";
import { getAllBlogPosts } from "@/lib/blog";
import {
  SITE_DESCRIPTION,
  SITE_NAME,
  getLocaleFromSegment,
  getLocalizedAlternates,
} from "@/lib/seo";

type BlogIndexPageProps = {
  params: Promise<{ locale: string }>;
};

export async function generateMetadata({
  params,
}: BlogIndexPageProps): Promise<Metadata> {
  const { locale } = await params;
  const currentLocale = getLocaleFromSegment(locale);
  const tLocale = await getTranslations({
    locale: currentLocale,
    namespace: "dashboard",
  });
  const title = `${tLocale("navBlog")} | ${SITE_NAME}`;

  return {
    title,
    description: SITE_DESCRIPTION,
    alternates: getLocalizedAlternates("blog"),
    openGraph: {
      title,
      description: SITE_DESCRIPTION,
      type: "website",
      locale: currentLocale === "fi" ? "fi_FI" : "en_US",
      siteName: SITE_NAME,
    },
    twitter: {
      card: "summary_large_image",
      title,
      description: SITE_DESCRIPTION,
    },
  };
}

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export const dynamicParams = false;

export default async function BlogIndexPage({ params }: BlogIndexPageProps) {
  const { locale } = await params;
  const currentLocale = getLocaleFromSegment(locale);

  setRequestLocale(currentLocale);

  const t = await getTranslations({
    locale: currentLocale,
    namespace: "dashboard",
  });
  const posts = await getAllBlogPosts(currentLocale);
  const localeCode = currentLocale === "fi" ? "fi-FI" : "en-US";

  return (
    <main className="mx-auto flex w-full max-w-4xl flex-1 px-4 py-10 sm:px-6">
      <section className="w-full space-y-6">
        <header className="space-y-2">
          <h1 className="text-3xl font-semibold text-foreground">
            {t("navBlog")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("blogPlaceholder")}
          </p>
        </header>

        {posts.length === 0 ? (
          <p className="rounded-xl border border-border bg-card p-6 text-sm text-muted-foreground">
            {t("blogNoPostsLabel")}
          </p>
        ) : (
          <div className="space-y-3">
            {posts.map((post) => (
              <article
                key={post.slug}
                className="rounded-xl border border-border bg-card p-5 transition-colors hover:bg-muted/40"
              >
                <Link
                  href={`/${currentLocale}/blog/${post.slug}`}
                  className="block"
                >
                  <h2 className="text-lg font-semibold text-foreground">
                    {post.title}
                  </h2>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {new Intl.DateTimeFormat(localeCode, {
                      dateStyle: "medium",
                    }).format(new Date(post.publishedAt))}
                  </p>
                  <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                    {post.description}
                  </p>
                </Link>
              </article>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

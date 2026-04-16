import Link from "next/link";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getTranslations, setRequestLocale } from "next-intl/server";

import { renderBlogMarkdownToSafeHtml } from "@/components/blog-post-content";
import { routing } from "@/i18n/routing";
import { getAllBlogPosts, getBlogPostBySlug } from "@/lib/blog";
import {
  SITE_NAME,
  getLocaleFromSegment,
  getLocalizedAlternates,
} from "@/lib/seo";
import { toAbsoluteUrl } from "@/lib/site-url";

type BlogDetailPageProps = {
  params: Promise<{ locale: string; slug: string }>;
};

export async function generateMetadata({
  params,
}: BlogDetailPageProps): Promise<Metadata> {
  const { locale, slug } = await params;
  const currentLocale = getLocaleFromSegment(locale);
  const post = await getBlogPostBySlug(currentLocale, slug);

  if (!post) {
    return {
      title: `Post not found | ${SITE_NAME}`,
      robots: {
        index: false,
        follow: false,
      },
      alternates: getLocalizedAlternates(`blog/${slug}`),
    };
  }

  const alternates = getLocalizedAlternates(`blog/${post.slug}`);

  return {
    title: `${post.title} | ${SITE_NAME}`,
    description: post.description,
    alternates,
    openGraph: {
      title: post.title,
      description: post.description,
      type: "article",
      locale: currentLocale === "fi" ? "fi_FI" : "en_US",
      siteName: SITE_NAME,
      url: toAbsoluteUrl(`/${currentLocale}/blog/${post.slug}`),
      publishedTime: new Date(post.publishedAt).toISOString(),
      tags: post.tags,
    },
    twitter: {
      card: "summary_large_image",
      title: post.title,
      description: post.description,
    },
  };
}

export async function generateStaticParams() {
  const allPosts = await Promise.all(
    routing.locales.map(async (locale) => {
      const posts = await getAllBlogPosts(locale);
      return posts.map((post) => ({ locale, slug: post.slug }));
    }),
  );

  return allPosts.flat();
}

export const dynamicParams = false;

export default async function BlogDetailPage({ params }: BlogDetailPageProps) {
  const { locale, slug } = await params;
  const currentLocale = getLocaleFromSegment(locale);
  setRequestLocale(currentLocale);

  const t = await getTranslations({
    locale: currentLocale,
    namespace: "dashboard",
  });
  const post = await getBlogPostBySlug(currentLocale, slug);

  if (!post) {
    notFound();
  }

  const localeCode = currentLocale === "fi" ? "fi-FI" : "en-US";
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    headline: post.title,
    description: post.description,
    datePublished: new Date(post.publishedAt).toISOString(),
    dateModified: new Date(post.publishedAt).toISOString(),
    inLanguage: localeCode,
    author: {
      "@type": "Person",
      name: "Karo Tammela",
      url: toAbsoluteUrl(`/${currentLocale}`),
    },
    publisher: {
      "@type": "Organization",
      name: SITE_NAME,
      url: toAbsoluteUrl("/"),
    },
    mainEntityOfPage: toAbsoluteUrl(`/${currentLocale}/blog/${post.slug}`),
    keywords: post.tags.join(", "),
  };

  return (
    <main className="mx-auto flex w-full max-w-4xl flex-1 px-4 py-10 sm:px-6">
      <article className="w-full rounded-xl border border-border bg-card p-6 sm:p-8">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />

        <Link
          href={`/${currentLocale}/blog`}
          className="text-sm font-medium text-primary hover:text-primary/80"
        >
          {t("blogBackLabel")}
        </Link>

        <h1 className="mt-4 text-3xl font-semibold text-foreground">
          {post.title}
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          {new Intl.DateTimeFormat(localeCode, { dateStyle: "medium" }).format(
            new Date(post.publishedAt),
          )}
        </p>
        <p className="mt-4 text-base text-muted-foreground">
          {post.description}
        </p>

        <div
          className="blog-prose mt-8"
          dangerouslySetInnerHTML={{
            __html: renderBlogMarkdownToSafeHtml(post.body),
          }}
        />
      </article>
    </main>
  );
}

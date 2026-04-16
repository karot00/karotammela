import type { MetadataRoute } from "next";

import { routing } from "@/i18n/routing";
import { getAllBlogPosts } from "@/lib/blog";
import { toAbsoluteUrl } from "@/lib/site-url";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const staticRoutes: MetadataRoute.Sitemap = [
    {
      url: toAbsoluteUrl("/"),
      lastModified: new Date(),
      changeFrequency: "weekly",
      priority: 1,
    },
    ...routing.locales.flatMap((locale) => [
      {
        url: toAbsoluteUrl(`/${locale}`),
        lastModified: new Date(),
        changeFrequency: "weekly" as const,
        priority: 0.9,
      },
      {
        url: toAbsoluteUrl(`/${locale}/blog`),
        lastModified: new Date(),
        changeFrequency: "daily" as const,
        priority: 0.8,
      },
      {
        url: toAbsoluteUrl(`/${locale}/privacy`),
        lastModified: new Date(),
        changeFrequency: "monthly" as const,
        priority: 0.5,
      },
    ]),
  ];

  const blogEntries = await Promise.all(
    routing.locales.map(async (locale) => {
      const posts = await getAllBlogPosts(locale);

      return posts.map((post) => ({
        url: toAbsoluteUrl(`/${locale}/blog/${post.slug}`),
        lastModified: new Date(post.publishedAt),
        changeFrequency: "monthly" as const,
        priority: 0.7,
      }));
    }),
  );

  return [...staticRoutes, ...blogEntries.flat()];
}

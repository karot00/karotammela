import { readdir, readFile } from "node:fs/promises";
import path from "node:path";

import matter from "gray-matter";
import { z } from "zod";

export const blogLocaleSchema = z.enum(["fi", "en"]);

const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

const blogFrontmatterSchema = z.object({
  title: z.string().min(1),
  description: z.string().min(1),
  publishedAt: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
  slug: z.string().regex(slugPattern),
  draft: z.boolean().optional().default(false),
  tags: z.array(z.string().min(1)).optional().default([]),
});

export type BlogLocale = z.infer<typeof blogLocaleSchema>;
export type BlogFrontmatter = z.infer<typeof blogFrontmatterSchema>;

export type BlogPost = BlogFrontmatter & {
  locale: BlogLocale;
  body: string;
  sourcePath: string;
};

export type BlogPaginationResult = {
  page: number;
  pageSize: number;
  total: number;
  pages: number;
  hasPrev: boolean;
  hasNext: boolean;
  items: BlogPost[];
};

const BLOG_CONTENT_ROOT = path.join(process.cwd(), "content", "blog");

function isNodeErrorWithCode(
  error: unknown,
): error is NodeJS.ErrnoException & { code: string } {
  return (
    error instanceof Error &&
    "code" in error &&
    typeof (error as { code?: unknown }).code === "string"
  );
}

function normalizePage(input: number): number {
  if (!Number.isFinite(input)) {
    return 1;
  }

  return Math.max(1, Math.floor(input));
}

function isDraftVisible(draft: boolean) {
  if (process.env.NODE_ENV === "production") {
    return !draft;
  }

  return true;
}

function assertUniqueSlugs(posts: BlogPost[], locale: BlogLocale) {
  const seen = new Set<string>();

  for (const post of posts) {
    if (seen.has(post.slug)) {
      throw new Error(
        `Duplicate blog slug '${post.slug}' for locale '${locale}'.`,
      );
    }

    seen.add(post.slug);
  }
}

function getPublishedDateValue(post: BlogPost): number {
  return new Date(post.publishedAt).getTime();
}

function sortPosts(posts: BlogPost[]): BlogPost[] {
  return [...posts].sort((a, b) => {
    const dateDelta = getPublishedDateValue(b) - getPublishedDateValue(a);

    if (dateDelta !== 0) {
      return dateDelta;
    }

    return a.slug.localeCompare(b.slug, "en");
  });
}

async function readLocaleMarkdownFiles(locale: BlogLocale): Promise<string[]> {
  const localeDir = path.join(BLOG_CONTENT_ROOT, locale);

  try {
    const entries = await readdir(localeDir, { withFileTypes: true });

    return entries
      .filter((entry) => entry.isFile() && entry.name.endsWith(".md"))
      .map((entry) => path.join(localeDir, entry.name));
  } catch (error) {
    if (isNodeErrorWithCode(error) && error.code === "ENOENT") {
      return [];
    }

    throw error;
  }
}

function parsePostFile(
  filePath: string,
  locale: BlogLocale,
  source: string,
): BlogPost {
  const { data, content } = matter(source);
  const parsed = blogFrontmatterSchema.parse(data);
  const publishedDate = new Date(parsed.publishedAt);

  if (Number.isNaN(publishedDate.getTime())) {
    throw new Error(
      `Invalid publishedAt date '${parsed.publishedAt}' in ${filePath}.`,
    );
  }

  return {
    ...parsed,
    locale,
    body: content.trim(),
    sourcePath: filePath,
  };
}

export async function getAllBlogPosts(locale: BlogLocale): Promise<BlogPost[]> {
  const validLocale = blogLocaleSchema.parse(locale);
  const filePaths = await readLocaleMarkdownFiles(validLocale);

  const parsedPosts = await Promise.all(
    filePaths.map(async (filePath) => {
      const source = await readFile(filePath, "utf8");

      return parsePostFile(filePath, validLocale, source);
    }),
  );

  const visiblePosts = parsedPosts.filter((post) => isDraftVisible(post.draft));
  const sortedPosts = sortPosts(visiblePosts);

  assertUniqueSlugs(sortedPosts, validLocale);

  return sortedPosts;
}

export async function getBlogPostBySlug(
  locale: BlogLocale,
  slug: string,
): Promise<BlogPost | null> {
  if (!slugPattern.test(slug)) {
    return null;
  }

  const posts = await getAllBlogPosts(locale);

  return posts.find((post) => post.slug === slug) ?? null;
}

export function paginateBlogPosts(
  posts: BlogPost[],
  requestedPage: number,
  pageSize = 10,
): BlogPaginationResult {
  const safePageSize = Math.max(1, Math.floor(pageSize));
  const total = posts.length;
  const pages = Math.max(1, Math.ceil(total / safePageSize));
  const page = Math.min(normalizePage(requestedPage), pages);
  const startIndex = (page - 1) * safePageSize;

  return {
    page,
    pageSize: safePageSize,
    total,
    pages,
    hasPrev: page > 1,
    hasNext: page < pages,
    items: posts.slice(startIndex, startIndex + safePageSize),
  };
}

export function normalizeBlogViewQuery(input: {
  view?: string;
  page?: string;
  post?: string;
}): { view: "overview" | "blog"; page: number; post: string | null } {
  const view = input.view === "blog" ? "blog" : "overview";
  const pageValue = Number(input.page);
  const page = Number.isFinite(pageValue) ? normalizePage(pageValue) : 1;

  if (!input.post) {
    return { view, page, post: null };
  }

  return {
    view,
    page,
    post: slugPattern.test(input.post) ? input.post : null,
  };
}

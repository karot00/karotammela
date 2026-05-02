import { cookies } from "next/headers";
import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { redirect } from "next/navigation";

import { UnlockedDashboard } from "@/components/unlocked-dashboard";
import type { AiPulseData } from "@/components/ai-pulse/ai-pulse-shell";
import {
  getAllBlogPosts,
  getBlogPostBySlug,
  normalizeBlogViewQuery,
  paginateBlogPosts,
} from "@/lib/blog";
import { getChangelog } from "@/lib/changelog";
import { getCachedDashboardStats } from "@/lib/db/stats-cache";
import {
  getAvailableTickers,
  getLatestTrends,
  getStockHistory,
} from "@/lib/db/queries";
import { getLocaleFromSegment, getLocalizedAlternates } from "@/lib/seo";
import { verifyUnlockCookieValue } from "@/lib/security/unlock-cookie";
import { trackServerEvent } from "@/lib/telemetry/events";
import { AI_TICKERS } from "@/lib/ai/stocks-fetcher";

type DashboardPageProps = {
  params: Promise<{ locale: string }>;
  searchParams: Promise<{
    view?: string;
    page?: string;
    post?: string;
  }>;
};

type StatsResponse = {
  totalAttempts: number;
  unlockedCount: number;
  directUnlockCount: number;
  highestLevel: number;
  avgMessagesToUnlock: number;
  latestUnlockAt: string | null;
};

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const currentLocale = getLocaleFromSegment(locale);
  const t = await getTranslations({
    locale: currentLocale,
    namespace: "dashboard",
  });

  return {
    title: t("title"),
    description: t("description") || undefined,
    alternates: getLocalizedAlternates("dashboard"),
    robots: {
      index: false,
      follow: false,
    },
  };
}

async function getStats(): Promise<StatsResponse | null> {
  if (!process.env.TURSO_DATABASE_URL) {
    console.warn(
      "[dashboard] TURSO_DATABASE_URL is missing; returning null stats.",
    );
    return null;
  }

  try {
    const stats = await getCachedDashboardStats();

    return {
      ...stats,
      latestUnlockAt: stats.latestUnlockAt
        ? stats.latestUnlockAt.toISOString()
        : null,
    };
  } catch (error) {
    console.error("[dashboard] Failed to read stats", error);
    return null;
  }
}

async function getAiPulseData(): Promise<AiPulseData | null> {
  if (!process.env.TURSO_DATABASE_URL) {
    return null;
  }

  try {
    const [trends, availableTickers] = await Promise.all([
      getLatestTrends(),
      getAvailableTickers(),
    ]);

    const initialTicker =
      availableTickers.length > 0 ? (availableTickers[0] ?? "NVDA") : "NVDA";
    const initialStockData = await getStockHistory(initialTicker, 365);

    // Map ticker strings back to { ticker, name } using AI_TICKERS as source of truth
    const tickerMap = new Map<string, string>(
      AI_TICKERS.map((t) => [t.ticker, t.name]),
    );
    const displayTickers =
      availableTickers.length > 0
        ? availableTickers.map((t) => ({
            ticker: t,
            name: tickerMap.get(t) ?? t,
          }))
        : AI_TICKERS.map((t) => ({ ticker: t.ticker, name: t.name }));

    return {
      trends,
      initialTicker,
      initialStockData,
      availableTickers: displayTickers,
    };
  } catch (err) {
    console.error("[dashboard] Failed to fetch AI Pulse data", err);
    return null;
  }
}

export default async function DashboardPage({
  params,
  searchParams,
}: DashboardPageProps) {
  const { locale } = await params;
  const query = await searchParams;
  const normalizedBlogState = normalizeBlogViewQuery(query);

  setRequestLocale(locale);
  const cookieStore = await cookies();
  const unlockCookie = cookieStore.get("karot_unlock")?.value;
  const secret = process.env.UNLOCK_COOKIE_SECRET;

  if (!unlockCookie || !secret) {
    redirect(`/${locale}`);
  }

  const payload = verifyUnlockCookieValue(unlockCookie, secret);

  if (!payload) {
    redirect(`/${locale}`);
  }

  trackServerEvent("dashboard.opened", {
    locale,
    sessionId: payload.sessionId,
  });

  const t = await getTranslations("dashboard");
  const stats = await getStats();
  const posts = await getAllBlogPosts(locale === "fi" ? "fi" : "en");
  const paginatedPosts = paginateBlogPosts(posts, normalizedBlogState.page, 10);
  const changelog = getChangelog(locale === "fi" ? "fi" : "en");
  const aiPulse =
    normalizedBlogState.view === "ai-pulse" ? await getAiPulseData() : null;

  const selectedPost = normalizedBlogState.post
    ? await getBlogPostBySlug(
        locale === "fi" ? "fi" : "en",
        normalizedBlogState.post,
      )
    : (paginatedPosts.items[0] ?? null);

  return (
    <UnlockedDashboard
      locale={locale}
      initialView={normalizedBlogState.view}
      changelog={changelog}
      aiPulse={aiPulse}
      blog={{
        page: paginatedPosts.page,
        pageSize: paginatedPosts.pageSize,
        total: paginatedPosts.total,
        pages: paginatedPosts.pages,
        hasPrev: paginatedPosts.hasPrev,
        hasNext: paginatedPosts.hasNext,
        requestedPost: normalizedBlogState.post,
        selectedPost: selectedPost
          ? {
              title: selectedPost.title,
              description: selectedPost.description,
              publishedAt: selectedPost.publishedAt,
              slug: selectedPost.slug,
              draft: selectedPost.draft,
              tags: selectedPost.tags,
              body: selectedPost.body,
            }
          : null,
        items: paginatedPosts.items.map((post) => ({
          title: post.title,
          description: post.description,
          publishedAt: post.publishedAt,
          slug: post.slug,
          draft: post.draft,
          tags: post.tags,
        })),
      }}
      stats={
        stats
          ? {
              ...stats,
            }
          : null
      }
      copy={{
        badge: t("badge"),
        title: t("title"),
        description: t("description"),
        navigationLabel: t("navigationLabel"),
        navOverview: t("navOverview"),
        navProjects: t("navProjects"),
        navTech: t("navTech"),
        navBlog: t("navBlog"),
        navChangelog: t("navChangelog"),
        navSettings: t("navSettings"),
        menuLabel: t("menuLabel"),
        homeLinkLabel: t("homeLinkLabel"),
        aboutTitle: t("aboutTitle"),
        aboutBody: t("aboutBody"),
        profileImageAlt: t("profileImageAlt"),
        toolkitTitle: t("toolkitTitle"),
        codingModelsLabel: t("codingModelsLabel"),
        chatbotModelLabel: t("chatbotModelLabel"),
        analyticsTitle: t("analyticsTitle"),
        totalAttemptsLabel: t("totalAttemptsLabel"),
        unlockedCountLabel: t("unlockedCountLabel"),
        directUnlockCountLabel: t("directUnlockCountLabel"),
        avgMessagesToUnlockLabel: t("avgMessagesToUnlockLabel"),
        latestUnlockLabel: t("latestUnlockLabel"),
        latestUnlockNever: t("latestUnlockNever"),
        sourceOffline: t("sourceOffline"),
        contactTitle: t("contactTitle"),
        contactDescription: t("contactDescription"),
        contactAvailabilityEyebrow: t("contactAvailabilityEyebrow"),
        contactAvailabilityStatus: t("contactAvailabilityStatus"),
        contactConnectLabel: t("contactConnectLabel"),
        contactGithubLabel: t("contactGithubLabel"),
        contactLinkedinLabel: t("contactLinkedinLabel"),
        projectsTitle: t("projectsTitle"),
        projectOneTitle: t("projectOneTitle"),
        projectOneDescription: t("projectOneDescription"),
        projectTwoTitle: t("projectTwoTitle"),
        projectTwoDescription: t("projectTwoDescription"),
        projectThreeTitle: t("projectThreeTitle"),
        projectThreeDescription: t("projectThreeDescription"),
        projectFourTitle: t("projectFourTitle"),
        projectFourDescription: t("projectFourDescription"),
        projectGithubLabel: t("projectGithubLabel"),
        techTitle: t("techTitle"),
        techCategories: {
          frontend: t("techCategories.frontend"),
          backendDb: t("techCategories.backendDb"),
          infrastructure: t("techCategories.infrastructure"),
          tools: t("techCategories.tools"),
        },
        techItems: {
          nextjs: {
            name: t("techItems.nextjs.name"),
            description: t("techItems.nextjs.description"),
          },
          tailwindCss: {
            name: t("techItems.tailwindCss.name"),
            description: t("techItems.tailwindCss.description"),
          },
          postgresql: {
            name: t("techItems.postgresql.name"),
            description: t("techItems.postgresql.description"),
          },
          sqlite: {
            name: t("techItems.sqlite.name"),
            description: t("techItems.sqlite.description"),
          },
          vercel: {
            name: t("techItems.vercel.name"),
            description: t("techItems.vercel.description"),
          },
          cloudflare: {
            name: t("techItems.cloudflare.name"),
            description: t("techItems.cloudflare.description"),
          },
          github: {
            name: t("techItems.github.name"),
            description: t("techItems.github.description"),
          },
          resend: {
            name: t("techItems.resend.name"),
            description: t("techItems.resend.description"),
          },
          vsCode: {
            name: t("techItems.vsCode.name"),
            description: t("techItems.vsCode.description"),
          },
          kiloCode: {
            name: t("techItems.kiloCode.name"),
            description: t("techItems.kiloCode.description"),
          },
        },
        blogPlaceholder: t("blogPlaceholder"),
        blogPreviousLabel: t("blogPreviousLabel"),
        blogNextLabel: t("blogNextLabel"),
        blogBackLabel: t("blogBackLabel"),
        blogPageLabel: t("blogPageLabel"),
        blogNoPostsLabel: t("blogNoPostsLabel"),
        blogMissingPostLabel: t("blogMissingPostLabel"),
        blogShareLabel: t("blogShareLabel"),
        blogShareCopiedLabel: t("blogShareCopiedLabel"),
        blogShareErrorLabel: t("blogShareErrorLabel"),
        blogDraftBadgeLabel: t("blogDraftBadgeLabel"),
        settingsLanguageTitle: t("settingsLanguageTitle"),
        settingsLanguageDescription: t("settingsLanguageDescription"),
        settingsCookieTitle: t("settingsCookieTitle"),
        settingsCookieDescription: t("settingsCookieDescription"),
        settingsThemeTitle: t("settingsThemeTitle"),
        settingsThemeDescription: t("settingsThemeDescription"),
        privacyPolicyLink: t("privacyPolicyLink"),
        lightModeLabel: t("lightModeLabel"),
        darkModeLabel: t("darkModeLabel"),
        contactNameLabel: t("contactNameLabel"),
        contactEmailLabel: t("contactEmailLabel"),
        contactCompanyLabel: t("contactCompanyLabel"),
        contactMessageLabel: t("contactMessageLabel"),
        contactSubmitLabel: t("contactSubmitLabel"),
        contactPendingLabel: t("contactPendingLabel"),
        contactSuccessLabel: t("contactSuccessLabel"),
        contactErrorLabel: t("contactErrorLabel"),
        changelogTitle: t("changelogTitle"),
        changelogLead: t("changelogLead"),
        changelogEmptyLabel: t("changelogEmptyLabel"),
        changelogTypeAdded: t("changelogTypeAdded"),
        changelogTypeChanged: t("changelogTypeChanged"),
        changelogTypeFixed: t("changelogTypeFixed"),
        changelogTypeRemoved: t("changelogTypeRemoved"),
        navAiPulse: t("navAiPulse"),
        aiPulseTitle: t("aiPulseTitle"),
        aiPulseDescription: t("aiPulseDescription"),
        aiPulseTrendsTitle: t("aiPulseTrendsTitle"),
        aiPulseStocksTitle: t("aiPulseStocksTitle"),
        aiPulseTickerLabel: t("aiPulseTickerLabel"),
        aiPulseNoTrendsLabel: t("aiPulseNoTrendsLabel"),
        aiPulseLastUpdatedLabel: t("aiPulseLastUpdatedLabel"),
        aiPulseSourceLabel: t("aiPulseSourceLabel"),
        aiPulseLoadingLabel: t("aiPulseLoadingLabel"),
      }}
    />
  );
}

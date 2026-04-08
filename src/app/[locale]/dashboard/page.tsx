import { cookies } from "next/headers";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { redirect } from "next/navigation";

import { UnlockedDashboard } from "@/components/unlocked-dashboard";
import { getCachedDashboardStats } from "@/lib/db/stats-cache";
import { verifyUnlockCookieValue } from "@/lib/security/unlock-cookie";
import { trackServerEvent } from "@/lib/telemetry/events";

type DashboardPageProps = {
  params: Promise<{ locale: string }>;
};

type StatsResponse = {
  totalAttempts: number;
  unlockedCount: number;
  highestLevel: number;
  latestUnlockAt: string | null;
};

async function getStats(): Promise<StatsResponse | null> {
  if (!process.env.TURSO_DATABASE_URL) {
    return null;
  }

  try {
    const stats = await getCachedDashboardStats();

    return {
      ...stats,
      latestUnlockAt: stats.latestUnlockAt ? stats.latestUnlockAt.toISOString() : null,
    };
  } catch {
    return null;
  }
}

export default async function DashboardPage({ params }: DashboardPageProps) {
  const { locale } = await params;
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

  return (
    <UnlockedDashboard
      locale={locale}
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
        highestLevelLabel: t("highestLevelLabel"),
        latestUnlockLabel: t("latestUnlockLabel"),
        latestUnlockNever: t("latestUnlockNever"),
        sourceOffline: t("sourceOffline"),
        contactTitle: t("contactTitle"),
        contactDescription: t("contactDescription"),
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
        blogPlaceholder: t("blogPlaceholder"),
        settingsLanguageTitle: t("settingsLanguageTitle"),
        settingsLanguageDescription: t("settingsLanguageDescription"),
        settingsThemeTitle: t("settingsThemeTitle"),
        settingsThemeDescription: t("settingsThemeDescription"),
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
      }}
    />
  );
}

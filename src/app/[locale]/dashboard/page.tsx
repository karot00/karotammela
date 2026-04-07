import { cookies } from "next/headers";
import { getTranslations } from "next-intl/server";
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
  latestUnlockAt: Date | null;
};

async function getStats(locale: string): Promise<StatsResponse | null> {
  if (!process.env.TURSO_DATABASE_URL) {
    return null;
  }

  try {
    return await getCachedDashboardStats(locale as "en" | "fi");
  } catch {
    return null;
  }
}

export default async function DashboardPage({ params }: DashboardPageProps) {
  const { locale } = await params;
  const cookieStore = await cookies();
  const unlockCookie = cookieStore.get("karot_unlock")?.value;
  const secret = process.env.UNLOCK_COOKIE_SECRET;

  if (!unlockCookie || !secret) {
    redirect(`/${locale}`);
  }

  const payload = verifyUnlockCookieValue(unlockCookie, secret);

  if (!payload || payload.locale !== locale) {
    redirect(`/${locale}`);
  }

  trackServerEvent("dashboard.opened", {
    locale,
    sessionId: payload.sessionId,
  });

  const t = await getTranslations("dashboard");
  const stats = await getStats(locale);

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
        livePulseTitle: t("livePulseTitle"),
        totalAttemptsLabel: t("totalAttemptsLabel"),
        unlockedCountLabel: t("unlockedCountLabel"),
        highestLevelLabel: t("highestLevelLabel"),
        latestUnlockLabel: t("latestUnlockLabel"),
        latestUnlockNever: t("latestUnlockNever"),
        sourceOffline: t("sourceOffline"),
        labTitle: t("labTitle"),
        labProjectOneTitle: t("labProjectOneTitle"),
        labProjectOneDescription: t("labProjectOneDescription"),
        labProjectTwoTitle: t("labProjectTwoTitle"),
        labProjectTwoDescription: t("labProjectTwoDescription"),
        labProjectThreeTitle: t("labProjectThreeTitle"),
        labProjectThreeDescription: t("labProjectThreeDescription"),
        dossierTitle: t("dossierTitle"),
        dossierBody: t("dossierBody"),
        dossierLocationLabel: t("dossierLocationLabel"),
        dossierLocation: t("dossierLocation"),
        dossierContactLabel: t("dossierContactLabel"),
        dossierContact: t("dossierContact"),
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

import { use } from "react";
import { cookies } from "next/headers";
import { useTranslations } from "next-intl";
import { setRequestLocale } from "next-intl/server";

import { HeroSection } from "@/components/hero-section";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { SentinelTerminal } from "@/components/sentinel-terminal";
import { verifyUnlockCookieValue } from "@/lib/security/unlock-cookie";

type Props = {
  params: Promise<{ locale: string }>;
};

export default function LocaleHomePage({ params }: Props) {
  const { locale } = use(params);
  setRequestLocale(locale);

  const t = useTranslations("home");
  const cookieStore = use(cookies());
  const unlockCookie = cookieStore.get("karot_unlock")?.value;
  const secret = process.env.UNLOCK_COOKIE_SECRET;
  const hasUnlockedBefore = Boolean(unlockCookie && secret && verifyUnlockCookieValue(unlockCookie, secret));

  return (
    <main className="flex flex-1 px-6 py-10 sm:py-16">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 sm:gap-10">
        <div className="flex justify-end">
          <LocaleSwitcher />
        </div>

        <HeroSection
          badge={t("phaseLabel")}
          intro={t("intro")}
          body1={t("body1")}
          body2={t("body2")}
          body3={t("body3")}
        />

        <SentinelTerminal
          title={t("sentinel.title")}
          subtitle={t("sentinel.subtitle")}
          meterLabel={t("sentinel.meterLabel")}
          inputPlaceholder={t("sentinel.inputPlaceholder")}
          sendLabel={t("sentinel.sendLabel")}
          resetLabel={t("sentinel.resetLabel")}
          initialMessage={t("sentinel.initialMessage")}
          unlockedLabel={t("sentinel.unlockedLabel")}
          unlockedCta={t("sentinel.unlockedCta")}
          pendingLabel={t("sentinel.pendingLabel")}
          errorLabel={t("sentinel.errorLabel")}
          hasUnlockedBefore={hasUnlockedBefore}
          returnOverlayTitle={t("sentinel.returnOverlayTitle")}
          returnOverlayBody={t("sentinel.returnOverlayBody")}
          playAgainLabel={t("sentinel.playAgainLabel")}
          goDashboardLabel={t("sentinel.goDashboardLabel")}
        />
      </div>
    </main>
  );
}

import { useTranslations } from "next-intl";

import { HeroSection } from "@/components/hero-section";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { SentinelTerminal } from "@/components/sentinel-terminal";

export default function LocaleHomePage() {
  const t = useTranslations("home");

  return (
    <main className="flex flex-1 px-6 py-10 sm:py-16">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 sm:gap-10">
        <div className="flex justify-end">
          <LocaleSwitcher />
        </div>

        <HeroSection
          badge={t("phaseLabel")}
          title={t("title")}
          description={t("description")}
          primaryCta={t("primaryCta")}
          secondaryCta={t("secondaryCta")}
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
        />
      </div>
    </main>
  );
}

"use client";

import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { useCookieConsent } from "@/modules/cookie-consent";

export function CookieConsentSettingsTrigger() {
  const t = useTranslations("cookieConsent");
  const { openSettings } = useCookieConsent();

  return (
    <Button
      type="button"
      variant="link"
      className="h-auto p-0 text-sm"
      onClick={openSettings}
      aria-label={t("settingsTriggerAriaLabel")}
    >
      {t("settingsTrigger")}
    </Button>
  );
}

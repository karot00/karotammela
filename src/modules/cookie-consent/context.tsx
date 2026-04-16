"use client";

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";

import { CONSENT_SCHEMA_VERSION } from "@/modules/cookie-consent/constants";
import { isBannerRequired } from "@/modules/cookie-consent/gate";
import {
  createDefaultConsentState,
  readConsentFromBrowser,
  writeConsentToBrowser,
} from "@/modules/cookie-consent/storage";
import type {
  ConsentCategoryPreferences,
  ConsentState,
} from "@/modules/cookie-consent/types";

type CookieConsentProviderProps = PropsWithChildren<{
  initialConsent?: ConsentState;
}>;

type CookieConsentContextValue = {
  consent: ConsentState;
  isSettingsOpen: boolean;
  isBannerVisible: boolean;
  acceptAll: () => void;
  rejectNonEssential: () => void;
  savePreferences: (categories: Partial<ConsentCategoryPreferences>) => void;
  openSettings: () => void;
  closeSettings: () => void;
};

const CookieConsentContext = createContext<CookieConsentContextValue | null>(
  null,
);

function buildConsentState(
  categories: Partial<ConsentCategoryPreferences>,
): ConsentState {
  return {
    version: CONSENT_SCHEMA_VERSION,
    updatedAt: new Date().toISOString(),
    categories: {
      essential: true,
      functional: categories.functional === true,
      analytics: categories.analytics === true,
      marketing: categories.marketing === true,
    },
  };
}

export function CookieConsentProvider({
  children,
  initialConsent,
}: CookieConsentProviderProps) {
  const [consent, setConsent] = useState<ConsentState>(
    () =>
      initialConsent ??
      (typeof document === "undefined"
        ? createDefaultConsentState()
        : readConsentFromBrowser()),
  );
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);

  const persistConsent = useCallback((nextState: ConsentState) => {
    setConsent(nextState);
    writeConsentToBrowser(nextState);
  }, []);

  const acceptAll = useCallback(() => {
    persistConsent(
      buildConsentState({
        essential: true,
        functional: true,
        analytics: true,
        marketing: true,
      }),
    );
    setIsSettingsOpen(false);
  }, [persistConsent]);

  const rejectNonEssential = useCallback(() => {
    persistConsent(
      buildConsentState({
        essential: true,
        functional: false,
        analytics: false,
        marketing: false,
      }),
    );
    setIsSettingsOpen(false);
  }, [persistConsent]);

  const savePreferences = useCallback(
    (categories: Partial<ConsentCategoryPreferences>) => {
      persistConsent(
        buildConsentState({
          ...consent.categories,
          ...categories,
          essential: true,
        }),
      );
      setIsSettingsOpen(false);
    },
    [consent.categories, persistConsent],
  );

  const openSettings = useCallback(() => {
    setIsSettingsOpen(true);
  }, []);

  const closeSettings = useCallback(() => {
    setIsSettingsOpen(false);
  }, []);

  const value = useMemo<CookieConsentContextValue>(
    () => ({
      consent,
      isSettingsOpen,
      isBannerVisible: isBannerRequired(consent),
      acceptAll,
      rejectNonEssential,
      savePreferences,
      openSettings,
      closeSettings,
    }),
    [
      acceptAll,
      closeSettings,
      consent,
      isSettingsOpen,
      openSettings,
      rejectNonEssential,
      savePreferences,
    ],
  );

  return (
    <CookieConsentContext.Provider value={value}>
      {children}
    </CookieConsentContext.Provider>
  );
}

export function useCookieConsent(): CookieConsentContextValue {
  const context = useContext(CookieConsentContext);

  if (!context) {
    throw new Error(
      "useCookieConsent must be used within CookieConsentProvider",
    );
  }

  return context;
}

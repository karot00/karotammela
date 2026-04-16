"use client";

import {
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import {
  CONSENT_CATEGORY_CONFIG,
  CONSENT_INVENTORY,
  useCookieConsent,
  type ConsentCategory,
} from "@/modules/cookie-consent";

const FOCUSABLE_SELECTOR =
  "a[href], button:not([disabled]), textarea, input, select, [tabindex]:not([tabindex='-1'])";

export function CookieConsentPreferences() {
  const t = useTranslations("cookieConsent");
  const {
    consent,
    isSettingsOpen,
    closeSettings,
    savePreferences,
    rejectNonEssential,
    acceptAll,
  } = useCookieConsent();

  const titleId = useId();
  const panelRef = useRef<HTMLDivElement | null>(null);

  const [draftCategories, setDraftCategories] = useState(consent.categories);
  const [hasDraftChanges, setHasDraftChanges] = useState(false);

  const currentDraft = hasDraftChanges ? draftCategories : consent.categories;

  const resetDraft = useCallback(() => {
    setHasDraftChanges(false);
    setDraftCategories(consent.categories);
  }, [consent.categories]);

  const handleCloseSettings = useCallback(() => {
    resetDraft();
    closeSettings();
  }, [closeSettings, resetDraft]);

  useEffect(() => {
    if (!isSettingsOpen) {
      return;
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    const panel = panelRef.current;
    const firstFocusable =
      panel?.querySelector<HTMLElement>(FOCUSABLE_SELECTOR);
    firstFocusable?.focus();

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        handleCloseSettings();
        return;
      }

      if (event.key !== "Tab") {
        return;
      }

      const focusableNodes = panel
        ? Array.from(panel.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR))
        : [];

      if (focusableNodes.length === 0) {
        return;
      }

      const first = focusableNodes[0];
      const last = focusableNodes[focusableNodes.length - 1];
      const activeElement = document.activeElement as HTMLElement | null;

      if (!event.shiftKey && activeElement === last) {
        event.preventDefault();
        first.focus();
      }

      if (event.shiftKey && activeElement === first) {
        event.preventDefault();
        last.focus();
      }
    }

    document.addEventListener("keydown", onKeyDown);

    return () => {
      document.removeEventListener("keydown", onKeyDown);
      document.body.style.overflow = previousOverflow;
    };
  }, [handleCloseSettings, isSettingsOpen]);

  const groupedInventory = useMemo(() => {
    return CONSENT_CATEGORY_CONFIG.map((categoryConfig) => ({
      category: categoryConfig.id,
      required: categoryConfig.required,
      title: t(categoryConfig.titleKey),
      description: t(categoryConfig.descriptionKey),
      items: CONSENT_INVENTORY.filter(
        (item) => item.category === categoryConfig.id,
      ),
    }));
  }, [t]);

  if (!isSettingsOpen) {
    return null;
  }

  return (
    <div
      className="fixed inset-0 z-[60] flex items-end justify-center overflow-y-auto bg-black/55 p-3 sm:items-center sm:p-4"
      role="presentation"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          handleCloseSettings();
        }
      }}
    >
      <section
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className="my-auto flex max-h-[96svh] w-full max-w-3xl flex-col overflow-hidden rounded-2xl border border-border bg-card shadow-[0_24px_60px_rgba(0,0,0,0.42)]"
      >
        <header className="border-b border-border/80 px-5 py-4 sm:px-6">
          <p className="text-[0.7rem] font-semibold tracking-[0.16em] text-primary uppercase">
            {t("preferences.eyebrow")}
          </p>
          <h2
            id={titleId}
            className="mt-1 text-lg font-semibold text-card-foreground"
          >
            {t("preferences.title")}
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {t("preferences.description")}
          </p>
        </header>

        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-4 pb-5 sm:px-6 sm:pb-6">
          {groupedInventory.map((entry) => {
            const toggleId = `consent-toggle-${entry.category}`;
            const enabled = currentDraft[entry.category as ConsentCategory];

            return (
              <article
                key={entry.category}
                className="rounded-xl border border-border bg-muted/20 p-4"
              >
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <h3 className="text-sm font-semibold text-card-foreground">
                      {entry.title}
                    </h3>
                    <p className="mt-1 text-sm text-muted-foreground">
                      {entry.description}
                    </p>
                  </div>

                  <label
                    htmlFor={toggleId}
                    className="inline-flex items-center gap-3 self-start"
                  >
                    <input
                      id={toggleId}
                      type="checkbox"
                      checked={enabled}
                      disabled={entry.required}
                      onChange={(event) => {
                        const nextChecked = event.currentTarget.checked;

                        setHasDraftChanges(true);
                        setDraftCategories((previous) => ({
                          ...previous,
                          [entry.category]: entry.required ? true : nextChecked,
                        }));
                      }}
                      className="h-4 w-4 rounded border-border accent-primary disabled:cursor-not-allowed disabled:opacity-70"
                    />
                    <span className="text-xs font-medium text-muted-foreground">
                      {entry.required
                        ? t("categories.alwaysActive")
                        : t("categories.optional")}
                    </span>
                  </label>
                </div>

                {entry.items.length > 0 ? (
                  <details className="mt-3 rounded-lg border border-border/80 bg-card/60">
                    <summary className="cursor-pointer px-3 py-2 text-xs font-medium text-muted-foreground sm:text-sm">
                      {t("inventory.detailsLabel")}
                    </summary>
                    <div className="overflow-hidden border-t border-border/80">
                      <table className="w-full border-collapse text-left text-xs sm:text-sm">
                        <thead className="bg-muted/40 text-muted-foreground">
                          <tr>
                            <th className="px-3 py-2 font-medium">
                              {t("inventory.name")}
                            </th>
                            <th className="px-3 py-2 font-medium">
                              {t("inventory.provider")}
                            </th>
                            <th className="px-3 py-2 font-medium">
                              {t("inventory.purpose")}
                            </th>
                            <th className="px-3 py-2 font-medium">
                              {t("inventory.duration")}
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          {entry.items.map((item) => (
                            <tr
                              key={item.name}
                              className="border-t border-border/80"
                            >
                              <td className="px-3 py-2 text-card-foreground">
                                {item.name}
                              </td>
                              <td className="px-3 py-2 text-muted-foreground">
                                {item.provider}
                              </td>
                              <td className="px-3 py-2 text-muted-foreground">
                                {item.purpose}
                              </td>
                              <td className="px-3 py-2 text-muted-foreground">
                                {item.duration}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </details>
                ) : (
                  <p className="mt-3 text-xs text-muted-foreground">
                    {t("inventory.emptyCategory")}
                  </p>
                )}
              </article>
            );
          })}
        </div>

        <footer className="sticky bottom-0 flex flex-wrap items-center justify-end gap-2 border-t border-border/80 bg-card px-5 py-4 pb-[max(1rem,env(safe-area-inset-bottom))] sm:px-6">
          <Button
            type="button"
            size="sm"
            variant="ghost"
            onClick={handleCloseSettings}
          >
            {t("actions.cancel")}
          </Button>
          <Button
            type="button"
            size="sm"
            variant="secondary"
            onClick={rejectNonEssential}
          >
            {t("actions.rejectNonEssential")}
          </Button>
          <Button
            type="button"
            size="sm"
            variant="secondary"
            onClick={acceptAll}
          >
            {t("actions.acceptAll")}
          </Button>
          <Button
            type="button"
            size="sm"
            onClick={() => {
              savePreferences(currentDraft);
              setHasDraftChanges(false);
            }}
          >
            {t("actions.savePreferences")}
          </Button>
        </footer>
      </section>
    </div>
  );
}

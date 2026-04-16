import type { ConsentCategory } from "@/modules/cookie-consent/types";

export type ConsentCategoryConfig = {
  id: ConsentCategory;
  required: boolean;
  titleKey: string;
  descriptionKey: string;
};

export type ConsentStorageType = "cookie" | "localStorage" | "sessionStorage";

export type ConsentInventoryItem = {
  name: string;
  provider: string;
  purpose: string;
  duration: string;
  type: ConsentStorageType;
  category: ConsentCategory;
  required: boolean;
};

export const CONSENT_CATEGORY_CONFIG: ReadonlyArray<ConsentCategoryConfig> = [
  {
    id: "essential",
    required: true,
    titleKey: "categories.essential.title",
    descriptionKey: "categories.essential.description",
  },
  {
    id: "functional",
    required: false,
    titleKey: "categories.functional.title",
    descriptionKey: "categories.functional.description",
  },
  {
    id: "analytics",
    required: false,
    titleKey: "categories.analytics.title",
    descriptionKey: "categories.analytics.description",
  },
  {
    id: "marketing",
    required: false,
    titleKey: "categories.marketing.title",
    descriptionKey: "categories.marketing.description",
  },
];

export const CONSENT_INVENTORY: ReadonlyArray<ConsentInventoryItem> = [
  {
    name: "karot_unlock",
    provider: "karotammela.fi",
    purpose: "Sentinel unlock verification for dashboard access",
    duration: "14 days",
    type: "cookie",
    category: "essential",
    required: true,
  },
  {
    name: "NEXT_LOCALE",
    provider: "karotammela.fi",
    purpose: "Persist selected language between visits",
    duration: "Session and refresh persistence",
    type: "cookie",
    category: "functional",
    required: false,
  },
  {
    name: "karot-theme",
    provider: "karotammela.fi",
    purpose: "Persist preferred visual theme",
    duration: "Until cleared",
    type: "localStorage",
    category: "functional",
    required: false,
  },
  {
    name: "Umami cookieless analytics",
    provider: "Umami",
    purpose: "Measure page views and engagement without cookies",
    duration: "No cookies used",
    type: "cookie",
    category: "analytics",
    required: false,
  },
];

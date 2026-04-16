import type { ConsentState } from "@/modules/cookie-consent/types";

export const CONSENT_COOKIE_NAME = "karot_consent";
export const CONSENT_SCHEMA_VERSION = 1;
export const CONSENT_COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 180;
export const CONSENT_UNSET_UPDATED_AT = "1970-01-01T00:00:00.000Z";

export const DEFAULT_CONSENT_CATEGORIES: ConsentState["categories"] = {
  essential: true,
  functional: false,
  analytics: false,
  marketing: false,
};

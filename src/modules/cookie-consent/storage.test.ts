import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  CONSENT_COOKIE_NAME,
  CONSENT_SCHEMA_VERSION,
  CONSENT_UNSET_UPDATED_AT,
} from "@/modules/cookie-consent/constants";
import {
  createDefaultConsentState,
  parseConsentState,
  readConsentFromDocumentCookie,
  serializeConsentState,
} from "@/modules/cookie-consent/storage";

describe("cookie consent storage", () => {
  it("returns default state when cookie value is missing", () => {
    const parsed = parseConsentState(undefined);

    assert.deepEqual(parsed, createDefaultConsentState());
  });

  it("returns default state for invalid payload", () => {
    const parsed = parseConsentState("%7Bbroken");

    assert.deepEqual(parsed, createDefaultConsentState());
  });

  it("normalizes legacy payload to current schema version", () => {
    const now = new Date("2026-01-12T08:00:00.000Z");
    const legacyPayload = encodeURIComponent(
      JSON.stringify({
        version: 0,
        categories: {
          essential: false,
          functional: true,
          analytics: false,
          marketing: true,
        },
      }),
    );

    const parsed = parseConsentState(legacyPayload, now);

    assert.equal(parsed.version, CONSENT_SCHEMA_VERSION);
    assert.equal(parsed.updatedAt, now.toISOString());
    assert.deepEqual(parsed.categories, {
      essential: true,
      functional: true,
      analytics: false,
      marketing: true,
    });
  });

  it("roundtrips consent state through serializer and parser", () => {
    const source = {
      version: CONSENT_SCHEMA_VERSION,
      updatedAt: "2026-02-03T11:22:33.000Z",
      categories: {
        essential: true,
        functional: false,
        analytics: true,
        marketing: false,
      },
    };

    const serialized = serializeConsentState(source);
    const parsed = parseConsentState(serialized);

    assert.deepEqual(parsed, source);
  });

  it("reads consent cookie from document.cookie string", () => {
    const consentValue = serializeConsentState({
      version: CONSENT_SCHEMA_VERSION,
      updatedAt: "2026-03-21T10:00:00.000Z",
      categories: {
        essential: true,
        functional: true,
        analytics: true,
        marketing: false,
      },
    });
    const documentCookie = `theme=dark; ${CONSENT_COOKIE_NAME}=${consentValue}; other=1`;

    const parsed = readConsentFromDocumentCookie(documentCookie);

    assert.equal(parsed.categories.functional, true);
    assert.equal(parsed.categories.analytics, true);
    assert.equal(parsed.categories.marketing, false);
  });

  it("keeps unset marker for default state", () => {
    const defaultState = createDefaultConsentState();

    assert.equal(defaultState.updatedAt, CONSENT_UNSET_UPDATED_AT);
  });
});

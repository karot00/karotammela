import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  CONSENT_SCHEMA_VERSION,
  CONSENT_UNSET_UPDATED_AT,
} from "@/modules/cookie-consent/constants";
import {
  isBannerRequired,
  isCategoryAllowed,
} from "@/modules/cookie-consent/gate";
import type { ConsentState } from "@/modules/cookie-consent/types";

function createState(overrides?: Partial<ConsentState>): ConsentState {
  return {
    version: CONSENT_SCHEMA_VERSION,
    updatedAt: "2026-04-10T09:15:00.000Z",
    categories: {
      essential: true,
      functional: false,
      analytics: false,
      marketing: false,
    },
    ...overrides,
  };
}

describe("cookie consent gate", () => {
  it("always allows essential category", () => {
    const state = createState({
      categories: {
        essential: true,
        functional: false,
        analytics: false,
        marketing: false,
      },
    });

    assert.equal(isCategoryAllowed(state, "essential"), true);
  });

  it("respects optional category toggles", () => {
    const state = createState({
      categories: {
        essential: true,
        functional: true,
        analytics: false,
        marketing: true,
      },
    });

    assert.equal(isCategoryAllowed(state, "functional"), true);
    assert.equal(isCategoryAllowed(state, "analytics"), false);
    assert.equal(isCategoryAllowed(state, "marketing"), true);
  });

  it("requires banner when consent is unset", () => {
    const state = createState({ updatedAt: CONSENT_UNSET_UPDATED_AT });

    assert.equal(isBannerRequired(state), true);
  });

  it("hides banner after consent transitions (accept or reject)", () => {
    const rejectedState = createState({
      categories: {
        essential: true,
        functional: false,
        analytics: false,
        marketing: false,
      },
    });
    const acceptedState = createState({
      categories: {
        essential: true,
        functional: true,
        analytics: true,
        marketing: true,
      },
    });

    assert.equal(isBannerRequired(rejectedState), false);
    assert.equal(isBannerRequired(acceptedState), false);
  });
});

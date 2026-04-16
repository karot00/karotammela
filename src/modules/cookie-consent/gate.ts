import { CONSENT_UNSET_UPDATED_AT } from "@/modules/cookie-consent/constants";
import type {
  ConsentCategory,
  ConsentState,
} from "@/modules/cookie-consent/types";

export function isCategoryAllowed(
  state: ConsentState,
  category: ConsentCategory,
): boolean {
  if (category === "essential") {
    return true;
  }

  return state.categories[category] === true;
}

export function isBannerRequired(state: ConsentState): boolean {
  return state.updatedAt === CONSENT_UNSET_UPDATED_AT;
}

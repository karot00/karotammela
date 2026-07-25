package security

import (
	"encoding/json"
	"net/url"
	"time"
)

// Consent cookie contract — a faithful port of the Next.js reference
// (src/modules/cookie-consent/constants.ts, storage.ts, gate.ts) so the same
// `karot_consent` cookie is honored identically by both stacks behind the
// Tech Switcher. The value is percent-encoded JSON written by client-side
// code; it is intentionally not signed because it carries display
// preferences only (no authorization), matching the reference.

const (
	// ConsentCookieName is the shared cookie key (CONSENT_COOKIE_NAME).
	ConsentCookieName = "karot_consent"
	// ConsentSchemaVersion is the current payload schema
	// (CONSENT_SCHEMA_VERSION). Older payloads are normalized up.
	ConsentSchemaVersion = 1
	// ConsentCookieMaxAgeSeconds is the 180-day cookie lifetime
	// (CONSENT_COOKIE_MAX_AGE_SECONDS).
	ConsentCookieMaxAgeSeconds = 60 * 60 * 24 * 180
	// ConsentUnsetUpdatedAt marks "no decision stored yet"
	// (CONSENT_UNSET_UPDATED_AT). The banner is required exactly while
	// UpdatedAt equals this sentinel.
	ConsentUnsetUpdatedAt = "1970-01-01T00:00:00.000Z"
)

// consentTimestampFormat matches Date#toISOString() output (millisecond
// precision, UTC "Z"), which both stacks write.
const consentTimestampFormat = "2006-01-02T15:04:05.000Z"

// ConsentCategories holds the four preference buckets. Essential is always
// true: both stacks force it on normalize and on write.
type ConsentCategories struct {
	Essential  bool `json:"essential"`
	Functional bool `json:"functional"`
	Analytics  bool `json:"analytics"`
	Marketing  bool `json:"marketing"`
}

// ConsentState is the karot_consent JSON payload.
type ConsentState struct {
	Version    int               `json:"version"`
	UpdatedAt  string            `json:"updatedAt"`
	Categories ConsentCategories `json:"categories"`
}

// DefaultConsentState mirrors createDefaultConsentState(): unset marker, all
// optional categories off.
func DefaultConsentState() ConsentState {
	return ConsentState{
		Version:   ConsentSchemaVersion,
		UpdatedAt: ConsentUnsetUpdatedAt,
		Categories: ConsentCategories{
			Essential: true,
		},
	}
}

// partialConsent mirrors the lenient PartialConsentInput shape: UpdatedAt is
// `any` so non-string garbage normalizes like the TypeScript
// `typeof x === "string"` check, and Categories is a json.RawMessage so a
// payload without the key ("no legacy shape") falls back to the default
// state while `"categories": null` still counts as a stored shape.
type partialConsent struct {
	UpdatedAt  any             `json:"updatedAt"`
	Categories json.RawMessage `json:"categories"`
}

type partialConsentCategories struct {
	Functional any `json:"functional"`
	Analytics  any `json:"analytics"`
	Marketing  any `json:"marketing"`
}

// ParseConsentState mirrors parseConsentState(): any missing, malformed, or
// shape-less value yields the default state; recognized payloads are
// normalized to the current schema with essential forced on. `now` is used
// when the stored updatedAt is absent or unparseable, matching the
// reference's `Date.parse` fallback. Values written by either stack are ISO
// (RFC3339) timestamps, which is what the parse accepts.
func ParseConsentState(value string, now time.Time) ConsentState {
	if value == "" {
		return DefaultConsentState()
	}
	// PathUnescape matches decodeURIComponent semantics (percent-decoding
	// without treating '+' as space).
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return DefaultConsentState()
	}
	var parsed partialConsent
	if err := json.Unmarshal([]byte(decoded), &parsed); err != nil {
		return DefaultConsentState()
	}
	if parsed.Categories == nil {
		return DefaultConsentState()
	}

	var cats partialConsentCategories
	// A null or non-object "categories" leaves zero values, matching the
	// reference's optional-chaining normalization.
	_ = json.Unmarshal(parsed.Categories, &cats)

	updatedAt := now.UTC().Format(consentTimestampFormat)
	if s, ok := parsed.UpdatedAt.(string); ok {
		if ts, err := time.Parse(time.RFC3339, s); err == nil {
			updatedAt = ts.UTC().Format(consentTimestampFormat)
		}
	}

	return ConsentState{
		Version:   ConsentSchemaVersion,
		UpdatedAt: updatedAt,
		Categories: ConsentCategories{
			Essential:  true,
			Functional: asBool(cats.Functional),
			Analytics:  asBool(cats.Analytics),
			Marketing:  asBool(cats.Marketing),
		},
	}
}

// SerializeConsentState mirrors serializeConsentState(): JSON, then
// percent-encoding. Used by tests and any future server-side writer; both
// stacks' parsers accept each other's encoding.
func SerializeConsentState(state ConsentState) string {
	b, err := json.Marshal(state)
	if err != nil {
		return ""
	}
	return url.PathEscape(string(b))
}

// IsConsentCategoryAllowed mirrors isCategoryAllowed(): essential is always
// allowed; optional categories require an explicit true.
func IsConsentCategoryAllowed(state ConsentState, category string) bool {
	switch category {
	case "essential":
		return true
	case "functional":
		return state.Categories.Functional
	case "analytics":
		return state.Categories.Analytics
	case "marketing":
		return state.Categories.Marketing
	}
	return false
}

// IsConsentBannerRequired mirrors isBannerRequired(): the banner shows only
// until the first stored decision (updatedAt moves off the unset sentinel).
func IsConsentBannerRequired(state ConsentState) bool {
	return state.UpdatedAt == ConsentUnsetUpdatedAt
}

func asBool(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

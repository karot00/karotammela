package security

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"
)

// The cases below are ported from src/modules/cookie-consent/storage.test.ts
// and gate.test.ts so the Go stack normalizes and gates the shared
// karot_consent cookie exactly like the Next.js reference.

func TestParseConsentStateReturnsDefaultWhenMissing(t *testing.T) {
	parsed := ParseConsentState("", time.Now())
	if parsed != DefaultConsentState() {
		t.Fatalf("expected default state, got %+v", parsed)
	}
}

func TestParseConsentStateReturnsDefaultForInvalidPayload(t *testing.T) {
	parsed := ParseConsentState("%7Bbroken", time.Now())
	if parsed != DefaultConsentState() {
		t.Fatalf("expected default state, got %+v", parsed)
	}
}

func TestParseConsentStateNormalizesLegacyPayload(t *testing.T) {
	now := time.Date(2026, 1, 12, 8, 0, 0, 0, time.UTC)
	legacyPayload := url.PathEscape(`{"version":0,"categories":{"essential":false,"functional":true,"analytics":false,"marketing":true}}`)

	parsed := ParseConsentState(legacyPayload, now)

	if parsed.Version != ConsentSchemaVersion {
		t.Fatalf("version = %d, want %d", parsed.Version, ConsentSchemaVersion)
	}
	if parsed.UpdatedAt != "2026-01-12T08:00:00.000Z" {
		t.Fatalf("updatedAt = %q, want %q", parsed.UpdatedAt, "2026-01-12T08:00:00.000Z")
	}
	want := ConsentCategories{Essential: true, Functional: true, Analytics: false, Marketing: true}
	if parsed.Categories != want {
		t.Fatalf("categories = %+v, want %+v", parsed.Categories, want)
	}
}

func TestConsentStateRoundTripsThroughSerializerAndParser(t *testing.T) {
	source := ConsentState{
		Version:   ConsentSchemaVersion,
		UpdatedAt: "2026-02-03T11:22:33.000Z",
		Categories: ConsentCategories{
			Essential:  true,
			Functional: false,
			Analytics:  true,
			Marketing:  false,
		},
	}

	parsed := ParseConsentState(SerializeConsentState(source), time.Now())
	if parsed != source {
		t.Fatalf("roundtrip = %+v, want %+v", parsed, source)
	}
}

func TestDefaultConsentStateKeepsUnsetMarker(t *testing.T) {
	if DefaultConsentState().UpdatedAt != ConsentUnsetUpdatedAt {
		t.Fatalf("default updatedAt = %q, want unset sentinel %q",
			DefaultConsentState().UpdatedAt, ConsentUnsetUpdatedAt)
	}
}

// "categories": null counts as a stored shape in the reference (the key is
// present), so the payload normalizes instead of falling back to the
// unset sentinel — and the banner therefore stays hidden.
func TestParseConsentStateNullCategoriesIsAStoredShape(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	value := url.PathEscape(`{"updatedAt":"2026-04-10T09:15:00.000Z","categories":null}`)

	parsed := ParseConsentState(value, now)

	if parsed.UpdatedAt != "2026-04-10T09:15:00.000Z" {
		t.Fatalf("updatedAt = %q, want stored value", parsed.UpdatedAt)
	}
	want := ConsentCategories{Essential: true}
	if parsed.Categories != want {
		t.Fatalf("categories = %+v, want %+v", parsed.Categories, want)
	}
	if IsConsentBannerRequired(parsed) {
		t.Fatal("banner must stay hidden once a categories shape was stored")
	}
}

func TestParseConsentStateDropsNonBooleanGarbage(t *testing.T) {
	value := url.PathEscape(`{"updatedAt":"2026-02-03T11:22:33.000Z","categories":{"functional":"yes","analytics":1,"marketing":true}}`)

	parsed := ParseConsentState(value, time.Now())

	want := ConsentCategories{Essential: true, Functional: false, Analytics: false, Marketing: true}
	if parsed.Categories != want {
		t.Fatalf("categories = %+v, want %+v", parsed.Categories, want)
	}
	if parsed.UpdatedAt != "2026-02-03T11:22:33.000Z" {
		t.Fatalf("updatedAt = %q, want stored value", parsed.UpdatedAt)
	}
}

func TestParseConsentStateReplacesUnparseableUpdatedAtWithNow(t *testing.T) {
	now := time.Date(2026, 7, 24, 1, 2, 3, 0, time.UTC)
	value := url.PathEscape(`{"updatedAt":"not-a-date","categories":{"functional":true}}`)

	parsed := ParseConsentState(value, now)

	if parsed.UpdatedAt != "2026-07-24T01:02:03.000Z" {
		t.Fatalf("updatedAt = %q, want now fallback", parsed.UpdatedAt)
	}
}

// Serialized output must be decodable by the reference parser contract:
// percent-encoded JSON with the four category keys.
func TestSerializeConsentStateProducesPercentEncodedJSON(t *testing.T) {
	state := ConsentState{
		Version:   ConsentSchemaVersion,
		UpdatedAt: "2026-03-21T10:00:00.000Z",
		Categories: ConsentCategories{
			Essential: true, Functional: true, Analytics: true, Marketing: false,
		},
	}

	serialized := SerializeConsentState(state)
	decoded, err := url.PathUnescape(serialized)
	if err != nil {
		t.Fatalf("serialized value is not percent-encoded: %v", err)
	}
	var probe map[string]any
	if err := json.Unmarshal([]byte(decoded), &probe); err != nil {
		t.Fatalf("decoded value is not JSON: %v", err)
	}
	if _, ok := probe["categories"]; !ok {
		t.Fatal("serialized payload is missing the categories key")
	}
}

func consentGateState(updatedAt string, cats ConsentCategories) ConsentState {
	return ConsentState{Version: ConsentSchemaVersion, UpdatedAt: updatedAt, Categories: cats}
}

func TestIsConsentCategoryAllowedAlwaysAllowsEssential(t *testing.T) {
	state := consentGateState("2026-04-10T09:15:00.000Z", ConsentCategories{Essential: true})
	if !IsConsentCategoryAllowed(state, "essential") {
		t.Fatal("essential must always be allowed")
	}
}

func TestIsConsentCategoryAllowedRespectsOptionalToggles(t *testing.T) {
	state := consentGateState("2026-04-10T09:15:00.000Z", ConsentCategories{
		Essential: true, Functional: true, Analytics: false, Marketing: true,
	})
	if !IsConsentCategoryAllowed(state, "functional") {
		t.Fatal("functional should be allowed")
	}
	if IsConsentCategoryAllowed(state, "analytics") {
		t.Fatal("analytics should be denied")
	}
	if !IsConsentCategoryAllowed(state, "marketing") {
		t.Fatal("marketing should be allowed")
	}
}

func TestIsConsentBannerRequiredWhenUnset(t *testing.T) {
	state := consentGateState(ConsentUnsetUpdatedAt, ConsentCategories{Essential: true})
	if !IsConsentBannerRequired(state) {
		t.Fatal("banner must be required while consent is unset")
	}
}

func TestIsConsentBannerHiddenAfterAcceptOrReject(t *testing.T) {
	rejected := consentGateState("2026-04-10T09:15:00.000Z", ConsentCategories{Essential: true})
	accepted := consentGateState("2026-04-10T09:15:00.000Z", ConsentCategories{
		Essential: true, Functional: true, Analytics: true, Marketing: true,
	})
	if IsConsentBannerRequired(rejected) {
		t.Fatal("banner must hide after rejection")
	}
	if IsConsentBannerRequired(accepted) {
		t.Fatal("banner must hide after acceptance")
	}
}

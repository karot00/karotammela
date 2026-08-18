package security

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

var vipCookieTestNow = time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)

const vipCookieTestSecret = "vip-cookie-test-secret-0123456789abcdef"

func TestVIPCookieRoundTrip(t *testing.T) {
	value := CreateVIPCookieValue(vipCookieTestSecret, vipCookieTestNow)
	if value == "" {
		t.Fatal("CreateVIPCookieValue returned empty value")
	}
	if strings.Count(value, ".") != 1 {
		t.Fatalf("value %q is not payload.signature", value)
	}

	got := VerifyVIPCookieValue(value, vipCookieTestSecret, vipCookieTestNow)
	if got == nil {
		t.Fatal("fresh cookie rejected")
	}
	if got.Version != VIPCookieVersion {
		t.Errorf("Version = %d, want %d", got.Version, VIPCookieVersion)
	}
	if got.IssuedAt != vipCookieTestNow.Unix() {
		t.Errorf("IssuedAt = %d, want %d", got.IssuedAt, vipCookieTestNow.Unix())
	}
	wantExp := vipCookieTestNow.Add(VIPCookieLifetime).Unix()
	if got.ExpiresAt != wantExp {
		t.Errorf("ExpiresAt = %d, want %d", got.ExpiresAt, wantExp)
	}

	// Valid until the last second before expiry.
	if VerifyVIPCookieValue(value, vipCookieTestSecret, time.Unix(wantExp-1, 0)) == nil {
		t.Error("cookie rejected one second before expiry")
	}
	// Dead at and after expiry (server-side, not browser-side).
	if VerifyVIPCookieValue(value, vipCookieTestSecret, time.Unix(wantExp, 0)) != nil {
		t.Error("cookie accepted exactly at expiry")
	}
	if VerifyVIPCookieValue(value, vipCookieTestSecret, time.Unix(wantExp+1, 0)) != nil {
		t.Error("cookie accepted after expiry")
	}
}

// TestVIPCookieGoldenVector pins the wire format so an accidental change to
// payload encoding, key order, or signing input is caught. Vector computed
// with secret/now below (see cmd note in the PR that introduced it).
func TestVIPCookieGoldenVector(t *testing.T) {
	value := CreateVIPCookieValue(vipCookieTestSecret, vipCookieTestNow)
	const want = "eyJ2IjoxLCJpYXQiOjE3ODcwNTQ0MDAsImV4cCI6MTc4NzE0MDgwMH0.EY2nP-4FKw_qRikRnUiVG2U5NAfheNNGHeUvrO7Y8Hw"
	if value != want {
		t.Fatalf("cookie value drifted:\n got: %s\nwant: %s", value, want)
	}
}

func TestVIPCookieRotationInvalidatesOldCookies(t *testing.T) {
	value := CreateVIPCookieValue("old-secret-old-secret-old-secret-0123", vipCookieTestNow)
	if VerifyVIPCookieValue(value, "new-secret-new-secret-new-secret-0123", vipCookieTestNow) != nil {
		t.Error("cookie accepted after VIP_COOKIE_SECRET rotation")
	}
}

func TestVIPCookieTamperRejected(t *testing.T) {
	value := CreateVIPCookieValue(vipCookieTestSecret, vipCookieTestNow)
	parts := strings.SplitN(value, ".", 2)

	tamperPayload := func(mutate func(p *VIPCookiePayload)) string {
		raw, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		var p VIPCookiePayload
		if err := json.Unmarshal(raw, &p); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		mutate(&p)
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return base64.RawURLEncoding.EncodeToString(b) + "." + parts[1]
	}

	cases := map[string]string{
		"extended expiry":   tamperPayload(func(p *VIPCookiePayload) { p.ExpiresAt += 86400 }),
		"backdated issued":  tamperPayload(func(p *VIPCookiePayload) { p.IssuedAt -= 86400 }),
		"version bump":      tamperPayload(func(p *VIPCookiePayload) { p.Version++ }),
		"signature flipped": parts[0] + "." + parts[1][:len(parts[1])-2] + "AA",
		"signature empty":   parts[0] + ".",
		"payload empty":     "." + parts[1],
		"no separator":      parts[0] + parts[1],
		"garbage":           "not-a-cookie",
		"empty":             "",
	}
	for name, v := range cases {
		if VerifyVIPCookieValue(v, vipCookieTestSecret, vipCookieTestNow) != nil {
			t.Errorf("%s: tampered cookie accepted", name)
		}
	}
}

func TestVIPCookieWrongVersionRejected(t *testing.T) {
	raw, err := json.Marshal(map[string]any{"v": VIPCookieVersion + 1, "iat": vipCookieTestNow.Unix(), "exp": vipCookieTestNow.Add(time.Hour).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	enc := base64.RawURLEncoding.EncodeToString(raw)
	value := enc + "." + signPayload(enc, vipCookieTestSecret)
	if VerifyVIPCookieValue(value, vipCookieTestSecret, vipCookieTestNow) != nil {
		t.Error("future-version cookie accepted")
	}
}

func TestVIPCookieFutureIssuedRejected(t *testing.T) {
	// Mint at now+1h, verify at now: beyond skew tolerance.
	value := CreateVIPCookieValue(vipCookieTestSecret, vipCookieTestNow.Add(time.Hour))
	if VerifyVIPCookieValue(value, vipCookieTestSecret, vipCookieTestNow) != nil {
		t.Error("future-issued cookie accepted")
	}
	// Within skew tolerance it remains acceptable.
	value = CreateVIPCookieValue(vipCookieTestSecret, vipCookieTestNow.Add(time.Minute))
	if VerifyVIPCookieValue(value, vipCookieTestSecret, vipCookieTestNow) == nil {
		t.Error("cookie issued within clock skew rejected")
	}
}

func TestVIPCookieEmptySecretRejectsEverything(t *testing.T) {
	value := CreateVIPCookieValue(vipCookieTestSecret, vipCookieTestNow)
	if VerifyVIPCookieValue(value, "", vipCookieTestNow) != nil {
		t.Error("cookie verified with empty secret")
	}
}

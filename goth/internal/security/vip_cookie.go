package security

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

// VIPCookieName is the VIP portal session cookie. It is deliberately separate
// from karot_unlock so the portal can be rotated or revoked independently of
// the public dashboard unlock state (plan §6.2).
const VIPCookieName = "karot_vip"

// VIPCookieVersion is the signed-payload version. Verification accepts only
// the current version, so bumping it invalidates every outstanding cookie
// without touching VIP_COOKIE_SECRET.
const VIPCookieVersion = 1

// VIPCookieLifetime bounds a VIP session. Expiry is verified server-side on
// every request, not only by the browser (plan §6.2).
const VIPCookieLifetime = 24 * time.Hour

// vipMaxClockSkew tolerates modest clock drift between minter and verifier
// when rejecting future-issued payloads.
const vipMaxClockSkew = 2 * time.Minute

// VIPCookiePayload is the signed, versioned VIP session payload.
type VIPCookiePayload struct {
	Version   int   `json:"v"`
	IssuedAt  int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`
}

// marshalVIPPayload renders the payload with no HTML escaping and no trailing
// newline, matching the unlock-cookie byte form so signatures are stable.
func marshalVIPPayload(p VIPCookiePayload) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(p); err != nil {
		return nil
	}
	return bytes.TrimRight(buf.Bytes(), "\n")
}

// CreateVIPCookieValue builds the `payload.signature` cookie value for a fresh
// VIP session starting at now. The signature is HMAC-SHA-256 over the base64url
// payload using VIP_COOKIE_SECRET (never the plaintext access code).
func CreateVIPCookieValue(secret string, now time.Time) string {
	encoded := marshalVIPPayload(VIPCookiePayload{
		Version:   VIPCookieVersion,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(VIPCookieLifetime).Unix(),
	})
	if encoded == nil {
		return ""
	}
	enc := toBase64URL(encoded)
	return enc + "." + signPayload(enc, secret)
}

// VerifyVIPCookieValue validates a VIP cookie value against secret at time now.
// It returns nil for any malformed, tampered, expired, future-issued,
// wrong-version, or wrong-secret value. Callers must treat nil as unauthorized;
// hiding controls in HTML is never authorization (plan §6.2).
func VerifyVIPCookieValue(value, secret string, now time.Time) *VIPCookiePayload {
	if secret == "" {
		return nil
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil
	}
	expected := signPayload(parts[0], secret)
	a := []byte(parts[1])
	b := []byte(expected)
	if len(a) != len(b) || subtle.ConstantTimeCompare(a, b) != 1 {
		return nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil
	}
	var decoded struct {
		Version   *int   `json:"v"`
		IssuedAt  *int64 `json:"iat"`
		ExpiresAt *int64 `json:"exp"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil
	}
	if decoded.Version == nil || decoded.IssuedAt == nil || decoded.ExpiresAt == nil {
		return nil
	}
	if *decoded.Version != VIPCookieVersion {
		return nil
	}
	// Server-side expiry: the session is dead at and after ExpiresAt.
	if now.Unix() >= *decoded.ExpiresAt {
		return nil
	}
	// Reject future-issued payloads beyond skew (forged or replayed values).
	if *decoded.IssuedAt > now.Add(vipMaxClockSkew).Unix() {
		return nil
	}
	if *decoded.IssuedAt >= *decoded.ExpiresAt {
		return nil
	}
	return &VIPCookiePayload{
		Version:   *decoded.Version,
		IssuedAt:  *decoded.IssuedAt,
		ExpiresAt: *decoded.ExpiresAt,
	}
}

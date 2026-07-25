package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// UnlockPayload is the signed cookie payload.
type UnlockPayload struct {
	SessionID  string `json:"sessionId"`
	Locale     string `json:"locale"`
	UnlockedAt int64  `json:"unlockedAt"`
}

func toBase64URL(input []byte) string {
	return base64.RawURLEncoding.EncodeToString(input)
}

func signPayload(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return toBase64URL(mac.Sum(nil))
}

// marshalPayload renders the payload in the exact byte form of the reference's
// JSON.stringify (src/lib/security/unlock-cookie.ts): no HTML escaping of
// '<', '>' or '&', keys in struct order, no trailing newline. Byte parity
// keeps a Go-minted cookie identical to a Node-minted one for the same input.
func marshalPayload(payload UnlockPayload) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(payload); err != nil {
		return nil
	}
	return bytes.TrimRight(buf.Bytes(), "\n")
}

// CreateUnlockCookieValue builds the `payload.signature` cookie value.
func CreateUnlockCookieValue(payload UnlockPayload, secret string) string {
	encoded := marshalPayload(payload)
	if encoded == nil {
		return ""
	}
	enc := toBase64URL(encoded)
	return enc + "." + signPayload(enc, secret)
}

// VerifyUnlockCookieValue validates and decodes a cookie value.
//
// Validation semantics mirror verifyUnlockCookieValue exactly (Phase 12.5f,
// resolving the former mismatch explicitly): the HMAC comparison is the
// security gate; afterwards each field is checked for PRESENCE and JSON TYPE
// only — an empty sessionId/locale string and an unlockedAt of 0 are accepted
// precisely because the reference's `typeof` checks accept them. A missing
// key, null, or wrong JSON type is rejected, like the reference.
func VerifyUnlockCookieValue(value, secret string) *UnlockPayload {
	// Split on "." and use only the first two segments, mirroring the
	// reference's destructuring of value.split(".") — trailing segments are
	// ignored by both stacks (golden vector "extra-segments-ignored").
	parts := strings.Split(value, ".")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
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
	// Pointer fields distinguish "absent/null" from "present zero value",
	// matching the reference's typeof guards.
	var decoded struct {
		SessionID  *string  `json:"sessionId"`
		Locale     *string  `json:"locale"`
		UnlockedAt *float64 `json:"unlockedAt"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil
	}
	if decoded.SessionID == nil || decoded.Locale == nil || decoded.UnlockedAt == nil {
		return nil
	}
	return &UnlockPayload{
		SessionID:  *decoded.SessionID,
		Locale:     *decoded.Locale,
		UnlockedAt: int64(*decoded.UnlockedAt),
	}
}

// Rate limiter (in-memory), ported from src/lib/security/rate-limit.ts and
// generalized behind a pluggable store so every abuse-prone endpoint (contact,
// sentinel, AI Pulse) shares one bounded, distributed-ready limiter.

// RateLimitResult is the outcome of a single rate-limit check.
type RateLimitResult struct {
	Allowed      bool
	Remaining    int
	RetryAfterMs int64
}

// RateLimitStore is the pluggable backend behind EnforceRateLimit. The default
// is a bounded in-memory store suitable for a single instance; a distributed
// backend (e.g. Redis) can implement this interface and be installed with
// SetRateLimitStore for multi-instance deployments behind a load balancer.
type RateLimitStore interface {
	// Take records one hit against bucketKey and reports whether it is within
	// limit. windowMs is the bucket lifetime; now is the current unix-millis.
	Take(bucketKey string, limit int, windowMs, now int64) RateLimitResult
}

type rateEntry struct {
	count   int
	resetAt int64
}

// memoryStore is a bounded in-memory RateLimitStore. It evicts expired buckets
// and, when still over capacity, the soonest-expiring buckets, so memory stays
// bounded even under a key-space flood (e.g. spoofed X-Forwarded-For values).
type memoryStore struct {
	mu         sync.Mutex
	store      map[string]*rateEntry
	maxBuckets int
}

func newMemoryStore(maxBuckets int) *memoryStore {
	return &memoryStore{store: map[string]*rateEntry{}, maxBuckets: maxBuckets}
}

// Take implements RateLimitStore.
func (l *memoryStore) Take(bucketKey string, limit int, windowMs, now int64) RateLimitResult {
	l.mu.Lock()
	defer l.mu.Unlock()

	if e, ok := l.store[bucketKey]; !ok || e.resetAt <= now {
		l.store[bucketKey] = &rateEntry{count: 1, resetAt: now + windowMs}
		l.evictLocked(now)
		return RateLimitResult{Allowed: true, Remaining: max(limit-1, 0), RetryAfterMs: windowMs}
	}

	e := l.store[bucketKey]
	e.count++
	if e.count > limit {
		return RateLimitResult{Allowed: false, Remaining: 0, RetryAfterMs: retryMs(e.resetAt - now)}
	}
	return RateLimitResult{Allowed: true, Remaining: max(limit-e.count, 0), RetryAfterMs: retryMs(e.resetAt - now)}
}

var (
	storeMu      sync.RWMutex
	defaultStore RateLimitStore = newMemoryStore(4096)
)

// SetRateLimitStore installs a custom (e.g. distributed) backend. Passing nil
// restores the default bounded in-memory store. Safe for concurrent use.
func SetRateLimitStore(s RateLimitStore) {
	storeMu.Lock()
	defer storeMu.Unlock()
	if s == nil {
		defaultStore = newMemoryStore(4096)
		return
	}
	defaultStore = s
}

func currentStore() RateLimitStore {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return defaultStore
}

// EnforceRateLimit records a hit for scope:key against the active store and
// returns whether the bucket is allowed and retry info.
func EnforceRateLimit(scope, key string, limit int, window time.Duration) (allowed bool, remaining int, retryAfterMs int64) {
	res := currentStore().Take(scope+":"+key, limit, window.Milliseconds(), time.Now().UnixMilli())
	return res.Allowed, res.Remaining, res.RetryAfterMs
}

// evictLocked drops expired buckets and, if still over capacity, the
// oldest-by-reset entries. Caller must hold l.mu.
func (l *memoryStore) evictLocked(now int64) {
	for k, e := range l.store {
		if e.resetAt <= now {
			delete(l.store, k)
		}
	}
	if len(l.store) <= l.maxBuckets {
		return
	}
	// Over capacity: remove the entries expiring soonest first.
	type kv struct {
		key     string
		resetAt int64
	}
	items := make([]kv, 0, len(l.store))
	for k, e := range l.store {
		items = append(items, kv{k, e.resetAt})
	}
	// Simple selection sort of the few oldest resetAt values.
	for i := 0; i < len(items) && len(l.store) > l.maxBuckets; i++ {
		minIdx := i
		for j := i + 1; j < len(items); j++ {
			if items[j].resetAt < items[minIdx].resetAt {
				minIdx = j
			}
		}
		items[i], items[minIdx] = items[minIdx], items[i]
		delete(l.store, items[i].key)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func retryMs(remaining int64) int64 {
	if remaining < 1000 {
		return 1000
	}
	return remaining
}

// GetClientIP extracts the client IP from proxy headers.
func GetClientIP(headers map[string]string) string {
	if v, ok := headers["x-forwarded-for"]; ok && v != "" {
		parts := strings.Split(v, ",")
		return strings.TrimSpace(parts[0])
	}
	if v, ok := headers["x-real-ip"]; ok && v != "" {
		return v
	}
	return "unknown"
}

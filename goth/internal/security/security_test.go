package security

import (
	"sync"
	"testing"
	"time"
)

func TestEnforceRateLimitAllowsThenBlocks(t *testing.T) {
	SetRateLimitStore(nil) // fresh default store
	t.Cleanup(func() { SetRateLimitStore(nil) })

	const limit = 3
	for i := 1; i <= limit; i++ {
		allowed, remaining, retry := EnforceRateLimit("test-scope", "1.2.3.4", limit, time.Minute)
		if !allowed {
			t.Fatalf("request %d: allowed=false, want true", i)
		}
		if want := limit - i; remaining != want {
			t.Errorf("request %d: remaining=%d, want %d", i, remaining, want)
		}
		if retry <= 0 {
			t.Errorf("request %d: retryAfterMs=%d, want >0", i, retry)
		}
	}

	allowed, remaining, retry := EnforceRateLimit("test-scope", "1.2.3.4", limit, time.Minute)
	if allowed {
		t.Fatal("over-limit request: allowed=true, want false")
	}
	if remaining != 0 {
		t.Errorf("over-limit remaining=%d, want 0", remaining)
	}
	if retry < 1000 {
		t.Errorf("over-limit retryAfterMs=%d, want >=1000", retry)
	}

	// A different key has its own independent bucket.
	if allowed, _, _ := EnforceRateLimit("test-scope", "9.9.9.9", limit, time.Minute); !allowed {
		t.Error("distinct key should have its own bucket")
	}
	// A different scope with the same key is also independent.
	if allowed, _, _ := EnforceRateLimit("other-scope", "1.2.3.4", limit, time.Minute); !allowed {
		t.Error("distinct scope should have its own bucket")
	}
}

func TestExpiredWindowResets(t *testing.T) {
	SetRateLimitStore(nil)
	t.Cleanup(func() { SetRateLimitStore(nil) })

	// 1 ms window: the second call lands in a fresh window.
	if allowed, _, _ := EnforceRateLimit("exp", "k", 1, time.Millisecond); !allowed {
		t.Fatal("first request should be allowed")
	}
	time.Sleep(3 * time.Millisecond)
	if allowed, _, _ := EnforceRateLimit("exp", "k", 1, time.Millisecond); !allowed {
		t.Fatal("request after window reset should be allowed")
	}
}

// countingStore is a trivial distributed-ready backend used to prove the
// limiter routes through the installed RateLimitStore.
type countingStore struct {
	mu    sync.Mutex
	calls int
	last  string
}

func (c *countingStore) Take(bucketKey string, limit int, windowMs, now int64) RateLimitResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	c.last = bucketKey
	return RateLimitResult{Allowed: false, Remaining: 0, RetryAfterMs: 4200}
}

func TestSetRateLimitStorePluggable(t *testing.T) {
	cs := &countingStore{}
	SetRateLimitStore(cs)
	t.Cleanup(func() { SetRateLimitStore(nil) })

	allowed, _, retry := EnforceRateLimit("svc", "abc", 10, time.Minute)
	if allowed {
		t.Error("custom store returned Allowed=false; EnforceRateLimit should honor it")
	}
	if retry != 4200 {
		t.Errorf("retryAfterMs=%d, want 4200 from custom store", retry)
	}
	if cs.calls != 1 {
		t.Errorf("store calls=%d, want 1", cs.calls)
	}
	if cs.last != "svc:abc" {
		t.Errorf("bucketKey=%q, want svc:abc", cs.last)
	}

	// nil restores the default bounded in-memory store.
	SetRateLimitStore(nil)
	if allowed, _, _ := EnforceRateLimit("svc", "abc", 10, time.Minute); !allowed {
		t.Error("after reset the default store should allow a fresh key")
	}
}

func TestMemoryStoreBounded(t *testing.T) {
	s := newMemoryStore(8)
	now := time.Now().UnixMilli()
	// Insert far more distinct keys than capacity, all in a long window.
	for i := 0; i < 1000; i++ {
		s.Take(string(rune('a'+i%26))+"-"+time.Duration(i).String(), 5, 60_000, now)
	}
	s.mu.Lock()
	size := len(s.store)
	s.mu.Unlock()
	if size > 8 {
		t.Errorf("store size=%d, want <= maxBuckets 8 (bounded eviction failed)", size)
	}
}

func TestGetClientIP(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{"forwarded-for first hop", map[string]string{"x-forwarded-for": "203.0.113.9, 10.0.0.1"}, "203.0.113.9"},
		{"real-ip fallback", map[string]string{"x-real-ip": "198.51.100.4"}, "198.51.100.4"},
		{"unknown", map[string]string{}, "unknown"},
	}
	for _, tc := range cases {
		if got := GetClientIP(tc.headers); got != tc.want {
			t.Errorf("%s: GetClientIP=%q, want %q", tc.name, got, tc.want)
		}
	}
}

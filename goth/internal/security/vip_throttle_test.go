package security

import (
	"testing"
	"time"
)

func TestVIPLoginThrottleBudget(t *testing.T) {
	th := NewVIPLoginThrottle()
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	ip := "203.0.113.7"

	for i := 1; i <= vipLoginAttemptLimit; i++ {
		if ok, retry := th.Allow(ip, now); !ok {
			t.Fatalf("attempt %d denied early (retry %v)", i, retry)
		}
	}
	ok, retry := th.Allow(ip, now)
	if ok {
		t.Fatal("6th attempt allowed")
	}
	if retry != vipLoginBaseWindow {
		t.Errorf("first lockout retry = %v, want %v", retry, vipLoginBaseWindow)
	}

	// Denied throughout the cooldown, even with fresh timestamps.
	mid := now.Add(vipLoginBaseWindow / 2)
	if ok, _ = th.Allow(ip, mid); ok {
		t.Error("attempt allowed during cooldown")
	}
	// After the cooldown lapses the budget resets.
	after := now.Add(vipLoginBaseWindow + time.Second)
	if ok, _ = th.Allow(ip, after); !ok {
		t.Error("attempt denied after cooldown lapsed")
	}
}

func TestVIPLoginThrottleEscalates(t *testing.T) {
	th := NewVIPLoginThrottle()
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	ip := "203.0.113.8"

	exhaust := func(at time.Time) time.Duration {
		t.Helper()
		var retry time.Duration
		for i := 0; i <= vipLoginAttemptLimit; i++ {
			_, retry = th.Allow(ip, at)
		}
		return retry
	}

	first := exhaust(now)
	if first != vipLoginBaseWindow {
		t.Fatalf("first lockout = %v, want %v", first, vipLoginBaseWindow)
	}

	second := exhaust(now.Add(first + time.Second))
	if second != 2*vipLoginBaseWindow {
		t.Fatalf("second lockout = %v, want %v", second, 2*vipLoginBaseWindow)
	}

	third := exhaust(now.Add(first + second + time.Minute))
	if third != 4*vipLoginBaseWindow {
		t.Fatalf("third lockout = %v, want %v", third, 4*vipLoginBaseWindow)
	}
}

func TestVIPLoginThrottleCooldownCapped(t *testing.T) {
	th := NewVIPLoginThrottle()
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	ip := "203.0.113.9"

	cursor := now
	var retry time.Duration
	for i := 0; i < 12; i++ {
		for j := 0; j <= vipLoginAttemptLimit; j++ {
			_, retry = th.Allow(ip, cursor)
		}
		cursor = cursor.Add(retry + time.Second)
	}
	if retry != vipLoginMaxCooldown {
		t.Errorf("escalated cooldown = %v, want cap %v", retry, vipLoginMaxCooldown)
	}
}

func TestVIPLoginThrottleLockoutsDecay(t *testing.T) {
	th := NewVIPLoginThrottle()
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	ip := "203.0.113.10"

	var retry time.Duration
	for i := 0; i <= vipLoginAttemptLimit; i++ {
		_, retry = th.Allow(ip, now)
	}
	if retry != vipLoginBaseWindow {
		t.Fatalf("first lockout = %v, want %v", retry, vipLoginBaseWindow)
	}

	// More than the decay period later, the escalation state is forgotten and
	// the next lockout is back to the base window.
	later := now.Add(vipLoginLockoutDecay + time.Hour)
	for i := 0; i <= vipLoginAttemptLimit; i++ {
		_, retry = th.Allow(ip, later)
	}
	if retry != vipLoginBaseWindow {
		t.Errorf("post-decay lockout = %v, want base %v", retry, vipLoginBaseWindow)
	}
}

func TestVIPLoginThrottleIsolatedPerIP(t *testing.T) {
	th := NewVIPLoginThrottle()
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)

	for i := 0; i <= vipLoginAttemptLimit; i++ {
		th.Allow("203.0.113.11", now)
	}
	if ok, _ := th.Allow("203.0.113.12", now); !ok {
		t.Error("unrelated IP denied because another IP exhausted its budget")
	}
}

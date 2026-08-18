package security

import (
	"sync"
	"time"
)

type VIPChatLimiter struct{}

func NewVIPChatLimiter() *VIPChatLimiter { return &VIPChatLimiter{} }

func (l *VIPChatLimiter) Allow(key string, now time.Time) (bool, int64) {
	ok, _, retry := EnforceRateLimit("vip-chat", key, 20, time.Hour)
	return ok, retry
}

// VIP login throttling (threat T2): a per-IP attempt budget with an
// escalating temporary cooldown after each exhausted budget. The dedicated
// throttle (rather than the generic fixed-window limiter) is what makes the
// cooldown grow: 10 min after the first lockout, 20 min after the second,
// 40 min after the third, and so on up to a 6 h cap.

const (
	vipLoginAttemptLimit   = 5
	vipLoginBaseWindow     = 10 * time.Minute
	vipLoginMaxCooldown    = 6 * time.Hour
	vipLoginLockoutDecay   = 24 * time.Hour
	vipLoginMaxThrottleIPs = 4096
)

type vipLoginState struct {
	attempts      int
	windowEnd     int64
	cooldownUntil int64
	lockouts      int
	lastLockoutAt int64
	lastSeenAt    int64
}

// VIPLoginThrottle limits VIP access-code attempts per client IP. It is safe
// for concurrent use and bounded in memory.
type VIPLoginThrottle struct {
	mu          sync.Mutex
	states      map[string]*vipLoginState
	limit       int
	baseWindow  time.Duration
	maxCooldown time.Duration
}

// NewVIPLoginThrottle builds a throttle with the plan's defaults: 5 attempts
// per 10-minute window, escalating cooldowns capped at 6 hours.
func NewVIPLoginThrottle() *VIPLoginThrottle {
	return &VIPLoginThrottle{
		states:      map[string]*vipLoginState{},
		limit:       vipLoginAttemptLimit,
		baseWindow:  vipLoginBaseWindow,
		maxCooldown: vipLoginMaxCooldown,
	}
}

// Allow records one login attempt for ip at time now. It reports whether the
// attempt may proceed and, when denied, how long the caller should wait.
func (t *VIPLoginThrottle) Allow(ip string, now time.Time) (allowed bool, retryAfter time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	nowSec := now.Unix()
	t.evictLocked(nowSec)

	s, ok := t.states[ip]
	if !ok {
		s = &vipLoginState{}
		t.states[ip] = s
	}
	s.lastSeenAt = nowSec

	if nowSec < s.cooldownUntil {
		return false, time.Duration(s.cooldownUntil-nowSec) * time.Second
	}

	if nowSec >= s.windowEnd {
		s.attempts = 0
		s.windowEnd = nowSec + int64(t.baseWindow.Seconds())
		if s.lastLockoutAt > 0 && nowSec-s.lastLockoutAt > int64(vipLoginLockoutDecay.Seconds()) {
			s.lockouts = 0
		}
	}

	s.attempts++
	if s.attempts <= t.limit {
		return true, 0
	}

	// Budget exhausted: escalate. Cooldown doubles with each lockout and is
	// capped; the shift is bounded so repeated lockouts cannot overflow.
	s.lockouts++
	s.lastLockoutAt = nowSec
	shift := s.lockouts - 1
	if shift > 30 {
		shift = 30
	}
	cooldown := t.baseWindow << uint(shift)
	if cooldown <= 0 || cooldown > t.maxCooldown {
		cooldown = t.maxCooldown
	}
	s.cooldownUntil = nowSec + int64(cooldown.Seconds())
	s.attempts = 0
	return false, cooldown
}

// evictLocked drops entries that have been idle longer than the lockout-decay
// period (their escalation history is expired anyway) and, if still over
// capacity, arbitrary further entries so memory stays bounded under a
// spoofed-IP flood. Entries that merely lapsed their current window or
// cooldown are kept so returning clients still face their escalated cooldowns.
// Caller must hold t.mu.
func (t *VIPLoginThrottle) evictLocked(nowSec int64) {
	decaySec := int64(vipLoginLockoutDecay.Seconds())
	for ip, s := range t.states {
		if nowSec-s.lastSeenAt > decaySec {
			delete(t.states, ip)
		}
	}
	for ip := range t.states {
		if len(t.states) <= vipLoginMaxThrottleIPs {
			break
		}
		delete(t.states, ip)
	}
}

package security

import (
	"sync"
	"time"
)

type VIPResourceGate struct {
	sem    chan struct{}
	mu     sync.Mutex
	starts int
	reset  time.Time
	limit  int
	window time.Duration
}

func NewVIPResourceGate(concurrency, limit int, window time.Duration) *VIPResourceGate {
	return &VIPResourceGate{sem: make(chan struct{}, concurrency), limit: limit, window: window}
}

func (g *VIPResourceGate) Acquire(now time.Time) bool {
	select {
	case g.sem <- struct{}{}:
	default:
		return false
	}
	g.mu.Lock()
	if g.reset.IsZero() || !now.Before(g.reset) {
		g.starts, g.reset = 0, now.Add(g.window)
	}
	if g.starts >= g.limit {
		g.mu.Unlock()
		<-g.sem
		return false
	}
	g.starts++
	g.mu.Unlock()
	return true
}

func (g *VIPResourceGate) Release() { <-g.sem }

var vipLoginGate = NewVIPResourceGate(2, 30, 10*time.Minute)
var vipChatGate = NewVIPResourceGate(3, 100, time.Hour)

func AcquireVIPLogin(now time.Time) bool { return vipLoginGate.Acquire(now) }
func ReleaseVIPLogin()                   { vipLoginGate.Release() }
func AcquireVIPChat(now time.Time) bool  { return vipChatGate.Acquire(now) }
func ReleaseVIPChat()                    { vipChatGate.Release() }

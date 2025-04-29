package mux

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type selectiveRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

type visitor struct {
	tokens     int
	lastAccess time.Time
}

func NewSelectiveRateLimiter() *selectiveRateLimiter {
	rl := &selectiveRateLimiter{
		visitors: make(map[string]*visitor),
	}
	go rl.cleanupExpiredVisitors(10 * time.Minute)
	return rl
}

func (m *selectiveRateLimiter) Invoke(c *RouteContext, next HandlerFunc) {
	opts := c.Options
	if opts == nil || opts.RateLimit <= 0 {
		next(c)
		return
	}

	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	v := m.getVisitor(ip, c.Request.Pattern, opts.RateLimit, opts.RateInterval)

	if v.tokens <= 0 {
		c.Problem(&ProblemDetails{
			Title:    "Rate limit exceeded",
			Detail:   "You have exceeded the allowed number of requests. Please try again later.",
			Status:   http.StatusTooManyRequests,
			Type:     "https://httpstatuses.com/429",
			Instance: getInstanceURI(c.Request),
		})
		return
	}

	v.tokens--
	next(c)
}

func (m *selectiveRateLimiter) getVisitor(ip, key string, limit int, interval time.Duration) *visitor {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := ip + ":" + key
	v, ok := m.visitors[id]
	now := time.Now()

	if !ok {
		v = &visitor{tokens: limit - 1, lastAccess: now}
		m.visitors[id] = v
		return v
	}

	elapsed := now.Sub(v.lastAccess)
	v.lastAccess = now

	refill := int(elapsed / interval)
	if refill > 0 {
		v.tokens += refill
		if v.tokens > limit {
			v.tokens = limit
		}
	}

	return v
}

func (m *selectiveRateLimiter) cleanupExpiredVisitors(maxIdle time.Duration) {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for key, v := range m.visitors {
			if now.Sub(v.lastAccess) > maxIdle {
				delete(m.visitors, key)
			}
		}
		m.mu.Unlock()
	}
}

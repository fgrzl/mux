package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"

	routerpkg "github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
)

type SelectiveRateLimiter struct {
	mu            sync.Mutex
	visitors      map[string]*visitor
	cleanupTicker time.Duration
}

type visitor struct {
	tokens     int
	lastAccess time.Time
}

// ---- Options Pattern ----

type RateLimiterOptions struct {
	CleanupInterval time.Duration
}

type RateLimiterOption func(*RateLimiterOptions)

func WithCleanupInterval(d time.Duration) RateLimiterOption {
	return func(o *RateLimiterOptions) {
		o.CleanupInterval = d
	}
}

// UseRateLimiter adds rate limiting middleware to the router with the given options.
func UseRateLimiter(r *routerpkg.Router, opts ...RateLimiterOption) {
	limiter := NewSelectiveRateLimiter(opts...)
	r.Middleware = append(r.Middleware, limiter)
}

func NewSelectiveRateLimiter(opts ...RateLimiterOption) *SelectiveRateLimiter {
	config := &RateLimiterOptions{
		CleanupInterval: 10 * time.Minute,
	}
	for _, opt := range opts {
		opt(config)
	}

	rl := &SelectiveRateLimiter{
		visitors:      make(map[string]*visitor),
		cleanupTicker: config.CleanupInterval,
	}
	go rl.cleanupExpiredVisitors()
	return rl
}

// ---- Middleware ----

func (m *SelectiveRateLimiter) Invoke(c routing.RouteContext, next routerpkg.HandlerFunc) {

	opts := c.Options()
	if opts == nil || opts.RateLimit <= 0 {
		next(c)
		return
	}

	ip, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		// If we can't parse the host:port, use the entire RemoteAddr as IP
		ip = c.Request().RemoteAddr
	}
	v := m.getVisitor(ip, c.Request().Pattern, opts.RateLimit, opts.RateInterval)

	if v.tokens <= 0 {
		instance := c.Request().RequestURI
		c.Problem(&routing.ProblemDetails{
			Title:    "Rate limit exceeded",
			Detail:   "You have exceeded the allowed number of requests. Please try again later.",
			Status:   http.StatusTooManyRequests,
			Type:     "https://httpstatuses.com/429",
			Instance: &instance,
		})
		return
	}

	v.tokens--
	next(c)
}

func (m *SelectiveRateLimiter) getVisitor(ip, key string, limit int, interval time.Duration) *visitor {
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

func (m *SelectiveRateLimiter) cleanupExpiredVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for key, v := range m.visitors {
			if now.Sub(v.lastAccess) > m.cleanupTicker {
				delete(m.visitors, key)
			}
		}
		m.mu.Unlock()
	}
}

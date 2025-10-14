package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// SelectiveRateLimiter implements a simple token-bucket rate limiter keyed by client IP and route pattern.
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

// RateLimiterOptions configures the rate limiter behavior.
type RateLimiterOptions struct {
	CleanupInterval time.Duration
}

// RateLimiterOption applies a configuration to RateLimiterOptions.
type RateLimiterOption func(*RateLimiterOptions)

// WithCleanupInterval configures how often to purge inactive visitors.
func WithCleanupInterval(d time.Duration) RateLimiterOption {
	return func(o *RateLimiterOptions) {
		o.CleanupInterval = d
	}
}

// UseRateLimiter adds rate limiting middleware to the router with the given options.
func UseRateLimiter(r *router.Router, opts ...RateLimiterOption) {
	limiter := NewSelectiveRateLimiter(opts...)
	r.Use(limiter)
}

// NewSelectiveRateLimiter constructs a SelectiveRateLimiter with optional configuration.
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

const rateLimitTitle = "Rate limit exceeded"
const rateLimitDetail = "You have exceeded the allowed number of requests. Please try again later."

func (m *SelectiveRateLimiter) Invoke(c routing.RouteContext, next router.HandlerFunc) {

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
	// Use the configured route pattern (from options) as the visitor key part.
	// Reading pattern from the RouteOptions ensures we use the canonical route
	// identifier registered at startup rather than any request-derived value.
	v := m.getVisitor(ip, opts.Pattern, opts.RateLimit, opts.RateInterval)

	if v.tokens <= 0 {
		instance := c.Request().RequestURI
		c.Problem(&routing.ProblemDetails{
			Title:    rateLimitTitle,
			Detail:   rateLimitDetail,
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
		// First access consumes one token; initialize to limit-1 so the
		// caller's subsequent decrement models immediate consumption.
		v = &visitor{tokens: limit - 1, lastAccess: now}
		m.visitors[id] = v
		return v
	}

	elapsed := now.Sub(v.lastAccess)
	v.lastAccess = now

	// Guard against zero interval which would cause divide-by-zero panic.
	refill := 0
	if interval > 0 {
		refill = int(elapsed / interval)
	}
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

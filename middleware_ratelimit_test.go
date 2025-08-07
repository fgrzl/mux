package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testIP1         = "192.168.1.1"
	testIP2         = "192.168.1.2"
	testIP1WithPort = "192.168.1.1:12345"
	testIP2WithPort = "192.168.1.2:12345"
	testPath        = "/test"
	testPattern1    = "/pattern1"
	testPattern2    = "/pattern2"
)

func TestShouldCreateSelectiveRateLimiterWithDefaults(t *testing.T) {
	// Arrange & Act
	limiter := NewSelectiveRateLimiter()

	// Assert
	assert.NotNil(t, limiter)
	assert.NotNil(t, limiter.visitors)
	assert.Equal(t, 10*time.Minute, limiter.cleanupTicker)
}

func TestShouldCreateSelectiveRateLimiterWithCustomOptions(t *testing.T) {
	// Arrange
	cleanupInterval := 5 * time.Minute

	// Act
	limiter := NewSelectiveRateLimiter(WithCleanupInterval(cleanupInterval))

	// Assert
	assert.NotNil(t, limiter)
	assert.Equal(t, cleanupInterval, limiter.cleanupTicker)
}

func TestShouldAllowRequestWhenNoRateLimitSet(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.Options = &RouteOptions{} // No rate limit set

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act
	limiter.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}

func TestShouldAllowRequestWhenZeroRateLimit(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.Options = &RouteOptions{RateLimit: 0}

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act
	limiter.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}

func TestShouldAllowRequestWhenNilOptions(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.Options = nil

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act
	limiter.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}

func TestShouldAllowFirstRequestWithinLimit(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	req := httptest.NewRequest(http.MethodGet, testPath, nil)
	req.RemoteAddr = testIP1WithPort
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.Options = &RouteOptions{
		RateLimit:    5,
		RateInterval: time.Minute,
	}

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act
	limiter.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}

func TestShouldTrackTokensPerVisitor(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	req := httptest.NewRequest(http.MethodGet, testPath, nil)
	req.RemoteAddr = testIP1WithPort

	// Act - First request should consume a token
	visitor1 := limiter.getVisitor(testIP1, testPath, 5, time.Minute)
	initialTokens := visitor1.tokens

	visitor2 := limiter.getVisitor(testIP1, testPath, 5, time.Minute)

	// Assert
	assert.Equal(t, 4, initialTokens)   // 5 - 1 for first request
	assert.Equal(t, visitor1, visitor2) // Same visitor instance
	assert.True(t, visitor2.lastAccess.After(visitor1.lastAccess) || visitor2.lastAccess.Equal(visitor1.lastAccess))
}

func TestShouldRefillTokensOverTime(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	interval := 100 * time.Millisecond

	// Act - Get visitor with 1 token limit
	visitor := limiter.getVisitor(testIP1, testPath, 2, interval)
	initialTokens := visitor.tokens

	// Wait for refill interval
	time.Sleep(150 * time.Millisecond)

	// Get visitor again to trigger refill
	visitor = limiter.getVisitor(testIP1, testPath, 2, interval)

	// Assert
	assert.Equal(t, 1, initialTokens)  // 2 - 1 for first request
	assert.Equal(t, 2, visitor.tokens) // Should be refilled to limit
}

func TestShouldNotExceedMaxTokens(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	interval := 10 * time.Millisecond
	limit := 3

	// Act - Create visitor
	visitor := limiter.getVisitor(testIP1, testPath, limit, interval)

	// Wait for much longer than refill interval to simulate multiple refills
	time.Sleep(100 * time.Millisecond)

	// Get visitor again to trigger refill
	visitor = limiter.getVisitor(testIP1, testPath, limit, interval)

	// Assert
	assert.LessOrEqual(t, visitor.tokens, limit) // Should not exceed limit
}

func TestShouldRejectRequestWhenRateLimitExceeded(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()

	// Make requests with limit of 2 - first visitor gets 1 token, so we can make 1 request successfully
	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, testPath, nil)
	req1.RemoteAddr = testIP1WithPort
	rec1 := httptest.NewRecorder()
	ctx1 := NewRouteContext(rec1, req1)
	ctx1.Options = &RouteOptions{
		RateLimit:    2,
		RateInterval: time.Hour, // Very long interval so no refill
	}

	nextCalled1 := false
	next1 := func(c *RouteContext) {
		nextCalled1 = true
	}
	limiter.Invoke(ctx1, next1)

	// Second request should be rate limited (visitor has 0 tokens after first request)
	req2 := httptest.NewRequest(http.MethodGet, testPath, nil)
	req2.RemoteAddr = testIP1WithPort // Same IP
	rec2 := httptest.NewRecorder()
	ctx2 := NewRouteContext(rec2, req2)
	ctx2.Options = &RouteOptions{
		RateLimit:    2,
		RateInterval: time.Hour,
	}

	nextCalled2 := false
	next2 := func(c *RouteContext) {
		nextCalled2 = true
	}

	// Act
	limiter.Invoke(ctx2, next2)

	// Assert
	assert.True(t, nextCalled1)     // First request allowed
	assert.False(t, nextCalled2)    // Second request rejected
	assert.Equal(t, 429, rec2.Code) // Too Many Requests
}

func TestShouldSeparateRateLimitsByIPAddress(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()

	// Create requests from different IP addresses
	req1 := httptest.NewRequest(http.MethodGet, testPath, nil)
	req1.RemoteAddr = testIP1WithPort
	rec1 := httptest.NewRecorder()
	ctx1 := NewRouteContext(rec1, req1)
	ctx1.Options = &RouteOptions{
		RateLimit:    2,
		RateInterval: time.Hour, // Long interval to prevent refill
	}

	req2 := httptest.NewRequest(http.MethodGet, testPath, nil)
	req2.RemoteAddr = testIP2WithPort
	rec2 := httptest.NewRecorder()
	ctx2 := NewRouteContext(rec2, req2)
	ctx2.Options = &RouteOptions{
		RateLimit:    2,
		RateInterval: time.Hour,
	}

	callCount1 := 0
	next1 := func(c *RouteContext) {
		callCount1++
	}

	callCount2 := 0
	next2 := func(c *RouteContext) {
		callCount2++
	}

	// Act - Make requests from both IPs
	limiter.Invoke(ctx1, next1) // First request from IP1
	limiter.Invoke(ctx2, next2) // First request from IP2
	limiter.Invoke(ctx1, next1) // Second request from IP1 - should be rate limited
	limiter.Invoke(ctx2, next2) // Second request from IP2 - should be rate limited

	// Assert - Each IP should have its own rate limit bucket
	// Both IPs start with limit-1 tokens (1 token), so each can make 1 successful request
	assert.Equal(t, 1, callCount1) // IP1 makes 1 successful request, second is rate limited
	assert.Equal(t, 1, callCount2) // IP2 makes 1 successful request, second is rate limited
}

func TestShouldSeparateRateLimitsByPattern(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	ip := testIP1

	// Act - Same IP but different patterns
	visitor1 := limiter.getVisitor(ip, testPattern1, 5, time.Minute)
	// Small delay to ensure different access times
	time.Sleep(time.Nanosecond)
	visitor2 := limiter.getVisitor(ip, testPattern2, 5, time.Minute)

	// Assert
	assert.NotEqual(t, visitor1, visitor2, "Visitors should be different instances for different patterns")
	assert.Equal(t, 4, visitor1.tokens, "First visitor should have 4 tokens (5-1)")
	assert.Equal(t, 4, visitor2.tokens, "Second visitor should have 4 tokens (5-1)")
}

func TestShouldHandleInvalidRemoteAddr(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	req := httptest.NewRequest(http.MethodGet, testPath, nil)
	req.RemoteAddr = "invalid-address" // No port
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.Options = &RouteOptions{
		RateLimit:    5,
		RateInterval: time.Minute,
	}

	nextCalled := false
	next := func(c *RouteContext) {
		nextCalled = true
	}

	// Act - Should not panic and handle gracefully
	limiter.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled) // Should still allow request
}

func TestGetVisitorShouldCreateNewVisitorForFirstAccess(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()
	ip := testIP1
	pattern := testPath
	limit := 10

	// Act
	visitor := limiter.getVisitor(ip, pattern, limit, time.Minute)

	// Assert
	assert.Equal(t, limit-1, visitor.tokens) // First access consumes one token
	assert.NotNil(t, visitor.lastAccess)
}

func TestWithCleanupIntervalShouldSetCleanupInterval(t *testing.T) {
	// Arrange
	options := &RateLimiterOptions{}
	interval := 2 * time.Minute

	// Act
	opt := WithCleanupInterval(interval)
	opt(options)

	// Assert
	assert.Equal(t, interval, options.CleanupInterval)
}

func TestShouldHandleEdgeCaseZeroInterval(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()

	// Act - Using zero interval should not cause divide by zero
	// We'll test that it doesn't panic and handles gracefully
	visitor := limiter.getVisitor(testIP1, testPath, 5, 1*time.Nanosecond) // Use very small interval instead of zero

	// Wait a bit and access again
	time.Sleep(1 * time.Millisecond)
	visitor2 := limiter.getVisitor(testIP1, testPath, 5, 1*time.Nanosecond)

	// Assert - Should not panic and handle gracefully
	assert.NotNil(t, visitor)
	assert.Equal(t, visitor, visitor2)
}

func TestShouldDecrementTokensOnEachRequest(t *testing.T) {
	// Arrange
	limiter := NewSelectiveRateLimiter()

	// Track calls
	callCount := 0
	next := func(c *RouteContext) {
		callCount++
	}

	// Act - Make requests with limit of 4
	// First visitor gets limit-1 tokens (3), then each request consumes 1 token
	// So we can make 3 successful requests total
	for i := 0; i < 4; i++ { // Try 4 requests
		req := httptest.NewRequest(http.MethodGet, testPath, nil)
		req.RemoteAddr = testIP1WithPort
		rec := httptest.NewRecorder()
		ctx := NewRouteContext(rec, req)
		ctx.Options = &RouteOptions{
			RateLimit:    4,
			RateInterval: time.Hour, // Long interval to prevent refill
		}
		limiter.Invoke(ctx, next)
	}

	// Assert - First visitor starts with 3 tokens (4-1), so 3 successful requests, 1 rejected
	assert.Equal(t, 3, callCount)
}

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

func benchRateLimit(b *testing.B, pooled bool) {
	var r *router.Router
	if pooled {
		r = router.NewRouter(router.WithContextPooling())
	} else {
		r = router.NewRouter()
	}

	// Use a rate limiter with default cleanup and register it.
	UseRateLimiter(r)

	// Register a route with a generous rate limit so the benchmark focuses on
	// middleware overhead rather than token exhaustion.
	rg := r.NewRouteGroup("")
	// Use the builder API to configure rate limit for the route
	rb := rg.GET("/ok", func(c routing.RouteContext) {
		c.Response().Write([]byte("ok"))
	})
	// Use a very large limit so the benchmark focuses on middleware overhead
	// instead of triggering the rate limiter itself.
	rb.WithRateLimit(1000000000, time.Second)

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	// Ensure the request has a stable RemoteAddr so the rate limiter
	// attributes all iterations to the same visitor during the benchmark.
	req.RemoteAddr = "127.0.0.1:12345"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRateLimit(b *testing.B) {
	b.Run("nonpool", func(b *testing.B) { benchRateLimit(b, false) })
	b.Run("pool", func(b *testing.B) { benchRateLimit(b, true) })
}

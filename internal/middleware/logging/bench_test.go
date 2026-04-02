package logging

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/bench"
	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
)

// noopHandler is a slog.Handler that drops all logs to avoid I/O in benchmarks.
type noopHandler struct{}

func (noopHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (noopHandler) Handle(context.Context, slog.Record) error { return nil }
func (noopHandler) WithAttrs([]slog.Attr) slog.Handler        { return noopHandler{} }
func (noopHandler) WithGroup(string) slog.Handler             { return noopHandler{} }

func benchLogging(b *testing.B, pooled bool) {
	// Install a no-op default logger to remove backend cost from the benchmark
	slog.SetDefault(slog.New(noopHandler{}))

	var r *router.Router
	if pooled {
		r = router.NewRouter(router.WithContextPooling())
	} else {
		r = router.NewRouter()
	}
	rg := r.NewRouteGroup("")
	rg.GET("/ok", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte("ok")) })
	UseLogging(r)

	_, req := bench.NewRecorderRequest(http.MethodGet, "/ok")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkLogging(b *testing.B) {
	b.Run("nonpool", func(b *testing.B) { benchLogging(b, false) })
	b.Run("pool", func(b *testing.B) { benchLogging(b, true) })
}

// Additional benchmark variant that exercises multiple concurrent paths
func BenchmarkLoggingVariedPaths(b *testing.B) {
	slog.SetDefault(slog.New(noopHandler{}))

	for _, pooled := range []bool{false, true} {
		name := "nonpool"
		if pooled {
			name = "pool"
		}
		b.Run(name, func(b *testing.B) {
			var r *router.Router
			if pooled {
				r = router.NewRouter(router.WithContextPooling())
			} else {
				r = router.NewRouter()
			}
			UseLogging(r)
			rg := r.NewRouteGroup("")
			// A few routes to ensure params/static/wildcard do not change logging overhead.
			rg.GET("/ok", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte("ok")) })
			rg.GET("/user/:id", func(c routing.RouteContext) {
				if id, ok := c.Param("id"); ok {
					_, _ = c.Response().Write([]byte(id))
				} else {
					_, _ = c.Response().Write([]byte(""))
				}
			})
			rg.GET("/files/*path", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte("f")) })

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rr := httptest.NewRecorder()
				// rotate among the three
				switch i % 3 {
				case 0:
					r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ok", nil))
				case 1:
					r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/user/123", nil))
				default:
					r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/files/a/b/c", nil))
				}
			}
		})
	}
}

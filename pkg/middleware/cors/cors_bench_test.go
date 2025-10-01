package cors

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

func init() {
	// Silence structured logs during benchmarks to avoid polluting output and costing time.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
}

// BenchmarkCORS_Invoke measures Invoke overhead for simple and preflight requests.
func BenchmarkCORSInvoke(b *testing.B) {
	cases := []struct {
		name  string
		opts  Options
		setup func(r *http.Request)
	}{
		{
			name: "Simple_Wildcard",
			opts: Options{AllowedOrigins: []string{"*"}},
			setup: func(r *http.Request) {
				r.Header.Set("Origin", "https://example.com")
			},
		},
		{
			name: "Simple_Specific",
			opts: Options{AllowedOrigins: []string{"https://example.com"}},
			setup: func(r *http.Request) {
				r.Header.Set("Origin", "https://example.com")
			},
		},
		{
			name: "Preflight_Specific",
			opts: Options{AllowedOrigins: []string{"https://example.com"}},
			setup: func(r *http.Request) {
				r.Method = http.MethodOptions
				r.Header.Set("Origin", "https://example.com")
				r.Header.Set("Access-Control-Request-Method", "POST")
				r.Header.Set("Access-Control-Request-Headers", "X-Custom")
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			m := newCORS(tc.opts)
			next := func(c routing.RouteContext) {}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
				rec := httptest.NewRecorder()
				if tc.setup != nil {
					tc.setup(req)
				}
				ctx := routing.NewRouteContext(rec, req)
				m.Invoke(ctx, next)
			}
		})
	}
}

// BenchmarkCORS_RouterPipeline measures middleware in a router pipeline.
func BenchmarkCORSRouterPipeline(b *testing.B) {
	r := router.NewRouter()
	UseCORS(r, Options{AllowedOrigins: []string{"*"}})
	r.GET("/test", func(c routing.RouteContext) { c.NoContent() })

	b.Run("Pipeline_Wildcard", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			req.Header.Set("Origin", "https://example.com")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})
}

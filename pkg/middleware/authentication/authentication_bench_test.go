package authentication

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/cookiejar"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

func init() {
	// Silence structured logs during benchmarks to avoid polluting output and costing time.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
}

// BenchmarkAuthentication_Invoke measures middleware.Invoke overhead for several flows.
func BenchmarkAuthentication_Invoke(b *testing.B) {
	cases := []struct {
		name  string
		setup func(r *http.Request)
	}{
		{
			name: "Cookie_Valid",
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: cookiejar.GetUserCookieName(), Value: "valid-token"})
			},
		},
		{
			name: "Bearer_Valid",
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer valid-token")
			},
		},
		{
			name:  "Anonymous_Allowed",
			setup: func(r *http.Request) {},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			// setup middleware with validator that accepts "valid-token"
			m := &authenticationMiddleware{provider: &defaultTokenProvider{validateFn: func(token string) (claims.Principal, error) {
				if token == "valid-token" {
					return newMockPrincipal("bench-user"), nil
				}
				return nil, errors.New("invalid")
			}}}

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
				// for the anonymous allowed case, mark Options accordingly
				if tc.name == "Anonymous_Allowed" {
					opts := ctx.Options()
					if opts == nil {
						// route contexts in tests can have nil Options; set via helper if available
					} else {
						opts.AllowAnonymous = true
					}
				}
				m.Invoke(ctx, next)
			}
		})
	}
}

// BenchmarkAuthentication_RouterPipeline measures the middleware in a router pipeline.
func BenchmarkAuthentication_RouterPipeline(b *testing.B) {
	r := router.NewRouter()
	UseAuthentication(r, WithValidator(func(token string) (claims.Principal, error) {
		if token == "valid-token" {
			return newMockPrincipal("bench-user"), nil
		}
		return nil, errors.New("invalid")
	}), WithTokenTTL(30*time.Minute))

	r.GET("/test", func(c routing.RouteContext) { c.NoContent() })

	b.Run("Pipeline_Bearer", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})
}

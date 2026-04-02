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
	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/cookiekit"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

func init() {
	// Silence structured logs during benchmarks to avoid polluting output and costing time.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
}

const (
	benchURL             = "http://example.com/test"
	bearerPrefix         = "Bearer "
	validToken           = "valid-token"
	caseCookieValid      = "Cookie_Valid"
	caseBearerValid      = "Bearer_Valid"
	caseAnonymousAllowed = "Anonymous_Allowed"
)

// BenchmarkAuthentication_Invoke measures middleware.Invoke overhead for several flows.
func BenchmarkAuthenticationInvoke(b *testing.B) {
	cases := []struct {
		name  string
		setup func(r *http.Request)
	}{
		{
			name: caseCookieValid,
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: cookiekit.GetUserCookieName(), Value: validToken})
			},
		},
		{
			name: caseBearerValid,
			setup: func(r *http.Request) {
				r.Header.Set(common.HeaderAuthorization, bearerPrefix+validToken)
			},
		},
		{
			name: caseAnonymousAllowed,
			setup: func(r *http.Request) {
				// noop
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			m := newBenchAuthMiddleware()
			next := func(c routing.RouteContext) {
				// noop
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ctx := newBenchRouteContext(tc.setup)
				// for the anonymous allowed case, mark Options accordingly
				if tc.name == caseAnonymousAllowed {
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
func BenchmarkAuthenticationRouterPipeline(b *testing.B) {
	r := router.NewRouter()
	UseAuthentication(r, WithValidator(func(token string) (claims.Principal, error) {
		if token == validToken {
			return newMockPrincipal("bench-user"), nil
		}
		return nil, errors.New("invalid")
	}), WithTokenTTL(30*time.Minute))

	r.GET("/test", func(c routing.RouteContext) { c.NoContent() })

	b.Run("Pipeline_Bearer", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodGet, benchURL, nil)
			req.Header.Set(common.HeaderAuthorization, bearerPrefix+validToken)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})
}

package authorization

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// BenchmarkAuthorization_Invoke measures middleware.Invoke overhead for various checks.
func BenchmarkAuthorizationInvoke(b *testing.B) {
	// helper to create a context with a user having roles/scopes
	makeCtx := func() *routing.DefaultRouteContext {
		return newAuthzCtx(newDefaultAuthzUser(), nil)
	}

	cases := []struct {
		name  string
		setup func(m *authorizationMiddleware)
		run   func(m *authorizationMiddleware, ctx *routing.DefaultRouteContext)
	}{
		{
			name: "NoReq_NoOp",
			setup: func(m *authorizationMiddleware) {
				// noop
			},
			run: func(m *authorizationMiddleware, ctx *routing.DefaultRouteContext) {
				m.Invoke(ctx, noop)
			},
		},
		{
			name:  "Role_Match",
			setup: func(m *authorizationMiddleware) { m.options = &AuthorizationOptions{Roles: []string{"admin"}} },
			run: func(m *authorizationMiddleware, ctx *routing.DefaultRouteContext) {
				m.Invoke(ctx, noop)
			},
		},
		{
			name:  "Role_NoMatch",
			setup: func(m *authorizationMiddleware) { m.options = &AuthorizationOptions{Roles: []string{"missing"}} },
			run: func(m *authorizationMiddleware, ctx *routing.DefaultRouteContext) {
				m.Invoke(ctx, noop)
			},
		},
		{
			name:  "Scope_Match",
			setup: func(m *authorizationMiddleware) { m.options = &AuthorizationOptions{Scopes: []string{"read"}} },
			run: func(m *authorizationMiddleware, ctx *routing.DefaultRouteContext) {
				m.Invoke(ctx, noop)
			},
		},
		{
			name: "Permission_Interpolate",
			setup: func(m *authorizationMiddleware) {
				m.options = &AuthorizationOptions{Permissions: []string{"resource:{id}:read"}}
			},
			run: func(m *authorizationMiddleware, ctx *routing.DefaultRouteContext) {
				// set a param used during interpolation using the params pool to avoid benchmark noise
				params := routing.AcquireParams()
				params.Set("id", "42")
				ctx.SetParamsSlice(params)
				m.Invoke(ctx, noop)
				// clear and release params to avoid leaking allocations into subsequent iterations
				ctx.SetParamsSlice(nil)
				routing.ReleaseParams(params)
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			m := &authorizationMiddleware{}
			tc.setup(m)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ctx := makeCtx()
				tc.run(m, ctx)
			}
		})
	}
}

func noop(c routing.RouteContext) {
	c.NoContent()
}

// BenchmarkAuthorization_RouterPipeline measures middleware in a router pipeline.
func BenchmarkAuthorizationRouterPipeline(b *testing.B) {
	r := router.NewRouter()
	UseAuthorization(r, WithRoles("admin"))
	r.GET("/test", func(c routing.RouteContext) { c.NoContent() })

	b.Run("Pipeline_Role", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			// include a user in the context by using middleware chain externally is non-trivial;
			// for pipeline benchmark we'll just hit the route which will have no user and thus be forbidden,
			// still exercises middleware overhead.
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})
}

// Micro-benchmark focusing solely on permission interpolation from Params slice.
func BenchmarkInterpolatePermissions_Slice(b *testing.B) {
	ps := routing.AcquireParams()
	defer routing.ReleaseParams(ps)
	ps.Set("id", "42")
	permissions := []string{"resource:{id}:read", "resource:{id}:write"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolatePermissions(ps, permissions)
	}
}

// Micro-benchmark simulating the legacy map-based interpolation: build a map from Params per-call
// and perform interpolation using the map. This shows the cost of the per-request conversion.
func BenchmarkInterpolatePermissions_MapAlloc(b *testing.B) {
	ps := routing.AcquireParams()
	defer routing.ReleaseParams(ps)
	ps.Set("id", "42")
	permissions := []string{"resource:{id}:read", "resource:{id}:write"}

	// local helper using map[string]string replacements
	interpolateUsingMap := func(replacements map[string]string, permission string) string {
		var result strings.Builder
		var start int
		inPlaceholder := false
		for i, ch := range permission {
			if ch == '{' {
				inPlaceholder = true
				start = i + 1
			} else if ch == '}' && inPlaceholder {
				inPlaceholder = false
				placeholder := permission[start:i]
				replaced := placeholder
				for k, v := range replacements {
					if strings.EqualFold(k, placeholder) {
						replaced = v
						break
					}
				}
				result.WriteString(replaced)
			} else if !inPlaceholder {
				result.WriteRune(ch)
			}
		}
		return result.String()
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate legacy behavior: create a new map from params each request
		replacements := make(map[string]string, ps.Len())
		for j := 0; j < ps.Len(); j++ {
			p := (*ps)[j]
			replacements[p.Key] = p.Value
		}
		// perform interpolation for all permissions
		for _, perm := range permissions {
			_ = interpolateUsingMap(replacements, perm)
		}
	}
}

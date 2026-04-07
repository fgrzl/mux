package authorization

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
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

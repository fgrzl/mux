package mux

import (
	"net/http"

	internalcommon "github.com/fgrzl/mux/internal/common"
	internalrouter "github.com/fgrzl/mux/internal/router"
)

type RouteGroup struct {
	inner *internalrouter.RouteGroup
}

func wrapRouteGroup(inner *internalrouter.RouteGroup) *RouteGroup {
	if inner == nil {
		return nil
	}
	return &RouteGroup{inner: inner}
}

func (g *RouteGroup) Configure(configure func(*RouteGroup)) error {
	return g.inner.Configure(func(_ *internalrouter.RouteGroup) {
		if configure != nil {
			configure(g)
		}
	})
}

func (g *RouteGroup) Use(middleware ...Middleware) *RouteGroup {
	g.inner.Use(toInternalMiddlewares(middleware)...)
	return g
}

func (g *RouteGroup) Services() *ServiceRegistry {
	inner := g.inner.Services()
	return newServiceRegistry(
		func(key ServiceKey, svc any) { inner.Register(internalcommon.ServiceKey(key), svc) },
		func(key ServiceKey) (any, bool) { return inner.Get(internalcommon.ServiceKey(key)) },
	)
}

func (g *RouteGroup) Service(key ServiceKey, svc any) *RouteGroup {
	g.inner.WithService(internalcommon.ServiceKey(key), svc)
	return g
}

func (g *RouteGroup) WithPathParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "path", description, example, true)
}

func (g *RouteGroup) WithQueryParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "query", description, example, false)
}

func (g *RouteGroup) WithRequiredQueryParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "query", description, example, true)
}

func (g *RouteGroup) WithHeaderParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "header", description, example, false)
}

func (g *RouteGroup) WithRequiredHeaderParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "header", description, example, true)
}

func (g *RouteGroup) WithCookieParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "cookie", description, example, false)
}

func (g *RouteGroup) WithRequiredCookieParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "cookie", description, example, true)
}

func (g *RouteGroup) RequireRoles(roles ...string) *RouteGroup {
	g.inner.RequireRoles(roles...)
	return g
}

func (g *RouteGroup) RequireScopes(scopes ...string) *RouteGroup {
	g.inner.RequireScopes(scopes...)
	return g
}

func (g *RouteGroup) RequirePermission(perms ...string) *RouteGroup {
	g.inner.RequirePermission(perms...)
	return g
}

func (g *RouteGroup) WithTags(tags ...string) *RouteGroup {
	g.inner.WithTags(tags...)
	return g
}

func (g *RouteGroup) WithSummary(summary string) *RouteGroup {
	g.inner.WithSummary(summary)
	return g
}

func (g *RouteGroup) WithDescription(description string) *RouteGroup {
	g.inner.WithDescription(description)
	return g
}

func (g *RouteGroup) WithSecurity(sec SecurityRequirement) *RouteGroup {
	g.inner.WithSecurity(toInternalSecurityRequirement(sec))
	return g
}

func (g *RouteGroup) AllowAnonymous() *RouteGroup {
	g.inner.AllowAnonymous()
	return g
}

func (g *RouteGroup) Deprecated() *RouteGroup {
	g.inner.Deprecated()
	return g
}

func (g *RouteGroup) Group(prefix string) *RouteGroup {
	return wrapRouteGroup(g.inner.NewRouteGroup(prefix))
}

func (g *RouteGroup) Handle(method, pattern string, handler http.Handler) *RouteBuilder {
	return wrapRouteBuilder(g.inner.Handle(method, pattern, handler))
}

func (g *RouteGroup) HandleFunc(method, pattern string, handler http.HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.HandleFunc(method, pattern, handler))
}

func (g *RouteGroup) GET(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.GET(pattern, adaptHandler(handler)))
}

func (g *RouteGroup) POST(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.POST(pattern, adaptHandler(handler)))
}

func (g *RouteGroup) PUT(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.PUT(pattern, adaptHandler(handler)))
}

func (g *RouteGroup) PATCH(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.PATCH(pattern, adaptHandler(handler)))
}

func (g *RouteGroup) DELETE(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.DELETE(pattern, adaptHandler(handler)))
}

func (g *RouteGroup) HEAD(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.HEAD(pattern, adaptHandler(handler)))
}

func (g *RouteGroup) OPTIONS(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.OPTIONS(pattern, adaptHandler(handler)))
}

func (g *RouteGroup) TRACE(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.TRACE(pattern, adaptHandler(handler)))
}

func (g *RouteGroup) Healthz() *RouteBuilder {
	return wrapRouteBuilder(g.inner.Healthz())
}

func (g *RouteGroup) HealthzWithReady(isReady func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(g.inner.HealthzWithReady(adaptReadyCheck(isReady)))
}

func (g *RouteGroup) Livez() *RouteBuilder {
	return wrapRouteBuilder(g.inner.Livez())
}

func (g *RouteGroup) LivezWithCheck(isLive func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(g.inner.LivezWithCheck(adaptReadyCheck(isLive)))
}

func (g *RouteGroup) Readyz() *RouteBuilder {
	return wrapRouteBuilder(g.inner.Readyz())
}

func (g *RouteGroup) ReadyzWithCheck(isReady func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(g.inner.ReadyzWithCheck(adaptReadyCheck(isReady)))
}

func (g *RouteGroup) Startupz() *RouteBuilder {
	return wrapRouteBuilder(g.inner.Startupz())
}

func (g *RouteGroup) StartupzWithCheck(hasStarted func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(g.inner.StartupzWithCheck(adaptReadyCheck(hasStarted)))
}

func (g *RouteGroup) StaticFallback(pattern, dir, fallback string) *RouteBuilder {
	return wrapRouteBuilder(g.inner.StaticFallback(pattern, dir, fallback))
}

func (g *RouteGroup) addRouteParam(name, in, description string, example any, required bool) *RouteGroup {
	g.inner.WithParam(name, in, description, example, required)
	return g
}

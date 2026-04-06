package mux

import (
	"net/http"

	internalcommon "github.com/fgrzl/mux/internal/common"
	internalrouter "github.com/fgrzl/mux/internal/router"
)

// RouteGroup registers a set of routes that share a path prefix and inherited
// defaults such as middleware, services, auth requirements, and OpenAPI
// metadata.
type RouteGroup struct {
	inner *internalrouter.RouteGroup
}

func wrapRouteGroup(inner *internalrouter.RouteGroup) *RouteGroup {
	if inner == nil {
		return nil
	}
	return &RouteGroup{inner: inner}
}

// Configure runs nested registration against the group and returns setup-time
// validation errors instead of panicking.
func (g *RouteGroup) Configure(configure func(*RouteGroup)) error {
	return g.inner.Configure(func(_ *internalrouter.RouteGroup) {
		if configure != nil {
			configure(g)
		}
	})
}

// Use attaches middleware to every route in the group.
func (g *RouteGroup) Use(middleware ...Middleware) *RouteGroup {
	g.inner.Use(toInternalMiddlewares(middleware)...)
	return g
}

// Services returns the group-scoped service registry. Services registered here
// are inherited by child groups and routes. Prefer it when registering more
// than one dependency.
func (g *RouteGroup) Services() *ServiceRegistry {
	inner := g.inner.Services()
	return newServiceRegistry(
		func(key ServiceKey, svc any) { inner.Register(internalcommon.ServiceKey(key), svc) },
		func(key ServiceKey) (any, bool) { return inner.Get(internalcommon.ServiceKey(key)) },
	)
}

// Service registers a service on the group so child groups and routes can
// resolve it. It is the singular convenience form of
// Services().Register(...), which is the preferred API when multiple services
// need to be registered together.
func (g *RouteGroup) Service(key ServiceKey, svc any) *RouteGroup {
	g.inner.WithService(internalcommon.ServiceKey(key), svc)
	return g
}

// WithPathParam documents a required path parameter inherited by routes in the
// group. The name should match route placeholders used beneath the group.
func (g *RouteGroup) WithPathParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "path", description, example, true)
}

// WithQueryParam documents an optional query parameter inherited by routes in
// the group.
func (g *RouteGroup) WithQueryParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "query", description, example, false)
}

// WithRequiredQueryParam documents a required query parameter inherited by
// routes in the group.
func (g *RouteGroup) WithRequiredQueryParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "query", description, example, true)
}

// WithHeaderParam documents an optional header parameter inherited by routes in
// the group.
func (g *RouteGroup) WithHeaderParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "header", description, example, false)
}

// WithRequiredHeaderParam documents a required header parameter inherited by
// routes in the group.
func (g *RouteGroup) WithRequiredHeaderParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "header", description, example, true)
}

// WithCookieParam documents an optional cookie parameter inherited by routes in
// the group.
func (g *RouteGroup) WithCookieParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "cookie", description, example, false)
}

// WithRequiredCookieParam documents a required cookie parameter inherited by
// routes in the group.
func (g *RouteGroup) WithRequiredCookieParam(name, description string, example any) *RouteGroup {
	return g.addRouteParam(name, "cookie", description, example, true)
}

// RequireRoles marks every route in the group as requiring the provided roles
// when authorization middleware is enabled.
func (g *RouteGroup) RequireRoles(roles ...string) *RouteGroup {
	g.inner.RequireRoles(roles...)
	return g
}

// RequireScopes marks every route in the group as requiring the provided
// scopes when authorization middleware is enabled.
func (g *RouteGroup) RequireScopes(scopes ...string) *RouteGroup {
	g.inner.RequireScopes(scopes...)
	return g
}

// RequirePermission marks every route in the group as requiring the provided
// permissions when authorization middleware is enabled.
func (g *RouteGroup) RequirePermission(perms ...string) *RouteGroup {
	g.inner.RequirePermission(perms...)
	return g
}

// WithTags applies default OpenAPI tags to routes in the group.
func (g *RouteGroup) WithTags(tags ...string) *RouteGroup {
	g.inner.WithTags(tags...)
	return g
}

// WithSummary applies a default OpenAPI summary to routes in the group.
func (g *RouteGroup) WithSummary(summary string) *RouteGroup {
	g.inner.WithSummary(summary)
	return g
}

// WithDescription applies a default OpenAPI description to routes in the
// group.
func (g *RouteGroup) WithDescription(description string) *RouteGroup {
	g.inner.WithDescription(description)
	return g
}

// WithSecurity applies default OpenAPI security requirements to routes in the
// group. Pair it with auth middleware for runtime enforcement.
func (g *RouteGroup) WithSecurity(sec SecurityRequirement) *RouteGroup {
	g.inner.WithSecurity(toInternalSecurityRequirement(sec))
	return g
}

// AllowAnonymous clears inherited authentication requirements for routes in
// this group.
func (g *RouteGroup) AllowAnonymous() *RouteGroup {
	g.inner.AllowAnonymous()
	return g
}

// Deprecated marks routes in this group as deprecated by default in generated
// OpenAPI operations.
func (g *RouteGroup) Deprecated() *RouteGroup {
	g.inner.Deprecated()
	return g
}

// Group creates a nested route group beneath prefix. Child groups inherit the
// parent prefix, middleware, services, auth requirements, and metadata.
func (g *RouteGroup) Group(prefix string) *RouteGroup {
	return wrapRouteGroup(g.inner.NewRouteGroup(prefix))
}

// Handle registers a raw http.Handler and returns a RouteBuilder for further
// decoration. Use RouteContextFromRequest inside raw handlers when you need mux
// route state such as params or scoped services.
func (g *RouteGroup) Handle(method, pattern string, handler http.Handler) *RouteBuilder {
	return wrapRouteBuilder(g.inner.Handle(method, pattern, handler))
}

// HandleFunc registers a raw http.HandlerFunc and returns a RouteBuilder for
// further decoration.
func (g *RouteGroup) HandleFunc(method, pattern string, handler http.HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.HandleFunc(method, pattern, handler))
}

// GET registers a GET route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (g *RouteGroup) GET(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.GET(pattern, adaptHandler(handler)))
}

// POST registers a POST route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (g *RouteGroup) POST(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.POST(pattern, adaptHandler(handler)))
}

// PUT registers a PUT route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (g *RouteGroup) PUT(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.PUT(pattern, adaptHandler(handler)))
}

// PATCH registers a PATCH route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (g *RouteGroup) PATCH(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.PATCH(pattern, adaptHandler(handler)))
}

// DELETE registers a DELETE route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (g *RouteGroup) DELETE(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.DELETE(pattern, adaptHandler(handler)))
}

// HEAD registers a HEAD route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (g *RouteGroup) HEAD(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.HEAD(pattern, adaptHandler(handler)))
}

// OPTIONS registers an OPTIONS route and returns a RouteBuilder for middleware
// and OpenAPI decoration.
func (g *RouteGroup) OPTIONS(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.OPTIONS(pattern, adaptHandler(handler)))
}

// TRACE registers a TRACE route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (g *RouteGroup) TRACE(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(g.inner.TRACE(pattern, adaptHandler(handler)))
}

// Healthz registers a GET /healthz probe under the group prefix that always
// reports ready.
func (g *RouteGroup) Healthz() *RouteBuilder {
	return wrapRouteBuilder(g.inner.Healthz())
}

// HealthzWithReady registers a GET /healthz probe under the group prefix with
// a custom readiness check.
func (g *RouteGroup) HealthzWithReady(isReady func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(g.inner.HealthzWithReady(adaptReadyCheck(isReady)))
}

// Livez registers a GET /livez probe under the group prefix that always
// reports live.
func (g *RouteGroup) Livez() *RouteBuilder {
	return wrapRouteBuilder(g.inner.Livez())
}

// LivezWithCheck registers a GET /livez probe under the group prefix with a
// custom liveness check.
func (g *RouteGroup) LivezWithCheck(isLive func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(g.inner.LivezWithCheck(adaptReadyCheck(isLive)))
}

// Readyz registers a GET /readyz probe under the group prefix that always
// reports ready.
func (g *RouteGroup) Readyz() *RouteBuilder {
	return wrapRouteBuilder(g.inner.Readyz())
}

// ReadyzWithCheck registers a GET /readyz probe under the group prefix with a
// custom readiness check.
func (g *RouteGroup) ReadyzWithCheck(isReady func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(g.inner.ReadyzWithCheck(adaptReadyCheck(isReady)))
}

// Startupz registers a GET /startupz probe under the group prefix that always
// reports started.
func (g *RouteGroup) Startupz() *RouteBuilder {
	return wrapRouteBuilder(g.inner.Startupz())
}

// StartupzWithCheck registers a GET /startupz probe under the group prefix
// with a custom startup check.
func (g *RouteGroup) StartupzWithCheck(hasStarted func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(g.inner.StartupzWithCheck(adaptReadyCheck(hasStarted)))
}

// StaticFallback serves files matched by pattern from dir and falls back to
// fallback when no file matches. This is useful for SPA shells such as
// "/app/**" backed by an index.html fallback.
func (g *RouteGroup) StaticFallback(pattern, dir, fallback string) *RouteBuilder {
	return wrapRouteBuilder(g.inner.StaticFallback(pattern, dir, fallback))
}

func (g *RouteGroup) addRouteParam(name, in, description string, example any, required bool) *RouteGroup {
	g.inner.WithParam(name, in, description, example, required)
	return g
}

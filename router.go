package mux

import (
	"fmt"
	"net/http"

	internalcommon "github.com/fgrzl/mux/internal/common"
	internalrouter "github.com/fgrzl/mux/internal/router"
)

// Router registers routes, middleware, services, and top-level API metadata
// for a mux application.
type Router struct {
	inner *internalrouter.Router
}

// NewRouter constructs a router with the supplied API metadata and runtime
// options. The typical startup flow is NewRouter(...), Configure(...), then
// NewServer(...).Listen(ctx).
func NewRouter(opts ...RouterOption) *Router {
	return &Router{inner: internalrouter.NewRouter(toInternalRouterOptions(opts)...)}
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.inner.ServeHTTP(w, req)
}

// Configure runs startup registration with validation errors returned instead
// of panicking. Prefer it for application setup and tests that want one
// explicit error check after route registration.
func (r *Router) Configure(configure func(*Router)) error {
	return r.inner.Configure(func(_ *internalrouter.Router) {
		if configure != nil {
			configure(r)
		}
	})
}

// Use attaches global middleware to the router. Call it during startup before
// serving requests.
func (r *Router) Use(middleware ...Middleware) *Router {
	for _, mw := range middleware {
		r.inner.Use(middlewareAdapter{mw: mw})
	}
	return r
}

// Services returns the root service registry. Services registered here are
// inherited by groups and routes unless they are overridden farther down the
// tree. Prefer it when registering more than one dependency.
func (r *Router) Services() *ServiceRegistry {
	inner := r.inner.Services()
	return newServiceRegistry(
		func(key ServiceKey, svc any) { inner.Register(internalcommon.ServiceKey(key), svc) },
		func(key ServiceKey) (any, bool) { return inner.Get(internalcommon.ServiceKey(key)) },
	)
}

// Service registers a service on the root router and returns the router for
// fluent startup configuration. It is the singular convenience form of
// Services().Register(...), which is the preferred API when multiple services
// need to be registered together.
func (r *Router) Service(key ServiceKey, svc any) *Router {
	r.inner.WithService(internalcommon.ServiceKey(key), svc)
	return r
}

// Group creates a route group rooted at prefix. Child routes inherit router
// middleware, services, auth requirements, and metadata defaults.
func (r *Router) Group(prefix string) *RouteGroup {
	return wrapRouteGroup(r.inner.NewRouteGroup(prefix))
}

// Handle registers a raw http.Handler and returns a RouteBuilder for further
// decoration. Use RouteContextFromRequest inside raw handlers when you need mux
// route state such as params or scoped services.
func (r *Router) Handle(method, pattern string, handler http.Handler) *RouteBuilder {
	return wrapRouteBuilder(r.inner.Handle(method, pattern, handler))
}

// HandleFunc registers a raw http.HandlerFunc and returns a RouteBuilder for
// further decoration.
func (r *Router) HandleFunc(method, pattern string, handler http.HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.HandleFunc(method, pattern, handler))
}

// GET registers a GET route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (r *Router) GET(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.GET(pattern, adaptHandler(handler)))
}

// POST registers a POST route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (r *Router) POST(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.POST(pattern, adaptHandler(handler)))
}

// PUT registers a PUT route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (r *Router) PUT(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.PUT(pattern, adaptHandler(handler)))
}

// PATCH registers a PATCH route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (r *Router) PATCH(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.PATCH(pattern, adaptHandler(handler)))
}

// DELETE registers a DELETE route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (r *Router) DELETE(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.DELETE(pattern, adaptHandler(handler)))
}

// HEAD registers a HEAD route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (r *Router) HEAD(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.HEAD(pattern, adaptHandler(handler)))
}

// OPTIONS registers an OPTIONS route and returns a RouteBuilder for middleware
// and OpenAPI decoration.
func (r *Router) OPTIONS(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.OPTIONS(pattern, adaptHandler(handler)))
}

// TRACE registers a TRACE route and returns a RouteBuilder for middleware and
// OpenAPI decoration.
func (r *Router) TRACE(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.TRACE(pattern, adaptHandler(handler)))
}

// Healthz registers a GET /healthz probe that always reports ready.
func (r *Router) Healthz() *RouteBuilder {
	return wrapRouteBuilder(r.inner.Healthz())
}

// HealthzWithReady registers a GET /healthz probe with a custom readiness
// check.
func (r *Router) HealthzWithReady(isReady func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(r.inner.HealthzWithReady(adaptReadyCheck(isReady)))
}

// Livez registers a GET /livez probe that always reports live.
func (r *Router) Livez() *RouteBuilder {
	return wrapRouteBuilder(r.inner.Livez())
}

// LivezWithCheck registers a GET /livez probe with a custom liveness check.
func (r *Router) LivezWithCheck(isLive func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(r.inner.LivezWithCheck(adaptReadyCheck(isLive)))
}

// Readyz registers a GET /readyz probe that always reports ready.
func (r *Router) Readyz() *RouteBuilder {
	return wrapRouteBuilder(r.inner.Readyz())
}

// ReadyzWithCheck registers a GET /readyz probe with a custom readiness
// check.
func (r *Router) ReadyzWithCheck(isReady func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(r.inner.ReadyzWithCheck(adaptReadyCheck(isReady)))
}

// Startupz registers a GET /startupz probe that always reports started.
func (r *Router) Startupz() *RouteBuilder {
	return wrapRouteBuilder(r.inner.Startupz())
}

// StartupzWithCheck registers a GET /startupz probe with a custom startup
// check.
func (r *Router) StartupzWithCheck(hasStarted func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(r.inner.StartupzWithCheck(adaptReadyCheck(hasStarted)))
}

// StaticFallback serves files matched by pattern from dir and falls back to
// fallback when no file matches. This is useful for SPA shells backed by an
// index.html fallback.
func (r *Router) StaticFallback(pattern, dir, fallback string) *RouteBuilder {
	return wrapRouteBuilder(r.inner.StaticFallback(pattern, dir, fallback))
}

// GenerateSpecWithGenerator creates an OpenAPI specification from the router's
// registered routes. Give each documented route a stable OperationID and
// explicit body and response metadata when you want generated clients and AI
// tooling to stay predictable.
func GenerateSpecWithGenerator(gen *Generator, rtr *Router) (*OpenAPISpec, error) {
	if gen == nil || gen.inner == nil {
		return nil, fmt.Errorf("generator is nil")
	}
	if rtr == nil || rtr.inner == nil {
		return nil, fmt.Errorf("router is nil")
	}
	info, err := rtr.inner.InfoObject()
	if err != nil {
		return nil, err
	}
	routes, err := rtr.inner.Routes()
	if err != nil {
		return nil, err
	}
	spec, err := gen.inner.GenerateSpecFromRoutes(info, routes)
	if err != nil {
		return nil, err
	}
	return wrapOpenAPISpec(spec), nil
}

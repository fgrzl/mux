package mux

import (
	"fmt"
	"net/http"
	"time"

	internalbuilder "github.com/fgrzl/mux/internal/builder"
	internalcommon "github.com/fgrzl/mux/internal/common"
	internalrouter "github.com/fgrzl/mux/internal/router"
	internalrouting "github.com/fgrzl/mux/internal/routing"
)

type RouterOption struct {
	apply internalrouter.RouterOption
}

func WithTitle(title string) RouterOption {
	return RouterOption{apply: internalrouter.WithTitle(title)}
}

func WithSummary(summary string) RouterOption {
	return RouterOption{apply: internalrouter.WithSummary(summary)}
}

func WithDescription(description string) RouterOption {
	return RouterOption{apply: internalrouter.WithDescription(description)}
}

func WithTermsOfService(url string) RouterOption {
	return RouterOption{apply: internalrouter.WithTermsOfService(url)}
}

func WithVersion(version string) RouterOption {
	return RouterOption{apply: internalrouter.WithVersion(version)}
}

func WithContact(name, url, email string) RouterOption {
	return RouterOption{apply: internalrouter.WithContact(name, url, email)}
}

func WithLicense(name, url string) RouterOption {
	return RouterOption{apply: internalrouter.WithLicense(name, url)}
}

func WithClientURL(url string) RouterOption {
	return RouterOption{apply: internalrouter.WithClientURL(url)}
}

func WithContextPooling() RouterOption {
	return RouterOption{apply: internalrouter.WithContextPooling()}
}

func WithHeadFallbackToGet() RouterOption {
	return RouterOption{apply: internalrouter.WithHeadFallbackToGet()}
}

func WithMaxBodyBytes(n int64) RouterOption {
	return RouterOption{apply: internalrouter.WithMaxBodyBytes(n)}
}

func toInternalRouterOptions(opts []RouterOption) []internalrouter.RouterOption {
	if len(opts) == 0 {
		return nil
	}
	internal := make([]internalrouter.RouterOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internal = append(internal, opt.apply)
		}
	}
	return internal
}

type Router struct {
	inner *internalrouter.Router
}

func NewRouter(opts ...RouterOption) *Router {
	return &Router{inner: internalrouter.NewRouter(toInternalRouterOptions(opts)...)}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.inner.ServeHTTP(w, req)
}

func (r *Router) Configure(configure func(*Router)) error {
	return r.inner.Configure(func(_ *internalrouter.Router) {
		if configure != nil {
			configure(r)
		}
	})
}

func (r *Router) Use(middleware ...Middleware) *Router {
	for _, mw := range middleware {
		r.inner.Use(middlewareAdapter{mw: mw})
	}
	return r
}

func (r *Router) Services() *ServiceRegistry {
	inner := r.inner.Services()
	return newServiceRegistry(
		func(key ServiceKey, svc any) { inner.Register(internalcommon.ServiceKey(key), svc) },
		func(key ServiceKey) (any, bool) { return inner.Get(internalcommon.ServiceKey(key)) },
	)
}

func (r *Router) Service(key ServiceKey, svc any) *Router {
	r.inner.WithService(internalcommon.ServiceKey(key), svc)
	return r
}

func (r *Router) Group(prefix string) *RouteGroup {
	return wrapRouteGroup(r.inner.NewRouteGroup(prefix))
}

func (r *Router) Handle(method, pattern string, handler http.Handler) *RouteBuilder {
	return wrapRouteBuilder(r.inner.Handle(method, pattern, handler))
}

func (r *Router) HandleFunc(method, pattern string, handler http.HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.HandleFunc(method, pattern, handler))
}

func (r *Router) GET(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.GET(pattern, adaptHandler(handler)))
}

func (r *Router) POST(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.POST(pattern, adaptHandler(handler)))
}

func (r *Router) PUT(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.PUT(pattern, adaptHandler(handler)))
}

func (r *Router) PATCH(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.PATCH(pattern, adaptHandler(handler)))
}

func (r *Router) DELETE(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.DELETE(pattern, adaptHandler(handler)))
}

func (r *Router) HEAD(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.HEAD(pattern, adaptHandler(handler)))
}

func (r *Router) OPTIONS(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.OPTIONS(pattern, adaptHandler(handler)))
}

func (r *Router) TRACE(pattern string, handler HandlerFunc) *RouteBuilder {
	return wrapRouteBuilder(r.inner.TRACE(pattern, adaptHandler(handler)))
}

func (r *Router) Healthz() *RouteBuilder {
	return wrapRouteBuilder(r.inner.Healthz())
}

func (r *Router) HealthzWithReady(isReady func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(r.inner.HealthzWithReady(adaptReadyCheck(isReady)))
}

func (r *Router) Livez() *RouteBuilder {
	return wrapRouteBuilder(r.inner.Livez())
}

func (r *Router) LivezWithCheck(isLive func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(r.inner.LivezWithCheck(adaptReadyCheck(isLive)))
}

func (r *Router) Readyz() *RouteBuilder {
	return wrapRouteBuilder(r.inner.Readyz())
}

func (r *Router) ReadyzWithCheck(isReady func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(r.inner.ReadyzWithCheck(adaptReadyCheck(isReady)))
}

func (r *Router) Startupz() *RouteBuilder {
	return wrapRouteBuilder(r.inner.Startupz())
}

func (r *Router) StartupzWithCheck(hasStarted func(RouteContext) bool) *RouteBuilder {
	return wrapRouteBuilder(r.inner.StartupzWithCheck(adaptReadyCheck(hasStarted)))
}

func (r *Router) StaticFallback(pattern, dir, fallback string) *RouteBuilder {
	return wrapRouteBuilder(r.inner.StaticFallback(pattern, dir, fallback))
}

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

func (g *RouteGroup) Tags(tags ...string) *RouteGroup {
	g.inner.WithTags(tags...)
	return g
}

func (g *RouteGroup) Summary(summary string) *RouteGroup {
	g.inner.WithSummary(summary)
	return g
}

func (g *RouteGroup) Description(description string) *RouteGroup {
	g.inner.WithDescription(description)
	return g
}

func (g *RouteGroup) Security(sec SecurityRequirement) *RouteGroup {
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

type RouteBuilder struct {
	inner *internalbuilder.RouteBuilder
}

func wrapRouteBuilder(inner *internalbuilder.RouteBuilder) *RouteBuilder {
	if inner == nil {
		return nil
	}
	return &RouteBuilder{inner: inner}
}

func (b *RouteBuilder) AllowAnonymous() *RouteBuilder {
	b.inner.AllowAnonymous()
	return b
}

func (b *RouteBuilder) Use(middleware ...Middleware) *RouteBuilder {
	b.inner.Use(toInternalMiddlewares(middleware)...)
	return b
}

func (b *RouteBuilder) Services() *ServiceRegistry {
	inner := b.inner.Services()
	return newServiceRegistry(
		func(key ServiceKey, svc any) { inner.Register(internalcommon.ServiceKey(key), svc) },
		func(key ServiceKey) (any, bool) { return inner.Get(internalcommon.ServiceKey(key)) },
	)
}

func (b *RouteBuilder) Service(key ServiceKey, svc any) *RouteBuilder {
	b.inner.WithService(internalcommon.ServiceKey(key), svc)
	return b
}

func (b *RouteBuilder) RequirePermission(perms ...string) *RouteBuilder {
	b.inner.RequirePermission(perms...)
	return b
}

func (b *RouteBuilder) RequireRoles(roles ...string) *RouteBuilder {
	b.inner.RequireRoles(roles...)
	return b
}

func (b *RouteBuilder) RequireScopes(scopes ...string) *RouteBuilder {
	b.inner.RequireScopes(scopes...)
	return b
}

func (b *RouteBuilder) RateLimit(limit int, interval time.Duration) *RouteBuilder {
	b.inner.WithRateLimit(limit, interval)
	return b
}

func (b *RouteBuilder) OperationID(id string) *RouteBuilder {
	b.inner.WithOperationID(id)
	return b
}

func (b *RouteBuilder) Summary(summary string) *RouteBuilder {
	b.inner.WithSummary(summary)
	return b
}

func (b *RouteBuilder) Description(description string) *RouteBuilder {
	b.inner.WithDescription(description)
	return b
}

func (b *RouteBuilder) Tags(tags ...string) *RouteBuilder {
	b.inner.WithTags(tags...)
	return b
}

func (b *RouteBuilder) ExternalDocs(url, description string) *RouteBuilder {
	b.inner.WithExternalDocs(url, description)
	return b
}

func (b *RouteBuilder) Security(sec SecurityRequirement) *RouteBuilder {
	b.inner.WithSecurity(toInternalSecurityRequirement(sec))
	return b
}

func (b *RouteBuilder) Deprecated() *RouteBuilder {
	b.inner.WithDeprecated()
	return b
}

func (b *RouteBuilder) WithPathParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "path", description, example, true)
}

func (b *RouteBuilder) WithQueryParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "query", description, example, false)
}

func (b *RouteBuilder) WithRequiredQueryParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "query", description, example, true)
}

func (b *RouteBuilder) WithHeaderParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "header", description, example, false)
}

func (b *RouteBuilder) WithRequiredHeaderParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "header", description, example, true)
}

func (b *RouteBuilder) WithCookieParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "cookie", description, example, false)
}

func (b *RouteBuilder) WithRequiredCookieParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "cookie", description, example, true)
}

func (b *RouteBuilder) AcceptJSON(example any) *RouteBuilder {
	b.inner.WithJsonBody(example)
	return b
}

func (b *RouteBuilder) AcceptJSONOneOf(examples ...any) *RouteBuilder {
	b.inner.WithOneOfJsonBody(examples...)
	return b
}

func (b *RouteBuilder) AcceptJSONAnyOf(examples ...any) *RouteBuilder {
	b.inner.WithAnyOfJsonBody(examples...)
	return b
}

func (b *RouteBuilder) AcceptJSONAllOf(examples ...any) *RouteBuilder {
	b.inner.WithAllOfJsonBody(examples...)
	return b
}

func (b *RouteBuilder) AcceptForm(example any) *RouteBuilder {
	b.inner.WithFormBody(example)
	return b
}

func (b *RouteBuilder) AcceptMultipart(example any) *RouteBuilder {
	b.inner.WithMultipartBody(example)
	return b
}

func (b *RouteBuilder) Responds(code int, example any) *RouteBuilder {
	b.inner.WithResponse(code, example)
	return b
}

func (b *RouteBuilder) OK(example any) *RouteBuilder {
	b.inner.WithOKResponse(example)
	return b
}

func (b *RouteBuilder) Created(example any) *RouteBuilder {
	b.inner.WithCreatedResponse(example)
	return b
}

func (b *RouteBuilder) Accepted(example any) *RouteBuilder {
	b.inner.WithAcceptedResponse(example)
	return b
}

func (b *RouteBuilder) NoContent() *RouteBuilder {
	b.inner.WithNoContentResponse()
	return b
}

func (g *RouteGroup) addRouteParam(name, in, description string, example any, required bool) *RouteGroup {
	g.inner.WithParam(name, in, description, example, required)
	return g
}

func (b *RouteBuilder) addRouteParam(name, in, description string, example any, required bool) *RouteBuilder {
	b.inner.WithParam(name, in, description, example, required)
	return b
}

func adaptHandler(handler HandlerFunc) internalrouting.HandlerFunc {
	if handler == nil {
		return nil
	}
	return func(c internalrouting.RouteContext) {
		handler(wrapRouteContext(c))
	}
}

func adaptReadyCheck(check func(RouteContext) bool) func(internalrouting.RouteContext) bool {
	if check == nil {
		return nil
	}
	return func(c internalrouting.RouteContext) bool {
		return check(wrapRouteContext(c))
	}
}

func toInternalMiddlewares(middleware []Middleware) []internalrouting.Middleware {
	if len(middleware) == 0 {
		return nil
	}
	internal := make([]internalrouting.Middleware, 0, len(middleware))
	for _, mw := range middleware {
		if mw != nil {
			internal = append(internal, middlewareAdapter{mw: mw})
		}
	}
	return internal
}

type middlewareAdapter struct {
	mw Middleware
}

func (m middlewareAdapter) Invoke(c internalrouting.RouteContext, next internalrouting.HandlerFunc) {
	if m.mw == nil {
		next(c)
		return
	}
	m.mw.Invoke(wrapRouteContext(c), func(nextCtx RouteContext) {
		unwrapped := unwrapRouteContext(nextCtx)
		if unwrapped == nil {
			unwrapped = c
		}
		next(unwrapped)
	})
}

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

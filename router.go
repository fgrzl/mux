package mux

import (
	"fmt"
	"net/http"

	internalcommon "github.com/fgrzl/mux/internal/common"
	internalrouter "github.com/fgrzl/mux/internal/router"
)

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

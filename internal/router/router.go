// Package mux provides a lightweight, modular HTTP router for Go with middleware,
// request binding, OpenAPI 3.1 generation, structured responses, and flexible auth support.
package router

import (
	"log/slog"
	"net/http"

	openapi "github.com/fgrzl/mux/internal/openapi"
	"github.com/fgrzl/mux/internal/registry"
	"github.com/fgrzl/mux/internal/routing"
)

type RouteContext = routing.RouteContext

// NewRouter creates a new Router with the given options.
func NewRouter(opts ...RouterOption) *Router {
	options := &RouterOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return &Router{
		RouteGroup: RouteGroup{
			prefix:        "",
			routeRegistry: registry.NewRouteRegistry(),
		},
		options: options,
	}
}

// APIOption is a generic function type for configuring API components.
type APIOption[T any] = func(api T)

// HandlerFunc defines the signature for HTTP request handlers.
type HandlerFunc func(c routing.RouteContext)

// RouteKey uniquely identifies a route by its HTTP method and pattern.
type RouteKey struct {
	Method  string
	Pattern string
}

// Action represents an HTTP action with its method and handler.
type Action struct {
	Method  string
	Handler HandlerFunc
}

// Middleware defines the interface for HTTP middleware components.
type Middleware interface {
	Invoke(c routing.RouteContext, next HandlerFunc)
}

// Router is the main HTTP router that handles routing and middleware execution.
type Router struct {
	RouteGroup
	options *RouterOptions
	// Middleware is exported so internal packages and tests can register middleware.
	middleware []Middleware
}

// Use registers a middleware with the router.
func (rtr *Router) Use(m Middleware) {
	rtr.middleware = append(rtr.middleware, m)
}

// NewRouteGroup creates a new route group with the specified prefix.
// The prefix will be added to all routes using this router (e.g., /api/v1).
func (rtr *Router) NewRouteGroup(prefix string) *RouteGroup {
	prefix = normalizeRoute(prefix, "/")
	return newRouteGroupBase(prefix, rtr.routeRegistry)
}

// ServeHTTP implements http.Handler.
func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	c := routing.NewRouteContext(w, r)

	// Panic recovery
	defer func() {
		if err := recover(); err != nil {
			slog.ErrorContext(c, "panic recovered in ServeHTTP", "error", err, "path", r.URL.Path, "method", r.Method)
			c.ServerError("Internal Server Error", "An unexpected error occurred")
		}
	}()

	// lookup the route options using the pattern and method
	options, params, ok := rtr.routeRegistry.Load(r.URL.Path, r.Method)
	if !ok {
		slog.DebugContext(c, "not found", "path", r.URL.Path, "method", r.Method)
		c.NotFound()
		return
	}

	c.SetOptions(options)
	c.SetParams(params)
	c.SetClientURL(rtr.options.clientURL)

	// start the pipeline
	var next HandlerFunc
	index := 0
	mw := rtr.middleware
	next = func(c routing.RouteContext) {
		if index < len(mw) {
			current := mw[index]
			index++
			current.Invoke(c, next)
		} else {
			c.Options().Handler(c)
		}
	}
	next(c)
}

func (rtr *Router) InfoObject() (*openapi.InfoObject, error) {
	return rtr.options.openapi, nil
}

func (rtr *Router) Routes() ([]openapi.RouteData, error) {
	root := rtr.routeRegistry.Root()
	return collectRoutesFromNode(root)
}

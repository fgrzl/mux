// Package mux provides a lightweight, modular HTTP router for Go with middleware,
// request binding, OpenAPI 3.1 generation, structured responses, and flexible auth support.
package mux

import (
	"log/slog"
	"net/http"
)

// NewRouter creates a new Router with the given options.
func NewRouter(opts ...RouterOption) *Router {
	options := &RouterOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return &Router{
		RouteGroup: RouteGroup{
			prefix:   "",
			registry: NewRouteRegistry(),
		},
		options: options,
	}
}

// APIOption is a generic function type for configuring API components.
type APIOption[T any] = func(api T)

// HandlerFunc defines the signature for HTTP request handlers.
type HandlerFunc func(c *RouteContext)

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
	Invoke(ctx *RouteContext, next HandlerFunc)
}

// Router is the main HTTP router that handles routing and middleware execution.
type Router struct {
	RouteGroup
	options    *RouterOptions
	middleware []Middleware
}

// NewRouteGroup creates a new route group with the specified prefix.
// The prefix will be added to all routes using this router (e.g., /api/v1).
func (rtr *Router) NewRouteGroup(prefix string) *RouteGroup {
	prefix = normalizeRoute(prefix, "/")
	return newRouteGroupBase(prefix, rtr.registry)
}

// ServeHTTP implements http.Handler.
func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewRouteContext(w, r)

	// Panic recovery
	defer func() {
		if err := recover(); err != nil {
			slog.ErrorContext(c, "panic recovered in ServeHTTP", "error", err, "path", r.URL.Path, "method", r.Method)
			c.ServerError("Internal Server Error", "An unexpected error occurred")
		}
	}()

	// lookup the route options using the pattern and method
	options, params, ok := rtr.registry.Load(r.URL.Path, r.Method)
	if !ok {
		slog.DebugContext(c, "not found", "path", r.URL.Path, "method", r.Method)
		c.NotFound()
		return
	}

	c.Options = options
	c.Params = params

	// start the pipeline
	var next HandlerFunc
	index := 0
	middleware := rtr.middleware
	next = func(c *RouteContext) {
		if index < len(middleware) {
			current := middleware[index]
			index++
			current.Invoke(c, next)
		} else {
			c.Options.Handler(c)
		}
	}
	next(c)
}

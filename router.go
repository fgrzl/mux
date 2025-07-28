package mux

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func NewRouter(opts ...RouterOption) *Router {
	options := &RouterOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return &Router{
		RouteGroup: RouteGroup{
			prefix:       "",
			registry:     NewRouteRegistry(),
			authProvider: options.authProvider,
		},
		options:      options,
		authProvider: options.authProvider,
	}
}

type APIOption[T any] = func(api T)

type HandlerFunc func(c *RouteContext)

type RouteKey struct {
	Method  string
	Pattern string
}

type Action struct {
	Method  string
	Handler HandlerFunc
}

type Middleware interface {
	Invoke(ctx *RouteContext, next HandlerFunc)
}

type Router struct {
	RouteGroup
	options      *RouterOptions
	authProvider AuthProvider
	middleware   []Middleware
}

// the prefix that will be add to all routes using this router (i.e. /api/v1)
func (rtr *Router) NewRouteGroup(prefix string) *RouteGroup {
	prefix = normalizeRoute(prefix, "/")
	return newRouteGroupBase(prefix, rtr.authProvider, rtr.registry)
}

// ServeHTTP implements http.Handler.
func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewRouteContext(w, r)

	// Panic recovery
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic recovered in ServeHTTP", "error", err, "path", r.URL.Path, "method", r.Method)

			// Check if we're in development mode
			env := os.Getenv("GO_ENV")
			if env == "" {
				env = os.Getenv("ENVIRONMENT")
			}

			if env == "development" || env == "dev" {
				// In development, provide detailed error information
				c.ServerError("Panic Recovered", fmt.Sprintf("A panic occurred: %v", err))
			} else {
				// In production, provide generic error message
				c.ServerError("Internal Server Error", "An unexpected error occurred")
			}
		}
	}()

	// lookup the route options using the pattern and method
	options, params, ok := rtr.registry.Load(r.URL.Path, r.Method)
	if !ok {
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

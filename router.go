package mux

import (
	"net/http"
	"strings"
)

type HandlerFunc func(c *RouteContext)

type RouteKey struct {
	Method  string
	Pattern string
}

type Router struct {
	prefix     string
	middleware []Middleware
	registry   *RouteRegistry
}

// the prefix that will be add to all routes using this router (i.e. /api/v1)
func NewRouter(prefix string) *Router {
	prefix = normalizeRoute(prefix, "/")
	rtr := &Router{
		prefix:   prefix,
		registry: NewRouteRegistry(),
	}
	return rtr
}

type Middleware interface {
	Invoke(ctx *RouteContext, next HandlerFunc)
}

// ServeHTTP implements http.Handler.
func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	c := NewRouteContext(w, r)

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

type Action struct {
	Method  string
	Handler HandlerFunc
}

func (s *Router) HEAD(pattern string, handler HandlerFunc) *RouteBuilder {
	return s.registerRoute(http.MethodHead, pattern, handler)
}

func (s *Router) GET(pattern string, handler HandlerFunc) *RouteBuilder {
	return s.registerRoute(http.MethodGet, pattern, handler)
}

func (s *Router) POST(pattern string, handler HandlerFunc) *RouteBuilder {
	return s.registerRoute(http.MethodPost, pattern, handler)
}

func (s *Router) PUT(pattern string, handler HandlerFunc) *RouteBuilder {
	return s.registerRoute(http.MethodPut, pattern, handler)
}

func (s *Router) DELETE(pattern string, handler HandlerFunc) *RouteBuilder {
	return s.registerRoute(http.MethodDelete, pattern, handler)
}

func (rtr *Router) registerRoute(method string, pattern string, handler HandlerFunc) *RouteBuilder {

	pattern = normalizeRoute(pattern, rtr.prefix)

	options := &RouteOptions{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	}

	rtr.registry.Register(pattern, method, options)

	return &RouteBuilder{Options: options}
}

func normalizeRoute(route string, prefix string) string {
	// Remove extra slashes from prefix and route to prevent "//"
	prefix = strings.TrimRight(prefix, "/")
	route = strings.TrimLeft(route, "/")

	// Prepend the prefix if not already present
	if !strings.HasPrefix(route, prefix) {
		route = prefix + "/" + route
	}

	// Ensure route starts with a "/"
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}

	// Ensure route ends with a "/"
	if !strings.HasSuffix(route, "/") {
		route = route + "/"
	}

	// Replace multiple slashes with a single slash
	route = strings.ReplaceAll(route, "//", "/")

	return route
}

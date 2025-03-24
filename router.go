package mux

import (
	"net/http"
)

func NewRouter() *Router {
	return &Router{
		RouteGroup: RouteGroup{
			prefix:   "",
			registry: NewRouteRegistry(),
		}}
}

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
	middleware []Middleware
}

// the prefix that will be add to all routes using this router (i.e. /api/v1)
func (rtr *Router) NewRouteGroup(prefix string) *RouteGroup {
	prefix = normalizeRoute(prefix, "/")
	return &RouteGroup{
		prefix:   prefix,
		registry: rtr.registry,
	}
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

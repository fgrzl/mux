package mux

import (
	"net/http"
	"time"

	"github.com/fgrzl/claims/jwtkit"
)

func NewRouterOptions() *RouterOptions {
	return &RouterOptions{}
}

type OpenAPIMetadata struct {
	Title       string
	Description string
	Version     string
}

type RouterOptions struct {
	authProvider AuthProvider
	openapi      *OpenAPIMetadata
}

func (o *RouterOptions) WithAuth(signer jwtkit.Signer, ttl *time.Duration) *RouterOptions {
	o.authProvider = NewAuthProvider(signer, ttl)
	return o
}

func (o *RouterOptions) WithOpenAPIMetadata(title, version, description string) *RouterOptions {
	o.openapi = &OpenAPIMetadata{
		Title:       title,
		Version:     version,
		Description: description,
	}
	return o
}

func NewRouter(options *RouterOptions) *Router {

	if options == nil {
		options = &RouterOptions{}
	}

	return &Router{
		RouteGroup: RouteGroup{
			prefix:       "",
			registry:     NewRouteRegistry(),
			authProvider: options.authProvider,
		},
		authProvider: options.authProvider,
	}
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
	authProvider AuthProvider
	middleware   []Middleware
}

// the prefix that will be add to all routes using this router (i.e. /api/v1)
func (rtr *Router) NewRouteGroup(prefix string) *RouteGroup {
	prefix = normalizeRoute(prefix, "/")
	return &RouteGroup{
		prefix:       prefix,
		authProvider: rtr.authProvider,
		registry:     rtr.registry,
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

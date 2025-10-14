// Package mux provides a lightweight, modular HTTP router for Go with middleware,
// request binding, OpenAPI 3.1 generation, structured responses, and flexible auth support.
package router

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync/atomic"

	openapi "github.com/fgrzl/mux/pkg/openapi"
	"github.com/fgrzl/mux/pkg/registry"
	"github.com/fgrzl/mux/pkg/routing"
)

// RouteContext is an alias for routing.RouteContext exposed for callers.
type RouteContext = routing.RouteContext

// NewRouter creates a new Router with the given options.
func NewRouter(opts ...RouterOption) *Router {
	options := &RouterOptions{}
	for _, opt := range opts {
		opt(options)
	}

	r := &Router{
		RouteGroup: RouteGroup{
			prefix:        "",
			routeRegistry: registry.NewRouteRegistry(),
		},
		options: options,
	}
	// initialize pipeline with a default final handler to avoid storing nil
	// into atomic.Value (which panics). The handler will call the route's
	// configured handler when executed. We also store the current middleware
	// count to detect changes and rebuild lazily in ServeHTTP.
	defaultHandler := func(c routing.RouteContext) {
		c.Options().Handler(c)
	}
	r.pipeline.Store(pipelineCache{h: HandlerFunc(defaultHandler), mwCount: 0})
	return r
}

// HandlerFunc defines the signature for HTTP request handlers.
// HandlerFunc defines the signature for HTTP request handlers.
// The function receives a RouteContext which contains request/response helpers
// and route-specific options.
type HandlerFunc func(c routing.RouteContext)

// RouteKey uniquely identifies a route by its HTTP method and pattern.
// RouteKey uniquely identifies a route by its HTTP method and pattern.
type RouteKey struct {
	Method  string
	Pattern string
}

// Action represents an HTTP action with its method and handler.
// Action represents an HTTP action with its method and handler.
type Action struct {
	Method  string
	Handler HandlerFunc
}

// Middleware defines the interface for HTTP middleware components.
// Middleware defines the interface for HTTP middleware components.
// Implementations should perform pre/post processing and call next when ready.
type Middleware interface {
	Invoke(c routing.RouteContext, next HandlerFunc)
}

// Router is the main HTTP router that handles routing and middleware execution.
// Router is the main HTTP router that handles routing and middleware execution.
type Router struct {
	RouteGroup
	options *RouterOptions
	// Middleware is exported so internal packages and tests can register middleware.
	middleware []Middleware
	// pipeline caches the composed middleware chain (HandlerFunc). It is
	// rebuilt when middleware are added via Use. Stored with atomic.Value
	// to avoid per-request locking and allocations.
	pipeline atomic.Value // holds pipelineCache
}

// pipelineCache stores the composed handler and the middleware count used to build it.
type pipelineCache struct {
	h       HandlerFunc
	mwCount int
}

// headWriter wraps a ResponseWriter and discards body writes. Used when serving
// HEAD requests via the GET handler so headers and status codes are preserved
// but no body is written.
type headWriter struct{ http.ResponseWriter }

func (hw headWriter) Write(p []byte) (int, error) { return len(p), nil }

// routeOutcome captures the result of resolving a request to a route.
type routeOutcome int

const (
	routeOutcomeResolved routeOutcome = iota
	routeOutcomeMethodNotAllowed
	routeOutcomeNotFound
)

// routeResolution collects data produced while resolving a request.
type routeResolution struct {
	options      *routing.RouteOptions
	params       routing.RouteParams
	details      registry.LoadDetails
	suppressBody bool
}

// Use registers a middleware with the router.
func (rtr *Router) Use(m Middleware) {
	rtr.middleware = append(rtr.middleware, m)
	// Compose the pipeline immediately and cache it so the first request
	// doesn't pay for pipeline construction.
	mw := rtr.middleware
	var final HandlerFunc = func(c routing.RouteContext) {
		c.Options().Handler(c)
	}
	for i := len(mw) - 1; i >= 0; i-- {
		m := mw[i]
		next := final
		final = func(c routing.RouteContext) {
			m.Invoke(c, next)
		}
	}
	rtr.pipeline.Store(pipelineCache{h: final, mwCount: len(mw)})
}

// NewRouteGroup creates a new route group with the specified prefix.
// The prefix will be added to all routes using this router (e.g., /api/v1).
// NewRouteGroup creates a new route group with the specified prefix.
// The prefix will be added to all routes using this router (e.g., /api/v1).
func (rtr *Router) NewRouteGroup(prefix string) *RouteGroup {
	prefix = normalizeRoute(prefix, "/")
	return newRouteGroupBase(prefix, rtr.routeRegistry)
}

// ServeHTTP implements http.Handler.
// ServeHTTP implements http.Handler.
func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Inline path normalization (avoid function call overhead)
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	// Fast path: try exact static route lookup before context acquisition
	// This saves ~20ns on static routes by avoiding unnecessary allocations
	if opt, ok := rtr.routeRegistry.LoadExact(r.URL.Path, r.Method); ok {
		c := rtr.acquireRouteContext(w, r)
		defer rtr.recoverAndRelease(w, r, c)
		c.SetOptions(opt)
		rtr.executePipeline(c)
		return
	}

	// Standard path: full route resolution
	c := rtr.acquireRouteContext(w, r)
	defer rtr.recoverAndRelease(w, r, c)

	res, outcome := rtr.resolveRoute(r, c)
	switch outcome {
	case routeOutcomeNotFound:
		slog.DebugContext(c, "not found", "path", r.URL.Path, "method", r.Method)
		c.NotFound()
		return
	case routeOutcomeMethodNotAllowed:
		if res.details.Allow != "" {
			w.Header().Set("Allow", res.details.Allow)
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rtr.configureContext(c, w, res)
	rtr.executePipeline(c)
}

func (rtr *Router) acquireRouteContext(w http.ResponseWriter, r *http.Request) *routing.DefaultRouteContext {
	if rtr.options != nil && rtr.options.ContextPooling {
		return routing.AcquireContext(w, r)
	}
	return routing.NewRouteContext(w, r)
}

func (rtr *Router) recoverAndRelease(w http.ResponseWriter, r *http.Request, c *routing.DefaultRouteContext) {
	if rec := recover(); rec != nil {
		logCtx := context.Background()
		if r != nil && r.Context() != nil {
			logCtx = r.Context()
		}
		stack := debug.Stack()
		slog.ErrorContext(logCtx, "panic recovered in ServeHTTP", "error", rec, "path", safeURLPath(r), "method", safeMethod(r), "stack", string(stack))
		if c != nil {
			c.ServerError("Internal Server Error", "An unexpected error occurred")
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
	if rtr.options != nil && rtr.options.ContextPooling && c != nil {
		routing.ReleaseContext(c)
	}
}

func safeURLPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return r.URL.Path
}

func safeMethod(r *http.Request) string {
	if r == nil {
		return ""
	}
	return r.Method
}

func (rtr *Router) resolveRoute(r *http.Request, c *routing.DefaultRouteContext) (routeResolution, routeOutcome) {
	res, outcome := rtr.resolveInitialRoute(r, c)
	if r.Method == http.MethodHead && rtr.shouldFallbackToGet() {
		originalOutcome := outcome
		rtr.applyHeadFallback(r, &res)
		if res.details.Found && res.details.MethodOK {
			outcome = routeOutcomeResolved
		} else {
			outcome = originalOutcome
		}
	}
	return res, outcome
}

func (rtr *Router) resolveInitialRoute(r *http.Request, c *routing.DefaultRouteContext) (routeResolution, routeOutcome) {
	if r == nil || r.URL == nil {
		return routeResolution{}, routeOutcomeNotFound
	}

	res := routeResolution{}
	if params := c.Params(); params != nil {
		res.params = params
	}

	path := r.URL.Path
	method := r.Method

	node := rtr.routeRegistry.FindNode(path)
	switch {
	case node == nil || len(node.RouteOptions) == 0:
		return rtr.resolveWithDetailedLookup(path, method, &res)
	default:
		return rtr.resolveNodeWithOptions(node, path, method, &res)
	}
}

func (rtr *Router) resolveWithDetailedLookup(path, method string, res *routeResolution) (routeResolution, routeOutcome) {
	tmp := routing.AcquireRouteParams()
	opt, det := rtr.routeRegistry.LoadDetailedInto(path, method, tmp)
	res.details = det
	if !det.Found {
		routing.ReleaseRouteParams(tmp)
		res.params = nil
		return *res, routeOutcomeNotFound
	}
	res.params = tmp
	if !det.MethodOK {
		return *res, routeOutcomeMethodNotAllowed
	}
	res.options = opt
	return *res, routeOutcomeResolved
}

func (rtr *Router) resolveNodeWithOptions(node *routing.RouteNode, path, method string, res *routeResolution) (routeResolution, routeOutcome) {
	opt, ok := node.RouteOptions[method]
	if !ok {
		res.details = registry.LoadDetails{Found: true, MethodOK: false, Allow: node.AllowHeader}
		return *res, routeOutcomeMethodNotAllowed
	}

	res.details = registry.LoadDetails{Found: true, MethodOK: true}
	res.options = opt
	if node.HasParams {
		params := res.params
		if params == nil {
			// Use parameter count to pre-allocate map with correct capacity
			if node.ParamCount > 0 {
				params = routing.AcquireRouteParamsWithCapacity(node.ParamCount)
			} else {
				params = rtr.newRouteParams()
			}
		}
		if opt2, ok2 := rtr.routeRegistry.LoadInto(path, method, params); ok2 {
			res.options = opt2
		}
		res.params = params
	}
	return *res, routeOutcomeResolved
}

func (rtr *Router) applyHeadFallback(r *http.Request, res *routeResolution) {
	if r == nil || r.URL == nil {
		return
	}

	params := res.params
	if params == nil {
		params = routing.AcquireRouteParams()
	}

	getNode := rtr.routeRegistry.FindNodeInto(r.URL.Path, params)
	if getNode == nil {
		res.params = params
		return
	}

	if opt, ok := getNode.RouteOptions[http.MethodGet]; ok {
		res.options = opt
		res.details = registry.LoadDetails{Found: true, MethodOK: true}
		res.suppressBody = true
		res.params = params
		return
	}

	res.params = params
}

func (rtr *Router) shouldFallbackToGet() bool {
	return rtr.options != nil && rtr.options.HeadFallbackToGet
}

func (rtr *Router) newRouteParams() routing.RouteParams {
	if rtr.options != nil && rtr.options.ContextPooling {
		return routing.AcquireRouteParams()
	}
	return make(routing.RouteParams, 2)
}

func (rtr *Router) configureContext(c *routing.DefaultRouteContext, w http.ResponseWriter, res routeResolution) {
	c.SetOptions(res.options)
	params := res.params
	switch {
	case len(params) > 0:
		c.SetParams(params)
	case params == nil && !rtr.usingContextPooling():
		c.SetParams(nil)
	case params != nil && c.Params() == nil:
		c.SetParams(params)
	}

	if rtr.options != nil {
		c.SetClientURL(rtr.options.clientURL)
		c.SetMaxBodyBytes(rtr.options.MaxBodyBytes)
	}

	if res.suppressBody {
		c.SetResponse(headWriter{ResponseWriter: w})
	}
}

func (rtr *Router) usingContextPooling() bool {
	return rtr.options != nil && rtr.options.ContextPooling
}

func (rtr *Router) executePipeline(c routing.RouteContext) {
	mw := rtr.middleware
	if v := rtr.pipeline.Load(); v != nil {
		if pc, ok := v.(pipelineCache); ok && pc.h != nil && pc.mwCount == len(mw) {
			pc.h(c)
			return
		}
	}

	final := rtr.buildPipeline(mw)
	rtr.pipeline.Store(pipelineCache{h: final, mwCount: len(mw)})
	final(c)
}

func (rtr *Router) buildPipeline(mw []Middleware) HandlerFunc {
	final := HandlerFunc(func(c routing.RouteContext) {
		c.Options().Handler(c)
	})
	for i := len(mw) - 1; i >= 0; i-- {
		middleware := mw[i]
		next := final
		final = func(c routing.RouteContext) {
			middleware.Invoke(c, next)
		}
	}
	return final
}

func (rtr *Router) InfoObject() (*openapi.InfoObject, error) {
	return rtr.options.openapi, nil
}

// Routes returns a list of OpenAPI route metadata collected from the registry.
func (rtr *Router) Routes() ([]openapi.RouteData, error) {
	root := rtr.routeRegistry.Root()
	return collectRoutesFromNode(root)
}

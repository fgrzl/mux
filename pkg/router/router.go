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
			validation:    routing.NewValidationState(),
		},
		options: options,
	}
	// initialize pipeline with a default final handler to avoid storing nil
	// into atomic.Value (which panics). The handler will call the route's
	// configured handler when executed. We also store the current middleware
	// count to detect changes and rebuild lazily in ServeHTTP.
	defaultHandler := func(c routing.RouteContext) {
		invokeRouteHandler(c)
	}
	r.pipeline.Store(pipelineCache{h: HandlerFunc(defaultHandler), mwCount: 0})
	return r
}

// HandlerFunc defines the signature for HTTP request handlers.
// HandlerFunc defines the signature for HTTP request handlers.
// The function receives a RouteContext which contains request/response helpers
// and route-specific options.
type HandlerFunc = routing.HandlerFunc

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
type Middleware = routing.Middleware

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

// Safe switches the router's configuration tree into non-panicking validation
// mode and returns the router for fluent startup configuration.
// Prefer Configure in new application code when you want a single error return
// for startup validation.
func (rtr *Router) Safe() *Router {
	rtr.RouteGroup.Safe()
	return rtr
}

// Configure runs startup configuration with validation errors returned instead
// of panicking. This is the recommended entry point for application setup that
// wants one explicit error check after route registration.
func (rtr *Router) Configure(configure func(*Router)) error {
	if configure == nil {
		return nil
	}

	original := rtr.RouteGroup.validationState()
	configured := original.WithPanicOnError(false)
	rtr.RouteGroup.validation = configured
	defer func() {
		rtr.RouteGroup.validation = original
	}()

	configure(rtr)
	return configured.Err()
}

// Errors returns accumulated configuration errors for the router tree.
func (rtr *Router) Errors() []error {
	return rtr.RouteGroup.Errors()
}

// Err returns accumulated configuration errors for the router tree.
func (rtr *Router) Err() error {
	return rtr.RouteGroup.Err()
}

// Services returns a fluent registry for configuring scoped services on the
// root router.
func (rtr *Router) Services() *routing.ServiceRegistry {
	return rtr.RouteGroup.Services()
}

// WithService registers a service on the root router and returns the router
// for fluent startup configuration.
func (rtr *Router) WithService(key routing.ServiceKey, svc any) *Router {
	rtr.RouteGroup.WithService(key, svc)
	return rtr
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
	paramsSlice  *routing.Params // Optimized slice-based parameter storage
	details      registry.LoadDetails
	suppressBody bool
}

// Use registers a middleware with the router.
//
// IMPORTANT: Use() must be called during application startup, before calling
// http.ListenAndServe() or starting any concurrent request handling. The middleware
// list is not protected by a mutex for performance reasons, so calling Use() while
// the router is handling requests will cause data races and undefined behavior.
//
// Safe usage pattern:
//
//	rtr := router.NewRouter()
//	rtr.Use(&myMiddleware{})    // OK: during startup
//	rtr.GET("/api/users", ...)  // OK: during startup
//	http.ListenAndServe(":8080", rtr) // Now serving requests
//
// Unsafe - DO NOT DO THIS:
//
//	go func() {
//	    rtr.Use(&myMiddleware{}) // UNSAFE: concurrent modification!
//	}()
func (rtr *Router) Use(m Middleware) {
	rtr.middleware = append(rtr.middleware, m)
	// Compose the pipeline immediately and cache it so the first request
	// doesn't pay for pipeline construction.
	mw := rtr.middleware
	var final HandlerFunc = func(c routing.RouteContext) {
		invokeRouteHandler(c)
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
	group := newRouteGroupBase(prefix, rtr.routeRegistry)
	group.copyDefaults(&rtr.RouteGroup)
	return group
}

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
		// Manual release instead of defer for ~5-10ns improvement
		rtr.configureContext(c, w, routeResolution{options: opt})

		// Skip middleware pipeline if no middleware configured (~20-30ns faster)
		if len(rtr.middleware) == 0 {
			rtr.executeHandlerWithRecover(opt.EffectiveHandler(), c, w, r)
			rtr.releaseContext(c)
		} else {
			rtr.executePipelineWithRecover(c, w, r)
			rtr.releaseContext(c)
		}
		return
	}

	// Standard path: full route resolution
	c := rtr.acquireRouteContext(w, r)

	res, outcome := rtr.resolveRoute(r, c)
	switch outcome {
	case routeOutcomeNotFound:
		c.NotFound()
		rtr.releaseContext(c)
		return
	case routeOutcomeMethodNotAllowed:
		if res.details.Allow != "" {
			w.Header().Set("Allow", res.details.Allow)
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		rtr.releaseContext(c)
		return
	}

	rtr.configureContext(c, w, res)

	// Skip middleware pipeline if no middleware configured (~20-30ns faster)
	if len(rtr.middleware) == 0 {
		rtr.executeHandlerWithRecover(res.options.EffectiveHandler(), c, w, r)
		rtr.releaseContext(c)
	} else {
		rtr.executePipelineWithRecover(c, w, r)
		rtr.releaseContext(c)
	}
}

func (rtr *Router) acquireRouteContext(w http.ResponseWriter, r *http.Request) *routing.DefaultRouteContext {
	if rtr.options != nil && rtr.options.ContextPooling {
		return routing.AcquireContext(w, r)
	}
	return routing.NewRouteContext(w, r)
}

// releaseContext returns the context to the pool if pooling is enabled
func (rtr *Router) releaseContext(c *routing.DefaultRouteContext) {
	if rtr.options != nil && rtr.options.ContextPooling && c != nil {
		routing.ReleaseContext(c)
	}
}

// executeHandlerWithRecover executes a handler directly with panic recovery (no middleware)
func (rtr *Router) executeHandlerWithRecover(handler routing.HandlerFunc, c *routing.DefaultRouteContext, w http.ResponseWriter, r *http.Request) {
	defer func() {
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
	}()
	handler(c)
}

// executePipelineWithRecover executes the pipeline with panic recovery
// This is separate from executePipeline to allow non-panic paths to skip defer overhead
func (rtr *Router) executePipelineWithRecover(c *routing.DefaultRouteContext, w http.ResponseWriter, r *http.Request) {
	defer func() {
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
	}()
	rtr.executePipeline(c)
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

func invokeRouteHandler(c routing.RouteContext) {
	if c == nil {
		return
	}
	options := c.Options()
	if options == nil {
		return
	}
	handler := options.EffectiveHandler()
	if handler == nil {
		return
	}
	handler(c)
}

func (rtr *Router) resolveRoute(r *http.Request, c *routing.DefaultRouteContext) (routeResolution, routeOutcome) {
	res, outcome := rtr.resolveInitialRoute(r, c)
	if r.Method == http.MethodHead && rtr.shouldFallbackToGet() && outcome == routeOutcomeMethodNotAllowed {
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
	// Use optimized slice-based params from context, or create one if needed
	if paramsSlice := c.ParamsSlice(); paramsSlice != nil {
		res.paramsSlice = paramsSlice
	} else {
		// Allocate params for non-pooled contexts
		params := &routing.Params{}
		c.SetParamsSlice(params)
		res.paramsSlice = params
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
	opt, det := rtr.routeRegistry.LoadDetailedIntoSlice(path, method, res.paramsSlice)
	res.details = det
	if !det.Found {
		return *res, routeOutcomeNotFound
	}
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
		// Extract parameters using optimized slice-based storage
		if opt2, ok2 := rtr.routeRegistry.LoadIntoSlice(path, method, res.paramsSlice); ok2 {
			res.options = opt2
		}
	}
	return *res, routeOutcomeResolved
}

func (rtr *Router) applyHeadFallback(r *http.Request, res *routeResolution) {
	if r == nil || r.URL == nil {
		return
	}

	getNode := rtr.routeRegistry.FindNodeIntoSlice(r.URL.Path, res.paramsSlice)
	if getNode == nil {
		return
	}

	if opt, ok := getNode.RouteOptions[http.MethodGet]; ok {
		res.options = opt
		res.details = registry.LoadDetails{Found: true, MethodOK: true}
		res.suppressBody = true
	}
}

func (rtr *Router) shouldFallbackToGet() bool {
	return rtr.options != nil && rtr.options.HeadFallbackToGet
}

func (rtr *Router) configureContext(c *routing.DefaultRouteContext, w http.ResponseWriter, res routeResolution) {
	c.SetOptions(res.options)
	if res.options != nil {
		res.options.ApplyServices(c)
	}

	// paramsSlice is already set on the context from resolveRoute

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
		invokeRouteHandler(c)
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
	if rtr.options == nil {
		return nil, nil
	}
	return openapi.CloneInfoObject(rtr.options.openapi), nil
}

// Routes returns a list of OpenAPI route metadata collected from the registry.
func (rtr *Router) Routes() ([]openapi.RouteData, error) {
	root := rtr.routeRegistry.Root()
	return collectRoutesFromNode(root)
}

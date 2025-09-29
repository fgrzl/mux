// Package mux provides a lightweight, modular HTTP router for Go with middleware,
// request binding, OpenAPI 3.1 generation, structured responses, and flexible auth support.
package router

import (
	"log/slog"
	"net/http"
	"sync/atomic"

	openapi "github.com/fgrzl/mux/pkg/openapi"
	"github.com/fgrzl/mux/pkg/registry"
	"github.com/fgrzl/mux/pkg/routing"
)

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
func (rtr *Router) NewRouteGroup(prefix string) *RouteGroup {
	prefix = normalizeRoute(prefix, "/")
	return newRouteGroupBase(prefix, rtr.routeRegistry)
}

// ServeHTTP implements http.Handler.
func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire route context (pooled if enabled)
	var c *routing.DefaultRouteContext
	if rtr.options != nil && rtr.options.ContextPooling {
		c = routing.AcquireContext(w, r)
	} else {
		c = routing.NewRouteContext(w, r)
	}

	// Panic recovery; then ensure context is released back to pool if enabled
	defer func() {
		if err := recover(); err != nil {
			slog.ErrorContext(c, "panic recovered in ServeHTTP", "error", err, "path", r.URL.Path, "method", r.Method)
			c.ServerError("Internal Server Error", "An unexpected error occurred")
		}
		if rtr.options != nil && rtr.options.ContextPooling {
			routing.ReleaseContext(c)
		}
	}()

	// lookup the route options using the pattern and method
	// Reuse the context's params map if available to avoid allocs
	var (
		params  routing.RouteParams
		options *routing.RouteOptions
		det     registry.LoadDetails
	)
	if p := c.Params(); p != nil {
		params = p
	}
	// First, find the node without forcing param allocation
	node := rtr.routeRegistry.FindNode(r.URL.Path)
	if node != nil {
		if len(node.RouteOptions) == 0 {
			// No terminal options; compute LoadDetails for Allow/header behavior.
			tmp := routing.AcquireRouteParams()
			_, det = rtr.routeRegistry.LoadDetailedInto(r.URL.Path, r.Method, tmp)
			if det.Found {
				params = tmp
			} else {
				routing.ReleaseRouteParams(tmp)
			}
		} else if opt, ok := node.RouteOptions[r.Method]; ok {
			// Method allowed at this node. Only allocate/populate params if the node has params.
			if node.HasParams {
				if params == nil {
					if rtr.options != nil && rtr.options.ContextPooling {
						params = routing.AcquireRouteParams()
					} else {
						params = make(routing.RouteParams, 2)
					}
				}
				if opt2, ok2 := rtr.routeRegistry.LoadInto(r.URL.Path, r.Method, params); ok2 {
					options = opt2
					det = registry.LoadDetails{Found: true, MethodOK: true}
				} else {
					// Fallback: still set options even if LoadInto failed unexpectedly
					options = opt
					det = registry.LoadDetails{Found: true, MethodOK: true}
				}
			} else {
				// No params for this pattern (static, wildcard, catch-all)
				options = opt
				det = registry.LoadDetails{Found: true, MethodOK: true}
			}
		} else {
			det = registry.LoadDetails{Found: true, MethodOK: false, Allow: node.AllowHeader}
		}
	} else {
		// No node matched; compute LoadDetails to drive 404/405 logic
		tmp := routing.AcquireRouteParams()
		_, det = rtr.routeRegistry.LoadDetailedInto(r.URL.Path, r.Method, tmp)
		if det.Found {
			params = tmp
		} else {
			routing.ReleaseRouteParams(tmp)
		}
	}
	// Optional HEAD->GET fallback when no explicit HEAD route is registered
	suppressBody := false
	// HEAD fallback: if enabled, attempt to serve via GET regardless of initial match state
	if r.Method == http.MethodHead && rtr.options.HeadFallbackToGet {
		// If we haven't already found a node, attempt to find one for GET without
		// allocating another full traversal when possible. Reuse params if available.
		if params == nil {
			params = routing.AcquireRouteParams()
		}
		// Find GET node; if it's the same node we already inspected this is cheap,
		// otherwise it performs a single traversal.
		getNode := rtr.routeRegistry.FindNodeInto(r.URL.Path, params)
		if getNode != nil {
			if opt, ok := getNode.RouteOptions[http.MethodGet]; ok {
				options = opt
				det = registry.LoadDetails{Found: true, MethodOK: true}
				suppressBody = true
			}
		}
	}
	if !det.Found || !det.MethodOK {
		// If path matches but method not allowed, return 405 with Allow header
		if det.Found && !det.MethodOK {
			if det.Allow != "" {
				w.Header().Set("Allow", det.Allow)
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		slog.DebugContext(c, "not found", "path", r.URL.Path, "method", r.Method)
		c.NotFound()
		return
	}

	c.SetOptions(options)
	// only set params map on context if we had one populated
	if len(params) > 0 {
		c.SetParams(params)
	} else if params == nil && !(rtr.options != nil && rtr.options.ContextPooling) {
		// for non-pooled contexts, ensure params is empty
		c.SetParams(nil)
	} else if params != nil && c.Params() == nil {
		// attach the newly created map to the context
		c.SetParams(params)
	}
	c.SetClientURL(rtr.options.clientURL)
	// apply max body size option to the context
	c.SetMaxBodyBytes(rtr.options.MaxBodyBytes)

	// If suppressBody (HEAD fallback), wrap the ResponseWriter to discard body writes
	if suppressBody {
		c.SetResponse(headWriter{ResponseWriter: w})
	}

	// start the pipeline: try to use cached pipeline if middleware haven't changed
	mw := rtr.middleware
	if v := rtr.pipeline.Load(); v != nil {
		if pc, ok := v.(pipelineCache); ok && pc.h != nil && pc.mwCount == len(mw) {
			pc.h(c)
			return
		}
	}

	// Build pipeline and cache it. Building composes the middleware from last
	// to first so each middleware receives the next handler.
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
	// Cache the composed pipeline for future requests with the current mw count
	rtr.pipeline.Store(pipelineCache{h: final, mwCount: len(mw)})
	final(c)
}

func (rtr *Router) InfoObject() (*openapi.InfoObject, error) {
	return rtr.options.openapi, nil
}

func (rtr *Router) Routes() ([]openapi.RouteData, error) {
	root := rtr.routeRegistry.Root()
	return collectRoutesFromNode(root)
}

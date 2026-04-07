package routing

import (
	"strings"
	"time"

	openapi "github.com/fgrzl/mux/internal/openapi"
)

// HandlerFunc is the handler signature used by routing package.
// It accepts a RouteContext so implementations can be defined within
// the routing package without importing mux, avoiding cycles.
type HandlerFunc = func(RouteContext)

// Middleware defines a request middleware that can wrap a handler.
// Implementations may perform work before and/or after calling next.
type Middleware interface {
	Invoke(c RouteContext, next HandlerFunc)
}

// RouteOptions holds both runtime routing data (handler, auth etc.) and the
// OpenAPI Operation object that will be rendered into the spec.
// The Operation from the openapi package is embedded so code that expects
// fields like Parameters/Responses continues to work.
type RouteOptions struct {
	// ---- runtime routing metadata ----
	Method  string
	Pattern string
	Handler HandlerFunc
	// Middleware stores route-scoped middleware, including middleware
	// inherited from enclosing RouteGroups.
	Middleware []Middleware
	// Services stores route-scoped services, including services inherited
	// from enclosing RouteGroups.
	Services map[ServiceKey]any

	// ---- runtime operations ----
	AllowAnonymous bool
	Roles          []string
	Scopes         []string
	Permissions    []string
	RateLimit      int
	RateInterval   time.Duration

	// ---- OpenAPI documentation ----
	openapi.Operation

	// ParamIndex is a runtime index of parameters for fast lookups.
	// Key format: strings.ToLower(in+":"+name)
	ParamIndex      map[string]*openapi.ParameterObject
	handlerPipeline HandlerFunc
}

// EffectiveHandler returns the handler that should execute for this route,
// including any scoped middleware registered on the RouteOptions.
func (o *RouteOptions) EffectiveHandler() HandlerFunc {
	if o == nil {
		return nil
	}
	if o.handlerPipeline != nil {
		return o.handlerPipeline
	}
	if len(o.Middleware) > 0 {
		o.rebuildHandlerPipeline()
		if o.handlerPipeline != nil {
			return o.handlerPipeline
		}
	}
	return o.Handler
}

// HasMiddleware reports whether the route has any scoped middleware attached.
func (o *RouteOptions) HasMiddleware() bool {
	return o != nil && len(o.Middleware) > 0
}

// HasServices reports whether the route has any scoped services attached.
func (o *RouteOptions) HasServices() bool {
	return o != nil && len(o.Services) > 0
}

// AppendMiddleware appends scoped middleware to this route and rebuilds the
// cached handler pipeline.
func (o *RouteOptions) AppendMiddleware(middleware ...Middleware) {
	if o == nil || len(middleware) == 0 {
		return
	}
	o.Middleware = append(o.Middleware, middleware...)
	o.rebuildHandlerPipeline()
}

// SetMiddleware replaces the route's scoped middleware and rebuilds the
// cached handler pipeline.
func (o *RouteOptions) SetMiddleware(middleware []Middleware) {
	if o == nil {
		return
	}
	if len(middleware) == 0 {
		o.Middleware = nil
		o.handlerPipeline = nil
		return
	}
	o.Middleware = append(o.Middleware[:0], middleware...)
	o.rebuildHandlerPipeline()
}

// SetService registers a scoped service on this route.
func (o *RouteOptions) SetService(key ServiceKey, svc any) {
	if o == nil || key == "" || svc == nil {
		return
	}
	if o.Services == nil {
		o.Services = make(map[ServiceKey]any)
	}
	o.Services[key] = svc
}

// SetServices replaces the route's scoped services using a shallow copy.
func (o *RouteOptions) SetServices(services map[ServiceKey]any) {
	if o == nil {
		return
	}
	if len(services) == 0 {
		o.Services = nil
		return
	}
	cloned := make(map[ServiceKey]any, len(services))
	for key, svc := range services {
		if key == "" || svc == nil {
			continue
		}
		cloned[key] = svc
	}
	if len(cloned) == 0 {
		o.Services = nil
		return
	}
	o.Services = cloned
}

// ApplyServices sets the route's scoped services on the request context.
func (o *RouteOptions) ApplyServices(c RouteContext) {
	if o == nil || c == nil || len(o.Services) == 0 {
		return
	}
	for key, svc := range o.Services {
		c.SetService(key, svc)
	}
}

func (o *RouteOptions) rebuildHandlerPipeline() {
	if o == nil {
		return
	}
	final := o.Handler
	if final == nil {
		o.handlerPipeline = nil
		return
	}
	for i := len(o.Middleware) - 1; i >= 0; i-- {
		middleware := o.Middleware[i]
		if middleware == nil {
			continue
		}
		next := final
		final = func(c RouteContext) {
			middleware.Invoke(c, next)
		}
	}
	o.handlerPipeline = final
}

// BuildParamIndex constructs a lowercase parameter index keyed by "in:name".
// It returns nil if params is empty or nil.
func BuildParamIndex(params []*openapi.ParameterObject) map[string]*openapi.ParameterObject {
	if len(params) == 0 {
		return nil
	}
	idx := make(map[string]*openapi.ParameterObject, len(params))
	for _, p := range params {
		if p == nil {
			continue
		}
		key := strings.ToLower(p.In + ":" + p.Name)
		idx[key] = p
	}
	return idx
}

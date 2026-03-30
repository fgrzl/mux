package servicelocator

import (
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// ---- Service Middleware ----

// ServiceSetterOptions holds the services to be set on the RouteContext.
//
// Deprecated: use router.Services().Register(...) or the Services() registry
// on route groups and route builders instead.
type ServiceSetterOptions struct {
	Services map[routing.ServiceKey]any
}

// ServiceSetterOption configures ServiceSetterOptions.
//
// Deprecated: use router.Services().Register(...) or the Services() registry
// on route groups and route builders instead.
type ServiceSetterOption func(*ServiceSetterOptions)

// WithService adds a service to the options.
//
// Deprecated: use router.Services().Register(...) or the Services() registry
// on route groups and route builders instead.
func WithService(key routing.ServiceKey, svc any) ServiceSetterOption {
	return func(opts *ServiceSetterOptions) {
		if opts.Services == nil {
			opts.Services = make(map[routing.ServiceKey]any)
		}
		opts.Services[key] = svc
	}
}

// UseServices adds middleware that sets services on the RouteContext.
//
// Deprecated: use (*router.Router).Services().Register(...) or the Services()
// registry on route groups and route builders instead.
func UseServices(rtr *router.Router, opts ...ServiceSetterOption) {
	if rtr == nil {
		return
	}
	options := &ServiceSetterOptions{}
	for _, opt := range opts {
		opt(options)
	}
	rtr.Services().RegisterMany(options.Services)
	rtr.Use(&serviceSetterMiddleware{options: options})
}

// serviceSetterMiddleware implements middleware that injects services into the route context.
type serviceSetterMiddleware struct {
	options *ServiceSetterOptions
}

// Invoke implements the Middleware interface, setting services on the RouteContext.
func (m *serviceSetterMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	if m.options == nil || m.options.Services == nil {
		next(c)
		return
	}
	for k, v := range m.options.Services {
		c.SetService(k, v)
	}
	next(c)
}

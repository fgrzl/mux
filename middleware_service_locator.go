package mux

// ---- Service Middleware ----

// ServiceKey is the type for service map keys.
type ServiceKey string

// ServiceSetterOptions holds the services to be set on the RouteContext.
type ServiceSetterOptions struct {
	Services map[ServiceKey]any
}

// ServiceSetterOption configures ServiceSetterOptions.
type ServiceSetterOption func(*ServiceSetterOptions)

// WithService adds a service to the options.
func WithService(key ServiceKey, svc any) ServiceSetterOption {
	return func(opts *ServiceSetterOptions) {
		if opts.Services == nil {
			opts.Services = make(map[ServiceKey]any)
		}
		opts.Services[key] = svc
	}
}

// UseServices adds middleware that sets services on the RouteContext.
func (rtr *Router) UseServices(opts ...ServiceSetterOption) {
	options := &ServiceSetterOptions{}
	for _, opt := range opts {
		opt(options)
	}
	rtr.middleware = append(rtr.middleware, &serviceSetterMiddleware{options: options})
}

type serviceSetterMiddleware struct {
	options *ServiceSetterOptions
}

func (m *serviceSetterMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	if m.options == nil || m.options.Services == nil {
		next(c)
		return
	}
	for k, v := range m.options.Services {
		c.SetService(k, v)
	}
	next(c)
}

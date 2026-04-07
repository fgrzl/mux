package routing

// ServiceRegistry provides a small fluent API for registering and inspecting
// scoped services on routers, route groups, and route builders.
type ServiceRegistry struct {
	register func(ServiceKey, any)
	lookup   func(ServiceKey) (any, bool)
}

// NewServiceRegistry constructs a ServiceRegistry using caller-provided
// registration and lookup functions.
func NewServiceRegistry(register func(ServiceKey, any), lookup func(ServiceKey) (any, bool)) *ServiceRegistry {
	return &ServiceRegistry{register: register, lookup: lookup}
}

// Register adds or replaces a service and returns the registry for chaining.
func (r *ServiceRegistry) Register(key ServiceKey, svc any) *ServiceRegistry {
	if r == nil || r.register == nil {
		return r
	}
	r.register(key, svc)
	return r
}

// RegisterMany adds all provided services and returns the registry for chaining.
func (r *ServiceRegistry) RegisterMany(services map[ServiceKey]any) *ServiceRegistry {
	if r == nil {
		return r
	}
	for key, svc := range services {
		r.Register(key, svc)
	}
	return r
}

// Get returns a previously registered service if present.
func (r *ServiceRegistry) Get(key ServiceKey) (any, bool) {
	if r == nil || r.lookup == nil {
		return nil, false
	}
	return r.lookup(key)
}

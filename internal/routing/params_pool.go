package routing

import "sync"

// paramsPool reuses small RouteParams maps to reduce allocations for transient
// lookups such as temporary detailed loads or when filling params into pooled
// contexts.
var paramsPool = sync.Pool{
	New: func() any { return make(RouteParams, 2) },
}

// paramsSlicePool reuses Params slices to reduce allocations for route parameter extraction.
// Pre-allocated with capacity 4 which covers most real-world routing scenarios (1-3 params typical).
var paramsSlicePool = sync.Pool{
	New: func() any {
		p := make(Params, 0, 4)
		return &p
	},
}

// AcquireRouteParams returns a RouteParams map from the pool. Callers must
// return it with ReleaseRouteParams when they no longer need it and are not
// attaching it to a long-lived structure.
func AcquireRouteParams() RouteParams {
	return paramsPool.Get().(RouteParams)
}

// AcquireRouteParamsWithCapacity returns a RouteParams map with the specified
// capacity. For capacity <= 2, it uses the pool. For larger capacities, it
// allocates a new map with the exact capacity to avoid growth allocations.
func AcquireRouteParamsWithCapacity(capacity int) RouteParams {
	if capacity <= 2 {
		return paramsPool.Get().(RouteParams)
	}
	return make(RouteParams, capacity)
}

// ReleaseRouteParams clears the map and returns it to the pool. The map must
// not be used after calling ReleaseRouteParams.
func ReleaseRouteParams(m RouteParams) {
	if m == nil {
		return
	}
	for k := range m {
		delete(m, k)
	}
	// Only pool maps with small size to avoid memory bloat
	// Maps don't have a cap() function, so we check length
	if len(m) <= 4 {
		paramsPool.Put(m)
	}
}

// AcquireParams returns a pooled Params slice with capacity 4.
// The returned slice has length 0 but capacity for 4 parameters.
// Callers must return it with ReleaseParams when done.
func AcquireParams() *Params {
	return paramsSlicePool.Get().(*Params)
}

// ReleaseParams resets the slice and returns it to the pool.
// The slice must not be used after calling ReleaseParams.
func ReleaseParams(p *Params) {
	if p == nil {
		return
	}
	p.Reset()
	// Only pool slices with reasonable capacity to avoid memory bloat
	if cap(*p) <= 8 {
		paramsSlicePool.Put(p)
	}
}

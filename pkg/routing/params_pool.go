package routing

import "sync"

// paramsPool reuses small RouteParams maps to reduce allocations for transient
// lookups such as temporary detailed loads or when filling params into pooled
// contexts.
var paramsPool = sync.Pool{
	New: func() any { return make(RouteParams, 2) },
}

// AcquireRouteParams returns a RouteParams map from the pool. Callers must
// return it with ReleaseRouteParams when they no longer need it and are not
// attaching it to a long-lived structure.
func AcquireRouteParams() RouteParams {
	return paramsPool.Get().(RouteParams)
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
	paramsPool.Put(m)
}

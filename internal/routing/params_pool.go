package routing

import "sync"

// paramsSlicePool reuses Params slices to reduce allocations for route parameter extraction.
// Pre-allocated with capacity 4 which covers most real-world routing scenarios (1-3 params typical).
var paramsSlicePool = sync.Pool{
	New: func() any {
		p := make(Params, 0, 4)
		return &p
	},
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

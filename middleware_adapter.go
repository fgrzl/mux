package mux

import internalrouting "github.com/fgrzl/mux/internal/routing"

type middlewareAdapter struct {
	mw Middleware
}

func (m middlewareAdapter) Invoke(c internalrouting.RouteContext, next internalrouting.HandlerFunc) {
	if m.mw == nil {
		next(c)
		return
	}
	m.mw.Invoke(wrapRouteContext(c), func(nextCtx RouteContext) {
		unwrapped := unwrapRouteContext(nextCtx)
		if unwrapped == nil {
			unwrapped = c
		}
		next(unwrapped)
	})
}

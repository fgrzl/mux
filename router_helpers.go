package mux

import internalrouting "github.com/fgrzl/mux/internal/routing"

func adaptHandler(handler HandlerFunc) internalrouting.HandlerFunc {
	if handler == nil {
		return nil
	}
	return func(c internalrouting.RouteContext) {
		handler(wrapRouteContext(c))
	}
}

func adaptReadyCheck(check func(RouteContext) bool) func(internalrouting.RouteContext) bool {
	if check == nil {
		return nil
	}
	return func(c internalrouting.RouteContext) bool {
		return check(wrapRouteContext(c))
	}
}

func toInternalMiddlewares(middleware []Middleware) []internalrouting.Middleware {
	if len(middleware) == 0 {
		return nil
	}
	internal := make([]internalrouting.Middleware, 0, len(middleware))
	for _, mw := range middleware {
		if mw != nil {
			internal = append(internal, middlewareAdapter{mw: mw})
		}
	}
	return internal
}

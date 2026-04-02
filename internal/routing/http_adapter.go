package routing

import (
	"context"
	"net/http"
)

type routeContextRequestKey struct{}

// HTTPHandler adapts a standard-library http.Handler into a routing HandlerFunc.
func HTTPHandler(handler http.Handler) HandlerFunc {
	if handler == nil {
		return nil
	}
	return func(c RouteContext) {
		c.SetRequest(c.Request())
		handler.ServeHTTP(c.Response(), c.Request())
	}
}

// HTTPHandlerFunc adapts a standard-library http.HandlerFunc into a routing HandlerFunc.
func HTTPHandlerFunc(handler http.HandlerFunc) HandlerFunc {
	if handler == nil {
		return nil
	}
	return HTTPHandler(handler)
}

// RouteContextFromRequest returns the active RouteContext attached to an HTTP request.
func RouteContextFromRequest(r *http.Request) (RouteContext, bool) {
	if r == nil {
		return nil, false
	}
	routeCtx, ok := r.Context().Value(routeContextRequestKey{}).(RouteContext)
	if !ok || routeCtx == nil {
		return nil, false
	}
	return routeCtx, true
}

func bindRouteContextToRequest(r *http.Request, routeCtx RouteContext) (*http.Request, context.Context) {
	base := context.Background()
	if r != nil && r.Context() != nil {
		base = r.Context()
	}
	bound := context.WithValue(base, routeContextRequestKey{}, routeCtx)
	if r == nil {
		return nil, bound
	}
	return r.WithContext(bound), bound
}

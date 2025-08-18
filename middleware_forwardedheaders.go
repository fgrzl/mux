package mux

// forwardedHeadersMiddleware processes X-Forwarded-* headers to restore original client information.
type forwardedHeadersMiddleware struct{}

// Invoke implements the Middleware interface, processing X-Forwarded-Proto and X-Forwarded-For headers.
func (m *forwardedHeadersMiddleware) Invoke(c RouteContext, next HandlerFunc) {

	if proto := c.Request().Header.Get(HeaderXForwardedProto); proto != "" {
		c.Request().URL.Scheme = proto
	}
	if ip := c.Request().Header.Get(HeaderXForwardedFor); ip != "" {
		c.Request().RemoteAddr = ip
	}
	next(c)
}

// UseForwardedHeaders adds middleware that processes X-Forwarded-* headers.
func UseForwardedHeaders(router *Router) {
	router.middleware = append(router.middleware, &forwardedHeadersMiddleware{})
}

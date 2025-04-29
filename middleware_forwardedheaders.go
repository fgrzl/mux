package mux

type forwardedHeadersMiddleware struct{}

func (m *forwardedHeadersMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	if proto := c.Request.Header.Get("X-Forwarded-Proto"); proto != "" {
		c.Request.URL.Scheme = proto
	}
	if ip := c.Request.Header.Get("X-Forwarded-For"); ip != "" {
		c.Request.RemoteAddr = ip
	}
	next(c)
}

func (rtr *Router) UseForwardedHeaders() {
	rtr.middleware = append(rtr.middleware, &forwardedHeadersMiddleware{})
}

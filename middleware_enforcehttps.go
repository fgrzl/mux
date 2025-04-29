package mux

import "net/http"

type enforceHTTPSMiddleware struct{}

func (m *enforceHTTPSMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	if c.Request.URL.Scheme != "https" {
		target := "https://" + c.Request.Host + c.Request.URL.RequestURI()
		http.Redirect(c.Response, c.Request, target, http.StatusMovedPermanently)
		return
	}
	next(c)
}

func (rtr *Router) UseEnforceHTTPS() {
	rtr.middleware = append(rtr.middleware, &enforceHTTPSMiddleware{})
}

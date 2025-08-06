package mux

import "net/http"

// enforceHTTPSMiddleware redirects HTTP requests to HTTPS.
type enforceHTTPSMiddleware struct{}

// Invoke implements the Middleware interface, redirecting HTTP requests to HTTPS.
func (m *enforceHTTPSMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	if c.Request.URL.Scheme != "https" {
		target := "https://" + c.Request.Host + c.Request.URL.RequestURI()
		http.Redirect(c.Response, c.Request, target, http.StatusMovedPermanently)
		return
	}
	next(c)
}

// UseEnforceHTTPS adds middleware that redirects HTTP requests to HTTPS.
func (rtr *Router) UseEnforceHTTPS() {
	rtr.middleware = append(rtr.middleware, &enforceHTTPSMiddleware{})
}

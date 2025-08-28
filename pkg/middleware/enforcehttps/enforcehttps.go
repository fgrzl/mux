package enforcehttps

import (
	"net/http"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// enforceHTTPSMiddleware redirects HTTP requests to HTTPS.
type enforceHTTPSMiddleware struct{}

// Invoke implements the Middleware interface, redirecting HTTP requests to HTTPS.
func (m *enforceHTTPSMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	if c.Request().URL.Scheme != "https" {
		target := "https://" + c.Request().Host + c.Request().URL.RequestURI()
		http.Redirect(c.Response(), c.Request(), target, http.StatusMovedPermanently)
		return
	}
	next(c)
}

// UseEnforceHTTPS adds middleware that redirects HTTP requests to HTTPS.
func UseEnforceHTTPS(rtr *router.Router) {
	rtr.Use(&enforceHTTPSMiddleware{})
}

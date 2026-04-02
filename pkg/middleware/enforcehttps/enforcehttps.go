package enforcehttps

import (
	"net/http"
	"strings"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// enforceHTTPSMiddleware redirects HTTP requests to HTTPS.
type enforceHTTPSMiddleware struct{}

// isHTTPS checks if the request is over HTTPS by examining:
// 1. The TLS connection state (most reliable)
// 2. X-Forwarded-Proto header (for reverse proxy scenarios)
// 3. Forwarded header proto= directive (RFC 7239)
func isHTTPS(r *http.Request) bool {
	// Direct TLS connection is the most reliable indicator
	if r.TLS != nil {
		return true
	}

	// Check X-Forwarded-Proto header (commonly set by reverse proxies)
	if proto := r.Header.Get("X-Forwarded-Proto"); strings.EqualFold(proto, "https") {
		return true
	}

	// Check RFC 7239 Forwarded header for proto=https
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		// Simple parse: look for proto=https or proto="https"
		lower := strings.ToLower(fwd)
		if strings.Contains(lower, "proto=https") || strings.Contains(lower, "proto=\"https\"") {
			return true
		}
	}

	return false
}

// Invoke implements the Middleware interface, redirecting HTTP requests to HTTPS.
func (m *enforceHTTPSMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	if !isHTTPS(c.Request()) {
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

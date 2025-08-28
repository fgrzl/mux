package forwardheaders

import (
	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
)

// forwardedHeadersMiddleware processes X-Forwarded-* headers to restore original client information.
type forwardedHeadersMiddleware struct{}

// Invoke implements the Middleware interface, processing X-Forwarded-Proto and X-Forwarded-For headers.
func (m *forwardedHeadersMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {

	if proto := c.Request().Header.Get(common.HeaderXForwardedProto); proto != "" {
		c.Request().URL.Scheme = proto
	}
	if ip := c.Request().Header.Get(common.HeaderXForwardedFor); ip != "" {
		c.Request().RemoteAddr = ip
	}
	next(c)
}

// UseForwardedHeaders adds middleware that processes X-Forwarded-* headers.
func UseForwardedHeaders(rtr *router.Router) {
	rtr.Use(&forwardedHeadersMiddleware{})
}

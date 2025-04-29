package mux

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type otelMiddleware struct {
	operation string
}

func (m *otelMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	handler := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Request = r  // ensure updated context is set
		c.Response = w // not strictly needed but for completeness
		next(c)
	}), m.operation)

	handler.ServeHTTP(c.Response, c.Request)
}

func (rtr *Router) UseOpenTelemetry(operation string) {
	if operation == "" {
		operation = "http.server"
	}

	rtr.middleware = append(rtr.middleware, &otelMiddleware{operation: operation})
}

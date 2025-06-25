package mux

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ---- Functional Options ----

type OpenTelemetryOptions struct {
	Operation string
}

type OpenTelemetryOption func(*OpenTelemetryOptions)

func WithOperation(name string) OpenTelemetryOption {
	return func(o *OpenTelemetryOptions) {
		o.Operation = name
	}
}

func (rtr *Router) UseOpenTelemetry(opts ...OpenTelemetryOption) {
	options := &OpenTelemetryOptions{Operation: "http.server"}
	for _, opt := range opts {
		opt(options)
	}
	rtr.middleware = append(rtr.middleware, &otelMiddleware{operation: options.Operation})
}

// ---- Middleware ----

type otelMiddleware struct {
	operation string
}

func (m *otelMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	handler := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Request = r
		c.Response = w
		next(c)
	}), m.operation)

	handler.ServeHTTP(c.Response, c.Request)
}

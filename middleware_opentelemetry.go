package mux

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ---- Functional Options ----

// OpenTelemetryOptions configures the OpenTelemetry middleware behavior.
type OpenTelemetryOptions struct {
	Operation string
}

// OpenTelemetryOption is a function type for configuring OpenTelemetry options.
type OpenTelemetryOption func(*OpenTelemetryOptions)

// WithOperation sets the operation name for OpenTelemetry tracing.
func WithOperation(name string) OpenTelemetryOption {
	return func(o *OpenTelemetryOptions) {
		o.Operation = name
	}
}

// UseOpenTelemetry adds OpenTelemetry tracing and metrics middleware.
func (rtr *Router) UseOpenTelemetry(opts ...OpenTelemetryOption) {
	options := &OpenTelemetryOptions{Operation: "http.server"}
	for _, opt := range opts {
		opt(options)
	}
	rtr.middleware = append(rtr.middleware, &otelMiddleware{operation: options.Operation})
}

// otelMiddleware provides OpenTelemetry integration for HTTP requests.
type otelMiddleware struct {
	operation string
}

// Invoke implements the Middleware interface, adding OpenTelemetry tracing to HTTP requests.
func (m *otelMiddleware) Invoke(c RouteContext, next HandlerFunc) {
	handler := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next(c)
	}), m.operation)

	handler.ServeHTTP(c.Response(), c.Request())
}

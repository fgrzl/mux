package opentelemetry

import (
	"context"
	"net/http"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
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
func UseOpenTelemetry(rtr *router.Router, opts ...OpenTelemetryOption) {
	options := &OpenTelemetryOptions{Operation: "http.server"}
	for _, opt := range opts {
		opt(options)
	}
	rtr.Use(newOTELMiddleware(options.Operation))
}

// otelMiddleware provides OpenTelemetry integration for HTTP requests.
type otelMiddleware struct {
	operation string
	handler   http.Handler
}

// Invoke implements the Middleware interface, adding OpenTelemetry tracing to HTTP requests.
func (m *otelMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	// Lazy init handler for cases where tests construct the middleware directly.
	if m.handler == nil {
		m.handler = buildOTELHandler(m.operation)
	}
	// Attach per-request data (RouteContext and next) into the request context so the
	// prebuilt handler can retrieve them without capturing per-request closures.
	data := &otelData{c: c, next: next}
	reqWithCtx := c.Request().WithContext(context.WithValue(c, otelNextKey{}, data))
	m.handler.ServeHTTP(c.Response(), reqWithCtx)
}

// newOTELMiddleware constructs an otelMiddleware with a pre-wired handler.
func newOTELMiddleware(operation string) *otelMiddleware {
	mw := &otelMiddleware{operation: operation}
	mw.handler = buildOTELHandler(operation)
	return mw
}

// Shared context key and data payload for passing next and RouteContext.
type otelNextKey struct{}
type otelData struct {
	c    routing.RouteContext
	next router.HandlerFunc
}

func buildOTELHandler(operation string) http.Handler {
	return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Context().Value(otelNextKey{}); v != nil {
			if data, ok := v.(*otelData); ok && data.c != nil && data.next != nil {
				// Prefer a named interface over anonymous for clarity/maintainability
				type requestSetter interface{ SetRequest(*http.Request) }
				if dc, ok2 := data.c.(requestSetter); ok2 {
					dc.SetRequest(r)
				}
				data.next(data.c)
				return
			}
		}
	}), operation)
}

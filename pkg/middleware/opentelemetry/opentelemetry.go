package opentelemetry

import (
	"context"
	"net/http"
	"strings"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	if isWebSocketUpgrade(c.Request()) {
		next(c)
		return
	}

	// Lazy init handler for cases where tests construct the middleware directly.
	if m.handler == nil {
		m.handler = buildOTELHandler(m.operation)
	}
	// Attach per-request data (RouteContext and next) into the request context so the
	// prebuilt handler can retrieve them without capturing per-request closures.
	data := &otelData{c: c, next: next}

	// Use the request context as the base context for values.
	reqCtx := c.Request().Context()
	if opts := c.Options(); opts != nil {
		if opts.Pattern != "" {
			reqCtx = context.WithValue(reqCtx, otelRoutePatternKey{}, opts.Pattern)
		}
		if opts.Method != "" {
			reqCtx = context.WithValue(reqCtx, otelMethodKey{}, opts.Method)
		}
	}

	reqWithCtx := c.Request().WithContext(context.WithValue(reqCtx, otelNextKey{}, data))
	m.handler.ServeHTTP(c.Response(), reqWithCtx)
}

// newOTELMiddleware constructs an otelMiddleware with a pre-wired handler.
func newOTELMiddleware(operation string) *otelMiddleware {
	mw := &otelMiddleware{operation: operation}
	mw.handler = buildOTELHandler(operation)
	return mw
}

// Shared context keys and data payload for passing next, RouteContext, and route metadata.
type otelNextKey struct{}
type otelRoutePatternKey struct{}
type otelMethodKey struct{}

type otelData struct {
	c    routing.RouteContext
	next router.HandlerFunc
}

// RequestSetter is implemented by RouteContext types that can update their
// underlying *http.Request. Having a named interface here improves readability
// and makes type assertions clearer and testable.
type RequestSetter interface{ SetRequest(*http.Request) }

func buildOTELHandler(operation string) http.Handler {
	formatter := func(op string, r *http.Request) string {
		method, pattern := extractRouteTraceData(r.Context())
		if method != "" && pattern != "" {
			return method + " " + pattern
		}
		return op
	}

	return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enrichSpanWithRouteAttributes(r.Context())
		if v := r.Context().Value(otelNextKey{}); v != nil {
			if data, ok := v.(*otelData); ok && data.c != nil && data.next != nil {
				// Use the named RequestSetter interface so callers and tests can
				// rely on a clear, documented contract instead of an anonymous type.
				if dc, ok2 := data.c.(RequestSetter); ok2 {
					dc.SetRequest(r)
				}
				data.next(data.c)
				return
			}
		}
	}), operation, otelhttp.WithSpanNameFormatter(formatter))
}

func extractRouteTraceData(ctx context.Context) (method string, pattern string) {
	if v := ctx.Value(otelMethodKey{}); v != nil {
		if s, ok := v.(string); ok {
			method = s
		}
	}
	if v := ctx.Value(otelRoutePatternKey{}); v != nil {
		if s, ok := v.(string); ok {
			pattern = s
		}
	}
	return method, pattern
}

func enrichSpanWithRouteAttributes(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	method, pattern := extractRouteTraceData(ctx)
	attrs := make([]attribute.KeyValue, 0, 3)
	if pattern != "" {
		attrs = append(attrs,
			attribute.String("http.route", pattern),
			attribute.String("mux.route.pattern", pattern),
		)
	}
	if method != "" {
		attrs = append(attrs, attribute.String("http.request.method", method))
	}
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
}

func isWebSocketUpgrade(r *http.Request) bool {
	if r == nil {
		return false
	}

	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return false
	}

	for _, value := range r.Header.Values("Connection") {
		for _, token := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(token), "Upgrade") {
				return true
			}
		}
	}

	return false
}

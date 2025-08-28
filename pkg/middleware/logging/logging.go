package logging

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// ---- Functional Options ----

// LoggingOptions configures the logging middleware behavior.
type LoggingOptions struct{}

// LoggingOption is a function type for configuring logging options.
type LoggingOption func(*LoggingOptions)

// UseLogging adds structured request/response logging middleware.
func UseLogging(rtr *router.Router, opts ...LoggingOption) {
	options := &LoggingOptions{}
	for _, opt := range opts {
		opt(options)
	}
	rtr.Use(&loggingMiddleware{options: options})
}

// ---- Middleware ----

// loggingMiddleware provides structured HTTP request/response logging.
type loggingMiddleware struct {
	options *LoggingOptions
}

// Invoke implements the Middleware interface, logging request details with structured logging.
func (m *loggingMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	start := time.Now()
	rec := &statusRecorder{ResponseWriter: c.Response()}
	c.SetResponse(rec)

	next(c)

	slog.InfoContext(c, "http_request",
		slog.String("method", c.Request().Method),
		slog.String("path", c.Request().URL.Path),
		slog.Int("status", rec.Status),
		slog.String("remote", c.Request().RemoteAddr),
		slog.String("user_agent", c.Request().UserAgent()),
		slog.Duration("duration", time.Since(start)),
	)
}

// ---- Helpers ----

// statusRecorder wraps http.ResponseWriter to capture the response status code.
type statusRecorder struct {
	http.ResponseWriter
	Status int
}

// WriteHeader captures the status code and forwards it to the underlying ResponseWriter.
func (r *statusRecorder) WriteHeader(code int) {
	r.Status = code
	r.ResponseWriter.WriteHeader(code)
}

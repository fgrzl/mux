package logging

import (
	"html"
	"log/slog"
	"net/http"
	"sync"
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
	logger  *slog.Logger
}

// Invoke implements the Middleware interface, logging request details with structured logging.
func (m *loggingMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	start := time.Now()
	rec := &statusRecorder{ResponseWriter: c.Response()}
	c.SetResponse(rec)

	next(c)

	// Use pooled attribute slice and LogAttrs to minimize allocations.
	logger := m.logger
	if logger == nil {
		logger = slog.Default()
		m.logger = logger
	}
	if !logger.Enabled(c, slog.LevelInfo) {
		return
	}

	req := c.Request()
	duration := time.Since(start)

	bufp := attrPool.Get().(*[]slog.Attr)
	attrs := (*bufp)[:0]
	// Sanitize potentially user-controlled values before logging. This helper
	// escapes HTML special chars and enforces a small maximum length to avoid
	// accidental log amplification.
	safePath := `"` + sanitizeForLog(req.URL.Path) + `"`
	safeUA := `"` + sanitizeForLog(req.UserAgent()) + `"`
	attrs = append(attrs,
		slog.String("method", req.Method),
		slog.String("path", safePath),
		slog.Int("status", rec.StatusCode()),
		slog.String("remote", req.RemoteAddr),
		slog.String("user_agent", safeUA),
		slog.Duration("duration", duration),
	)
	logger.LogAttrs(c, slog.LevelInfo, "http_request", attrs...)
	// reset and return to pool
	*bufp = attrs[:0]
	attrPool.Put(bufp)
}

// sanitizeForLog applies HTML escaping and truncates long values to a sane length.
func sanitizeForLog(s string) string {
	const max = 256
	esc := html.EscapeString(s)
	if len(esc) > max {
		return esc[:max] + "..."
	}
	return esc
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

// Write ensures status defaults to 200 if not set, then writes the body.
func (r *statusRecorder) Write(p []byte) (int, error) {
	if r.Status == 0 {
		r.Status = http.StatusOK
	}
	return r.ResponseWriter.Write(p)
}

// StatusCode returns the captured status code, defaulting to 200 if none was written.
func (r *statusRecorder) StatusCode() int {
	if r.Status == 0 {
		return http.StatusOK
	}
	return r.Status
}

// attrPool reuses []slog.Attr buffers to avoid per-request slice allocations.
var attrPool = sync.Pool{New: func() any {
	b := make([]slog.Attr, 0, 8)
	return &b
}}

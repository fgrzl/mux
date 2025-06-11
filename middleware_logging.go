package mux

import (
	"log/slog"
	"net/http"
	"time"
)

type LoggingOptions struct{}

func (rtr *Router) UseLogging(options *LoggingOptions) {
	rtr.middleware = append(rtr.middleware, &loggingMiddleware{options: options})
}

type loggingMiddleware struct {
	options *LoggingOptions
}

func (m *loggingMiddleware) Invoke(c *RouteContext, next HandlerFunc) {

	start := time.Now()
	rec := &statusRecorder{ResponseWriter: c.Response}

	c.Response = rec

	next(c)

	slog.InfoContext(c, "http_request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
		slog.Int("status", rec.Status),
		slog.String("remote", c.Request.RemoteAddr),
		slog.String("user_agent", c.Request.UserAgent()),
		slog.Duration("duration", time.Since(start)),
	)
}

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.Status = code
	r.ResponseWriter.WriteHeader(code)
}

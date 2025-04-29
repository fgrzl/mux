package mux

import (
	"log/slog"
)

type LoggingOptions struct{}

func (rtr *Router) UseLogging(options *LoggingOptions) {
	rtr.middleware = append(rtr.middleware, &loggingMiddleware{options: options})
}

type loggingMiddleware struct {
	options *LoggingOptions
}

func (m *loggingMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	slog.DebugContext(c, "request started",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)

	next(c)

	slog.DebugContext(c, "request ended",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)
}

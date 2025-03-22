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
	slog.Debug("Request started: %s %s\n", c.Request.Method, c.Request.URL.Path)
	next(c)
	slog.Debug("Request ended: %s %s\n", c.Request.Method, c.Request.URL.Path)
}

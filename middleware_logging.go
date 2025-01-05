package mux

import (
	"fmt"
)

type LoggingOptions struct{}

func (rtr *Router) UseLogging(options *LoggingOptions) {
	rtr.middleware = append(rtr.middleware, &loggingMiddleware{options: options})
}

type loggingMiddleware struct {
	options *LoggingOptions
}

func (m *loggingMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	fmt.Printf("Request started: %s %s\n", c.Request.Method, c.Request.URL.Path)
	next(c)
	fmt.Printf("Request ended: %s %s\n", c.Request.Method, c.Request.URL.Path)
}

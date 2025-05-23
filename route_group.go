package mux

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type RouteGroup struct {
	prefix       string
	authProvider AuthProvider
	registry     *RouteRegistry
}

func (rg *RouteGroup) HEAD(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodHead, pattern, handler)
}

func (rg *RouteGroup) GET(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodGet, pattern, handler)
}

func (rg *RouteGroup) POST(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodPost, pattern, handler)
}

func (rg *RouteGroup) PUT(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodPut, pattern, handler)
}

func (rg *RouteGroup) DELETE(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodDelete, pattern, handler)
}

func (rg *RouteGroup) Healthz() *RouteBuilder {
	return rg.HealthzWithReady(func() bool { return true })
}

func (rg *RouteGroup) HealthzWithReady(isReady func() bool) *RouteBuilder {
	return rg.registerRoute(http.MethodGet, "/healthz", func(c *RouteContext) {
		c.Response.Header().Set("Content-Type", "text/plain")

		if isReady() {
			c.Response.WriteHeader(http.StatusOK)
			c.Response.Write([]byte("ok"))
			return
		}

		c.Response.WriteHeader(http.StatusServiceUnavailable)
		c.Response.Write([]byte("not ready"))
	}).AllowAnonymous()
}

// StaticFallback registers a GET route that serves static files from the given directory.
// If the requested file does not exist or is a directory, the fallback file (typically index.html)
// will be served instead. This is useful for SPAs using client-side routing.
//
// Example:
//
//	r.StaticFallback("/**", "static", "static/index.html")
//
// For a subpath mount:
//
//	r.StaticFallback("/portal/**", "static", "static/index.html")
func (rg *RouteGroup) StaticFallback(pattern, dir, fallback string) *RouteBuilder {
	prefix := strings.TrimSuffix(pattern, "**")
	prefix = strings.TrimRight(prefix, "/")

	handler := func(c *RouteContext) {
		requestPath := c.Request.URL.Path
		trimmed := strings.TrimPrefix(requestPath, prefix)
		trimmed = strings.TrimPrefix(trimmed, "/")
		fullPath := filepath.Join(dir, trimmed)

		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			http.ServeFile(c.Response, c.Request, fallback)
			return
		}

		// Construct a new request with the stripped path for static serving
		r := *c.Request
		r.URL.Path = "/" + trimmed
		http.FileServer(http.Dir(dir)).ServeHTTP(c.Response, &r)
	}

	return rg.registerRoute(http.MethodGet, pattern, handler)
}

func (rg *RouteGroup) registerRoute(method string, pattern string, handler HandlerFunc) *RouteBuilder {

	pattern = normalizeRoute(pattern, rg.prefix)

	options := &RouteOptions{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	}

	rg.registry.Register(pattern, method, options)

	return &RouteBuilder{Options: options}
}

func normalizeRoute(route string, prefix string) string {
	// Remove extra slashes from prefix and route to prevent "//"
	prefix = strings.TrimRight(prefix, "/")
	route = strings.TrimLeft(route, "/")

	// Prepend the prefix if not already present
	if !strings.HasPrefix(route, prefix) {
		route = prefix + "/" + route
	}

	// Ensure route starts with a "/"
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}

	// Ensure route ends with a "/"
	if !strings.HasSuffix(route, "/") {
		route = route + "/"
	}

	// Replace multiple slashes with a single slash
	route = strings.ReplaceAll(route, "//", "/")

	return route
}

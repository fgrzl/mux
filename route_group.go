package mux

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type RouteGroup struct {
	prefix   string
	registry *RouteRegistry
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

// StaticFallback registers a GET route that serves static files from the given directory.
// If the requested file does not exist or is a directory, the fallback file (typically index.html)
// will be served instead. This is useful for SPAs using client-side routing.
//
// Example:
//
//	r.StaticFallback("/", "static", "static/index.html")
//
// For a subpath mount:
//
//	r.StaticFallback("/portal", "static", "static/index.html")
func (rg *RouteGroup) StaticFallback(pattern, dir, fallback string) *RouteBuilder {
	strip := strings.TrimRight(pattern, "/")
	fs := http.StripPrefix(strip, http.FileServer(http.Dir(dir)))

	handler := func(c *RouteContext) {
		cleanPath := path.Clean(c.Request.URL.Path)
		fullPath := filepath.Join(dir, cleanPath)

		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			http.ServeFile(c.Response, c.Request, fallback)
			return
		}

		fs.ServeHTTP(c.Response, c.Request)
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

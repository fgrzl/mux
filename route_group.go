package mux

import (
	"net/http"
	"strings"
)

type RouteGroup struct {
	prefix   string
	registry *RouteRegistry
}

func (rg *RouteGroup) GET(s string, param any) {
	panic("unimplemented")
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

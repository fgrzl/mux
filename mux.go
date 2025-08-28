package mux

import (
	"fmt"

	openapi "github.com/fgrzl/mux/internal/openapi"
	"github.com/fgrzl/mux/internal/router"
	routing "github.com/fgrzl/mux/internal/routing"
)

// Re-export core router types and constructors.
type Router = router.Router

// NewRouter creates a new Router using the internal router implementation.
func NewRouter(opts ...router.RouterOption) *Router {
	return router.NewRouter(opts...)
}

// HandlerFunc is the handler signature used by the router.
type HandlerFunc = routing.HandlerFunc

// RouteContext is the request-scoped context passed to handlers.
type RouteContext = routing.RouteContext

// RouteOptions re-exported from routing package.
type RouteOptions = routing.RouteOptions

// ServiceKey and token provider key
type ServiceKey = routing.ServiceKey

const ServiceKeyTokenProvider = routing.ServiceKeyTokenProvider

// OpenAPI types promoted for public API convenience.
type Schema = openapi.Schema
type SecurityRequirement = openapi.SecurityRequirement

// (HTTP constants such as MimeJSON and HeaderXForwarded* are defined in http_constants.go)

var (
	NewGenerator   = openapi.NewGenerator
	WithPathPrefix = openapi.WithPathPrefix
	WithExamples   = openapi.WithExamples
)

// GenerateSpecWithGenerator is a convenience wrapper that collects routes from
// the provided Router and runs the given openapi.Generator to produce a spec.
func GenerateSpecWithGenerator(gen *openapi.Generator, rtr *router.Router) (*openapi.OpenAPISpec, error) {
	if gen == nil || rtr == nil {
		return nil, fmt.Errorf("generator or router is nil")
	}

	info, err := rtr.InfoObject()
	if err != nil {
		return nil, err
	}

	routes, err := rtr.Routes()
	if err != nil {
		return nil, err
	}

	return gen.GenerateSpecFromRoutes(info, routes)
}

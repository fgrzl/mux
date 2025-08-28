package mux

import (
	"fmt"

	"github.com/fgrzl/mux/internal/router"
	routing "github.com/fgrzl/mux/internal/routing"
	"github.com/fgrzl/mux/internal/tokenizer"
	"github.com/fgrzl/mux/pkg/common"
	openapi "github.com/fgrzl/mux/pkg/openapi"
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

const ServiceKeyTokenProvider = common.ServiceKeyTokenProvider

// OpenAPI types promoted for public API convenience.
type Schema = openapi.Schema
type SecurityRequirement = openapi.SecurityRequirement

// OpenAPISpec is the exported alias for the internal OpenAPI spec structure.
// Consumers can use this type when interacting with the spec produced by
// GenerateSpecWithGenerator or the internal openapi package.
type OpenAPISpec = openapi.OpenAPISpec

// NewOpenAPISpec is a convenience constructor exposed to consumers.
var NewOpenAPISpec = openapi.NewOpenAPISpec

// ProblemDetails mirrors the public error payload used across the router and
// middleware. It is an alias to the internal common.ProblemDetails type so
// consumers can reference it when declaring response types.
type ProblemDetails = common.ProblemDetails

// DefaultProblem is the package-level default problem instance used by the
// library; exported for convenience in tests and consumers.
var DefaultProblem = common.DefaultProblem

// (HTTP constants such as MimeJSON and HeaderXForwarded* are defined in http_constants.go)

var (
	// Expose generator helpers from the public pkg/openapi package.
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

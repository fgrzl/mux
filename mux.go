package mux

import (
	"fmt"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/middleware/authentication"
	"github.com/fgrzl/mux/pkg/middleware/authorization"
	"github.com/fgrzl/mux/pkg/middleware/compression"
	"github.com/fgrzl/mux/pkg/middleware/enforcehttps"
	"github.com/fgrzl/mux/pkg/middleware/exportcontrol"
	"github.com/fgrzl/mux/pkg/middleware/forwardheaders"
	"github.com/fgrzl/mux/pkg/middleware/logging"
	"github.com/fgrzl/mux/pkg/middleware/opentelemetry"
	"github.com/fgrzl/mux/pkg/middleware/ratelimit"
	"github.com/fgrzl/mux/pkg/middleware/servicelocator"
	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/pkg/tokenizer"
)

// Router is the underlying router implementation used by this package.
//
// Router is re-exported so consumers can construct and use routers without
// importing the internal router package directly.
type Router = router.Router

// NewRouter is an alias to the internal router constructor.
//
// Using a variable alias keeps the public API as a direct reference to the
// underlying implementation and matches the project's preferred style.
var NewRouter = router.NewRouter

// HandlerFunc is the handler signature used by the router.
type HandlerFunc = routing.HandlerFunc

// RouteContext is the request-scoped context passed to handlers.
type RouteContext = routing.RouteContext

// RouteOptions is the per-route configuration type.
type RouteOptions = routing.RouteOptions

// ServiceKey identifies named services stored on a RouteContext.
type ServiceKey = routing.ServiceKey

// ServiceKeyTokenProvider is the service key used for the token provider.
const ServiceKeyTokenProvider = tokenizer.ServiceKeyTokenProvider

// Schema is an OpenAPI schema alias for convenience when generating specs.
type Schema = openapi.Schema

// SecurityRequirement mirrors openapi.SecurityRequirement.
type SecurityRequirement = openapi.SecurityRequirement

// OpenAPISpec is the OpenAPI specification produced by the generator.
type OpenAPISpec = openapi.OpenAPISpec

// NewOpenAPISpec constructs a new OpenAPISpec using the openapi package.
var NewOpenAPISpec = openapi.NewOpenAPISpec

// ProblemDetails is the canonical problem details payload used across the
// router and middleware. It is re-exported for convenience.
type ProblemDetails = common.ProblemDetails

// DefaultProblem is the package-level default problem instance used by the
// library; exported for convenience in tests and consumers.
var DefaultProblem = common.DefaultProblem

// NewGenerator is a variable that references the openapi.NewGenerator function.
// It can be used to create a new OpenAPI generator instance via this package.
// By exposing the constructor as a package-level variable, callers can replace
// it (for example in tests) with an alternative implementation or mock.
var NewGenerator = openapi.NewGenerator

// WithPathPrefix returns an option that sets a path prefix applied to all generated
// OpenAPI operation paths. It is an alias of openapi.WithPathPrefix, provided so
// callers of this package can configure the prefix without importing the openapi package.
var WithPathPrefix = openapi.WithPathPrefix

// WithExamples is an option that attaches example values to generated OpenAPI
// schema elements. It is an alias for openapi.WithExamples and provides the
// same behavior and accepted value forms as that symbol. Use WithExamples when
// you want to include illustrative examples in the OpenAPI documentation emitted
// by this package's generation helpers.
var WithExamples = openapi.WithExamples

// Re-export router options so callers can use them via the top-level mux package.
var WithTitle = router.WithTitle
var WithDescription = router.WithDescription
var WithVersion = router.WithVersion

// GenerateSpecWithGenerator collects routes from rtr and uses gen to produce
// an OpenAPI spec. It returns an error if either argument is nil or if the
// router fails to provide route information.
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

// LoggingOptions is an alias of logging.LoggingOptions.
type LoggingOptions = logging.LoggingOptions

// LoggingOption is an alias of logging.LoggingOption.
type LoggingOption = logging.LoggingOption

// UseLogging is an alias to logging.UseLogging.
var UseLogging = logging.UseLogging

// AuthOption is an alias of authentication.AuthOption.
type AuthOption = authentication.AuthOption

// AuthenticationOptions is an alias of authentication.AuthenticationOptions.
type AuthenticationOptions = authentication.AuthenticationOptions

// UseAuthentication is an alias to authentication.UseAuthentication.
var UseAuthentication = authentication.UseAuthentication

// UseAuthenticationWithProvider is an alias to authentication.UseAuthenticationWithProvider.
var UseAuthenticationWithProvider = authentication.UseAuthenticationWithProvider

// AuthZOption is an alias of authorization.AuthZOption.
type AuthZOption = authorization.AuthZOption

// AuthorizationOptions is an alias of authorization.AuthorizationOptions.
type AuthorizationOptions = authorization.AuthorizationOptions

// UseAuthorization is an alias to authorization.UseAuthorization.
var UseAuthorization = authorization.UseAuthorization

// CompressionOptions is an alias of compression.CompressionOptions.
type CompressionOptions = compression.CompressionOptions

// CompressionOption is an alias of compression.CompressionOption.
type CompressionOption = compression.CompressionOption

// UseCompression is an alias to compression.UseCompression.
var UseCompression = compression.UseCompression

// UseEnforceHTTPS is an alias to enforcehttps.UseEnforceHTTPS.
var UseEnforceHTTPS = enforcehttps.UseEnforceHTTPS

// ExportControlOptions is an alias of exportcontrol.ExportControlOptions.
type ExportControlOptions = exportcontrol.ExportControlOptions

// ExportControlOption is an alias of exportcontrol.ExportControlOption.
type ExportControlOption = exportcontrol.ExportControlOption

// UseExportControl is an alias to exportcontrol.UseExportControl.
var UseExportControl = exportcontrol.UseExportControl

// UseForwardedHeaders is an alias to forwardheaders.UseForwardedHeaders.
var UseForwardedHeaders = forwardheaders.UseForwardedHeaders

// OpenTelemetryOptions is an alias of opentelemetry.OpenTelemetryOptions.
type OpenTelemetryOptions = opentelemetry.OpenTelemetryOptions

// OpenTelemetryOption is an alias of opentelemetry.OpenTelemetryOption.
type OpenTelemetryOption = opentelemetry.OpenTelemetryOption

// UseOpenTelemetry is an alias to opentelemetry.UseOpenTelemetry.
var UseOpenTelemetry = opentelemetry.UseOpenTelemetry

// RateLimiterOptions is an alias of ratelimit.RateLimiterOptions.
type RateLimiterOptions = ratelimit.RateLimiterOptions

// RateLimiterOption is an alias of ratelimit.RateLimiterOption.
type RateLimiterOption = ratelimit.RateLimiterOption

// UseRateLimiter is an alias to ratelimit.UseRateLimiter.
var UseRateLimiter = ratelimit.UseRateLimiter

// ServiceSetterOptions is an alias of servicelocator.ServiceSetterOptions.
type ServiceSetterOptions = servicelocator.ServiceSetterOptions

// ServiceSetterOption is an alias of servicelocator.ServiceSetterOption.
type ServiceSetterOption = servicelocator.ServiceSetterOption

// UseServices is an alias to servicelocator.UseServices.
var UseServices = servicelocator.UseServices

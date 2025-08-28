package mux

import (
	"fmt"

	"github.com/fgrzl/mux/pkg/builder"
	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/cookiejar"
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

// RouteGroup is a type alias for the package's underlying route-group implementation.
// It represents a collection of routes that share common configuration such as a
// URL path prefix and middleware. RouteGroup provides methods to register handlers,
// compose middleware, and create nested subgroups to structure related routes.
// This alias exists to maintain API stability and allow the concrete implementation
// to be refactored without requiring changes from callers.
type RouteGroup = router.RouteGroup

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

// --- Common MIME types ---
// Re-export common MIME type constants for convenience.
const (
	MimeJSON              = common.MimeJSON
	MimeXML               = common.MimeXML
	MimeFormURLEncoded    = common.MimeFormURLEncoded
	MimeMultipartFormData = common.MimeMultipartFormData
	MimeTextPlain         = common.MimeTextPlain
	MimeTextHTML          = common.MimeTextHTML
	MimeTextCSV           = common.MimeTextCSV
	MimeOctetStream       = common.MimeOctetStream
	MimePDF               = common.MimePDF
	MimeZIP               = common.MimeZIP
	MimePNG               = common.MimePNG
	MimeJPEG              = common.MimeJPEG
	MimeGIF               = common.MimeGIF
	MimeSVG               = common.MimeSVG
	MimeWebP              = common.MimeWebP
	MimeMP4               = common.MimeMP4
	MimeMP3               = common.MimeMP3
	MimeWAV               = common.MimeWAV
	MimeOGG               = common.MimeOGG
	MimeJSONAPI           = common.MimeJSONAPI
	MimeOpenAPI           = common.MimeOpenAPI
	MimeYAML              = common.MimeYAML
	MimeProblemJSON       = common.MimeProblemJSON
)

// --- Common HTTP headers ---
const (
	HeaderContentType        = common.HeaderContentType
	HeaderContentDisposition = common.HeaderContentDisposition
	HeaderXForwardedFor      = common.HeaderXForwardedFor
	HeaderXForwardedProto    = common.HeaderXForwardedProto
	HeaderXRealIP            = common.HeaderXRealIP
	HeaderAcceptEncoding     = common.HeaderAcceptEncoding
	HeaderContentEncoding    = common.HeaderContentEncoding
	HeaderUserAgent          = common.HeaderUserAgent
	HeaderAuthorization      = common.HeaderAuthorization
	HeaderLocation           = common.HeaderLocation
	HeaderSetCookie          = common.HeaderSetCookie
	HeaderCookie             = common.HeaderCookie
	HeaderAccept             = common.HeaderAccept
	HeaderRetryAfter         = common.HeaderRetryAfter
	HeaderCacheControl       = common.HeaderCacheControl
	HeaderETag               = common.HeaderETag
	HeaderContentLength      = common.HeaderContentLength
	HeaderTransferEncoding   = common.HeaderTransferEncoding
)

// --- Cookie defaults and accessors ---
// DefaultUserCookieName is the default app session cookie name. Alias of
// cookiejar.DefaultUserCookieName.
const DefaultUserCookieName = cookiejar.DefaultUserCookieName

// DefaultTwoFactorCookieName is the default two-factor cookie name. Alias of
// cookiejar.DefaultTwoFactorCookieName.
const DefaultTwoFactorCookieName = cookiejar.DefaultTwoFactorCookieName

// DefaultIdpUserCookieName is the default identity-provider session cookie
// name. Alias of cookiejar.DefaultIdpUserCookieName.
const DefaultIdpUserCookieName = cookiejar.DefaultIdpUserCookieName

// GetUserCookieName returns the current application session cookie name.
var GetUserCookieName = cookiejar.GetUserCookieName

// SetAppSessionCookieName sets the application session cookie name.
var SetAppSessionCookieName = cookiejar.SetAppSessionCookieName

// GetTwoFactorCookieName returns the current two-factor authentication cookie name.
var GetTwoFactorCookieName = cookiejar.GetTwoFactorCookieName

// SetTwoFactorCookieName sets the two-factor authentication cookie name.
var SetTwoFactorCookieName = cookiejar.SetTwoFactorCookieName

// GetIdpSessionCookieName returns the current identity-provider session cookie name.
var GetIdpSessionCookieName = cookiejar.GetIdpSessionCookieName

// SetIdpSessionCookieName sets the identity-provider session cookie name.
var SetIdpSessionCookieName = cookiejar.SetIdpSessionCookieName

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

// WithTitle sets the API title on the underlying router. It is an alias of
// router.WithTitle so callers can configure the router via this package.
var WithTitle = router.WithTitle

// WithDescription sets the API description on the underlying router. It is an
// alias of router.WithDescription so callers can configure the router via
// this package.
var WithDescription = router.WithDescription

// WithVersion sets the API version on the underlying router. It is an alias
// of router.WithVersion so callers can configure the router via this
// package.
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

// --- Logging middleware ---
// LoggingOptions configures the logging middleware. Alias of
// logging.LoggingOptions.
type LoggingOptions = logging.LoggingOptions

// LoggingOption is a single functional option for logging middleware.
type LoggingOption = logging.LoggingOption

// UseLogging installs the logging middleware on a router. Alias of
// logging.UseLogging.
var UseLogging = logging.UseLogging

// --- Authentication middleware ---
// AuthOption is a functional option used to configure the authentication
// middleware. Alias of authentication.AuthOption.
type AuthOption = authentication.AuthOption

// AuthenticationOptions holds configuration for the authentication
// middleware. Alias of authentication.AuthenticationOptions.
type AuthenticationOptions = authentication.AuthenticationOptions

// UseAuthentication installs the authentication middleware on a router.
// Alias of authentication.UseAuthentication.
var UseAuthentication = authentication.UseAuthentication

// UseAuthenticationWithProvider installs authentication middleware using a
// specific token provider. Alias of
// authentication.UseAuthenticationWithProvider.
var UseAuthenticationWithProvider = authentication.UseAuthenticationWithProvider

// WithTokenTTL sets the token time-to-live duration for the authentication
// middleware. Alias of authentication.WithTokenTTL.
var WithTokenTTL = authentication.WithTokenTTL

// WithValidator sets the token validator function used by the
// authentication middleware. Alias of authentication.WithValidator.
var WithValidator = authentication.WithValidator

// WithTokenCreator sets the token creation function used by the
// authentication middleware. Alias of authentication.WithTokenCreator.
var WithTokenCreator = authentication.WithTokenCreator

// --- Authorization middleware ---
// AuthZOption is a functional option used to configure the authorization
// middleware. Alias of authorization.AuthZOption.
type AuthZOption = authorization.AuthZOption

// AuthorizationOptions holds configuration for the authorization
// middleware. Alias of authorization.AuthorizationOptions.
type AuthorizationOptions = authorization.AuthorizationOptions

// UseAuthorization installs the authorization middleware on a router.
// Alias of authorization.UseAuthorization.
var UseAuthorization = authorization.UseAuthorization

// WithRoles configures required roles for the authorization middleware.
// Alias of authorization.WithRoles.
var WithRoles = authorization.WithRoles

// WithScopes configures required scopes for the authorization middleware.
// Alias of authorization.WithScopes.
var WithScopes = authorization.WithScopes

// WithPermissions configures required permissions for the authorization
// middleware. Alias of authorization.WithPermissions.
var WithPermissions = authorization.WithPermissions

// WithRoleChecker sets a custom role-checking function for authorization.
// Alias of authorization.WithRoleChecker.
var WithRoleChecker = authorization.WithRoleChecker

// WithScopeChecker sets a custom scope-checking function for authorization.
// Alias of authorization.WithScopeChecker.
var WithScopeChecker = authorization.WithScopeChecker

// WithPermissionChecker sets a custom permission-checking function for
// authorization. Alias of authorization.WithPermissionChecker.
var WithPermissionChecker = authorization.WithPermissionChecker

// --- Compression middleware ---
// CompressionOptions configures the compression middleware. Alias of
// compression.CompressionOptions.
type CompressionOptions = compression.CompressionOptions

// CompressionOption is a single functional option for compression middleware.
type CompressionOption = compression.CompressionOption

// UseCompression installs the compression middleware on a router. Alias of
// compression.UseCompression.
var UseCompression = compression.UseCompression

// --- Enforce HTTPS middleware ---
// UseEnforceHTTPS installs middleware that redirects HTTP to HTTPS. Alias of
// enforcehttps.UseEnforceHTTPS.
var UseEnforceHTTPS = enforcehttps.UseEnforceHTTPS

// --- Export-control middleware ---
// ExportControlOptions configures the export-control middleware. Alias of
// exportcontrol.ExportControlOptions.
type ExportControlOptions = exportcontrol.ExportControlOptions

// ExportControlOption is a single functional option for export-control.
type ExportControlOption = exportcontrol.ExportControlOption

// UseExportControl installs the export-control middleware on a router.
// Alias of exportcontrol.UseExportControl.
var UseExportControl = exportcontrol.UseExportControl

// WithGeoIPDatabase supplies a geoip2 reader to the export-control middleware
// for country lookups. Alias of exportcontrol.WithGeoIPDatabase.
var WithGeoIPDatabase = exportcontrol.WithGeoIPDatabase

// --- Forwarded headers middleware ---
// UseForwardedHeaders installs the middleware which handles forwarded
// headers (X-Forwarded-*) on a router. Alias of
// forwardheaders.UseForwardedHeaders.
var UseForwardedHeaders = forwardheaders.UseForwardedHeaders

// --- OpenTelemetry middleware ---
// OpenTelemetryOptions configures the OpenTelemetry middleware. Alias of
// opentelemetry.OpenTelemetryOptions.
type OpenTelemetryOptions = opentelemetry.OpenTelemetryOptions

// OpenTelemetryOption is a single functional option for OpenTelemetry.
type OpenTelemetryOption = opentelemetry.OpenTelemetryOption

// UseOpenTelemetry installs the OpenTelemetry middleware on a router.
// Alias of opentelemetry.UseOpenTelemetry.
var UseOpenTelemetry = opentelemetry.UseOpenTelemetry

// WithOperation sets the operation name used by the OpenTelemetry middleware
// when creating spans. Alias of opentelemetry.WithOperation.
var WithOperation = opentelemetry.WithOperation

// --- Rate-limiter middleware ---
// RateLimiterOptions configures the rate limiter middleware. Alias of
// ratelimit.RateLimiterOptions.
type RateLimiterOptions = ratelimit.RateLimiterOptions

// RateLimiterOption is a single functional option for the rate limiter.
type RateLimiterOption = ratelimit.RateLimiterOption

// UseRateLimiter installs the rate-limiter middleware on a router. Alias of
// ratelimit.UseRateLimiter.
var UseRateLimiter = ratelimit.UseRateLimiter

// WithCleanupInterval configures the cleanup interval used by the
// rate-limiter middleware. Alias of ratelimit.WithCleanupInterval.
var WithCleanupInterval = ratelimit.WithCleanupInterval

// --- Service locator middleware ---
// ServiceSetterOptions configures the service-locator middleware. Alias of
// servicelocator.ServiceSetterOptions.
type ServiceSetterOptions = servicelocator.ServiceSetterOptions

// ServiceSetterOption is a single functional option for the service locator.
type ServiceSetterOption = servicelocator.ServiceSetterOption

// UseServices installs the service locator middleware on a router. Alias of
// servicelocator.UseServices.
var UseServices = servicelocator.UseServices

// WithService adds a service instance to be injected by the service
// locator middleware. Alias of servicelocator.WithService.
var WithService = servicelocator.WithService

// --- Builder helpers ---
// RouteBuilder provides a fluent interface for configuring routes. Alias of
// builder.RouteBuilder.
type RouteBuilder = builder.RouteBuilder

// Route bootstraps a new RouteBuilder. Alias of builder.Route.
var Route = builder.Route

// --- Router options and helpers ---
// RouterOptions configures the top-level router. Alias of
// router.RouterOptions.
type RouterOptions = router.RouterOptions

// RouterOption is the functional option type used to configure the router.
type RouterOption = router.RouterOption

// WithClientURL sets the client URL on the router. Alias of
// router.WithClientURL.
var WithClientURL = router.WithClientURL

// WithSummary sets the API summary on the underlying router. Alias of
// router.WithSummary.
var WithSummary = router.WithSummary

// WithTermsOfService sets the terms-of-service URL on the router. Alias of
// router.WithTermsOfService.
var WithTermsOfService = router.WithTermsOfService

// WithContact sets the contact information for the API. Alias of
// router.WithContact.
var WithContact = router.WithContact

// WithLicense sets the license information for the API. Alias of
// router.WithLicense.
var WithLicense = router.WithLicense

// Middleware is the interface used by router middleware implementations.
// Alias of router.Middleware.
type Middleware = router.Middleware

// --- OpenAPI generator types ---
// Generator represents the OpenAPI generator. Alias of openapi.Generator.
type Generator = openapi.Generator

// GeneratorOption is a functional option for configuring Generator. Alias
// of openapi.GeneratorOption.
type GeneratorOption = openapi.GeneratorOption

// --- Tokenizer types ---
// TokenProvider defines the minimal interface for creating and validating
// tokens. Alias of tokenizer.TokenProvider.
type TokenProvider = tokenizer.TokenProvider

package mux

import (
	"time"

	internalbuilder "github.com/fgrzl/mux/internal/builder"
	internalcommon "github.com/fgrzl/mux/internal/common"
)

// RouteBuilder decorates a registered route with middleware, auth
// requirements, scoped services, and OpenAPI metadata. Methods are chainable
// and mutate the registered route in place.
type RouteBuilder struct {
	inner *internalbuilder.RouteBuilder
}

func wrapRouteBuilder(inner *internalbuilder.RouteBuilder) *RouteBuilder {
	if inner == nil {
		return nil
	}
	return &RouteBuilder{inner: inner}
}

// AllowAnonymous clears inherited authentication requirements for this route.
// Use it for public endpoints such as login, callbacks, webhooks, and probes.
func (b *RouteBuilder) AllowAnonymous() *RouteBuilder {
	b.inner.AllowAnonymous()
	return b
}

// Use attaches middleware that runs only for this route after router and group
// middleware.
func (b *RouteBuilder) Use(middleware ...Middleware) *RouteBuilder {
	b.inner.Use(toInternalMiddlewares(middleware)...)
	return b
}

// Services returns the route-scoped service registry. Services registered here
// are available to the handler and override parent registrations for the same
// key. Prefer it when registering more than one dependency.
func (b *RouteBuilder) Services() *ServiceRegistry {
	inner := b.inner.Services()
	return newServiceRegistry(
		func(key ServiceKey, svc any) { inner.Register(internalcommon.ServiceKey(key), svc) },
		func(key ServiceKey) (any, bool) { return inner.Get(internalcommon.ServiceKey(key)) },
	)
}

// Service registers a service that is visible only to this route and returns
// the builder for chaining. It is the singular convenience form of
// Services().Register(...), which is the preferred API when multiple services
// need to be registered together.
func (b *RouteBuilder) Service(key ServiceKey, svc any) *RouteBuilder {
	b.inner.WithService(internalcommon.ServiceKey(key), svc)
	return b
}

// RequirePermission marks the route as requiring the provided permissions when
// authorization middleware is enabled.
func (b *RouteBuilder) RequirePermission(perms ...string) *RouteBuilder {
	b.inner.RequirePermission(perms...)
	return b
}

// RequireRoles marks the route as requiring the provided roles when
// authorization middleware is enabled.
func (b *RouteBuilder) RequireRoles(roles ...string) *RouteBuilder {
	b.inner.RequireRoles(roles...)
	return b
}

// RequireScopes marks the route as requiring the provided scopes when
// authorization middleware is enabled.
func (b *RouteBuilder) RequireScopes(scopes ...string) *RouteBuilder {
	b.inner.RequireScopes(scopes...)
	return b
}

// WithRateLimit applies a route-local rate limit. Prefer shared middleware when
// multiple routes should draw from the same budget.
func (b *RouteBuilder) WithRateLimit(limit int, interval time.Duration) *RouteBuilder {
	b.inner.WithRateLimit(limit, interval)
	return b
}

// WithOperationID sets a stable, unique OpenAPI operationId for this route.
// Provide one for every documented route so generators and AI tooling can
// refer to the operation consistently.
func (b *RouteBuilder) WithOperationID(id string) *RouteBuilder {
	b.inner.WithOperationID(id)
	return b
}

// WithSummary sets the one-line OpenAPI summary for this route.
func (b *RouteBuilder) WithSummary(summary string) *RouteBuilder {
	b.inner.WithSummary(summary)
	return b
}

// WithDescription sets the longer OpenAPI description for this route.
func (b *RouteBuilder) WithDescription(description string) *RouteBuilder {
	b.inner.WithDescription(description)
	return b
}

// WithTags groups the route under the provided OpenAPI tags. Reuse the same
// tag strings across related endpoints for cleaner generated docs and clients.
func (b *RouteBuilder) WithTags(tags ...string) *RouteBuilder {
	b.inner.WithTags(tags...)
	return b
}

// WithExternalDocs links the route to additional documentation in generated
// OpenAPI output.
func (b *RouteBuilder) WithExternalDocs(url, description string) *RouteBuilder {
	b.inner.WithExternalDocs(url, description)
	return b
}

// WithSecurity declares OpenAPI security requirements for this route. This
// documents required schemes and scopes; pair it with auth middleware for
// runtime enforcement.
func (b *RouteBuilder) WithSecurity(sec SecurityRequirement) *RouteBuilder {
	b.inner.WithSecurity(toInternalSecurityRequirement(sec))
	return b
}

// WithDeprecated marks the route as deprecated in generated OpenAPI output.
func (b *RouteBuilder) WithDeprecated() *RouteBuilder {
	b.inner.WithDeprecated()
	return b
}

// WithPathParam documents a path parameter and marks it required. The name
// should match the placeholder used in the route pattern.
func (b *RouteBuilder) WithPathParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "path", description, example, true)
}

// WithQueryParam documents an optional query parameter for this route.
func (b *RouteBuilder) WithQueryParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "query", description, example, false)
}

// WithRequiredQueryParam documents a required query parameter for this route.
func (b *RouteBuilder) WithRequiredQueryParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "query", description, example, true)
}

// WithHeaderParam documents an optional header parameter for this route.
func (b *RouteBuilder) WithHeaderParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "header", description, example, false)
}

// WithRequiredHeaderParam documents a required header parameter for this
// route.
func (b *RouteBuilder) WithRequiredHeaderParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "header", description, example, true)
}

// WithCookieParam documents an optional cookie parameter for this route.
func (b *RouteBuilder) WithCookieParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "cookie", description, example, false)
}

// WithRequiredCookieParam documents a required cookie parameter for this
// route.
func (b *RouteBuilder) WithRequiredCookieParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "cookie", description, example, true)
}

// WithJsonBody documents a required application/json request body using
// example to infer both schema and example payload. Prefer concrete structs
// over generic maps when you want stable generated schemas.
func (b *RouteBuilder) WithJsonBody(example any) *RouteBuilder {
	b.inner.WithJsonBody(example)
	return b
}

// WithOneOfJsonBody documents an application/json request body whose schema
// may match any one of the supplied examples.
func (b *RouteBuilder) WithOneOfJsonBody(examples ...any) *RouteBuilder {
	b.inner.WithOneOfJsonBody(examples...)
	return b
}

// WithAnyOfJsonBody documents an application/json request body whose schema
// may include any combination of the supplied examples.
func (b *RouteBuilder) WithAnyOfJsonBody(examples ...any) *RouteBuilder {
	b.inner.WithAnyOfJsonBody(examples...)
	return b
}

// WithAllOfJsonBody documents an application/json request body that must
// satisfy all of the supplied example schemas.
func (b *RouteBuilder) WithAllOfJsonBody(examples ...any) *RouteBuilder {
	b.inner.WithAllOfJsonBody(examples...)
	return b
}

// WithFormBody documents an application/x-www-form-urlencoded request body
// from the provided example.
func (b *RouteBuilder) WithFormBody(example any) *RouteBuilder {
	b.inner.WithFormBody(example)
	return b
}

// WithMultipartBody documents a multipart/form-data request body from the
// provided example.
func (b *RouteBuilder) WithMultipartBody(example any) *RouteBuilder {
	b.inner.WithMultipartBody(example)
	return b
}

// WithResponse documents a response for the given status code. When example is
// nil the response is documented without a body; otherwise the example drives
// schema inference and example generation. Prefer a named helper when one
// exists so generated APIs read like HTTP semantics.
func (b *RouteBuilder) WithResponse(code int, example any) *RouteBuilder {
	b.inner.WithResponse(code, example)
	return b
}

// WithOKResponse documents a 200 OK response for this route.
func (b *RouteBuilder) WithOKResponse(example any) *RouteBuilder {
	b.inner.WithOKResponse(example)
	return b
}

// WithCreatedResponse documents a 201 Created response for this route.
func (b *RouteBuilder) WithCreatedResponse(example any) *RouteBuilder {
	b.inner.WithCreatedResponse(example)
	return b
}

// WithAcceptedResponse documents a 202 Accepted response for this route.
func (b *RouteBuilder) WithAcceptedResponse(example any) *RouteBuilder {
	b.inner.WithAcceptedResponse(example)
	return b
}

// WithNoContentResponse documents a 204 No Content response for this route.
func (b *RouteBuilder) WithNoContentResponse() *RouteBuilder {
	b.inner.WithNoContentResponse()
	return b
}

// WithNotFoundResponse documents a 404 Not Found response for this route.
func (b *RouteBuilder) WithNotFoundResponse() *RouteBuilder {
	b.inner.WithNotFoundResponse()
	return b
}

// WithConflictResponse documents a 409 Conflict response for this route. Use
// it when writes can fail because of duplicate, stale, or incompatible state.
func (b *RouteBuilder) WithConflictResponse() *RouteBuilder {
	b.inner.WithConflictResponse()
	return b
}

// WithBadRequestResponse documents a 400 Bad Request response for validation
// and parsing failures.
func (b *RouteBuilder) WithBadRequestResponse() *RouteBuilder {
	b.inner.WithBadRequestResponse()
	return b
}

// WithUnauthorizedResponse documents a 401 Unauthorized response. It does not
// install authentication middleware by itself.
func (b *RouteBuilder) WithUnauthorizedResponse() *RouteBuilder {
	b.inner.WithUnauthorizedResponse()
	return b
}

// WithForbiddenResponse documents a 403 Forbidden response. It does not
// install authorization middleware by itself.
func (b *RouteBuilder) WithForbiddenResponse() *RouteBuilder {
	b.inner.WithForbiddenResponse()
	return b
}

// WithStandardErrors documents the default client-error set used by mux route
// helpers: 400 and 404. Add auth or conflict helpers separately when the route
// can also return 401, 403, or 409.
func (b *RouteBuilder) WithStandardErrors() *RouteBuilder {
	b.inner.WithStandardErrors()
	return b
}

// WithMovedPermanentlyResponse documents a 301 Moved Permanently redirect with
// no response body.
func (b *RouteBuilder) WithMovedPermanentlyResponse() *RouteBuilder {
	b.inner.WithMovedPermanentlyResponse()
	return b
}

// WithFoundResponse documents a 302 Found redirect with no response body.
func (b *RouteBuilder) WithFoundResponse() *RouteBuilder {
	b.inner.WithFoundResponse()
	return b
}

// WithSeeOtherResponse documents a 303 See Other redirect with no response
// body. Use it for POST-Redirect-GET flows.
func (b *RouteBuilder) WithSeeOtherResponse() *RouteBuilder {
	b.inner.WithSeeOtherResponse()
	return b
}

// WithTemporaryRedirectResponse documents a 307 Temporary Redirect with no
// response body while preserving the original HTTP method.
func (b *RouteBuilder) WithTemporaryRedirectResponse() *RouteBuilder {
	b.inner.WithTemporaryRedirectResponse()
	return b
}

// WithPermanentRedirectResponse documents a 308 Permanent Redirect with no
// response body while preserving the original HTTP method.
func (b *RouteBuilder) WithPermanentRedirectResponse() *RouteBuilder {
	b.inner.WithPermanentRedirectResponse()
	return b
}

func (b *RouteBuilder) addRouteParam(name, in, description string, example any, required bool) *RouteBuilder {
	b.inner.WithParam(name, in, description, example, required)
	return b
}

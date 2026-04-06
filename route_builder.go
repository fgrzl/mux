package mux

import (
	"time"

	internalbuilder "github.com/fgrzl/mux/internal/builder"
	internalcommon "github.com/fgrzl/mux/internal/common"
)

type RouteBuilder struct {
	inner *internalbuilder.RouteBuilder
}

func wrapRouteBuilder(inner *internalbuilder.RouteBuilder) *RouteBuilder {
	if inner == nil {
		return nil
	}
	return &RouteBuilder{inner: inner}
}

func (b *RouteBuilder) AllowAnonymous() *RouteBuilder {
	b.inner.AllowAnonymous()
	return b
}

func (b *RouteBuilder) Use(middleware ...Middleware) *RouteBuilder {
	b.inner.Use(toInternalMiddlewares(middleware)...)
	return b
}

func (b *RouteBuilder) Services() *ServiceRegistry {
	inner := b.inner.Services()
	return newServiceRegistry(
		func(key ServiceKey, svc any) { inner.Register(internalcommon.ServiceKey(key), svc) },
		func(key ServiceKey) (any, bool) { return inner.Get(internalcommon.ServiceKey(key)) },
	)
}

func (b *RouteBuilder) Service(key ServiceKey, svc any) *RouteBuilder {
	b.inner.WithService(internalcommon.ServiceKey(key), svc)
	return b
}

func (b *RouteBuilder) RequirePermission(perms ...string) *RouteBuilder {
	b.inner.RequirePermission(perms...)
	return b
}

func (b *RouteBuilder) RequireRoles(roles ...string) *RouteBuilder {
	b.inner.RequireRoles(roles...)
	return b
}

func (b *RouteBuilder) RequireScopes(scopes ...string) *RouteBuilder {
	b.inner.RequireScopes(scopes...)
	return b
}

func (b *RouteBuilder) WithRateLimit(limit int, interval time.Duration) *RouteBuilder {
	b.inner.WithRateLimit(limit, interval)
	return b
}

func (b *RouteBuilder) WithOperationID(id string) *RouteBuilder {
	b.inner.WithOperationID(id)
	return b
}

func (b *RouteBuilder) WithSummary(summary string) *RouteBuilder {
	b.inner.WithSummary(summary)
	return b
}

func (b *RouteBuilder) WithDescription(description string) *RouteBuilder {
	b.inner.WithDescription(description)
	return b
}

func (b *RouteBuilder) WithTags(tags ...string) *RouteBuilder {
	b.inner.WithTags(tags...)
	return b
}

func (b *RouteBuilder) WithExternalDocs(url, description string) *RouteBuilder {
	b.inner.WithExternalDocs(url, description)
	return b
}

func (b *RouteBuilder) WithSecurity(sec SecurityRequirement) *RouteBuilder {
	b.inner.WithSecurity(toInternalSecurityRequirement(sec))
	return b
}

func (b *RouteBuilder) WithDeprecated() *RouteBuilder {
	b.inner.WithDeprecated()
	return b
}

func (b *RouteBuilder) WithPathParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "path", description, example, true)
}

func (b *RouteBuilder) WithQueryParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "query", description, example, false)
}

func (b *RouteBuilder) WithRequiredQueryParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "query", description, example, true)
}

func (b *RouteBuilder) WithHeaderParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "header", description, example, false)
}

func (b *RouteBuilder) WithRequiredHeaderParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "header", description, example, true)
}

func (b *RouteBuilder) WithCookieParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "cookie", description, example, false)
}

func (b *RouteBuilder) WithRequiredCookieParam(name, description string, example any) *RouteBuilder {
	return b.addRouteParam(name, "cookie", description, example, true)
}

func (b *RouteBuilder) WithJsonBody(example any) *RouteBuilder {
	b.inner.WithJsonBody(example)
	return b
}

func (b *RouteBuilder) WithOneOfJsonBody(examples ...any) *RouteBuilder {
	b.inner.WithOneOfJsonBody(examples...)
	return b
}

func (b *RouteBuilder) WithAnyOfJsonBody(examples ...any) *RouteBuilder {
	b.inner.WithAnyOfJsonBody(examples...)
	return b
}

func (b *RouteBuilder) WithAllOfJsonBody(examples ...any) *RouteBuilder {
	b.inner.WithAllOfJsonBody(examples...)
	return b
}

func (b *RouteBuilder) WithFormBody(example any) *RouteBuilder {
	b.inner.WithFormBody(example)
	return b
}

func (b *RouteBuilder) WithMultipartBody(example any) *RouteBuilder {
	b.inner.WithMultipartBody(example)
	return b
}

func (b *RouteBuilder) WithResponse(code int, example any) *RouteBuilder {
	b.inner.WithResponse(code, example)
	return b
}

func (b *RouteBuilder) WithOKResponse(example any) *RouteBuilder {
	b.inner.WithOKResponse(example)
	return b
}

func (b *RouteBuilder) WithCreatedResponse(example any) *RouteBuilder {
	b.inner.WithCreatedResponse(example)
	return b
}

func (b *RouteBuilder) WithAcceptedResponse(example any) *RouteBuilder {
	b.inner.WithAcceptedResponse(example)
	return b
}

func (b *RouteBuilder) WithNoContentResponse() *RouteBuilder {
	b.inner.WithNoContentResponse()
	return b
}

func (b *RouteBuilder) addRouteParam(name, in, description string, example any, required bool) *RouteBuilder {
	b.inner.WithParam(name, in, description, example, required)
	return b
}

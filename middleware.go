package mux

import (
	"context"
	"time"

	"github.com/fgrzl/claims"
	internalauthentication "github.com/fgrzl/mux/internal/middleware/authentication"
	internalauthorization "github.com/fgrzl/mux/internal/middleware/authorization"
	internalcompression "github.com/fgrzl/mux/internal/middleware/compression"
	internalcors "github.com/fgrzl/mux/internal/middleware/cors"
	internalenforcehttps "github.com/fgrzl/mux/internal/middleware/enforcehttps"
	internalexportcontrol "github.com/fgrzl/mux/internal/middleware/exportcontrol"
	internalforwardheaders "github.com/fgrzl/mux/internal/middleware/forwardheaders"
	internallogging "github.com/fgrzl/mux/internal/middleware/logging"
	internalopentelemetry "github.com/fgrzl/mux/internal/middleware/opentelemetry"
	internalratelimit "github.com/fgrzl/mux/internal/middleware/ratelimit"
	internalopenapi "github.com/fgrzl/mux/internal/openapi"
	internalrouting "github.com/fgrzl/mux/internal/routing"
	"github.com/oschwald/geoip2-golang"
)

type GeneratorOption struct {
	apply internalopenapi.GeneratorOption
}

func WithOpenAPIExamples() GeneratorOption {
	return GeneratorOption{apply: internalopenapi.WithExamples()}
}

func WithOpenAPIPathPrefix(prefix string) GeneratorOption {
	return GeneratorOption{apply: internalopenapi.WithPathPrefix(prefix)}
}

type Generator struct {
	inner *internalopenapi.Generator
}

func NewGenerator(opts ...GeneratorOption) *Generator {
	internalOpts := make([]internalopenapi.GeneratorOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	return &Generator{inner: internalopenapi.NewGenerator(internalOpts...)}
}

func UseLogging(rtr *Router) {
	internallogging.UseLogging(rtr.inner)
}

func UseCompression(rtr *Router) {
	internalcompression.UseCompression(rtr.inner)
}

type CORSOption struct {
	apply internalcors.CORSOption
}

func WithCORSAllowedOrigins(origins ...string) CORSOption {
	return CORSOption{apply: internalcors.WithAllowedOrigins(origins...)}
}

func WithCORSAllowedMethods(methods ...string) CORSOption {
	return CORSOption{apply: internalcors.WithAllowedMethods(methods...)}
}

func WithCORSAllowedHeaders(headers ...string) CORSOption {
	return CORSOption{apply: internalcors.WithAllowedHeaders(headers...)}
}

func WithCORSExposeHeaders(headers ...string) CORSOption {
	return CORSOption{apply: internalcors.WithExposeHeaders(headers...)}
}

func WithCORSCredentials(allow bool) CORSOption {
	return CORSOption{apply: internalcors.WithCredentials(allow)}
}

func WithCORSOriginWildcard(patterns ...string) CORSOption {
	return CORSOption{apply: internalcors.WithOriginWildcard(patterns...)}
}

func WithCORSMaxAge(seconds int) CORSOption {
	return CORSOption{apply: internalcors.WithMaxAge(seconds)}
}

func UseCORS(rtr *Router, opts ...CORSOption) {
	internalOpts := make([]internalcors.CORSOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	internalcors.UseCORS(rtr.inner, internalOpts...)
}

type ForwardedHeadersOption struct {
	apply internalforwardheaders.ForwardHeadersOption
}

func WithForwardedTrustAll() ForwardedHeadersOption {
	return ForwardedHeadersOption{apply: internalforwardheaders.WithTrustAll()}
}

func WithForwardedTrustedProxies(proxies ...string) ForwardedHeadersOption {
	return ForwardedHeadersOption{apply: internalforwardheaders.WithTrustedProxies(proxies...)}
}

func WithForwardedRespectHeader(respect bool) ForwardedHeadersOption {
	return ForwardedHeadersOption{apply: internalforwardheaders.WithRespectForwarded(respect)}
}

func UseForwardedHeaders(rtr *Router, opts ...ForwardedHeadersOption) {
	internalOpts := make([]internalforwardheaders.ForwardHeadersOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	internalforwardheaders.UseForwardedHeaders(rtr.inner, internalOpts...)
}

type OpenTelemetryOption struct {
	apply internalopentelemetry.OpenTelemetryOption
}

func WithTelemetryOperation(operation string) OpenTelemetryOption {
	return OpenTelemetryOption{apply: internalopentelemetry.WithOperation(operation)}
}

func UseOpenTelemetry(rtr *Router, opts ...OpenTelemetryOption) {
	internalOpts := make([]internalopentelemetry.OpenTelemetryOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	internalopentelemetry.UseOpenTelemetry(rtr.inner, internalOpts...)
}

type AuthOption struct {
	apply internalauthentication.AuthOption
}

func WithAuthTokenTTL(ttl time.Duration) AuthOption {
	return AuthOption{apply: internalauthentication.WithTokenTTL(ttl)}
}

func WithAuthValidator(fn func(string) (claims.Principal, error)) AuthOption {
	return AuthOption{apply: internalauthentication.WithValidator(fn)}
}

func WithAuthTokenCreator(fn func(claims.Principal, time.Duration) (string, error)) AuthOption {
	return AuthOption{apply: internalauthentication.WithTokenCreator(fn)}
}

func WithAuthAppSessionCookieName(name string) AuthOption {
	return AuthOption{apply: internalauthentication.WithAppSessionCookieName(name)}
}

func WithAuthTwoFactorCookieName(name string) AuthOption {
	return AuthOption{apply: internalauthentication.WithTwoFactorCookieName(name)}
}

func WithAuthIDPSessionCookieName(name string) AuthOption {
	return AuthOption{apply: internalauthentication.WithIDPSessionCookieName(name)}
}

func WithAuthCSRFProtection() AuthOption {
	return AuthOption{apply: internalauthentication.WithCSRFProtection()}
}

func WithAuthRateLimiter(fn func(string) bool) AuthOption {
	return AuthOption{apply: internalauthentication.WithRateLimiter(fn)}
}

func WithAuthTokenRevocationChecker(fn func(string) bool) AuthOption {
	return AuthOption{apply: internalauthentication.WithTokenRevocationChecker(fn)}
}

func WithAuthIssuerValidator(issuer string) AuthOption {
	return AuthOption{apply: internalauthentication.WithIssuerValidator(issuer)}
}

func WithAuthAudienceValidator(audience string) AuthOption {
	return AuthOption{apply: internalauthentication.WithAudienceValidator(audience)}
}

func WithAuthContextEnricher(fn func(context.Context, claims.Principal) context.Context) AuthOption {
	return AuthOption{apply: internalauthentication.WithContextEnricher(func(c internalrouting.RouteContext) {
		if fn == nil || c == nil {
			return
		}
		principal := c.User()
		enriched := fn(c.Request().Context(), principal)
		if enriched == nil {
			return
		}
		c.SetRequest(c.Request().WithContext(enriched))
	})}
}

func UseAuthentication(rtr *Router, opts ...AuthOption) {
	internalOpts := make([]internalauthentication.AuthOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	internalauthentication.UseAuthentication(rtr.inner, internalOpts...)
}

func UseAuthenticationWithProvider(rtr *Router, provider TokenProvider, opts ...AuthOption) {
	internalOpts := make([]internalauthentication.AuthOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	internalauthentication.UseAuthenticationWithProvider(rtr.inner, provider, internalOpts...)
}

func NewInMemoryRateLimiter(maxAttempts int, window time.Duration) func(string) bool {
	return internalauthentication.NewInMemoryRateLimiter(maxAttempts, window)
}

type AuthorizationOption struct {
	apply internalauthorization.AuthZOption
}

func WithAuthorizationRoles(roles ...string) AuthorizationOption {
	return AuthorizationOption{apply: internalauthorization.WithRoles(roles...)}
}

func WithAuthorizationScopes(scopes ...string) AuthorizationOption {
	return AuthorizationOption{apply: internalauthorization.WithScopes(scopes...)}
}

func WithAuthorizationPermissions(perms ...string) AuthorizationOption {
	return AuthorizationOption{apply: internalauthorization.WithPermissions(perms...)}
}

func WithAuthorizationRoleChecker(fn func(claims.Principal, []string) bool) AuthorizationOption {
	return AuthorizationOption{apply: internalauthorization.WithRoleChecker(fn)}
}

func WithAuthorizationScopeChecker(fn func(claims.Principal, []string) bool) AuthorizationOption {
	return AuthorizationOption{apply: internalauthorization.WithScopeChecker(fn)}
}

func WithAuthorizationPermissionChecker(fn func(claims.Principal, []string) bool) AuthorizationOption {
	return AuthorizationOption{apply: internalauthorization.WithPermissionChecker(fn)}
}

func UseAuthorization(rtr *Router, opts ...AuthorizationOption) {
	internalOpts := make([]internalauthorization.AuthZOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	internalauthorization.UseAuthorization(rtr.inner, internalOpts...)
}

func UseEnforceHTTPS(rtr *Router) {
	internalenforcehttps.UseEnforceHTTPS(rtr.inner)
}

type ExportControlOption struct {
	apply internalexportcontrol.ExportControlOption
}

func WithExportControlGeoIPDatabase(db *geoip2.Reader) ExportControlOption {
	return ExportControlOption{apply: internalexportcontrol.WithGeoIPDatabase(db)}
}

func UseExportControl(rtr *Router, opts ...ExportControlOption) {
	internalOpts := make([]internalexportcontrol.ExportControlOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	internalexportcontrol.UseExportControl(rtr.inner, internalOpts...)
}

type RateLimiterOption struct {
	apply internalratelimit.RateLimiterOption
}

func WithRateLimitCleanupInterval(interval time.Duration) RateLimiterOption {
	return RateLimiterOption{apply: internalratelimit.WithCleanupInterval(interval)}
}

type RateLimiter struct {
	inner *internalratelimit.SelectiveRateLimiter
}

func NewRateLimiter(opts ...RateLimiterOption) *RateLimiter {
	internalOpts := make([]internalratelimit.RateLimiterOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	return &RateLimiter{inner: internalratelimit.NewSelectiveRateLimiter(internalOpts...)}
}

func NewRateLimiterWithContext(ctx context.Context, opts ...RateLimiterOption) *RateLimiter {
	internalOpts := make([]internalratelimit.RateLimiterOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internalOpts = append(internalOpts, opt.apply)
		}
	}
	return &RateLimiter{inner: internalratelimit.NewSelectiveRateLimiterWithContext(ctx, internalOpts...)}
}

func (r *RateLimiter) Stop() {
	if r != nil && r.inner != nil {
		r.inner.Stop()
	}
}

func (r *RateLimiter) Invoke(c MutableRouteContext, next HandlerFunc) {
	if r == nil || r.inner == nil {
		next(c)
		return
	}
	innerCtx := unwrapRouteContext(c)
	if innerCtx == nil {
		next(c)
		return
	}
	r.inner.Invoke(innerCtx, func(nextCtx internalrouting.RouteContext) {
		next(wrapRouteContext(nextCtx))
	})
}

func UseRateLimiter(rtr *Router, opts ...RateLimiterOption) {
	rtr.Use(NewRateLimiter(opts...))
}

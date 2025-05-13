package mux

import (
	"time"
)

type RouteBuilder struct {
	Pattern string
	Options *RouteOptions
}

func (rb *RouteBuilder) AllowAnonymous() *RouteBuilder {
	rb.Options.AllowAnonymous = true
	return rb
}

func (rb *RouteBuilder) RequirePermission(resource string, permissions ...string) *RouteBuilder {
	rb.Options.Permissions = append(rb.Options.Permissions, permissions...)
	return rb
}

func (rb *RouteBuilder) RequireRoles(roles ...string) *RouteBuilder {
	rb.Options.Roles = append(rb.Options.Roles, roles...)
	return rb
}

func (rb *RouteBuilder) RequireScopes(scopes ...string) *RouteBuilder {
	rb.Options.Scopes = append(rb.Options.Scopes, scopes...)
	return rb
}

func (rb *RouteBuilder) WithRateLimit(limit int, interval time.Duration) *RouteBuilder {
	rb.Options.RateLimit = limit
	rb.Options.RateInterval = interval
	return rb
}

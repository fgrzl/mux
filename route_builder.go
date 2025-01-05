package mux

type RouteBuilder struct {
	Pattern string
	Options *RouteOptions
}

func (rb *RouteBuilder) AllowAnonymous() *RouteBuilder {
	rb.Options.AllowAnonymous = true
	return rb
}

func (rb *RouteBuilder) RequirePermission(resource string, permissions ...string) *RouteBuilder {
	for _, p := range permissions {
		rb.Options.Permissions = append(rb.Options.Permissions, p)
	}
	return rb
}

func (rb *RouteBuilder) RequireRoles(roles ...string) *RouteBuilder {
	for _, r := range roles {
		rb.Options.Roles = append(rb.Options.Roles, r)
	}
	return rb
}

func (rb *RouteBuilder) RequireScopes(scopes ...string) *RouteBuilder {
	for _, s := range scopes {
		rb.Options.Scopes = append(rb.Options.Scopes, s)
	}
	return rb
}
